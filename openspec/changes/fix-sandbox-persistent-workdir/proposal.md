## Why

`uf sandbox create --backend podman` copies the project
directory into a named volume at `/workspace/<project-name>/`
via `podman cp`, but `buildPersistentRunArgs()` never sets
`--workdir` or the `WORKSPACE` environment variable. The
container image's entrypoint defaults to `cd /workspace`,
so OpenCode starts one directory level above the actual
project.

This is the persistent-mode variant of the same bug fixed
in PR #123 for the ephemeral path. PR #123 added `WORKSPACE`
env var passing to `buildRunArgs()` (ephemeral), but the
fix was never applied to `buildPersistentRunArgs()`
(persistent).

## What Changes

Add `--workdir` and `-e WORKSPACE=` to
`buildPersistentRunArgs()` so the container starts in
`/workspace/<project-name>/` — the directory where
`podman cp` placed the project source.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `buildPersistentRunArgs()`: Sets `--workdir` and
  `WORKSPACE` env var to `/workspace/<project-basename>`
  so the entrypoint's `cd "$WORKSPACE"` lands in the
  project directory.

### Removed Capabilities

(none)

## Impact

- `internal/sandbox/podman.go` — `buildPersistentRunArgs()`
  gains `--workdir` and `WORKSPACE` env var (~5 lines)
- `internal/sandbox/sandbox_test.go` — new/updated
  assertions for persistent args tests
- No CLI, config, or user-facing API changes
- No container image changes (entrypoint already reads
  `WORKSPACE`)

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This is an internal plumbing fix within the sandbox
package. No inter-hero artifact interfaces are affected.

### II. Composability First

**Assessment**: PASS

The fix is contained within `buildPersistentRunArgs()`
and does not introduce new dependencies. The Podman
backend remains independently usable.

### III. Observable Quality

**Assessment**: PASS

The `WORKSPACE` env var is observable via
`podman exec <container> env | grep WORKSPACE` and the
working directory is observable via `pwd` inside the
container. Both are machine-parseable.

### IV. Testability

**Assessment**: PASS

The fix is testable via existing `buildPersistentRunArgs`
unit tests — assertions verify `--workdir` and
`WORKSPACE=` are present in the generated argument list.
No external services required.
