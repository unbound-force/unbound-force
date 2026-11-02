<!-- capability: review-dispatch
     change: split-host-resolution-failure-classes
     originating_issue: 914
     Archive ordering: this change MUST archive after
     host-fallback-provenance — the MODIFIED 'Host-Source
     Model Resolution' (RD-FR-037) and 'Advisor Host-Model
     Fallback' entries supersede that change's ADDED and
     MODIFIED text per requirement name and archive order.
     The MODIFY applies against the materialized
     predecessor text and narrows only the host-resolution
     failure clause. Strategy and rationale: design.md
     (D7). -->

Requirement numbering: RD-FR-038 is allocated for this change.
RD-FR-039 through RD-FR-048 are reserved for future
review-dispatch changes. The RD-FR-001..037 range is allocated
across the review-dispatch capability's unarchived and archived
deltas (RD-FR-001..008 review-council-multi-model-fanout,
RD-FR-025..027 fix-review-dispatch-recovery, RD-FR-028..031
derive-per-run-timeout, RD-FR-032..036
review-dispatch-diff-manifest, RD-FR-037
host-fallback-provenance); the review-matrix capability carries
RD-FR-009..010 and RD-FR-019..024, with RD-FR-011..018 reserved
there. The unnumbered `dispatch_agent_run Tool` requirement lives
in the dispatch-agent-run capability delta (shared-agent-executor,
as MODIFIED by host-fallback-provenance); it is not changed here.
The model-matrix structure requirements (MM-FR-001..008) live in
the model-matrix capability delta. Requirement ids MUST stay
inside the allocated range.

## ADDED Requirements

### Requirement: Host-Resolution Failure Classification

**ID: RD-FR-038** This requirement MUST be satisfied.

The shared dispatch core (`executeDispatchRun`) MUST classify
every `resolveHostModel()` failure through one shared
classification helper called by both branches that invoke host
resolution — the host-source branch (RD-FR-037) and the advisor
host-model fallback. The two branches MUST NOT carry duplicated
classification logic.

A structural failure — an error that is an instance of
`StructuralHostModelError` (`.opencode/lib/agent-executor.ts`) —
MUST fail the run with error code `host_model_structural`,
`retryable: false`, and a message that preserves both the shared
failure's message and its remediation string, rendered as
`<message>. Remediation: <remediation>; retrying will not fix
this.` The message MUST pass through the shared
`sanitizeInvocationError()` redaction, which redacts Bearer
tokens, credential-shaped assignments, URL query strings, and
home paths.

Any other resolution failure is transient and MUST keep the
existing class: error code `host_model_unavailable`,
`retryable: true`, and a sanitized message that carries the
failure's text and no remediation suffix.

Consumers MUST be able to distinguish the classes from
`error.code` and `error.retryable` alone; they MUST NOT need to
parse message prose. A resolution aborted via the cancellation
signal MUST take precedence over classification and fail with
error code `cancelled`.

Both classified failures are pre-execution early exits: no run
file is written. On such an early exit, `retryable` means caller
re-invocation only. `retryable: true` means the caller MAY
re-invoke the dispatch once the lookup recovers;
`retryable: false` means re-invocation cannot succeed until the
structural cause is fixed. An early exit is NOT a DR-FR-002
redispatch target, because no run file exists for
`redispatch_failed_runs` to target. Consolidation of a dispatch
session whose included plan entry early-exited MUST fail closed
with an error containing `included plan entry
<agent>/<sequence> has no run` and an `INCONCLUSIVE` verdict. A
`redispatch_failed_runs` re-run that early-exits is distinct: it
MUST restore the prior run file, report outcome `early_exit`
carrying the classified error, and return the session to its
pre-attempt state.

The recorded forms of the failure — the `dispatch_agent_run`
result and the `redispatch_failed_runs` early-exit result entry
that repeats the outcome error — MUST preserve the classified
`code`, `retryable`, and remediation-carrying sanitized message,
rather than reducing the failure to its bare message text.

#### Scenario: Structural failure on the host-source branch

- **GIVEN** a `source: "host"` plan entry with null model and
  null tier
- **AND** the parent session's current message is not an
  assistant message
- **WHEN** `resolveHostModel()` throws
  `StructuralHostModelError`
