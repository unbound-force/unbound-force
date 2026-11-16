---
tag: seam-verification
author: jay-flowers
category: gotcha
created_at: 2026-11-16T07:50:17Z
identity: seam-verification-20261116T075017-jay-flowers
tier: draft
---

tool-inventory --seams classifies a test file's seam layer by filename (registration-surface.test.ts -> registration, golden-output-shape.test.ts -> output-shape), so a Layer-3 round-trip test that lives inside golden-output-shape.test.ts is reported under the output-shape layer and the round-trip layer shows up as `missing` for the tool it covers. In the synthetic-compaction-continue-prompt-rec review (2026-11-16), both record_human_decision and relay_spec_review_gate reported `missing: ["round-trip"]` while the round-trips were actually covered — golden-output-shape.test.ts's "round-trips a record_human_decision write through checkpoint ref validation and get_progress" executes the full registry-path write->read pair. The `missing` array is filename-heuristic evidence, not a verdict: a determinism reviewer must read the flagged tests (or grep the test names for "round-trip") before reporting a DE-007 finding, and should treat a missing-layer flag on a tool whose tests cluster in a differently-named file as a classification artifact first.
