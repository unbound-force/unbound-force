---
tag: dependency-management
author: jay-flowers
category: pattern
created_at: 2026-11-05T14:58:07Z
identity: dependency-management-20261105T145807-jay-flowers
tier: draft
---

When removing a verification artifact that carries multiple contract dimensions (e.g., a package-lock.json carrying version, integrity, AND license evidence), triage each dimension separately before deleting its assertions: map each dimension to a surviving verification path (installed node_modules package.json, manifest pins, doctor checks) or record its intentional drop in the spec delta's Previously paragraph and the design decisions. Observed in PR 957 (stop-shipping-plugin-lock): deleting assertLockedDirectPackage/assertLockedSecurityOverride to remove the lock silently dropped the license-verification contract ("all four dependencies MUST retain their MIT license") that lived only inside those assertions, while the version dimension carried forward as manifest pins — the spec rewrite and design.md recorded the lock/integrity/reproducibility trade-offs but not the license gate. Key distinction: a static testdata provenance record is not a live gate — it asserts a hand-maintained record against hardcoded expectations and cannot detect a license change in the actually-resolved dependency tree. Recipe: before deleting assertions tied to a removed artifact, enumerate every contract dimension those assertions enforce and give each one an explicit destination (carried, replaced, or documented-drop).
