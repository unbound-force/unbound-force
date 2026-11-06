---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:17:45Z
identity: test-quality-20261105T161745-jay-flowers
tier: draft
---

A payload array cardinality cap declared on two parallel validation surfaces (runtime Zod schema + registry JSON Schema) needs a direct bound+1 rejection test on each surface. The producing path's early truncation (slice at cap + overflow warning) makes the schema rejection unreachable from the happy path, so per-entry shape tests and truncation tests leave the declared bound unverified — and nothing pins both surfaces to the same number. Pattern from PR #949 (RD-FR-054, proposals max 512): submit a 513-element array of individually-valid entries directly to each surface and assert rejection — .opencode/test/finalize-error-boundaries.test.ts errorsFor() for the Zod surface, internal/schemas/review_dispatch_test.go schemas.ValidateBytes for the JSON Schema surface. The same gap pattern applies to every dual-surface array bound in the review-dispatch contract (findings 512, plan 512, advisories 512, runs 64).
