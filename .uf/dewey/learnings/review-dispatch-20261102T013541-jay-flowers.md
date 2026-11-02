---
tag: review-dispatch
author: jay-flowers
category: context
created_at: 2026-11-02T01:35:41Z
identity: review-dispatch-20261102T013541-jay-flowers
tier: draft
---

Issue #760 (review-dispatch: second dispatch in one opencode session reusing the first's correlation ID and stale session-metadata) is remediated in unbound-force HEAD as verified 2026-11-01. .opencode/plugins/review-dispatch/index.ts now (1) mints a fresh correlation ID per dispatchReviewRuns aggregate invocation and re-points sessionCorrelationMap before prompt building (design D1 of fix-review-dispatch-correlation-release, issue #794), so a second workflow in the same conversation gets its own dispatch session, and (2) persists session_metadata whenever supplied, replacing the earlier session-metadata.json file (design D5 of fix-review-dispatch-recovery) — together implementing both of the issue's suggested fix directions. Direct dispatch_agent_run calls still share one correlation per parent session by design, now safe because metadata refreshes on every supply. The pre-fix failure signature (consolidate_dispatch failing closed with "artifact commit must equal the immutable reviewed head SHA" plus "run has no included plan entry" after a second dispatch) and the manual recovery procedure are recorded in learning/review-dispatch-20261019T182401-jay-flowers. The issue was still open per the 2026-10-25 Dewey index — it is a close candidate once the fixing changes are confirmed shipped and verified in a live two-dispatch session.
