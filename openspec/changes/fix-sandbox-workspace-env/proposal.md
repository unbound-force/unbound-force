## Why

PR #119 (`sandbox-parent-mount`) mounts the project's
parent directory at `/workspace` and sets `--workdir
/workspace/<basename>` so OpenCode starts in the
correct project subdirectory. However, the container
image's `entrypoint.sh` runs `cd "$WORKSPACE"` where
`WORKSPACE` defaults to `/workspace`. This overrides
the `--workdir` setting, causing OpenCode to start in
the parent directory instead of the project.

For example, when running `uf sandbox start` from
`/Users/j/Projects/org/gcal-organizer`:

- **Expected**: OpenCode CWD = `/workspace/gcal-organizer`
- **Actual**: OpenCode CWD = `/workspace` (the parent)

## What Changes

Pass `-e WORKSPACE=/workspace/<basename>` in
`buildRunArgs()` when parent mount is active. The
container's `entrypoint.sh` already reads the
`WORKSPACE` environment variable and `cd`s to it,
so no image changes are needed.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `buildRunArgs()`: When parent mount is active, passes
  `-e WORKSPACE=/workspace/<basename>` to the container
  so the entrypoint `cd`s to the correct project
  subdirectory.

### Removed Capabilities
- None.

## Impact

- `internal/sandbox/config.go`: Add one `-e` flag pair
  to `buildRunArgs()` when `useParentMount()` is true.
- `internal/sandbox/sandbox_test.go`: Update existing
  parent-mount tests to verify `WORKSPACE` env var,
  add test for `--no-parent` case (no `WORKSPACE`).

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

No inter-hero communication affected. This is a
container configuration fix.

### II. Composability First

**Assessment**: PASS

The sandbox remains independently usable. The fix is
backward compatible — `--no-parent` mode does not set
`WORKSPACE` (entrypoint uses its default `/workspace`).

### III. Observable Quality

**Assessment**: PASS

No output format changes.

### IV. Testability

**Assessment**: PASS

The change is testable via `buildRunArgs()` output
inspection — same pattern as existing tests. No
external services required.
