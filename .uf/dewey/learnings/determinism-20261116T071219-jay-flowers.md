---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-16T07:12:19Z
identity: determinism-20261116T071219-jay-flowers
tier: draft
---

tool-inventory --seams classifies a test file's layer by its filename role (registration-surface.test.ts → registration, golden-output-shape.test.ts → output-shape), so a genuine write→read round-trip test living inside golden-output-shape.test.ts still surfaces as missing: ["round-trip"] for the tool it exercises. In the synthetic-compaction-continue-prompt-rec review (2026-11-16), record_human_decision and relay_spec_review_gate both carried missing round-trip flags while golden-output-shape.test.ts:2147 ("round-trips a record_human_decision write through checkpoint ref validation and get_progress") executed exactly that layer through the registry dispatch path. A missing-layer flag is filename-derived evidence, not a verdict: dismiss it only after reading the referencing tests, and expect round-trip coverage to live inside golden-output-shape suites whenever the change follows the DE-007 task-7.1 layout.
