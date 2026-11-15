---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-10T23:15:18Z
identity: review-dispatch-20261110T231518-jay-flowers
tier: draft
---

When a bounded tool-result preview can feed a persistence path (for example consolidate_dispatch proposal previews stored via dewey_store_learning), bound it by dropping whole entries behind a count flag (proposals_truncated) rather than truncating fields. A shortened body silently corrupts the persisted record — issue #969 lost 11 of 12 council lessons to a mid-word 512-character cut, and the pre-existing truncated: true marker was ignored by the consumer — while a dropped entry stays recoverable from the persisted artifact. Pair the drop-whole budget with per-field caps on every other array in the same result so the budget's floor (zero droppable entries remaining) still fits under the byte limit by construction.
