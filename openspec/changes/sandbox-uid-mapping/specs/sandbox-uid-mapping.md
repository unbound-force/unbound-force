## ADDED Requirements

### FR-001: Default UID Mapping on All Podman Runs

All `podman run` invocations from the sandbox MUST
include `--userns=keep-id:uid=1000,gid=1000` by
default. This applies to both `buildRunArgs()`
(ephemeral containers) and `buildPersistentRunArgs()`
(persistent workspaces).

#### Scenario: Ephemeral sandbox on Linux
- **GIVEN** a Linux host with rootless Podman >= 4.3
- **WHEN** `uf sandbox start` is executed
- **THEN** the `podman run` command includes
  `--userns=keep-id:uid=1000,gid=1000`
- **AND** files in `/workspace` appear owned by
  `dev:dev` (UID 1000:GID 1000) inside the container

#### Scenario: Persistent workspace on Linux
- **GIVEN** a Linux host with rootless Podman >= 4.3
- **WHEN** `uf sandbox create` is executed
- **THEN** the `podman run` command for the
  persistent container includes
  `--userns=keep-id:uid=1000,gid=1000`

#### Scenario: Direct mode writable
- **GIVEN** a running sandbox in `direct` mode with
  UID mapping enabled
- **WHEN** the container's `dev` user writes a file
  to `/workspace`
- **THEN** the write succeeds (no permission denied)

#### Scenario: Git operates without safe.directory
- **GIVEN** a running sandbox with UID mapping enabled
- **WHEN** `git status` is run inside the container
- **THEN** git operates normally without requiring
  `git config --global --add safe.directory`

### FR-002: macOS Podman Machine UID Probe

On macOS (`platform.OS == "darwin"`), the sandbox
MUST probe whether the Podman machine's virtiofs is
mapping UIDs correctly before starting the main
container.

#### Scenario: macOS probe succeeds
- **GIVEN** a macOS host with a properly configured
  Podman machine
- **WHEN** `uf sandbox start` is executed
- **THEN** a probe container runs
  `stat -c '%u' /test` on the mounted project dir
- **AND** the output is `1000`
- **AND** `PlatformConfig.UIDMapSupported` is set
  to `true`
- **AND** the main container starts normally with
  `--userns=keep-id:uid=1000,gid=1000`

#### Scenario: macOS probe fails
- **GIVEN** a macOS host with an unconfigured Podman
  machine (virtiofs maps files as UID 0)
- **WHEN** `uf sandbox start` is executed
- **AND** the `--uidmap` flag is NOT set
- **THEN** the probe container output is NOT `1000`
- **AND** `Start()` returns an error containing
  "Podman machine UID mapping is not configured"
- **AND** the error message includes remediation
  steps (podman machine stop/rm/init/start)
- **AND** the error message suggests `--uidmap` as
  an escape hatch

#### Scenario: macOS probe error (fail-safe)
- **GIVEN** a macOS host where the probe container
  fails to start
- **WHEN** `uf sandbox start` is executed
- **THEN** `probeUIDMapping()` returns `false`
- **AND** the same remediation error is shown
  (fail-safe: assume mapping is broken)

#### Scenario: Linux skips probe
- **GIVEN** a Linux host
- **WHEN** `uf sandbox start` is executed
- **THEN** no probe container is run
- **AND** `PlatformConfig.UIDMapSupported` is `true`
  unconditionally

### FR-003: `--uidmap` CLI Flag Override

A `--uidmap` boolean flag MUST be available on
`uf sandbox start` and `uf sandbox create`. When set,
explicit `--uidmap`/`--gidmap` arguments MUST replace
`--userns=keep-id:uid=1000,gid=1000`.

#### Scenario: `--uidmap` flag on start
- **GIVEN** a macOS host with unconfigured Podman
  machine
- **WHEN** `uf sandbox start --uidmap` is executed
- **THEN** the `podman run` command includes
  `--uidmap 1000:0:1`, `--uidmap 0:1:1000`,
  `--uidmap 1001:1001:64536` (and corresponding
  `--gidmap` entries)
- **AND** the command does NOT include `--userns`
- **AND** the macOS probe check is bypassed
- **AND** the container starts successfully

#### Scenario: `--uidmap` from config
- **GIVEN** `.uf/config.yaml` contains
  `sandbox: { uid_map: true }`
- **WHEN** `uf sandbox start` is executed without
  `--uidmap` flag
- **THEN** the `--uidmap`/`--gidmap` arguments are
  used (config applies)

#### Scenario: `--uidmap` from env var
- **GIVEN** `UF_SANDBOX_UIDMAP=1` is set in the
  environment
- **WHEN** `uf sandbox start` is executed
- **THEN** the `--uidmap`/`--gidmap` arguments are
  used

#### Scenario: `--uidmap` flag on create
- **GIVEN** any platform
- **WHEN** `uf sandbox create --uidmap` is executed
- **THEN** the persistent container uses
  `--uidmap`/`--gidmap` arguments

### FR-004: Post-Copy Ownership Fix

`PodmanBackend.Create()` MUST run
`chown -R dev:dev /workspace` inside the container
after `podman cp` and before the health check.

#### Scenario: chown after copy succeeds
- **GIVEN** a `uf sandbox create` operation that
  successfully copies project files via `podman cp`
- **WHEN** the copy completes
- **THEN** the sandbox runs
  `podman exec <ctr> chown -R dev:dev /workspace`
