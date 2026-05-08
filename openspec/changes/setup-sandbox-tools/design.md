## Context

`uf setup` installs 13 tools (OpenCode, Gaze, gh, Node,
OpenSpec, uv, Specify, Replicator, replicator setup,
Ollama, Dewey, golangci-lint, govulncheck). `uf doctor`
validates 10 check groups plus a conditional DevPod group.

Neither command handles Podman or DevPod installation.
The existing DevPod doctor group only checks binary
presence and devcontainer.json existence -- it does not
verify version or provider configuration.

The DevPod ecosystem changed: the standalone `podman`
provider was removed. Users must alias the Docker
provider: `devpod provider add docker --name podman -o DOCKER_COMMAND=podman`.

## Goals / Non-Goals

### Goals

- Install Podman via `uf setup` with platform awareness
  (macOS Podman machine init, Linux direct install)
- Install DevPod via `uf setup` with Homebrew
- Automatically configure DevPod's Podman provider alias
- Add Podman as a required tool in `uf doctor` with
  version check (>= 4.3) and runtime health validation
- Validate Podman machine state on macOS (exists and
  running) and Podman daemon responsiveness on Linux
- Enhance DevPod doctor checks with version (>= 0.5.0)
  and provider registration validation
- Add install hints and URLs for both tools
- Smoke-test Podman after setup installation

### Non-Goals

- Podman machine lifecycle management beyond initial
  `machine init && machine start` during setup
- Docker backend support for DevPod (only Podman alias)
- Modifying the sandbox package itself (the
  `autoDetectBackend` routing bug is a separate change)
- Adding Podman Desktop GUI installation

## Decisions

### D1: Podman as required in doctor

Podman is classified as `required: true` in
`coreToolSpecs` because the sandbox is becoming
non-optional. Doctor will report Fail severity when
Podman is missing.

### D2: DevPod as recommended in doctor

DevPod is classified as `recommended: true` (Warn if
missing) in the conditional DevPod check group. DevPod
adds persistent workspace management on top of Podman
but the ephemeral sandbox path works without it.

### D3: DOCKER_COMMAND over DOCKER_HOST

The provider alias uses `-o DOCKER_COMMAND=podman`
rather than `-o DOCKER_HOST=unix:///run/user/$UID/podman/podman.sock`.
DOCKER_COMMAND is simpler, portable across platforms,
and avoids UID-dependent socket paths that vary between
Linux and macOS Podman machine configurations.

### D4: Setup step positioning

New steps are inserted after Ollama (current step 10)
and before Dewey (current step 11), grouping container
infrastructure tools together. New ordering:

| Step | Tool |
|------|------|
| 1-10 | Unchanged (OpenCode through Ollama) |
| 11 | Podman |
| 12 | DevPod |
| 13 | DevPod provider configuration |
| 14 | Dewey (was 11) |
| 15 | golangci-lint (was 12) |
| 16 | govulncheck (was 13) |

Step count label updates from `[N/13]` to `[N/16]`.

### D5: Provider detection via ExecCmd

The provider check parses `devpod provider list` output
using exact name matching on the first column. The
output is split by lines, then each line is split by
whitespace, and the first field is compared exactly to
"podman". This avoids false positives from providers
with "podman" as a substring (e.g., "podman-custom").
DevPod's JSON output format is not stable for
`provider list`, so table parsing is used.

### D6: macOS Podman machine guard

On macOS, Podman requires a VM (Podman machine). After
installing Podman, setup checks if a machine exists
(`podman machine list --format '{{.Name}}'`). If no
machine exists, it runs `podman machine init` (with a
180-second timeout to prevent indefinite hangs on slow
networks) and `podman machine start`. This is a
best-effort step -- failures are reported but do not
block subsequent steps.

### D7: Podman version parsing

Podman version output is `podman version X.Y.Z`. The
parser follows the same pattern as `parseGoVersion` and
`parseNodeVersion` in the existing doctor code: extract
the version string, split on dots, compare major.minor
against the minimum (4.3).

### D7a: DevPod version parsing

DevPod version output is a single line: `v0.X.Y` (e.g.,
`v0.6.15`). The parser strips the leading `v`, splits
on dots, and compares major.minor against the minimum
(0.5). Pre-release suffixes (e.g., `0.6.15-beta`) are
handled by truncating at the first hyphen, matching the
existing sandbox parser pattern in `devpod.go:252`.

### D7b: GOOS injection for doctor Options

The doctor `Options` struct gains a `GOOS string` field
that overrides `runtime.GOOS` when non-empty. This
matches the existing pattern in the setup `Options`
struct (line 63 of `setup.go`) and keeps the two sibling
packages consistent. The sandbox package uses a richer
`Platform *PlatformConfig` struct, but doctor's needs
are simpler (only OS string needed for branching) and
the lightweight `GOOS string` approach avoids
over-engineering.

