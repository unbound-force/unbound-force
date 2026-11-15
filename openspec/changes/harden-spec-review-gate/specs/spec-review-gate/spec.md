<!-- capability: spec-review-gate
     change: harden-spec-review-gate
     originating_issue: 1078
     Mechanical enforcement of the spec-review gate: consolidation
     persistence before model resumption, per-finding decision
     gate with post-decision control flow, iteration re-dispatch
     guard, deterministic orchestrator relay, bypass-path
     regression coverage, and resume-probe adjudication.
     Design rationale and rejected alternatives: design.md. -->

Requirement numbering: HARDEN-001 through HARDEN-006 are
allocated for this change. HARDEN-007 through HARDEN-010 are
reserved for future spec-review-gate changes. The sibling change
`spec-review-per-finding-gate` (issue #777) allocates the
unprefixed FR-001..FR-007 requirements of the same capability in
its unarchived delta; archived capability requirements live under
`openspec/specs/<capability>/spec.md`. Requirement ids MUST stay
inside the allocated range.

## ADDED Requirements

### Requirement: Consolidation Persistence Before Model Resumption

**ID: HARDEN-001** This requirement MUST be satisfied.

The system MUST persist the result of `consolidate_dispatch` for
spec-review to progress state before control returns to the
model. The persisted record MUST include `verdict`,
`correlation_id`, `iteration`, `finding_count`, and `finding_ids`
(one identity per finding, so per-finding decisions can be
tracked). `get_progress` MUST reflect this persisted iteration; a
stale `current_step: spec-review` from a prior iteration MUST NOT
be returned after consolidation. This requirement's text is the
single normative shape for the persisted record and for the
per-finding decision contracts; scenario record excerpts are
illustrative subsets.

#### Scenario: REQUEST CHANGES consolidation is recorded

- **GIVEN** a child has called `dispatch_review_runs` →
  `consolidate_dispatch` for spec-review iteration 3 with verdict
  REQUEST CHANGES, correlation
  `5355acba-62f4-4859-8775-7294b3e8702c`, 1 HIGH + 7 MEDIUM/LOW
- **WHEN** `consolidate_dispatch` completes
- **THEN** progress state MUST contain
  `spec-review {iteration: 3, verdict: REQUEST_CHANGES,
  correlation_id: 5355acba-…, finding_count: 8,
  finding_ids: [ …8 finding identities… ]}`
- **AND** a subsequent `get_progress` MUST return `iteration: 3`
  (not iteration 2) and MUST NOT show a stale step

#### Scenario: APPROVE still advances

- **GIVEN** spec-review iteration 3 consolidates to APPROVE
- **WHEN** consolidation is recorded
- **THEN** progress state MUST reflect `verdict: APPROVE` and the
  orchestrator MAY advance past spec-review without per-finding
  decisions

### Requirement: Per-Finding Gate Before Advancing Spec-Review

**ID: HARDEN-002** This requirement MUST be satisfied.

When the latest recorded spec-review result is non-APPROVE
(REQUEST CHANGES), `checkpoint` MUST refuse to advance
`spec-review` to the next phase/step unless every finding in that
result has a recorded per-finding decision of `apply` (with a
reference to the applied fix), `override` (with a non-empty
reason), or `pause`. The refusal MUST be mechanical
(tool-enforced), not dependent on the model's prose claim of
"addressed".

Post-decision control flow: an `apply` decision MUST be followed
by a spec-review iteration N+1 that consolidates to APPROVE, or a
recorded human override set, before `checkpoint` may advance
`spec-review`. A decision set of `override` entries alone (no
`apply` decisions) permits advancement with the overrides
recorded; a `pause` decision holds the gate — the run exits
through the paused exit path and does not advance.

#### Scenario: Checkpoint refuses advance without decisions

- **GIVEN** progress state holds spec-review iteration 3 REQUEST
  CHANGES with 8 findings and no per-finding decisions
- **WHEN** the child (or any caller) attempts `checkpoint`
  advancing `spec-review` (e.g., marking spec-review done or
  starting implementation)
- **THEN** `checkpoint` MUST return an error indicating missing
  per-finding decisions and MUST NOT advance the step

#### Scenario: Apply decisions require a subsequent APPROVE iteration

- **GIVEN** progress state holds spec-review iteration 3 REQUEST
  CHANGES whose 8 findings all have recorded `apply` decisions
  with fix references, and the fixes are applied
- **WHEN** `checkpoint` is called to advance `spec-review` before
  a subsequent iteration consolidates to APPROVE
- **THEN** it MUST refuse with an error indicating a pending
  approving iteration and MUST NOT advance the step
- **AND GIVEN** spec-review iteration 4 subsequently consolidates
  to APPROVE
- **WHEN** `checkpoint` is called to advance `spec-review`
- **THEN** it MUST succeed

#### Scenario: Overrides are audited

- **GIVEN** a finding is to be dismissed without a spec fix
- **WHEN** its decision is recorded
- **THEN** it MUST be `override` with a non-empty reason, and the
  checkpointed record MUST include
  `overrides: [{finding_id, reason}]` (empty when no overrides)

#### Scenario: All-override decision set — advance allowed

- **GIVEN** every finding in iteration 3's REQUEST CHANGES has a
  recorded `override` decision with reason, and no finding has an
  `apply` decision
- **WHEN** `checkpoint` is called to advance `spec-review`
- **THEN** it MUST succeed and record the decisions (the
  overrides array) alongside the iteration's verdict

### Requirement: Iteration Re-Dispatch Guard

**ID: HARDEN-003** This requirement MUST be satisfied.

A re-dispatch request for a spec-review iteration number that
already has a persisted verdict MUST be refused with an
`iteration_already_recorded` error naming the recorded iteration
and its verdict. Re-rolling the same iteration number until
APPROVE is forbidden. The guard MUST enforce the iteration cap
(refuse when the cap would be exceeded). A caller that needs a
fresh review MUST dispatch the next iteration number.

#### Scenario: Same-iteration re-dispatch is refused

- **GIVEN** progress state already holds spec-review iteration 3
  with verdict REQUEST CHANGES (correlation `5355acba-…`)
- **WHEN** the orchestrator prompts a new child to "re-dispatch
  iteration 3"
- **THEN** the dispatch layer MUST return an
  `iteration_already_recorded` error naming iteration 3 and its
  recorded verdict, and MUST NOT overwrite iteration 3's REQUEST
  CHANGES with a new APPROVE under the same iteration number

#### Scenario: Normal next-iteration dispatch succeeds

- **GIVEN** iteration 3 is recorded as REQUEST CHANGES with a
  complete decision set whose `apply` fixes are applied
- **WHEN** spec-review iteration 4 is dispatched
- **THEN** it MUST proceed and its result MUST be recorded as
  iteration 4 (not overwriting iteration 3)

### Requirement: Unconditional Orchestrator Relay After Child Return

**ID: HARDEN-004** This requirement MUST be satisfied.

When a delegated `/uf.unleash` child returns (exit `success` or
otherwise), `/uf.run-pipeline-plan` MUST make one unconditional
intent-level call: `relay_spec_review_gate` (registered in the
uf-workflow plugin). The tool — not command prose — reads
progress state and the dispatch artifact
(`.uf/artifacts/dispatch/<correlation>.json`) and computes the
pause state: an undecided non-APPROVE spec-review iteration →
`{gate_paused: true, relay: …}`; a decided gate, an approved
gate, or no spec-review record → `{gate_paused: false}`. The
orchestrator MUST present the rendered relay to the human when
the tool returns `gate_paused: true` and MUST proceed only when
the tool returns `gate_paused: false`. The rendered relay —
every finding's severity, description, file:line, and the
correlation id — MUST be produced by the tool from the
artifact, without an LLM call, and MUST be artifact-sourced,
never child-text sourced: a compaction summary, empty text, or
a claim that the findings were "addressed" cannot change it.
The child's returned text is never the decision input, and no
prose-evaluated condition may sit between the child's return
and the relay call.

#### Scenario: Compaction-summary return relays from artifact

- **GIVEN** child `ses_e254aedb8ffeqlLkdOgSpKNj60` consolidated
  REQUEST CHANGES (8 findings, correlation `5355acba-…`) then
  edited specs and returned at 17:29 a compaction summary with no
  gate report, and `get_progress` shows the recorded iteration 3
  REQUEST CHANGES
- **WHEN** the orchestrator processes the child's return
- **THEN** its unconditional `relay_spec_review_gate` call MUST
  return `gate_paused: true` and render, from the dispatch
  artifact for `5355acba-…`, all 8 findings (1 HIGH + 7
  MEDIUM/LOW) with correlation id, for presentation of the
  gate (apply/override/pause) to the human — it MUST NOT
  dispatch a new child to re-roll iteration 3 to APPROVE

#### Scenario: Correct REQUEST CHANGES text — still artifact-sourced

- **GIVEN** a child returns REQUEST CHANGES in its text
- **WHEN** the orchestrator relays to the human
- **THEN** the relayed finding list MUST be the output of
  `relay_spec_review_gate` (artifact-sourced, correlation id
  included), not the child's text

#### Scenario: No spec-review record — relay reports not paused

- **GIVEN** a `/uf.unleash` child returns before any spec-review
  dispatch was recorded (progress state holds no spec-review
  entry)
- **WHEN** the orchestrator makes its unconditional
  `relay_spec_review_gate` call
- **THEN** the tool MUST return `gate_paused: false` and the
  orchestrator proceeds by its other state checks

### Requirement: Regression Coverage for Bypass Path

**ID: HARDEN-005** This requirement MUST be satisfied.

Regression tests MUST cover the bypass path: "child consolidates
REQUEST CHANGES, edits spec files, returns non-gate text (e.g.,
compaction summary); `get_progress` would previously show stale
state; orchestrator must not lose findings".

#### Scenario: Bypass path is tested

- **GIVEN** a fixture replaying child
  `ses_e254aedb8ffeqlLkdOgSpKNj60`'s behavior (REQUEST CHANGES →
  3 edits to spec.md → compaction summary return)
