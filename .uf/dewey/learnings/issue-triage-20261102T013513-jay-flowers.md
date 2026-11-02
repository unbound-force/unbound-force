---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:35:13Z
identity: issue-triage-20261102T013513-jay-flowers
tier: draft
---

Unbound Force meta repo triage pattern: additions to auto-discovered convention packs (.opencode/uf/packs/*.md) are user-facing governance changes, not spec artifacts. Review dispatch loads every pack in that directory into each Divisor review prompt, so new rules (e.g., determinism.md DE-005/DE-006, Intent-Level Orchestration) alter repo-wide review-council behavior; the AGENTS.md documentation-issue gate (current-repo issue before PR merge) therefore applies, and the internal-refactoring / test-only / CI-only / spec-artifact exemptions do not. Precedent: #720 (determinism-codification) and #762 (router-pattern-codification), same artifact class. Two execution corollaries: (1) docs issues scoped to a pack version drift quickly — determinism.md moved 1.1.0 → 1.4.0 (adding DE-007/DE-008) within two weeks of #762 being filed — so documentation passes should enumerate the live rule inventory or scope their listings explicitly rather than target a frozen version; (2) verify the docs surface before assuming work remains — repo greps may show no user-facing docs enumerate the rule inventory at all, leaving the CHANGELOG entry as the substantive record.
