---
tag: documentation-maintenance
author: jay-flowers
category: pattern
created_at: 2026-11-06T20:17:01Z
identity: documentation-maintenance-20261106T201701-jay-flowers
tier: draft
---

Persona-count literals in documentation drift on every agent addition. unbound-force enumerates its Divisor council roster in at least four places (docs/heroes.md "9-persona review council", docs/usage.md "divisor-* (9 agents)" and "9-persona review council", docs/architecture.md) alongside divergent repo-wide counts (AGENTS.md "17 active", README.md "18 agents") and a separate scaffold-scoped count (QUICKSTART.md "11 agents ... reviewer council (9 personas)", which describes what uf init deploys and is unaffected by repo-local personas). Prefer count-free phrasing ("personas are discovered from .opencode/agents/divisor-*.md at review time") or a single canonical roster table, and distinguish repo-local personas from the scaffolded set; when a change adds a persona, grep docs/ for the old count in the same change.
