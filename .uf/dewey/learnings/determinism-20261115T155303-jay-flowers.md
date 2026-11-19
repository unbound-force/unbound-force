---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-15T15:53:03Z
identity: determinism-20261115T155303-jay-flowers
tier: draft
---

tool-inventory's per-tool `missing: ["round-trip"]` seam flag can under-report coverage when write→read tests live in regression suites the layer heuristic classifies as "other". In harden-spec-review-gate (PR #1094), dispatch_review_runs, relay_spec_review_gate, and uf_unleash_detect_resume all reported round-trip missing, yet spec-review-gate.test.ts carried an explicitly named "round-trips the consolidated record through get_progress, checkpoint, and the relay (DE-007 layer 3)" test, a real-tool dispatch-guard refusal test, and unleash-resume.test.ts wrote gate state then invoked the real detect_resume tool. Before reporting a DE-007 violation, verify substance: look for a test that persists state through the recording primitive or writer tool and then invokes the real reader tool surface — the heuristic's layer labels are evidence, not the verdict.
