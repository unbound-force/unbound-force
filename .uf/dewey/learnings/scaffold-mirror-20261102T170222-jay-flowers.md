---
tag: scaffold-mirror
author: jay-flowers
category: pattern
created_at: 2026-11-02T17:02:22Z
identity: scaffold-mirror-20261102T170222-jay-flowers
tier: draft
---

In unified diffs, the `index <old-blob>..<new-blob>` header line gives cheap cryptographic proof that two mirrored files are byte-identical: when a canonical asset and its scaffold mirror (e.g. .opencode/plugins/review-dispatch/index.ts and internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts) show identical blob hashes in their respective diff headers, both files were byte-identical before and after the change — mirror sync can be confirmed during review without running the drift gate or diffing the files separately. Verified in the split-host-resolution-failure-classes review where both files carried `index 2a73b366..2b78ac4b`.
