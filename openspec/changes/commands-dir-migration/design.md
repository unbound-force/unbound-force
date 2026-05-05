## Context

The `uf init` scaffold engine deploys slash command files
to `.opencode/command/` (singular). OpenCode documentation
and OpenSpec now use `.opencode/commands/` (plural). OpenCode
runtime supports both via a `{command,commands}` glob, so the
mismatch is cosmetic but produces confusing split-directory
layouts. See `proposal.md` for full motivation and
constitution alignment.

The scaffold engine (`internal/scaffold/scaffold.go`) uses
`embed.FS` at `internal/scaffold/assets/` with path
mapping via `mapAssetPath()`. Files under `opencode/command/`
are classified as tool-owned by `isToolOwned()` and
auto-updated on content diff. The `Run()` function
executes sub-tools (`specify init`, `openspec init`,
`gaze init`) after deploying embedded assets. Sub-tools
that have not been updated continue to write to the old
`command/` path.

## Goals / Non-Goals

### Goals
- `uf init` deploys command files to `.opencode/commands/`
- Existing repos with `.opencode/command/` are
  automatically migrated during `uf init`
- `/uf-init` slash command provides equivalent migration
  when run inside OpenCode without the Go binary
- `uf doctor` warns on legacy `.opencode/command/`
- `validate-hero-contract.sh` accepts either directory
  with a deprecation warning for the old path
- Migration is idempotent, re-runnable, and tolerant of
  partial failures

### Non-Goals
- Fixing `specify init` (Speckit upstream, separate change)
- Fixing `gaze init` (separate change in unbound-force/gaze)
- Fixing `replicator init` (separate change)
- Fixing `muti-mind` init commands (separate change)
- Modifying historical/archived specs

## Decisions

### D1: Rename embedded asset directory

Rename `internal/scaffold/assets/opencode/command/` to
`internal/scaffold/assets/opencode/commands/`. The generic
`mapAssetPath()` prefix mapping (`opencode/` -> `.opencode/`)
handles the new path without changes. `isToolOwned()` and
`isDivisorAsset()` need their string prefixes updated.

**Rationale**: Single source-of-truth change. The `embed.FS`
walk produces the new paths automatically.

### D2: Migration runs after sub-tools

`migrateCommandDir()` runs after `initSubTools()` in the
`Run()` execution order. This ensures files created by
`specify init` and `gaze init` in the old `command/`
directory are caught and moved to `commands/`.

Execution order:
1. Scaffold deploys embedded assets to `commands/` (new)
2. `initSubTools()` runs (`specify init` writes to old path)
3. `migrateCommandDir()` moves everything from old to new
4. `printSummary()` includes migration result

**Rationale**: Post-sub-tool migration is self-healing.
Every `uf init` run corrects the directory layout regardless
of upstream tool behavior.

### D3: Atomic rename when possible, merge when both exist

When only `.opencode/command/` exists (no `commands/`):
use `os.Rename()` for an atomic directory rename.

When both directories exist: per-file merge. Files unique
to the old directory are moved. Duplicate files with
identical content are removed from the old directory.
Duplicate files with different content: the `commands/`
version is kept (it was deployed by the scaffold moments
ago with the latest embedded content), the old copy is
removed, and a warning is printed suggesting `/uf-init`
for AI-assisted resolution.

**Rationale**: `os.Rename()` is fast and atomic on the
same filesystem. The merge path handles the common case
where scaffold deployed to `commands/` and sub-tools
created files in `command/`.

### D4: os.Rename with read+write+remove fallback

Individual file moves use `os.Rename()` first. On failure
(e.g., cross-device), fall back to read -> write -> remove.

**Rationale**: Most moves are same-filesystem (same
`.opencode/` parent). The fallback handles edge cases
without adding complexity to the common path.

### D5: Symlink detection and skip

If `.opencode/command` is a symlink (detected via
`os.Lstat()`), skip migration entirely with a warning.

**Rationale**: Users who created a symlink made a
deliberate filesystem choice. Automatically modifying
symlinks risks data loss in the linked target.

### D6: DivisorOnly mode skips migration

`uf init --divisor` is a subset deployment into another
repo. Migration is a project-level concern, not a Divisor
deployment concern. The migration function returns `nil`
(no `subToolResult`) in DivisorOnly mode.

**Rationale**: Renaming directories in a foreign repo
during a subset deployment is a side effect with
unexpected blast radius.

### D7: Silent no-op when nothing to migrate

When `.opencode/command/` does not exist, the function
returns `nil` -- no `subToolResult` is added to the
summary. This keeps the output clean after the first
successful migration.

**Rationale**: Most `uf init` runs after migration will
have nothing to report. Verbose no-op messages add noise.

### D8: Hero contract accepts either, warns on legacy

`validate-hero-contract.sh` checks for `.opencode/commands/`
first. If found, passes. If not found, checks for
`.opencode/command/`. If found, passes with a deprecation
warning. If neither, fails.

**Rationale**: Existing hero repos (gaze, website) have
not been migrated yet. A hard requirement would break
contract validation immediately. The warning nudges
migration without blocking.

### D9: /uf-init provides complementary migration

The `/uf-init` slash command gains a new Step 0 that
performs the same migration using bash tools (`mv`,
`rmdir`). This handles the case where a user runs
`/uf-init` inside OpenCode without first running the
Go binary. All subsequent steps use `commands/` paths.

**Rationale**: Defense in depth. Two independent migration
paths (Go binary + slash command) maximize the chance of
a clean migration regardless of the user's entry point.

### D10: Conflict warning references /uf-init

When a file exists in both directories with different
content, the warning message suggests running `/uf-init`
for AI-assisted resolution. The Go binary cannot
intelligently merge Markdown content, but the AI agent
can.

**Rationale**: Connects the deterministic tool (Go binary)
to the intelligent tool (AI agent) for cases that need
judgment.

## Risks / Trade-offs

### R1: Upstream tools re-create old directory

`specify init` and `gaze init` continue writing to
`command/`. Every `uf init` run migrates their output,
but users who run those tools independently between
`uf init` runs will have files in the old directory.
OpenCode loads both, so functionality is unaffected.

**Mitigation**: File upstream issues. The `uf init`
migration is a self-healing safety net.

### R2: Large mechanical diff

~480+ string references change from `command/` to
`commands/`. The diff will be large but is almost
entirely mechanical find-replace.

**Mitigation**: Task structure separates Go source
changes from Markdown changes for reviewability.
Scaffold drift tests validate asset synchronization.

### R3: Git history discontinuity

`git log --follow` may lose history across the directory
rename for individual files. Git detects renames by
content similarity, so identical-content moves are
tracked, but modified-and-moved files may not be.

**Mitigation**: This is a known git limitation. The
rename is a one-time event. Historical context is
preserved in this design document and the OpenSpec
change artifacts.

### R4: Race condition in /uf-init migration

If the user runs `/uf-init` while another process is
writing to `.opencode/command/`, files could be missed.

**Mitigation**: Extremely unlikely in practice. The
migration is idempotent and can be re-run.
