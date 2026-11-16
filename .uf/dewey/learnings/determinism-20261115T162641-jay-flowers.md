---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-15T16:26:41Z
identity: determinism-20261115T162641-jay-flowers
tier: draft
as_of: "unbound-force@f1393b41a3438503096e2003e0f6801a157d2be7"
---

DE-007 spec-review method for review-dispatch plugin changes: before judging a tasks.md's seam coverage, map the three seam layers to the tests that actually pin them. In this repo: registration-surface.test.ts is generic (dynamic tool enumeration; per-parameter locks exist only for specific contracts such as consolidate_dispatch correlation_id), scratch-smoke.test.ts pins registered tool-name lists, summary-retrieval.test.ts pins the compact-result key order and the SUMMARY_LIMITS literals (including errorChars), and golden-output-shape.test.ts pins shapes for the documented call sites. A change that extends a tool input enum or adds an output key needs tasked updates only for the layers that pin the modified contract — the generic registration and tool-name layers stay green untouched. Verified on the review-dispatch-errors-cap-loses-unrecov spec review (2026-11-15): tasks 4.1-4.5 covered every pinning layer for the failure_record key, the errors whole-or-absent projection, and the read_dispatch_artifact errors section; no registration-surface update was needed.
