---
tag: regression-protection
author: jay-flowers
category: pattern
created_at: 2026-11-07T16:56:26Z
identity: regression-protection-20261107T165626-jay-flowers
tier: draft
---

Spec-mandated content-preservation behaviors need composed Run()-level regression tests, not only unit tests of the merge helper. Precedent: PR 957 (stop-shipping-plugin-lock) — SC-FR-003's devDeps-preservation wiring in the scaffold deploy loop had only direct unit tests of preservePluginDevDependencies, while the adjacent foreign-lock scenario got a full Run()-level test (TestSCFR003_ForeignLockSurvivesRun). Unit tests of the helper cannot detect wiring regressions: a dropped call, a wrong relPath guard, or a reorder past the bytes.Equal skip comparison leaves the suite green while user-owned content is silently stripped. When composing the test, seed a target that differs from the merged result (e.g., an older pin plus devDependencies) so the write path actually runs and the file reports Updated — a byte-identical seed only exercises the skip path and misses reorder regressions.
