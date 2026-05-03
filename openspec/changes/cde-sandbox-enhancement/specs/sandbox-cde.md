## ADDED Requirements

### Requirement: DevPodBackend implementation

A `DevPodBackend` struct MUST implement the `Backend`
interface by delegating to the `devpod` CLI. The backend
MUST use `devcontainer.json` as its workspace
configuration format.

Methods:
- `Create(opts)`: Run `devpod up` with project dir,
  Podman provider, gateway env vars, and workspace name
- `Start(opts)`: Run `devpod up` to resume a stopped
  workspace
- `Stop(opts)`: Run `devpod stop` with workspace name
- `Destroy(opts)`: Run `devpod delete` with workspace
  name and `--force` flag
- `Status(opts)`: Run `devpod status` and parse output
- `Attach(opts)`: Run `opencode attach` against the
  workspace's OpenCode server
- `Name()`: Return `"devpod"`

`DevPodBackend` MUST enforce a minimum DevPod version
of 0.5.0 via `parseDevPodVersion()`, following the
`parsePodmanVersion()` pattern.

`DevPodBackend.Create()` MUST pre-flight check for
`podman` in PATH before calling `devpod up`.

#### Scenario: Create workspace with DevPod

- **GIVEN** `devpod` >= 0.5.0 is in PATH and
  `.devcontainer/devcontainer.json` exists
- **WHEN** the user runs
  `uf sandbox create --backend devpod`
- **THEN** `devpod up` MUST be called with the project
  directory, `--provider podman`, gateway env vars
  via `--workspace-env` (when gateway is active), and
  workspace name `uf-sandbox-<project-name>`

#### Scenario: DevPod version too old

- **GIVEN** `devpod` is in PATH with version 0.4.x
- **WHEN** `DevPodBackend.Create()` is called
- **THEN** it MUST return an error with the current
  and required versions

#### Scenario: Podman not installed for DevPod

- **GIVEN** `devpod` is in PATH but `podman` is not
- **WHEN** `DevPodBackend.Create()` is called
- **THEN** it MUST return an error with Podman install
  instructions

#### Scenario: Create fails mid-execution

- **GIVEN** `devpod up` returns a non-zero exit code
- **WHEN** `DevPodBackend.Create()` runs
- **THEN** it MUST propagate the error with the devpod
  output for diagnostics

#### Scenario: Missing devcontainer.json

- **GIVEN** `.devcontainer/devcontainer.json` does NOT
  exist
- **WHEN** `DevPodBackend.Create()` is called
- **THEN** it MUST return an error suggesting
  `uf sandbox init`

#### Scenario: DevPod not installed

- **GIVEN** `devpod` is not in PATH
- **WHEN** `ResolveBackend()` is called with
  `--backend devpod`
- **THEN** it MUST return an error with the message
  containing install instructions

#### Scenario: Stop and resume workspace

- **GIVEN** a DevPod workspace exists and is running
- **WHEN** the user runs `uf sandbox stop` then
  `uf sandbox start`
- **THEN** `devpod stop` MUST be called, then
  `devpod up` MUST resume the existing workspace
  without recreating it

#### Scenario: Attach to DevPod workspace

- **GIVEN** a DevPod workspace is running
- **WHEN** `Attach()` is called
- **THEN** `opencode attach` MUST be called with the
  workspace's forwarded server URL (default:
  `http://localhost:4096`)

#### Scenario: Parse DevPod status output

- **GIVEN** `devpod status` returns JSON with
  `"state": "Running"`
- **WHEN** `Status()` parses the output
- **THEN** `WorkspaceStatus.Running` MUST be true,
  `WorkspaceStatus.Name` MUST be
  `"uf-sandbox-<project>"`,
  `WorkspaceStatus.Backend` MUST be `"devpod"`

### Requirement: DevPod persistent workspace detection

`isPersistentWorkspace()` MUST be extended to detect
DevPod workspaces. When `devpod` is in PATH, it MUST
call `devpod status uf-sandbox-<project> --output json`
to check for an existing DevPod workspace. If the
status command returns a workspace (any state), the
workspace is considered persistent.

The DevPod check is guarded by `LookPath("devpod")` —
if DevPod is not installed, only the Podman volume
check applies.

#### Scenario: DevPod workspace detected as persistent

- **GIVEN** a DevPod workspace named
  `uf-sandbox-<project>` exists
- **WHEN** `isPersistentWorkspace()` is called
- **THEN** it MUST return true

