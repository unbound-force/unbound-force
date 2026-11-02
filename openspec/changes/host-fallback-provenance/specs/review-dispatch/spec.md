<!-- capability: review-dispatch
     change: host-fallback-provenance
     originating_issue: 901
     Archive ordering: this change MUST archive after
     advisor-host-model-fallback — the MODIFIED 'Advisor
     Host-Model Fallback' entry supersedes that change's ADDED
     'Non-advisor with empty profile rejected' scenario text
     per requirement name and archive order; the MODIFY applies
     against the materialized predecessor text and excludes
     only source "host". Strategy and rationale: design.md
     (D6). -->

Requirement numbering: RD-FR-037 is allocated for this change.
RD-FR-038 through RD-FR-048 are reserved for future
review-dispatch changes. The RD-FR-001..036 range is allocated
across two capabilities' unarchived deltas: the review-dispatch
capability carries RD-FR-001..008
(review-council-multi-model-fanout), RD-FR-025..027
(fix-review-dispatch-recovery), RD-FR-028..031
(derive-per-run-timeout), and RD-FR-032..036
(review-dispatch-diff-manifest); the review-matrix capability
carries RD-FR-009..010 (derive-per-run-timeout's review-matrix
delta) and RD-FR-019..024 (optional-profile-models), with
RD-FR-011..018 reserved for future review-matrix changes per
that delta's own numbering note. The unnumbered
`dispatch_agent_run Tool` requirement — including the
tier-default clause RD-FR-037 narrows for `source: "host"` —
lives in the dispatch-agent-run capability delta at
`openspec/changes/shared-agent-executor/specs/dispatch-agent-run.md`.
The model-matrix structure requirements (MM-FR-001 through
MM-FR-008) live in the model-matrix capability's change delta at
`openspec/changes/model-matrix-delegation/specs/model-matrix/spec.md`
(no archived `specs/model-matrix/spec.md` exists yet).

## ADDED Requirements

### Requirement: Host-Source Model Resolution

**ID: RD-FR-037** This requirement MUST be satisfied.

The shared dispatch core (`executeDispatchRun`) MUST resolve the
execution model for a validated `source: "host"` entry that
carries no direct model bypass by calling the shared
`resolveHostModel()` (`.opencode/lib/agent-executor.ts`) with the
calling context's session and message identifiers. Host-source
resolution MUST NOT consult the review matrix profiles: the
planner emits host-fallback entries with null model and null
tier, meaning "use the active host model", and the dispatch core
MUST NOT silently default the tier or substitute a configured
profile model. The `tier ?? "standard"` default MUST remain
reachable only for entries whose source is not `"host"`.

The host-source run MUST persist host-run provenance per
RD-FR-008: `requested_model`, `provider`, `model_id`, and
`variant` null; `resolved_parent_model` non-null after successful
lookup, forming the exact model identity replayed to the child;
`resolved_parent_variant` variant-or-null per the writer contract
of `resolveHostModel()`. The plan-entry model-provenance match and
`model_mismatch` computation MUST then succeed for host-source
runs without any finalizer change.

A host-source run whose host resolution fails MUST fail with
error code `host_model_unavailable` and `retryable: true`, with
an error message passed through the shared
`sanitizeInvocationError()` redaction
(`.opencode/lib/agent-executor.ts`) — the same sanitization the
advisor branch applies, redacting Bearer tokens,
credential-shaped assignments, URL query strings, and home
paths. A host resolution aborted via the cancellation signal
MUST fail with error code `cancelled`. Both failure paths are
pre-execution early exits: no run file is written for them.

The `source` field comparison MUST use strict string equality
(`=== "host"`) with no case normalization, trimming, or
coercion. Unexpected values MUST fall through to the existing
tier-based resolution path. A `source: "host"` entry that also
supplies a direct model bypass MUST execute the direct bypass
unchanged (the bypass branch precedes source-based resolution);
the persisted provenance then follows the non-host shape and
finalization applies its existing host invariants.

`dispatch_review_runs` and `redispatch_failed_runs` MUST resolve
identical host-source plan entries identically through the
shared `resolvePlanEntryRuntime` mapping: a null model and null
tier pass neither field, and the shared core resolves the host
model on both paths.

#### Scenario: Host run inherits the parent model

- **GIVEN** a dispatch plan entry with `source: "host"`,
  `model: null`, and `tier: null`
- **AND** the parent session's current assistant message carries
  an actual model identity and an active variant
- **WHEN** the shared dispatch core resolves the model
- **THEN** it calls `resolveHostModel()` and executes the child
  with the resolved model and variant
- **AND** the persisted run records `requested_model`, `provider`,
  `model_id`, and `variant` as null
- **AND** `resolved_parent_model` is the host model's combined
  identity and `resolved_parent_variant` is its active variant

#### Scenario: Host run without an active variant

- **GIVEN** a dispatch plan entry with `source: "host"`,
  `model: null`, and `tier: null`
- **AND** the parent session's current assistant message carries
  a model identity with no active variant
- **WHEN** the shared dispatch core resolves the model
- **THEN** the run executes on the resolved host model
- **AND** the persisted run records `resolved_parent_variant`
  as null
- **AND** finalization raises no defect for the host run

#### Scenario: Host run ignores a configured profile model

- **GIVEN** a dispatch plan entry with `source: "host"`,
  `model: null`, and `tier: null`
- **AND** the review matrix (or a per-developer override)
  configures a model for the `standard` profile
- **WHEN** the shared dispatch core resolves the model
- **THEN** the run executes on the resolved host model, not the
  configured profile model
- **AND** the persisted run carries host-run provenance (null
  requested model, non-null resolved parent model after
  successful lookup)

