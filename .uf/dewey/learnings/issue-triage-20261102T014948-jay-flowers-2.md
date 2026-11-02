---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:49:48Z
identity: issue-triage-20261102T014948-jay-flowers-2
tier: draft
---

Bug-issue triage must verify cited mechanics semantically against current HEAD, not positionally: line-number citations rot as files evolve, and parallel fix changes can remediate an open issue's mechanism while the issue stays open. Verifying unbound-force#746 (review-dispatch one dispatch session per conversation vs. uf.address-feedback per-item Tier 2 cycles), the issue's cited index.ts:2714-2733 at f0327a9 resolved to unrelated finalization code at current HEAD; grepping the mechanism (isFirstCall, session-metadata persistence) plus a Dewey prior-learnings search surfaced the remediation record: fix-review-dispatch-correlation-release (#794 design D1 — mint a fresh correlation per dispatch_review_runs invocation and re-point the session map) and fix-review-dispatch-recovery (design D5 — persist session_metadata whenever supplied, last-writer-wins), both confirmed in current code comments, with the consuming command updated in lockstep (uf.address-feedback §2.3.4 now captures each item's correlation_id and forwards it unconditionally to consolidation). Structural rule: a tool-contract × command-contract conflict is fully remedified only when both sides of the pair change together. A bug that was valid when filed stays VALID in triage after remediation — route it to closure via the fixing change instead of queueing duplicate fix work.
