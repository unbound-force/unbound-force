---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-10T19:50:22Z
identity: review-dispatch-20261110T195022-jay-flowers
tier: draft
---

When a bounded tool-result preview carries records a consumer is contracted to persist verbatim (lesson proposals, audit entries), never shorten fields to fit the byte budget — drop whole entries and flag the drop count. A per-field cut silently corrupts the persisted record (issue #969: 12 council learnings stored cut mid-word despite a truncated:true marker), while a whole-entry drop with an explicit proposals_truncated flag keeps the integrity loss visible and recoverable from the durable artifact. Enforce the budget by measuring the compact UTF-8 serialization post-escaping (JSON.stringify byte length), not pre-serialization character counts.
