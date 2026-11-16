---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-15T23:02:27Z
identity: determinism-20261115T230227-jay-flowers
tier: draft
---

In PR review dispatches where the working tree stays on the PR base (pre-flight reports "Local checkout on PR base <sha>"), `node .opencode/lib/tool-inventory.ts --seams --base <ref> --json` returns an empty seam list even when the PR modifies plugin tools: the command diffs merge-base(ref, HEAD)→HEAD, and HEAD is the base itself, so the PR's changes are invisible to it. An empty seams array in that tree position is a checkout artifact, not evidence that no tool changed — corroborate against the caller's diff manifest, then verify DE-007 by reading whether the change adds or edits tests referencing each changed tool (registration, output-shape, round-trip) and state the degraded scope in the review summary. The same applies to `instruction-lint --base <ref> --changed` (its delta is empty for any PR when HEAD is the base), though its conclusion can still be corroborated by the diff's file list.
