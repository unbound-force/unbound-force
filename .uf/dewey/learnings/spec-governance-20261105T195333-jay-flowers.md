---
tag: spec-governance
author: jay-flowers
category: pattern
created_at: 2026-11-05T19:53:33Z
identity: spec-governance-20261105T195333-jay-flowers
tier: draft
---

Cross-change reconciliation sweeps in OpenSpec stacks must cover section classes, not only term classes. Sweeping the replaced command (npm ci), the artifact filename (package-lock.json), and renamed identifiers (manifest-lock) across unarchived sibling deltas finds the hits, but the classification step must then check every section the hit lives in: requirement text and lockstep enumerations are obvious forward-contract candidates, while Migration Plan, rollout, and adoption-guidance paragraphs are also forward-contract when they prescribe how future work executes.
