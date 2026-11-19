---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-15T10:18:59Z
identity: determinism-20261115T101859-jay-flowers
tier: draft
as_of: "unbound-force@8145ac2d9e933e918c7e47434e8955e2cb13cfa6"
---

DE-007 seam evidence reports `round-trip` as missing for dispatch_review_runs on every review that modifies the tool, but the layer is covered in substance: .opencode/test/dispatch-review-runs.test.ts's RD-CF-001 block ("cross-session consolidation by forwarded correlation_id", ~line 1370) drives the real dispatch_review_runs and reads the persisted session back through consolidate_dispatch from a different tool session. tool-inventory's classifyTest (.opencode/lib/tool-inventory.ts:463-466) recognizes the round-trip layer only from a file name matching /round-?trip/ or a test title pairing the tool name with the literal "round-trip" token, so cross-tool execute() chains in an "other"-classified file are invisible to it. Adjudication pattern: before reporting a CRITICAL DE-007 on a missing round-trip layer, read the referencing "other"-layer test files for write-then-read execute() chains (dispatch tool → consolidate/status tool by correlation_id); report the violation only when no substantive coverage exists, and recommend titling one covering test with the "round-trip" token so the evidence and the substance agree.
