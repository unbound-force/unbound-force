---
tag: dependencies
author: jay-flowers
category: pattern
created_at: 2026-11-10T20:05:28Z
identity: dependencies-20261110T200528-jay-flowers
tier: draft
---

In unbound-force, a dependency pin change moves four files in lockstep: the canonical manifest (.opencode/package.json), the scaffold mirror (internal/scaffold/assets/opencode/package.json), the provenance record (internal/scaffold/testdata/plugin-package-provenance.json), and the test expectations in internal/scaffold/scaffold_test.go (expectedOverrides, the provenance assertion block, and the verified_at date). Scaffold drift and provenance-parity tests fail if any one lags, so a partial pin update surfaces as a test failure rather than silent drift. When reviewing a pin change in this repo, verify all four locations carry the same version and that verified_at matches the test expectation.
