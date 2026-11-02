---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:35:17Z
identity: issue-triage-20261102T013517-jay-flowers
tier: draft
---

In unbound-force repos, adding rules to a convention pack under .opencode/uf/packs/ is a user-facing change: review dispatch auto-discovers every pack and includes it in all Divisor review prompts, so a new rule (e.g., determinism.md DE-005/DE-006, Intent-Level Orchestration) alters repo-wide review-council behavior. The AGENTS.md documentation-issue gate therefore applies, with established precedent of one docs-tracking issue per governance change filed against the current repo (determinism-codification filed #720; router-pattern-codification filed #762, referencing originating issue #758). determinism.md self-identifies as a repo-local governance pack (not a scaffold asset deployed to hero repos), so the current-repo documentation gate — not the org constitution's cross-repo website gate — is the applicable one. When triaging such issues, verify the cited rules exist in the current pack and spec delta, confirm no duplicate docs issues exist, and treat the either/or acceptance criterion (docs mention the new rules or explicitly scope to the prior rule set) as objectively verifiable.
