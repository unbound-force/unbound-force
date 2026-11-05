---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-05T15:44:23Z
identity: test-quality-20261105T154423-jay-flowers
tier: draft
---

Threshold contracts need exact-edge tests, not only calibrated ones. The RD-FR-055 proposal-dedup tests were calibrated on the real issue #941 session (duplicate pair Jaccard 0.39, distinct pairs at most 0.10), which left the contractual 0.3 boundary unpinned: a >= to > comparison regression, or a threshold drift anywhere in (0.19, 0.39), passed the whole suite. To pin a rational threshold, construct the minimal integer fixture: Jaccard k/(2n-k) = 0.3 solves to 13k = 6n, so two 13-token sets sharing exactly 6 tokens yield 6/20 = 0.3; 6/20 rounds to the same IEEE-754 double as the literal 0.3, so the assertion is deterministic. Before reporting a boundary as unpinned, check whether an existing test pins it incidentally: the tie-break test's two 6-token Greek-letter sets with different token sequences already pin that exactly-6-token proposals merge via Jaccard rather than exact-sequence equality.
