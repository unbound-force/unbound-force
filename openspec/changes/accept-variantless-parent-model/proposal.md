## Why

`validateRunModel` in review-dispatch finalization required
successful host and host-resolved advisor runs to carry both
`resolved_parent_model` and `resolved_parent_variant` non-null,
rejecting with "requires resolved parent model and variant".
That validator predated the variant-optional writer contract
from `resolveHostModel()` (commit 716f2036, absent-variant
fallback): resolution legitimately succeeds with no variant
when the provider does not populate one, and the persistence
layer writes `host.variant ?? null` under
`NullableVariantSchema`. A provider such as
`enmaas/rits/zai-org/glm-5-3` therefore produces dispatches
that finalization fail-closed on, yielding an INCONCLUSIVE
verdict — observed in the wild on an unbound-force/dewey PR
review (#135).

The requirement record made it worse: 'Finalization Model
Validation' lives as the completed-but-unarchived
`advisor-host-model-fallback` delta (no synced
`openspec/specs/review-dispatch/` exists, so the unarchived
delta stack is the only requirement record), and it was
authored from the over-strict validator. The
`review-council-multi-model-fanout` delta carries the same
over-strict both-non-null text. A future archive sync would
propagate the stale MUST into a canonical spec contradicting
the code, and an implementer working from the deltas would
reintroduce the fail-closed bug.

The review council on `quick/accept-variantless-parent-model`
returned REQUEST CHANGES with two HIGH findings:
spec-governance (the unreconciled requirement record) and
documentation-pipeline (the missing documentation-issue
reference). This change is the fix-loop response.

## What Changes

Spec record only — the code was already corrected and
council-reviewed: `validateRunModel` now rejects successful
host and host-resolved advisor runs ONLY when
`resolved_parent_model` is null, and the failure message reads
"requires resolved parent model". This change adds a minimal
MODIFY delta on 'Finalization Model Validation' so the
unarchived requirement record matches the variant-or-null
schema contract, states the corrected failure-message
expectation, and supersedes the over-strict both-non-null text
in the `advisor-host-model-fallback` and
`review-council-multi-model-fanout` deltas at archive time
(neither old change is edited in place). The CHANGELOG entry
references documentation issue #891 recording the docs-gate
outcome.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `review-dispatch`: finalization model validation for
  successful host and host-resolved advisor runs requires a
  non-null `resolved_parent_model` only;
  `resolved_parent_variant` conforms to the variant-or-null
  schema contract and MAY be null when the provider reports
  no variant.

### Removed Capabilities
- None.

## Impact

- **Spec record**: `openspec/changes/accept-variantless-
  parent-model/specs/review-dispatch/spec.md` — MODIFIED
  'Finalization Model Validation'
- **No code changes**: the reviewed implementation (plugin,
  scaffold mirror, finalize error-boundary tests) stays
  untouched; this change records its contract
- **Supersession**: the over-strict text in the unarchived
  `advisor-host-model-fallback` and
  `review-council-multi-model-fanout` deltas is superseded at
  archive time, not edited in place
- **CHANGELOG.md**: Unreleased/Fixed entry references
  documentation issue #891
- **Documentation impact**: documentation issue #891
  (https://github.com/unbound-force/unbound-force/issues/891)
  records the docs-gate outcome for the user-facing behavior
  restoration

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.4.0).

### I. Autonomous Collaboration

**Assessment**: PASS

The change reconciles a requirement record with the corrected
artifact contract. It introduces no runtime coupling; it edits
spec text and a changelog entry only, and communicates the
corrected contract through the existing unarchived-delta
stack that becomes the canonical spec at archive time.

### II. Composability First

**Assessment**: PASS

The corrected requirement keeps review-dispatch's finalization
contract aligned with the shared `resolveHostModel()` writer in
`agent-executor.ts` rather than diverging from it. No new
dependencies; each hero remains independently usable.

### III. Observable Quality

**Assessment**: PASS

The requirement now states exactly what the machine-parseable
run records carry: `resolved_parent_model` non-null on
success, `resolved_parent_variant` variant-or-null per the
persisted schema. The failure message expectation
("requires resolved parent model") matches the shipped
validator output, so spec, code, and tests observe one
contract.

### IV. Testability

**Assessment**: PASS

Every scenario in the delta maps 1:1 to an existing green test
in `.opencode/test/finalize-error-boundaries.test.ts`
(variantless host pass, variantless host-resolved advisor
pass, null-model rejection for both sources). The record adds
no untestable behavior.

### V. Security by Default

**Assessment**: PASS

The essential identity control is unchanged:
`resolved_parent_model` remains required non-null and is
cross-checked against `reported_model`. The change relaxes
only the variant half of a redundant both-non-null check on
tool-owned local artifacts; it introduces no new privilege or
input paths.

### VI. Deterministic Execution

**Assessment**: PASS

The change records an existing deterministic validation rule;
it adds no procedural prose requiring LLM re-reasoning. The
spec record itself is the deterministic artifact that
prevents a future implementer from reintroducing the
fail-closed bug from stale deltas.
