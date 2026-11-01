---
tag: golden-tests
author: jay-flowers
category: gotcha
created_at: 2026-10-31T13:31:11Z
identity: golden-tests-20261031T133111-jay-flowers
tier: draft
---

A dispatched review child may execute with the working tree checked out at the PR BASE (merge-base), not the head — verify which commit the tree is on before treating grep/read line numbers as head-state positions. Verified during the forward-aggregate-correlation-id council review (PR #883, 2026-10-31): grepping .opencode/commands/*.md for golden-test site labels returned line numbers 13-35 lower than every refreshed label in post-edit regions (parse_tasks at 908 vs label 924, format_unleash_demo at 1755 vs label 1790, finalize_review_dispatch at 531 vs 543), which initially reads as label corruption and would produce a false finding. Discriminator: read a line the diff adds (e.g. the toolContext(sessionID = "parent-session") fixture signature in dispatch-review-runs.test.ts, or the file's total line count vs numstat) — its absence proves the tree is at base. Correct verification method: map each label to a base grep hit by subtracting the diff's cumulative per-hunk offsets (per-file numstat net confirms the total); when every label lands exactly on a base grep hit plus its offset, the labels are correct at head. All ~25 labels in golden-output-shape.test.ts and dispatch-review-runs.test.ts passed this check exactly. Rule: before flagging file:line label drift in a review, first prove the tree state, then re-derive head positions from the diff — never compare labels against base-state grep output directly.