- **THEN** the run fails with error code `host_model_structural`
- **AND** `retryable` is false
- **AND** the message carries `Remediation: ` followed by the
  shared failure's remediation string and ends with `retrying
  will not fix this.`
- **AND** the message is the `sanitizeInvocationError()` form
- **AND** no run file is written

#### Scenario: Structural failure on the advisor fallback

- **GIVEN** an advisor plan entry whose review-matrix profile has
  no model
- **AND** the parent session's assistant message carries no model
  identity
- **WHEN** `resolveHostModel()` throws
  `StructuralHostModelError`
- **THEN** the run fails with error code `host_model_structural`
- **AND** `retryable` is false
- **AND** the message preserves the remediation string
- **AND** no run file is written

#### Scenario: Invalid parent variant is structural

- **GIVEN** a host-resolution call whose parent assistant message
  carries a present-but-invalid variant
- **WHEN** `resolveHostModel()` throws
  `StructuralHostModelError`
- **THEN** the run fails with error code `host_model_structural`
- **AND** `retryable` is false
- **AND** the message names the invalid-variant remediation

#### Scenario: Transient failure keeps the retryable class

- **GIVEN** a host-resolution call whose session-message lookup
  fails transiently
- **WHEN** `resolveHostModel()` throws a plain error that is not
  a `StructuralHostModelError`
- **THEN** the run fails with error code `host_model_unavailable`
- **AND** `retryable` is true
- **AND** the message is the sanitized failure text with no
  remediation suffix
- **AND** no run file is written

#### Scenario: Abort takes precedence over classification

- **GIVEN** a host-resolution call that would throw
  `StructuralHostModelError`
- **WHEN** the cancellation signal aborts the call
- **THEN** the run fails with error code `cancelled`
- **AND** `retryable` is false
- **AND** no run file is written

#### Scenario: Early exit is not a redispatch target

- **GIVEN** a host-source and an advisor dispatch that both
  early-exited during host resolution, so neither wrote a run
  file
- **WHEN** `redispatch_failed_runs` runs for that session
- **THEN** neither early-exited plan entry has a run file to
  target
- **AND** a named agent with no run file reports `not_found`
- **AND** consolidation fails closed with `included plan entry
  <agent>/<sequence> has no run`
- **AND** the final verdict is `INCONCLUSIVE`

#### Scenario: Re-run early exit reports the classified error

- **GIVEN** a failed host-source run with a persisted prompt and
  plan entry
- **WHEN** `redispatch_failed_runs` re-executes it and host
  resolution throws `StructuralHostModelError`
- **THEN** the result entry reports outcome `early_exit`
- **AND** its `error.code` is `host_model_structural`
- **AND** `retryable` is false
- **AND** the error message carries the remediation string
- **AND** the prior run file is restored to its active path

## MODIFIED Requirements

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

A host-source run whose host resolution fails MUST classify the
failure per RD-FR-038 (Host-Resolution Failure Classification): a
transient lookup failure fails with error code
`host_model_unavailable` and `retryable: true`, while a
structural cause fails with error code `host_model_structural`
and `retryable: false`, carrying the shared failure's remediation
string. Both messages MUST pass through the shared
`sanitizeInvocationError()` redaction
(`.opencode/lib/agent-executor.ts`) — the same sanitization the
advisor branch applies, redacting Bearer tokens,
credential-shaped assignments, URL query strings, and home paths.
A host resolution aborted via the cancellation signal MUST fail
with error code `cancelled`. All failure paths are pre-execution
early exits: no run file is written for them, and `retryable`
means caller re-invocation only per RD-FR-038.

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

#### Scenario: Transient host resolution failure

- **GIVEN** a host-source plan entry and a parent session whose
  current message cannot be resolved because of a transient
  lookup failure
- **WHEN** `resolveHostModel()` fails with a non-structural error
- **THEN** the run fails with error code `host_model_unavailable`
- **AND** `retryable` is true
- **AND** the error message is the `sanitizeInvocationError()`
  form of the resolution failure (no Bearer tokens,
  credential-shaped assignments, URL query strings, or home
  paths)
- **AND** no run file is written

#### Scenario: Structural host resolution failure

- **GIVEN** a host-source plan entry and a parent session whose
  current message is not an assistant message, carries no model
  identity, or carries an invalid variant
- **WHEN** `resolveHostModel()` throws
  `StructuralHostModelError`
- **THEN** the run fails with error code `host_model_structural`
- **AND** `retryable` is false
- **AND** the message preserves the remediation string in the
  `<message>. Remediation: <remediation>; retrying will not fix
  this.` shape
- **AND** the message is the `sanitizeInvocationError()` form
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

Previously: the failure clause routed every host-resolution
failure — transient and structural alike — to error code
`host_model_unavailable` with `retryable: true`, and the recorded
message carried only the error's message text, dropping a
structural failure's remediation string.

### Requirement: Advisor Host-Model Fallback

When an advisor run's review-matrix profile omits a `model`
field, `dispatch_agent_run` MUST resolve the parent session's
active model via `resolveHostModel` and use it as the execution
model. The run MUST use host-run provenance semantics: null
requested model provenance with resolved parent model
provenance.

A host-resolution failure on the advisor fallback MUST classify
the failure per RD-FR-038, identically to the host-source branch:
a transient lookup failure fails with `host_model_unavailable`
and `retryable: true`, while a structural cause fails with
`host_model_structural` and `retryable: false`, carrying the
shared failure's remediation string. Both messages MUST pass
through the shared `sanitizeInvocationError()` redaction.

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

#### Scenario: Transient host resolution failure

- **GIVEN** the review matrix profile has no `model`
- **AND** `source: "advisor"`
- **WHEN** `resolveHostModel` fails transiently (session message
  unavailable)
- **THEN** `dispatch_agent_run` returns a failed result with
  error code `host_model_unavailable`
- **AND** `retryable` is `true`
- **AND** the error message is a sanitized availability error
  that MUST NOT leak internal session details (session IDs,
  message IDs, internal API paths)

#### Scenario: Structural host resolution failure

- **GIVEN** the review matrix profile has no `model`
- **AND** `source: "advisor"`
- **WHEN** `resolveHostModel` throws `StructuralHostModelError`
- **THEN** `dispatch_agent_run` returns a failed result with
  error code `host_model_structural`
- **AND** `retryable` is `false`
- **AND** the error message preserves the remediation string
- **AND** no run file is written

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
  RD-FR-037, or fails with the RD-FR-038 classification
  (`host_model_structural` or `host_model_unavailable`)
  — never `tier_model_unavailable`

The `source` field comparison MUST use strict string equality
(`=== "advisor"` for the advisor branch, `=== "host"` for the
host branch per RD-FR-037) with no case normalization,
trimming, or coercion. Unexpected values (empty string, `null`,
`undefined`, case variations like `"Advisor"` or `"ADVISOR"`)
MUST fall through to the `tier_model_unavailable` rejection
path; `source: "host"` is excluded from that fallthrough per
RD-FR-037.

Previously: the requirement's "Host resolution failure" scenario
routed every advisor host-resolution failure to
`host_model_unavailable` with `retryable: true` and a sanitized
message, with no structural class and no remediation
preservation.

## REMOVED Requirements

None.
