---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:35:56Z
identity: issue-triage-20261101T233556-jay-flowers
tier: draft
---

A documentation-gap issue claiming "constraint X is documented here but missing there" can be verified objectively with one repo-wide grep of the constraint's distinctive term: matches in implementation surfaces (skill/command files, plugin code, tests, spec deltas) establish that the constraint exists and is mechanically enforced, while zero matches in the named presentation surfaces confirm the gap. On issue #826 ("risk floors missing from docs/usage.md workflow table and AGENTS.md Quick row"), grep "risk floor" matched tier-routing SKILL.md, uf.unleash.md, uf-workflow/index.ts (ROUTE_RISK_FLOOR enforcement), unleash-route-decision.test.ts, and the unleash-tier-routing spec delta, with zero matches in docs/usage.md and AGENTS.md — verifying both halves of the claim in a single pass, corroborated by the validated documentation-pipeline compiled article's "Still-open gap" entry. Complement the grep with a Dewey semantic search for the constraint term: validated compiled articles and prior learnings frequently record the originating PR and the exact surfaces that were missed.
