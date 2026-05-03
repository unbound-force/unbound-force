package sandbox

import (
	"fmt"
	"path/filepath"
)

// Backend constants for workspace provisioning.
const (
	// BackendAuto auto-detects the backend.
	BackendAuto = "auto"

	// BackendPodman selects the Podman backend.
	BackendPodman = "podman"

	// BackendDevPod selects the DevPod backend.
	BackendDevPod = "devpod"

	// ModePersistent is the mode for persistent workspaces
	// (named volumes or DevPod storage).
	ModePersistent = "persistent"

	// DefaultConfigPath is the default sandbox config file
	// path relative to the project directory.
	DefaultConfigPath = ".uf/sandbox.yaml"
)

// Backend defines the interface for workspace lifecycle
// operations. Implementations handle the infrastructure-
// specific details of provisioning, starting, stopping,
// and destroying workspaces.
//
// Design decision: Strategy pattern per SOLID Open/Closed
// Principle. Adding a new backend (e.g., Docker, K8s
// direct) requires only a new implementation, not
// modification of the orchestration layer.
type Backend interface {
	// Create provisions a persistent workspace with the
	// project's source code and toolchain. Returns an error
	// if the workspace already exists.
	Create(opts Options) error

	// Start starts a stopped workspace without losing
	// state. For PodmanBackend, falls back to ephemeral
	// mode if no persistent workspace exists.
	Start(opts Options) error

	// Stop stops a running workspace while preserving
	// all state. Idempotent — returns nil if already
	// stopped.
	Stop(opts Options) error

	// Destroy permanently deletes the workspace and all
	// associated state. Idempotent — returns nil if no
	// workspace exists.
	Destroy(opts Options) error

	// Status returns the current state of the workspace.
	// Returns a zero-value WorkspaceStatus if no workspace
	// exists.
	Status(opts Options) (WorkspaceStatus, error)

	// Attach connects the TUI to the running workspace's
	// OpenCode server.
	Attach(opts Options) error

	// Name returns the backend identifier (e.g., "podman").
	Name() string
}

// ResolveBackend selects the appropriate Backend
// implementation based on Options, environment, and
// configuration.
//
// Resolution order:
//  1. --backend flag (explicit selection)
//  2. UF_SANDBOX_BACKEND env var
//  3. .uf/sandbox.yaml backend field
//  4. Auto-detect: Podman (default)
func ResolveBackend(opts Options) (Backend, error) {
	opts.defaults()

	// Determine the requested backend name from flag > env > config > auto.
	backendName := opts.BackendName
	if backendName == "" {
		backendName = opts.Getenv("UF_SANDBOX_BACKEND")
	}
	if backendName == "" {
		cfg, _ := LoadConfig(opts)
		if cfg.Backend != "" {
			backendName = cfg.Backend
		}
	}
	if backendName == "" {
		backendName = BackendAuto
	}

	switch backendName {
	case BackendPodman:
		if _, err := opts.LookPath("podman"); err != nil {
			return nil, fmt.Errorf("podman not found, install: brew install podman")
		}
		return &PodmanBackend{}, nil

	case BackendDevPod:
		if _, err := opts.LookPath("devpod"); err != nil {
			return nil, fmt.Errorf(
				"devpod not found, install: https://devpod.sh/docs/getting-started/install")
		}
		return &DevPodBackend{}, nil

	case "che":
		// Migration error: Che backend was removed in favor of DevPod.
		return nil, fmt.Errorf(
			"che backend removed, use --backend devpod instead. Install DevPod: https://devpod.sh/docs/getting-started/install")

	case BackendAuto:
		return autoDetectBackend(opts)

	default:
		return nil, fmt.Errorf(
			"unknown backend: %s, use 'auto', 'podman', or 'devpod'", backendName)
	}
}

// autoDetectBackend selects the best available backend.
// Prefers DevPod when devpod is in PATH AND
// .devcontainer/devcontainer.json exists in the project
// directory. Falls back to Podman otherwise.
func autoDetectBackend(opts Options) (Backend, error) {
	if _, err := opts.LookPath("devpod"); err == nil {
		dcPath := filepath.Join(opts.ProjectDir,
			".devcontainer", "devcontainer.json")
		if _, readErr := opts.ReadFile(dcPath); readErr == nil {
			return &DevPodBackend{}, nil
		}
	}
	return &PodmanBackend{}, nil
}
