---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-15T06:08:53Z
identity: determinism-20261115T060853-jay-flowers
tier: draft
as_of: "unbound-force@e5e077b35f0c7984517024431e696737e330b080"
---

tool-inventory seam evidence reports missing: ["round-trip"] for read-side tool changes (guards, relays, resume probes) even when the write→read path is fully covered elsewhere: the layer classifier attributes round-trip to test files that reference the tool, and the writer's suite (or a fixture that drives the real tool against recorded state) carries the coverage. DE-007 adjudication must first ask whether the changed behavior writes state at all — relay_spec_review_gate and uf_unleash_detect_resume write none, and dispatch_review_runs' new specReviewDispatchGuard only reads — then locate the write→read coverage across suites (spec-review-gate.test.ts drives the actual dispatch tool against state written by appendSpecReviewIterationToState and asserts the iteration_already_recorded refusal before correlation minting). Verified on harden-spec-review-gate (opsx/harden-spec-review-gate, 2026-11-15): all three missing-round-trip entries were read-side changes with covered read paths, not seam gaps.
