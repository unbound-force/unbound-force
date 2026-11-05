---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-05T20:31:23Z
identity: determinism-20261105T203123-jay-flowers
tier: draft
---

New instruction is new instruction: when a fix adds git-choreography prose to a command file, AGENTS.md's Determinism gate and determinism.md DE-001/DE-005 apply at authoring time — the 'existing prose migrates opportunistically' carve-out covers only pre-existing text, not sections added by the current change. Incident-style or bugfix-shaped changes feel exempt but are not. Two corollaries from PR #955 feedback item 7: (1) review-time severity for a prose-form violation follows the packs (DE-001/DE-005 = CRITICAL) even when the reviewer files the item [low] and scopes remediation 'longer term' — severity classifies the violation class; the thread label classifies fix timing. Keeping these separate preserves the governance signal without overriding the maintainer's triage. (2) When the fix's scope cannot absorb the tool extraction, the increment must still record a tracked, actionable follow-up (issue or change record — the item-3-dispatch pattern), because an untracked deferral leaves a standing MUST-rule violation as the instruction's permanent form and doubles every future edge-case fix across the canonical command and its drift-enforced scaffold mirror.
