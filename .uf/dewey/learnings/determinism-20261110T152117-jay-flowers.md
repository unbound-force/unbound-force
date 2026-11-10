---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-10T15:21:17Z
identity: determinism-20261110T152117-jay-flowers
tier: draft
---

Determinism Step 1 evidence degrades silently when the review is dispatched from a checkout on the PR's base side: `.uf/scripts/instruction-lint --json --base <base> --changed` exits 0 with an empty `delta.files` (and `tool-inventory --seams` returns an empty list) even though the caller's diff names changed instruction files — the delta compares HEAD against the merge base, and a base-side checkout has nothing between them. Observed on PR #1005: AGENTS.md was modified +3/-1 in the caller's diff, the on-disk AGENTS.md carried the pre-change text, and AGENTS.md nonetheless appeared in the linter's whole-repo token map — proving the empty delta was checkout state, not file-set exclusion. Adjudication: scope the review from the caller's diff, take per-file counts from the caller's changed-files list, note the limitation in the review summary, and never read the empty delta as 'no instruction files changed'.
