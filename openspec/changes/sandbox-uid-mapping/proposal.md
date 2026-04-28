## Why

The `uf sandbox` container mounts the host project
directory into the container at `/workspace`. On
rootless Podman, file ownership inside the container
does not match the container's `dev` user (UID 1000)
because no user namespace mapping is configured. This
manifests differently on each supported platform:

**macOS (Podman machine + virtiofs)**: Files appear as
`root:nobody` (UID 0:65534) inside the container
because the Podman machine VM's virtiofs layer does
not preserve the macOS user's identity. The container's
`dev` user cannot write to the mounted project
directory even in `direct` mode.

**Fedora/RHEL (native rootless Podman)**: The host
user's UID (typically 1000) is mapped to UID 0 (root)
inside the container by default. Files owned by the
host user appear as root-owned. The container's `dev`
user (UID 1000 inside the container) maps to a
subordinate UID on the host and cannot write to
host-owned files.

Neither `buildRunArgs()` nor `buildPersistentRunArgs()`
passes `--userns`, `--user`, `--uidmap`, or `--gidmap`
flags to `podman run`. The result is that `direct`
mode is broken (container cannot write to host files),
and tools inside the container (git, opencode, gaze)
see ownership mismatches that cause warnings or
failures (e.g., `git` refuses to operate on a repo
with "dubious ownership").

This is a blocking issue for the sandbox feature --
discovered while running a live sandbox session on
macOS with rootless Podman, where the git
safe.directory workaround was required and file writes
were impossible. Related design discussion in
Issue #108 (gcloud credential strategy) explored
five options; this change implements the
`--userns=keep-id:uid=1000,gid=1000` approach
(supersedes Issue #108 Option E with the extended
`:uid=N,gid=N` syntax available since Podman 4.3).

## What Changes

### Platform-Aware UID Mapping Strategy

Add a three-tier UID mapping strategy that selects
the correct Podman user namespace flags based on the
host platform:

1. **Fedora/Linux (default)**: Add
   `--userns=keep-id:uid=1000,gid=1000` to all
   `podman run` invocations. This maps the calling
   user's UID/GID to UID/GID 1000 inside the
   container (the `dev` user), making mounted files
   appear correctly owned. No additional configuration
   required.

2. **macOS (Podman machine with virtiofs)**: Detect
   whether the Podman machine is configured for
   correct UID mapping by running a lightweight probe
   container. If the probe fails (files appear as
   `root:nobody`), error with an actionable message
   explaining how to configure the Podman machine.
   If the probe passes, use
   `--userns=keep-id:uid=1000,gid=1000` (same as
   Linux).

3. **macOS with `--uidmap` override**: When the user
   passes `--uidmap` on the command line, use explicit
   `--uidmap` / `--gidmap` arguments instead of
   `--userns=keep-id`. This provides an escape hatch
   for macOS users whose Podman machine configuration
   cannot be changed, or for non-standard container
   images with a different user UID.

### Podman Machine Configuration Detection

Extend `DetectPlatform()` with a probe that runs a
short-lived container on macOS to check whether
virtiofs is mapping UIDs correctly. If the mapping is
wrong, `Start()` returns an error with remediation
steps instead of silently proceeding with broken file
ownership.

### Post-Copy Ownership Fix for Persistent Workspaces

In `PodmanBackend.Create()`, after `podman cp` copies
source files into the named volume, run
`podman exec <container> chown -R dev:dev /workspace`
to fix ownership. `podman cp` operates outside the
user namespace mapping, so copied files inherit the
caller's UID (root inside the container) rather than
the mapped `dev` user.

## Capabilities

### New Capabilities
- `uid-mapping`: All `podman run` invocations include
  `--userns=keep-id:uid=1000,gid=1000` on Linux,
  ensuring the container's `dev` user owns mounted
  files.
- `podman-machine-detection`: On macOS, the sandbox
  detects whether the Podman machine's virtiofs is
  configured for correct UID mapping and errors with
  actionable guidance if not.
