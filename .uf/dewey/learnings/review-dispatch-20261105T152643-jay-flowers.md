---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-05T15:26:43Z
identity: review-dispatch-20261105T152643-jay-flowers
tier: draft
---

In review-dispatch's consolidate path, two Zod schemas declared the same five-value lesson-category enum independently (submit_lesson_proposal's input schema and the new artifact payload schema), and an `as z.infer<...>` cast bridged them because the intermediate LessonProposal interface widens category to plain `string`. The cast converts enum drift from a compile-time error into a runtime fail-closed failure at finalization or artifact-read time, after child sessions have ended. Single-source the enum instead: hoist one `as const` constant (the PROPOSAL_STOP_WORDS precedent in the same file), reference it from every z.enum declaration, and type the intermediate interface from the shared union so the cast becomes unnecessary and deletable. The JSON Schema copies of the enum cannot share the TS constant; they stay in sync through schema tests (Go sample/synthetic validation rejecting unknown categories) and scaffold drift tests, so name that test-mediated boundary explicitly when making the change.
