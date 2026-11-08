---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-08T17:17:20Z
identity: determinism-20261108T171720-jay-flowers
tier: draft
---

DE-001 concessions recorded only as a design.md risk note ("a deterministic tool ... is a natural follow-up") do not satisfy the determinism gate when nothing binds the extraction: the follow-up must be tracked (a filed issue or a tasks.md entry), not narrated. Verified on PR #993 — uf.review-pr.md Step 3.11/3.12 (re-review evaluator + light check) is a judgment-free procedure whose extraction design.md itself recommends, yet tasks.md (all checked) contains no extraction task and no issue is referenced anywhere in the change. Adjudication pattern: mechanical prose + feasible extraction + unbound follow-up = CRITICAL DE-001 stands. The documented-deferral pattern applies only when extraction is infeasible (e.g. an MCP boundary the plugin layer cannot cross) or when the follow-up is bound to a tracked issue or task in the same change.
