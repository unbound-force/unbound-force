---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-01T21:57:05Z
identity: review-dispatch-20261101T215705-jay-flowers
tier: draft
---

When a producer contract legitimately makes a provenance field optional, downstream fail-closed validators must be updated in lockstep or they reject valid data: unbound-force's review-dispatch finalizer rejected successful host and host-resolved-advisor runs whose resolved_parent_variant was null — dewey PR #135 reported INCONCLUSIVE with "requires resolved parent model and variant" under the variantless parent model enmaas/rits/zai-org/glm-5-3 — while the writer contract, resolveHostModel() in .opencode/lib/agent-executor.ts, deliberately omits the variant when the provider reports none. The validator predated the writer contract and was the stale artifact. Fail-closed provenance validation is the correct posture, but it must distinguish "optional field legitimately absent" from "required provenance missing": a control that rejects legitimate runs erodes trust in its own INCONCLUSIVE signals. Fix pattern (accept-variantless-parent-model, PR #898): relax only the genuinely optional field, keep the hard requirement on the required one (null resolved_parent_model still fails with "requires resolved parent model"), and mirror the corrected validation into the scaffold copy.
