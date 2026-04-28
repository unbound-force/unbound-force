// Package sandbox implements containerized OpenCode session
// management via Podman. It provides Start, Stop, Attach,
// Extract, and Status operations for isolated development
// environments. All external dependencies are injected for
// testability per Constitution Principle IV.
package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
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

	// HealthTimeout is the maximum time to wait for the
	// OpenCode server health check (FR-005).
	HealthTimeout = 60 * time.Second

	// ModeIsolated mounts the project directory read-only.
	ModeIsolated = "isolated"

	// ModeDirect mounts the project directory read-write.
	ModeDirect = "direct"
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
	"ANTHROPIC_API_KEY":            true,
	"ANTHROPIC_VERTEX_PROJECT_ID":  true,
	"CLAUDE_CODE_USE_VERTEX":       true,
	"GOOGLE_CLOUD_PROJECT":         true,
	"VERTEX_LOCATION":              true,
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

// googleCloudCredentialMounts returns -v and -e flags for
// Google Cloud authentication inside the container (FR-021,
// FR-022). Two strategies:
//
//  1. If GOOGLE_APPLICATION_CREDENTIALS is set and the file
//     exists, mount it read-only and set the env var to the
//     container-internal path.
//  2. If GOOGLE_APPLICATION_CREDENTIALS is not set, mount the
//     entire ~/.config/gcloud/ directory read-write so the
//     auth library can read credentials and write refreshed
//     access tokens (needed for authorized_user credentials).
//
// When gatewayActive is true, all gcloud credential mounts
// are skipped because the gateway handles authentication
// on the host side (FR-011).
func googleCloudCredentialMounts(opts Options, platform PlatformConfig, gatewayActive bool) []string {
	if gatewayActive {
		return nil
	}
	var args []string

	// Strategy 1: explicit service account key file.
	if gac := opts.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); gac != "" {
		if _, err := os.Stat(gac); err == nil {
			containerPath := "/home/dev/.config/gcloud/service-account.json"
			mount := fmt.Sprintf("%s:%s:ro", gac, containerPath)
			if platform.SELinux {
				mount += ",Z"
			}
			args = append(args, "-v", mount)
			args = append(args, "-e", "GOOGLE_APPLICATION_CREDENTIALS="+containerPath)
		}
		return args
	}

	// Strategy 2: mount entire gcloud config directory.
	// The authorized_user credential type needs access_tokens.db
	// and credentials.db for token refresh. Mount read-write so
	// the auth library can update cached access tokens.
	home, err := os.UserHomeDir()
	if err != nil {
		return args
	}
	gcloudDir := filepath.Join(home, ".config", "gcloud")
	if _, err := os.Stat(gcloudDir); err == nil {
		containerPath := "/home/dev/.config/gcloud"
		mount := fmt.Sprintf("%s:%s", gcloudDir, containerPath)
		if platform.SELinux {
			mount += ":Z"
		}
		args = append(args, "-v", mount)
		// Set GOOGLE_APPLICATION_CREDENTIALS to the ADC file
		// inside the mounted directory so the auth library
		// finds it without searching.
		adcContainer := containerPath + "/application_default_credentials.json"
		args = append(args, "-e", "GOOGLE_APPLICATION_CREDENTIALS="+adcContainer)
	}
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

	// UID/GID mapping (before volume mounts).
	args = append(args, uidMappingArgs(opts)...)

	// Volume mounts.
	args = append(args, buildVolumeMounts(opts, platform)...)

	// Environment variables (gateway-aware).
	args = append(args, forwardedEnvVars(opts, gatewayActive)...)

	// Gateway env vars or credential mounts.
	if gatewayActive {
		args = append(args, gatewayEnvVars(gatewayPort)...)
	}

	// Google Cloud credential mounts (FR-021, FR-022).
	// Skipped when gateway is active.
	args = append(args, googleCloudCredentialMounts(opts, platform, gatewayActive)...)

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
