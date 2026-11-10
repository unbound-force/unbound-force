---
tag: openspec-governance
author: jay-flowers
category: gotcha
created_at: 2026-11-07T20:25:30Z
identity: openspec-governance-20261107T202530-jay-flowers
tier: draft
---

Retroactive supersession records that claim exhaustive coverage (e.g. 'the inline annotations mark every contradicted span') must derive their census mechanically — grep the retired token across the affected tree and account for every hit — not compile it per-file by recall. In PR 948's retroactive-push-skill-contracts record, the manual census annotated the latest-generation twin of each contradicted span (spawn-subtask-classification-params determinism-policy L188-196; model-matrix-delegation unleash collision L189-196) but missed the identical older spans in delegation-skill-enforcement (determinism-policy L121-129; unleash collision L115-122), leaving task 2.2's enumeration at three of four files carrying the legacy root and design.md's 'every contradicted span' mitigation claim factually false. The trap: contradicted text often exists in BOTH the older and newer deltas of a capability stack (ADDED then MODIFIED), and per-file recall finds only one copy. Per-occurrence grep enumeration — a DE-001-class mechanical procedure — closes the class; fixing only the instance a reviewer names leaves the claim still false.
