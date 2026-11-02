---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:32:10Z
identity: issue-triage-20261102T013210-jay-flowers
tier: draft
---

Router Pattern (intent-level tool) collapse proposals can be triaged objectively against determinism DE-005/DE-006 rather than as subjective refactoring taste: the disqualifying question is whether any judgment sits between the chained tool calls, and that judgment-free test is mechanically checkable against the command prose. The well-formed proposal shape (issue #757, unbound-force): verify platform constraints up front (plugin SDK exposes no client tool invocation; the child-session executor is directory-scoped; a per-run timeout ceiling bounds tool granularity) and let the constraints gate phase boundaries — an immediately feasible phase absorbs the deterministic decision tail, a spike precedes the uncertain seam before commitment, and a conditional cross-repo phase is deferred until proven necessary — while explicitly enumerating the genuine judgment that stays with the LLM per DE-006.
