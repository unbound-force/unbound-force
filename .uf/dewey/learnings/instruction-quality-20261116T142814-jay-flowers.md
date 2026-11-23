---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-16T14:28:14Z
identity: instruction-quality-20261116T142814-jay-flowers
tier: draft
as_of: "unbound-force@2318d58fcc5416396271851f3e9bf2ef7c021faf"
---

Crash-recovery instructions around idempotent tools must name the full input set, not a processed subset. uf.address-feedback Section 1.7 told a crash-recovering run to re-pass "items whose replies were already posted" to post_feedback_replies; followed literally, the tool's per-item records made every passed item a no-op and it recorded the reply-comments completion checkpoint with pending replies never posted (issue #1023). The fix (openspec/changes/fix-feedback-reply-followups, FRR-FR-003): instruct "pass the full confirmed item list again, unchanged — including items whose replies were already posted" and state the tool's idempotence guarantees (skips posted replies and resolved threads; posts pending replies; resolves threads left open). Pattern: when a tool deduplicates work via per-item records, re-entry guidance is "re-pass the full confirmed input" — naming any subset silently converts recovery into a no-op. Residual edge to design for: the instruction still needs a named durable source for the pending items' confirmed payloads; decision counts and posted-reply records alone let a recovering agent reconstruct only the already-processed subset.