- **AND** a progress message "Fixing workspace
  permissions..." is printed
- **AND** all files in `/workspace` are owned by
  `dev:dev`

#### Scenario: chown fails triggers cleanup
- **GIVEN** a `uf sandbox create` operation where
  `podman cp` succeeds
- **WHEN** `chown -R dev:dev /workspace` fails
- **THEN** the container is removed
  (`podman rm -f <ctr>`)
- **AND** the volume is removed
  (`podman volume rm <vol>`)
- **AND** an error is returned containing
  "failed to fix permissions"

### FR-005: Detection Precedence

The UID mapping strategy MUST follow this precedence:

1. `--uidmap` flag (or config/env) -> explicit
   `--uidmap`/`--gidmap` args
2. macOS + probe fails -> error with remediation
3. Default -> `--userns=keep-id:uid=1000,gid=1000`

#### Scenario: Flag overrides probe failure
- **GIVEN** a macOS host where the probe would fail
- **WHEN** `uf sandbox start --uidmap` is executed
- **THEN** the `--uidmap` flag takes precedence
- **AND** no probe error is raised
- **AND** the container starts with explicit mapping

#### Scenario: Linux uses default without probe
- **GIVEN** a Linux host
- **WHEN** `uf sandbox start` is executed
- **THEN** no probe runs
- **AND** `--userns=keep-id:uid=1000,gid=1000` is
  used

### FR-006: Shared UID Mapping Args Helper

A shared helper function `uidMappingArgs(opts Options)
[]string` MUST be used by both `buildRunArgs()` and
`buildPersistentRunArgs()` to avoid duplicating the
flag generation logic.

#### Scenario: Helper returns keep-id by default
- **GIVEN** `Options{UIDMap: false}`
- **WHEN** `uidMappingArgs()` is called
- **THEN** it returns
  `["--userns=keep-id:uid=1000,gid=1000"]`

#### Scenario: Helper returns uidmap when overridden
- **GIVEN** `Options{UIDMap: true}`
- **WHEN** `uidMappingArgs()` is called
- **THEN** it returns a 12-element slice with
  `--uidmap` and `--gidmap` pairs
- **AND** it does NOT include `--userns`

### FR-010: Podman Version Check

`Start()` MUST validate that the installed Podman
version is >= 4.3 before using the extended
`--userns=keep-id:uid=N,gid=N` syntax. If the
version is too old, return an actionable error.

#### Scenario: Podman version too old
- **GIVEN** Podman 4.2.0 is installed
- **WHEN** `uf sandbox start` is executed
- **THEN** `Start()` returns an error containing
  "Podman >= 4.3 required"
- **AND** the error includes upgrade instructions

#### Scenario: Podman version sufficient
- **GIVEN** Podman 4.3.0 or later is installed
- **WHEN** `uf sandbox start` is executed
- **THEN** the version check passes silently
- **AND** the container starts normally

### FR-011: Rootful Podman Guard for `--uidmap`

When `--uidmap` is set, `Start()` MUST detect whether
Podman is running in rootful mode (via `podman info
--format '{{.Host.Security.Rootless}}'`). If rootful
is detected, MUST return an error rejecting `--uidmap`
because the explicit UID mapping grants container
UID 1000 actual host root access under rootful Podman.

#### Scenario: `--uidmap` rejected under rootful
- **GIVEN** rootful Podman (`Rootless: false`)
- **WHEN** `uf sandbox start --uidmap` is executed
- **THEN** `Start()` returns an error containing
  "--uidmap is only safe under rootless Podman"
- **AND** the container is NOT started

#### Scenario: `--uidmap` accepted under rootless
- **GIVEN** rootless Podman (`Rootless: true`)
- **WHEN** `uf sandbox start --uidmap` is executed
- **THEN** the rootful check passes
- **AND** the container starts with explicit mapping

### FR-012: Platform Injection for Testability

`Options` MUST include a `Platform *PlatformConfig`
field. When non-nil, `Start()` uses the provided
`PlatformConfig` instead of calling
`DetectPlatform()`. This allows tests to exercise
macOS detection logic on Linux CI runners.

#### Scenario: Platform injection overrides detection
- **GIVEN** `Options.Platform` is set to a
  `PlatformConfig{OS: "darwin", UIDMapSupported: false}`
- **WHEN** `Start()` is called on a Linux host
- **THEN** `Start()` uses the injected platform
- **AND** the macOS probe error is returned
- **AND** `DetectPlatform()` is NOT called

## MODIFIED Requirements

### FR-007: `buildRunArgs()` Includes User Namespace

Previously: `buildRunArgs()` did not include any user
namespace or UID mapping flags.

Now: `buildRunArgs()` MUST call `uidMappingArgs()`
and include the returned flags before the image
argument.

### FR-008: `buildPersistentRunArgs()` Includes
  User Namespace

Previously: `buildPersistentRunArgs()` did not include
any user namespace or UID mapping flags.

Now: `buildPersistentRunArgs()` MUST call
`uidMappingArgs()` and include the returned flags
before the image argument.

### FR-009: `PlatformConfig` Extended

Previously: `PlatformConfig` contained `OS`, `Arch`,
and `SELinux` fields.

Now: `PlatformConfig` MUST additionally contain
`UIDMapSupported bool`. Set to `true` on Linux
unconditionally, set by probe result on macOS.

## REMOVED Requirements

None.
