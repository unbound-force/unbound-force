---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T00:13:21Z
identity: issue-triage-20261102T001321-jay-flowers
tier: draft
---

Documentation-companion issues for a new workflow mechanism should scope a terminology-collision check against the user-facing docs tree, not just coverage. Triage of unbound-force/unbound-force#819 (/uf.unleash build/test fix loop, PR #815 companion) found that user-facing docs describe /uf.unleash in docs/usage.md, docs/architecture.md, docs/configuration.md, and docs/heroes.md, yet none cover the new automatic build/test fix loop — while the phrase "fix loop" is already canonicalized in the same docs for a different mechanism: /uf.fix-ci's human-gated "Gated fix loop" (docs/usage.md:101, docs/architecture.md:550). The docs work therefore needs both new coverage and a distinct canonical term for the automatic loop, or readers will conflate automatic causality-classified delegation with human-gated CI diagnosis (content pack VB-003, terminology consistency). Recipe: when triaging or executing a docs-companion issue, grep the docs tree for the mechanism's short name; a hit belonging to an unrelated feature means the issue should mandate a disambiguating canonical term rather than reusing the colliding phrase.
