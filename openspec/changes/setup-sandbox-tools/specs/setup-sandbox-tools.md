## ADDED Requirements

### Requirement: Setup installs Podman

`uf setup` MUST install Podman when it is not present.
On macOS, setup MUST use `brew install podman`. On Linux,
setup MUST use `brew install podman` when Homebrew is
available. When Homebrew is not available, setup MUST
skip installation and report a download URL.

#### Scenario: Podman not installed, Homebrew available

- **GIVEN** Podman is not in PATH
- **AND** Homebrew is available
- **WHEN** `uf setup` runs
- **THEN** Podman is installed via `brew install podman`
- **AND** the step result reports "installed" with
  detail "via Homebrew"

#### Scenario: Podman already installed

- **GIVEN** Podman is in PATH
- **WHEN** `uf setup` runs
- **THEN** the step result reports "already installed"
- **AND** no install command is executed

#### Scenario: No Homebrew available

- **GIVEN** Podman is not in PATH
- **AND** Homebrew is not available
- **WHEN** `uf setup` runs
- **THEN** the step result reports "skipped"
- **AND** detail includes a download URL

### Requirement: Setup initializes Podman machine on macOS

On macOS, after Podman installation, `uf setup` MUST
check for an existing Podman machine. If no machine
exists, setup MUST run `podman machine init` and
`podman machine start`. The `podman machine init`
command MUST have a timeout of 180 seconds to prevent
indefinite hangs on slow networks. Machine init or
start failures MUST NOT block subsequent setup steps.

#### Scenario: macOS, no Podman machine exists

- **GIVEN** the platform is macOS
- **AND** Podman is installed
- **AND** no Podman machine exists
- **WHEN** `uf setup` runs the Podman step
- **THEN** `podman machine init` is executed
- **AND** `podman machine start` is executed

#### Scenario: macOS, Podman machine already exists

- **GIVEN** the platform is macOS
- **AND** Podman is installed
- **AND** a Podman machine already exists
- **WHEN** `uf setup` runs the Podman step
- **THEN** machine init is skipped

#### Scenario: Linux, no machine needed

- **GIVEN** the platform is Linux
- **AND** Podman is installed
- **WHEN** `uf setup` runs the Podman step
- **THEN** no machine init is attempted

#### Scenario: macOS, machine init fails

- **GIVEN** the platform is macOS
- **AND** Podman is installed
- **AND** no Podman machine exists
- **AND** `podman machine init` fails
- **WHEN** `uf setup` runs the Podman step
- **THEN** the step result reports "installed"
- **AND** detail includes "machine init failed"
- **AND** subsequent setup steps continue

#### Scenario: macOS, machine start fails

- **GIVEN** the platform is macOS
- **AND** Podman is installed
- **AND** `podman machine init` succeeds
- **AND** `podman machine start` fails
- **WHEN** `uf setup` runs the Podman step
- **THEN** the step result reports "installed"
- **AND** detail includes "machine start failed"
- **AND** subsequent setup steps continue

### Requirement: Setup installs DevPod

`uf setup` MUST install DevPod when it is not present.
Setup MUST use `brew install devpod` when Homebrew is
available. When Homebrew is not available, setup MUST
skip installation and report the DevPod download URL.

#### Scenario: DevPod not installed, Homebrew available

- **GIVEN** DevPod is not in PATH
- **AND** Homebrew is available
- **WHEN** `uf setup` runs
- **THEN** DevPod is installed via `brew install devpod`
- **AND** the step result reports "installed"

#### Scenario: DevPod already installed

- **GIVEN** DevPod is in PATH
- **WHEN** `uf setup` runs
- **THEN** the step result reports "already installed"

#### Scenario: No Homebrew available for DevPod

- **GIVEN** DevPod is not in PATH
- **AND** Homebrew is not available
- **WHEN** `uf setup` runs
- **THEN** the step result reports "skipped"
- **AND** detail includes
  "https://devpod.sh/docs/getting-started/install"

### Requirement: Setup configures DevPod Podman provider

`uf setup` MUST configure a DevPod provider named
"podman" when both DevPod and Podman are installed and
no provider named "podman" is registered. Provider
detection MUST use exact name matching on the first
column of `devpod provider list` output (not substring
matching). Setup MUST use
`devpod provider add docker --name podman -o DOCKER_COMMAND=podman`.

#### Scenario: Both tools installed, no podman provider

- **GIVEN** DevPod is in PATH
- **AND** Podman is in PATH
- **AND** `devpod provider list` output does not contain
  a provider named "podman"
- **WHEN** `uf setup` runs the provider config step
- **THEN** `devpod provider add docker --name podman -o DOCKER_COMMAND=podman` is executed
- **AND** the step result reports "installed"

#### Scenario: Provider already registered

- **GIVEN** DevPod is in PATH
- **AND** `devpod provider list` output contains a
  provider named "podman"
- **WHEN** `uf setup` runs the provider config step
- **THEN** no provider add command is executed
- **AND** the step result reports "already installed"

#### Scenario: DevPod not available

- **GIVEN** DevPod is not in PATH
- **WHEN** `uf setup` runs the provider config step
- **THEN** the step is skipped with detail "no devpod"

