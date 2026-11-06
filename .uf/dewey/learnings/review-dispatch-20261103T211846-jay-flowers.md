---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-03T21:18:46Z
identity: review-dispatch-20261103T211846-jay-flowers
tier: draft
---

Bounded file readers that reject symlinks with lstat and then read with readFile leave a TOCTOU window: readFile re-resolves the path by name, so a file swapped to a symlink (or grown past a size cap) between the check and the read bypasses both guards. Observed in unbound-force PR #949's read_dispatch_artifact (RD-FR-057). To make the check and the read operate on the same inode, open a file handle first (node:fs/promises open with the numeric O_NOFOLLOW flag from fs.constants) and fstat the handle before
