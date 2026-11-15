---
tag: code-review
author: jay-flowers
category: gotcha
created_at: 2026-11-10T20:28:29Z
identity: code-review-20261110T202829-jay-flowers
tier: draft
---

In a child code-review session, the live PR head reported by fetch_pr_review_context can lag the frozen input context's head_sha when the branch carries unpushed local fix commits (observed: frozen head 744a39bb included the effect-security-override commit, while live PR #1006 head 7c012279 predated it — its file list omitted the override files and the plugin-transitive CI check still failed there). Treat the frozen diff as the review scope, and source change-record evidence (e.g., DE-004 delta checks) from the artifacts inside the diff itself rather than from the live PR body alone.
