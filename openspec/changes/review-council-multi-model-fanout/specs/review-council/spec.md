## ADDED Requirements

### Requirement: Review Council Argument Grammar

**ID: RC-FR-001** This requirement MUST be satisfied.

Review council MUST parse case-sensitive tokens in any order before
discovery. It MUST accept at most one exact mode token, `code` or
`specs`; at most one exact `--full` token; and at most one PR token.

A PR token MUST contain ASCII digits only and represent a value from 1
through 999999. Leading zeroes MUST be accepted and normalized before
lookup. One PR token MUST imply code mode. An explicit valid mode MUST
otherwise win, followed by existing auto-detection.

The command MUST reject `specs` with a PR token, multiple PR tokens,
repeated flags, conflicting modes, and unknown tokens before dispatch.
It MUST reject zero, signs, decimals, Unicode digits, and values above
999999. Token matching MUST remain case-sensitive.

#### Scenario: Full flag before PR number

- **GIVEN** arguments are `--full 42`
- **WHEN** parsing completes
- **THEN** PR 42 uses code mode
- **AND** full-panel mode is true

#### Scenario: Specs with PR is invalid

- **GIVEN** arguments are `specs 42`
- **WHEN** parsing begins
- **THEN** the command reports an argument error
- **AND** no reviewer discovery or dispatch occurs

#### Scenario: Leading zeroes normalize

- **GIVEN** arguments are `000042 --full code`
- **WHEN** parsing completes
- **THEN** PR 42 uses code mode
- **AND** full-panel mode is true

#### Scenario: Unicode digits are invalid

- **GIVEN** the PR token contains non-ASCII digit characters
- **WHEN** parsing begins
- **THEN** the command reports an argument error
- **AND** no reviewer discovery or dispatch occurs

#### Scenario: PR value outside the range

- **GIVEN** the PR token is `0` or `1000000`
- **WHEN** parsing begins
- **THEN** the command reports an argument error
- **AND** no reviewer discovery or dispatch occurs

### Requirement: Review Council Dispatch Plan

**ID: RC-FR-002** This requirement MUST be satisfied.

Code and spec modes MUST load `dispatch-advisor`, validate and display
its version 1 plan, and execute included runs. A PR invocation MUST
use resolved base-to-head context. Plan validation failure MUST produce
`INCONCLUSIVE` before child sessions start.

#### Scenario: Explicit adversary fan-out

- **GIVEN** the plan contains two ordered adversary runs
- **WHEN** review council delegates review
- **THEN** both runs execute within concurrency limits
- **AND** no unplanned run is created

### Requirement: Multi-Run Finding Consolidation

**ID: RC-FR-003** This requirement MUST be satisfied.

Review council MUST deduplicate findings with the same normalized file
and root cause. One consolidated finding MUST retain all contributing
agents, models, variants, sources, and sequences. Model-level
deduplication MUST precede existing cross-persona root-cause and
compound-severity processing.

#### Scenario: Same defect from two models

- **GIVEN** two runs report the same unchecked input
- **WHEN** findings are consolidated
- **THEN** one finding is emitted
- **AND** both run identities remain attached

#### Scenario: Different defects in one file

- **GIVEN** two runs report different root causes in one file
- **WHEN** findings are consolidated
- **THEN** both findings remain separate

### Requirement: Review Model Self-Report

**ID: RC-FR-004** This requirement MUST be satisfied.

Every delegated prompt MUST require `**Model**: <family>`. The final
report MUST retain authoritative requested model and optional variant
for each run and finding. A conflicting self-report MUST be noted and
MUST NOT overwrite request provenance.

For a host run, requested model and variant MUST remain null. The final
report MUST instead retain the resolved parent model and active variant
plus the reported child model. A self-report MUST NOT overwrite any of
those authoritative fields.

#### Scenario: Self-report mismatch

- **GIVEN** request provenance names one model
- **AND** response text self-reports another
- **WHEN** the report is rendered
- **THEN** request provenance remains authoritative
- **AND** the mismatch is recorded

## MODIFIED Requirements

### Requirement: Agent Discovery Step

**ID: RC-FR-005** This requirement MUST be satisfied.

The council MUST discover `.opencode/agents/divisor-*.md` before
planning. It MUST load reviewer metadata only from the closed
`.uf/reviewer-capabilities.yaml` version 1 manifest. The manifest MUST
contain one unique entry per known persona with `agent`, `capability`,
and ordered `scopes`. Capability MUST be `review` or `content`. Scope
values MUST come from `security`, `cli-ux`, `test-quality`,
`documentation`, `ci-cd`, `dependencies`, and `standard`.

The council MUST pass only discovered agents with valid `review`
entries to `plan_review_dispatch`. It MUST invoke only included plan
runs and MUST NOT hardcode the invocation list. Unknown personas MUST
remain eligible only through an explicit valid manifest entry with
intersecting scope. Content personas MUST be listed but never
dispatched. Missing, duplicate, conflicting, or invalid entries MUST
be reported and MUST NOT produce a run. Agent frontmatter MUST NOT
carry reviewer capability or scope metadata.

