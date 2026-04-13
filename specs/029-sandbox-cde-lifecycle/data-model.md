# Data Model: Sandbox CDE Lifecycle

**Branch**: `029-sandbox-cde-lifecycle` | **Date**: 2026-04-13

## New Types

### Backend

The core abstraction for workspace provisioning. Two
implementations: `PodmanBackend` and `CheBackend`.

```go
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
    // project's source code and toolchain. Idempotent —
    // returns an error if the workspace already exists.
    Create(opts Options) error

    // Start starts a stopped workspace without losing
    // state. Returns an error if no workspace exists
    // (must call Create first).
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
```

**Relationships**: Selected by `ResolveBackend()` based
on `--backend` flag, `UF_CHE_URL` env var, or auto-
detection. Consumed by all `uf sandbox` subcommands.

---

### WorkspaceStatus

Extended version of the existing `ContainerStatus` that
supports both Podman and CDE backends.

```go
// WorkspaceStatus represents the current state of a
// sandbox workspace, regardless of backend. Extends
// the Spec 028 ContainerStatus with backend-agnostic
// fields.
type WorkspaceStatus struct {
    // Exists is true when a workspace has been created
    // (via `uf sandbox create`).
    Exists bool

    // Running is true when the workspace is active.
    Running bool

    // Backend is the backend type ("podman" or "che").
    Backend string

    // Name is the workspace name
    // (e.g., "uf-sandbox-myproject").
    Name string

    // ID is the workspace identifier (container ID for
    // Podman, workspace ID for Che). Short form.
    ID string

    // Image is the container image or devfile used.
    Image string

    // Mode is the workspace mode. For Podman: "isolated"
    // or "direct". For CDE: always "persistent".
    Mode string

    // ProjectDir is the source project directory (host
    // path for Podman, repo URL for CDE).
    ProjectDir string

    // ServerURL is the OpenCode server URL. For Podman:
    // http://localhost:4096. For CDE: the Che endpoint URL.
    ServerURL string

    // DemoEndpoints lists exposed demo port URLs.
    DemoEndpoints []DemoEndpoint

    // StartedAt is the workspace start time.
    StartedAt string

    // ExitCode is set when the workspace has stopped.
    // -1 when running. Only applicable for Podman.
    ExitCode int

    // Persistent is true when the workspace uses named
    // volumes or CDE storage (survives stop/start).
    Persistent bool
}
```

**Relationships**: Produced by `Backend.Status()`,
consumed by `FormatWorkspaceStatus()` for display.
Supersedes `ContainerStatus` for new lifecycle commands
(existing `ContainerStatus` retained for backward
compatibility with Spec 028 code paths).

---

### DemoEndpoint

Represents an exposed port for demo review.

```go
// DemoEndpoint represents an exposed port in the
// workspace accessible for demo review.
type DemoEndpoint struct {
    // Name is the endpoint name from the devfile
    // (e.g., "demo-web", "demo-api").
    Name string

    // Port is the container-internal port number.
    Port int

    // URL is the externally accessible URL. For Podman:
    // http://localhost:<port>. For CDE: the Che route URL.
    URL string

    // Protocol is "http" or "https".
    Protocol string
}
```

**Relationships**: Populated from devfile endpoints
(CDE) or `--demo-ports` flag / config file (Podman).
Included in `WorkspaceStatus.DemoEndpoints`.

---

### SandboxConfig

Persistent configuration loaded from `.uf/sandbox.yaml`.

```go
// SandboxConfig is the persistent sandbox configuration
// loaded from `.uf/sandbox.yaml`. Provides defaults for
// CDE URL, Ollama endpoint, backend selection, and demo
// port mappings.
type SandboxConfig struct {
    // Che contains CDE backend configuration.
    Che CheConfig `yaml:"che"`

    // Ollama contains Ollama endpoint configuration.
    Ollama OllamaConfig `yaml:"ollama"`

    // Backend is the default backend: "auto", "podman",
    // or "che". Default: "auto".
    Backend string `yaml:"backend"`

    // DemoPorts lists port numbers to expose for demos
    // (Podman only; CDE uses devfile endpoints).
    DemoPorts []int `yaml:"demo_ports"`
}

// CheConfig contains Eclipse Che connection settings.
type CheConfig struct {
    // URL is the Che/Dev Spaces instance URL.
    // Can also be set via UF_CHE_URL env var.
    URL string `yaml:"url"`

    // Token is the authentication token for REST API.
    // Can also be set via UF_CHE_TOKEN env var.
    // Only needed when chectl is not available.
    Token string `yaml:"token"`
}

// OllamaConfig contains Ollama endpoint settings.
type OllamaConfig struct {
    // Host is the Ollama endpoint URL. Overrides the
    // default host.containers.internal:11434 for CDE
    // deployments where that hostname doesn't resolve.
    Host string `yaml:"host"`
}
```

**Relationships**: Loaded by `LoadConfig()` from
`.uf/sandbox.yaml`. Merged with CLI flags and env vars
(flag > env > config > default). Consumed by
`ResolveBackend()` for backend selection.

---

### PodmanBackend

```go
// PodmanBackend implements Backend for local Podman
// containers with named volumes for persistent state.
type PodmanBackend struct{}
```

