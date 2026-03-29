## Why

The `/review-council` Code Review Mode has a step 1 that
says "replicate CI checks locally" but the instruction is
a single vague paragraph. In practice, this step is easy
to skip or partially execute. More critically:

- The Divisor Testing agent (`divisor-testing`) has
  `bash: false` -- it audits test *quality* by reading
  files but **cannot run tests**. There is no automated
  local test execution gate before pushing.
- Gaze (the Tester hero) produces CRAP scores, quality
  metrics, and classification data, but is never invoked
  as part of the review council workflow. The Divisor
  Testing agent reviews test structure without access to
  actual quality metrics.
- There are no git hooks (`pre-push`, `pre-commit`) in
  the project. The only automated test gate is CI on
  the PR, which runs *after* pushing.

This means the review council can APPROVE code that
doesn't compile or pass tests.

## What Changes

Rewrite `/review-council` Code Review Mode step 1 into
an explicit two-phase hard gate:

**Phase 1a -- CI Checks (mandatory, hard gate):**
Read `.github/workflows/` to extract CI commands, execute
them locally. If any fail, STOP -- do not invoke Divisor
agents.

**Phase 1b -- Gaze Quality Analysis (conditional):**
If `gaze` is available, invoke the `gaze-reporter` agent
in `full` mode. Pass the Gaze report as additional
context to Divisor agents so they can reference concrete
CRAP scores and quality metrics in their review.

Update step 2 to pass Gaze results (when available) as
context to each Divisor agent, particularly
`divisor-testing`.

## Capabilities

### New Capabilities
- None (strengthens existing capability)

### Modified Capabilities
- `review-council-code-review`: Step 1 becomes an
  explicit two-phase hard gate with CI execution and
  conditional Gaze integration. Divisor agents receive
  Gaze quality data as review context.

### Removed Capabilities
- None

## Impact

- **Files modified**: `.opencode/command/review-council.md`
  and its scaffold asset copy
- **No Go code changes**: Markdown command file only
- **No agent file changes**: Divisor agents remain
  read-only (`bash: false`). Gaze reporter is invoked
  as-is via the Task tool.
- **Spec Review Mode**: Untouched -- CI and Gaze steps
  apply only to Code Review Mode
- **Workflow impact**: `/review-council` becomes a
  stronger local quality gate. The expected workflow
  is `/review-council` (local gate) then `/finale`
  (finalize and merge).

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

Gaze produces a self-describing quality report (JSON
output, provenance metadata). The Divisor agents consume
this report as context input, maintaining artifact-based
communication. No runtime coupling is introduced -- Gaze
runs independently and its output is passed as text.

### II. Composability First

**Assessment**: PASS

Gaze integration is conditional -- if `gaze` is not
installed, the review proceeds without it. The CI check
phase uses whatever commands are defined in the CI
workflow files, not a hardcoded list. The command works
in any repo with a `.github/workflows/` directory.

### III. Observable Quality

**Assessment**: PASS

This change directly strengthens observable quality by
ensuring quality claims (test passage, CRAP scores,
coverage metrics) are backed by automated, reproducible
evidence *before* the review council renders its verdict.

### IV. Testability

**Assessment**: N/A

This is a Markdown command file change. The scaffold
drift detection test verifies the file is properly
embedded. No new Go code is introduced.
