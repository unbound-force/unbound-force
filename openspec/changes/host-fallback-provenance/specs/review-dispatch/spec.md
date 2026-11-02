<!-- capability: review-dispatch
     change: host-fallback-provenance
     originating_issue: 901 -->

Requirement numbering: RD-FR-037 is allocated for this change.
RD-FR-038 through RD-FR-048 are reserved for future
review-dispatch changes. The dispatch-agent-run tool-surface
requirements live in the `review-dispatch` capability (RD-FR-001
through RD-FR-036, allocated by earlier changes); the
model-matrix structure requirements live in
`specs/model-matrix/spec.md` under the `model-matrix` capability.

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
error code `host_model_unavailable` and `retryable: true`, with a
sanitized error message that MUST NOT leak internal session
details (session ids, message ids, internal API paths). A host
resolution aborted via the cancellation signal MUST fail with
error code `cancelled`. Both failure paths are pre-execution
early exits: no run file is written for them.

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
- **AND** the error message is sanitized and leaks no internal
  session details
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

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
