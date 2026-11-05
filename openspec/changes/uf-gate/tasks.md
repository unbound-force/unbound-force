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
  ("1.0.0") and `ProducerName` ("uf-gate"). (GATE-FR-003)

## 2. Check functions

- [x] 2.1 Create `internal/gate/checks.go` with check
  functions. Each function has signature
  `func(ctx runContext) CheckResult`, accepting the project
  directory and the run's once-per-run change resolution
  (GATE-FR-011 as amended):
  - `checkSpecExists`: probe openspec
    `proposal.md` and speckit `spec.md` (GATE-FR-004)
  - `checkPlanExists`: probe openspec `design.md`
    and speckit `plan.md` (GATE-FR-005)
  - `checkTasksExist`: probe openspec `tasks.md`
    and speckit `tasks.md` (GATE-FR-006)
  - `checkCoverageStrategy`: scan plan/tasks content
    for coverage strategy keywords — heading containing
    "coverage" (case-insensitive, negated headings such
    as "No coverage strategy" excluded), body containing
    "coverage target", "unit test", "integration test",
    "e2e test", or line matching `coverage.*%` (GATE-FR-006)
  - `checkTasksComplete`: scan tasks.md for `- [ ]`
    lines outside fenced code blocks — if any unchecked
    boxes remain, check fails (GATE-FR-007)
  - `checkReviewRun`: look for `<!-- spec-review: passed -->`
    or `<!-- code-review: passed -->` marker in
    tasks.md (GATE-FR-007)
  - `checkReviewPass`: look for
    `<!-- code-review: passed -->` marker in
    tasks.md (GATE-FR-008)
  - `checkNotOnMain`: exec `git rev-parse
    --abbrev-ref HEAD` using `exec.Command` with
    explicit argument separation (no `sh -c`) and
    verify not `main`. Handle git unavailable and
    non-git directory with descriptive error messages.
    Detached HEAD passes (not `main`). (GATE-FR-008)
  Define `phaseChecks` map from phase string to
  `[]CheckFunc` slice. (GATE-FR-002)

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
  error). (GATE-FR-002, GATE-FR-009, GATE-FR-010)

## 4. Output formatting

- [x] 4.1 [P] Create `internal/gate/format.go` with
  `FormatJSON(report, writer)` and
  `FormatText(report, writer)` functions following the
  doctor pattern. Text format uses checkmark/cross
  symbols. Check results go to stdout; validation and
  internal errors go to stderr. (GATE-FR-003, D7)

## 5. CLI wiring

- [x] 5.1 Create `cmd/unbound-force/gate.go` with
  `gateParams` struct (including stdout and stderr
  `io.Writer` fields), `runGate(params)` function, and
  `newGateCmd()` cobra constructor. Flags: `--phase`
  (required), `--format` (default "text", validated
  against "text"/"json"), `--dir` (default ".",
  validated as existing directory). Register in
  `main.go` with `root.AddCommand(newGateCmd())`.
  (GATE-FR-001, GATE-FR-002, GATE-FR-003, GATE-FR-010, GATE-FR-011)

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
  including provenance field population. (GATE-FR-004
  through GATE-FR-011)
- [x] 6.2 [P] Create `cmd/unbound-force/gate_test.go`
  with tests for `runGate()` verifying JSON and text
  output, invalid phase rejection, missing phase flag,
  invalid format rejection, non-existent directory
  rejection, exit code 2 for internal errors.
  (GATE-FR-001, GATE-FR-002, GATE-FR-003, GATE-FR-009, GATE-FR-010)

## 7. Coverage strategy

- [x] 7.1 Document and verify coverage against target.

Unit tests cover all check functions in isolation and
`Run()` integration. The CLI layer is tested via
`runGate()` with `io.Writer` injection. No integration
or e2e tests needed — the command is read-only
filesystem inspection. Git-dependent checks use
`git init` inside `t.TempDir()` for isolation.

Coverage achieved: 92.7% statement coverage for
`internal/gate/` (measured with `go test -cover`
at the review-fix head, including the unmapped
fallback, negated-heading, fenced-example, and
CI-ref validation regression tests);
the lowest per-function figure outside the
deliberately-untriggered `init` panic guard is
80.0% (FormatJSON, whose marshal-error branch is
unreachable with the current report types).
Justification: this is a governance gate command
where false negatives mean constitution violations
pass unchecked. The check functions are pure
functions with clear input/output contracts, making
high coverage both achievable and appropriate for
the risk level. The enforced CI ratchet for
`internal/gate/` is the 80% helper-scope floor in
coverage-gate.json; the measured figure exceeds
that floor, and a regression below 80% fails the
CI coverage gate.

## 8. Documentation

- [x] 8.1 Update AGENTS.md project structure to include
  `internal/gate/` entry.
- [x] 8.2 Add CHANGELOG.md entry for the new `gate`
  subcommand.
- [x] 8.3 [P] File documentation issue against the
  current repo for user-facing `uf gate` usage docs.
  (Filed: #583; website tracking: unbound-force/website#343
  — the website CLI-reference update for `uf gate`, labeled
  `docs` and `blog` — per the constitution's Cross-Repo
  Documentation rule)

<!-- spec-review: passed -->
<!-- code-review: passed -->
