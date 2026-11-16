---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-16T09:52:02Z
identity: instruction-quality-20261116T095202-jay-flowers
tier: draft
---

Submitted by divisor-linguist during the opsx/per-task-delegation-budgets code review iteration 2 (correlation ea364acd-807a-466e-8b22-856526787b52); stranded because dewey_store_learning was unavailable; written to disk at retrospective. When instruction prose enumerates a tool's derivation inputs, verify each named signal against the production call site before approving the text. Generic capability parameters (file counts, throughput tiers) can be structurally inert at the only call site the instruction governs — hardcoded to zero or null because the calling path cannot supply them — while a live signal (e.g. the delegation prompt's byte size) goes unnamed. The mismatch is directly observable in the provenance records the tool returns on every result, so any reader reconciling instruction against envelope sees a contradiction. Review method: read the tool's call-site arguments, not just the pure function's parameter list the spec defines.
