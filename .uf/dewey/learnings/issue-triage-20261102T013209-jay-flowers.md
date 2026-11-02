---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:32:09Z
identity: issue-triage-20261102T013209-jay-flowers
tier: draft
---

Router-Pattern collapse issues in unbound-force triage cleanly as VALID/enhancement when they map to determinism.md DE-005 (a fixed judgment-free tool-call sequence MUST collapse into a single intent-level tool) and cite verifiable platform constraints. The validation checklist that held for issue #757 (uf.unleash Step 6 implement-lifecycle collapse): (1) the collapsed sequence contains no judgment steps — every branch is a lookup on the prior call's structured result; (2) platform constraints are cited, not assumed (@opencode-ai/plugin SDK exposes no tool-invocation method on its client, so Replicator MCP calls stay LLM-driven; the 1,800s per-agent-run timeout ceiling forbids whole-phase mega-tools; executeAgentSession is directory-scoped with abort linkage — corroborated by the shipped shared-agent-executor and unleash-deterministic-tools OpenSpec changes); (3) acceptance criteria are testable (tool-enforced failure-stop, byte-identical .opencode/ ↔ internal/scaffold/assets/opencode/ mirrors, make plugin-test + make test); (4) scope matches the OpenSpec tactical tier with one opsx change per phase; (5) cross-repo contingencies are deferred to the sibling repo rather than bundled into the current issue. Two follow-through expectations at spec time: per DE-007 the implementing tasks.md must carry seam-verification tasks (registration surface, documented output shape, round-trips), and Phase-2-style spikes must account for the parented-child permission-ask wedge (learning agent-wedge-20261030T212253-jay-flowers) when running workers in worktree directories.
