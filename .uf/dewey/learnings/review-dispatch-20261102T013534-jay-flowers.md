---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:35:34Z
identity: review-dispatch-20261102T013534-jay-flowers
tier: draft
---

Issue #760's two-dispatch failure mode (a second review dispatch in one opencode session reusing the first's correlation ID; session_metadata persisted first-call-only; consolidate_dispatch failing closed with "artifact commit must equal the immutable reviewed head SHA" plus "run has no included plan entry") is REMEDIATED at current HEAD — the manual session-metadata.json rewrite recovery described in earlier review-dispatch learnings (2026-10-15, 2026-10-19) is obsolete and MUST NOT be applied. Both root-cause components landed: (1) fix-review-dispatch-recovery design D5 — .opencode/plugins/review-dispatch/index.ts ~3467-3482 persists session_metadata whenever supplied, replacing the earlier file (no isFirstCall gate); (2) fix-review-dispatch-correlation-release design D1 (issue #794, landed 2026-10-23) — dispatchReviewRuns mints a fresh correlation per invocation and re-points sessionCorrelationMap immediately after input validation, before any session-creating side effect (index.ts ~5588-5603), so each workflow from one conversation gets its own session dir and session-less outcomes fail closed as not_found. Tests cover the mint in four suites (.opencode/test/dispatch-review-runs.test.ts, dispatch-agent-run.test.ts, consolidate-dispatch.test.ts, dispatch-status.test.ts); openspec/changes/fix-review-dispatch-correlation-release/ and CHANGELOG record the change. Triage rule: recovery-procedure learnings phrased as "until issue #N is fixed, apply this workaround" must be re-verified against current code before use — semantic search surfaces them with high similarity long after the fix lands, and applying an obsolete manual-recovery procedure (rewriting a dispatch session's metadata by hand) to already-fixed state is itself a data-integrity risk.
