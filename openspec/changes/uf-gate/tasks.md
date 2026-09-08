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

## 1. Models and types

- [x] 1.1 [P] Create `internal/gate/models.go` with
  `CheckResult` struct (Name, Passed, Message fields),
  `GateReport` struct (Version, Producer, Timestamp,
  Branch, Phase, Passed, Checks, Summary), and `Summary`
  struct (Total, Passed, Failed). Include JSON tags on
  all fields. Add provenance constants: `SchemaVersion`
  ("1.0.0") and `ProducerName` ("uf-gate"). (FR-003)

## 2. Check functions

- [x] 2.1 Create `internal/gate/checks.go` with check
  functions. Each function has signature
  `func(dir string) CheckResult`:
  - `checkSpecExists`: probe openspec
    `proposal.md` and speckit `spec.md` (FR-004)
  - `checkPlanExists`: probe openspec `design.md`
    and speckit `plan.md` (FR-005)
  - `checkTasksExist`: probe openspec `tasks.md`
    and speckit `tasks.md` (FR-006)
  - `checkCoverageStrategy`: scan plan/tasks content
    for coverage strategy keywords — heading containing
    "coverage" (case-insensitive), body containing
    "coverage target", "unit test", "integration test",
    "e2e test", or line matching `coverage.*%` (FR-006)
  - `checkTasksComplete`: scan tasks.md for `- [ ]`
    lines — if any unchecked boxes remain, check fails
    (FR-007)
  - `checkReviewRun`: look for `<!-- spec-review: passed -->`
    or `<!-- code-review: passed -->` marker in
    tasks.md (FR-007)
  - `checkReviewPass`: look for
    `<!-- code-review: passed -->` marker in
    tasks.md (FR-008)
  - `checkNotOnMain`: exec `git rev-parse
    --abbrev-ref HEAD` using `exec.Command` with
    explicit argument separation (no `sh -c`) and
    verify not `main`. Handle git unavailable and
    non-git directory with descriptive error messages.
    Detached HEAD passes (not `main`). (FR-008)
  Define `phaseChecks` map from phase string to
  `[]CheckFunc` slice. (FR-002)

## 3. Core gate logic

- [x] 3.1 Create `internal/gate/gate.go` with `Options`
  struct (TargetDir, Phase, Format, Stdout, Stderr),
  `Run()` function that validates phase, canonicalizes
  `TargetDir` via `filepath.Abs` + `filepath.EvalSymlinks`,
  validates it is an existing directory, executes check
  functions from the phase registry, populates provenance
  fields (version, producer, timestamp, branch), builds
  `GateReport`, returns error if any check fails.
  Exit code 0 (pass), 1 (check failure), 2 (internal
  error). (FR-002, FR-009, FR-010)

## 4. Output formatting

- [x] 4.1 [P] Create `internal/gate/format.go` with
  `FormatJSON(report, writer)` and
  `FormatText(report, writer)` functions following the
  doctor pattern. Text format uses checkmark/cross
  symbols. Check results go to stdout; validation and
  internal errors go to stderr. (FR-003, D7)

## 5. CLI wiring

- [x] 5.1 Create `cmd/unbound-force/gate.go` with
  `gateParams` struct (including stdout and stderr
  `io.Writer` fields), `runGate(params)` function, and
  `newGateCmd()` cobra constructor. Flags: `--phase`
  (required), `--format` (default "text", validated
  against "text"/"json"), `--dir` (default ".",
  validated as existing directory). Register in
  `main.go` with `root.AddCommand(newGateCmd())`.
  (FR-001, FR-002, FR-003, FR-010, FR-011)

## 6. Tests

- [x] 6.1 [P] Create `internal/gate/gate_test.go` with
  tests for each check function using `t.TempDir()`
  scaffolded artifact trees. Cover:
  - specify phase pass/fail with both openspec and
    speckit layouts (fixture: `openspec/changes/test/
    proposal.md` for openspec, `specs/001-test/spec.md`
    for speckit)
  - plan phase pass/fail
  - implement phase pass/fail including coverage
    strategy detection (positive: heading with
    "coverage", keyword "unit test", line with
    "coverage target: >= 80%"; negative: no keywords)
  - review phase pass/fail: tasks-complete (all `[x]`
    vs remaining `[ ]`), review-run (marker present
    vs absent)
  - pr phase pass/fail: review-pass (marker present
    vs absent), not-on-main (use `git init` +
    `git checkout -b` inside `t.TempDir()` for
    isolated git repository fixtures)
  - git error paths: non-git directory, detached HEAD
  Test `Run()` integration with multiple phases,
  including provenance field population. (FR-004
  through FR-011)
- [x] 6.2 [P] Create `cmd/unbound-force/gate_test.go`
  with tests for `runGate()` verifying JSON and text
  output, invalid phase rejection, missing phase flag,
  invalid format rejection, non-existent directory
  rejection, exit code 2 for internal errors.
  (FR-001, FR-002, FR-003, FR-009, FR-010)

## 7. Coverage strategy

- [x] 7.1 Document and verify coverage meets target.

Unit tests cover all check functions in isolation and
`Run()` integration. The CLI layer is tested via
`runGate()` with `io.Writer` injection. No integration
or e2e tests needed — the command is read-only
filesystem inspection. Git-dependent checks use
`git init` inside `t.TempDir()` for isolation.

Coverage target: >= 90% line coverage for
`internal/gate/` package. Justification: this is a
governance gate command where false negatives mean
constitution violations pass unchecked. The check
functions are pure functions with clear input/output
contracts, making high coverage both achievable and
appropriate for the risk level.

## 8. Documentation

- [ ] 8.1 Update AGENTS.md project structure to include
  `internal/gate/` entry.
- [ ] 8.2 Add CHANGELOG.md entry for the new `gate`
  subcommand.
- [ ] 8.3 [P] File documentation issue against the
  current repo for user-facing `uf gate` usage docs.

<!-- spec-review: passed -->
<!-- code-review: passed -->
