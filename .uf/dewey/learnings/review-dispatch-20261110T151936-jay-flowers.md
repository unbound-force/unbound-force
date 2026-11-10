---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-10T15:19:36Z
identity: review-dispatch-20261110T151936-jay-flowers
tier: draft
---

Scaffold mirror parity (canonical .opencode/ file vs internal/scaffold/assets/ copy) can be verified at review time without reading both files: the git diff index lines carry blob hashes, and identical `index <old>..<new>` hashes across the canonical and mirror file sections prove byte identity (verified on PR #1007: uf-workflow/index.ts and uf.address-feedback.md canonical/mirror pairs shared identical index hashes 9e585e327 and be8075a69). The Go drift tests (task42CanonicalAssets / TestEmbeddedAssets_MatchSource) are the enforcement mechanism; the diff blob hashes are the reviewer's cheap evidence. Use this instead of re-reading 1000-line mirror files.
