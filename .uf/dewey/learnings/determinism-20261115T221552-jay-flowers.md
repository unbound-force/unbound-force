---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-15T22:15:52Z
identity: determinism-20261115T221552-jay-flowers
tier: draft
as_of: "unbound-force@6a553043168ccaffc4677672b76b30f5453aa1b5"
evidence:
  - ".opencode/lib/tool-inventory.ts"
  - ".opencode/test/read-dispatch-artifact.test.ts"
  - ".opencode/test/consolidate-dispatch.test.ts"
  - ".uf/artifacts/dispatch/444d957c-7bee-4d37-86ce-e126eb844edf.json"
---

Code-review DE-007 adjudication for review-dispatch plugin tools: tool-inventory's seam layer classifier is path-regex based (only files matching /golden|output-shape/ credit the output-shape layer; round-trip needs a matching path or a test title naming the tool), so a tool whose documented output shape is pinned exhaustively in its own dedicated test file reports `missing: ["output-shape"]` as a false positive. Verified on the review-dispatch-errors-cap-loses-unrecov code review (2026-11-15): read_dispatch_artifact showed output-shape missing while read-dispatch-artifact.test.ts pinned the new errors-section return shape with toMatchObject over every documented field plus an Object.keys exact-set assertion (the Layer-2 golden discipline), a description/args parity test, and consolidate-dispatch.test.ts carried the production writer→reader round-trip (explicitly citing the DE-007 writer/reader seam). Before flagging CRITICAL DE-007 on a `missing` layer, check the layer's substance in the tool's dedicated test file — the name heuristic cannot see it. The need trigger also matters: output-shape is needed only when an instruction file documents the tool's output fields, and read_dispatch_artifact's paging fields (next_offset, status) documented in uf.unleash.md and peers are unchanged by the diff and already pinned in that same dedicated file since PR #949's feedback.
