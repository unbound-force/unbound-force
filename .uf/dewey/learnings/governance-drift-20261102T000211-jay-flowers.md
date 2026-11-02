---
tag: governance-drift
author: jay-flowers
category: gotcha
created_at: 2026-11-02T00:02:11Z
identity: governance-drift-20261102T000211-jay-flowers
tier: draft
---

Fixing a hardcoded set enumeration in an instruction template requires sweeping the entire file for count-derived references, not just the cited lines. Verified during #824 triage: unbound-force's constitution-check agent (.opencode/agents/constitution-check.md) hardcodes the org-principle count in four places — the procedure loop (line 43 "For each of the four org principles (I, II, III, IV)"), the output-template repeat marker (line 95), the report summary field (line 99 "count, always 4"), and the ALIGNED decision criteria (line 108 "All four findings"). Patching only the two cited lines would leave the report template and verdict logic computing over a stale cardinality, so a governance verdict can return ALIGNED while silently omitting active principles (V Security by Default and VI Deterministic Execution at constitution v1.4.0). Prefer deriving the enumeration from the canonical document at run time ("for each principle in the org constitution's Core Principles section") per determinism pack DE-001, rather than replacing one stale count with the next.
