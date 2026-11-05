---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-05T15:44:29Z
identity: test-quality-20261105T154429-jay-flowers
tier: draft
---

Numeric contract thresholds need exact-edge tests, not just calibrated fixtures. PR 949's RD-FR-055 proposal-dedup tests bracket the Jaccard 0.3 merge threshold only within (0.189, 0.39] — no-merge fixtures score ~0.10/~0.189, merge fixtures 0.39–1.0 — so a silent constant drift to 0.35 or a >= → > operator regression passes the suite. The pinning fixtures: two token sets of 13 unique non-stop-word tokens sharing exactly 6 (6/20 = 0.3; IEEE-754 correctly-rounded 6/20 equals the literal 0.3, so the case is deterministic) must merge, and an exactly-5-token high-Jaccard different-sequence pair must not (sub-floor sets match only on equal token sequences). Also watch incidental pinning: the 6-token floor's exactly-6 edge was held only by fixtures named for length tie-breaking, which a fixture refactor could silently drop.
