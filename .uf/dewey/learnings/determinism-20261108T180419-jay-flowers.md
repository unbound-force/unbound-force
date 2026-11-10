---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-08T18:04:19Z
identity: determinism-20261108T180419-jay-flowers
tier: draft
---

DE-001 admits no deferral by design note. Recording a prose→tool extraction as a "natural follow-up" in a change's design.md Risks section does not exempt mechanical instruction added by that same change: the determinism pack's opportunistic-migration carve-out covers pre-existing prose only, so a fetch + condition-gate + deterministic-selection + prescribed-output step shipped as new prose is a CRITICAL DE-001/DE-003/DE-005 violation even when the risk note acknowledges it. Either ship the extraction with the change (plus DE-007 seam tests: registration surface, documented output shape, round-trips) or bind the follow-up with a tracked issue and a tasks.md entry naming the extraction and its seam-verification layers — an unbound "natural follow-up" note enforces nothing and leaves reviewers no mechanism to verify the follow-up ever lands.
