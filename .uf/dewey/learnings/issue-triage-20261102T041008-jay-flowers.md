---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:10:08Z
identity: issue-triage-20261102T041008-jay-flowers
tier: draft
---

Convention-addition triage gap-check (unbound-force/unbound-force): verify a claimed guidance gap against both AGENTS.md and the applicable convention packs before accepting it. The AGENTS.md Testing Conventions section covers framework, assertions, naming, isolation, and drift detection but says nothing about in-file pattern reuse, and go.md TC-006 encourages table-driven tests without instructing agents to extend existing tables rather than reimplement equivalent logic in standalone test functions — the exact gap that produced a medium DRY/code-organization finding and a fix cycle on PR #594. Project-specific Go rules also have a second designated home beyond AGENTS.md: the empty .opencode/uf/packs/go-custom.md Custom Rules section (CR-NNN prefix) loads during Go-pack implementation and review, while AGENTS.md stays in every agent's base context — placement between the two is an implementation decision to surface, not a validity blocker.
