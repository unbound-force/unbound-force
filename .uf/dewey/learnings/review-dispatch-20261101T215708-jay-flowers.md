---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-01T21:57:08Z
identity: review-dispatch-20261101T215708-jay-flowers
tier: draft
---

A validation gate that encodes a stricter contract than the writer producing the data is a stale artifact that fails closed on legitimate input. In unbound-force review-dispatch, resolveHostModel() (.opencode/lib/agent-executor.ts, commit 716f2036) deliberately returns variant: undefined when the provider reports no variant, but validateRunModel still required both resolved_parent_model and resolved_parent_variant non-null — so host and host-resolved-advisor runs from variantless providers (e.g. enmaas/rits/zai-org/glm-5-3) fail-closed at finalize_review_dispatch and misreport INCONCLUSIVE (observed on unbound-force/dewey PR #135; fixed by the accept-variantless-parent-model change). Structural lesson: when one logical contract is implemented in two places (writer and validator), the halves drift independently — regression-test them together, and on conflict the writer's contract is the reference; the validator is the stale artifact. Triage verification pattern for docs-gate issues that record a fix's user-facing deltas (applied to issue #891, 2026-11-01): grep the claimed message change in both the live plugin and its scaffold mirror, read the writer function's signature/JSDoc, and confirm the change's OpenSpec artifacts and CHANGELOG entry cite the issue — every #891 claim verified against .opencode/plugins/review-dispatch/index.ts:2323/2332, the byte-identical scaffold mirror, and the Unreleased CHANGELOG entry (lines 1939-1955).
