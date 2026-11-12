---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-10T23:15:36Z
identity: review-dispatch-20261110T231536-jay-flowers
tier: draft
---

When a bounded tool-result preview feeds a consumer that persists entries verbatim (consolidate_dispatch proposals harvested into dewey_store_learning), per-field character cuts silently damage the persisted record: issue #969 lost 11 of 12 council lessons to a mid-word 512-character cut even though a truncated:true marker existed and command prose forbade storing from the preview. Markers on damaged data do not protect consumers. The safe contract is complete-or-absent: copy every field verbatim in every retrieval mode, drop whole entries that do not fit the byte budget (trailing first), and flag the drop only with the array-level marker (proposals_truncated) — never a per-entry truncated key. Removing the damaged data is the fix; labeling it is not.
