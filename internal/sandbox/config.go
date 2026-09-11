// Package sandbox implements containerized OpenCode session
// management via Podman. It provides Start, Stop, Attach,
// Extract, and Status operations for isolated development
// environments. All external dependencies are injected for
// testability per Constitution Principle IV.
package sandbox

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Container and resource defaults.
const (
	// ContainerName is the fixed name for the sandbox
	// container. Only one sandbox is supported at a time.
	ContainerName = "uf-sandbox"

	// DefaultImage is the default container image.
	DefaultImage = "quay.io/unbound-force/opencode-dev:latest"

	// DefaultMemory is the default memory limit.
	DefaultMemory = "8g"

	// DefaultCPUs is the default CPU limit.
	DefaultCPUs = "4"

	// DefaultServerPort is the OpenCode server port.
	DefaultServerPort = 4096

	// ModeIsolated mounts the project directory read-only.
	ModeIsolated = "isolated"

	// ModeDirect mounts the project directory read-write.
	ModeDirect = "direct"

	// DefaultIDE is the default IDE value for DevPod
	// workspaces. "none" means no IDE is opened.
	DefaultIDE = "none"

	// HealthTimeout is the maximum time to wait for the
	// OpenCode server health check (FR-005).
	HealthTimeout = 60 * time.Second
)

// forwardedAPIKeys lists environment variable names that are
// forwarded from the host to the container using Podman's
// -e VAR syntax (value read from host environment at runtime).
var forwardedAPIKeys = []string{
	"ANTHROPIC_API_KEY",
	"OPENAI_API_KEY",
	"GEMINI_API_KEY",
	"OPENROUTER_API_KEY",
	// Google Vertex AI (FR-020).
	"GOOGLE_CLOUD_PROJECT",
	"VERTEX_LOCATION",
	// Anthropic via Vertex (Claude models on GCP).
	"ANTHROPIC_VERTEX_PROJECT_ID",
	"CLAUDE_CODE_USE_VERTEX",
}

// gatewaySkippedKeys lists environment variable names that
// are NOT forwarded to the container when the gateway is
// active. The gateway handles authentication for these
// providers, so their credentials must not leak into the
// container (FR-011).
var gatewaySkippedKeys = map[string]bool{
	"ANTHROPIC_API_KEY":           true,
	"ANTHROPIC_VERTEX_PROJECT_ID": true,
	"CLAUDE_CODE_USE_VERTEX":      true,
	"GOOGLE_CLOUD_PROJECT":        true,
	"VERTEX_LOCATION":             true,
}

// gatewayEnvVars returns -e flag pairs for the gateway's
// container-internal URL and auth token. The container uses
// host.containers.internal to reach the host's gateway
// process (FR-011).
func gatewayEnvVars(port int) []string {
	return []string{
		"-e", fmt.Sprintf("ANTHROPIC_BASE_URL=http://host.containers.internal:%d", port),
		"-e", "ANTHROPIC_API_KEY=gateway",
	}
}

// DefaultConfig resolves image, memory, and CPU settings from
// flag values → environment variables → constant defaults.
// Flag values (already set on opts) take highest precedence.
func DefaultConfig(opts Options) Options {
	if opts.Image == "" {
		if envImg := opts.Getenv("UF_SANDBOX_IMAGE"); envImg != "" {
			opts.Image = envImg
		} else {
			opts.Image = DefaultImage
		}
	}
	if opts.Memory == "" {
		opts.Memory = DefaultMemory
	}
	if opts.CPUs == "" {
		opts.CPUs = DefaultCPUs
	}
	if opts.Mode == "" {
		opts.Mode = ModeIsolated
	}
	if opts.IDE == "" {
		if envIDE := opts.Getenv("UF_SANDBOX_IDE"); envIDE != "" {
			opts.IDE = envIDE
		} else {
			opts.IDE = DefaultIDE
		}
	}
	return opts
}

// forwardedEnvVars returns -e flag pairs for API keys and
// the Ollama host override. API keys use -e VAR syntax so
// Podman reads the value from the host environment. Ollama
// host is set explicitly to the container-internal hostname
// that resolves to the host machine (per research.md R7).
//
// When gatewayActive is true, provider-specific keys
// (ANTHROPIC_API_KEY, ANTHROPIC_VERTEX_PROJECT_ID,
// CLAUDE_CODE_USE_VERTEX, etc.) are skipped because the
// gateway handles authentication (FR-011). Non-proxied keys
// (OPENAI_API_KEY, GEMINI_API_KEY, OPENROUTER_API_KEY) are
// always forwarded.
func forwardedEnvVars(opts Options, gatewayActive bool) []string {
	var args []string
	for _, key := range forwardedAPIKeys {
		if gatewayActive && gatewaySkippedKeys[key] {
			continue
		}
		if v := opts.Getenv(key); v != "" {
			args = append(args, "-e", key)
		}
	}
	// Always set OLLAMA_HOST to the container-internal
	// hostname so containerized tools can reach the host's
	// Ollama instance.
	args = append(args, "-e", "OLLAMA_HOST=host.containers.internal:11434")
	return args
}

