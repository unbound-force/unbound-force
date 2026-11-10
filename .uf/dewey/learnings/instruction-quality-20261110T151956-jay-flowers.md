---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-10T15:19:56Z
identity: instruction-quality-20261110T151956-jay-flowers
tier: draft
---

When mechanical choreography is extracted from a command into a tool, audit the command's residual prose for subset-scoped instructions that predate the tool. An old skip/filter rule ('items marked as fully executed (comment-posted) in the cache MUST be skipped') tends to be rewritten as a pass-through rule ('pass items whose replies were already posted to the tool again unchanged') that names only the previously-filtered subset as the object of the verb — inviting a restrictive reading that drops the rest of the input set. The tool then computes completion over the passed subset only (complete = posted === passed-total), records its checkpoint, and silently loses the un-passed items. Fix pattern: scope the instruction to the full confirmed list and call out the previously-filtered subset as included ('pass the full confirmed item list again, unchanged — including items whose replies were already posted'). Observed in PR #1007, uf.address-feedback.md Section 1.7 vs postFeedbackReplies in .opencode/lib/feedback-replies.ts.
