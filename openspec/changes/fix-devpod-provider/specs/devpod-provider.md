## ADDED Requirements

### Requirement: DevPod Create succeeds without podman in PATH

`DevPodBackend.Create()` MUST NOT require `podman` to be
present in PATH. The docker provider aliased as `podman`
resolves the container runtime binary via `DOCKER_PATH`
in its own provider configuration, not via PATH lookup.

#### Scenario: Create workspace when podman is absent from PATH

- **GIVEN** DevPod >= 0.5.0 is installed
- **AND** `LookPath("podman")` returns an error
- **AND** `.devcontainer/devcontainer.json` exists
- **AND** the docker provider is registered as `podman`
  by `uf setup`
- **WHEN** `DevPodBackend.Create()` is called
- **THEN** the call MUST proceed to `devpod up` without error
- **AND** the `devpod up` arguments MUST include
  `--provider podman`

## MODIFIED Requirements

### Requirement: DevPod Create provider flag

`DevPodBackend.Create()` MUST pass `--provider podman` to
`devpod up`. This references the provider **name** registered
by `uf setup` (`devpod provider add docker --name podman`),
not the removed standalone `loft-sh/devpod-provider-podman`.

The `--provider podman` flag at line 97 is already correct
and MUST NOT be changed. See design decision D2.

#### Scenario: Successful workspace creation

- **GIVEN** DevPod >= 0.5.0 is installed
- **AND** `.devcontainer/devcontainer.json` exists
- **WHEN** `DevPodBackend.Create()` is called with valid
  options
- **THEN** `devpod up` MUST be invoked with `--provider podman`
- **AND** the workspace MUST be created with the expected
  name `uf-sandbox-<project-name>`

### Requirement: DevPod Create GoDoc accuracy

The `Create()` and `devpodWorkspaceName()` GoDoc comments
MUST accurately reflect the docker-aliased-as-podman
provider model.

Previously: GoDoc referenced "Podman persistent workspace
naming convention (D5)", "podman in PATH (DevPod Podman
provider requirement)", and "--provider podman" without
explaining the aliasing model.

#### Scenario: GoDoc reflects provider model

- **GIVEN** the source file `internal/sandbox/devpod.go`
- **WHEN** a developer reads the GoDoc for `Create()` and
  `devpodWorkspaceName()`
- **THEN** the GoDoc MUST explain the docker provider
  aliased as `podman` model
- **AND** the `podman in PATH` pre-flight item MUST be
  removed from the pre-flight list
- **AND** no references to the removed standalone Podman
  provider MUST remain in the affected doc comments

### Requirement: Test assertions match production behavior

`TestDevPodCreate_Success` MUST assert `--provider podman`
(this is already correct and must remain). It MUST also
verify success when `podman` is absent from PATH.

Both `TestDevPodCreate_NotInstalled` (line 3689) and
`TestDevPodCreate_PodmanNotInstalled` (line 3992) MUST be
replaced with a single `TestDevPodCreate_NoPodmanInPath`
verifying `Create()` succeeds when podman is absent from
PATH. See design decision D3.

Previously: `TestDevPodCreate_Success` asserted
`--provider podman` (correct, unchanged).
`TestDevPodCreate_NotInstalled` and
`TestDevPodCreate_PodmanNotInstalled` both asserted that
missing podman caused an error.

#### Scenario: TestDevPodCreate_Success asserts provider name

- **GIVEN** the test `TestDevPodCreate_Success`
- **AND** `LookPath("podman")` returns error (podman absent)
- **WHEN** the test captures the `devpod up` arguments
- **THEN** the captured arguments MUST contain
  `--provider podman`
- **AND** `Create()` MUST succeed without error

#### Scenario: TestDevPodCreate_NoPodmanInPath succeeds

- **GIVEN** a test where `LookPath("podman")` returns error
- **AND** all other preconditions are satisfied
- **WHEN** `DevPodBackend.Create()` is called
- **THEN** the call MUST succeed without error

#### Scenario: Create workspace when provider not configured

- **GIVEN** DevPod >= 0.5.0 is installed
- **AND** no provider named `podman` is registered
- **WHEN** `DevPodBackend.Create()` is called
- **THEN** `devpod up` MUST return an error
- **AND** the error MUST be surfaced verbatim to the user

### Requirement: Actionable error on devpod up failure

When `devpod up` fails, the error message MUST include a
diagnostic hint directing the user to `uf doctor` and
`uf setup`.

#### Scenario: devpod up fails with diagnostic hint

- **GIVEN** `devpod up` returns a non-zero exit code
- **AND** the workspace is not running
- **WHEN** the error is surfaced to the user
- **THEN** the error MUST include the original `devpod up`
  output
- **AND** the error MUST include "Run 'uf doctor' to
  diagnose or 'uf setup' to configure"

## REMOVED Requirements

### Requirement: podman in PATH pre-flight check

The `LookPath("podman")` guard in `DevPodBackend.Create()`
is removed. The docker provider aliased as `podman` does
not require `podman` to be present in PATH — it resolves
the container runtime via `DOCKER_PATH` in its provider
configuration.

Previously: `Create()` returned an error containing
"podman not found" when `LookPath("podman")` failed.

Reason: The pre-flight check enforced a now-incorrect
prerequisite. `uf setup` configures
`devpod provider add docker --name podman -o DOCKER_PATH=podman`,
which wires the container runtime within DevPod's provider
config. PATH lookup for `podman` is neither necessary nor
sufficient for the provider to work.
