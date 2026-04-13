# Research: Sandbox CDE Lifecycle

**Branch**: `029-sandbox-cde-lifecycle` | **Date**: 2026-04-13
**Spec**: [spec.md](spec.md)

## R1: Backend Interface Pattern — Strategy vs. Direct Dispatch

**Decision**: Introduce a `Backend` interface in
`internal/sandbox/` with two implementations:
`PodmanBackend` and `CheBackend`. The existing
`Start()`/`Stop()` functions become methods on
`PodmanBackend`. New `Create()`/`Destroy()` functions
dispatch to the selected backend via the interface.

```go
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

**Rationale**: The Strategy pattern (SOLID Open/Closed
Principle) allows adding new backends (e.g., Docker,
Kubernetes direct) without modifying the orchestration
layer. The existing `Options` struct is extended with a
`Backend` field that selects the implementation. This
follows the established codebase pattern where all
external dependencies are injectable — the backend
itself becomes an injectable dependency.

**Alternatives considered**:
- Switch statement on backend name: Violates Open/Closed.
  Adding a new backend requires modifying the switch.
- Separate packages per backend (`sandbox/podman/`,
  `sandbox/che/`): Over-engineered for two backends.
  A single package with interface + two implementations
  is sufficient. YAGNI for package-level separation.
- Functional dispatch (map of functions): Loses the
  cohesion of grouping related operations. An interface
  is clearer.

**Source**: Dewey learning `028-sandbox-command-1`
(Options/ExecCmd pattern), Replicator system-design
skill (dependency injection via constructor/Options).

---

## R2: Eclipse Che Workspace Provisioning — chectl vs. REST API

**Decision**: Support both `chectl` CLI and Che REST API,
with `chectl` as the primary path and REST API as the
fallback. Detection order:
1. If `chectl` is in PATH → use CLI
2. If `UF_CHE_URL` is set → use REST API directly
3. Neither → CDE backend unavailable, fall back to Podman

**chectl workflow**:
```bash
# Create workspace from devfile
chectl workspace:create --devfile=devfile.yaml \
  --name=uf-<project-name>

# Start workspace
chectl workspace:start --name=uf-<project-name>

# Stop workspace
chectl workspace:stop --name=uf-<project-name>

# Delete workspace
chectl workspace:delete --name=uf-<project-name> --yes
```

**REST API workflow** (Che Server API v2):
```
POST /api/workspace/devfile
  Body: devfile YAML content
  → workspace ID

PATCH /api/workspace/{id}/runtime
  Body: {"status": "RUNNING"}

PATCH /api/workspace/{id}/runtime
  Body: {"status": "STOPPED"}

DELETE /api/workspace/{id}
```

**Rationale**: `chectl` handles authentication, endpoint
discovery, and error formatting. It's the recommended
tool for Che workspace management. The REST API fallback
supports environments where `chectl` is not installed
but the Che server is accessible (e.g., OpenShift Dev
Spaces with only a URL and token).

**Alternatives considered**:
- REST API only: Requires implementing authentication,
  token refresh, and error handling that `chectl` already
  provides. More work, more surface area for bugs.
- `chectl` only: Excludes environments where only the
  REST API is available (e.g., CI/CD pipelines, remote
  servers without `chectl`).

**Source**: Eclipse Che documentation, `chectl` CLI
reference, Che Server API v2 specification.

---

## R3: Podman Named Volumes for Persistent State

**Decision**: Replace the ephemeral bind-mount pattern
from Spec 028 with named volumes for the `create`
lifecycle. Named volumes persist across container
stop/start/rm cycles.

**Volume naming convention**: `uf-sandbox-<project-name>`
where `<project-name>` is derived from the project
directory basename (sanitized: lowercase, alphanumeric
+ hyphens only).

**Volume lifecycle**:
```bash
# Create: provision named volume + initial container
podman volume create uf-sandbox-myproject
podman run -d --name uf-sandbox-myproject \
  -v uf-sandbox-myproject:/workspace \
  -p 4096:4096 \
  <image>
# Copy project source into volume on first create
podman cp <project-dir>/. uf-sandbox-myproject:/workspace/

# Start: restart existing container
podman start uf-sandbox-myproject

# Stop: stop without removing
podman stop uf-sandbox-myproject

