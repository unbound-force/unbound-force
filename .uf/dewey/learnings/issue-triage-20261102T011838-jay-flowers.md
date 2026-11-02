---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:18:38Z
identity: issue-triage-20261102T011838-jay-flowers
tier: draft
---

Triage of follow-up issues filed from review-council advisories must be grounded in the current tree, not the issue's filing-time state: the advisory's fix may have already landed via the parent change's PR, which typically cites only the parent issue number in code comments and closing keywords — leaving the follow-up issue open, untriaged, and unlinked after resolution. Issue #771 (verify replacement run-file persistence before reporting redispatched, filed 2026-10-20) is the precedent: PR #776 landed the exact proposed fix the same day (ExecuteDispatchRunOutcome.persisted flag, redispatch_failed_runs persistence gate + prior-file restore, write-temp-then-rename, dispatch-recovery.test.ts persistence-failure test) while its code comments cite only parent issue #751. A landed fix does not make the report INVALID — the verdict stays VALID (the report was accurate and actionable when filed) and the recommended disposition becomes close-as-completed with a link to the landing PR, after verifying the PR's closing keywords actually referenced the follow-up. Rule: before ruling on any advisory-derived issue, grep the tree for the proposed fix's artifacts (flag names, gate comments, test titles); resolution status changes the maintainer action, not the verdict.
