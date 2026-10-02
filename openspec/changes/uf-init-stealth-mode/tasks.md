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

## 1. Scaffold engine (`internal/scaffold/scaffold.go`)

- [x] 1.1 Add `Stealth bool` and `Check bool` fields to `Options` with
  GoDoc explaining stealth scaffolds tool-owned files git-invisibly
  without modifying tracked files and check reports the expected tree
  state without persisting
- [x] 1.2 Add `writeGitInfoExclude(opts *Options) subToolResult` helper
  that appends the scaffolded paths (`.opencode/`, `.specify/`,
  `openspec/`, `.uf/`, `opencode.json`) to `.git/info/exclude`,
  idempotent via a marker comment, resolving the path through
  `git rev-parse --git-path info/exclude` (not a hardcoded
  `.git/info/exclude`), appending any newly-created untracked paths
  discovered during init (so the exclusion list is complete for all
  sub-tool output), and returning a clear error when the target is
  unwritable
- [x] 1.3 Add `snapshotTrackedFiles` / `restoreTrackedFiles` helpers
  capturing and restoring the content bytes AND mode/permission bits of
  ALL tracked files (permission bits via per-file `os.Lstat`, not
  `os.Stat` which follows symlinks; symlinks restored as symlinks), not
  just `.gitignore` and `AGENTS.md`, preserving permission bits on
  restore
- [x] 1.4 Wire `--stealth` through `Run`: verify inside a git repo via
  `git rev-parse --is-inside-work-tree` (error before any writes if
  not), detect already-tracked scaffolded paths and refuse with a clear
  error, skip `ensureGitignore` and `ensureAGENTSmdPackSection`, call
  `writeGitInfoExclude` BEFORE scaffolding, wrap `initSubTools`
  delegation with snapshot/restore, roll back scaffolded files +
  marker-delimited exclusion block + un-stage any paths staged during
  this init on any failure, warn when a stale marker-delimited stealth
  exclusion block is present during normal init, and extend
  `printSummary` to report working-tree cleanliness as JSON
- [x] 1.5 Add a `--check` path (stealth-only) that reports the expected
  tree state as JSON without persisting anything, and is re-runnable
  after init (combined with `--stealth`) to detect a missing/stale
  `.git/info/exclude` block

## 2. CLI wiring (`cmd/unbound-force/main.go`)

- [x] 2.1 Add `stealth` and `check` to `initParams`, map them into
  `scaffold.Options`, add `--stealth` and `--check` flags to
  `newInitCmd`, and update the command help text to document that
  stealth exclusion is local-only and not preserved on re-clone

## 3. Tests

Tests 3.1–3.14 all modify the new `scaffold_stealth_test.go`, so they
run sequentially to avoid merge conflicts. See design.md "Test
Strategy" for the unit vs integration classification and the ≥90%
statement-coverage target on the stealth paths.

- [x] 3.1 Add table-driven tests asserting stealth mode leaves
  `.gitignore` and `AGENTS.md` byte-identical while normal mode still
  modifies them (new `scaffold_stealth_test.go`)
- [x] 3.2 Add tests asserting `.git/info/exclude` is written with the
  scaffolded paths, the tracked `.gitignore` gains no new entry, a
  second stealth init appends no duplicate entries, and a newly-created
  path beginning with `!`/`#` is emitted as a literal (escaped) pattern
  and still excluded
- [x] 3.3 Add tests faking ONLY `ExecCmd` (to simulate a sub-tool
  writing `.gitignore`) while `ReadFile`/`WriteFile` remain the real
  `os` functions, asserting the on-disk `.gitignore` bytes AND
  mode/permission bits (on a non-default-mode file, e.g. `0755`/`0600`)
  are restored, and that a tracked symlink is restored as a symlink
  with the same link target
- [x] 3.4 Add a normal-mode regression test asserting non-stealth
  behavior is byte-identical to the current implementation
- [x] 3.5 Add tests asserting `--check` reports cleanliness (as JSON
  including `tool`/`version`/`timestamp` and, when available,
  `ref`/`commit`) without writing any file, exits non-zero when the
  tree is dirty or the exclusion is missing/stale and 0 when clean, and
  that `uf init --check` without `--stealth` is a clear error
- [x] 3.6 Add an integration test (real `git` binary in `t.TempDir()`)
  asserting `git status --porcelain` is empty after stealth init
- [x] 3.7 Add an integration test asserting a non-git directory errors
  non-zero with a message referencing the git requirement and writes no
  files; and a unit test (faked `ExecCmd` returning exit 127) asserting
  the git-not-installed path reports a distinct actionable error and
  writes no files
- [x] 3.8 Add an integration test injecting a mid-`Run` failure and
  asserting tracked files are byte-identical (bytes and mode), the
  index is restored (paths staged during init are un-staged while
  pre-existing unrelated staged changes survive), a pre-existing
  marker-delimited exclusion block survives while the current
  invocation's block is removed, `git status --porcelain` is empty, and
  a non-nil error is returned; plus a case asserting that a rollback-
  step failure reports a clear error naming residual dirty paths
- [x] 3.9 Add a test asserting scaffolded paths are present and readable
  at their standard locations after stealth init (pipeline operability)
- [x] 3.10 Add a test asserting already-tracked scaffolded paths are
  refused with a clear error naming the tracked path(s) and no files
  written, including under `uf init --stealth --force` (force MUST NOT
  bypass the refusal)
- [x] 3.11 Add a test asserting re-runnable `--check` detects a wiped
  `.git/info/exclude` and reports the now-visible scaffolded paths, and
  that a partially-edited block (one path removed) reports that
  specific path as now-visible
- [x] 3.12 Add tests asserting `writeGitInfoExclude` returns a clear
  error and leaves `git status --porcelain` empty when the exclusion
  target is unwritable (simulated portably via a faked `ExecCmd`
  resolving `--git-path info/exclude` to a path inside a non-writable
  or missing-parent directory, not `chmod 000` which is ineffective
  under root), and that the exclusion path is resolved via
  `git rev-parse --git-path info/exclude` (faked `ExecCmd` returning a
  non-default git-dir path)
- [x] 3.13 Add a table-driven test asserting the `--stealth` ×
  {`--force`, `--divisor`, `--lang`, `--dry-run`} compositions: for
  each, assert (a) the functional outcome (overwrite-untracked /
  divisor-subset / language-honored / `opencode.json`-suppressed) AND
  (b) the exclusion block is written and `git status --porcelain` is
  empty (for `--dry-run`, `opencode.json` suppressed while scaffolding
  remains hidden); plus that the `--stealth` help text documents the
  local-only non-portability (the `--check` case is covered by 3.5;
  non-stealth parity by 3.4)
- [x] 3.14 Add a test asserting normal `uf init` (without `--stealth`)
  warns when a stale marker-delimited stealth exclusion block is
  present in `.git/info/exclude`

## 4. Verification and documentation

- [x] 4.1 Verify constitution alignment (all five principles:
  Autonomous Collaboration, Composability First, Observable Quality,
  Testability, Security by Default) against the completed implementation
- [x] 4.2 File a cross-repo documentation issue in
  `unbound-force/website` (per constitution Cross-Repo Documentation)
  to track the new `--stealth` flag in the `uf init` docs
- [x] 4.3 Add a CHANGELOG.md entry for the new `--stealth` flag
- [x] 4.4 Run `make check` (lint, test with `-race -count=1`, build)
  and confirm the coverage ratchet is satisfied (≥90% statement
  coverage on the stealth paths) and local CI parity before marking
  complete

<!-- spec-review: passed -->
<!-- code-review: passed -->
