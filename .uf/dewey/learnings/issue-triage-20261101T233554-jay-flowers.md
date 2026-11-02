---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:35:54Z
identity: issue-triage-20261101T233554-jay-flowers
tier: draft
---

Documentation-gap issues claiming a constraint is "missing from" summary docs can be verified with a single repo-wide grep of the constraint term: the match set establishes the authoritative sources (and surfaces any locations the issue under-cites), while non-matches confirm the claimed gaps. Verified on unbound-force issue #826 (tier-routing risk floors missing from docs/usage.md workflow table and AGENTS.md Quick row): grep for the risk-floor term confirmed documentation in .opencode/skills/tier-routing/SKILL.md (Phase 4, four escalate-only signals), .opencode/commands/uf.unleash.md, and mechanical enforcement in .opencode/plugins/uf-workflow/index.ts (ROUTE_RISK_FLOOR rejection with table-driven tests), while docs/usage.md and AGENTS.md returned zero matches — and the validated documentation-pipeline compiled article plus learning documentation-pipeline-20261020T195900 independently recorded the same gap from the PR #784 audit. The grep also cheaply bounds scope: website docs already carried the floors, so the two cited locations were the complete gap.
