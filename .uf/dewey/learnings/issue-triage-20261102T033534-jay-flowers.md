---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:35:34Z
identity: issue-triage-20261102T033534-jay-flowers
tier: draft
---

Triaging a prose-consolidation issue filed before a governance amendment requires validating the problem while re-routing the mechanism. Issue #661 (unbound-force/unbound-force, filed 2026-10-05) proposes deduplicating the session-resume guard, MANDATORY GATE blocks, and shell-injection temp-file patterns into .opencode/skills/session-resume/SKILL.md. The duplication is real and measurable (DE-004 token deltas; guard/gate text greppable across uf.review-council, uf.review-pr, uf.address-feedback, uf.triage-issue, uf.finale, uf.unleash), but three governance layers post-date or coexist with the filing and constrain the mechanism: (1) Constitution Principle VI / determinism pack DE-001 and DE-003 route the mechanical temp-file pattern and the prescribed gate output to executable tools with tests (precedent: uf_finale_create_pr's secure temp file, format_unleash_exit's rendered blocks), leaving only the guard's judgment guidance unambiguously skill-prose; (2) shipped openspec requirements (harden-implement-gates, port-uf-rebase-pr command conventions) mandate session-resume guard language structurally present in command files (CRITICAL RULE block within 5 lines of gate instructions), so replacing inline blocks with one-line skill references must preserve or explicitly amend those requirements; (3) scaffold drift-detection means the dedup sweep must cover internal/scaffold mirrors, not just .opencode/commands (extends learning documentation-pipeline-20261019T202640-jay-flowers). Triage outcome: VALID problem, enhancement, objective — with mechanism reconciliation assigned to the spec phase rather than invalidating the issue.
