## ADDED Requirements

### Requirement: init-force-reindex

`uf init --force` MUST re-run `dewey index` when the
`.dewey/` directory already exists, refreshing the
search index with current repo content.

#### Scenario: force re-index on existing workspace
- **GIVEN** `.dewey/` already exists and `dewey` is
  in PATH
- **WHEN** the user runs `uf init --force`
- **THEN** `dewey index` is executed with progress
  message `"Re-indexing Dewey sources..."` and the
  result is reported as `"re-indexed"`

#### Scenario: no force, existing workspace skipped
- **GIVEN** `.dewey/` already exists
- **WHEN** the user runs `uf init` (no `--force`)
- **THEN** dewey init and index are both skipped
  silently (existing behavior preserved)

## MODIFIED Requirements

### Requirement: setup-step-count

`uf setup` MUST have 13 steps (down from 15). Steps
13-14 (`initDewey`, `indexDewey`) are removed. The
final step (`runUnboundInit`, previously step 15)
becomes step 13.

Previously: `uf setup` had 15 steps with dewey
init/index at steps 13-14.

#### Scenario: setup step count
- **GIVEN** a fresh environment with all tools
  available
- **WHEN** the user runs `uf setup`
- **THEN** progress messages show `[1/13]` through
  `[13/13]` with no dewey init or index steps

### Requirement: setup-no-dewey-workspace

`uf setup` MUST NOT directly run `dewey init` or
`dewey index`. Dewey workspace initialization is
delegated to `uf init` (called at setup's final step).

Previously: `uf setup` ran `dewey init` at step 13
and `dewey index` at step 14 before running `uf init`.

#### Scenario: setup delegates dewey to init
- **GIVEN** dewey is installed but `.dewey/` does not
  exist
- **WHEN** the user runs `uf setup`
- **THEN** `.dewey/` is created by `uf init` at the
  final step (step 13), not by setup directly

## REMOVED Requirements

### Requirement: setup-initDewey

The `initDewey()` function in `setup.go` is removed.
Dewey workspace creation is handled by `initSubTools()`
in `scaffold.go`.

### Requirement: setup-indexDewey

The `indexDewey()` function in `setup.go` is removed.
Dewey indexing is handled by `initSubTools()` in
`scaffold.go`.
