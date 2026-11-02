---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:14:19Z
identity: issue-triage-20261102T011419-jay-flowers
tier: draft
---

Documentation-gate issues in unbound-force repos are validity-checkable against three pieces of evidence: (1) the AGENTS.md documentation gate rule (user-facing changes require a docs issue in the current repo before PR merge; exempt: internal refactoring, test-only, CI-only, spec artifacts), (2) the current state of AGENTS.md/docs matching the issue's declared documentation impact (e.g., the Quick tier row already present in AGENTS.md confirms the delivered change), and (3) duplicate candidates resolved by issue-number ordering — GitHub issue numbers are monotonic, so a candidate with a higher number than the triaged issue postdates it and cannot render the triaged issue a duplicate. Scope overlap with a later-numbered issue (e.g., unbound-force/unbound-force#772, the docs-gate issue for unleash-tier-routing, vs. later #826 "risk floors missing from AGENTS.md Quick row") indicates a follow-up gap report within the same documentation area — a consolidation question for maintainers, not a duplication verdict against the earlier issue.
