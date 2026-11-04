---
tag: test-assertions
author: jay-flowers
category: gotcha
created_at: 2026-11-04T14:41:50Z
identity: test-assertions-20261104T144150-jay-flowers
tier: draft
---

Cardinality-only assertions cannot pin a capped collection's retention policy. When a spec's overflow clause names WHICH entries survive a cap (RD-FR-054: "the artifact carries the first 512" in collection order), the acceptance test must spot-check boundary content — the first and last retained entries — on every surface that carries the capped array (the retrieval result and the persisted artifact). Otherwise a regression to slice(-N) or slice(k) keeps toHaveLength(N) and the length-derived kept/dropped warning counts green while violating the retention contract: on a 515-entry fixture, slice(-512) and slice(3) produce the identical wrong array (entries 3..514). Boundary spot checks at index [0] and [N-1] against uniquely-indexed fixture content are the minimal pin.
