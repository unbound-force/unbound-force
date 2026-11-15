<!-- capability: review-dispatch -->

## ADDED Requirements

> **Numbering**: RD-FR-063 through RD-FR-065 are allocated for
> this change. The highest previously allocated review-dispatch
> requirement is RD-FR-062
> (`review-dispatch-a-reviewer-with-empty-ou`). This change
> builds on the execution contracts of
> `optional-profile-models` (RD-FR-019 through RD-FR-024),
> `advisor-host-model-fallback` (#693),
> `host-fallback-provenance` (RD-FR-037), and
> `split-host-resolution-failure-classes` (RD-FR-038).

### Requirement: Planner Signals Host-Model Inheritance

**ID: RD-FR-063** This requirement MUST be satisfied.

When `plan_review_dispatch` emits an include-entry whose
`source` is `"advisor"` and whose resolved profile has no
configured `model` — from base-run resolution, the full panel,
or opt-in risk augmentation — the plan MUST carry an advisory
naming the profile (the tier when the profile is a tier name),
stating that the run will inherit the parent session's active
model at execution (advisor host-model fallback, #693), and
naming the remediation: configure `profiles.<name>.model` in
`.uf/review-matrix.yaml` or the gitignored override.

The advisory MUST NOT change `plan.status` — the plan stays
`ready`; the fallback is sanctioned execution — and MUST NOT
be truncated. Emission MUST be deterministic: one advisory per
distinct profile, advisory order sorted by profile name, and
the affected agents listed in the plan's agent order.
Include-entries whose `source` is `"host"` MUST NOT trigger
this advisory.

#### Scenario: Advisor run on a model-less profile

- **GIVEN** a review matrix where `profiles.standard` has no
  `model` and `defaults.code` is `standard`
- **AND** an agent listed in `advisor.code` with no explicit
  run
- **WHEN** `plan_review_dispatch` builds the plan for a code
  review
- **THEN** the agent's include-entry has `source: "advisor"`
  and `model: null`
- **AND** the plan `status` is `ready`
- **AND** the plan carries exactly one advisory for profile
  `standard` naming the agent, the host-model inheritance
  fallback, and the remediation

#### Scenario: Full panel on a model-less standard profile

- **GIVEN** a review matrix where `profiles.standard` has no
  `model`
- **WHEN** `plan_review_dispatch` builds a full-panel plan
- **THEN** every review-capable persona's include-entry has
  `source: "advisor"` and `model: null`
- **AND** the plan carries one advisory for profile `standard`
  naming the affected agents, not one advisory per entry

#### Scenario: Risk augmentation selects a model-less profile

- **GIVEN** a matrix where risk augmentation is enabled and the
  adjacent profile for the change's tier has no `model`
- **AND** the caller opts in to augmentation
- **WHEN** the plan is built
- **THEN** the augmentation run's include-entry has
  `source: "advisor"` and `model: null`
- **AND** the plan carries an advisory naming the adjacent
  profile and the host-model inheritance fallback

#### Scenario: Profile with a model emits no advisory

- **GIVEN** a review matrix where every profile referenced by
  the plan has a configured `model`
- **WHEN** the plan is built
- **THEN** the plan carries no host-model-inheritance advisory

#### Scenario: Host-source entries stay self-describing

- **GIVEN** an agent absent from both `runs` and `advisor`
  configuration for the mode
- **WHEN** the plan is built
- **THEN** the agent's host-fallback include-entry has
  `model: null` and its existing reason naming the active host
  model
- **AND** no additional advisory is emitted for it

#### Scenario: Advisory emission is deterministic

- **GIVEN** the same matrix, discovered agents, and change
  inputs
- **WHEN** the plan is built twice
- **THEN** the advisory arrays are identical, with one entry
  per distinct model-less profile sorted by profile name and
  affected agents in plan order

### Requirement: Planner Fails Closed on Unexecutable Explicit Runs

**ID: RD-FR-064** This requirement MUST be satisfied.

When a matrix `runs.<mode>.<agent>` entry uses the `profile`
form and the named profile has no configured `model`, the
planner MUST reject the entry before dispatch: the entry
records a validation error naming the profile and the
remediation, its `decision` is `skip`, and the plan returns
`status: "inconclusive"` with `workflow_result:
"INCONCLUSIVE"` and a plan error naming the agent and profile
— riding the existing invalid-explicit-run mechanism.

An explicit-source run never reaches host-model inheritance:
execution deterministically rejects it with non-retryable
`tier_model_unavailable` before any child session exists, so a
plan that emits it as `ready` is unexecutable as written.
Explicit entries using the `model` form, and advisor- or
host-destination runs on the same model-less profile, are
unaffected (RD-FR-063 governs the advisor case).

#### Scenario: Explicit profile run against a model-less profile

- **GIVEN** a matrix where `runs.code.divisor-architect` is
  `[{ profile: heavy }]` and `profiles.heavy` has no `model`
- **WHEN** `plan_review_dispatch` builds the plan
- **THEN** the plan `status` is `"inconclusive"` and
  `workflow_result` is `"INCONCLUSIVE"`
- **AND** the plan errors name `divisor-architect`, the
  `heavy` profile, and the remediation
- **AND** the entry's `decision` is `"skip"` with a
  non-empty `validation_errors`

#### Scenario: Direct-model explicit runs are unaffected

- **GIVEN** a matrix where `runs.code.divisor-architect` is
  `[{ model: provider/model }]`
- **WHEN** the plan is built
- **THEN** the entry is an include-entry with that model and
  no new validation error

#### Scenario: Advisor runs on the same profile stay includable

- **GIVEN** a matrix where `profiles.heavy` has no `model` and
  an advisor-listed agent is pinned to `heavy` via `always`
- **WHEN** the plan is built
- **THEN** the advisor entry remains an include-entry with
  `model: null`
- **AND** the plan follows RD-FR-063 (advisory, `ready`)
  rather than this requirement

#### Scenario: Unknown profile keeps today's rejection

- **GIVEN** a matrix where an explicit run references a
  profile absent from `profiles`
- **WHEN** the plan is built
- **THEN** the existing `invalid-explicit-run` rejection
  applies unchanged

### Requirement: All-Runs-Failed Summary Carries Failure Causes

**ID: RD-FR-065** This requirement MUST be satisfied.

`dispatch_review_runs` MUST capture every failed run's cause
into the result `warnings`: a fulfilled-but-failed run —
including pre-execution early exits such as
`tier_model_unavailable`, `host_model_structural`,
`host_model_unavailable`, and `unknown_agent` — records a
warning naming the agent, run sequence, error code, and error
message; a rejected batch promise keeps the existing
`run rejected:` warning. Capture applies on every completion,
including partial failure.

When the dispatch completes with `runs_failed > 0` and
`runs_completed === 0`, the result MUST set `message` to the
first run-failure cause (first failure in batch order),
falling back to the first warning when no run cause exists, so
the caller-visible one-line summary renders it under the
existing #946 contract (one line, at most 512 characters
including the ellipsis, `retryable` omitted when unknown). The
`no failure message reported` placeholder remains only when
both are absent. A dispatch with at least one completed run
keeps its `ok` status; causes remain visible through
`runs_failed` and the captured warnings.

#### Scenario: All runs rejected on an unresolvable tier

- **GIVEN** a dispatch whose every included run fails
  pre-execution with `tier_model_unavailable` naming a
  model-less profile
- **WHEN** `dispatch_review_runs` completes
- **THEN** the result `status` is `"error"` with
  `runs_completed: 0`
- **AND** `message` names the first failure's error code and
  profile
- **AND** the one-line summary ends with that cause rather
  than `no failure message reported`

#### Scenario: Early-exit causes are captured as warnings

- **GIVEN** a dispatch where one run early-exits with
  `host_model_unavailable` and another is rejected before
  execution with an unexpected throw
- **WHEN** the dispatch completes
- **THEN** `warnings` contains an entry naming the first
  run's agent, sequence, `host_model_unavailable`, and its
  message
- **AND** `warnings` contains the existing `run rejected:`
  entry for the thrown run

#### Scenario: Partial failure keeps the ok summary

- **GIVEN** a dispatch where one run fails and another
  completes
- **WHEN** the dispatch completes
- **THEN** the result `status` is `"ok"` with no `message`
  fold
- **AND** the failed run's cause is still captured in
  `warnings`

#### Scenario: Long causes stay bounded

- **GIVEN** an all-failed dispatch whose first cause exceeds
  512 characters after line collapsing
- **WHEN** the summary line is rendered
- **THEN** the rendered cause is cut to at most 512
  characters including the ellipsis

#### Scenario: First-cause selection is deterministic

- **GIVEN** the same plan, prompts, and failing children
- **WHEN** the dispatch runs
- **THEN** the selected `message` is the first failure in
  batch order, identically on every run

## REMOVED Requirements

None.
