---
tag: address-feedback
author: jay-flowers
category: pattern
created_at: 2026-11-16T15:42:05Z
identity: address-feedback-20261116T154205-jay-flowers
tier: draft
as_of: "unbound-force@c2dfd4b8f15f9ae106c599aef9a0418bf075dd93"
---

When a documentation change must define a behavioral terminal state (quick branch quick/address-feedback-skip-branch-terminal-state, 2026-11-16), ground the definition in three sources before writing prose: the tool implementation (.opencode/lib/feedback-replies.ts — postFeedbackReplies computes complete = posted===total && resolved===resolvable and records the reply-comments checkpoint only when complete; joinFiledIssues renders failed issue filings as nulls "recording reality rather than an error"; write_feedback_triage_artifact clears the command's progress state on success, making the artifact write terminal), the test suite (.opencode/test/feedback-replies.test.ts pins the skip-branch item shape reply "issue_failed"/issue "failed"/thread "reply_not_posted" and the null-pass-through artifact rendering), and the pre-extraction git history (git show 923ef221~1:.opencode/commands/uf.address-feedback.md — the original hand-written Section 4.7 said FILE fields come "from the issue_number and issue_url that post_feedback_replies returned for it", settling that the docs describe tool behavior, not aspiration). The definition that emerged — a complete: false result ends the run at Section 4.6 with no artifact and no progress clear, and the artifact's FILE fields render only from complete runs (joined identity) or the skip-posting path (null/null with a decision-based filed count) — passed the review council with no requested changes precisely because every clause was tool-verifiable.
