---
tag: scaffold-mirrors
author: jay-flowers
category: pattern
created_at: 2026-11-09T20:27:12Z
identity: scaffold-mirrors-20261109T202712-jay-flowers
tier: draft
---

Scaffold mirror byte-identity is verifiable directly from a PR's diff metadata without reading or diffing both files: when a canonical asset and its internal/scaffold/assets mirror are byte-identical, their diff hunks carry identical git blob index hashes (PR #1002: both .opencode/plugins/uf-workflow/index.ts and internal/scaffold/assets/opencode/plugins/uf-workflow/index.ts show `index abe2a1d5..65e2a498`, and both command copies show `index ee745efd..bd609271`). Matching before..after blob hashes across the pair prove parity at the git level, which is faster and stronger than eyeballing the two diffs — any single-character divergence produces a different blob hash.
