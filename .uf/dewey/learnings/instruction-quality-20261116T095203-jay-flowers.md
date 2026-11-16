---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-16T09:52:03Z
identity: instruction-quality-20261116T095203-jay-flowers
tier: draft
---

Submitted by divisor-linguist during the opsx/per-task-delegation-budgets code review iteration 2 (correlation ea364acd-807a-466e-8b22-856526787b52); stranded because dewey_store_learning was unavailable; written to disk at retrospective. Compressing a spec clause like "no new global default is introduced; top-level callers without a budgeted ancestor behave as before" into command prose as "unbudgeted top-level invocations keep the existing delegation defaults" drops the spec's scoping and creates a two-way ambiguity. In the per-task-delegation-budgets change, the spec's "behave as before" was scoped to idle semantics, but the compressed plural "defaults" let a reader conclude the flat 30-minute wall clock also survived at top level — hiding whether the change's headline behavior applies in the primary (top-level /uf.unleash) scenario. Fix pattern: scope preserved-defaults clauses to the specific default preserved ("the existing idle-window default; the derived wall clock above still applies") instead of an unscoped plural.
