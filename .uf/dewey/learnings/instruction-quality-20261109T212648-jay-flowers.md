---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-09T21:26:48Z
identity: instruction-quality-20261109T212648-jay-flowers
tier: draft
---

Fixing an instruction-lint dead-reference finding (IL-010) requires two independent checks: liveness (the replacement command/skill exists — the linter verifies this mechanically) and semantic fit with the state the surrounding handler remediates. In unbound-force PR #1001, blocked-state handlers in opsx-apply.md and openspec-apply-change/SKILL.md pointed at nonexistent continue commands; the correct replacement was the artifact-generating propose command (/opsx-propose, openspec-propose), because state "blocked (missing artifacts)" is remediated by generating artifacts — and opsx-propose's own guardrails (skip branch creation when already on opsx/<name>; ask continue-or-new when a change with that name exists) make the route safe for an existing blocked change. A reference that is merely live but semantically wrong (e.g., an archive or explore command) would have passed the linter while mis-routing the user.