- `uidmap-override`: New `--uidmap` flag on
  `uf sandbox start` and `uf sandbox create` allows
  explicit UID/GID mapping for non-standard setups
  (primarily macOS escape hatch).
- `post-copy-chown`: `PodmanBackend.Create()` fixes
  file ownership after `podman cp` so the `dev` user
  owns all workspace files.

### Modified Capabilities
- `buildRunArgs()`: Adds `--userns=keep-id:uid=1000,
  gid=1000` (or `--uidmap`/`--gidmap` when override
  is active) before the image argument.
- `buildPersistentRunArgs()`: Same UID mapping flags.
- `PodmanBackend.Create()`: Adds `chown` step after
  `podman cp`.
- `DetectPlatform()`: Extended with Podman machine
  UID mapping detection on macOS.
- `Options` struct: New `UIDMap` bool field.
- `SandboxConfig`: New `uid_map` config field.

### Removed Capabilities
- None.

## Impact

### Files Modified

| File | Change |
|------|--------|
| `internal/sandbox/config.go` | Add `--userns=keep-id` to `buildRunArgs()`, extract `uidMappingArgs()` helper, add `--uidmap`/`--gidmap` when `UIDMap` is set |
| `internal/sandbox/podman.go` | Add `--userns=keep-id` to `buildPersistentRunArgs()`, add `chown` after `podman cp` |
| `internal/sandbox/detect.go` | Add `UIDMapSupported` to `PlatformConfig`, add `probeUIDMapping()` on macOS |
| `internal/sandbox/sandbox.go` | Add `UIDMap` field to `Options`, integrate detection into `Start()` |
| `cmd/unbound-force/sandbox.go` | Add `--uidmap` flag to `start` and `create` commands |
| `internal/config/config.go` | Add `UIDMap` field to `SandboxConfig` |
| `internal/sandbox/sandbox_test.go` | Tests for UID mapping flags, macOS detection, chown step |

### Behavioral Changes

- All `podman run` invocations now include user
  namespace mapping flags. Files inside the container
  appear owned by `dev:dev` (UID 1000:GID 1000)
  instead of `root:nobody` or `root:root`.
- On macOS, if the Podman machine is not configured
  for correct UID mapping, `uf sandbox start` fails
  with an actionable error instead of silently
  producing broken file ownership.
- `uf sandbox create` now runs `chown -R dev:dev
  /workspace` after copying source files.
- New `--uidmap` flag available on `start` and
  `create` subcommands.

### No Breaking Changes

- Isolated mode (read-only) behavior unchanged --
  the UID mapping makes `git` and read operations
  work correctly without the `safe.directory`
  workaround.
- Direct mode now works correctly (files are writable
  by the container's `dev` user).
- Persistent workspace lifecycle unchanged.
- Gateway integration unchanged.
- CDE (Eclipse Che) backend unchanged.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

This change is internal to the sandbox package. It
does not affect inter-hero artifact formats,
communication protocols, or artifact envelope schemas.
The sandbox is infrastructure -- heroes never interact
with it directly.

### II. Composability First

**Assessment**: PASS

The sandbox remains independently usable (`uf sandbox
start`). The `--uidmap` flag is optional with a
sensible default. No new mandatory dependencies are
introduced. The `--userns=keep-id` flag is a standard
Podman feature available in all supported versions
(>= 4.3).

### III. Observable Quality

**Assessment**: PASS

This change improves observability by replacing a
silent failure (wrong file ownership with no error)
with an explicit error on macOS when the Podman
machine is misconfigured. The error message includes
remediation steps.

### IV. Testability

**Assessment**: PASS

All detection logic uses the existing injectable
function pattern (`ExecCmd`). UID mapping flag
generation is testable by inspecting the
`buildRunArgs()` output. macOS detection is testable
by injecting mock `ExecCmd` responses. The `chown`
step is testable via the same `ExecCmd` injection.
No external services required.
