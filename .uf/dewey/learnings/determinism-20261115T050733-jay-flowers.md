---
tag: determinism
author: jay-flowers
category: reference
created_at: 2026-11-15T05:07:33Z
identity: determinism-20261115T050733-jay-flowers
tier: draft
---

instruction-lint --json --base <merge-base> --changed emits a delta block with per-file line/token deltas and totals that is paste-ready for change records. DE-004 asks change records (proposal, tasks, PR description, or CHANGELOG) to state those numbers whenever instruction text changes; running the linter delta at edit time and pasting the totals closes the gap without hand-counting. In PR #1087 the uf.review-council.md edit (+6 lines, +96 tokens) appeared in no change record, persisting a prior-round advisory.
