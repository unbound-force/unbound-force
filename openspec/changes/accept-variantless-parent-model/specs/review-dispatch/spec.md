## MODIFIED Requirements

### Requirement: Finalization Model Validation

Previously: `validateRunModel` required successful host and
host-resolved advisor runs to carry both
`resolved_parent_model` and `resolved_parent_variant` non-null,
rejecting otherwise with a message requiring resolved parent
model and variant.

Modified: `validateRunModel` MUST validate three branches:

1. `source: "host"` — null requested provenance; on success
   `resolved_parent_model` MUST be non-null.
   `resolved_parent_variant` conforms to the variant-or-null
   schema contract (`NullableVariantSchema`) and MAY be null
   when the provider reports no variant.
2. `source: "advisor"` with `requested_model: null` —
   host-resolved advisor, validated identically to host runs
   (null requested provenance; on success
   `resolved_parent_model` MUST be non-null;
   `resolved_parent_variant` MAY be null).
3. All other sources — unchanged (non-null requested
   provenance, null resolved provenance)

The variant-or-null sanction matches the writer contract of
`resolveHostModel()` in `.opencode/lib/agent-executor.ts`
(commit 716f2036, absent-variant fallback): resolution
succeeds with no variant when the provider does not populate
one, and the persistence layer writes `host.variant ?? null`.
A successful host or host-resolved advisor run with a null
`resolved_parent_model` MUST be rejected with a message
containing "requires resolved parent model".

This delta supersedes the over-strict both-non-null text in
the unarchived `advisor-host-model-fallback` and
`review-council-multi-model-fanout` review-dispatch deltas at
archive time; those changes are not edited in place.

#### Scenario: Host-resolved advisor passes validation without a variant

- **GIVEN** a finalization payload with a successful run where
  `source: "advisor"` and `requested_model: null`
- **AND** `resolved_parent_model` is non-null
  (`enmaas/rits/zai-org/glm-5-3`, observed in the wild)
- **AND** `resolved_parent_variant` is null because the
  provider reports no variant
- **AND** the plan entry has `model: null` and `variant: null`
- **WHEN** `finalize_review_dispatch` validates the run
- **THEN** validation succeeds

#### Scenario: Host run passes validation without a variant

- **GIVEN** a finalization payload with a successful run where
  `source: "host"`
- **AND** `resolved_parent_model` is non-null and
  `resolved_parent_variant` is null
- **WHEN** `finalize_review_dispatch` validates the run
- **THEN** validation succeeds

#### Scenario: Host-resolved advisor without resolved parent model fails

- **GIVEN** a successful run with `source: "advisor"` and
  `requested_model: null`
- **AND** `resolved_parent_model` is null
- **WHEN** `finalize_review_dispatch` validates the run
- **THEN** validation fails with a message containing
  "requires resolved parent model"