#### Scenario: No DevPod workspace

- **GIVEN** `devpod` is in PATH but no workspace named
  `uf-sandbox-<project>` exists
- **WHEN** `isPersistentWorkspace()` is called
- **THEN** it MUST return the result of the Podman
  volume check (existing behavior)

### Requirement: Gateway integration for Create

The top-level `Create()` dispatch function MUST call
`autoStartGateway()` before delegating to the backend.
The gateway port and active flag MUST be stored on
`Options.GatewayPort` and `Options.GatewayActive`.

#### Scenario: Create with Vertex AI

- **GIVEN** `CLAUDE_CODE_USE_VERTEX=1` and
  `ANTHROPIC_VERTEX_PROJECT_ID` are set
- **WHEN** `uf sandbox create` is called
- **THEN** the gateway MUST be started before workspace
  creation, and gateway env vars MUST be passed to the
  backend

#### Scenario: Create without cloud provider

- **GIVEN** no cloud provider env vars are set
- **WHEN** `uf sandbox create` is called
- **THEN** the gateway MUST NOT be started, workspace
  creation MUST proceed normally

### Requirement: Gateway integration for persistent Start

The persistent workspace branch of `Start()` MUST call
`autoStartGateway()` before delegating to the backend's
`Start()` method.

#### Scenario: Resume persistent workspace with gateway

- **GIVEN** a persistent workspace exists and Vertex AI
  env vars are set
- **WHEN** `uf sandbox start` is called
- **THEN** the gateway MUST be auto-started before
  resuming the workspace

### Requirement: Options struct gateway fields

`Options` MUST include `GatewayPort int` and
`GatewayActive bool` fields. These MUST be set by
dispatch functions after `autoStartGateway()` returns.

### Requirement: Persistent Podman gateway env injection

`buildPersistentRunArgs()` MUST read
`opts.GatewayActive` and `opts.GatewayPort` and inject
gateway env vars when active.

Previously: `buildPersistentRunArgs()` called
`forwardedEnvVars(opts, false)` unconditionally.

#### Scenario: Persistent Podman with gateway

- **GIVEN** a persistent Podman workspace is being
  created and `opts.GatewayActive` is true
- **WHEN** `buildPersistentRunArgs()` assembles args
- **THEN** the args MUST include
  `-e ANTHROPIC_BASE_URL=http://host.containers.internal:53147`
  and MUST NOT include provider-specific keys

### Requirement: Devcontainer scaffolding command

`uf sandbox init` MUST generate
`.devcontainer/devcontainer.json` in the project root.
The command MUST accept `--image` (default from config),
`--demo-ports` (int slice), and `--force` (overwrite)
flags.

The generated devcontainer MUST include:
- `image` set to the container image
- `forwardPorts` with port 4096 plus demo ports
- `containerEnv` with `ANTHROPIC_BASE_URL` and
  `ANTHROPIC_API_KEY=gateway`
- `remoteUser` set to `dev`

Idempotent: skip if `.devcontainer/devcontainer.json`
exists unless `--force` is set.

#### Scenario: Scaffold devcontainer

- **GIVEN** no `.devcontainer/` directory exists
- **WHEN** `uf sandbox init` is called
- **THEN** `.devcontainer/devcontainer.json` MUST be
  created with image, ports, and gateway env vars

#### Scenario: Already exists

- **GIVEN** `.devcontainer/devcontainer.json` exists
- **WHEN** `uf sandbox init` is called without `--force`
- **THEN** the command MUST skip and print a message

#### Scenario: Force overwrite

- **GIVEN** `.devcontainer/devcontainer.json` exists
- **WHEN** `uf sandbox init --force` is called
- **THEN** the existing file MUST be overwritten

### Requirement: Extract behavior for persistent workspaces

`Extract()` MUST replace the Che-specific
`resolveCheURL()` check with a general persistent
workspace check. For all persistent workspaces (both
Podman and DevPod), `Extract()` MUST return early with
the message: "This is a persistent workspace — changes
are on the host filesystem or use git push."

Previously: checked `resolveCheURL(opts)` to detect
CDE workspaces specifically.

#### Scenario: Extract on persistent DevPod workspace

- **GIVEN** a DevPod persistent workspace exists
- **WHEN** `Extract()` is called
- **THEN** it MUST return early with a message
  indicating changes are accessible via git

### Requirement: DevPod doctor checks

