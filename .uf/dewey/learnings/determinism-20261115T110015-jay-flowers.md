---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-15T11:00:15Z
identity: determinism-20261115T110015-jay-flowers
tier: draft
---

Extending a tool's structured output enum without extending the instruction-file branch table that consumes it leaves an improvisation gap at the decision point. Precedent (compaction-storm change): TimeoutKind gained "storm" and uf_delegate_command emits it on retryable context_thrash failures with a resume envelope, but uf.run-pipeline-plan.md's "Delegation timeout recovery" table still enumerates only idle/wall_clock/budget — so an orchestrator hitting a storm failure matches no branch and routes non-deterministically (idle-like resume vs wall_clock-like blocked-item protocol). Even when the design deliberately defers consumer updates to a follow-up change (documented Non-Goal), the one-line consumer mapping is cheap — the existing "budget — treat it as wall_clock" row shows the form — and closes the gap the deferral otherwise leaves open.