### D8: Podman runtime health check in doctor

After Podman passes the version check in `coreToolSpecs`,
a post-check validates that Podman is actually functional
by running `podman info`. This follows the Ollama
post-check pattern (where after presence/version, the
embedding model is verified).

The post-check is platform-aware:

**macOS**: Before `podman info`, check
`podman machine list --format '{{.Name}}'` to verify a
machine exists. If no machine exists, report Fail with
hint: "No Podman machine found. Run: podman machine init && podman machine start".
If a machine exists but `podman info` fails, report Fail
with hint: "Podman machine may not be running. Run: podman machine start".

**Linux**: Run `podman info` directly. If it fails,
report Fail with hint:
"Podman not responding. Check: systemctl --user status podman.socket".

The post-check uses `ExecCmd` injection, so tests can
simulate all failure modes without real Podman.

### D8a: Docker-to-Podman shim detection

After the Podman runtime health check, doctor checks
whether `docker` is in PATH. If found, it resolves the
binary path via `opts.EvalSymlinks` (already injected
on Options) and checks if the resolved path contains
"podman". This detects the common pattern where
`/usr/local/bin/docker -> /opt/podman/bin/podman`.

The check is informational only (Pass severity in both
cases). When docker is a Podman shim, the user gets
confirmation that docker commands will route to Podman.
When docker is real Docker, the user is informed that
the sandbox uses Podman, not Docker — avoiding
confusion about which container runtime is active.

If `docker` is not in PATH, the check is silently
skipped (no result emitted) to keep output clean.

Uses `opts.LookPath` for presence detection and
`opts.EvalSymlinks` for symlink resolution — both
already injectable on Options.

### D9: Setup smoke test after Podman install

After Podman installation (and machine init on macOS),
setup runs `podman info` as a smoke test to verify the
installation is functional. This is a best-effort
check: failures are reported as warnings but do not
block subsequent steps. Users who see the warning may
need to restart their terminal or manually start the
Podman machine.

### D10: Testability (Constitution IV)

All new functions accept `*Options` with injected
`LookPath`, `ExecCmd`, and `ReadFile`. No real binaries
or network calls in tests. Provider detection is tested
by injecting canned `devpod provider list` output into
`ExecCmd`. Podman runtime checks are tested by injecting
canned `podman info` and `podman machine list` output.

## Risks / Trade-offs

### R1: DevPod provider list output format

The provider check parses human-readable table output
from `devpod provider list`. If DevPod changes its
output format, the check may produce false negatives.
Mitigation: the check uses exact first-column matching
which is resilient to column width and ordering changes.
If `devpod provider list` itself fails, the check
degrades to a skip with warning rather than a false
negative.

### R2: macOS Podman machine init is slow

`podman machine init` downloads a VM image (~300MB) and
can take 30-60 seconds. This may surprise users during
setup. Mitigation: setup prints progress messages and
the step is skipped if a machine already exists.

### R3: Homebrew-only install path

Both Podman and DevPod use Homebrew as the primary
install method. Users without Homebrew get a skip with
download URL. This matches the existing pattern for
Ollama, Replicator, and other tools.

### R4: podman info may fail after fresh install

On macOS, `podman info` may fail immediately after
`brew install podman` if the Podman machine hasn't
fully started. The setup smoke test reports this as a
warning rather than a failure. Doctor will catch the
issue on subsequent runs.

### R5: Version parser duplication (accepted)

The doctor package needs `parsePodmanVersion` and
`parseDevPodVersion` functions that duplicate logic
from `internal/sandbox/detect.go` and `devpod.go`.
This duplication is intentional and accepted: the
parsers are ~10 lines each, and coupling doctor to
sandbox would violate package independence. The doctor
parsers follow the doctor package's existing pattern
(`versionParse func(output string) (string, error)` +
`versionCheck func(version string, min string) bool`)
rather than the sandbox package's `(int, int, error)`
pattern. Both parsers MUST handle the same edge cases
as their sandbox counterparts (leading `v` prefix for
DevPod, pre-release suffixes). If a third consumer
appears, extraction to `internal/toolversion/` should
be considered.

## Coverage Strategy

All new code is tested at the unit level via dependency
injection (Constitution IV). No integration or e2e tests
are required -- all external tool interactions are
injected via `Options` struct fields. The existing
`Run()` integration test pattern covers step ordering.
Coverage target: maintain existing CI ratchet (no
regression). New functions must have test coverage for
happy path, error path, and platform branching.
