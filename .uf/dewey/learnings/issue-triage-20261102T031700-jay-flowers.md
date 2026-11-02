---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:17:00Z
identity: issue-triage-20261102T031700-jay-flowers
tier: draft
---

Triage check for plugin-tool proposals in unbound-force repos: evaluate the proposed tool shape against the determinism pack's intent-level orchestration rules, not just its token-saving goal. DE-006 [SHOULD] names "verdict mapping" among routing decisions that should execute inside an intent-level tool as constants or pure functions; DE-005 [MUST] collapses judgment-free tool chains (call lookup, interpret the structured result, invoke the next tool) into a single intent-level tool. A standalone agent-callable lookup tool satisfies DE-001 extraction but preserves the agent-as-orchestrator hop DE-005 removes. Issue #671 (map_verdict) exemplifies the pattern: valid problem, typed contract, testable acceptance criteria — yet the member-comment disposition re-scopes the verdict mapping into post_github_review (#664) as an internal constant, with constants in .opencode/lib/ shared with consolidate_dispatch's canonicalDecision projection for a single source of truth. Testing corollary: prose→tool migration issues must carry DE-007's three seam layers (registration surface, golden output shape, round-trip/command-layer invocation pinning) in their test acceptance criteria — "tests cover all mappings" alone is incomplete.
