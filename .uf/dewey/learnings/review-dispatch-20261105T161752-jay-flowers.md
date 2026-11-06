---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:17:52Z
identity: review-dispatch-20261105T161752-jay-flowers
tier: draft
---

When a change adds a cardinality (or numeric) bound to two parallel validation surfaces — the plugin's Zod schema and the JSON Schema registry — the boundary test must assert the N+1 rejection on BOTH surfaces directly. A truncating producer's overflow test does not verify the validation surfaces: the producer's output is already capped (e.g., consolidate_dispatch keeps the first 512 proposals and emits a proposals_overflow warning), so the strict-schema rejection a direct caller would hit is never exercised. Demonstrated on PR #949 (RD-FR-054): entry-level length boundaries were tested on both surfaces (16_384 valid / 16_385 rejected) while the array-level 513-entry rejection was tested on neither, leaving the Zod .max(512) and JSON Schema maxItems 512 caps unverified. Fix shape: one 513-entry rejection assertion through the Zod parse (finalizeReviewDispatch/errorsFor asserting the payload.proposals path fragment, matching the file's existing assertion style) and one through the JSON Schema validate (schemas.ValidateBytes on a 513-element array), placed in the same test files that cover the field's entry shapes.
