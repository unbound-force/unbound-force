## Why

`dewey init` and `dewey index` are repo-specific
operations (they initialize and index *this* repo's
content), but they currently run in both `uf setup`
(steps 13-14) and `uf init` (`initSubTools()`). This
creates duplication: during `uf setup`, dewey init/index
run twice -- once directly and once indirectly via
`uf init` at the final step (though init skips because
`.dewey/` already exists).

The established principle (from Spec 017) is:
- `uf setup` = system-level (install binaries)
- `uf init` = repo-level (configure this repo)

Dewey workspace initialization is repo-level work and
should live exclusively in `uf init`.

Additionally, `uf init` currently skips dewey index
entirely on re-runs (when `.dewey/` exists). After
adding new files to the repo, the index goes stale
with no way to refresh it via `uf init`.

## What Changes

1. Remove `initDewey()` and `indexDewey()` functions
   from `uf setup` (steps 13-14)
2. Renumber setup steps from 15 to 13
3. Add force re-index to `uf init`: when `.dewey/`
   already exists and `--force` is set, re-run
   `dewey index` to refresh the search index

## Capabilities

### New Capabilities
- `init-force-reindex`: `uf init --force` re-indexes
  Dewey sources when `.dewey/` already exists

### Modified Capabilities
- `setup-step-count`: Setup reduces from 15 to 13
  steps (dewey init/index removed, delegated to init)
- `init-dewey-ownership`: `uf init` is the sole owner
  of dewey workspace initialization and indexing

### Removed Capabilities
- `setup-dewey-init`: `uf setup` no longer directly
  runs `dewey init`
- `setup-dewey-index`: `uf setup` no longer directly
  runs `dewey index`

## Impact

- **Files modified**: `internal/setup/setup.go`,
  `internal/setup/setup_test.go`,
  `internal/scaffold/scaffold.go`,
  `internal/scaffold/scaffold_test.go`
- **No new files**: Pure refactoring + small feature
  (force re-index)
- **Net code reduction**: ~45 lines removed from setup,
  ~10 lines added to scaffold
- **User-visible behavior**: `uf setup` has 13 steps
  instead of 15. `uf init --force` now re-indexes
  Dewey. Dewey workspace is still initialized during
  `uf setup` (via init at the final step).

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

No changes to artifact-based communication or
inter-hero interfaces. Refactoring of CLI tool
internals.

### II. Composability First

**Assessment**: PASS

Dewey remains optional -- `uf init` only runs
dewey init/index when the dewey binary is available.
The change strengthens composability by consolidating
repo-specific initialization in the repo-specific
command.

### III. Observable Quality

**Assessment**: PASS

`uf init` reports dewey status via `subToolResult`
(initialized, completed, skipped, failed). Progress
messages ("Initializing Dewey workspace...",
"Indexing Dewey sources...") provide real-time
feedback. `uf setup` continues to report dewey binary
installation status.

### IV. Testability

**Assessment**: PASS

All dewey operations use injectable `ExecCmd` and
`LookPath` for test isolation. New force-reindex
behavior is tested via `t.TempDir()` with injected
dependencies. Setup tests are simplified (fewer
mocked steps).
