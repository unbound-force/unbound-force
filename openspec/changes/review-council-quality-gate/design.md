## Context

The `/review-council` command's Code Review Mode step 1
currently reads:

> "Replicate CI checks locally before delegating to
> council agents. Read `.github/workflows/` to identify
> the exact commands CI runs, then execute those same
> commands. Any failure is a CRITICAL finding that must
> be fixed before the council review begins."

This is a single paragraph that conflates instruction
with guidance. In practice, it's ambiguous about whether
failures are a hard stop or just a finding to report.
The Divisor Testing agent cannot run tests (`bash: false`)
and has no access to Gaze quality data. There are no
git hooks providing a local test gate.

## Goals / Non-Goals

### Goals
- Make the CI check a hard gate that stops the review
  council workflow on failure
- Integrate Gaze quality analysis into Code Review Mode
  when `gaze` is available
- Pass Gaze report data to Divisor agents as review
  context so they can reference concrete metrics
- Keep Gaze optional -- missing Gaze is informational,
  not blocking

### Non-Goals
- Modifying Divisor agent files (they stay read-only)
- Modifying the Gaze reporter agent
- Adding git hooks (separate concern)
- Modifying `/finale` (it watches remote CI as a
  second safety net)
- Adding quality gates to Spec Review Mode (specs don't
  compile or have test coverage)
- Making Gaze a hard requirement for the review council

## Decisions

### D1: Two-phase structure (1a + 1b)

Split step 1 into Phase 1a (CI, mandatory) and Phase 1b
(Gaze, conditional). This separates the hard gate
(build/test must pass) from the quality enrichment
(CRAP scores and metrics improve review quality but
aren't blocking).

Per Constitution Principle II (Composability First),
Gaze is independently installable and the review council
works without it.

### D2: CI commands from workflow files, not hardcoded

Phase 1a reads `.github/workflows/` to extract the
exact CI commands. This ensures the local gate matches
what CI actually runs, even if CI steps are added or
changed in the future. Per AGENTS.md: "Do not rely on
a memorized list of commands -- always derive them from
the workflow files, which are the source of truth."

### D3: Gaze invoked via gaze-reporter agent

Phase 1b invokes Gaze through the existing
`gaze-reporter` agent (subagent_type: `gaze-reporter`)
in `full` mode. This reuses the existing agent's
formatting, error handling, and graceful degradation
rather than running raw `gaze` commands. The agent's
output is captured and included in the Divisor agent
prompts.

Per Constitution Principle I (Autonomous Collaboration),
Gaze produces a self-describing report that the Divisor
agents consume as context input. No runtime coupling.

### D4: Gaze context enriches Divisor agent prompts

When Gaze results are available, they are appended to
each Divisor agent's review prompt as a "Quality
Context" section. This gives `divisor-testing` access
to actual CRAP scores, coverage percentages, and
quadrant distributions. Other Divisor agents benefit
too -- `divisor-architect` can reference complexity
metrics, `divisor-sre` can reference coverage ratchets.

### D5: Hard stop on CI failure, not just a finding

If `go build`, `go vet`, or `go test` fails in Phase
1a, the command MUST stop and report the failure. It
MUST NOT proceed to invoke Divisor agents. Rationale:
reviewing code that doesn't compile is wasted work.
The current wording ("must be fixed before the council
review begins") is rewritten as an explicit STOP
instruction.

## Risks / Trade-offs

### Risk: Local test execution adds latency

Running the full test suite locally before the review
council adds 30-120 seconds to the workflow. For this
project (16 packages, ~3s total), the latency is
minimal. For larger projects, this could be significant.

**Mitigation**: The test suite is fast in this project.
For future projects with slow tests, the CI commands
are derived from workflow files -- if the project uses
a fast subset for PR checks, that's what runs locally.

### Risk: Gaze report may be large

A full Gaze report on a large codebase could be
hundreds of lines. Passing the entire report as context
to 5 Divisor agents in parallel multiplies the token
usage.

**Mitigation**: The Gaze report is already formatted
as a concise summary with tables and one-line
interpretations (per the gaze-reporter agent contract).
The full report is passed once; each Divisor agent
receives the same copy. Token overhead is bounded by
the report size, not the codebase size.

### Trade-off: Gaze is optional, not mandatory

If Gaze isn't installed, the Divisor Testing agent
reviews test quality without access to actual CRAP
scores or coverage data. This is the current behavior
-- the change makes it strictly better when Gaze is
present, not worse when it's absent.
