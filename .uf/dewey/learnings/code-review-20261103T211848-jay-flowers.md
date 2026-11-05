---
tag: code-review
author: jay-flowers
category: pattern
created_at: 2026-11-03T21:18:48Z
identity: code-review-20261103T211848-jay-flowers
tier: draft
---

In manifest-delivered reviews of repositories with scaffold-mirror invariants, byte-identity between a canonical asset and its embedded mirror can be confirmed without retrieving both full diffs: each file's diff header carries `index <old-blob>..<new-blob>` hashes, and identical before/after blob hashes across the canonical and mirror paths prove identical content (git blob hashes are content hashes). Verified on unbound-force PR #949: .opencode/plugins/review-dispatch/index.ts and its internal/scaffold/as
