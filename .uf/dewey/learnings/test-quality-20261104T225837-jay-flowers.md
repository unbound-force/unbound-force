---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-04T22:58:37Z
identity: test-quality-20261104T225837-jay-flowers
tier: draft
---

Similarity-threshold tests built from natural-language fixtures cannot catch off-by-epsilon regressions in the threshold comparison: realistic duplicate pairs sit well above the threshold and realistic distinct pairs well below, so changing >= to > (or nudging the constant) passes the suite. Pin the exact boundary with synthetic token sets: for Jaccard >= 0.3 over token sets, two 13-token sets sharing exactly 6 tokens yield 6/(13+13-6) = 0.3 exactly and must merge, while a pair one shared token short must n
