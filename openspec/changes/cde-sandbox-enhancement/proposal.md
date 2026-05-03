## Why

`uf sandbox start` provides a seamless one-command
experience for ephemeral Podman containers: it auto-
detects cloud providers, starts the gateway proxy, pulls
the container image, injects credentials, waits for
health, and attaches. However, the gateway proxy is NOT
wired into persistent workspaces (`uf sandbox create`),
and the CDE backend (Eclipse Che) requires a full
Kubernetes cluster infrastructure that is impractical
for local single-developer use.

Research revealed three problems:

1. **Che requires Kubernetes**: Modern Eclipse Che (7.x)
   needs a K8s cluster + Che Operator + DevWorkspace
   Operator + Che Server — ~10GB RAM just for
   infrastructure before your workspace even starts.
   The lightweight single-container mode from Che 6
   was architecturally removed.

2. **kubedock cannot host Che**: kubedock emulates the
   Docker API for running containers INSIDE a Che
   workspace — it cannot host Che itself. The Che
   backend's assumption that kubedock bridges this gap
   is invalid.

3. **Gateway not wired into persistent paths**: Both
   `Create()` and persistent `Start()` skip
   `autoStartGateway()` entirely. The Vertex AI proxy
   that makes ephemeral sandbox so convenient is
   unavailable in persistent workspaces.

**DevPod** (loft-sh/devpod, 14.9k stars, MPL-2.0) is
a client-only tool that runs `devcontainer.json`-based
workspaces directly in Docker/Podman — no K8s required.
It is the direct spiritual successor to Che 6's
lightweight single-container mode and uses the
industry-standard devcontainer spec.

## What Changes

### Replace Che backend with DevPod backend

Remove `CheBackend` and all Che-specific code
(`che.go`, `BackendChe` constant, `resolveCheURL()`,
`addCheAuth()`, Che config fields). Replace with
`DevPodBackend` that delegates workspace lifecycle to
the `devpod` CLI. DevPod workspaces are created from
`devcontainer.json` and run in Podman via DevPod's
provider system.

### Gateway integration for persistent workspaces

Wire `autoStartGateway()` into the top-level `Create()`
and persistent `Start()` dispatch functions. Add
`GatewayPort` and `GatewayActive` fields to `Options`.
Update `buildPersistentRunArgs()` to inject gateway env
vars when active. This benefits both the DevPod backend
and the existing persistent Podman backend.

### Devcontainer scaffolding

Add `uf sandbox init` subcommand that generates a
`.devcontainer/devcontainer.json` for the project. Uses
the same container image, pre-configures gateway proxy
environment variables, and includes configurable demo
port forwarding. This is the industry-standard config
format consumed by DevPod, VS Code, GitHub Codespaces,
and others.

### CDE doctor checks updated for DevPod

Replace the Che-specific doctor check stubs with
DevPod checks: `devpod` binary presence, devcontainer
config existence. Context-sensitive — only shown when
DevPod is detected.

## Capabilities

### New Capabilities
- `DevPodBackend`: Workspace lifecycle via `devpod` CLI
  (Create, Start, Stop, Destroy, Status, Attach).
  Uses `devcontainer.json` and Podman provider.
- `uf sandbox init`: Scaffolds a
  `.devcontainer/devcontainer.json` with OpenCode image,
  gateway proxy env vars, and port forwarding.
- `BackendDevPod` constant and `--backend devpod` flag.
- DevPod doctor checks (binary presence, devcontainer
  config existence).

### Modified Capabilities
- `uf sandbox create`: Auto-starts the gateway proxy
  before delegating to any backend.
- `uf sandbox start` (persistent): Auto-starts gateway
  when resuming persistent workspaces.
- `buildPersistentRunArgs()`: Accepts gateway parameters
  for env var injection.
- `ResolveBackend()`: Replaces `BackendChe` with
  `BackendDevPod` in resolution logic.
- `autoDetectBackend()`: Prefers DevPod when `devpod`
  is in PATH; falls back to Podman.

### Removed Capabilities
- `CheBackend` and all Eclipse Che integration code.
- `BackendChe` constant.
- `CheURL` field on `Options`.
- `resolveCheURL()`, `addCheAuth()`,
  `cheWorkspaceName()` helpers.
- `cheWorkspaceInfo` struct and parse helpers.
- Che config fields in `SandboxConfig`.
- Che-specific doctor checks.

## Impact

### Files Affected

| Area | Changes |
|------|---------|
| `internal/sandbox/che.go` | DELETE entirely |
| `internal/sandbox/devpod.go` | NEW: DevPodBackend implementation |
| `internal/sandbox/backend.go` | Replace `BackendChe` with `BackendDevPod`, update `ResolveBackend()` and `autoDetectBackend()` |
| `internal/sandbox/sandbox.go` | Wire gateway into `Create()` and persistent `Start()`, add `GatewayPort`/`GatewayActive` to Options, remove `CheURL` |
| `internal/sandbox/podman.go` | Gateway params in `buildPersistentRunArgs()` |
| `internal/sandbox/config.go` | Remove Che config fields, add DevPod config |
| `internal/sandbox/workspace.go` | Remove Che references from `SandboxConfig` |
| `cmd/unbound-force/sandbox.go` | Add `init` subcommand, replace `--backend che` with `--backend devpod`, remove Che flags |
| `internal/doctor/checks.go` | Replace Che checks with DevPod checks |
| `internal/sandbox/sandbox_test.go` | Update tests for DevPod + gateway wiring |
| `internal/sandbox/devpod_test.go` | NEW: DevPodBackend tests |
| `cmd/unbound-force/sandbox_test.go` | Update for `init` subcommand + DevPod |
| `internal/doctor/doctor_test.go` | Update for DevPod checks |

### External Dependencies
- `devpod` binary (MPL-2.0) — used as external CLI
  tool via subprocess, no license impact on Apache-2.0
- No new Go module dependencies

## Constitution Alignment

### I. Autonomous Collaboration

**Assessment**: PASS

DevPod is an external CLI tool invoked via subprocess.
The gateway proxy runs as an independent host process.
DevPod workspaces communicate with the gateway through
`ANTHROPIC_BASE_URL` env vars. The devcontainer config
is a self-describing, version-controlled artifact.

### II. Composability First

**Assessment**: PASS

DevPod is optional — the Podman backend continues to
work independently. DevPod uses the devcontainer
standard (also consumed by VS Code, Codespaces).
Gateway integration activates only when a cloud
provider is detected. Users who never use DevPod see
zero changes except the gateway fix for persistent
Podman workspaces.

### III. Observable Quality

**Assessment**: PASS

Doctor checks produce structured output. DevPod
workspace status is observable via `uf sandbox status`.
Gateway status via `uf gateway status`. Devcontainer
config includes version markers.

### IV. Testability

**Assessment**: PASS

DevPodBackend uses the established `ExecCmd`/`LookPath`
injection pattern. All operations testable without a
running DevPod instance. Gateway wiring tests use the
same mock injection as existing tests.
<!-- scaffolded by uf vdev -->
