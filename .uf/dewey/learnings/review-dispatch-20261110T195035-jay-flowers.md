---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-10T19:50:35Z
identity: review-dispatch-20261110T195035-jay-flowers
tier: draft
---

When adding a byte budget to a bounded tool result, audit every non-droppable field for a cap and mirror the termination argument in the spec clause. buildSummaryRetrieval's RD-FR-049 budget terminates because verdict_reason is capped at 512 characters and every scalar is bounded, with a drop order over all four arrays; buildCompactFinalization's RD-FR-058 budget drops only proposals, so the uncapped `errors` array can still push the compact finalization result past 40,960 bytes on a validation-rejected payload — a spec MUST ("stays strictly under 40,960 bytes") with no enforcement mechanism on that path. Either cap the residual field (count plus per-string length, like the warnings projection) or scope the spec claim to the droppable arrays.
