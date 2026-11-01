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

The review council's fix loop drove this record to its final
shape: iteration 1 returned REQUEST CHANGES with two HIGH
findings — spec-governance (the unreconciled requirement record)
and documentation-pipeline (the missing documentation-issue
reference) — and iteration 2 approved the code while requesting
artifact remediations: a mechanically sound supersession
strategy, both documentation-gate issue citations, and the
convention-pack artifact set. Iteration 3 found the FMV half of
that strategy unsound — repo-wide, no delta ever ADDEDs
'Finalization Model Validation', so the MODIFIED entry was an
orphan that could never archive, and its scenario names shared
nothing with the fallback delta's — driving the final reshape:
FMV is ADDED (established outright, complete scenario set), and
only `review-council-multi-model-fanout` remains in the
archive-ordering constraint. This change is the fix-loop
response.

Workflow-record note: this OpenSpec artifact set rides the
quick-tier branch `quick/accept-variantless-parent-model`
because the council's fix loop demanded a durable spec record
after `/uf.unleash` tier routing had already created the
branch; re-routing would have orphaned the fix-loop history.
The branch-prefix-versus-artifact-shape divergence is
documented here as council-demanded escalation, and
normalizing it — recording a mid-flight tier-escalation path
in the AGENTS.md tier workflow, or routing council-mandated
spec records via `opsx/` branches — is a separate future
governance change, out of scope for this fix loop.

## What Changes

Spec record only — the code was already corrected and
council-reviewed: `validateRunModel` now rejects successful
host and host-resolved advisor runs ONLY when
`resolved_parent_model` is null, and the failure message reads
"requires resolved parent model". This change carries one ADD
and one MODIFY delta on the review-dispatch capability so the
unarchived requirement record matches the variant-or-null
schema contract: 'Finalization Model Validation' is ADDED,
establishing the corrected branches, scenario set, and
failure-message expectation outright — repo-wide no delta ever
ADDEDs that requirement, so no predecessor text can ever
materialize for a MODIFY to apply against, and the same-named
MODIFIED entry in the `advisor-host-model-fallback` delta is
itself an orphan (a pre-existing defect in that change,
documented but out of scope) — and 'Review Dispatch Artifact'
(RD-FR-008) is MODIFIED, restated in full with the corrected
host-lookup provenance clause (design.md D3). The MODIFIED
half supersedes per requirement name: it replaces the
over-strict host-lookup text the
`review-council-multi-model-fanout` delta carries under its
ADDED requirement of that name (design.md D2), and it is
archive-order dependent, so this change MUST archive after
`review-council-multi-model-fanout` (neither older change is
edited in place). The CHANGELOG entry references documentation
issues #891 and unbound-force/website#360 recording the
docs-gate outcome.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `review-dispatch`: finalization model validation for
  successful host and host-resolved advisor runs requires a
  non-null `resolved_parent_model` only;
  `resolved_parent_variant` conforms to the variant-or-null
  schema contract and MAY be null when the provider reports
  no variant. The 'Review Dispatch Artifact' requirement
  (RD-FR-008) carries the same correction in its host-lookup
  provenance clause.

### Removed Capabilities
- None.

## Impact

- **Spec record**: `openspec/changes/accept-variantless-
  parent-model/specs/review-dispatch/spec.md` — ADDED
  'Finalization Model Validation' and MODIFIED 'Review Dispatch
  Artifact' (RD-FR-008, restated in full with the corrected
  host-lookup provenance clause); design.md records the
  decisions (D1 variant-only relaxation, D2 ADDED FMV
  establishment plus same-name RDA supersession and archive
  ordering, D3 full restatement)
- **No code changes**: the reviewed implementation (plugin,
  scaffold mirror, finalize error-boundary tests) stays
  untouched; this change records its contract
- **Establishment, supersession, and archive ordering**: the
  ADDED 'Finalization Model Validation' entry establishes the
  corrected requirement outright — no delta ever ADDEDs that
  requirement, and the same-named MODIFIED entry in
  `advisor-host-model-fallback` is itself an orphan (that
  change's own defect, out of scope); the MODIFIED 'Review
  Dispatch Artifact' entry supersedes the over-strict
  host-lookup text in `review-council-multi-model-fanout`'s
  ADDED requirement of that name, per requirement name and
  archive-order dependent — this change MUST archive after
  `review-council-multi-model-fanout` (constraint stated in
  the delta preamble and design.md D2); neither older change
  is edited in place
- **CHANGELOG.md**: Unreleased/Fixed entry references
  documentation issues #891 and unbound-force/website#360
- **Documentation impact**: documentation issue #891
  (https://github.com/unbound-force/unbound-force/issues/891)
  and unbound-force/website tracking issue #360
  (https://github.com/unbound-force/website/issues/360) record
  the docs-gate outcome for the user-facing behavior
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

Every scenario in the ADDED 'Finalization Model Validation'
entry maps to an existing green test — the variantless host
and host-resolved-advisor passes and the null-model rejection
for both sources in
`.opencode/test/finalize-error-boundaries.test.ts`, and the
variant-present host-resolved-advisor pass in
`.opencode/test/finalize-review-dispatch.test.ts`; the 'Review
Dispatch Artifact' scenarios restated from the fanout delta
map to the existing artifact and finalization suites
unchanged. The record adds no untestable behavior.

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
