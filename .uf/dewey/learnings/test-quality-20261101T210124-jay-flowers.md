---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-01T21:01:24Z
identity: test-quality-20261101T210124-jay-flowers
tier: draft
---

In scenario-driven test design (OpenSpec/speckit), requirement bullets that state negative invariants ("X is never scaffolded", "Y is preserved") without a dedicated Given/When/Then scenario get skipped by tests built strictly from the scenario list. Indirect guards — exact-set asset-manifest tests and exact created-count assertions — do catch unexpected additions, but the lockstep-update workflow (adding the new path to the expected-asset manifest to make the suite green) silently converts the violated invariant into sanctioned behavior, with no test ever citing the spec clause that forbade it. When a spec requirement bullet carries a MUST that no scenario covers, add an explicit preservation/negative assertion to the behavior test: seed the file, then assert content is byte-unchanged and the path appears in no disposition list after every subsequent run. Observed in the model-matrix-scaffold-deployment change (SC-FR-010's .uf/model-matrix.override.yaml clause had no direct assertion in TestRun_ModelMatrixPreservedOnRerun while every scenario-backed clause did). (Lesson proposal from divisor-testing, #889 review-council dispatch.)