**Methods**: `Create`, `Start`, `Stop`, `Destroy`,
`Status`, `Attach`, `Name`.

**Key behaviors**:
- `Create`: `podman volume create` + `podman run` +
  `podman cp` to seed workspace
- `Start`: `podman start` (resume stopped container)
- `Stop`: `podman stop` (preserve container + volume)
- `Destroy`: `podman rm` + `podman volume rm`
- `Status`: `podman inspect` + `podman volume inspect`

---

### CheBackend

```go
// CheBackend implements Backend for Eclipse Che / Dev
// Spaces workspace provisioning.
type CheBackend struct {
    // cheURL is the Che server URL.
    cheURL string

    // useChectl is true when chectl is available.
    useChectl bool
}
```

**Methods**: `Create`, `Start`, `Stop`, `Destroy`,
`Status`, `Attach`, `Name`.

**Key behaviors**:
- `Create`: `chectl workspace:create --devfile=...` or
  REST API `POST /api/workspace/devfile`
- `Start`: `chectl workspace:start` or REST API PATCH
- `Stop`: `chectl workspace:stop` or REST API PATCH
- `Destroy`: `chectl workspace:delete` or REST API DELETE
- `Status`: `chectl workspace:list --output=json` or
  REST API GET
- `Attach`: `opencode attach <che-endpoint-url>`

---

## Extended Existing Types

### Options (extended)

New fields added to the existing `Options` struct:

```go
// --- New fields for Spec 029 ---

// BackendName selects the backend: "auto", "podman",
// or "che". Default: "auto" (auto-detect).
BackendName string

// WorkspaceName overrides the auto-generated workspace
// name. Default: "uf-sandbox-<project-name>".
WorkspaceName string

// DemoPorts lists additional ports to expose for demos
// (Podman only). Merged with config file ports.
DemoPorts []int

// ConfigPath is the path to .uf/sandbox.yaml.
// Default: "<ProjectDir>/.uf/sandbox.yaml".
ConfigPath string

// CheURL is the Eclipse Che server URL. Overrides
// config file. Can also be set via UF_CHE_URL env var.
CheURL string
```

**Backward compatibility**: All new fields have zero-
value defaults that preserve Spec 028 behavior. When
`BackendName` is empty or "auto" and no CDE is
configured, the Podman backend is selected with
ephemeral mode (existing behavior).

---

## Constants (extended)

```go
const (
    // BackendAuto auto-detects the backend.
    BackendAuto = "auto"

    // BackendPodman selects the Podman backend.
    BackendPodman = "podman"

    // BackendChe selects the Eclipse Che backend.
    BackendChe = "che"

    // ModePersistent is the mode for CDE workspaces.
    ModePersistent = "persistent"

    // DefaultConfigPath is the default sandbox config
    // file path relative to the project directory.
    DefaultConfigPath = ".uf/sandbox.yaml"
)
```

---

## Environment Variables (extended)

| Variable | Purpose | Default |
|----------|---------|---------|
| `UF_CHE_URL` | Eclipse Che server URL | (none) |
| `UF_CHE_TOKEN` | Che REST API auth token | (none) |
| `UF_OLLAMA_HOST` | Ollama endpoint override for CDE | `host.containers.internal:11434` |
| `UF_SANDBOX_BACKEND` | Default backend override | `auto` |

All existing env vars from Spec 028 are preserved.

---

## Type Relationships

```text
Options ──────────────────────────────────────────┐
  │                                               │
  ├─► LoadConfig() ──► SandboxConfig              │
  │                       │                       │
  ├─► ResolveBackend() ◄──┘                       │
  │       │                                       │
  │       ├─► PodmanBackend                       │
  │       │     ├─► Create() ──► named volume     │
  │       │     ├─► Start()  ──► podman start     │
  │       │     ├─► Stop()   ──► podman stop      │
  │       │     ├─► Destroy()──► volume rm        │
  │       │     └─► Status() ──► WorkspaceStatus  │
  │       │                                       │
  │       └─► CheBackend                          │
  │             ├─► Create() ──► chectl/REST API  │
  │             ├─► Start()  ──► chectl/REST API  │
  │             ├─► Stop()   ──► chectl/REST API  │
  │             ├─► Destroy()──► chectl/REST API  │
  │             └─► Status() ──► WorkspaceStatus  │
  │                                               │
  ├─► DetectPlatform() ──► PlatformConfig         │
  │                           │                   │
  ├─► buildRunArgs() ◄────────┘                   │
  │                                               │
  ├─► projectName() ──► workspace name            │
  │                                               │
  └─► Existing Spec 028 functions (backward compat)
        ├─► Start() (ephemeral)
        ├─► Stop()  (ephemeral)
        ├─► Attach()
        ├─► Extract()
        └─► Status() ──► ContainerStatus
```

---

## State Machine

```text
                    create
  [No Workspace] ──────────► [Stopped]
                                │
                          start │ ▲ stop
                                ▼ │
                             [Running]
                                │
                          destroy│
                                ▼
                         [No Workspace]

  Ephemeral mode (Spec 028 backward compat):
  [No Container] ──start──► [Running] ──stop──► [No Container]
```

The persistent lifecycle (`create`/`start`/`stop`/
`destroy`) is distinct from the ephemeral lifecycle
(`start`/`stop`). The `status` command reports which
mode is active.