Previously, the command searched for `reviewer-*.md` and invoked every
reviewer once through the host-model Task tool.

#### Scenario: Subset of personas present

- **GIVEN** only four manifested review agents exist on disk
- **WHEN** discovery and planning run
- **THEN** only files declaring review capability are eligible
- **AND** absent known personas are informational

#### Scenario: Content agent is discovered

- **GIVEN** a discovered Divisor agent has content capability
- **WHEN** discovery and planning run
- **THEN** the agent appears in the discovery summary
- **AND** no review run is planned for it

#### Scenario: No personas found

- **GIVEN** no valid Divisor agent file exists
- **WHEN** discovery runs
- **THEN** the result is `INCONCLUSIVE`
- **AND** dispatch does not begin

### Requirement: Discovery Summary in Final Report

**ID: RC-FR-006** This requirement MUST be satisfied.

The final report MUST list discovered personas and their capabilities,
included plan runs, content exclusions, policy skips with reasons,
validation errors, limit skips, and absent known roles. Absence,
content exclusions, and policy skips MUST be informational. Too little
successful coverage MUST follow RD-FR-007 cause precedence and produce
`UNAVAILABLE` only for availability-only causes.

Previously, the summary listed invoked reviewers and absent roles from
five-role reference set.

#### Scenario: Pruned and absent personas

- **GIVEN** one persona is pruned and one role is absent
- **WHEN** the report is rendered
- **THEN** both statuses and reasons are distinct
- **AND** neither creates a finding

### Requirement: Known Reviewer Role Descriptions

**ID: RC-FR-007** This requirement MUST be satisfied.

The manifest and role table MUST document all nine personas. Their
ordered entries MUST be:

- adversary: `review`; security, dependencies, standard
- architect: `review`; standard, cli-ux, ci-cd, documentation
- curator: `review`; documentation
- guard: `review`; standard, cli-ux, documentation
- SRE: `review`; ci-cd, dependencies, security
- testing: `review`; test-quality
- envoy, herald, and scribe: `content`; no scopes

Adversary and guard MUST always be eligible. Other review personas MUST
require scope intersection. Curator MUST require documentation or
user-facing classification. The table MUST inform advisor rules but
MUST NOT become the invocation list. Unknown review personas SHOULD
receive a prompt informed by an explicit valid manifest entry.
Canonical and scaffolded manifests MUST remain byte-identical.

Previously, the capability described five stale `reviewer-*` roles.

#### Scenario: Unknown reviewer role

- **GIVEN** a valid unknown persona is included
- **WHEN** its run is prompted
- **THEN** the prompt includes its declared scope

### Requirement: Council Verdict Policy

**ID: RC-FR-008** This requirement MUST be satisfied.

At least one successful assessment MUST exist. With successful results,
any REQUEST CHANGES MUST produce REQUEST CHANGES. Otherwise an advisory
MUST produce APPROVE WITH ADVISORIES. Only results without blocking or
advisory findings MAY produce APPROVE.

Failed runs MUST be informational when another assessment succeeds.
Zero successful runs MUST follow RD-FR-007 cause precedence.
Availability failures alone MUST produce `UNAVAILABLE`. Any policy,
plan, budget,
limit, persistence, or calculation cause MUST produce `INCONCLUSIVE`
and win over an availability cause. Either result MUST block automated
progression and request retry or human review.

The existing human confirmation gate for LOW/MEDIUM auto-fixes MUST
remain. HIGH/CRITICAL findings MUST NOT be auto-fixed. Advisory
reporting and iteration limits MUST remain unchanged.

Previously, one host-model verdict came from each discovered reviewer.

#### Scenario: One model blocks

- **GIVEN** one successful run approves and one requests changes
- **WHEN** the council consolidates verdicts
- **THEN** the verdict is REQUEST CHANGES

#### Scenario: Advisory-only council

- **GIVEN** successful runs have only LOW/MEDIUM advisories
- **AND** the user declines the fix gate
- **WHEN** the council finalizes
- **THEN** the verdict is APPROVE WITH ADVISORIES

#### Scenario: Every provider is unavailable

- **GIVEN** all runs fail only because providers are unavailable
- **WHEN** the council consolidates verdicts
- **THEN** the verdict is `UNAVAILABLE`
- **AND** approval is not emitted

#### Scenario: Availability and policy failures are mixed

- **GIVEN** one run has an unavailable provider
- **AND** plan validation also fails
- **WHEN** the council consolidates verdicts
- **THEN** the verdict is `INCONCLUSIVE`
- **AND** approval is not emitted

### Requirement: Iterative Fix Loop Scope

**ID: RC-FR-009** This requirement MUST be satisfied.

After confirmed fixes, the loop MUST recompute and validate the advisor
plan and rerun all included runs. It MUST NOT run absent or skipped
personas. The three-iteration limit and escalation policy MUST remain
unchanged.

Previously, the loop reran every discovered reviewer once.

#### Scenario: Plan changes after fixes

- **GIVEN** fixes change the classified diff profile
- **WHEN** the next iteration begins
- **THEN** the plan is recomputed from current evidence
- **AND** the iteration counter is not reset

## REMOVED Requirements

None.