#### Scenario: Host resolution failure

- **GIVEN** a host-source plan entry and a parent session whose
  current message cannot be resolved
- **WHEN** `resolveHostModel()` fails
- **THEN** the run fails with error code `host_model_unavailable`
- **AND** `retryable` is true
- **AND** the error message is the `sanitizeInvocationError()`
  form of the resolution failure (no Bearer tokens,
  credential-shaped assignments, URL query strings, or home
  paths)
- **AND** no run file is written

#### Scenario: Host resolution cancelled

- **GIVEN** a host-source plan entry
- **WHEN** `resolveHostModel()` is aborted via the cancellation
  signal
- **THEN** the run fails with error code `cancelled`
- **AND** no run file is written

#### Scenario: Host-fallback panel finalizes cleanly

- **GIVEN** a dispatch session whose plan includes an
  include-decision host-fallback entry (agent absent from both
  `runs[mode]` and `advisor[mode]`)
- **AND** the host run succeeds with host-run provenance
- **WHEN** `consolidate_dispatch` runs with provenance and calls
  the finalizer
- **THEN** `validateRunModel` raises no defect for the host run
- **AND** finalization succeeds with the native verdict

#### Scenario: Re-dispatched host run reuses host resolution

- **GIVEN** a failed host-source run persisted with its original
  prompt and plan entry
- **WHEN** `redispatch_failed_runs` re-executes it
- **THEN** the re-run resolves the host model through the same
  shared-core branch
- **AND** the replacement run file carries host-run provenance

#### Scenario: Non-host sources retain tier resolution

- **GIVEN** a `dispatch_agent_run` invocation with no `source`
  (defaulting to `explicit`), no model, and no tier
- **AND** the `standard` profile has no model configured
- **WHEN** the shared dispatch core resolves the model
- **THEN** it fails with `tier_model_unavailable` as before

#### Scenario: Direct bypass overrides host resolution

- **GIVEN** a `dispatch_agent_run` invocation with
  `source: "host"` and an explicit direct `model` bypass
- **WHEN** the shared dispatch core resolves the model
- **THEN** it executes the direct bypass unchanged and records
  the requested model
- **AND** host resolution is not consulted
- **AND** the persisted provenance follows the non-host shape,
  so finalization applies its existing host invariants and
  fails closed on the contradictory provenance input

## MODIFIED Requirements

### Requirement: Advisor Host-Model Fallback

Previously: the requirement's "Non-advisor with empty profile
rejected" scenario and its closing prose routed every source
other than `"advisor"` — including `"host"` — to the
`tier_model_unavailable` rejection path when the profile has no
model. Modified: the full replacement text below is verbatim
from the advisor-host-model-fallback delta except that
`source: "host"` is excluded from that rejection path (it
resolves the parent model per RD-FR-037 instead). Supersession
rationale: design.md (D6); archive ordering: this change MUST
archive after advisor-host-model-fallback, per the header.

When an advisor run's review-matrix profile omits a `model`
field, `dispatch_agent_run` MUST resolve the parent session's
active model via `resolveHostModel` and use it as the execution
model. The run MUST use host-run provenance semantics: null
requested model provenance with resolved parent model
provenance.

#### Scenario: Advisor inherits parent model

- **GIVEN** the review matrix profile for the requested tier
  has no `model` field
- **AND** the dispatch input has `source: "advisor"`
- **WHEN** `dispatch_agent_run` resolves the model
- **THEN** it calls `resolveHostModel` to obtain the parent
  session's active model identity and variant
- **AND** the run executes with the resolved model
- **AND** `requested_model` is `null`
- **AND** `resolved_parent_model` is the host model's combined
  identity
- **AND** `resolved_parent_variant` is the host model's active
  variant

#### Scenario: Host resolution failure

- **GIVEN** the review matrix profile has no `model`
- **AND** `source: "advisor"`
- **WHEN** `resolveHostModel` fails (session message
  unavailable)
- **THEN** `dispatch_agent_run` returns a failed result with
  error code `host_model_unavailable`
- **AND** `retryable` is `true`
- **AND** the error message is a sanitized availability error
  that MUST NOT leak internal session details (session IDs,
  message IDs, internal API paths)

#### Scenario: Host resolution cancelled

- **GIVEN** the review matrix profile has no `model`
- **AND** `source: "advisor"`
- **WHEN** `resolveHostModel` is aborted via signal
- **THEN** `dispatch_agent_run` returns a failed result with
  error code `cancelled`

#### Scenario: Non-advisor with empty profile rejected

- **GIVEN** the review matrix profile has no `model`
- **AND** `source` is neither `"advisor"` nor `"host"` (or is
  absent)
- **WHEN** `dispatch_agent_run` resolves the model
- **THEN** it returns `tier_model_unavailable` as before

#### Scenario: Host source with empty profile uses host model

- **GIVEN** the review matrix profile has no `model`
- **AND** `source` is `"host"`
- **WHEN** `dispatch_agent_run` resolves the model
- **THEN** it resolves the parent session's active model per
  RD-FR-037, or fails with retryable `host_model_unavailable`
  — never `tier_model_unavailable`

The `source` field comparison MUST use strict string equality
(`=== "advisor"` for the advisor branch, `=== "host"` for the
host branch per RD-FR-037) with no case normalization,
trimming, or coercion. Unexpected values (empty string, `null`,
`undefined`, case variations like `"Advisor"` or `"ADVISOR"`)
MUST fall through to the `tier_model_unavailable` rejection
path; `source: "host"` is excluded from that fallthrough per
RD-FR-037.

## REMOVED Requirements

None.
