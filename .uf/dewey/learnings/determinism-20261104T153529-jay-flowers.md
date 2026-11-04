---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-04T15:35:29Z
identity: determinism-20261104T153529-jay-flowers
tier: draft
---

DE-001 extraction feasibility boundary (PR 949 lesson storage): when a mechanical enumeration loop's terminal calls target a tool on a separate MCP server (dewey_store_learning on the Dewey server), the consuming plugin cannot execute the loop, and the AGENTS.md determinism gate's 'where an executable tool is feasible' qualifier governs — the residual agent-executed loop is an accepted state, not a gate violation. The retirement path is an intent-level batch-store tool on the tool-owning server that takes the proposal array as a validated parameter (one call replacing N), not a partial extraction on the consumer side. Interim resilience must still be prescribed in prose; PR 949's rewritten lesson-storage sections are the precedent: Dewey unavailable -> informational skip with verdict unchanged, artifact read error -> fail closed, empty proposals section -> none submitted, legacy plugin with preview-only proposals -> stranded report, and never store lessons from capped previews or no-provenance retrievals.
