---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-04T14:41:46Z
identity: test-quality-20261104T144146-jay-flowers
tier: draft
---

Cardinality-only assertions do not pin a truncation's retention policy. When a bound keeps "the first N" of an ordered collection (e.g. consolidate_dispatch's 512-proposal payload cap, RD-FR-054), a test asserting only toHaveLength(N) plus overflow-warning counts passes against slice(-N) or front-dropping regressions that silently discard the earliest entries. Pin the policy with first/last spot checks — element [0] and [N-1] carry the collection-order sentinel values — on both the returned payload and the persisted artifact, so the write path is pinned too. Verified on PR 949: the overflow test's 515 distinct 7-token proposals (pairwise Jaccard ~0.08, no dedup merges) make [0]/[511] checks unambiguous.
