---
tag: test-fixtures
author: jay-flowers
category: gotcha
created_at: 2026-10-31T19:52:45Z
identity: test-fixtures-20261031T195245-jay-flowers
tier: draft
---

Seed-then-modify-then-rerun disposition regression tests (TestRun_ConstitutionPreservedOnRerun, TestRun_ReviewerCapabilitiesPreservedOnRerun in unbound-force internal/scaffold/scaffold_test.go) discard the seeding run's *Result, which gaze reports as a P0 ReturnValue gap for the test. When the pinned spec scenario includes first-deploy semantics ("reported in the Created disposition list"), assert the seeded file's membership in the first run's Created list inside the same test — a single membership loop makes the test self-contained against the scenario and closes the gap. Discarding is defensible only when first-deploy membership is already pinned per-file elsewhere (TestRun_CreatesFiles iterates expectedInitiallyDeployedAssetPaths and asserts each deployed asset in result.Created); environment-tolerant count-range checks are weaker than per-file membership and should not be the coverage rationale.
