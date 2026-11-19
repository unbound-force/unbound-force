---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-16T15:41:01Z
identity: determinism-20261116T154101-jay-flowers
tier: draft
as_of: "unbound-force@c2dfd4b8f15f9ae106c599aef9a0418bf075dd93"
---

OpenSpec tasks.md files that annotate tasks with "classification:" lines (anticipating /uf.unleash implement-phase delegation through uf_delegate_command) should pre-state the full DE-008 binding for every deterministic task, not just the classification: the skill/script reference, or a substantive skill_absent_reason naming what was searched and what was found. The delegation tool enforces the contract fail-closed at implement time (SKILL_REFERENCE_REQUIRED / SKILL_NOT_FOUND), so a plan that carries only the classification half forces ad-hoc composition of the binding mid-implementation, and the recorded audit trail no longer traces back to the plan. In the unify-checknotonmain-with-the-shared-branch-resolution change (2026-11), task 3.1 modeled the complete pattern ("skill/script reference: repo-root `Makefile`, targets `build`, `lint`, `coverage-gate`") while deterministic tasks 1.4 and 3.3 carried classifications with no reference or absence reason — the Makefile test target runs ./... rather than scoped packages, and openspec validate has no repo wrapper, so honest absence reasons were the missing binding.
