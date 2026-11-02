---
tag: documentation-drift
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:14:05Z
identity: documentation-drift-20261102T021405-jay-flowers
tier: draft
---

Per-change documentation issues can go stale on arrival when the documented artifact grows incrementally after the founding change: unbound-force issue #720 (docs coverage for the determinism-codification change) describes .opencode/uf/packs/determinism.md as carrying "rules DE-001..DE-004" — accurate for the founding change — but the pack now carries DE-001..DE-008 after three later governance changes (router-pattern DE-005/DE-006, seam-verification DE-007, delegation-enforcement DE-008), each with its own docs issue (#762, #781, #791). Docs implemented from a founding-change issue must re-verify rule and artifact enumerations against the artifact's current state at implementation time, or scope the enumeration explicitly; acceptance criteria phrased as "mention X or explicitly scope the list" tolerate this drift, fixed enumerations in issue bodies do not.
