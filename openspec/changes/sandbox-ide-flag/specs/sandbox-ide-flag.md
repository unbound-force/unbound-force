## ADDED Requirements

### Requirement: IDE flag on sandbox create

`uf sandbox create` MUST accept an `--ide` flag that
specifies which IDE DevPod opens after provisioning
the workspace. The flag MUST be passed through to
`devpod up` as the `--ide` argument. Default: `"none"`.

#### Scenario: Create with VS Code

- **GIVEN** the user runs `uf sandbox create --backend devpod --ide vscode`
- **WHEN** the workspace is provisioned
- **THEN** `devpod up` is called with `--ide vscode`

Note: VS Code opening and the OpenCode server on port
4096 are expected DevPod behaviors, verified manually.
Unit tests verify only the argument passthrough.

#### Scenario: Create with default (no IDE flag)

- **GIVEN** the user runs `uf sandbox create --backend devpod`
- **AND** no `--ide` flag is provided
- **WHEN** the workspace is provisioned
- **THEN** `devpod up` is called with `--ide none`

#### Scenario: Create with invalid IDE value

- **GIVEN** the user runs `uf sandbox create --ide invalid`
- **WHEN** the command validates the IDE value
- **THEN** the command reports an error listing valid
  IDE values
- **AND** `devpod up` is not called

### Requirement: IDE flag on sandbox start

`uf sandbox start` MUST accept an `--ide` flag when
resuming a DevPod workspace. The flag MUST be passed
through to `devpod up --id <name> --ide <value>`.

#### Scenario: Resume with VS Code

- **GIVEN** a DevPod workspace exists
- **AND** the user runs `uf sandbox start --ide vscode`
- **WHEN** the workspace is resumed
- **THEN** `devpod up` is called with `--ide vscode`

#### Scenario: Resume with default

- **GIVEN** a DevPod workspace exists
- **AND** the user runs `uf sandbox start`
- **WHEN** the workspace is resumed
- **THEN** `devpod up` is called with `--ide none`

### Requirement: IDE value validation

The IDE value MUST be validated against DevPod's
supported IDE list: `none`, `vscode`, `openvscode`,
`fleet`, `jupyternotebook`, `cursor`. Invalid values
MUST produce an error before invoking `devpod up`.

#### Scenario: Valid IDE values accepted

- **GIVEN** the IDE value is one of: none, vscode,
  openvscode, fleet, jupyternotebook, cursor
- **WHEN** validation runs
- **THEN** the value is accepted

#### Scenario: Invalid IDE value rejected

- **GIVEN** the IDE value is "sublime"
- **WHEN** validation runs
- **THEN** an error is returned containing all valid
  IDE names (none, vscode, openvscode, fleet,
  jupyternotebook, cursor)

### Requirement: IDE resolution chain

The IDE value MUST be resolved from multiple sources
in priority order: `--ide` CLI flag > `UF_SANDBOX_IDE`
environment variable > `.uf/sandbox.yaml` `ide` field
> default `"none"`.

#### Scenario: CLI flag overrides env var

- **GIVEN** `UF_SANDBOX_IDE=fleet`
- **AND** the user passes `--ide vscode`
- **WHEN** the IDE value is resolved
- **THEN** the resolved value is `vscode`

#### Scenario: Env var used when no flag

- **GIVEN** `UF_SANDBOX_IDE=cursor`
- **AND** no `--ide` flag is provided
- **WHEN** the IDE value is resolved
- **THEN** the resolved value is `cursor`

#### Scenario: Config file used as fallback

- **GIVEN** `.uf/sandbox.yaml` contains `ide: vscode`
- **AND** no flag or env var is set
- **WHEN** the IDE value is resolved
- **THEN** the resolved value is `vscode`

### Requirement: IDE flag ignored for ephemeral sandbox

The `--ide` flag MUST only affect the DevPod backend.
When the sandbox runs in ephemeral Podman mode (no
`uf sandbox create`), the IDE flag MUST be silently
ignored.

#### Scenario: Ephemeral mode ignores IDE flag

- **GIVEN** no persistent workspace exists
- **AND** the user runs `uf sandbox start --ide vscode`
- **WHEN** the sandbox starts in ephemeral mode
- **THEN** the `--ide` flag is ignored
- **AND** the ephemeral Podman container starts normally

### Requirement: Attach detects persistent workspaces

