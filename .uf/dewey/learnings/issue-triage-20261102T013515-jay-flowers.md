---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:35:15Z
identity: issue-triage-20261102T013515-jay-flowers
tier: draft
---

In unbound-force/unbound-force, convention-pack rule additions are user-facing for triage purposes even with no CLI or agent-surface change: review dispatch auto-discovers every pack under .opencode/uf/packs/ into each Divisor review prompt, so new MUST/SHOULD rules alter repo-wide review-council behavior (DE-005/DE-006 shipped via PR #768). The AGENTS.md documentation gate is satisfied by a current-repo tracking issue per the #720 precedent (#762 for router-pattern-codification), and such an issue's acceptance criteria attach to the rules themselves, not the pack version cited at filing time — determinism.md has since advanced to v1.4.0 (adding DE-007/DE-008) without invalidating the issue's DE-005/DE-006 coverage ask.
