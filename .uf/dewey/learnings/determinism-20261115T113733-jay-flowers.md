---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-15T11:37:33Z
identity: determinism-20261115T113733-jay-flowers
tier: draft
---

DE-004 delta-evidence gotcha: when a review-fix iteration edits an instruction file after the change record was authored, the record must be updated with the instruction-lint delta rows. A proposal's Principle VI claim of "measurable deltas" citing behavior knob values (thresholds, margins) does not satisfy DE-004 — that rule wants instruction line/token counts. The evidence source is `.uf/scripts/instruction-lint --json --base <base> --changed`, whose `delta` block carries per-file lines/tokens/loadTokens and totals; paste those rows into the CHANGELOG entry or the proposal's Impact section whenever `delta.files` is non-empty.
