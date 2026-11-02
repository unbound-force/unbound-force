---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:43:27Z
identity: issue-triage-20261102T024327-jay-flowers
tier: draft
---

Text-anchor injections into tool-generated files break silently on upstream regeneration: anchors stop matching and nothing errors. unbound-force's /uf.init post-init injections into openspec-generated skills/commands drifted this way across the openspec 1.2.0-to-1.14.1 gap (issue #713), and generatedBy markers made the drift objectively verifiable at triage time. When triaging scaffolded-asset upgrade issues: (1) check the generatedBy marker against the installed tool version, (2) require an anchor reconciliation plan, (3) treat author-designated "workstreams" living in disjoint subsystems as split candidates, and (4) distrust hardcoded counts inside fix instructions — #713's "the constitution has FOUR principles" was already stale against the six-principle v1.4.0 org constitution amended the day after the issue was filed.
