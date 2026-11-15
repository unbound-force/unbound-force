---
tag: provenance-drift
author: jay-flowers
category: pattern
created_at: 2026-11-10T19:08:09Z
identity: provenance-drift-20261110T190809-jay-flowers
tier: draft
---

Parallel literal pins in separate tests only catch uncoordinated drift between a provenance record and the artifact it documents. When a provenance record claims to mirror a shipped artifact (e.g., internal/scaffold/testdata/plugin-package-provenance.json vs the package.json overrides block), the verifying test should cross-check record entries directly against the artifact's fields — as the packages loop in TestPluginPackageInputs_ApprovedMaintenanceProvenance does against manifest pins — not only against test-local constants. Without the direct linkage, a partial future bump that keeps each test internally consistent (manifest + one test's constants updated, record + the other test's constants stale) leaves record and artifact silently diverged while the suite stays green.
