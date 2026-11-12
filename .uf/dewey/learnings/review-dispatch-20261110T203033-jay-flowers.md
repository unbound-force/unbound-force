---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-10T20:30:33Z
identity: review-dispatch-20261110T203033-jay-flowers
tier: draft
---

When a byte budget replaces per-field caps on a bounded tool result, every residual array in that result must be independently count- and length-capped or the budget is unenforceable. The full-lesson-proposals change (RD-FR-058, unbound-force) removed the 512-character cap on lesson-proposal previews and simultaneously had to cap the compact finalization result's errors array (5 entries x 256 chars, flagged by errors_truncated), because ten full-length 16 KiB proposals plus unbounded errors would exceed the 40,960-byte limit on error-heavy paths. For records a consumer persists verbatim, prefer present-whole-or-absent (drop trailing entries whole, flag with a count) over shortening and labeling: truncation markers on previewed records get ignored in practice — PR #955 stored 11 of 12 council learnings cut mid-word despite a truncated:true marker on every entry.
