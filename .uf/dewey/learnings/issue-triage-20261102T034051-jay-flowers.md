---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:40:51Z
identity: issue-triage-20261102T034051-jay-flowers
tier: draft
---

When triaging instruction-consolidation issues that deduplicate protocol prose from command files into a shared skill, grep the protocol's tool names across both the command files and the skill — divergence between the copies is stronger triage evidence than duplication alone, because it shows the repeated "documented once" contract has already drifted into contradiction. Verified on unbound-force issue #663 (expand dispatch-advisor skill to full dispatch lifecycle): uf.address-feedback.md line 1241 still mandated execution through dispatch_agent_run while the dispatch-advisor skill (SKILL.md lines 96-97) mandated dispatch_review_runs and explicitly forbade individual dispatch_agent_run calls — the triplicated fail-closed protocol had silently split into conflicting versions. Two follow-ups for the implementing PR surface from the same grep: the skill has a scaffold mirror (internal/scaffold/assets/opencode/skills/dispatch-advisor/SKILL.md) whose byte parity drift tests make mirror sync mandatory, and the issue's step-4 wording (per-run dispatch_agent_run, PR #686) predates the dispatch_review_runs aggregation, so the expanded lifecycle doc must describe the current aggregated invocation.
