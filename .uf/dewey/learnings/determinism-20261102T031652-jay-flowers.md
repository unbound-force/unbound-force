---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:16:52Z
identity: determinism-20261102T031652-jay-flowers
tier: draft
---

When triaging or reviewing proposals to extract deterministic lookup tables from command prose into tools, check the consuming boundary before endorsing a standalone agent-callable lookup tool. The determinism pack (DE-005/DE-006, codified from the Router Pattern per #758/#716) names verdict mapping and severity sorting as lookup-table branching that is mechanical orchestration: it must execute inside intent-level tools as constants or pure functions, and a standalone lookup the agent calls and then acts on re-introduces the orchestration hop the Router Pattern removes. A standalone lookup tool is justified only when a consumer survives outside the consuming tool's path; shared mapping constants belong in .opencode/lib/ as a single source of truth alongside existing projections (e.g., consolidate_dispatch's canonicalDecision) rather than as a second table. Case: unbound-force #671 (standalone map_verdict) vs the re-scoped #664 (post_github_review internal constant).