#### Scenario: Podman not available

- **GIVEN** Podman is not in PATH
- **WHEN** `uf setup` runs the provider config step
- **THEN** the step is skipped with detail "no podman"

#### Scenario: Provider add fails

- **GIVEN** DevPod and Podman are in PATH
- **AND** no podman provider is registered
- **AND** `devpod provider add` fails
- **WHEN** `uf setup` runs the provider config step
- **THEN** the step result reports "failed"
- **AND** detail includes
  "devpod provider add docker --name podman -o DOCKER_COMMAND=podman"
- **AND** subsequent setup steps continue

#### Scenario: Provider list command fails

- **GIVEN** DevPod is in PATH
- **AND** `devpod provider list` fails (exit code != 0)
- **WHEN** `uf setup` runs the provider config step
- **THEN** the step result reports "skipped"
- **AND** detail includes a warning about provider
  list failure

### Requirement: Doctor checks Podman as required tool

`uf doctor` MUST check Podman as a required tool in
the Core Tools group. The check MUST validate presence
via `LookPath` and version via `podman --version` with
minimum version 4.3. Missing Podman MUST report Fail
severity.

#### Scenario: Podman installed, version sufficient

- **GIVEN** Podman is in PATH
- **AND** Podman version is >= 4.3
- **WHEN** `uf doctor` runs
- **THEN** the Podman check reports Pass
- **AND** the message includes the version number

#### Scenario: Podman installed, version too old

- **GIVEN** Podman is in PATH
- **AND** Podman version is < 4.3
- **WHEN** `uf doctor` runs
- **THEN** the Podman check reports Fail
- **AND** the install hint is "brew install podman"
  or "brew upgrade podman"

#### Scenario: Podman not installed

- **GIVEN** Podman is not in PATH
- **WHEN** `uf doctor` runs
- **THEN** the Podman check reports Fail
- **AND** the install hint is "brew install podman"

### Requirement: Doctor validates Podman runtime health

When Podman passes the presence and version checks,
`uf doctor` MUST perform a runtime health post-check
by running `podman info`. The post-check MUST be
platform-aware.

On macOS, doctor MUST first check for a Podman machine
via `podman machine list`. If no machine exists, the
check MUST report Fail with a hint to run
`podman machine init && podman machine start`. If a
machine exists but `podman info` fails, the check MUST
report Fail with a hint to run `podman machine start`.

On Linux, doctor MUST run `podman info` directly. If
it fails, the check MUST report Fail with a hint to
check `systemctl --user status podman.socket`.

#### Scenario: macOS, Podman machine running

- **GIVEN** the platform is macOS
- **AND** Podman is installed with version >= 4.3
- **AND** a Podman machine exists and is running
- **AND** `podman info` succeeds
- **WHEN** `uf doctor` runs
- **THEN** the Podman runtime check reports Pass
- **AND** the message indicates Podman is functional

#### Scenario: macOS, no Podman machine

- **GIVEN** the platform is macOS
- **AND** Podman is installed with version >= 4.3
- **AND** `podman machine list` returns no machines
- **WHEN** `uf doctor` runs
- **THEN** the Podman runtime check reports Fail
- **AND** the install hint is
  "podman machine init && podman machine start"

#### Scenario: macOS, Podman machine stopped

- **GIVEN** the platform is macOS
- **AND** Podman is installed with version >= 4.3
- **AND** a Podman machine exists
- **AND** `podman info` fails
- **WHEN** `uf doctor` runs
- **THEN** the Podman runtime check reports Fail
- **AND** the install hint is "podman machine start"

#### Scenario: Linux, Podman responsive

- **GIVEN** the platform is Linux
- **AND** Podman is installed with version >= 4.3
- **AND** `podman info` succeeds
- **WHEN** `uf doctor` runs
- **THEN** the Podman runtime check reports Pass

#### Scenario: Linux, Podman not responding

- **GIVEN** the platform is Linux
- **AND** Podman is installed with version >= 4.3
- **AND** `podman info` fails
- **WHEN** `uf doctor` runs
- **THEN** the Podman runtime check reports Fail
- **AND** the install hint mentions
  "systemctl --user status podman.socket"

#### Scenario: Podman version too old, skip runtime check

- **GIVEN** Podman is installed with version < 4.3
- **WHEN** `uf doctor` runs
- **THEN** the Podman version check reports Fail
- **AND** the runtime health post-check is skipped

### Requirement: Doctor detects docker-to-podman shim

When Podman passes its core tool checks, `uf doctor`
MUST check whether `docker` is in PATH and, if so,
resolve the binary via `EvalSymlinks` to determine if
it is a symlink or shim pointing to Podman. The check
MUST report an informational result.

If `docker` is not in PATH, the check SHOULD be
skipped silently (no result emitted).

#### Scenario: docker is a Podman symlink

- **GIVEN** `docker` is in PATH
- **AND** the resolved path of the `docker` binary
  contains "podman"
- **WHEN** `uf doctor` runs
- **THEN** the docker shim check reports Pass
- **AND** the message indicates "docker is a
  Podman shim"

