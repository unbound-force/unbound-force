---
tag: spec-artifacts
author: jay-flowers
category: gotcha
created_at: 2026-11-04T14:12:08Z
identity: spec-artifacts-20261104T141208-jay-flowers
tier: draft
---

A mid-change algorithm revision is only fully landed when every artifact that restates the rule is reconciled in the same increment: the original tasks.md task text (which reads as live instruction, not history), test names that embed the old threshold, and phase/decision numbering shifted by the insertion. In PR 949 (review-dispatch-consumer-bounds), the review-loop task 5.2 switched proposal dedup from overlap coefficient >= 0.5 with a 4-token floor to Jaccard >= 0.3 with a 6-token floor; design.md D2 recorded the supersession as rejected-alternative history, but task 1.2 restated the old rule uncited, one added test name kept "fewer than four tokens" while citing RD-FR-055, and the inserted phase 5 / design D7 landed physically before phase 4 / D6 without renumbering — a linear reader reconstructed two contradictory dedup rules. Fix pattern for review follow-ups that supersede an earlier task's rule: restate the final rule in the earlier task or add an explicit "(superseded by N.M)" cross-reference, rename tests whose names embed the replaced threshold, and renumber inserted phases/decisions so numbering follows execution order (MP-018).
