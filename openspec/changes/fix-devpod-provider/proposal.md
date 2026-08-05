## Why

`DevPodBackend.Create()` at `internal/sandbox/devpod.go:64-70`
gates on `LookPath("podman")`, requiring `podman` to be in
PATH before proceeding. This pre-flight check is incorrect:
`uf setup` registers the docker provider aliased as `podman`
(`devpod provider add docker --name podman -o DOCKER_PATH=podman`)
and the container runtime binary is resolved by DevPod's own
provider configuration, not by PATH lookup.

The `--provider podman` flag at line 97 is correct — it
references the provider **name** registered by `uf setup`,
not the removed standalone `loft-sh/devpod-provider-podman`.
Only the `LookPath("podman")` pre-flight check needs removal.

`uf setup` (`setup.go:1291-1341`) and `uf doctor`
(`checks.go:1433`) already use the docker-aliased-as-podman
model correctly. Only `DevPodBackend.Create()` enforces a
now-invalid PATH prerequisite, causing a reproducible failure
of `uf sandbox create --backend devpod` when `podman` is not
in PATH on an otherwise correctly configured system.

Fixes #431.

## What Changes

1. **Pre-flight check removal**: Remove the `LookPath("podman")`
   guard at `devpod.go:64-70`. The docker provider aliased as
   `podman` does not require `podman` in PATH —
   `DOCKER_PATH=podman` (configured by `uf setup`) wires the
   binary within DevPod's own provider config. The `devpod up`
   error path already surfaces errors verbatim via
   `fmt.Errorf`.

2. **Provider flag**: Keep `--provider podman` at line 97.
   This is the registered provider **name** (not the removed
   standalone provider). `uf setup` registers
   `devpod provider add docker --name podman`, and
   `uf doctor` validates `hasDevPodProvider(opts, "podman")`.
   Changing this to `--provider docker` would break — no
   provider named `docker` is registered.

3. **Diagnostic hint**: Add `"Run 'uf doctor' to diagnose
   or 'uf setup' to configure"` to the `devpod up` error
   path (line 126), replacing the removed install hint
   with broader guidance.

4. **GoDoc corrections**: Update stale references in
   `devpodWorkspaceName` (line 45) and `Create()`
   (lines 54, 58) to reflect the docker-aliased-as-podman
   model and removal of the PATH pre-flight.

5. **Test updates**: Update `TestDevPodCreate_Success` to
   verify success when podman is absent from PATH (keeping
   `--provider podman` assertion). Replace both
   `TestDevPodCreate_NotInstalled` (line 3689) and
   `TestDevPodCreate_PodmanNotInstalled` (line 3992) with a
   single `TestDevPodCreate_NoPodmanInPath` test verifying
   `Create()` succeeds when podman is absent from PATH.

## Capabilities

### New Capabilities

- None

### Modified Capabilities

- `DevPodBackend.Create()`: No longer requires `podman` in
  PATH. The `--provider podman` flag is retained — it
  references the docker provider registered under the name
  `podman` by `uf setup`, which is the correct provider model.

### Removed Capabilities

- `podman in PATH` pre-flight check in `DevPodBackend.Create()`:
  No longer required — the docker provider aliased as `podman`
  resolves the container runtime binary via `DOCKER_PATH`
  in its provider configuration, not via PATH lookup.

## Impact

- **Files**: `internal/sandbox/devpod.go`,
  `internal/sandbox/sandbox_test.go`
- **User-facing**: `uf sandbox create --backend devpod` will
  work on correctly configured systems (where `uf setup` has
  run) even when `podman` is not in PATH.
- **No API changes**: No exported function signatures change.
- **No configuration changes**: `uf setup` already configures
  the docker provider correctly. The `--provider podman` flag
  already matches the registered provider name.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change modifies a subprocess invocation flag and removes
a pre-flight check. It does not affect artifact-based
communication or inter-hero collaboration patterns.

### II. Composability First

**Assessment**: PASS

The fix removes a hard dependency on `podman` being in PATH
for the DevPod backend to function. After this change, the
DevPod backend delegates container runtime resolution to
DevPod's own provider configuration (the docker provider
aliased as `podman` with `DOCKER_PATH=podman`), which is the
correct composability boundary. The `uf sandbox` subsystem
remains independently usable.

### III. Observable Quality

**Assessment**: N/A

This change does not alter any output formats, provenance
metadata, or machine-parseable artifacts. Error messages
remain structured and actionable.

### IV. Testability

**Assessment**: PASS

All affected production paths have corresponding test
updates specified. The replacement test
(`TestDevPodCreate_NoPodmanInPath`) verifies the
behavioral change — that `Create()` succeeds without
podman in PATH. Both existing tests asserting the old
pre-flight error (`TestDevPodCreate_NotInstalled` at
line 3689 and `TestDevPodCreate_PodmanNotInstalled` at
line 3992) are replaced by this single test. The existing
`TestStart_PodmanMissing` (ephemeral Podman backend) is
explicitly out of scope and must remain unchanged,
preserving its correct regression coverage.

### V. Security by Default

**Assessment**: PASS

The docker provider aliased as `podman` routes through a
Docker-compatible socket. On Linux, the system Docker
socket (`/var/run/docker.sock`) is root-equivalent.
However, `uf setup` configures `DOCKER_PATH=podman`,
which routes through Podman's rootless socket. This
maintains the existing security posture. The security
boundary is maintained by `uf setup`, not by
`DevPodBackend.Create()`.

**Security invariant**: The provider's security posture
depends on `uf setup` having configured `DOCKER_PATH=podman`.
This change does not introduce or modify that dependency.
If `uf setup` has not been run, `devpod up` will fail
with a provider-not-found error (safe default — no silent
fallback to a root-equivalent socket).

The pre-existing latent injection concern in
`startServerViaSSH` (wsName concatenation at line 283) is
mitigated by `sanitizeRe` in `projectName()` and is out
of scope for this change.
