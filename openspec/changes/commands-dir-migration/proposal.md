## Why

OpenCode documentation canonicalizes `.opencode/commands/`
(plural) as the standard directory for slash commands.
OpenSpec already switched to this path (Fission-AI/OpenSpec
#748, PRs #953 and #760). However, `uf init` still deploys
command files to `.opencode/command/` (singular), creating a
split-directory situation where OpenSpec commands land in
`commands/` and all other commands land in `command/`.

OpenCode's runtime loads both directories via a
`{command,commands}` glob, so nothing is functionally broken.
But the inconsistency is confusing for users and produces
divergent directory layouts across repos.

## What Changes

Migrate the `uf init` scaffold engine and associated tooling
from `.opencode/command/` (singular) to `.opencode/commands/`
(plural), and add a migration mechanism for existing repos.

## Capabilities

### New Capabilities
- `migrateCommandDir()`: Automatic migration of existing
  `.opencode/command/` directories to `.opencode/commands/`
  during `uf init`. Handles rename (old-only), merge
  (both exist), symlink detection, and partial failure
  recovery. Idempotent and re-runnable.
- `/uf-init` Step 0: Command directory migration step in
  the slash command, providing the same migration when run
  inside OpenCode without the Go binary.
- Doctor legacy warning: `uf doctor` warns when the legacy
  `.opencode/command/` directory is detected, guiding users
  to run `uf init` to migrate.

### Modified Capabilities
- `uf init` scaffold engine: Deploys embedded command assets
  to `.opencode/commands/` instead of `.opencode/command/`.
  All path-mapping, tool-ownership, and Divisor-asset
  functions updated.
- `uf doctor`: Checks `.opencode/commands/` as the primary
  command directory.
- `validate-hero-contract.sh`: Accepts `.opencode/commands/`
  as the canonical path, warns on legacy `.opencode/command/`.
- `/uf-init` command: All ~15 path references updated from
  `command/` to `commands/`.
- Live command and agent Markdown files: Self-references
  updated to use the new path.

### Removed Capabilities
- None. The legacy `command/` path continues to be loaded
  by OpenCode's runtime. This change affects only what
  `uf init` produces and where it expects files.

## Impact

### Files Modified

**Go source** (production):
- `internal/scaffold/scaffold.go` -- `isToolOwned()`,
  `isDivisorAsset()`, new `migrateCommandDir()`
- `internal/doctor/checks.go` -- primary check path,
  legacy warning

**Embedded assets** (directory rename + content):
- `internal/scaffold/assets/opencode/command/` renamed to
  `internal/scaffold/assets/opencode/commands/` (8 files)
- Self-references within those 8 files updated

**Tests**:
- `internal/scaffold/scaffold_test.go` -- ~50+ path refs
- `internal/doctor/doctor_test.go` -- ~5 path refs
- ~11 new test functions for migration logic

**Live Markdown** (this repo's working copy):
- `.opencode/command/uf-init.md` -- ~15 refs + new Step 0
- `.opencode/command/unleash.md` -- ~4 refs
- `.opencode/command/review-council.md` -- ~1 ref
- `.opencode/command/cobalt-crush.md` -- ~2 refs
- `.opencode/agents/divisor-curator.md` -- 1 ref

**Project docs**:
- `AGENTS.md` -- ~10 refs (project structure, descriptions)
- `scripts/validate-hero-contract.sh` -- 4 refs

**NOT modified** (historical):
- Completed specs under `specs/` -- preserved as-is
- Archived OpenSpec changes -- preserved as-is

### Cross-Repo Impact
- Hero Interface Contract (Spec 002) validation script
  updated to accept `commands/` as canonical.
- Upstream issues to file: `unbound-force/gaze` and
  `unbound-force/replicator` for their `init` commands
  (separate changes, not blocked by this work).
- Speckit (`specify init`) is upstream and out of scope.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

This change does not alter artifact-based communication.
Command files are self-describing Markdown with YAML
frontmatter. The directory rename does not affect how
heroes produce or consume artifacts. The migration
function produces a `subToolResult` that is included in
the scaffold summary output, maintaining observability.

### II. Composability First

**Assessment**: PASS

The migration is tolerant of missing tools. If `specify
init` or `gaze init` are not installed, `uf init` still
deploys its own 8 command files to the new path. The
migration function gracefully handles files created by
any tool -- it moves everything without requiring
knowledge of which tool created which file. OpenCode
continues to load both directory names, so partial
migration is functionally safe.

### III. Observable Quality

**Assessment**: PASS

The migration reports its actions via `subToolResult`
(migrated count, skipped duplicates, warnings). Conflict
detection warns the user and suggests `/uf-init` for
AI-assisted resolution. Doctor checks produce actionable
warnings for the legacy directory. All output follows
the existing symbol conventions.

### IV. Testability

**Assessment**: PASS

The `migrateCommandDir()` function follows the established
injectable-dependency pattern (`opts.ReadFile`,
`opts.WriteFile`). All edge cases (rename, merge, symlink,
partial failure, idempotency, DivisorOnly skip) are
testable in isolation using `t.TempDir()`. No external
services or shared mutable state required.
