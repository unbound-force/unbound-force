---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-10-31T23:10:21Z
identity: test-quality-20261031T231021-jay-flowers
tier: draft
---

Schema-registry sample fixtures classified 'invalid' must be invalid at the layer the validation harness asserts. Loader-enforced rules — filesystem pack-existence checks, role-aware credential-population guards — are invisible to JSON Schema validation, so fixtures exercising them pass schema validation and silently satisfy an agreement-only harness. Cross-validation harnesses must assert per-fixture expected verdicts (filename or metadata convention), not merely identical verdicts between two validators; otherwise shared drift to accept an invalid shape goes undetected. On the model-matrix-delegation review (2026-10-31), task 2.5's harness asserted only v1.schema.json/Zod agreement while two of six task-4.2 invalid fixtures were loader-level invalid only, leaving MM-FR-008's 'invalid samples fail' scenario with no implementable test.
