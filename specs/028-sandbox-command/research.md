# Research: Sandbox Command

**Branch**: `028-sandbox-command` | **Date**: 2026-04-12
**Spec**: [spec.md](spec.md)

## R1: Existing CLI Pattern — Cobra Command Registration

**Decision**: Follow the established `newXxxCmd()` pattern
from `cmd/unbound-force/main.go`. Create a `newSandboxCmd()`
function that returns a `*cobra.Command` with subcommands
registered via `AddCommand()`. Each subcommand delegates to
a `runSandboxXxx(params)` function with a params struct
carrying `io.Writer` for testability.

**Rationale**: Every existing command (`init`, `doctor`,
`setup`) follows this exact pattern. The `runXxx(params)`
delegation enables unit testing without subprocess execution
or `os.Stdout` mocking. Dewey confirms this pattern across
Specs 007, 011, 013, and the Replicator repo.

**Alternatives considered**:
- Standalone binary (`uf-sandbox`): Rejected because the
  feature is part of the `uf` CLI, not a separate tool.
  Composability (Constitution II) means the sandbox is
  a capability of the existing binary.
- Plugin architecture: Over-engineered for a single
  command group. YAGNI.

**Source**: `cmd/unbound-force/main.go` lines 28-31,
Dewey: `specs/007-mx-f-architecture/research` R1.

---

## R2: Options/ExecCmd Injection Pattern

**Decision**: Create a `sandbox.Options` struct in
`internal/sandbox/` with injectable function fields:
`LookPath`, `ExecCmd`, `Getenv`, `Stdout`, `Stderr`.
Follow the exact pattern from `internal/setup/setup.go`
and `internal/doctor/doctor.go`. Include a `defaults()`
method that fills zero-value fields with production
implementations.

**Rationale**: This is the established testability pattern
across the codebase. Constitution Principle IV (Testability)
requires all components to be testable in isolation without
external services. The `LookPath`/`ExecCmd` injection
pattern has been validated across 6+ specs (011, 017, 023,
024, 027) and confirmed by Dewey semantic search.

**Key fields**:
- `LookPath func(string) (string, error)` — binary
  detection (podman, opencode)
- `ExecCmd func(string, ...string) ([]byte, error)` —
  subprocess execution (podman run, podman stop, etc.)
- `Getenv func(string) string` — env var reading
  (API keys, UF_SANDBOX_IMAGE, OLLAMA_HOST)
- `Stdout io.Writer` — user-facing output
- `Stderr io.Writer` — progress/status messages
- `ProjectDir string` — project to mount into container
- `Mode string` — "isolated" or "direct"
- `Detach bool` — skip auto-attach
- `Image string` — container image override
- `Memory string` — memory limit (default: "8g")
- `CPUs string` — CPU limit (default: "4")

**Source**: `internal/setup/setup.go` lines 24-60,
`internal/doctor/doctor.go` lines 18-59.

---

## R3: Platform Detection — Architecture and SELinux

**Decision**: Create a `PlatformConfig` struct and a
`DetectPlatform(opts)` function in `internal/sandbox/detect.go`.
Use `runtime.GOOS` and `runtime.GOARCH` for architecture.
For SELinux, check `/etc/selinux/config` for
`SELINUX=enforcing` and verify with `getenforce` command.

**Rationale**: The doctor package already uses
`runtime.GOOS + "/" + runtime.GOARCH` for platform
detection (see `environ.go` line 14). SELinux detection
requires both file check and command check because:
1. The config file may say "enforcing" but SELinux could
   be in permissive mode at runtime.
2. `getenforce` gives the runtime state.
3. On macOS, neither exists — skip SELinux entirely.

**PlatformConfig struct**:
```go
type PlatformConfig struct {
    OS       string // "darwin" or "linux"
    Arch     string // "arm64" or "amd64"
    SELinux  bool   // true if SELinux is enforcing
}
```

**Volume mount flag logic**:
- SELinux enforcing → append `:Z` to volume mounts
- SELinux disabled/absent → no suffix
- macOS → no suffix (no SELinux)

**Alternatives considered**:
- Always add `:Z`: Rejected because it causes warnings
  on non-SELinux systems and is incorrect on macOS.
- Use `sestatus` instead of `getenforce`: Rejected because
  `getenforce` is simpler (single word output) and more
  universally available on SELinux systems.

**Source**: `internal/doctor/environ.go` line 14,
Podman documentation on SELinux volume labels.

---

## R4: Container Lifecycle — Podman Commands

**Decision**: Map each sandbox subcommand to specific
Podman CLI invocations. Use the container name
`uf-sandbox` as the single-container identifier.

**Command mapping**:

| Subcommand | Podman Commands |
|------------|----------------|
| `start` | `podman inspect uf-sandbox` (check existing), `podman pull` (if needed), `podman run -d --name uf-sandbox ...` |
| `stop` | `podman stop uf-sandbox`, `podman rm uf-sandbox` |
| `attach` | `opencode attach http://localhost:4096` (not podman) |
| `extract` | `podman exec uf-sandbox git format-patch ...`, `git am` on host |
| `status` | `podman inspect uf-sandbox --format json` |

**Container run flags**:
```
podman run -d \
  --name uf-sandbox \
  --hostname uf-sandbox \
  -p 4096:4096 \
  -v <project>:/workspace[:ro][:Z] \
  -e ANTHROPIC_API_KEY \
  -e OPENAI_API_KEY \
  -e OLLAMA_HOST=host.containers.internal:11434 \
  --memory 8g \
  --cpus 4 \
  <image>
```

