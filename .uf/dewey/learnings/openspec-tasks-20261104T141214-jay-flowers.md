---
tag: openspec-tasks
author: jay-flowers
category: gotcha
created_at: 2026-11-04T14:12:14Z
identity: openspec-tasks-20261104T141214-jay-flowers
tier: draft
---

In an OpenSpec change, a review-followup task that supersedes an earlier task's specified algorithm must propagate the supersession back through the change's own artifacts in the same increment: restate or annotate the earlier task text with a supersession cross-reference, renumber phase headings and design decisions into execution order, and update test names that cite superseded parameters. Observed in unbound-force PR 949 (review-dispatch-consumer-bounds): after task 5.2 switched deduplication to Jaccard >= 0.3 with a 6-token floor, task 1.2 still specified the first-revision rule (overlap coefficient >= 0.5, 4-token floor) under an RD-FR-055 citation, phase headings ran 1,2,3,5,4, design decisions ran D1-D5,D7,D6, and one test name still said "fewer than four tokens" — while RD-FR-055, design D2, and PROPOSAL_DUPLICATE_JACCARD carried only the final rule. design.md D2 shows the correct pattern: state the final rule and document the first revision's failure inline. Practical rule: when a follow-up task changes a rule an earlier task specified, grep the change directory (and its tests) for the superseded parameters — thresholds, token counts, algorithm names — and update or annotate every hit before the review loop closes.
