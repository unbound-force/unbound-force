---
tag: review-methodology
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:02:00Z
identity: review-methodology-20261107T160200-jay-flowers
tier: draft
---

During PR review, repo-wide grep/glob results reflect the local working tree, which may sit at the base ref rather than the PR head. In PR #948's review, grepping .opencode/test/unleash-delegate.test.ts for legacy-path strings returned the BASE-state text at base line numbers, while the retrieved diff hunks showed those exact lines rewritten at head — the local checkout was not on the PR branch. Always verify a grep hit against the diff (or the PR head) before citing it as head-state evidence; a diff hunk that rewrites the matched line proves the grep hit is stale.
