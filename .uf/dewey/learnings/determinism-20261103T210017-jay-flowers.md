---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-03T21:00:17Z
identity: determinism-20261103T210017-jay-flowers
tier: draft
---

DE-004 line-delta claims go silently stale when a review-fix commit amends instruction text after the counts were recorded. In PR #944 (preflight-declared-output), the skill-size claim "589 -> 598 lines (+9)" appears identically in proposal.md, tasks.md, and CHANGELOG.md, but the final diff nets +10 (+15/-5): the base skill file is 589 lines on main and the head file is 599, so the recorded before/after pair is arithmetically inconsistent with the diff under any counting method (598 - 589 = 9 ≠ 10). The error propagates because each artifact copies the prediction without re-measuring; re-run the line count after the final fix commit and update every copy.
