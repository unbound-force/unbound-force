---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-16T12:51:32Z
identity: instruction-quality-20261116T125132-jay-flowers
tier: draft
as_of: "unbound-force@db9c0c04b4e23c28da65ae3f5c5209e85114182a"
---

Crash-recovery instruction that governs re-entry into an idempotent tool must direct passing the complete confirmed input list, never the already-processed subset. A subset-naming sentence ("pass items whose replies were already posted") reads as an exhaustive specification of what to pass: a literal re-run passes no pending items, the tool finds nothing new to do, and the completion checkpoint is recorded with pending work never executed. Name the full list ("pass the full confirmed item list again, unchanged — including items whose replies were already posted") and follow it with the tool's idempotence guarantees (skips posted replies and resolved threads from its own records, posts pending replies, resolves threads left open) so the executing agent can verify safety instead of re-deriving a pending subset. Fixed in uf.address-feedback.md Section 1.7 (FRR-FR-003, issue #1023).
