---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-10T20:30:17Z
identity: review-dispatch-20261110T203017-jay-flowers
tier: draft
---

When a bounded tool result previews records that a parent persists verbatim — lesson proposals, config snippets, audit entries — per-field truncation silently corrupts the persisted record even when the entry carries a truncated: true marker: PR #955 stored 11 of 12 council lessons cut mid-word despite the flag. Presentation data (findings, warnings) tolerates display caps; persistence records do not. The safe bound is whole-or-absent: copy every field verbatim, drop trailing entries whole under a post-assembly byte budget, and count drops with an overflow flag (proposals_truncated) so no retrieval mode can hand a consumer a shortened record to store. When auditing such a bound, verify three properties: the drop loop terminates (monotonic pops plus an explicit break once nothing remains droppable), every non-droppable field stays count- and length-capped so the budget holds once the droppable array empties, and the byte measure is UTF-8-correct (TextEncoder on the canonical JSON) so multibyte entries cannot slip past the budget.
