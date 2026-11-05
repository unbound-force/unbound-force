---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-11-03T21:18:50Z
identity: openspec-20261103T211850-jay-flowers
tier: draft
---

OpenSpec change artifacts accrete internal inconsistencies when review-iteration insertions preserve old section numbers for cross-reference stability: PR #949's tasks.md physically ordered phases 1, 2, 3, 5, 4 (the "Review follow-ups" phase 5 sits before the "Mirror, documentation, CI parity" phase 4), design.md ordered decisions D1–D5, D7, D6, and task 1.2 still described the superseded first-revision dedup algorithm (overlap coefficient 0.5, 4-token floor) that task 5.2 later replaced (Jaccard 0.3, 6-tok
