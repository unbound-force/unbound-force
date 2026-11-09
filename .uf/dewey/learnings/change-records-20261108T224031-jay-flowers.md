---
tag: change-records
author: jay-flowers
category: gotcha
created_at: 2026-11-08T22:40:31Z
identity: change-records-20261108T224031-jay-flowers
tier: draft
---

DE-004 change-record deltas should name their measurement source so they reconcile with the evidence tool. The divisor-determinism-persona proposal records "divisor-determinism is new, 246 lines; divisor-linguist shrinks 236 -> 203 lines (load set -3880 tokens)", but the raw diff counts are +263 lines for the new file and 262 -> 230 for divisor-linguist (+26/−58). If the recorded numbers come from a different measurement (e.g., linter-counted lines excluding blanks), the record should say which tool produced them; otherwise the Determinist's `instruction-lint --base` delta.totals check flags an apparent mismatch between the recorded deltas and the tool-computed ones.
