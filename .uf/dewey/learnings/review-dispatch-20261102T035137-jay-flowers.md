---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:51:37Z
identity: review-dispatch-20261102T035137-jay-flowers
tier: draft
---

Review-dispatch tier heuristics that count all diff lines uniformly over-weight byte-copy mirror syncs. In unbound-force, internal/scaffold/assets/ mirrors canonical .opencode/ and .uf/ files under the test-enforced TestEmbeddedAssets_MatchSource drift invariant, so a scaffold sync carries no original logic — yet PR #653's two sync files (uf.review-pr.md 315+/759-, review-matrix.yaml 3+/6-) contributed 1083 of 1337 diff lines and inflated the dispatch tier from lightweight to heavy. Repos with embedded-mirror invariants should classify mirror-sync paths as a low-signal category and/or exclude them from tier line counts, so review effort tracks the canonical source change — which is itself reviewed when it changes.
