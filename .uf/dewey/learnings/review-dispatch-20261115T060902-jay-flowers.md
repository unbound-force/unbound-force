---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-15T06:09:02Z
identity: review-dispatch-20261115T060902-jay-flowers
tier: draft
as_of: "unbound-force@e5e077b35f0c7984517024431e696737e330b080"
---

appendSpecReviewIterationToState's heal loop (which reconciles dispatch artifacts not yet present in progress state) runs BEFORE the caller's explicit append, so on the second and later consolidations of the same correlation the just-persisted artifact is appended by the heal pass and then again by the explicit append — duplicating the iteration. Fix pattern: make the append idempotent by correlation — after healing, if the entry's correlation_id is already recorded, write back the healed state and return the already-recorded iteration instead of appending a duplicate. Consolidation persists the artifact before its state write, so the heal path is guaranteed to observe every freshly finalized correlation. Verified in .opencode/test/spec-review-gate.test.ts (idempotent re-consolidation) on opsx/harden-spec-review-gate (2026-11-15).
