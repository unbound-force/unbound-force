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
- Wiring into FullSend harnesses (separate change)

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
type CheckFunc func(dir string) CheckResult

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
testable.

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
  "producer": "uf-gate",
  "timestamp": "2026-09-08T12:00:00Z",
  "branch": "opsx/uf-gate",
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
compatibility. The `producer`, `timestamp`, and `branch`
fields satisfy Constitution III's provenance requirement.
The `branch` field is best-effort — empty string when git
is unavailable or the directory is not a git repository.

**Rationale**: Constitution III (Observable Quality) — JSON
format with provenance metadata is required. The schema is
flat and stable, suitable for machine consumption by
FullSend harnesses.

### D5: Git state checks use exec, not library

The `checkNotOnMain` function shells out to `git` rather
than importing a Go git library:

```go
func checkNotOnMain(dir string) CheckResult {
    // exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
    // with dir set as the working directory
    // MUST use explicit argument separation (no sh -c)
}
```

**Error handling**: When git exec fails (git not installed,
not a git repository, unexpected output), the check returns
`CheckResult{Passed: false}` with a descriptive message.
Detached HEAD (`git rev-parse` returns "HEAD") is treated
as passing — it is not `main`.

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

The dual OpenSpec/Speckit detection could miss artifacts
in non-standard locations or custom schemas. Mitigated by
testing both detection paths and documenting expected
artifact locations in the command help text.

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
containing "coverage" (case-insensitive), body text
containing "coverage target", "unit test", "integration
test", or "e2e test", or a line matching `coverage.*%`.
False negatives are possible for non-standard phrasing.
Mitigated by documenting the expected keywords in the
command help text and testing with representative examples.
