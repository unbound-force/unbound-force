---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:16:50Z
identity: issue-triage-20261102T031650-jay-flowers
tier: draft
---

Tool-extraction triage check for the Config Lookup Tools pattern (#659 Phase 2): before endorsing a standalone agent-callable lookup tool, count its consumers and check determinism.md. DE-006 explicitly names verdict mapping, severity sorting, and exit-variant selection as routing that executes inside the consuming tool as constants or pure functions; a standalone lookup the agent calls and then acts on re-introduces the judgment-free orchestration chain that DE-005 (MUST) collapses into a single intent-level tool. Single-consumer lookups (e.g., verdict→GitHub-event feeding review posting) belong as internal constants of the consuming tool (post_github_review, #664). Multi-consumer lookups belong as shared constants in .opencode/lib/ — a single source of truth alongside consolidate_dispatch's canonicalDecision projection — because a second table drifts silently, and these mappings are gate-adjacent: a drifted verdict table posts the wrong GitHub event (e.g., COMMENT instead of REQUEST_CHANGES) and silently weakens the review gate.
