---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:35:46Z
identity: review-dispatch-20261102T033546-jay-flowers
tier: draft
---

Triage gotcha from unbound-force #658: an issue proposing a leaner intermediary agent (review-orchestrator to replace cobalt-crush-dev in /uf.unleash review dispatch) was invalidated nine days after filing by a better refactor that removed the intermediary entirely — openspec/changes/unleash-direct-review-dispatch (originating_issue 716, all tasks complete 2026-10-14) rewrote Steps 4 and 6 to call plan_review_dispatch + dispatch_review_runs + consolidate_dispatch directly in the parent session, and learning uf-unleash-20261006T004544-jay-flowers records the decision that direct deterministic dispatch beats ANY intermediary agent for review-council invocation. Rule: before validating an agent-replacement issue, grep the command file for the current agent name AND the dispatch tool calls to confirm the delegation path still exists — line-number references in issues go stale after command rewrites (uf.unleash.md grew 791 → 1961 lines across two rewrites). The remaining cobalt-crush-dev delegations (plan, tasks, phase-checkpoint fixes, code-fix application in review loops) are content work where the implementation persona is relevant and deliberately retained.
