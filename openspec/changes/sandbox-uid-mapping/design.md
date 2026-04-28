## Context

The `uf sandbox` command (Specs 028, 029) launches
containerized OpenCode sessions via rootless Podman.
The container image (`quay.io/unbound-force/opencode-
dev:latest`) defines a non-root user `dev` (UID 1000).
Host project directories are bind-mounted at
`/workspace`.

Currently, neither `buildRunArgs()` nor
`buildPersistentRunArgs()` passes any user namespace
or UID mapping flags to `podman run`. On rootless
Podman, this causes a UID mismatch between the host
user and the container's `dev` user:

- **macOS**: virtiofs maps files as `root:nobody`
  (UID 0:65534) inside the container.
- **Linux**: the host user's UID maps to UID 0 inside
  the container; the container's `dev` (UID 1000)
  maps to a subordinate UID with no access to
  host-owned files.

The result: `direct` mode is broken, `git` refuses to
operate without `safe.directory`, and tools inside the
container see ownership they cannot act on.

Issue #108 explored five options (`:U` mount flag,
credential copy, `--user`, `--userns=keep-id` without
extended syntax, and `--userns=keep-id` on all
platforms). This design selects the extended
`--userns=keep-id:uid=1000,gid=1000` syntax
(Podman >= 4.3) which maps the host user directly to
the container's `dev` user, avoiding all five options'
drawbacks.

## Goals / Non-Goals

### Goals
- Mounted files appear owned by `dev:dev` (UID
  1000:GID 1000) inside the container on both macOS
  and Linux
- `direct` mode works correctly (container can write
  to host files)
- `git` operates without `safe.directory` workaround
- macOS users with unconfigured Podman machines get
  an actionable error instead of silent breakage
- Users with non-standard setups have an escape hatch
  via `--uidmap`
- Files copied via `podman cp` (persistent workspaces)
  are owned by `dev:dev`
- All new behavior is testable via existing injection
  points (Principle IV)

### Non-Goals
- Automatic Podman machine configuration (too
  platform-specific and version-dependent -- document
  the steps instead)
- Supporting arbitrary container users (the image
  defines `dev` as UID 1000; other images are out of
  scope)
- Rootful Podman support (rootless is the target
  deployment model)
- Docker support (Podman only per Spec 028)
- Changing the container image's user definition

## Decisions

### D1: `--userns=keep-id:uid=1000,gid=1000` as Default

Add `--userns=keep-id:uid=1000,gid=1000` to all
`podman run` invocations (both `buildRunArgs()` and
`buildPersistentRunArgs()`). This maps the calling
user's UID/GID to UID/GID 1000 inside the container.

**Rationale**: `--userns=keep-id` is the idiomatic
rootless Podman solution for bind-mount ownership.
The extended syntax `:uid=1000,gid=1000` (available
since Podman 4.3) explicitly targets the container's
`dev` user regardless of the host user's UID. This
works on both Linux and macOS (when the Podman machine
is properly configured). Supersedes Issue #108
Options A-E by avoiding their individual drawbacks.

**Trade-off**: Requires Podman >= 4.3. Podman 4.3 was
released in October 2022 -- all supported Fedora and
macOS Homebrew versions ship >= 4.3.

**Version enforcement**: `Start()` MUST parse
`podman --version` output and compare >= 4.3 before
using the extended syntax. If too old, return an
actionable error: "Podman >= 4.3 required for
--userns=keep-id:uid=N,gid=N. Current: X.Y.Z.
Upgrade: brew upgrade podman (macOS) or
dnf upgrade podman (Fedora)."

**Testability**: Verifiable by inspecting the args
slice returned by `buildRunArgs()` -- no subprocess
execution needed. Version check testable via
injectable `ExecCmd`.

### D2: macOS Podman Machine UID Detection

On macOS (`platform.OS == "darwin"`), before starting
the container, probe whether the Podman machine is
mapping UIDs correctly.

**Detection strategy**: Run a lightweight probe
container using a minimal image (`busybox:latest`)
with an explicit `--entrypoint` override to prevent
the image's default entrypoint from executing:
```
podman run --rm \
  --entrypoint stat \
  --userns=keep-id:uid=1000,gid=1000 \
  -v <ProjectDir>:/test:ro \
  busybox:latest -c '%u' /test
```
If the output is `1000`, the mapping works. If it is
`0` or any other value, virtiofs is not mapping UIDs
correctly. Using `busybox:latest` (~5MB) avoids
pulling the full dev image (~1GB) for a simple stat
check and eliminates entrypoint security concerns.

Print a progress message before the probe:
`"Checking Podman machine UID mapping..."` to
`opts.Stderr` (consistent with other progress
messages in `Start()`).

**Decision**: Use the probe approach. Run a
lightweight probe container on macOS before the main
container start. Cache the result in
`PlatformConfig.UIDMapSupported` for the session.
The TOCTOU window (Podman machine state changing
between probe and container start) is a known
limitation — if the main container start fails
despite the probe succeeding, the error message
should suggest `--uidmap` as a fallback.

**Error message when detection fails**:
```
Error: Podman machine UID mapping is not configured
for file ownership compatibility.

Files in the container will appear as root:nobody,
preventing git and file operations.

To fix, recreate the Podman machine with:
  podman machine stop
  podman machine rm
  podman machine init --rootful=false
  podman machine start

If you cannot reconfigure the Podman machine, use
--uidmap to override:
  uf sandbox start --uidmap
```

**Testability**: The probe uses `opts.ExecCmd`, so
tests can inject mock responses without running
Podman.

