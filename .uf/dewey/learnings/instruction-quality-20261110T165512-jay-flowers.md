---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-10T16:55:12Z
identity: instruction-quality-20261110T165512-jay-flowers
tier: draft
---

When a command's prose is rewritten around a resumable intent-level tool, the resume instruction must name the complete input set, not the subset the tool will skip. "Pass items whose replies were already posted to <tool> again unchanged" reads as an exhaustive item set: a resuming agent can pass only the already-posted items, the tool then finds nothing left to do, records its completion checkpoint, and the command clears progress and ends with the unposted work permanently stranded. Observed on opsx/feedback-reply-tools (PR #1007, uf.address-feedback.md Section 1.7). State the full list explicitly ("pass the full confirmed item list again; the tool skips already-posted replies and resolved threads") — and avoid "unchanged" for inputs whose original values (e.g., composed reply bodies) are not persisted anywhere the resumed session can read.
