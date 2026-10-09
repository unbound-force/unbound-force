## ADDED Requirements

### Requirement: Advisor-Backed Issue Triage

**ID: DW-FR-001** This requirement MUST be satisfied.

`/uf.triage-issue` MUST replace its hardcoded five-agent panel with an
advisor-selected panel from the six review-capable personas and any
additional valid review-capable persona. It MUST execute plan runs
through `invoke_agent`.

Triage MUST derive selection and model tier only from the deterministic
RD-FR-002 and RD-FR-003 issue-content profile. It MUST use issue title,
nullable body, and ordered comments. It MUST NOT request, synthesize,
or infer code-diff metrics. The artifact MUST record the normalized
content hash, matched rule ids, content metrics, categories, and tier.

Model runs for one persona MUST first collapse through the existing
three-rule majority algorithm. The panel MUST apply the same algorithm
across persona votes. Failed runs MUST NOT vote, so fan-out MUST NOT
increase a persona's weight.

The command MUST emit a review-dispatch artifact. When no successful
persona vote exists, it MUST apply RD-FR-007 cause precedence. It MUST
NOT classify the issue as VALID or INVALID.

#### Scenario: Fanned persona has one vote

- **GIVEN** adversary runs twice and five personas run once
- **WHEN** triage consolidates model results
- **THEN** adversary produces one persona verdict
- **AND** the panel majority uses six persona votes

#### Scenario: Comments alter triage routing deterministically

- **GIVEN** title and body match no category
- **AND** an ordered comment contains `pipeline failure`
- **WHEN** triage planning runs twice with identical issue content
- **THEN** both plans select the `ci-cd` category
- **AND** both plans and content hashes are byte-equivalent

#### Scenario: No successful triage assessment

- **GIVEN** every planned triage run has only availability failures
- **WHEN** triage consolidates results
- **THEN** issue validity is `UNAVAILABLE`
- **AND** automated triage progression stops

### Requirement: Advisor-Backed Feedback Escalation

**ID: DW-FR-002** This requirement MUST be satisfied.

Tier 2 feedback MUST retain domain routing but obtain runs from the
advisor and execute them through `invoke_agent`. Tier 1 and the
fallback when no Divisor agents are deployed MUST remain unchanged.

The strictest successful Tier 2 recommendation MUST win, with ACCEPT
stricter than AUTHOR-DECIDES. Failed runs MUST NOT strengthen a result.
The command MUST emit a review-dispatch artifact. If planned Tier 2 has
no successful assessment, it MUST apply RD-FR-007 cause precedence.
It MUST mark Tier 2 unavailable and require retry or human review.

#### Scenario: Multi-domain feedback

- **GIVEN** feedback affects architecture and test strategy
- **WHEN** it enters Tier 2
- **THEN** architect and testing are requested from the advisor
- **AND** their runs follow the validated plan

#### Scenario: Strictest successful recommendation

- **GIVEN** one run recommends ACCEPT
- **AND** another recommends AUTHOR-DECIDES
- **WHEN** recommendations are consolidated
- **THEN** the Tier 2 recommendation is ACCEPT

#### Scenario: All Tier 2 runs fail

- **GIVEN** every Tier 2 run has a policy or limit failure
- **WHEN** feedback is consolidated
- **THEN** the result is `INCONCLUSIVE`
- **AND** it is not treated as AUTHOR-DECIDES

### Requirement: Advisor-Backed Speckit Test Review

**ID: DW-FR-003** This requirement MUST be satisfied.

`/speckit.testreview` MUST dispatch `divisor-testing` through the
advisor and `invoke_agent`. It MUST use code-review profiling limited
to the testing persona and preserve the test-review scope and verdict
contract.

The command MUST emit a review-dispatch artifact. No successful testing
assessment MUST apply RD-FR-007 cause precedence and block automated
progression. Failed runs MUST NOT become fabricated test findings.

#### Scenario: Test-review selection

- **GIVEN** `divisor-testing` is available
- **WHEN** test review delegates
- **THEN** the plan is restricted to `divisor-testing`
- **AND** configured explicit-first resolution applies

#### Scenario: Testing models unavailable

- **GIVEN** every testing run has only model availability failures
- **WHEN** test review consolidates results
- **THEN** the result is `UNAVAILABLE`
- **AND** no passing verdict is emitted

### Requirement: Workflow Provenance and Context

**ID: DW-FR-004** This requirement MUST be satisfied.

Each advisor-backed workflow MUST retain agent, model, variant, source,
and run sequence in human and artifact output. Deduplicated findings
MUST list all contributing runs.

Host runs MUST distinguish null requested model and variant from the
resolved parent model and active variant. They MUST also retain the
reported child model. Resolution failure MUST remain an availability
error rather than substitute configured defaults.

Commands that request sibling evidence MUST apply the `review-context`
confinement and lesson-proposal contracts. Commands that do not request
sibling evidence MUST NOT acquire it implicitly. Model self-reporting
MUST be requested at every direct dispatch site.

#### Scenario: Duplicate model finding

- **GIVEN** two runs report the same file and root cause
- **WHEN** the parent presents the finding
- **THEN** it appears once
- **AND** both run identities remain as provenance

## MODIFIED Requirements

### Requirement: Existing Triage Majority Policy

**ID: DW-FR-005** This requirement MUST be satisfied.

Issue triage MUST continue using its three ordered majority rules.
Voters MUST be successful consolidated persona verdicts, not raw runs.

Previously, one verdict came from each member of a hardcoded five-agent
panel.

#### Scenario: Majority across available personas

- **GIVEN** seven personas produce successful verdicts
- **WHEN** four return VALID and three return INVALID
- **THEN** overall issue validity is VALID

### Requirement: Existing Feedback Tier Boundary

**ID: DW-FR-006** This requirement MUST be satisfied.

Simple feedback MUST remain at Tier 1. Security, architecture,
multi-file, ambiguity, testing, performance, and operational concerns
MUST continue to enter Tier 2. Only Tier 2 dispatch and consolidation
change.

Previously, Tier 2 called selected agents through the host-model Task
tool.

#### Scenario: Simple feedback remains Tier 1

- **GIVEN** a naming preference has no cross-cutting impact
- **WHEN** the feedback is classified
- **THEN** it is assessed at Tier 1
- **AND** no advisor plan is created

## REMOVED Requirements

None.
