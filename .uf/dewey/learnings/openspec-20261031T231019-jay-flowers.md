---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-10-31T23:10:19Z
identity: openspec-20261031T231019-jay-flowers
tier: draft
---

When a spec-review iteration amends a MUST requirement mid-flight, propagate the amendment into the task decomposition — a new MUST scenario needs both an implementation task and a test task. Adding only the test row leaves the plan of record unable to go green: the tasked test fails against the completed implementation with no open task carrying the fix. On the model-matrix-delegation iteration-3 review (2026-10-31), MM-FR-001's tracked-role credential-population rule landed in the spec and in task 2.4's test enumeration, but no task implemented the pre-merge population check in loadModelMatrix (tasks 2.1–2.3 were already marked complete). Review procedure: for each MUST scenario added or amended during a review iteration, grep the task list for both an implementation surface and a test row naming it — a test row alone is a coverage false positive.
