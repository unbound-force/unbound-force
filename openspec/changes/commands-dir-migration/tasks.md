<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Rename Embedded Asset Directory

- [x] 1.1 Rename `internal/scaffold/assets/opencode/command/`
  to `internal/scaffold/assets/opencode/commands/` (8 files:
  agent-brief.md, cobalt-crush.md, constitution-check.md,
  finale.md, review-council.md, review-pr.md, uf-init.md,
  unleash.md). Use `git mv` to preserve history.

## 2. Update Scaffold Engine

All changes in `internal/scaffold/scaffold.go`.

- [x] 2.1 Update `isToolOwned()`: change prefix check from
  `"opencode/command/"` to `"opencode/commands/"`.
- [x] 2.2 Update `isDivisorAsset()`: change match from
  `"opencode/command/review-council.md"` to
  `"opencode/commands/review-council.md"`.
- [x] 2.3 Add `moveFile()` helper function: `os.Rename()`
  first, on error fall back to read -> write -> remove
  using `opts.ReadFile`/`opts.WriteFile`.
- [x] 2.4 Add `migrateCommandDir(opts *Options)
  *subToolResult` function implementing design decisions
  D2-D7, D10: DivisorOnly guard, symlink detection via
  `os.Lstat()`, atomic rename when only old dir exists,
  per-file merge when both exist, duplicate detection
  (identical=remove old, different=keep new+warn
  mentioning `/uf-init`), empty dir removal, silent
  no-op return.
- [x] 2.5 Wire `migrateCommandDir()` into `Run()`: call
  after `initSubTools()` returns, append result to
  `subResults` slice (only if non-nil). Must run before
  `printSummary()`.

## 3. Update Embedded Asset Content

Update self-references from `.opencode/command/` to
`.opencode/commands/` inside embedded scaffold copies.
These are under `internal/scaffold/assets/opencode/commands/`
(after Group 1 rename).

- [x] 3.1 [P] Update `uf-init.md`: ~15 path references
  throughout Steps 1, 2, 5, 6, 8 + add new Step 0
  (Command Directory Migration) before current Step 1.
  Step 0 uses bash to detect `.opencode/command/`, rename
  or merge to `.opencode/commands/`, report results. All
  subsequent steps use `commands/` paths.
- [x] 3.2 [P] Update `unleash.md`: ~4 references to
  command file paths (speckit.plan.md, speckit.tasks.md,
  review-council.md, speckit.implement.md).
- [x] 3.3 [P] Update `review-council.md`: ~1 reference
  in file scope list (`cmd/, .opencode/agents/,
  .opencode/command/` → `.opencode/commands/`).
- [x] 3.4 [P] Update `cobalt-crush.md`: ~2 references to
  speckit.implement.md and opsx-apply.md.
- [x] 3.5 [P] Update `review-pr.md`: check for and update
  any `.opencode/command/` references.
- [x] 3.6 [P] Update `agent-brief.md`: check for and update
  any `.opencode/command/` references.

## 4. Sync Live Files with Embedded Assets

After Group 1 (directory rename) and Group 3 (content
updates), the live `.opencode/command/` directory must be
renamed and its contents synchronized.

- [x] 4.1 Rename live directory `.opencode/command/` to
  `.opencode/commands/` using `git mv`. This moves all
  46 command files (8 scaffold-deployed + 38 from other
  tools).
- [x] 4.2 Copy the 8 updated embedded assets from
  `internal/scaffold/assets/opencode/commands/` to
  `.opencode/commands/` to ensure byte-identical sync
  (scaffold drift test requirement). Verify with diff.

## 5. Update Non-Scaffold Command Files

Update self-references in live command files that are
NOT scaffold-deployed (created by other tools, not
embedded). These are under `.opencode/commands/` (after
Group 4 rename).

- [x] 5.1 Search all `.opencode/commands/*.md` files for
  remaining `.opencode/command/` references (singular).
  Update any found. Expected files: `opsx-propose.md`,
  `opsx-apply.md`, `opsx-archive.md`, `opsx-explore.md`,
  `speckit.*.md` (9 files), `gaze.md`, `gaze-fix.md`,
  `muti-mind.*.md` (12 files), `workflow-*.md` (5 files),
  `forge.md`, `forge-status.md`, `handoff.md`, `inbox.md`,
  `org.md`.

## 6. Update Agent Files

- [x] 6.1 [P] Update `.opencode/agents/divisor-curator.md`:
  change `.opencode/command/` reference to
  `.opencode/commands/` in the change detection heuristic.
