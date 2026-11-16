---
tag: determinism-review
author: jay-flowers
category: gotcha
created_at: 2026-11-16T02:25:11Z
identity: determinism-review-20261116T022511-jay-flowers
tier: draft
---

Determinism reviews dispatched in manifest-delivery mode (diff served via read_diff_chunk) can execute with the working tree checked out at the PR merge base rather than the head: instruction-lint --base then reports an empty delta and tool-inventory --seams reports zero entries even though the PR diff changes .opencode/lib code. Empty Step 1 evidence in that mode reflects the tree state, not the change. Corroborate the changed-file list against the caller's diff manifest, confirm tree state by reading one changed file and comparing it to the diff's base lines, and fall back to reading the changed tests for DE-007 seam verification (the no-BASE degradation row's reading fallback).
