---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-25T21:37:00Z
identity: review-dispatch-20261125T213700-jay-flowers
tier: draft
as_of: "unbound-force@b7cade85f9611633678952880c8b75fa28393632"
---

In dispatched PR reviews, an instruction-lint run with --base <PR base> --changed that returns an empty delta.files while the caller's diff names instruction files means the working tree is not at the PR head — observed on unbound-force PR #1228, where the primary checkout sat at the merge base while pre-flight had linted in a detached worktree. An empty delta also scopes --changed findings to nothing, silently hiding IL-040/IL-041/IL-042 hints on the changed files (0 findings at the base tree vs 3 IL-040 hints after repositioning). Reposition with checkout_pr_head (it verifies HEAD against the frozen head SHA and never discards local commits), re-run the evidence commands, treat the re-run as the evidence of record, and disclose both runs in the review.
