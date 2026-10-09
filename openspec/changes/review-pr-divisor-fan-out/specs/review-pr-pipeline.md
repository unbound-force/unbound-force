## ADDED Requirements

### Requirement: Multi-Agent Divisor Review Fan-Out

The `/uf.review-pr` command MUST delegate PR diff analysis to multiple
Divisor review agents via `plan_review_dispatch` and `invoke_agent`,
replacing the single `general` Task subagent.

#### Scenario: Standard PR review dispatches to Divisor agents
- **GIVEN** a GitHub PR with metadata, CI checks, and a diff
- **WHEN** `/uf.review-pr` reaches the review phase (after pre-flight
  and review-context)
- **THEN** the command MUST discover available `divisor-*` agents,
  load the dispatch-advisor skill, call `plan_review_dispatch` with
  mode `code` and the changed files list, and invoke each planned agent
  via `invoke_agent`

#### Scenario: Dispatch plan respects diff characteristics
- **GIVEN** a PR diff touching security-sensitive paths (e.g.,
  credential handling, input validation)
- **WHEN** `plan_review_dispatch` evaluates the change profile
- **THEN** the plan MUST include `divisor-adversary` with a
  security-focused prompt

### Requirement: Council Verdict Consolidation

The `/uf.review-pr` command MUST consolidate agent findings into a
structured council verdict using `finalize_review_dispatch` with the
`"review-council"` command enum.

#### Scenario: All agents approve
- **GIVEN** all `invoke_agent` runs return with no CRITICAL or HIGH
  findings
- **WHEN** the command consolidates results via
  `finalize_review_dispatch`
- **THEN** the output MUST include a verdict of APPROVE or APPROVE WITH
  ADVISORIES with per-agent findings and run provenance

#### Scenario: One or more agents request changes
- **GIVEN** at least one `invoke_agent` run returns CRITICAL findings
- **WHEN** the command consolidates results
- **THEN** the output MUST include a verdict of REQUEST CHANGES and
  list all findings with severity, file, and line references

### Requirement: Pre-Flight in Parent Before Dispatch

The `/uf.review-pr` command MUST run pre-flight validation in `ci-aware`
mode in the parent command before agent discovery and dispatch.

#### Scenario: CI checks pass
- **GIVEN** Step 3 fetched CI checks and all passed
- **WHEN** pre-flight runs in `ci-aware` mode
- **THEN** pre-flight MUST report PASS and the pipeline MUST continue
  to agent discovery

#### Scenario: CI checks have failures on changed code
- **GIVEN** Step 3 fetched CI checks and some failed on files in the
  PR diff
- **WHEN** pre-flight runs in `ci-aware` mode
- **THEN** pre-flight MUST report the failures and the output MUST
  include a warning, but the pipeline SHOULD continue to agent
  discovery

### Requirement: Review Context in Parent Before Dispatch

The `/uf.review-pr` command MUST run the review-context skill
(Protocols 1-4) in the parent command before agent dispatch, including
issue linking from the PR diff (Protocol 4).

#### Scenario: PR body references an issue
- **GIVEN** a PR whose body contains "Fixes #42"
- **WHEN** review-context Protocol 1 runs
- **THEN** the linked issue details MUST be included in the review
  context passed to each `invoke_agent` prompt

## MODIFIED Requirements

### Requirement: Diff Fetch Order

The `/uf.review-pr` command MUST fetch the PR diff before running
review-context, not inside the analysis subagent.

Previously: Diff was fetched inside the `general` Task subagent (old
Step B) after CI checks and review-context.

#### Scenario: Diff available for review-context Protocol 4
- **GIVEN** a PR with linked issues in commit messages
- **WHEN** the command fetches the diff in Step 3.6
- **THEN** review-context Protocol 4 MUST be able to scan the diff for
  issue references before agent dispatch

### Requirement: Convention Pack Loading in Parent

The `/uf.review-pr` command MUST load applicable convention packs
(default, go, typescript, severity, content) in the parent command and
include their content in each `invoke_agent` prompt.

Previously: Convention packs were loaded inside the `general` Task
subagent.

#### Scenario: Go code change
- **GIVEN** a PR that changes Go source files
- **WHEN** the command loads convention packs before dispatch
- **THEN** the go convention pack content MUST be included in every
  agent prompt alongside default and severity packs

## REMOVED Requirements

### Requirement: Single General Subagent for PR Analysis

Previously: `/uf.review-pr` delegated all review analysis to a single
`general` Task subagent executing a monolithic prompt (Steps A-F of
the old command). This single-agent pattern is replaced by multi-agent
Divisor fan-out.