---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-16T13:30:13Z
identity: instruction-quality-20261116T133013-jay-flowers
tier: draft
as_of: "unbound-force@db9c0c04b4e23c28da65ae3f5c5209e85114182a"
---

Re-entry instructions that name a subset of items to re-pass ("pass items whose replies were already posted") read as an exclusive filter: an executing agent passes only that subset, the idempotent tool skips all of it as already-done, and the run records its completion checkpoint with the pending work never done. When a tool is idempotent over its own per-item record, name the full set and let the tool filter — "pass the full confirmed item list again, unchanged — including items whose replies were already posted" — and state the tool's skip-set (posted replies, resolved threads) next to it so the agent does not re-filter at the wrong layer. Pattern from the /uf.address-feedback Section 1.7 fix (issue #1023, MEDIUM 1).
