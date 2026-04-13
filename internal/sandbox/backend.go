package sandbox

import "fmt"

// Backend constants for workspace provisioning.
const (
	// BackendAuto auto-detects the backend.
	BackendAuto = "auto"

	// BackendPodman selects the Podman backend.
	BackendPodman = "podman"

	// BackendChe selects the Eclipse Che backend.
	BackendChe = "che"

	// ModePersistent is the mode for persistent workspaces
	// (named volumes or CDE storage).
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

	// Name returns the backend identifier ("podman" or
	// "che").
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
//  4. Auto-detect: CDE if chectl/UF_CHE_URL available,
//     Podman otherwise
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

	case BackendChe:
		cheURL := resolveCheURL(opts)
		hasChectl := false
		if _, err := opts.LookPath("chectl"); err == nil {
			hasChectl = true
		}
		if cheURL == "" && !hasChectl {
			return nil, fmt.Errorf(
				"CDE backend requested but not configured, set UF_CHE_URL or install chectl")
		}
		return &CheBackend{cheURL: cheURL, useChectl: hasChectl}, nil

	case BackendAuto:
		return autoDetectBackend(opts)

	default:
		return nil, fmt.Errorf(
			"unknown backend: %s, use 'auto', 'podman', or 'che'", backendName)
	}
}

// autoDetectBackend selects the best available backend.
// CDE is preferred when chectl or UF_CHE_URL is available;
// otherwise Podman is selected.
func autoDetectBackend(opts Options) (Backend, error) {
	cheURL := resolveCheURL(opts)
	hasChectl := false
	if _, err := opts.LookPath("chectl"); err == nil {
		hasChectl = true
	}

	// Prefer CDE when available.
	if cheURL != "" || hasChectl {
		return &CheBackend{cheURL: cheURL, useChectl: hasChectl}, nil
	}

	// Fall back to Podman.
	return &PodmanBackend{}, nil
}

// resolveCheURL returns the Che server URL from flag > env > config.
func resolveCheURL(opts Options) string {
	if opts.CheURL != "" {
		return opts.CheURL
	}
	if envURL := opts.Getenv("UF_CHE_URL"); envURL != "" {
		return envURL
	}
	cfg, _ := LoadConfig(opts)
	return cfg.Che.URL
}
