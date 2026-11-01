---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-01T13:46:27Z
identity: review-dispatch-20261101T134627-jay-flowers
tier: draft
---

When a manifest-delivered review diff's single chunk exceeds the read_diff_chunk display cap (~64KB — e.g. a 77KB new-file chunk), the persisted full tool output is a single JSON-escaped line that line-based tools cannot usefully read. Working technique: read the patch file itself with the Read tool's offset/limit in slices of ≤400 lines (larger slices get display-truncated), and derive new-file line numbers arithmetically from the hunk header — for a `@@ -0,0 +1,N @@` new-file hunk starting after the diff header, new-file line = patch line − (header offset); in PR #894's patch, model-matrix.ts content began at patch line 130, so new-file line = patch line − 130. Use the grep tool anchored to `^+` patterns on the patch file to locate function definitions and hunk boundaries cheaply (unanchored patterns can match very long diff lines and error out), and remember other review-diff-*.patch files in the same temp dir will match too — filter to the exact manifest path.
