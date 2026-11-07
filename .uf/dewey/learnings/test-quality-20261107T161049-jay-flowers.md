---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-07T16:10:49Z
identity: test-quality-20261107T161049-jay-flowers
tier: draft
---

When a spec mandates tool-level behavior (e.g., "uf init MUST keep a target's existing devDependencies on refresh"), a direct unit test of the extracted helper function does not verify the mandate — only the helper. The call-site wiring (merge applied to the right file, in the right order relative to the idempotency comparison, feeding the correct write/skip reporting) can regress while every unit test stays green. Ship a composed test through the entry point (Run()) alongside the unit tests: pre-stage the target state, run the real entry point, assert the deployed artifact content and the reported action. Precedent pair from PR #957: TestSCFR003_ForeignLockSurvivesRun (composed, Run()-level, catches wiring regressions) versus TestPreservePluginDevDependencies_Cases (unit-level only, would not catch a dropped call at the deployment loop).