`uf doctor` MUST include a "DevPod" check group when
DevPod is detected (`devpod` in PATH or config
`sandbox.backend == "devpod"`).

Checks:
1. `devpod` binary presence (install hint: DevPod
   install URL)
2. `.devcontainer/devcontainer.json` existence (install
   hint: `Run: uf sandbox init`)

#### Scenario: Checks hidden when not relevant

- **GIVEN** `devpod` is not in PATH and config backend
  is not `"devpod"`
- **WHEN** `uf doctor` is called
- **THEN** the DevPod check group MUST NOT appear

### Requirement: ResolveBackend DevPod support

`ResolveBackend()` MUST accept `"devpod"` as a valid
backend name. When `backendName == "devpod"`, it MUST
check for `devpod` in PATH and return a `DevPodBackend`.

`autoDetectBackend()` MUST prefer DevPod when BOTH
`devpod` is in PATH AND
`.devcontainer/devcontainer.json` exists in the project
directory. If either condition is false, fall back to
Podman.

Previously: preferred CDE when `chectl` or `UF_CHE_URL`
was available.

#### Scenario: Auto-detect selects DevPod

- **GIVEN** `devpod` is in PATH and
  `.devcontainer/devcontainer.json` exists
- **WHEN** `autoDetectBackend()` is called
- **THEN** it MUST return a `DevPodBackend`

#### Scenario: Auto-detect falls back to Podman

- **GIVEN** `devpod` is in PATH but
  `.devcontainer/devcontainer.json` does NOT exist
- **WHEN** `autoDetectBackend()` is called
- **THEN** it MUST return a `PodmanBackend`

### Requirement: Che migration error

When `backendName == "che"`, `ResolveBackend()` MUST
return an error with a migration message:
`"che backend removed, use --backend devpod instead.
Install DevPod: https://devpod.sh/docs/getting-started/install"`

## MODIFIED Requirements

### Requirement: buildPersistentRunArgs gateway support

`buildPersistentRunArgs()` MUST call
`forwardedEnvVars(opts, opts.GatewayActive)` and
append `gatewayEnvVars(opts.GatewayPort)` when
`opts.GatewayActive` is true.

Previously: called `forwardedEnvVars(opts, false)`.

### Requirement: Backend constants

`BackendChe` constant MUST be removed. `BackendDevPod`
constant (`"devpod"`) MUST be added.

Previously: `BackendChe = "che"`.

### Requirement: Options struct cleanup

`CheURL` field MUST be removed from `Options`.
`HTTPDo` documentation MUST be updated to remove
"Used for CDE REST API calls."
`BackendName` documentation MUST be updated to list
`"auto", "podman", "devpod"`.

Previously: `CheURL string` existed for Che server URL.
`HTTPDo` documented as CDE-specific. `BackendName`
listed `"che"` as valid value.

### Requirement: WorkspaceStatus and SandboxConfig
comment updates

All struct comments referencing `"che"` or `"CDE"` in
`workspace.go` MUST be updated to reference `"devpod"`
/ `"DevPod"` where applicable. This includes:
- `WorkspaceStatus.Backend`: `"podman" or "devpod"`
- `WorkspaceStatus.Mode`: remove CDE reference
- `WorkspaceStatus.ProjectDir`: remove CDE reference
- `WorkspaceStatus.ServerURL`: remove Che endpoint ref
- `SandboxConfig.Backend`: `"auto", "podman", "devpod"`

Previously: all referenced "che" and "CDE".

### Requirement: CLI help text updates

All user-facing help text in
`cmd/unbound-force/sandbox.go` MUST be updated to
replace Che/CDE references with DevPod. This includes
the parent `sandbox` command's Long description, the
`create` command's Long description, and the `destroy`
command's help text.

Previously: referenced "Eclipse Che / Dev Spaces".

## REMOVED Requirements

### Requirement: CheBackend

All Eclipse Che integration code MUST be removed:
`che.go` file, `BackendChe` constant, `CheURL` on
Options, `resolveCheURL()`, `addCheAuth()`,
`cheWorkspaceName()`, `cheWorkspaceInfo` struct,
`parseCheWorkspaceList()`, `parseCheWorkspaceJSON()`,
`extractCheServerURL()`, `extractCheEndpoints()`.

### Requirement: Che config fields

`SandboxConfig.Che` struct (URL, Token fields) MUST be
removed from workspace config parsing.
<!-- scaffolded by uf vdev -->
