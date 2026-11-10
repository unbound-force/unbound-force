---
tag: documentation-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-10T16:55:01Z
identity: documentation-pipeline-20261110T165501-jay-flowers
tier: draft
---

When a change reorders or relocates a decision gate that another unarchived OpenSpec change pins in a gate-inventory scenario, the invalidation is requirement-supersession class and must be recorded the same way: an interaction note in the changing proposal's Modified Capabilities section, or a scenario update in the pinned change before archive. Verified on PR #1007 (feedback-reply-tools): moving the /uf.address-feedback 4.6 resolve gate before posting left question-tool-presentation-contract's FR-008 "Section 4.6 resolve threads gate" scenario GIVEN ("reply comments were posted for accepted items") unsatisfiable, while the same proposal correctly recorded the AFS-FR-002 supersession for feedback-loop-scope — the asymmetry is the review gap. Without the note, archiving the pinned change promotes a scenario that contradicts the shipped flow into canonical specs. Generalizes documentation-pipeline-20261025T185035 (filed tracking-issue bodies go stale when mechanisms change) from issue bodies to unarchived sibling changes' spec scenarios.
