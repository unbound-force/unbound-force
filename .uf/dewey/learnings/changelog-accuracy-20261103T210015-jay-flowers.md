---
tag: changelog-accuracy
author: jay-flowers
category: gotcha
created_at: 2026-11-03T21:00:15Z
identity: changelog-accuracy-20261103T210015-jay-flowers
tier: draft
---

CHANGELOG DE-004 line-delta claims drift when they copy the spec-phase prediction instead of the measured diff. In the preflight-declared-output change (unbound-force/unbound-force PR #944), proposal.md and tasks.md predicted "589 -> 598 lines (+9)" for .opencode/skills/pre-flight/SKILL.md before implementation; the final PR diff was +15/-5 (net +10, post-change 599 lines, baseline verified on main), and the CHANGELOG entry repeated the predicted figure. Verify recorded skill-size deltas against git numstat
