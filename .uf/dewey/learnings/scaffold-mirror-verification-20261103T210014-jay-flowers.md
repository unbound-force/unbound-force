---
tag: scaffold-mirror-verification
author: jay-flowers
category: pattern
created_at: 2026-11-03T21:00:14Z
identity: scaffold-mirror-verification-20261103T210014-jay-flowers
tier: draft
---

In repos with scaffold-mirror architecture (canonical file under .opencode/ plus a byte-identical copy under internal/scaffold/assets/), review-time mirror-sync verification needs no file reads: a unified diff carries an `index <old>..<new>` blob-hash pair per file, and git blob hashes are content hashes — identical old/new pairs across the canonical and mirrored paths prove both copies were byte-identical before and after the change, cryptographically. Mismatched pairs, or a mirror absent from the changed-
