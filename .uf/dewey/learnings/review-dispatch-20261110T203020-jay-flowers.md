---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-10T20:30:20Z
identity: review-dispatch-20261110T203020-jay-flowers
tier: draft
---

A bounded tool-result preview that consumers persist verbatim must be complete-or-absent, never field-truncated. In unbound-force's review-dispatch plugin, proposal previews cut `information` to 512 characters with a per-entry `truncated: true` marker; PR #955's parent stored from the preview anyway and 11 of 12 harvested lessons were damaged mid-word. The durable fix (RD-FR-058, issue #969) removes the cut at the source: every field is projected verbatim, over-budget entries are dropped whole and counted by `proposals_truncated`, and a shared post-assembly byte-size check keeps both the summary retrieval and the compact finalization result strictly under 40,960 bytes. Markers on damaged data do not protect consumers; removing the damage does.
