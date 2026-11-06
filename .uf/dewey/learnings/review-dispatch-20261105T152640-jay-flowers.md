---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-05T15:26:40Z
identity: review-dispatch-20261105T152640-jay-flowers
tier: draft
---

An `as` cast between two separately declared Zod enums compiles cleanly while their value sets drift apart; the drift surfaces only at runtime as a fail-closed validation failure. In unbound-force review-dispatch (PR 949, RD-FR-054), PayloadLessonProposalSchema re-declared the five lesson-category values already declared by SubmitLessonProposalInputSchema and by the existing LESSON_CATEGORY_VALUES as-const tuple, so the consolidate path needed `category: p.category as z.infer<typeof PayloadLessonProposalSchema>["category"]` — runtime-sound only because partitionRunSubmissions re-validates proposals upstream. Accepted fix pattern: declare the value set once as an `as const` tuple, derive every Zod enum from it (z.enum(LESSON_CATEGORY_VALUES)), type the intermediate interface field as (typeof LESSON_CATEGORY_VALUES)[number] so the widened interface stops erasing the enum, and delete the cast — drift then fails compilation at the payload-construction site. Mirror the edit in the scaffold asset copy (byte-parity enforced by TestEmbeddedAssets_MatchSource). JSON Schema registry copies cannot reference the TS constant: keep them in sync via schema-behavior tests (valid/invalid category rejection cases) and, optionally, a parity assertion that the JSON enum equals the constant.
