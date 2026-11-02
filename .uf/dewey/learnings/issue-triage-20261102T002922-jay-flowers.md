---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:29:22Z
identity: issue-triage-20261102T002922-jay-flowers
tier: draft
---

Umbrella-issue decomposition in unbound-force triage: migration umbrellas (e.g., #797 OpenCode v2 migration) decompose into phase-tracked child issues that open with "Umbrella: #N" and cite umbrella decision IDs (e.g., D1/D3). When duplicate detection surfaces the umbrella as a title-keyword candidate for one of its phase children, the declared parent-child link is intentional decomposition, not duplication — triage the child on its own phase scope and exit criteria instead of closing it as a duplicate. Phase children feed the umbrella's next gate (e.g., Phase 4 dogfood sign-off feeding the Phase 5 cutover gate), so closing a phase child as a duplicate of its umbrella would sever the gate chain.