### D3: `--uidmap` CLI Flag as Escape Hatch

Add a `--uidmap` boolean flag to `uf sandbox start`
and `uf sandbox create`. When set, replace
`--userns=keep-id:uid=1000,gid=1000` with explicit
`--uidmap` and `--gidmap` arguments that map
container UID 1000 to the UID that owns the mounted
files (typically UID 0 on macOS virtiofs mounts).

The explicit mapping:
```
--uidmap 1000:0:1
--uidmap 0:1:1000
--uidmap 1001:1001:64536
--gidmap 1000:0:1
--gidmap 0:1:1000
--gidmap 1001:1001:64536
```

This maps (under rootless Podman):
- Container UID 1000 (dev) -> user namespace UID 0
  (= calling user's real UID on host, NOT actual
  root). This is the UID that virtiofs assigns to
  mounted files.
- Container UID 0 (root) -> user namespace UID 1
- Container UIDs 1001-65535 -> user namespace UIDs
  1001-65535

**Rootful Podman safety**: This mapping is ONLY safe
under rootless Podman. Under rootful Podman, "host
UID 0" means actual root — the mapping would grant
container UID 1000 real host root access (privilege
escalation). `Start()` MUST detect rootful Podman
(via `podman info --format
'{{.Host.Security.Rootless}}'`) and reject
`--uidmap` with a clear error if rootful is detected.

**Rationale**: This is the manual workaround for
macOS Podman machine configurations where
`--userns=keep-id` does not produce correct results
because virtiofs maps all files to UID 0 regardless.
The user opts in explicitly because this mapping is
fragile and depends on the specific virtiofs behavior.

**Config support**: Also configurable via
`.uf/config.yaml` under `sandbox.uid_map: true` or
env var `UF_SANDBOX_UIDMAP=1`. CLI flag takes
precedence.

**Testability**: Verifiable by inspecting args slice.

### D4: Post-Copy chown for Persistent Workspaces

In `PodmanBackend.Create()`, after the `podman cp`
step, run:
```
podman exec <ctrName> chown -R dev:dev /workspace
```

**Rationale**: `podman cp` operates outside the user
namespace. Even with `--userns=keep-id`, copied files
are owned by root inside the container. The `chown`
fixes this. The `dev` user and group exist in the
container image.

**Performance**: For large projects, `chown -R` may
take a few seconds. This is acceptable because
`Create()` is a one-time operation. Add a progress
message: "Fixing workspace permissions..."

**Partial failure**: If `chown` fails, clean up
(remove container and volume) and return error,
consistent with existing partial failure handling in
`Create()`.

**Testability**: The `chown` call goes through
`opts.ExecCmd`, which is injectable.

### D5: Detection Order in Start()

The UID mapping strategy selection follows this
precedence:

1. `--uidmap` flag (or config `sandbox.uid_map`) ->
   use explicit `--uidmap`/`--gidmap` args
2. macOS + probe fails -> error with remediation
3. Default -> `--userns=keep-id:uid=1000,gid=1000`

This precedence ensures:
- Explicit user override always wins
- macOS users get a clear error before wasting time
  on a broken sandbox
- Linux users get correct behavior with zero config

### D6: PlatformConfig Extension and Injection

Extend `PlatformConfig` with a new field:

```go
type PlatformConfig struct {
    OS              string
    Arch            string
    SELinux         bool
    UIDMapSupported bool  // true when keep-id works
}
```

On Linux, `UIDMapSupported` is always `true`.
On macOS, it is set by the probe (D2).

This keeps detection results in the existing
`PlatformConfig` struct, which is already threaded
through `buildRunArgs()` and
`buildPersistentRunArgs()`.

**Testability injection**: Add a `Platform
*PlatformConfig` field to `Options`. When non-nil,
`Start()` uses it instead of calling
`DetectPlatform()`. This allows tests to override
`runtime.GOOS` behavior and exercise macOS probe
and detection precedence logic on Linux CI. When
nil (zero value), `DetectPlatform()` is called as
before — no behavioral change for production code.

## Risks / Trade-offs

### R1: Podman Version Requirement

`--userns=keep-id:uid=1000,gid=1000` requires Podman
>= 4.3. Older versions support bare `--userns=keep-id`
but not the extended `:uid=N,gid=N` syntax.
Mitigation: all supported platforms (Fedora 38+,
macOS Homebrew) ship Podman >= 4.3. Add a version
check to `uf doctor` if needed in the future.

### R2: macOS Probe Latency

The detection probe (D2) runs a lightweight container
on macOS before the main sandbox start. This adds
~2-3 seconds to the first `uf sandbox start` on
macOS. Acceptable because it prevents a much more
confusing failure later. The probe is skipped entirely
on Linux.

### R3: chown Performance on Large Repos

`chown -R dev:dev /workspace` on a large monorepo
could take 5-10 seconds. This is a one-time cost
during `uf sandbox create`. Not a concern for typical
project sizes (< 10k files).

### R4: Explicit uidmap Fragility

The `--uidmap` escape hatch (D3) hardcodes a mapping
that assumes virtiofs assigns UID 0 to mounted files.
If a future Podman machine version changes this
behavior, the mapping breaks. Mitigation: `--uidmap`
is an opt-in escape hatch, not the default path. The
error message from D2 directs users to fix their
Podman machine first.

### R5: Named Volume Ownership

For persistent workspaces using named volumes
(not bind mounts), `--userns=keep-id` affects how
the volume is initialized. With `keep-id`, the volume
is created with the mapped UID. This should work
correctly but needs testing to confirm there are no
edge cases with volume reuse across different host
users.
