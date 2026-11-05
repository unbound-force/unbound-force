---
tag: supply-chain
author: jay-flowers
category: gotcha
created_at: 2026-11-04T20:08:34Z
identity: supply-chain-20261104T200834-jay-flowers
tier: draft
---

When a change removes a lockfile from an install pipeline, audit every install surface for the flags that made the locked regime safe, not just the primary one. In stop-shipping-plugin-lock (PR #957), the change kept --ignore-scripts on the uf init activation path and design.md listed it as the transitive-drift mitigation, but the `make plugin-test` path — the surface with the widest transitive tree (the vitest dev graph) and the one CI runs on every PR — switched to lockless `npm install` without --ignore-
