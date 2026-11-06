## Context

The proposal establishes a `uf gate` subcommand that enforces
constitution compliance at each workflow phase. This design
describes the technical approach, following the existing
`doctor` package as the structural pattern.

Today, the `doctor` command checks environment health (tool
availability, config validity). The `gate` command checks
workflow health (artifact presence, phase discipline). Both
produce structured results with pass/fail semantics and JSON
output, but they answer different questions:

- **doctor**: "Is my environment ready to work?"
- **gate**: "Am I allowed to proceed to this phase?"

## Goals / Non-Goals

### Goals
- Headless, non-interactive gate command with JSON output
- Phase-specific checks derived from the constitution
- Exit code 0 (pass) / 1 (fail) / 2 (internal error)
  for scripting
- Testable in isolation using `t.TempDir()` scaffolded trees
- Follows existing CLI patterns (params struct, `runGate`,
  `newGateCmd`)
- Usable as `validation_loop.script` in FullSend harnesses
- Runnable as local pre-commit or pre-PR hook

### Non-Goals
- Interactive prompts or remediation suggestions
- Modifying any artifacts or git state (read-only)
- Checking tool availability (that is `doctor`'s job)
- Enforcing code quality metrics (that is CI/Gaze's job)
- Wiring into FullSend harnesses (Tracked: #869)

## Decisions

### D1: Follow the doctor package pattern

The `internal/gate/` package mirrors `internal/doctor/`:

```
internal/gate/
  gate.go      -- Options, Result, Run()
  checks.go    -- per-phase check functions
  models.go    -- CheckResult, GateReport, Summary
  format.go    -- FormatJSON, FormatText
  gate_test.go -- tests with t.TempDir() scaffolded trees
```

The CLI layer in `cmd/unbound-force/gate.go` follows the
same params-struct pattern as doctor:

```go
type gateParams struct {
    targetDir string
    phase     string
    format    string
    stdout    io.Writer
    stderr    io.Writer
}

func runGate(p gateParams) error { ... }
func newGateCmd() *cobra.Command { ... }
```

**Model divergence note**: The gate `CheckResult` struct
(Name, Passed, Message) is intentionally simpler than the
doctor's `CheckResult` (Name, Severity, Message, Detail,
InstallHint, InstallURL). Gate checks are binary pass/fail
— there is no warn severity because constitution
compliance is not gradual. The doctor's `InstallHint` and
`InstallURL` fields are irrelevant to artifact checks.
A shared interface may be appropriate for future convergence
but is not needed for v1.

**Rationale**: Consistent CLI patterns reduce cognitive load
and make the codebase predictable. The doctor pattern is
proven and tested. Constitution IV (Testability) requires
`io.Writer` injection.

### D2: Phase enum with check registry

Phases are a string enum validated at the CLI layer:

```go
var validPhases = []string{
    "specify", "plan", "implement", "review", "pr",
}
```

Each phase maps to a slice of check functions:

```go
type CheckFunc func(ctx runContext) CheckResult

var phaseChecks = map[string][]CheckFunc{
    "specify":   {checkSpecExists},
    "plan":      {checkSpecExists, checkPlanExists},
    "implement": {checkSpecExists, checkPlanExists,
                  checkTasksExist, checkCoverageStrategy},
    "review":    {checkTasksComplete, checkReviewRun},
    "pr":        {checkReviewPass, checkNotOnMain},
}
```

Checks are cumulative — later phases include earlier checks
where appropriate. Each check function is independently
testable. The run context carries the project directory plus
the change resolution computed once per run, so every check
in a run evaluates a single branch snapshot and report
provenance matches the resolution the checks used.

**Phase simplification note**: The gate's five phases
(`specify`, `plan`, `implement`, `review`, `pr`) are
workflow milestones, not 1:1 mappings to the constitution's
full pipeline phases (`specify`, `clarify`, `plan`, `tasks`,
`analyze`, `checklist`, `implement`, `review`). The gate
groups related phases at enforcement boundaries — points
where a blocking check adds value. `clarify`, `tasks`,
`analyze`, and `checklist` are subsumed by the `plan` and
`implement` milestones. This simplification keeps the
gate's phase list manageable while covering all
constitution MUST rules.

**Rationale**: Constitution I (Autonomous Collaboration) —
check functions operate on filesystem artifacts without
runtime coupling. Constitution IV (Testability) — each
check is a pure function testable in isolation.

### D3: Artifact detection strategy

The gate command must find spec artifacts regardless of
whether they were created by Speckit or OpenSpec:

1. **OpenSpec**: Look for `openspec/changes/*/` with
   `.openspec.yaml` and the relevant artifact files
   (proposal.md, design.md, tasks.md).
2. **Speckit**: Look for `specs/NNN-*/` directories with
   spec.md, plan.md, tasks.md.

The check functions probe both locations and pass if
artifacts are found in either. The `--dir` flag sets the
project root for all lookups.

**Rationale**: Constitution II (Composability First) — the
gate works regardless of which workflow tool produced the
artifacts.

### D4: JSON output schema

The JSON output follows this structure, including
provenance metadata per Constitution III:

```json
{
  "version": "1.0.0",
  "tool_version": "dev",
  "producer": "uf-gate",
  "timestamp": "2026-09-08T12:00:00Z",
  "branch": "opsx/uf-gate",
  "commit": "1c79bf2452490bd73b8fcb4ce000991fefcbb7b0",
  "phase": "implement",
  "passed": false,
  "checks": [
    {
      "name": "spec-exists",
      "passed": true,
      "message": "Spec found at openspec/changes/uf-gate/proposal.md"
    },
    {
      "name": "coverage-strategy",
      "passed": false,
      "message": "No coverage strategy found in plan or tasks"
    }
  ],
  "summary": {
    "total": 4,
    "passed": 3,
    "failed": 1
  }
}
```

The `version` field tracks the schema version for backward
compatibility. The `producer`, `timestamp`, `branch`,
`commit`, and `tool_version` fields satisfy Constitution
III's provenance requirement. The `branch` and `commit`
fields are best-effort — empty string when git is
unavailable or the directory is not a git repository.

**Rationale**: Constitution III (Observable Quality) — JSON
format with provenance metadata is required. The schema is
flat and stable, suitable for machine consumption by
FullSend harnesses.

### D5: Git state checks use exec, not library

The `checkNotOnMain` function shells out to `git` rather
than importing a Go git library:

```go
func checkNotOnMain(ctx runContext) CheckResult {
    // Consumes the once-per-run branch resolution. detectBranch
    // shells out via execGit — explicit argument separation
    // (no sh -c), bounded timeout — and resolveChange carries
    // the result, including the CI-ref recovery for detached
    // HEADs, on the run context (design D2).
}
```

**Error handling**: When git exec fails (git not installed,
not a git repository, unexpected output), the branch
detection error is carried on the run context and the check
returns `CheckResult{Passed: false}` with a descriptive
message. Detached HEAD (`git rev-parse` returns "HEAD") is
treated as passing — it is not `main` — unless the CI-ref
recovery (risk R1) resolves it to `main`, in which case the
check fails.

**Rationale**: Constitution V (Security by Default) —
minimal dependencies. `exec.Command` with explicit argument
separation prevents shell injection. The `unbound-force`
CLI already assumes git is available. Adding a git library
would increase attack surface for a single branch-name
check. Constitution II (Composability) — no new
dependencies.

### D6: Text output for human use

When `--format text` (default), the gate command produces
colored terminal output similar to doctor:

```
Gate: implement

  ✓ spec-exists
  ✓ plan-exists
  ✓ tasks-exist
  ✗ coverage-strategy — No coverage strategy found

FAIL: 1 of 4 checks failed
```

This supports human use in local hooks and interactive
debugging.

### D7: Output channels

Check results (JSON or text format) MUST be written to
stdout. Validation errors and internal errors MUST be
written to stderr. This separation ensures that FullSend
harnesses can reliably parse JSON from stdout without
corruption from error messages.

## Risks / Trade-offs

### R1: Artifact location heuristics may miss edge cases

Resolution is branch-aware: when the current branch maps
to a change (`opsx/<name>` -> `openspec/changes/<name>/`,
`NNN-<name>` -> `specs/<name>/`), checks validate that
change only, so an unrelated change cannot mask a missing
artifact for the active change. `quick/<name>` branches
(the /uf.unleash quick tier) resolve strictly to no change —
the quick tier carries no spec artifacts, so unrelated
changes must not satisfy its checks either. On branches
without a change mapping, resolution falls back to one
first-alphabetical change directory for the whole run —
`openspec/changes/` is probed first and `specs/` only when
that layout holds no change directory — artifacts of
different changes are never combined to
satisfy a phase, and symlinked change directories are
followed so fallback matches strict resolution. A detached HEAD —
the default checkout state on CI runners — reports the
literal branch `HEAD` and would land in that fallback; the
resolver first recovers the branch from CI-provided refs
(`GITHUB_HEAD_REF`, then `GITHUB_REF_NAME`) so headless CI
runs keep strict branch-keyed resolution. Both paths can
miss artifacts in non-standard locations or custom
schemas. Mitigated by testing both detection paths and
documenting expected artifact locations in the command
help text.

### R2: Git exec dependency

Shelling out to `git` means the command fails if git is
not installed. This is acceptable because `uf` already
requires git for all workflows, and `doctor` checks for
git availability. When git is unavailable, the
`checkNotOnMain` check fails gracefully with a message
suggesting `uf doctor`.

### R3: Check granularity may need tuning

The initial check set is derived from the constitution's
explicit MUST rules. FullSend integration may reveal
additional checks needed. The check-registry pattern makes
adding new checks straightforward without changing the
architecture.

### R4: Coverage strategy detection is heuristic

Detecting "coverage strategy present" requires scanning
plan/tasks content for specific patterns: a heading
containing "coverage" (case-insensitive) that does not
negate it, body text containing "coverage target",
"unit test", "integration test", or "e2e test", or a
line matching `coverage.*%`. Signals that deny the
strategy do not count — negated headings ("No coverage
strategy"), negated body text ("No unit tests are
planned"), and measured zero-percent lines ("current
coverage: 0%") — and signals inside fenced code blocks
are documentation, not strategy, using the CommonMark
fence matching GATE-FR-007 applies to task lines. False
negatives remain possible for non-standard phrasing.
Mitigated by documenting the expected keywords and
exclusions in the command help text and testing with
representative examples.

### R5: Review markers are presence-of-text evidence

The `review` and `pr` phases verify review completion by
searching the active change's `tasks.md` for the literal
markers `<!-- spec-review: passed -->` and
`<!-- code-review: passed -->` (GATE-FR-008). This evidence
is only as strong as write access to `tasks.md`: a
committer can type the marker without running a review,
and the gate cannot distinguish the two. The mechanism is
what the spec mandates and matches the org's marker
convention, but consumers of `uf gate --phase pr` in CI
should not assume a stronger guarantee than "the marker is
present". Binding review checks to provenance-bearing
dispatch records (`.uf/artifacts/dispatch/`) is the
documented follow-up direction.