**Rationale**: Using `podman inspect` to check container
existence is more reliable than parsing `podman ps` output.
The `--name` flag ensures only one container with that name
can exist. `podman run` with `-d` (detach) starts the
container in the background, then we poll the health check
before attaching.

**Alternatives considered**:
- `podman compose`: Over-engineered for a single container.
- Docker compatibility: Podman is the explicit requirement
  per the spec. Docker users can alias `podman=docker`.

---

## R5: Health Check Polling

**Decision**: After `podman run`, poll
`http://localhost:4096/health` (or equivalent OpenCode
endpoint) with exponential backoff. Start at 500ms
intervals, double each attempt, cap at 5s intervals.
Total timeout: 60 seconds (per FR-005).

**Implementation**:
```go
func waitForHealth(opts *Options, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    interval := 500 * time.Millisecond
    maxInterval := 5 * time.Second

    for time.Now().Before(deadline) {
        // Check if container is still running
        // Try HTTP GET to health endpoint
        // If 200 OK, return nil
        time.Sleep(interval)
        if interval < maxInterval {
            interval *= 2
        }
    }
    return fmt.Errorf("health check timed out after %s", timeout)
}
```

**Rationale**: Exponential backoff prevents hammering the
server during startup while still detecting readiness
quickly. The 60-second timeout is generous enough for
image pull + container init + OpenCode server startup.

**Testability**: The health check function accepts the
`Options` struct with an injectable `HTTPGet` function,
so tests can inject a mock HTTP client that returns
success or failure without real network access. See
plan decision D3 for the rationale for direct HTTP
over `podman exec`.

**Alternatives considered**:
- Fixed interval polling: Wastes time with long intervals
  or hammers with short ones.
- Container health check (`--health-cmd`): Podman supports
  this but it adds complexity to the `podman run` command
  and requires polling `podman inspect` for health status
  anyway. Direct polling is simpler.

---

## R6: Change Extraction — git format-patch / git am

**Decision**: Use `git format-patch` inside the container
to generate patch files, copy them to the host via
`podman cp`, present a summary for review, then apply
via `git am` on confirmation.

**Extraction flow**:
1. `podman exec uf-sandbox git -C /workspace log --oneline
   origin/HEAD..HEAD` — list commits since mount point
2. `podman exec uf-sandbox git -C /workspace format-patch
   origin/HEAD..HEAD --stdout` — generate unified patch
3. Display patch summary (files changed, insertions,
   deletions) via `git apply --stat`
4. Prompt for confirmation
5. `git am` on host to apply the patch

**Rationale**: `git format-patch` preserves commit
metadata (author, message, timestamp). `git am` applies
patches as proper commits. This is the standard Git
workflow for transferring changes between repositories.

**Edge cases**:
- No commits since mount: Report "no changes to extract"
- Merge conflicts during `git am`: Report the conflict
  and suggest `git am --abort` to undo
- Container not running: Fail with "no sandbox running"

**Alternatives considered**:
- `git diff` + `git apply`: Loses commit metadata.
  Acceptable as a fallback if `format-patch` fails.
- `podman cp` of the entire workspace: Copies everything
  including untracked files, build artifacts, etc.
  Too coarse.
- Bind mount in direct mode: In direct mode, changes
  are already on the host — extraction is unnecessary.
  The `extract` command should detect this and inform
  the user.

---

## R7: Environment Variable Forwarding

**Decision**: Forward a curated list of environment
variables from the host to the container. Use `-e VAR`
(without value) syntax so Podman reads the value from
the host environment. Add `--env-file` support for
custom variables.

**Default forwarded variables**:
```
ANTHROPIC_API_KEY
OPENAI_API_KEY
GEMINI_API_KEY
OPENROUTER_API_KEY
OLLAMA_HOST=host.containers.internal:11434
```

**Rationale**: LLM API keys are the minimum required for
the agent to function. `OLLAMA_HOST` is set explicitly
to the container-internal hostname that resolves to the
host machine, enabling the containerized Dewey to reach
the host's Ollama instance.

**Security note**: The `-e VAR` syntax (without `=value`)
ensures API key values do not appear in `podman inspect`
command fields or shell history. However, the container
process may still log these values at runtime — this is
the container image's responsibility, not the CLI's.

**Testability**: The variable list is a package-level
slice, easily overridden in tests. The `Getenv` injection
on `Options` enables testing without real env vars.

---

## R8: File Organization

**Decision**: Three files in `internal/sandbox/`:

| File | Responsibility |
|------|---------------|
| `sandbox.go` | Core logic: Start, Stop, Attach, Extract, Status functions. Options struct, defaults. |
| `detect.go` | Platform detection: DetectPlatform, SELinux check, architecture mapping. |
| `config.go` | Container configuration: buildRunArgs, env var list, volume mount construction. |
| `sandbox_test.go` | Tests for all three files (single test file per Go convention for internal packages). |

**Rationale**: Separation of concerns per SOLID Single
Responsibility. `detect.go` is pure platform detection
(no container logic). `config.go` is pure configuration
building (no execution). `sandbox.go` orchestrates both.
This makes each file independently testable.

**Alternatives considered**:
- Single `sandbox.go`: Would exceed 500 lines and mix
  concerns. Rejected per clean code principles.
- Separate test files per source file: Go convention
  allows either pattern. A single test file is fine for
  an internal package with ~3 source files.
