---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-10T15:19:39Z
identity: documentation-pipeline-20261110T151939-jay-flowers
tier: draft
---

Verifying a two-repo documentation gate: the in-repo docs issue's body is itself evidence of the website half's scope. When the in-repo gate issue (e.g., unbound-force/unbound-force #1003 for PR #1007) enumerates website pages that still describe the pre-change behavior, the website-repo tracker is required content, not optional — its absence is a gate failure even though the current-repo half is satisfied. Search the website repo (unbound-force/website) across all states by three term classes, because website gate issues are titled independently of the change name: the change name (feedback-reply-tools), the tool or schema surface names (post_feedback_replies, feedback-triage), and the linked numbers (PR number, in-repo issue number, originating issue numbers). Zero matches across all three confirms the missing tracker; a title-only or open-state-only search can miss a filed-elsewhere or recently-closed tracker. Applied on PR #1007: in-repo #1003 satisfied the AGENTS.md current-repo gate while the constitution's Cross-Repo Documentation half had no tracker in any state.