// useParentMount returns true if the parent directory
// should be mounted instead of the project directory.
// Falls back to project-only mount when NoParent is set
// or when the parent is the filesystem root (FR-042).
func useParentMount(opts Options) bool {
	if opts.NoParent {
		return false
	}
	parent := filepath.Dir(opts.ProjectDir)
	return parent != "/" && parent != opts.ProjectDir
}

// buildVolumeMounts constructs -v flags for the workspace
// mount. By default, mounts the project's parent directory
// at /workspace so sibling repos are accessible via
// relative paths (e.g., ../dewey). The container's workdir
// is set to /workspace/<project-basename> by buildRunArgs.
// When NoParent is true or the project is at the filesystem
// root, mounts only the project directory (FR-040, FR-041,
// FR-042). Isolated mode uses :ro, SELinux uses :Z (FR-043).
func buildVolumeMounts(opts Options, platform PlatformConfig) []string {
	mountSource := opts.ProjectDir
	if useParentMount(opts) {
		mountSource = filepath.Dir(opts.ProjectDir)
	}
	mount := fmt.Sprintf("%s:/workspace", mountSource)
	if opts.Mode == ModeIsolated {
		mount += ":ro"
	}
	if platform.SELinux {
		mount += ",Z"
	}
	return []string{"-v", mount}
}

// uidMappingArgs returns the Podman user namespace flags for
// UID/GID mapping inside the container. By default, uses
// --userns=keep-id:uid=1000,gid=1000 which maps the host
// user to UID 1000 (the "dev" user) inside the container.
//
// When opts.UIDMap is true, returns explicit --uidmap/--gidmap
// flags instead. This is the fallback for macOS Podman machines
// where virtiofs does not support keep-id UID mapping.
func uidMappingArgs(opts Options) []string {
	if opts.UIDMap {
		return []string{
			"--uidmap", "1000:0:1",
			"--uidmap", "0:1:1000",
			"--uidmap", "1001:1001:64536",
			"--gidmap", "1000:0:1",
			"--gidmap", "0:1:1000",
			"--gidmap", "1001:1001:64536",
		}
	}
	return []string{"--userns=keep-id:uid=1000,gid=1000"}
}

// portMapping represents a host:container port pair parsed
// from a devcontainer.json forwardPorts entry.
type portMapping struct {
	host      int
	container int
}

// advanceStringLiteral copies a JSON string literal (opening
// quote already detected at src[i]) into out, including the
// opening and closing quotes, handling backslash escapes. It
// returns the index of the first character after the closing
// quote. Both stripJSONComments and stripTrailingCommas use
// this to skip string contents without misinterpreting their
// characters as comment markers or trailing commas.
func advanceStringLiteral(src string, out *strings.Builder, i int) int {
	out.WriteByte(src[i]) // opening quote
	i++
	for i < len(src) {
		out.WriteByte(src[i])
		if src[i] == '\\' {
			i++
			if i < len(src) {
				out.WriteByte(src[i])
			}
		} else if src[i] == '"' {
			break
		}
		i++
	}
	return i + 1 // past closing quote
}

// stripJSONComments removes single-line (//) and block (/* */)
// comments from JSONC input, preserving string contents. The
// devcontainer spec uses JSONC (JSON with Comments) as the
// canonical format for devcontainer.json.
func stripJSONComments(data []byte) []byte {
	src := string(data)
	var out strings.Builder
	out.Grow(len(src))
	i := 0
	for i < len(src) {
		// String literal — copy verbatim.
		if src[i] == '"' {
			i = advanceStringLiteral(src, &out, i)
			continue
		}
		// Line comment.
		if i+1 < len(src) && src[i] == '/' && src[i+1] == '/' {
			for i < len(src) && src[i] != '\n' {
				i++
			}
			continue
		}
		// Block comment.
		if i+1 < len(src) && src[i] == '/' && src[i+1] == '*' {
			i += 2
			for i+1 < len(src) && !(src[i] == '*' && src[i+1] == '/') {
				i++
			}
			if i+1 < len(src) {
				i += 2
			}
			continue
		}
		out.WriteByte(src[i])
		i++
	}
	return []byte(out.String())
}

// stripTrailingCommas removes trailing commas before ] or }
// in JSON input, preserving string contents. This handles
// the JSONC convention of allowing trailing commas.
func stripTrailingCommas(data []byte) []byte {
	src := string(data)
	var out strings.Builder
	out.Grow(len(src))
	i := 0
	for i < len(src) {
		// String literal — copy verbatim.
		if src[i] == '"' {
			i = advanceStringLiteral(src, &out, i)
			continue
		}
		// Trailing comma — skip if next non-whitespace is ] or }.
		if src[i] == ',' {
			j := i + 1
			for j < len(src) && (src[j] == ' ' || src[j] == '\t' || src[j] == '\n' || src[j] == '\r') {
				j++
			}
			if j < len(src) && (src[j] == ']' || src[j] == '}') {
				i++
				continue
			}
		}
		out.WriteByte(src[i])
		i++
	}
	return []byte(out.String())
}

