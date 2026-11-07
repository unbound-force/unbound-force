---
tag: review-process
author: jay-flowers
category: pattern
created_at: 2026-11-07T16:25:13Z
identity: review-process-20261107T162513-jay-flowers
tier: draft
---

Maintainer feedback items that bundle an in-PR content fix with a prospective CI gate must be split by workflow tier at triage: the content repair lands in the current PR (docs-only work needs no spec), while the CI gate routes to a separate spec-governed change because AGENTS.md's "What requires a spec" list includes CI changes and the commit-scope rule forbids mixing CI changes into a docs chore branch. The deferral is sound only with a tracked follow-up (issue or opened change referencing the review): on unbound-force, the learning-batch truncation defect recurred from PR #949 to PR #968 precisely because the first repair (cd603598) completed the sentences without shipping the requested gate. When assessing such feedback, also verify whether the literal ask is recoverable before planning the fix — a re-export instruction is moot when the source sessions and dispatch artifacts exist only on the authoring machine.
