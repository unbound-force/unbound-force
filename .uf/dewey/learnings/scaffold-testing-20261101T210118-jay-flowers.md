---
tag: scaffold-testing
author: jay-flowers
category: pattern
created_at: 2026-11-01T21:01:18Z
identity: scaffold-testing-20261101T210118-jay-flowers
tier: draft
---

Guard-test pattern for embed-FS template references: when a scaffolded template references other embedded assets by name (e.g., a model-matrix template whose `packs:` keys must map to pack files), enforce the reference closure with a dedicated test that parses the embedded template, iterates the referenced names, and fails naming each missing embedded asset path. Two design points make the guard trustworthy: fail loudly when the reference map is empty or absent, so a restructured template cannot pass vacuously; and sort the iterated keys, so failure output is deterministic across map iteration orders. Precedent: TestSCFR009_DeploymentCompletenessGuard (model-matrix-scaffold-deployment, issue #889), which turned the implicit slice-1 assumption that every `packs:` key resolves to an embedded pack into a checked invariant — and would have caught the shipped gap where the starter matrix mapped `meta-prompt` with no embedded pack file, leaving hero-repo matrices unloadable (MATRIX_INVALID). (Lesson proposal from divisor-architect, #889 review-council dispatch.)
