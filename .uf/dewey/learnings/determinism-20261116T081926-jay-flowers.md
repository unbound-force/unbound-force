---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-16T08:19:26Z
identity: determinism-20261116T081926-jay-flowers
tier: draft
---

Divisor determinism reviews: tool-inventory.ts seam evidence classifies tests file-name-first (LAYER_FILE patterns return before LAYER_TITLE test titles are consulted), so a round-trip test living in a file named golden-output-shape.test.ts is classified only as output-shape and the tool reports missing:["round-trip"] even when genuine round-trip coverage exists (e.g. golden-output-shape.test.ts:2147 "round-trips a record_human_decision write through checkpoint ref validation and get_progress" and :2223 for the orchestrator-relay envelope). Before reporting a DE-007 missing round-trip layer, grep the output-shape-named test files for round-trip titles — the missing entry can be a classification artifact rather than a coverage gap. Also check whether the tool writes state at all: a read-only tool (relay_spec_review_gate) needs no round-trip layer.
