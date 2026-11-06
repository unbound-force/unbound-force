---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-05T15:26:47Z
identity: review-dispatch-20261105T152647-jay-flowers
tier: draft
---

In .opencode/plugins/review-dispatch/index.ts the five-value lesson-category set ended up declared four ways by PR 949: the pre-existing SubmitLessonProposalInputSchema z.enum, the pre-existing hoisted `LESSON_CATEGORY_VALUES` as-const constant (already used by the summary-projection check), the new PayloadLessonProposalSchema z.enum, and the JSON Schema lessonProposal def. The new schema was joined to the loosely-typed `LessonProposal.category: string` field via an `as z.infer<...>` narrowing cast, which keeps the build green while the declarations drift and defers the mismatch to strict-schema rejection at finalization — the dispatch-stranding class. Discipline: before declaring any value-set enum in this codebase, grep for an existing `as const` constant of the same set (LESSON_CATEGORY_VALUES; UNLEASH_EXIT_VARIANTS precedent) and reference it from every z.enum and derived type — a new inline declaration of an already-constant set is a smell that will demand a silencing cast at its first join point. A TS constant cannot feed the generated/pinned JSON Schema artifacts; that boundary stays aligned via the Go schema test against the sample, not via the constant.