# Destroy: remove container + volume
podman rm uf-sandbox-myproject
podman volume rm uf-sandbox-myproject
```

**Key difference from Spec 028**: The existing `Start()`
creates an ephemeral container with a bind mount and
`Stop()` removes the container (destroying all state).
The new `Create()` provisions a named volume that
survives container removal. `Start()`/`Stop()` only
control the running state.

**Rationale**: Named volumes are the Podman/Docker
standard for persistent data. They survive container
lifecycle events. The volume name includes the project
name to support multiple concurrent sandboxes (one per
project), relaxing the single-container constraint from
Spec 028.

**Alternatives considered**:
- Bind mount to a host directory (e.g.,
  `~/.uf/sandbox/<project>/`): Works but creates
  permission issues on macOS (Podman VM boundary) and
  SELinux complications on Fedora. Named volumes are
  managed by Podman and avoid these issues.
- `podman commit` to save container state: Saves the
  entire filesystem as an image layer. Heavy, slow,
  and doesn't preserve running process state. Named
  volumes are lighter and more targeted.

**Source**: Podman documentation on named volumes,
Spec 028 research R4 (container lifecycle).

---

## R4: Container Naming — Per-Project vs. Global

**Decision**: Change the container name from the fixed
`uf-sandbox` (Spec 028) to `uf-sandbox-<project-name>`
to support multiple concurrent sandboxes.

**Project name derivation**:
```go
func projectName(dir string) string {
    name := filepath.Base(dir)
    // Sanitize: lowercase, replace non-alphanumeric with hyphens
    name = strings.ToLower(name)
    name = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(name, "-")
    name = strings.Trim(name, "-")
    if name == "" {
        name = "default"
    }
    return name
}
```

**Backward compatibility**: The existing `uf sandbox start`
(Spec 028 ephemeral mode) continues to use the fixed
`uf-sandbox` container name. The new `uf sandbox create`
uses `uf-sandbox-<project-name>`. The `status`, `attach`,
`stop`, and `extract` commands auto-detect which naming
scheme is in use by checking both names.

**Rationale**: Per the spec edge case: "Each project gets
its own named sandbox (`uf-sandbox-<project-name>`). The
single-container constraint from Spec 028 is relaxed to
single-container-per-project."

**Source**: Spec 029 Edge Cases section.

---

## R5: Bidirectional Git Sync

**Decision**: Git sync is the coordination mechanism
between the host and the CDE workspace. The agent
pushes from inside the workspace; the engineer pushes
from the host or Che IDE.

**Workspace-side setup** (during `create`):
1. Configure git remote to point to the same origin
2. Set up SSH/HTTPS credentials via Che secret injection
   or Podman env var forwarding
3. Create a workspace branch:
   `uf-sandbox-<project-name>-workspace`

**Sync protocol**:
- Agent pushes: Standard `git push` from inside workspace
- Engineer pulls: `git pull` on host to get agent changes
- Engineer pushes: `git push` from host; workspace does
  `git pull` on next `/unleash` run or manually
- Conflict detection: `git pull` with `--ff-only` inside
  workspace. If fast-forward fails, report conflict and
  require manual resolution.

**Rationale**: Git is the natural coordination mechanism.
Both the agent and engineer already use git. No custom
sync protocol needed. The workspace branch prevents
conflicts with the engineer's working branch.

**Alternatives considered**:
- `rsync` or `podman cp`: One-directional, loses git
  history, no conflict detection.
- Custom file watcher + sync daemon: Over-engineered.
  Git already solves this problem.
- Shared filesystem (NFS/CIFS): Requires infrastructure
  setup, doesn't work with CDE backends.

**Source**: Spec 029 US3 acceptance scenarios.

---

## R6: CDE Endpoint Exposure for Demos

**Decision**: CDE workspaces expose ports via the Che
endpoint mechanism. The devfile declares endpoints, and
Che creates routes/ingresses for them.

**Devfile endpoint declaration** (in project's
`devfile.yaml`):
```yaml
components:
  - name: dev
    container:
      endpoints:
        - name: opencode-server
          targetPort: 4096
          exposure: public
          protocol: https
        - name: demo-web
          targetPort: 3000
          exposure: public
          protocol: https
        - name: demo-api
          targetPort: 8080
          exposure: public
          protocol: https
```

**For Podman backend**: Port mapping via `-p` flags
(same as Spec 028 but extended for demo ports):
```bash
podman run -d \
  -p 4096:4096 \
  -p 3000:3000 \
  -p 8080:8080 \
  ...
