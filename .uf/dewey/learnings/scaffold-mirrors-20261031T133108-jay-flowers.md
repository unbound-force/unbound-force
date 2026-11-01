---
tag: scaffold-mirrors
author: jay-flowers
category: pattern
created_at: 2026-10-31T13:31:08Z
identity: scaffold-mirrors-20261031T133108-jay-flowers
tier: draft
---

Diff-based mirror-sync verification: when a git diff touches a canonical file and its scaffold mirror (e.g. .opencode/commands/uf.review-pr.md and internal/scaffold/assets/opencode/commands/uf.review-pr.md), identical `index <old>..<new>` blob-hash lines on both entries prove byte-identical content at base and head — mirror sync confirmed directly from the diff, no drift-test run needed. Git blob hashes are content-addressed, so equal hashes mean equal bytes (up to abbreviation-collision improbability; the drift tests remain the CI-authoritative gate). In PR #883 (opsx/forward-aggregate-correlation-id) all five command mirror pairs plus the plugin mirror (plugins/review-dispatch/index.ts, index 5ff98ede..9d1afb37 on both sides) carried identical index lines, closing the mirror question in seconds. Fall back to running the drift tests when a pair's index lines differ or a mirror is absent from the diff.
