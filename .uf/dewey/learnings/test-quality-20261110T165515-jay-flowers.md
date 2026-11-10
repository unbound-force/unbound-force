---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-10T16:55:15Z
identity: test-quality-20261110T165515-jay-flowers
tier: draft
---

Decision-set-driven branch logic needs a decision x state matrix audit before coverage is called done: enumerate every exported set member AND non-member against every state arm. The enum-exclusion cells — the decision NOT in the set — are the cells nobody writes tests for, and 'helpfully' extending the set later passes the whole suite. Verified on unbound-force PR 1007: feedback-replies.ts exports RESOLVABLE_DECISIONS = {accept, modify, file, reply} with ask/reject excluded per FRT-FR-001, but every `ask` test item used comment_kind "issue", so the ask-on-review-thread exclusion (thread stays open) had zero coverage while reject's exclusion was pinned by one happy-path item.
