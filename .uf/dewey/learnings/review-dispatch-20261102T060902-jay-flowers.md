---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-02T06:09:02Z
identity: review-dispatch-20261102T060902-jay-flowers
tier: draft
---

The #901 host-fallback provenance fix (unbound-force review-dispatch plugin, new requirement RD-FR-037, branch opsx/host-fallback-provenance, 2026-11-02) established the pattern for dispatcher/finalizer contract alignment: whenever the dispatch core gains a new run-source kind (source "host" joining "explicit" and "advisor"), the dispatcher must record exactly the provenance shape the finalizer's validateRunModel enforces for that source — for host runs, null requested_model/provider/model_id/variant plus a non-null resolved_parent_model after successful host lookup (RD-FR-008 as modified by the accept-variantless-parent-model change). The defect was dispatcher-side only; the finalizer's invariants were already correct, so the fix mirrored the #693 advisor branch's shared resolveHostModel() inheritance into a source === "host" branch (review-matrix profile bypass, sanitized retryable host_model_unavailable on resolution failure, cancelled on abort) rather than touching the finalizer or planner. Two operational notes for future applications of this pattern: keep the live plugin .opencode/plugins/review-dispatch/index.ts and the scaffold asset internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts byte-identical (the Go drift gate TestEmbeddedAssets_MatchSource enforces it; batch the mirror sync into the scaffold phase and accept an intentionally red drift gate between the live edit and the sync), and pin the behavior at the unit seam with a provenance-carrying consolidation test over a scratch dispatch session that includes a host-source run file — that test is the reproduction of the "dispatchable but unfinalizable" failure class that no per-run test can catch.