`Attach()` MUST check for persistent workspaces (Podman
named volume or DevPod workspace) before falling back to
the ephemeral container check. When a persistent
workspace exists, `Attach()` MUST delegate to the
resolved backend's `Attach()` method.

#### Scenario: Attach to DevPod workspace

- **GIVEN** a DevPod workspace exists and is running
- **AND** the user runs `uf sandbox attach`
- **WHEN** Attach checks for workspaces
- **THEN** it detects the DevPod workspace
- **AND** delegates to `DevPodBackend.Attach()`

#### Scenario: Attach with no workspace

- **GIVEN** no persistent workspace exists
- **AND** no ephemeral container is running
- **WHEN** the user runs `uf sandbox attach`
- **THEN** the command reports "no sandbox running"

### Requirement: DevPod Start waits for health

`DevPodBackend.Start()` MUST wait for the OpenCode
server health check (`waitForHealth`) after `devpod up`
returns and before attempting TUI attach. If the health
check times out, the command MUST print a warning and
return without error (the IDE may still be connected).

#### Scenario: Health check passes

- **GIVEN** the DevPod workspace is resumed
- **AND** the OpenCode server starts within the timeout
- **WHEN** `DevPodBackend.Start()` runs
- **THEN** `waitForHealth` succeeds
- **AND** the TUI attaches normally

#### Scenario: Health check times out

- **GIVEN** the DevPod workspace is resumed
- **AND** the OpenCode server does not respond
- **WHEN** `DevPodBackend.Start()` runs
- **THEN** a warning is printed suggesting
  `uf sandbox attach`
- **AND** the command returns without error

### Requirement: Devcontainer auto-starts OpenCode server

The devcontainer.json template MUST include a
`postStartCommand` that starts the OpenCode server
in the background. This ensures the server is running
when DevPod overrides the container entrypoint with
its own agent process.

#### Scenario: Server starts via postStartCommand

- **GIVEN** a DevPod workspace starts from the template
- **WHEN** DevPod runs the postStartCommand
- **THEN** the OpenCode server starts on port 4096
- **AND** the health check endpoint responds

### Requirement: Destroy handles ephemeral mode

`Destroy()` MUST check for persistent workspaces before
delegating to `ResolveBackend()`. When no persistent
workspace exists, `Destroy()` MUST handle ephemeral
cleanup directly (stop and remove the container) or
report that there is nothing to destroy. This prevents
`ResolveBackend()` from incorrectly selecting the
DevPod backend for ephemeral containers.

#### Scenario: Destroy ephemeral container

- **GIVEN** an ephemeral container is running
- **AND** no persistent workspace exists
- **WHEN** the user runs `uf sandbox destroy`
- **THEN** the container is stopped and removed
- **AND** `ResolveBackend()` is NOT called

#### Scenario: Destroy with no workspace

- **GIVEN** no persistent workspace exists
- **AND** no ephemeral container is running
- **WHEN** the user runs `uf sandbox destroy`
- **THEN** the command reports "No sandbox to destroy"
- **AND** returns without error

#### Scenario: Destroy persistent workspace

- **GIVEN** a persistent DevPod workspace exists
- **WHEN** the user runs `uf sandbox destroy`
- **THEN** `Destroy()` delegates to the resolved
  backend's `Destroy()` method

## MODIFIED Requirements

### Requirement: Options struct

Previously: No IDE field on `Options`.
Now: `Options` MUST include an `IDE string` field.

### Requirement: DevPodBackend.Create arguments

Previously: `--ide none` hardcoded in `devpod up` args.
Now: Uses `opts.IDE` (resolved, validated, defaulting
to `"none"`).

### Requirement: DevPodBackend.Start arguments

Previously: No `--ide` argument on resume.
Now: Passes `--ide opts.IDE` to `devpod up --id`.
Additionally waits for OpenCode server health before
TUI attach.

### Requirement: Attach dispatch

Previously: Only checked ephemeral Podman container.
Now: Checks `isPersistentWorkspace()` first and
delegates to backend for persistent workspaces.

### Requirement: Destroy dispatch

Previously: Always called `ResolveBackend()` which
could incorrectly select DevPod for ephemeral
containers.
Now: Checks `isPersistentWorkspace()` first. Handles
ephemeral cleanup directly when no persistent
workspace exists.

## REMOVED Requirements

None.
