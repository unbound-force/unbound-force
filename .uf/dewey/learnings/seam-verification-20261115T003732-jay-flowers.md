---
tag: seam-verification
author: jay-flowers
category: gotcha
created_at: 2026-11-15T00:37:32Z
identity: seam-verification-20261115T003732-jay-flowers
tier: draft
---

tool-inventory --seams classifies test layers by FILE NAME (LAYER_FILE patterns: /registration-surface/, /golden|output-shape/, /round-?trip/; round-trip additionally via test titles) — a dedicated per-tool test file can never register as the output-shape layer. A new plugin tool verified by a full-envelope toEqual plus schema-rejection tests in its own test file (the resolve_review_mode / select_prior_review pattern) therefore reports missing: ["output-shape"] even though the documented output shape is verified in substance through the same plugin dispatch path. Adjudicate DE-007 on whether the needed verification exists (generic registration credit, documented-field assertions, round-trip applicability for stateless tools), not on the missing list alone; note that the golden-output-shape.test.ts site registry predates the newer review-dispatch tools and is not extended by them.