// parseDevcontainerPorts reads .devcontainer/devcontainer.json
// from the project directory via opts.ReadFile and returns the
// forwardPorts array as port mappings. Returns nil with no
// error when the file is absent or contains no forwardPorts.
// Ports whose host port is already published (DefaultServerPort
// and demo ports) are excluded from the result to avoid
// duplicates.
//
// The devcontainer spec defines forwardPorts as
// Array<number | string>, where string values represent
// host:container port mappings (e.g., "8080:3000"). This
// function handles both forms and strips JSONC comments and
// trailing commas before parsing.
func parseDevcontainerPorts(opts Options, excludePorts map[int]bool) []portMapping {
	dcPath := filepath.Join(opts.ProjectDir,
		".devcontainer", "devcontainer.json")
	data, err := opts.ReadFile(dcPath)
	if err != nil {
		return nil
	}

	// Strip JSONC comments and trailing commas before
	// unmarshaling.
	data = stripTrailingCommas(stripJSONComments(data))

	var dc struct {
		ForwardPorts []json.RawMessage `json:"forwardPorts"`
	}
	if err := json.Unmarshal(data, &dc); err != nil {
		return nil
	}

	seen := make(map[int]bool)
	var ports []portMapping
	for _, raw := range dc.ForwardPorts {
		pm, ok := parsePortEntry(raw)
		if !ok {
			continue
		}
		if pm.host < 1 || pm.host > 65535 {
			continue
		}
		if pm.container < 1 || pm.container > 65535 {
			continue
		}
		if excludePorts[pm.host] || seen[pm.host] {
			continue
		}
		seen[pm.host] = true
		ports = append(ports, pm)
	}
	return ports
}

// parsePortEntry extracts the host and container ports from a
// single forwardPorts entry. Handles JSON numbers (8080) and
// strings ("8080" or "8080:3000"). For plain numbers and plain
// strings, host and container are the same. For "host:container"
// strings, returns distinct values. Returns false on failure.
func parsePortEntry(raw json.RawMessage) (portMapping, bool) {
	// Try as number first.
	var n float64
	if err := json.Unmarshal(raw, &n); err == nil {
		p := int(n)
		return portMapping{host: p, container: p}, true
	}

	// Try as string.
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return portMapping{}, false
	}

	// "host:container" format.
	if idx := strings.IndexByte(s, ':'); idx >= 0 {
		host, err1 := strconv.Atoi(s[:idx])
		container, err2 := strconv.Atoi(s[idx+1:])
		if err1 != nil || err2 != nil {
			return portMapping{}, false
		}
		return portMapping{host: host, container: container}, true
	}
	port, err := strconv.Atoi(s)
	if err != nil {
		return portMapping{}, false
	}
	return portMapping{host: port, container: port}, true
}

// buildRunArgs assembles the complete podman run argument list
// from Options and PlatformConfig. All values are passed as
// discrete exec.Command arguments — never shell-interpolated —
// preventing command injection (per contracts/sandbox-api.md).
//
// When gatewayActive is true, gateway env vars are added
// (ANTHROPIC_BASE_URL, ANTHROPIC_API_KEY=gateway) and credential
// mounts and provider API key forwarding are skipped (FR-011).
func buildRunArgs(opts Options, platform PlatformConfig, gatewayActive bool, gatewayPort int) []string {
	args := []string{
		"run", "-d",
		"--name", ContainerName,
		"--hostname", ContainerName,
		"-p", fmt.Sprintf("%d:%d", DefaultServerPort, DefaultServerPort),
	}

	// Devcontainer forwardPorts: read from
	// .devcontainer/devcontainer.json and publish any ports
	// not already covered by DefaultServerPort.
	excludePorts := map[int]bool{DefaultServerPort: true}
	for _, pm := range parseDevcontainerPorts(opts, excludePorts) {
		args = append(args, "-p", fmt.Sprintf("%d:%d", pm.host, pm.container))
	}

	// UID/GID mapping (before volume mounts).
	args = append(args, uidMappingArgs(opts)...)

	// Volume mounts.
	args = append(args, buildVolumeMounts(opts, platform)...)

	// Environment variables (gateway-aware).
	args = append(args, forwardedEnvVars(opts, gatewayActive)...)

	// Gateway env vars (container reaches host gateway via
	// host.containers.internal). Credential mounts removed
	// in favor of gateway-based credential isolation.
	if gatewayActive {
		args = append(args, gatewayEnvVars(gatewayPort)...)
	}

	// Resource limits.
	args = append(args, "--memory", opts.Memory)
	args = append(args, "--cpus", opts.CPUs)

	// Working directory: when parent mount is active,
	// set workdir to the project subdirectory within
	// the parent mount (FR-040). Also set WORKSPACE
	// env var so the entrypoint's cd "$WORKSPACE" goes
	// to the project, not the parent (FR-044).
	if useParentMount(opts) {
		projectSubdir := fmt.Sprintf("/workspace/%s",
			filepath.Base(opts.ProjectDir))
		args = append(args, "--workdir", projectSubdir)
		args = append(args, "-e",
			fmt.Sprintf("WORKSPACE=%s", projectSubdir))
	}

	// Image name (last argument).
	args = append(args, opts.Image)

	return args
}
