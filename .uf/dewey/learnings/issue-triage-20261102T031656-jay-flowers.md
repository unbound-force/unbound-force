---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:16:56Z
identity: issue-triage-20261102T031656-jay-flowers
tier: draft
---

When triaging proposals for standalone lookup tools (e.g., #671 map_verdict for verdict→GitHub-event mapping), check the determinism pack's Router Pattern rules before validating the mechanism, not just the goal. DE-006 names verdict mapping specifically as routing that should execute inside the consuming intent-level tool as constants; a standalone agent-callable lookup that feeds exactly one downstream tool keeps the agent as orchestrator and re-introduces the orchestration hop DE-005 removes. The proposal's goal (zero table-interpretation tokens, Constitution Principle VI / DE-001) can be fully preserved by internalizing the mapping in the consumer (post_github_review, #664) with constants shared via .opencode/lib/ for other internal projections (consolidate_dispatch canonicalDecision). Triage outcome pattern: valid problem + duplicate work item — supersede into the consuming tool's issue, and carry two requirements into that issue's scope so value is not lost: (a) single source of truth for the mapping constants shared across plugins, and (b) the command-file cleanup that realizes the quantified token savings.
