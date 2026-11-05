---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-05T15:44:26Z
identity: test-quality-20261105T154426-jay-flowers
tier: draft
---

Pin inclusive numeric contract boundaries with exact-edge fixtures, and enumerate every existing test's fixture values before calling a boundary covered. PR 949's RD-FR-055 dedup tests exercised the Jaccard >= 0.3 threshold only at 0.39-1.0 (merge) and <= 0.10 (keep), so a >= to > regression in nearIdenticalTokens passed the whole suite; two 13-token sets sharing exactly 6 tokens (6/20 = 0.3, the same IEEE-754 double as the literal) pin both the value and the inclusivity. Incidental coverage does not substitute for a boundary test: a tie-breaking test whose fixtures happen to hold exactly 6 tokens (Jaccard 1.0) pins the floor's upper side without revealing intent, while the lower side (a sub-6-token pair with Jaccard >= 0.3 must be kept) stayed pinned only at 2 tokens, leaving a floor regression from 6 to 5 undetected.