- **WHEN** the regression suite runs
- **THEN** it MUST assert (a) consolidation is persisted before
  the edits (state visible immediately after
  `consolidate_dispatch` returns), (b) `checkpoint` refuses
  advance without decisions, (c) stale `get_progress` cannot
  occur, and (d) `relay_spec_review_gate` renders 8 findings from
  the artifact with correlation `5355acba-…` (deterministic, no
  LLM) and no `overrides: []` is auto-written as if findings were
  never raised

### Requirement: Resume Probe Honors Recorded Adjudication

**ID: HARDEN-006** This requirement MUST be satisfied.

`uf_unleash_detect_resume` MUST honor a recorded per-finding
decision set for the latest recorded spec-review iteration as the
adjudication of that iteration:

- When the latest recorded iteration is non-APPROVE with a
  complete decision set of `override` entries (no `apply`
  decisions), the probe MUST report `spec_review` done, matching
  the advancement HARDEN-002 permits.
- When the latest recorded iteration is non-APPROVE with `apply`
  decisions, the probe MUST report `spec_review` done only when a
  subsequent iteration consolidates to APPROVE — and then MUST
  NOT re-fail the probe on dirty spec artifacts (the applied
  fixes dirty spec.md by design).
- When the latest recorded iteration is non-APPROVE with missing
  or partial decisions, the probe MUST report `spec_review`
  incomplete and name the undecided iteration; resume re-enters
  the gate rather than dispatching a fresh review.

#### Scenario: Override-only adjudication survives resume

- **GIVEN** progress state holds spec-review iteration 3 REQUEST
  CHANGES with all-override decisions recorded and `spec-review`
  advanced
- **WHEN** `uf_unleash_detect_resume` runs
- **THEN** it MUST report `spec_review` done and MUST NOT trigger
  a re-dispatch of the gate the human already adjudicated

#### Scenario: Undecided iteration reports incomplete

- **GIVEN** progress state holds spec-review iteration 3 REQUEST
  CHANGES with no recorded decisions
- **WHEN** `uf_unleash_detect_resume` runs
- **THEN** it MUST report `spec_review` incomplete, naming
  iteration 3 as the undecided iteration

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
