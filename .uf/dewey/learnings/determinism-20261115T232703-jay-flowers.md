---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-15T23:27:03Z
identity: determinism-20261115T232703-jay-flowers
tier: draft
as_of: "unbound-force@505dedf886fceeb2e552b7c1cf7633e2b6c36310"
---

A dispatched determinism review can run with the working tree still at the PR base commit: instruction-lint --base <sha> --changed then reports delta.files: [] with mergeBase equal to the supplied base, and tool-inventory --seams reports [], even though the caller's diff manifest lists changed instruction files and lib code; the parent's pre-flight corroborates with "SKIP — no in-scope files". Detection: mergeBase == supplied base AND empty delta.files while the caller's diff names .opencode/ instruction paths. Handling: treat the caller's diff manifest as the changed-file source of record; confirm the tree state by reading one changed file from disk (base content, no PR additions, confirms it); review changed prose by reading diff chunks plus the full base files; review DE-007 by reading whether the change adds or edits tests referencing the changed surfaces; record the scope limitation in the summary; do not attribute the evidence gap to the PR. Verified on the PR #1129 determinism review (tree at 505dedf, PR head a2cdb86).
