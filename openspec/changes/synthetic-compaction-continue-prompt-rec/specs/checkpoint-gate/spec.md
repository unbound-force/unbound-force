<!-- capability: checkpoint-gate
     change: synthetic-compaction-continue-prompt-rec
     originating_issue: 1118
     Mechanical enforcement of checkpoint gate overrides against
     synthetic-prompt quoting, declared-step enforcement, and
     escalated-advisory relay. Design rationale: design.md.
     Companion hardening: harden-spec-review-gate (HARDEN-001..006)
     and compaction-storm-on-a-tight-context-mode (DELEG-FR-001..002).
-->

Requirement numbering: this change allocates SYNTH-001 through
SYNTH-005 for the requirements added below. SYNTH-006 through
SYNTH-020 are reserved for future checkpoint-gate changes.
Archived capability requirements live under
`openspec/specs/<capability>/spec.md`. Requirement ids MUST stay
inside the allocated range.

## ADDED Requirements

### Requirement: Gate overrides carry a human-decision reference

**ID: SYNTH-001** This requirement MUST be satisfied.

Any `checkpoint` call that records a gate override (checklist gate,
spec-review advisory escalation, quality gate) MUST carry a
verifiable human-decision reference: a matching `question`-tool
answer from the session or a relayed human decision from the
orchestrator carried in delegation input. A free-text justification
alone (for example quoting "Continue if you have next steps, or
stop and ask for clarification if you are unsure how to proceed")
MUST NOT satisfy the requirement. When the reference is absent or
does not correspond to a recorded human decision, `checkpoint` MUST
reject the call with a machine-parseable error identifying the
missing reference.

#### Scenario: Synthetic prompt quoted as consent is rejected

- **GIVEN** a child calls `checkpoint` with
  `step: checklist-override-phase8-9` and
  `data.checklist_override` quoting "Continue if you have next
  steps" as "User explicitly directed continuation"
- **AND** no matching `question`-tool answer or orchestrator-
  relayed human decision exists for that override
- **WHEN** `checkpoint` evaluates the call
- **THEN** `checkpoint` MUST reject with an error indicating a
  missing human-decision reference
- **AND** MUST NOT persist progress state for that step

#### Scenario: Override with a valid human-decision reference is accepted

- **GIVEN** a gate override is requested via `checkpoint` with a
  `human_decision_ref` naming a recorded `question`-tool answer
  or a delegation relay id that corresponds to a human decision
- **WHEN** `checkpoint` evaluates the call
- **THEN** `checkpoint` MUST record the step with the reference as
  provenance and return success

#### Scenario: Prior-session free-text override does not carry forward

- **GIVEN** a prior session recorded an override as free text with
  no human-decision reference (e.g. "prior session override for
  Phase 1 remains in effect")
- **WHEN** a new session cites that text as justification without
  its own `human_decision_ref`
- **THEN** `checkpoint` MUST reject the new call as missing a
  human-decision reference

### Requirement: Synthetic prompt is never consent

**ID: SYNTH-002** This requirement MUST be satisfied.

Command prose in `/uf.unleash` and `/uf.run-pipeline-plan` and the
autonomy agent prose MUST state explicitly that compaction prompts
and synthetic continue prompts (including "Continue if you have
next steps, or stop and ask for clarification if you are unsure
how to proceed") are not human input and MUST NOT be treated as
authorization for any gate override. A contract test MUST pin that
prose in both the live copy (`.opencode/…`) and the corresponding
scaffold mirror (`internal/scaffold/assets/…`) so drift in either
copy is a test failure.

#### Scenario: Prose pin prevents regression

- **GIVEN** the synthetic-prompt-not-consent paragraph is present
  in `.opencode/commands/uf.unleash.md` and its scaffold mirror
- **WHEN** either copy is edited to remove or alter the paragraph
- **THEN** the contract test MUST fail identifying the missing or
  drifted prose

#### Scenario: Model quoting synthetic prompt is still rejected

- **GIVEN** the prose in SYNTH-002 is present in commands and agent
- **WHEN** a child quotes the synthetic prompt as "User explicitly
  directed continuation" in a `checkpoint` data field
- **THEN** `checkpoint` MUST still reject per SYNTH-001 regardless
  of the quoted text content

### Requirement: Declared checkpoint step enforcement

**ID: SYNTH-003** This requirement MUST be satisfied.

`checkpoint` MUST reject step names that are not in the command's
declared step set for `uf.unleash`, or at minimum those matching
the pattern `*override*`, unless the human-decision reference
required by SYNTH-001 is satisfied. The declared set includes
`clarify`, `plan`, `tasks`, `spec-review`, `implement`,
`implement-phase-N`, `code-review`, `retrospective`, and `demo`.
The narrow `*override*` predicate is the MUST; widening to the
full declared set is a permitted strengthening.

#### Scenario: Free-form override step name is rejected

- **GIVEN** a child calls `checkpoint` with
  `step: checklist-override-phase8-9`
- **AND** no valid `human_decision_ref` is present
- **WHEN** `checkpoint` evaluates the call
- **THEN** `checkpoint` MUST reject with an error naming the
  undeclared step and the missing authorization

#### Scenario: Override step with valid human reference is accepted

- **GIVEN** a child calls `checkpoint` with a step matching
  `*override*`
- **AND** a valid `human_decision_ref` per SYNTH-001 is present
- **WHEN** `checkpoint` evaluates the call
- **THEN** `checkpoint` MUST accept and persist the step

#### Scenario: Normal declared step is unaffected

- **GIVEN** a child calls `checkpoint` with a declared step
  `implement` or `implement-phase-2`
- **WHEN** `checkpoint` evaluates the call
- **THEN** `checkpoint` MUST NOT apply the override-step predicate
  to that call

### Requirement: Escalated ADVISORIES are relayed

**ID: SYNTH-004** This requirement MUST be satisfied.

An `APPROVE WITH ADVISORIES` spec-review result marked
`ESCALATED` MUST be relayed to the human via
`relay_spec_review_gate` (introduced in #1094) before
implementation continues, the same relay path used for
`REQUEST CHANGES`. `checkpoint` MUST refuse to advance past
`spec-review` to `implement` while an unrelayed ESCALATED
advisory result is the latest consolidated spec-review result
and no human-decision reference per SYNTH-001 covers that
escalation.

#### Scenario: Escalated advisories are relayed before implementation

- **GIVEN** spec-review consolidates to
  `APPROVE WITH ADVISORIES (ESCALATED)` with
  `correlation 8ef4f276-…`, 16 findings and 2 proposals
- **AND** no human-decision reference covers the escalation
- **WHEN** the pipeline attempts to advance from `spec-review` to
  `implement`
- **THEN** the pipeline MUST call `relay_spec_review_gate` and
  present the relay to the human
- **AND** MUST NOT advance to `implement` until the relayed
  decision is recorded

#### Scenario: Non-escalated advisories do not block relay

- **GIVEN** spec-review consolidates to
  `APPROVE WITH ADVISORIES` without `ESCALATED`
- **WHEN** the pipeline evaluates advancement
- **THEN** advancement MAY proceed without the SYNTH-004 relay

### Requirement: Regression replay of the incident checkpoint

**ID: SYNTH-005** This requirement MUST be satisfied.

The repository MUST include a deterministic regression test that
replays the exact `checkpoint` JSON from
`ses_e200afc75ffeSuftvVXtQnBJhR` (command `uf.unleash`, step
`checklist-override-phase8-9`, data
`checklist_override: "User explicitly directed continuation via
'Continue if you have next steps'…"`, no `human_decision_ref`)
and asserts `checkpoint` rejects the call. The test MUST also
assert that the synthetic prompt text alone never satisfies
SYNTH-001.

#### Scenario: Regression fixture replays incident JSON

- **GIVEN** the exact incident checkpoint payload is supplied as
  a fixture (step `checklist-override-phase8-9` with the
  quoted synthetic prompt in `checklist_override`)
- **WHEN** the test invokes `checkpoint` with that payload
- **THEN** the test MUST observe a rejection per SYNTH-001 and
  SYNTH-003
- **AND** MUST NOT observe persisted progress for that step
