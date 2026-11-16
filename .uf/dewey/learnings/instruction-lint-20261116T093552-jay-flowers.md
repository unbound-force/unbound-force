---
tag: instruction-lint
author: jay-flowers
category: gotcha
created_at: 2026-11-16T09:35:52Z
identity: instruction-lint-20261116T093552-jay-flowers
tier: draft
as_of: "unbound-force@62e674575e63293fa77d6db177303ca5a4fccfde"
---

Divisor Linguist runs dispatched with manifest diff delivery can land on a working tree pinned at the PR base SHA (head commit absent from the local object store). The prescribed `.uf/scripts/instruction-lint --json --base <sha> --changed` then exits 0 with a valid empty result and `delta.files: []`, so the PR's new prose is never linted — a zero-finding result that is a scope artifact, not evidence of cleanliness. Verify the delta before trusting it: when delta.files is empty but the caller's diff manifest names instruction files, disclose the unavailable linter coverage in the review summary and run the instruction-clarity checks directly on the chunks retrieved through read_diff_chunk. Line anchors must be derived from the diff hunk arithmetic (new-span start plus added-line offsets), since the on-disk file is the base version.
