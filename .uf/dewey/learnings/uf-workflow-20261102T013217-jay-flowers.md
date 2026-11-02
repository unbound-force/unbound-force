---
tag: uf-workflow
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:32:17Z
identity: uf-workflow-20261102T013217-jay-flowers
tier: draft
---

Feasibility boundary for Router-Pattern lifecycle collapse in the uf-workflow plugin suite (corroborated while triaging unbound-force/unbound-force#757): the @opencode-ai/plugin SDK (1.4.10) exposes no tool-invocation method on its client — Replicator MCP tools (forge worktree create/merge/cleanup, comms reservations) can only be invoked by the LLM, never from plugin code. Plugin code CAN spawn directory-scoped child sessions (executeAgentSession with query.directory — the mechanism behind uf_delegate_command and dispatch_review_runs). Agent runs are bounded by a 1,800s (1,800,000 ms) timeout ceiling, so intent-level tools must stay at per-task or per-batch granularity; whole-phase mega-tools exceed the ceiling. Any lifecycle-collapse design must therefore bracket LLM-driven MCP calls around plugin-native batch execution rather than absorbing the full lifecycle into one tool.
