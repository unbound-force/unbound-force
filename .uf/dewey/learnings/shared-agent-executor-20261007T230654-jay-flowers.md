---
tag: shared-agent-executor
author: jay-flowers
category: gotcha
created_at: 2026-10-07T23:06:54Z
identity: shared-agent-executor-20261007T230654-jay-flowers
tier: draft
---

When a function has multiple early-exit error paths that each construct identical provenance placeholder objects (as happened in dispatchAgentRun with 7 separate InvocationProvenance constructions), extracting a makeProvenance helper function dramatically reduces code duplication. The code review council (divisor-architect) flagged this as a MEDIUM DRY violation. The fix involved creating a makeProvenance(agent, requestedModel, requestedVariant, readOnly) helper and an earlyProvenance closure that captures parsed input fields. This reduced ~70 lines of repeated object literals to ~15 lines. Additionally, interface properties in TypeScript should be marked readonly for consistency when the interface represents injected dependencies (ExecutorDependencies) — the architect also flagged this inconsistency since all other interfaces in agent-executor.ts already used readonly.
