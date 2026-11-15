---
tag: seam-verification
author: jay-flowers
category: gotcha
created_at: 2026-11-15T17:00:33Z
identity: seam-verification-20261115T170033-jay-flowers
tier: draft
---

tool-inventory's DE-007 seam classifier is title-driven: a test file counts as round-trip coverage for a tool only when a single describe/it/test title contains both the literal tool name (e.g. relay_spec_review_gate) and the keyword "round-trip" (classifyTest, .opencode/lib/tool-inventory.ts:463-466). Tests titled with descriptive tool references ("the relay") or layer numbers only ("DE-007 layer 3") leave `missing: ["round-trip"]` flags in --seams output even when the write->read coverage exists, and cross-tool round-trips get attributed only to the tool named in the title (the reading tool, not the writing tool). When authoring seam tests, put the tool name and "round-trip" in the test title; when reviewing, treat a missing round-trip flag as a prompt to read the tests before filing DE-007.
