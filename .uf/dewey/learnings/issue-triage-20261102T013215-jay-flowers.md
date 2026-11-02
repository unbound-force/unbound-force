---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:32:15Z
identity: issue-triage-20261102T013215-jay-flowers
tier: draft
---

When triaging Router Pattern (intent-level tool collapse) proposals, validity hinges on three structural signals: (1) the target sequence is judgment-free — every branch is a lookup on a structured result (determinism pack DE-005); (2) the proposed tool boundaries fall on verified platform constraints — which operations are LLM-only MCP calls versus plugin-native child-session spawning, and whether granularity respects the 1,800s per-run timeout ceiling (per-task/per-batch, never whole-phase); and (3) scope is gated — spike-first phases and cross-repo contingencies deferred with explicit "only if needed" conditions rather than committed. Issue #757 (uf.unleash Step 6 per-task lifecycle collapse) exhibits all three: its phase boundaries (decision-tail collapse, batch spawn, cross-repo option) track the @opencode-ai/plugin SDK constraints exactly, making the feasibility claims objective rather than aspirational. Weak-signal variants to reject or clarify: proposals that name a tool but leave the boundary undefined, or that absorb judgment steps (conflict resolution, batch composition) into the tool against DE-006.
