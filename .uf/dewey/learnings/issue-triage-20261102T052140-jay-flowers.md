---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T05:21:40Z
identity: issue-triage-20261102T052140-jay-flowers
tier: draft
---

Composability claims verify differently at three layers, and triage must check each independently: (1) runtime coupling — unbound-force agent personas genuinely degrade gracefully when peers are absent (verified in cobalt-crush-dev.md and all divisor-* agent files via Dewey/Grep fallbacks); (2) install-time granularity — uf init offers only --divisor as a persona-level carve-out, leaving Cobalt-Crush, Mx-F-coach, and constitution-check bundle-only (verified in cmd/unbound-force/main.go flag registration and specs/006-cobalt-crush-architecture plan: "no separate install step"); (3) shared-contract coupling — convention packs (severity, default, ci, content, review-voice, meta-prompt) are dual-consumed vocabularies co-deployed unconditionally via alwaysDeployedPacks (internal/scaffold/scaffold.go). When triaging governance-claim issues, verify by symbol content (grep for flag registration, alwaysDeployedPacks) rather than cited line numbers — line citations drift as the codebase evolves (issue #559's scaffold.go refs drifted ~300-460 lines and its three-flag count grew to five after --stealth/--check landed) without undermining the substantive claim.
