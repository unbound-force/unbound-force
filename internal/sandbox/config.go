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
func forwardedEnvVars(opts Options) []string {
	var args []string
	for _, key := range forwardedAPIKeys {
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
//  2. If GOOGLE_APPLICATION_CREDENTIALS is not set, check for
//     gcloud Application Default Credentials at the standard
//     path (~/.config/gcloud/application_default_credentials.json)
//     and mount it if present.
func googleCloudCredentialMounts(opts Options, platform PlatformConfig) []string {
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

	// Strategy 2: gcloud ADC fallback.
	home, err := os.UserHomeDir()
	if err != nil {
		return args
	}
	adcPath := filepath.Join(home, ".config", "gcloud", "application_default_credentials.json")
	if _, err := os.Stat(adcPath); err == nil {
		containerPath := "/home/dev/.config/gcloud/application_default_credentials.json"
		mount := fmt.Sprintf("%s:%s:ro", adcPath, containerPath)
		if platform.SELinux {
			mount += ",Z"
		}
		args = append(args, "-v", mount)
	}
	return args
}

// buildVolumeMounts constructs -v flags for the project
// directory mount. Isolated mode uses :ro (read-only),
// direct mode uses read-write. The :Z suffix is appended
// when SELinux is enforcing (per research.md R3).
func buildVolumeMounts(opts Options, platform PlatformConfig) []string {
	mount := fmt.Sprintf("%s:/workspace", opts.ProjectDir)
	if opts.Mode == ModeIsolated {
		mount += ":ro"
	}
	if platform.SELinux {
		mount += ",Z"
	}
	return []string{"-v", mount}
}

// buildRunArgs assembles the complete podman run argument list
// from Options and PlatformConfig. All values are passed as
// discrete exec.Command arguments — never shell-interpolated —
// preventing command injection (per contracts/sandbox-api.md).
func buildRunArgs(opts Options, platform PlatformConfig) []string {
	args := []string{
		"run", "-d",
		"--name", ContainerName,
		"--hostname", ContainerName,
		"-p", fmt.Sprintf("%d:%d", DefaultServerPort, DefaultServerPort),
	}

	// Volume mounts.
	args = append(args, buildVolumeMounts(opts, platform)...)

	// Environment variables.
	args = append(args, forwardedEnvVars(opts)...)

	// Google Cloud credential mounts (FR-021, FR-022).
	args = append(args, googleCloudCredentialMounts(opts, platform)...)

	// Resource limits.
	args = append(args, "--memory", opts.Memory)
	args = append(args, "--cpus", opts.CPUs)

	// Image name (last argument).
	args = append(args, opts.Image)

	return args
}
