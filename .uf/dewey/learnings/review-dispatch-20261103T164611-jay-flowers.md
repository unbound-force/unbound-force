---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-03T16:46:11Z
identity: review-dispatch-20261103T164611-jay-flowers
tier: draft
---

When a strict `.strict()` zod schema guards a persisted artifact shape but the same schema also validates host-assembled finalization input, provenance fields the plugin itself emits (e.g., `read_only` on dispatch runs) leak into inputs and hard-reject payloads (unbound-force/unbound-force#931: 54 rejections across 39 sessions). Fix this at the input boundary with a `z.preprocess` that strips exactly the known plugin-emitted key before the strict schema, rather than widening the canonical schema: the persisted artifact contract stays strict and the input boundary stays explicit.
