---
tag: seam-verification
author: jay-flowers
category: gotcha
created_at: 2026-11-08T19:24:46Z
identity: seam-verification-20261108T192446-jay-flowers
tier: draft
---

A learning's "verified instance" must come from an executed observation, not from reading code shapes. On unbound-force PR #959, reading tool-inventory.test.ts's tuple-shaped PLUGINS list led a shipped learning to claim seamReport()'s genericPlugins resolves to an empty set for the real repository, while the real pre-existing registration-surface.test.ts declares its plugins as `{ id: "..." }` object entries that the extraction regex matches — all four plugins are credited at head. The repo-level discoverability test (run the evidence tool against the real repository and assert the coverage it should credit is credited) is the artifact that converts such claims from archaeology to falsifiable evidence; it would have caught the learning's error at authoring time, and it pins the convention against the rename/reshape drift that made the claim plausible.
