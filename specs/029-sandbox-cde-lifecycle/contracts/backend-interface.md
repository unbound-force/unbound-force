# Contract: Backend Interface

**Package**: `internal/sandbox`
**Date**: 2026-04-13

## Backend Interface

```go
// Backend defines the interface for workspace lifecycle
// operations. Implementations handle the infrastructure-
// specific details of provisioning, starting, stopping,
// and destroying workspaces.
type Backend interface {
    Create(opts Options) error
    Start(opts Options) error
    Stop(opts Options) error
    Destroy(opts Options) error
    Status(opts Options) (WorkspaceStatus, error)
    Attach(opts Options) error
    Name() string
}
```

---

## ResolveBackend

```go
// ResolveBackend selects the appropriate Backend
// implementation based on Options, environment, and
// configuration.
//
// Resolution order:
// 1. --backend flag (explicit selection)
// 2. UF_SANDBOX_BACKEND env var
// 3. .uf/sandbox.yaml backend field
// 4. Auto-detect: CDE if chectl/UF_CHE_URL available,
//    Podman otherwise
func ResolveBackend(opts Options) (Backend, error)
```

**Resolution matrix**:

| `--backend` | `UF_CHE_URL` | `chectl` in PATH | Result |
|-------------|-------------|-----------------|--------|
| `che` | set | any | CheBackend |
| `che` | unset | yes | CheBackend (discovers URL via chectl) |
| `che` | unset | no | Error: "CDE not configured" |
| `podman` | any | any | PodmanBackend |
| `auto` / empty | set | any | CheBackend |
| `auto` / empty | unset | yes | CheBackend |
| `auto` / empty | unset | no | PodmanBackend |

**Error behaviors**:

| Condition | Error message |
|-----------|--------------|
| `--backend che` but no CDE configured | "CDE backend requested but not configured. Set UF_CHE_URL or install chectl." |
| `--backend podman` but podman missing | "podman not found. Install: brew install podman" |
| Unknown backend name | "unknown backend: <name>. Use 'auto', 'podman', or 'che'." |

---

## PodmanBackend

### Create

```go
// Create provisions a persistent Podman workspace with
// named volumes. Seeds the workspace with the project's
// source code.
//
// Steps:
// 1. Check podman is in PATH
// 2. Verify no workspace exists (named volume check)
// 3. Create named volume: uf-sandbox-<project-name>
// 4. Start container with named volume mount
// 5. Copy project source into workspace
// 6. Wait for health check
//
// Returns error if workspace already exists.
func (b *PodmanBackend) Create(opts Options) error
```

**Preconditions**:
- `podman` in PATH
- No existing named volume `uf-sandbox-<project-name>`

**Postconditions**:
- Named volume `uf-sandbox-<project-name>` exists
- Container `uf-sandbox-<project-name>` is running
- Project source code is in `/workspace`
- OpenCode server is healthy on port 4096

**Partial failure cleanup**: If `Create()` fails after
step 3 (volume created) but before step 6 (health check
passes), the implementation SHOULD clean up the partial
state (remove the container and volume) so the engineer
can retry `create` without needing `destroy` first.

**Error behaviors**:

| Condition | Error message |
|-----------|--------------|
| Podman missing | "podman not found. Install: brew install podman" |
| Workspace exists | "sandbox already exists for <project>. Use `uf sandbox start` or `uf sandbox destroy` first." |
| Volume create fails | "failed to create volume: <error>" |
| Container start fails | "failed to start container: <error>" (partial state cleaned up) |
| Health timeout | "health check timed out after 60s" (partial state cleaned up) |

### Start

```go
// Start resumes a stopped persistent workspace.
// If no persistent workspace exists, falls back to
// ephemeral mode (Spec 028 behavior).
func (b *PodmanBackend) Start(opts Options) error
```

**Preconditions**:
- Named volume exists (persistent mode) OR no prior
  create (ephemeral mode)

**Postconditions**:
- Container is running
- OpenCode server is healthy

### Stop

```go
// Stop stops a running workspace. In persistent mode,
// the container is stopped but the named volume is
// preserved. In ephemeral mode, the container is
// stopped and removed (Spec 028 behavior).
func (b *PodmanBackend) Stop(opts Options) error
```

**Postconditions**:
- Persistent mode: container stopped, volume preserved
- Ephemeral mode: container removed

### Destroy

```go
// Destroy permanently deletes the workspace, container,
// and named volume. Idempotent.
func (b *PodmanBackend) Destroy(opts Options) error
```

**Postconditions**:
- No container `uf-sandbox-<project-name>` exists
- No volume `uf-sandbox-<project-name>` exists

---

## CheBackend

### Create

```go
// Create provisions a CDE workspace via Eclipse Che /
// Dev Spaces from the project's devfile.
//
// Steps:
// 1. Verify CDE access (chectl or REST API)
// 2. Locate devfile.yaml in project directory
// 3. Create workspace via chectl or REST API
// 4. Wait for workspace to reach RUNNING state
//
// Returns error if devfile is missing or CDE is
// unreachable.
func (b *CheBackend) Create(opts Options) error
```

**Preconditions**:
- CDE configured (chectl in PATH or UF_CHE_URL set)
- `devfile.yaml` exists in project directory

**Postconditions**:
- Che workspace `uf-<project-name>` exists and is running
- Endpoints are accessible

**Error behaviors**:

| Condition | Error message |
|-----------|--------------|
| No devfile | "devfile.yaml not found in <dir>. Add a devfile or use --backend podman." |
| CDE unreachable | "cannot reach Che at <url>. Check UF_CHE_URL and authentication." |
| Workspace exists | "CDE workspace already exists for <project>. Use `uf sandbox start` or `uf sandbox destroy`." |

### Start

```go
// Start starts a stopped CDE workspace.
func (b *CheBackend) Start(opts Options) error
```

### Stop

```go
// Stop stops a running CDE workspace. Workspace state
// is preserved by the CDE platform.
func (b *CheBackend) Stop(opts Options) error
```

### Destroy

```go
// Destroy permanently deletes the CDE workspace.
func (b *CheBackend) Destroy(opts Options) error
```

### Attach

```go
// Attach connects the TUI to the CDE workspace's
// OpenCode server via the Che endpoint URL.
func (b *CheBackend) Attach(opts Options) error
```

**Key difference from Podman**: The server URL is the
Che endpoint URL (e.g.,
`https://uf-myproject-opencode.apps.che.example.com`),
not `http://localhost:4096`.
