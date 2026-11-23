---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-15T19:59:56Z
identity: determinism-20261115T195956-jay-flowers
tier: draft
---

Prose→tool extraction where the new tool consumes a prior tool's per-item results (shipped pattern: write_feedback_triage_artifact consuming post_feedback_replies outcomes in /uf.address-feedback Section 4.7): carry a stable join key (comment_id) verbatim from the upstream tool's result type, bind items to outcomes inside the new tool with fail-closed duplicate/unmatched/type-mismatch checks, drop input-only keys at render time so the rendered payload satisfies an additionalProperties:false schema, and keep derived values (round numbering, summary counts) as tool-internal derivations that are never agent-supplied inputs. This keeps the agent's role to argument forwarding (judgment at the tool boundary, DE-006) while the join, counting, and envelope rendering stay tool-internal (DE-001), and it makes the spec's seam tests (args parity, golden output shape, write→read round-trip) the drift-proof record of the contract.
