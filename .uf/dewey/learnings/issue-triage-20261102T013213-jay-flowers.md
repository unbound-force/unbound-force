---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:32:13Z
identity: issue-triage-20261102T013213-jay-flowers
tier: draft
---

Router-Pattern collapse proposals — collapsing a fixed per-task tool sequence into a single intent-level tool — can be triaged against governance text rather than taste. Determinism pack rules DE-005/DE-006 enumerate exactly which routing concerns belong inside tools ("exit-variant selection, batch scheduling, failure-stop"), so a proposal whose absorbed sequence maps onto that enumeration is directly verifiable. unbound-force#757 (uf.unleash Step 6 per-task lifecycle: worktree create → spawn → wait → merge → conflict scan → cleanup → mark complete) models the full validation signal set: (1) the collapsed sequence is judgment-free, with every branch determined by the previous call's structured result; (2) the judgment boundary is explicit — batch composition, conflict resolution, and (until later phases) worktree MCP calls stay with the LLM; (3) platform constraints are verified rather than assumed — the @opencode-ai/plugin SDK exposes no MCP tool invocation, executeAgentSession is directory-scoped, and the 1,800s per-run timeout ceiling bounds granularity to per-task/per-batch; (4) scope is phase-gated with the cross-repo ask (unbound-force/replicator) deferred as conditional. A collapse proposal missing the explicit judgment boundary or the verified constraints warrants NEEDS-CLARIFICATION even when the sequence itself looks mechanical.
