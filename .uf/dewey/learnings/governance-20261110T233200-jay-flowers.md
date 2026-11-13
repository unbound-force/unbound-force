---
tag: governance
author: jay-flowers
category: decision
created_at: 2026-11-10T23:32:00Z
identity: governance-20261110T233200-jay-flowers
tier: draft
---

Governance floor for deferred determinism extractions (PR 1024 feedback triage): when a prose→tool extraction cannot land in the change that introduced the choreography, the deferral satisfies governance only when three elements are bound: (1) a tracked GitHub issue naming the extraction and its DE-007 seam layers (registration surface, documented output shape, write→read round-trips), (2) a documented principle tradeoff per the constitution's conflict-resolution rule when Principle V (ship the security fix now) is sequenced ahead of Principle VI (no prose choreography), and (3) the change record (CHANGELOG) pointing at the tracker instead of presenting the interim prose form as terminal. A narrated follow-up in a review thread or changelog maps to no actionable work — zero-waste violation and untracked gate debt. Also weigh scaffold amplification: command-template changes mirror byte-identically into internal/scaffold/assets/, so unextracted choreography ships to every hero repo via uf init and the deferral cost compounds per deployment.