- [x] 6.2 [P] Update scaffold copy
  `internal/scaffold/assets/opencode/agents/divisor-curator.md`:
  same change as 6.1. Verify byte-identical to live copy.

## 7. Update Doctor Checks

- [x] 7.1 [P] Update `internal/doctor/checks.go`:
  change `checkDirWithCount` path argument from
  `.opencode/command/` to `.opencode/commands/`. Add
  a legacy warning check: if `.opencode/command/` exists,
  append a warning result recommending `uf init`.
- [x] 7.2 [P] Update `internal/doctor/doctor_test.go`:
  change all `.opencode/command/` path strings to
  `.opencode/commands/` (~5 references). Add test for
  legacy directory warning.

## 8. Update Project Documentation

- [x] 8.1 [P] Update `AGENTS.md`: change project structure
  tree entry from `command/` to `commands/`, update
  command directory references throughout (~10 refs).
  Do NOT modify the Recent Changes entries for completed
  changes (historical accuracy).
- [x] 8.2 [P] Update `scripts/validate-hero-contract.sh`:
  change primary check to `.opencode/commands/`. Add
  fallback accepting `.opencode/command/` with a
  deprecation warning. Fail only when neither exists
  (~4 references).

## 9. Update Scaffold Tests

All changes in `internal/scaffold/scaffold_test.go`.

- [x] 9.1 Update `expectedAssetPaths`: change 8 entries
  from `opencode/command/` prefix to `opencode/commands/`.
- [x] 9.2 Update `knownNonEmbeddedFiles`: change all
  `.opencode/command/` entries to `.opencode/commands/`
  (~48 entries across Speckit, Gaze, Muti-Mind, OpenSpec,
  Workflow, Replicator, Dewey command files).
- [x] 9.3 Update test functions that reference command
  paths: `TestRun_CreatesFiles` (directory check),
  `TestCanonicalSources_AreEmbedded` (walk path),
  `TestMapAssetPath_Prefixes` (mapping test case),
  `TestIsDivisorAsset` (3 test cases), `TestIsToolOwned`
  (~10 test cases), `TestUFInitAsset_ContainsSpecificInstructions`
  (asset read path), and all Divisor/summary test helpers
  that create `.opencode/command/` paths (~15 functions).
- [x] 9.4 Add migration test functions (~12 tests):
  `TestMigrateCommandDir_NoOldDir` (silent no-op),
  `TestMigrateCommandDir_NeitherDir` (silent no-op),
  `TestMigrateCommandDir_RenameOnly` (atomic rename),
  `TestMigrateCommandDir_MergeUnique` (unique files moved),
  `TestMigrateCommandDir_MergeDupIdentical` (identical
  removed from old),
  `TestMigrateCommandDir_MergeDupDifferent` (commands/
  kept, warning printed),
  `TestMigrateCommandDir_Symlink` (skipped with warning),
  `TestMigrateCommandDir_DivisorOnly` (skipped, nil),
  `TestMigrateCommandDir_NonMDFiles` (non-.md left behind),
  `TestMigrateCommandDir_Idempotent` (re-run is silent),
  `TestMigrateCommandDir_PartialFailure` (partial move,
  warning, old dir kept),
  `TestMoveFile_FallbackOnRenameError` (injects Rename
  function that returns error, verifies read->write->remove
  fallback path executes correctly).

## 10. Build, Test, and Verify

- [x] 10.1 Run `make check` (build + test + vet + lint).
  Fix any failures.
- [x] 10.2 Verify scaffold drift: confirm all 8 embedded
  command assets are byte-identical to their live copies
  under `.opencode/commands/`. Run
  `TestCanonicalSources_AreEmbedded` specifically.
- [x] 10.3 Verify no stale references: search entire repo
  for `.opencode/command/` (singular, not followed by `s`).
  Exclude: completed specs under `specs/`, archived
  openspec changes, `go.sum`, and this change's own
  artifacts. All remaining hits must be intentional
  (e.g., migration function code referencing the old path).
- [x] 10.4 Constitution alignment verification: confirm
  migration function is testable in isolation (Principle
  IV), produces observable results via subToolResult
  (Principle III), and does not introduce mandatory
  cross-hero dependencies (Principle II).
- [x] 10.5 Website documentation check: search
  `unbound-force/website` for pages referencing
  `.opencode/command/`. If found, file a `docs:` GitHub
  issue via `gh issue create --repo unbound-force/website`.
  If no references found, skip (no issue needed).
<!-- spec-review: passed -->
<!-- code-review: passed -->
