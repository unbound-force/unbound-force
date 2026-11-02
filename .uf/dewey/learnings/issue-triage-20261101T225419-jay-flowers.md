---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:54:19Z
identity: issue-triage-20261101T225419-jay-flowers
tier: draft
---

Determinism-debt follow-up issues in unbound-force are self-verifying at triage time: when a change defers DE-001 mechanical-prose extraction as a Non-Goal, the proposal's Principle VI section names each deferred step and records the follow-up issue number ("tracked in #842" in openspec/changes/update-run-pipeline-plan/proposal.md), the consumer command file (.opencode/commands/uf.run-pipeline-plan.md) still carries those steps as prose, and the named remediation target exists with shipped precedent (uf-workflow createXxxTool factories with injected dependencies, success/failure JSON envelopes, verified via make plugin-test). Such issues triage as VALID/enhancement/objective with no split — the extractions share one command file, its byte-identical scaffold mirror under internal/scaffold/assets/, one plugin, and one test suite, so a single change carries them all.