#### Scenario: docker is real Docker

- **GIVEN** `docker` is in PATH
- **AND** the resolved path of the `docker` binary
  does not contain "podman"
- **WHEN** `uf doctor` runs
- **THEN** the docker shim check reports Pass
- **AND** the message indicates "Docker detected"
- **AND** the detail notes that the sandbox uses
  Podman, not Docker

#### Scenario: docker not in PATH

- **GIVEN** `docker` is not in PATH
- **WHEN** `uf doctor` runs
- **THEN** no docker shim check result is emitted

### Requirement: Setup verifies Podman after install

After Podman installation (and machine init on macOS),
`uf setup` SHOULD run `podman info` as a smoke test.
Smoke test failures MUST be reported as warnings and
MUST NOT block subsequent setup steps.

#### Scenario: Smoke test passes

- **GIVEN** Podman was just installed
- **AND** Podman machine was initialized (macOS)
- **AND** `podman info` succeeds
- **WHEN** `uf setup` runs the Podman step
- **THEN** the step result reports "installed"
- **AND** detail includes "verified"

#### Scenario: Smoke test fails

- **GIVEN** Podman was just installed
- **AND** `podman info` fails
- **WHEN** `uf setup` runs the Podman step
- **THEN** the step result reports "installed"
- **AND** detail includes "podman info failed"
- **AND** subsequent setup steps continue

### Requirement: Doctor checks DevPod version

When the DevPod check group is active, `uf doctor`
MUST validate DevPod version >= 0.5.0. Versions below
0.5.0 MUST report Warn severity.

#### Scenario: DevPod version sufficient

- **GIVEN** DevPod is installed
- **AND** DevPod version is >= 0.5.0
- **WHEN** `uf doctor` runs
- **THEN** the DevPod version check reports Pass

#### Scenario: DevPod version too old

- **GIVEN** DevPod is installed
- **AND** DevPod version is < 0.5.0
- **WHEN** `uf doctor` runs
- **THEN** the DevPod version check reports Warn
- **AND** an update hint is provided

### Requirement: Doctor Options GOOS injection

The doctor `Options` struct MUST include a `GOOS string`
field that overrides `runtime.GOOS` when non-empty.
When `GOOS` is empty, it MUST default to `runtime.GOOS`.
All platform-aware checks (Podman runtime health, install
hints) MUST use this field to enable cross-platform test
isolation per Constitution Principle IV.

#### Scenario: GOOS override in tests

- **GIVEN** `opts.GOOS` is set to "darwin"
- **AND** the test runs on Linux
- **WHEN** the Podman runtime health check runs
- **THEN** the check follows the macOS code path
  (machine list check before podman info)

#### Scenario: GOOS default

- **GIVEN** `opts.GOOS` is empty
- **WHEN** any platform-aware check runs
- **THEN** it uses `runtime.GOOS` as the platform

### Requirement: Doctor checks DevPod Podman provider

When the DevPod check group is active, `uf doctor`
MUST verify a provider named "podman" is registered.
Provider detection MUST use exact name matching on the
first column of `devpod provider list` output (not
substring matching). Missing provider MUST report Warn
severity with the fix command as install hint.

#### Scenario: Podman provider registered

- **GIVEN** DevPod is installed
- **AND** `devpod provider list` shows a provider
  named "podman"
- **WHEN** `uf doctor` runs
- **THEN** the provider check reports Pass

#### Scenario: Podman provider missing

- **GIVEN** DevPod is installed
- **AND** `devpod provider list` does not show a
  provider named "podman"
- **WHEN** `uf doctor` runs
- **THEN** the provider check reports Warn
- **AND** the install hint is
  `devpod provider add docker --name podman -o DOCKER_COMMAND=podman`

### Requirement: Install hints for Podman and DevPod

`homebrewInstallCmd`, `genericInstallCmd`, and
`installURL` MUST return appropriate values for
"podman" and "devpod" tool names.

#### Scenario: Homebrew install hint for Podman

- **GIVEN** the tool name is "podman"
- **WHEN** `homebrewInstallCmd` is called
- **THEN** it returns "brew install podman"

#### Scenario: Generic install hint for DevPod

- **GIVEN** the tool name is "devpod"
- **AND** Homebrew is not available
- **WHEN** `genericInstallCmd` is called
- **THEN** it returns a string containing
  "https://devpod.sh"

#### Scenario: Install URL for Podman

- **GIVEN** the tool name is "podman"
- **WHEN** `installURL` is called
- **THEN** it returns "https://podman.io"

#### Scenario: Install URL for DevPod

- **GIVEN** the tool name is "devpod"
- **WHEN** `installURL` is called
- **THEN** it returns "https://devpod.sh"

## MODIFIED Requirements

### Requirement: Setup step count

Previously: `Run()` executes 13 numbered steps.
Now: `Run()` MUST execute 16 numbered steps. All step
labels MUST use `[N/16]` format.

### Requirement: DevPod doctor check group

Previously: Checks devpod binary and devcontainer.json.
Now: Additionally MUST check DevPod version (>= 0.5.0)
and Podman provider registration.

## REMOVED Requirements

None.
