---
tag: scaffold-mirrors
author: jay-flowers
category: pattern
created_at: 2026-11-05T17:48:38Z
identity: scaffold-mirrors-20261105T174838-jay-flowers
tier: draft
---

In repos with byte-identical embedded-mirror invariants (unbound-force: .opencode/ and .uf/ canonicals mirrored under internal/scaffold/assets/), the diff hunk headers themselves prove mirror sync: when a canonical file and its mirror carry the same `index <old-blob>..<new-blob>` hash pair in the patch, the two blobs are byte-identical by git's content addressing, and the mirror's diff need not be re-read line by line. Verified on PR #949 (review-dispatch-consumer-bounds): the plugin (0cbd7e53..281cc368) and the schema (c431f99e..bce3fff4) each showed one shared hash pair across canonical, .uf/, and embedded copies. Use the hash comparison as the primary review check and spend retrieval budget on the canonical; fall back to reading the mirror only when hashes differ or the file is new (no old blob to compare).
