---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:14:32Z
identity: issue-triage-20261102T011432-jay-flowers
tier: draft
---

Triage discrimination for review-dispatch context-exhaustion issues requires mechanism-level comparison, not topic-level similarity: manifest/on-demand diff delivery (issue #880, RD-FR-032) bounds the diff channel of reviewer prompts; per-run prompt and sibling-evidence budget trimming (issue #770) bounds assembly volume; usage-aware skipping (budget_usd, extended to token budgets) bounds execution cost. Searching "review dispatch context too large" surfaces all three as candidates, but they are complementary rather than duplicative. The strongest validity signal for a deferred-follow-up issue is triple provenance — parent incident issue, the deferring change's Non-Goals entry, and the shipped enabling prerequisite — because each claim cross-checks against repository artifacts (D6 usage reporting in dispatch_status/consolidate_dispatch, budget_skipped run status, the fix-review-dispatch-recovery Non-Goals) without requiring implementation knowledge.
