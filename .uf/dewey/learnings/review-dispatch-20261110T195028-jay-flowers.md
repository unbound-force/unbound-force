---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-10T19:50:28Z
identity: review-dispatch-20261110T195028-jay-flowers
tier: draft
---

When a prose guard on tool output fails in practice, fix the tool so the damaged output cannot exist rather than labeling the damage. PR #955 stored 512-character-cut lesson proposals from consolidate_dispatch's summary preview despite RD-FR-056's written "never store from the preview" contract — the per-entry truncated: true marker was ignored by the consuming parent. The full-lesson-proposals fix (RD-FR-058, issue #969) removed the truncation at the source: proposal fields are copied verbatim, entries that do not fit the 40,960-byte budget are dropped whole and counted by proposals_truncated, and no retrieval mode can hand a consumer a partial body. Prose contracts and per-entry damage markers are interim mitigations; the mechanism of record is tool code that makes every consumer path safe.
