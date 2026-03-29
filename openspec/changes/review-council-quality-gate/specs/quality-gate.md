## ADDED Requirements

### Requirement: ci-hard-gate

In Code Review Mode, the `/review-council` command MUST
execute the CI workflow commands locally before invoking
any Divisor agent. If any CI command fails, the command
MUST stop and report the failure. It MUST NOT proceed
to Divisor agent delegation.

#### Scenario: CI passes, proceed to review
- **GIVEN** the user runs `/review-council` in Code
  Review Mode
- **WHEN** all CI commands (build, vet, test) pass
- **THEN** the command proceeds to Phase 1b and then
  to Divisor agent delegation

#### Scenario: CI fails, hard stop
- **GIVEN** the user runs `/review-council` in Code
  Review Mode
- **WHEN** `go test` fails with 2 test failures
- **THEN** the command reports the test failures as
  CRITICAL findings and stops without invoking any
  Divisor agent

### Requirement: ci-commands-from-workflow

The CI commands executed in Phase 1a MUST be derived
from the `.github/workflows/` directory. The command
MUST NOT use a hardcoded list of commands.

#### Scenario: derive commands from workflow
- **GIVEN** `.github/workflows/test.yml` defines
  `go build ./...`, `go vet ./...`, and
  `go test -race -count=1 ./...`
- **WHEN** Phase 1a executes
- **THEN** those exact commands are run locally in the
  same order

### Requirement: gaze-conditional-integration

In Code Review Mode, the `/review-council` command
SHOULD invoke the `gaze-reporter` agent in `full` mode
if the `gaze` binary is available. If `gaze` is not
available, the command MUST proceed without it and
note the absence as informational.

#### Scenario: Gaze available, run quality analysis
- **GIVEN** `which gaze` succeeds
- **WHEN** Phase 1a CI checks pass
- **THEN** the `gaze-reporter` agent is invoked in
  `full` mode and its output is captured for step 2

#### Scenario: Gaze not available, skip gracefully
- **GIVEN** `which gaze` fails (not installed)
- **WHEN** Phase 1a CI checks pass
- **THEN** the command proceeds to step 2 with a note:
  "Gaze not installed -- skipping quality analysis.
  Install with `brew install unbound-force/tap/gaze`."

### Requirement: gaze-context-to-divisor-agents

When Gaze quality data is available from Phase 1b, the
`/review-council` command MUST include the Gaze report
as additional context in each Divisor agent's review
prompt.

#### Scenario: Divisor agents receive Gaze context
- **GIVEN** Phase 1b produced a Gaze quality report
- **WHEN** the command delegates to Divisor agents in
  step 2
- **THEN** each agent's prompt includes a "Quality
  Context" section containing the Gaze report summary

#### Scenario: No Gaze data, agents review without it
- **GIVEN** Phase 1b was skipped (Gaze not installed)
- **WHEN** the command delegates to Divisor agents
- **THEN** agents receive their standard prompt without
  a "Quality Context" section

### Requirement: spec-review-mode-unchanged

Phase 1a (CI checks) and Phase 1b (Gaze analysis)
MUST NOT execute in Spec Review Mode. Spec Review Mode
MUST remain unchanged.

#### Scenario: Spec review skips CI and Gaze
- **GIVEN** the user runs `/review-council specs`
- **WHEN** the command enters Spec Review Mode
- **THEN** neither CI commands nor Gaze analysis are
  executed

## MODIFIED Requirements

### Requirement: code-review-step-1

Code Review Mode step 1 is restructured from a single
paragraph into two explicit sub-phases (1a and 1b)
with clear sequencing and hard-stop semantics.

Previously: "Replicate CI checks locally before
delegating to council agents. Read
`.github/workflows/` to identify the exact commands
CI runs, then execute those same commands. Any failure
is a CRITICAL finding that must be fixed before the
council review begins."

### Requirement: code-review-step-2

Code Review Mode step 2 is modified to include Gaze
report data (when available) in each Divisor agent's
review prompt as a "Quality Context" section.

Previously: Divisor agents received only the focus
area from the Known Reviewer Roles reference table.

## REMOVED Requirements

None.
