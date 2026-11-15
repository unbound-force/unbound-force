---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-10T21:26:00Z
identity: review-dispatch-20261110T212600-jay-flowers
tier: draft
---

When a bounded tool result swaps per-field character caps for a container-level byte budget that drops entries whole, the budget loop's early-exit invariant — every remaining field is independently capped — becomes load-bearing. The loop stops as soon as the droppable array is empty, so a field added later without a cap lets the result exceed the budget with no truncation flag. Pin the floor with a worst-case test: empty droppable array, every other array at its cap, assert the serialized total stays under budget.
