---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-10-31T22:10:04Z
identity: review-dispatch-20261031T221004-jay-flowers
tier: draft
---

Finalization validators must not enforce stricter field requirements than the persistence-layer schema they validate against. validateRunModel in review-dispatch rejected successful host and host-resolved advisor runs unless BOTH resolved_parent_model and resolved_parent_variant were non-null, while the writer contract (resolveHostModel in .opencode/lib/agent-executor.ts, absent-variant fallback returning variant: string | undefined) and NullableVariantSchema legitimately produce null variants when the provider reports none (e.g. enmaas/rits/zai-org/glm-5-3). The mismatch fail-closed real dispatches into INCONCLUSIVE verdicts (observed on unbound-force/dewey PR #135). Fix pattern: when relaxing a provenance check, keep the identity-bearing field required (resolved_parent_model stays non-null and cross-checked against reported_model via the expectedModel/model_mismatch machinery) and relax only fields the schema sanctions as null. Test pattern: cover the relaxed path with an end-to-end acceptance test through the full finalizer — the sample fixture (schemas/review-dispatch/samples/sample-review-dispatch.json) defaults resolved_parent_variant to null, so tests that leave the field untouched genuinely exercise the null-variant path including schema validation and artifact persistence; conversely, rejection tests should set the secondary field to a non-null value to prove the rejection is driven by the identity field alone.