```

**Rationale**: Che handles endpoint exposure natively via
Kubernetes ingress/routes. The devfile is the standard
mechanism. For Podman, explicit port mapping is the
equivalent. The `create` command reads the devfile to
discover which ports to map.

**Source**: Spec 029 FR-008, containerfile#3 dependency.

---

## R7: API Key Management — CDE vs. Podman

**Decision**: Two distinct credential injection strategies
based on backend:

**Podman backend** (unchanged from Spec 028):
- `-e VAR` syntax forwards host env vars
- Google Cloud credentials mounted as volumes
- OLLAMA_HOST set to `host.containers.internal:11434`

**CDE backend** (new):
- API keys injected via Che user preferences or
  Kubernetes secrets (per FR-014)
- No host env var forwarding (CDE runs on remote infra)
- OLLAMA_HOST configurable via `UF_OLLAMA_HOST` env var
  or `.uf/sandbox.yaml` config (per FR-015)

**Configuration file** (`.uf/sandbox.yaml`):
```yaml
# CDE backend configuration
che:
  url: https://che.example.com
  # Optional: token for REST API (if not using chectl)
  token: ${UF_CHE_TOKEN}

# Ollama endpoint override for CDE
ollama:
  host: http://ollama.internal:11434

# Default backend: auto, podman, or che
backend: auto

# Demo port mappings (Podman only; CDE uses devfile)
demo_ports:
  - 3000
  - 8080
```

**Security note**: The `token` field in `.uf/sandbox.yaml`
MUST NOT contain literal tokens. Use env var references
(`${UF_CHE_TOKEN}`) only. The implementation SHOULD
warn if a literal token is detected in the config file.
The `.uf/` directory is gitignored but defense-in-depth
requires avoiding plaintext credentials in any file.

**Rationale**: CDE environments run on remote
infrastructure where host env vars are not available.
Kubernetes secrets are the standard mechanism for
credential injection in cloud-native environments.
The config file provides a persistent, version-
controllable configuration.

**Source**: Spec 029 FR-014, FR-015, containerfile#4
dependency.

---

## R8: File Organization — Extended Package

**Decision**: Extend the existing `internal/sandbox/`
package with new files. Do not create a sub-package.

| File | Responsibility |
|------|---------------|
| `sandbox.go` | EXISTING: Core orchestration (Start, Stop, Attach, Extract, Status) — updated to dispatch via Backend |
| `detect.go` | EXISTING: Platform detection — unchanged |
| `config.go` | EXISTING: Container config — extended with named volume support |
| `backend.go` | NEW: Backend interface definition |
| `podman.go` | NEW: PodmanBackend implementation (Create, Destroy, persistent Start/Stop) |
| `che.go` | NEW: CheBackend implementation (Create, Destroy, Start, Stop via chectl/API) |
| `workspace.go` | NEW: Workspace state management (naming, config file parsing) |
| `sandbox_test.go` | EXISTING: Extended with new test cases |

**Rationale**: Keeping everything in one package
maintains the established pattern. The Backend interface
provides separation of concerns without package-level
boundaries. New files follow the existing naming
convention (lowercase, descriptive).

**Alternatives considered**:
- Sub-packages (`sandbox/podman/`, `sandbox/che/`):
  Over-engineered for two backends. Would require
  exporting types that are currently internal.
- Single file for both backends: Would exceed 500 lines.
  Separate files per backend is cleaner.

**Source**: Spec 028 research R8 (file organization),
Dewey learning `028-sandbox-command-1`.

---

## R9: Backward Compatibility with Spec 028

**Decision**: The existing `uf sandbox start` (ephemeral
mode) continues to work exactly as before. The new
`create`/`destroy` verbs are additive. The behavior
matrix:

| Command | Sandbox exists? | Behavior |
|---------|----------------|----------|
| `uf sandbox start` (no prior create) | No | Ephemeral mode (Spec 028 behavior) |
| `uf sandbox create` | No | Persistent mode (new) |
| `uf sandbox start` (after create) | Yes (stopped) | Resume persistent workspace |
| `uf sandbox stop` (ephemeral) | Running | Stop + remove container (Spec 028) |
| `uf sandbox stop` (persistent) | Running | Stop container, preserve volume |
| `uf sandbox destroy` | Exists | Remove container + volume |

**Detection logic**: Check for named volume
`uf-sandbox-<project-name>`. If it exists, the sandbox
is in persistent mode. If not, fall back to ephemeral
mode (Spec 028 behavior).

**Rationale**: Backward compatibility is required by
FR-018. Engineers who don't use `create`/`destroy` see
no behavior change. The detection logic is simple and
deterministic.

**Source**: Spec 029 FR-018, Spec 028 existing behavior.
