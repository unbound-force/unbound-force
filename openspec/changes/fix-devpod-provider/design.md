## Context

`DevPodBackend.Create()` gates on `LookPath("podman")` at
`devpod.go:64-70`, requiring `podman` to be in PATH. This
pre-flight check is incorrect for the current provider model.

**Provider name vs. type**: `uf setup` (`setup.go:1331`)
registers the docker provider under the **name** `podman`:
```
devpod provider add docker --name podman -o DOCKER_PATH=podman
```
When `devpod up --provider podman` runs, DevPod resolves the
provider by **name**, finding the docker-type provider aliased
as `podman`. The `--provider podman` flag at line 97 is
therefore correct — it matches the registered name.

The actual bug is the `LookPath("podman")` pre-flight check:
the docker provider aliased as `podman` resolves the container
runtime binary via `DOCKER_PATH` in its own configuration,
not via PATH lookup. The pre-flight enforces a now-invalid
prerequisite, blocking workspace creation on correctly
configured systems where `podman` is not in PATH.

`uf setup` and `uf doctor` (`checks.go:1433`, checking
`hasDevPodProvider(opts, "podman")`) already reflect the
correct model. Only `DevPodBackend.Create()` enforces the
stale PATH requirement.

Constitution alignment: see proposal.md. Key principle is
Composability First — removing the hard `podman` PATH
dependency restores the DevPod backend's standalone
functionality.

## Goals / Non-Goals

### Goals

- Restore `uf sandbox create --backend devpod` by removing
  the incorrect `podman in PATH` pre-flight check
- Keep `--provider podman` (the correct registered name)
- Update GoDoc to explain the docker-aliased-as-podman model
- Update tests to assert correct behavior and add regression
  coverage for the podman-absent-from-PATH happy path

### Non-Goals

- Redesigning the sandbox abstraction or backend interface
- Adding a docker socket reachability pre-flight check
  (letting `devpod up` report its own error is sufficient;
  the error is surfaced verbatim at line 126)
- Modifying `TestStart_PodmanMissing` — that test covers
  the Podman ephemeral backend's `Start()` at
  `sandbox.go:512`, which is correct and unrelated
- Addressing the latent `wsName` injection in
  `startServerViaSSH` (pre-existing, mitigated by
  `sanitizeRe`, tracked separately)

## Decisions

### D1: Remove pre-flight check entirely (do not replace)

The `LookPath("podman")` check at `devpod.go:64-70` is
removed without a replacement socket reachability check.

**Rationale**: The `backend.go:105-110` `ResolveBackend`
function already gates on `devpod` in PATH — this is the
correct architectural boundary for backend selection.
The docker provider aliased as `podman` resolves the
container runtime binary via `DOCKER_PATH` in its own
config, not via PATH lookup. Adding a docker socket check
in `Create()` would duplicate validation that belongs in
the DevPod binary itself. The `devpod up` error path
(line 126) already surfaces errors verbatim via
`fmt.Errorf("devpod up failed: %w: %s", ...)`, providing
actionable output.

**Trade-off**: The removed pre-flight check included an
install hint (`brew install podman`). After removal, if
the container runtime is unreachable, the user sees
`devpod up`'s own error message instead. This is acceptable
because `uf setup` is the canonical path for configuring
the container runtime, and its error messages are already
clear.

### D2: Keep `--provider podman` (do not change to docker)

The `--provider podman` flag at `devpod.go:97` is correct
and MUST NOT be changed. It references the provider
**name** registered by `uf setup`, not the removed
standalone provider.

**Rationale**: `uf setup` (`setup.go:1331`) runs:
`devpod provider add docker --name podman -o DOCKER_PATH=podman`
This registers a docker-type provider under the **name**
`podman`. `devpod up --provider podman` resolves by name,
finding this alias. `uf doctor` (`checks.go:1433`) validates
`hasDevPodProvider(opts, "podman")`. Changing to
`--provider docker` would fail — no provider named `docker`
is registered after `uf setup`.

### D3: Replace both podman-absent error tests with one success test

Both `TestDevPodCreate_NotInstalled` (line 3689) and
`TestDevPodCreate_PodmanNotInstalled` (line 3992) assert
that `Create()` returns a "podman not found" error. After
the pre-flight removal, both become invalid. They are
replaced with a single `TestDevPodCreate_NoPodmanInPath`
which asserts `Create()` **succeeds** when
`LookPath("podman")` returns an error.

**Rationale**: Simply deleting the tests leaves a gap —
a future contributor could re-introduce the podman
LookPath check and no test would catch it. The replacement
test asserts the intended behavior: DevPod backend does
not require podman in PATH. Two tests for the same
assertion is redundant; one replacement covers both.

### D4: TestStart_PodmanMissing is explicitly out of scope

This test exercises `Start()` (sandbox.go:479), which
dispatches to the ephemeral Podman backend path. The
`podman in PATH` check at `sandbox.go:512` is correct
for the ephemeral mode and must remain unchanged.

## Risks / Trade-offs

### Risk: Docker socket privilege model

The docker provider aliased as `podman` routes through a
Docker-compatible socket. On Linux, the system Docker
socket (`/var/run/docker.sock`) is root-equivalent.
`uf setup` mitigates this by configuring
`DOCKER_PATH=podman`, which routes through Podman's
rootless socket. This maintains the existing security
posture but is not enforced by `DevPodBackend.Create()` —
it depends on `uf setup` having run correctly.

**Security invariant**: The provider's security posture
depends on `uf setup` having configured
`DOCKER_PATH=podman`. This change does not introduce or
modify that dependency. If `uf setup` has not been run,
`devpod up --provider podman` will fail with "provider
not found" (safe default).

**Mitigation**: This is a pre-existing condition. The fix
does not introduce new privilege escalation; it aligns the
code with the already-deployed `uf setup` behavior.

### D5: Add diagnostic hint to `devpod up` error path

The `devpod up` error at line 126 is augmented with a
diagnostic hint: `"Run 'uf doctor' to diagnose or
'uf setup' to configure."` This replaces the removed
install hint (`brew install podman`) with guidance that
covers all failure modes, not just missing podman.

**Rationale**: The removed pre-flight check included
`Install: brew install podman`. After removal, raw
`devpod up` errors may be opaque. The diagnostic hint
directs users to `uf doctor` (which validates the full
provider configuration) and `uf setup` (which configures
everything). This is more useful than the old
podman-specific hint because it covers provider
misconfiguration, missing Docker socket, and other
failure modes.

### Risk: `uf setup` not run

If a user runs `uf sandbox create --backend devpod`
without first running `uf setup`, no provider named
`podman` will be registered, and `devpod up` will fail.
This is not a new failure mode — the old code would also
fail (either at the LookPath check or at `devpod up`).
The error surfaces verbatim at line 126.

**Mitigation**: `uf doctor` validates the provider
registration and provides the exact install command.
The `ResolveBackend` function at `backend.go:105-110`
already gates on `devpod` in PATH as the first check.
