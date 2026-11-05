---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-05T17:48:48Z
identity: test-quality-20261105T174848-jay-flowers
tier: draft
---

When a spec MUST requirement's implementation surface is command or skill prose, tool-side tests alone leave it regression-unprotected: pin the prose with a content gate. In unbound-force, command prose implementing RD-FR-056 is gated by TestCommandContracts_ArtifactSourcedLessons (clause strings plus stale-string removal checks), and skill prose is gateable per the TestDeterminismFirstSkill_SpawnToolParameters precedent (scaffold_test.go reads .opencode/skills/determinism-first/SKILL.md, scopes to a named section, and asserts the load-bearing strings). The preflight-declared-output change tested only the uf_preflight_declared_checks tool and left SKILL.md's Phase 5 "Project-declared check output" insertion rule ungated, so PF-FR-024's report-assembly scenarios had no automated verification. Rule of thumb: every spec scenario maps to either an executable test or a content gate; a scenario mapping to neither is a coverage gap, and scaffold mirror-parity tests only prove canonical==mirror, never canonical==spec.
