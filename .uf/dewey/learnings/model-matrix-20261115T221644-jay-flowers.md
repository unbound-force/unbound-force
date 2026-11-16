---
tag: model-matrix
author: jay-flowers
category: gotcha
created_at: 2026-11-15T22:16:44Z
identity: model-matrix-20261115T221644-jay-flowers
tier: draft
as_of: "unbound-force@6a553043168ccaffc4677672b76b30f5453aa1b5"
evidence:
  - "openspec/changes/review-dispatch-errors-cap-loses-unrecov/tasks.md"
---

On the @coffee workstation (2026-11-15), judgment-classified uf_delegate_command calls that rely on model-matrix classifier-cascade resolution fail with DELEGATE_FAILED ("Cannot connect to API... models.corp.redhat.com" — unreachable endpoint in the cascade). Two reliable workarounds: pass an explicit model that is known reachable (model: "enmaas/rits/zai-org/glm-5-3" worked all session) or a convention pack that resolves a profile (pack: "go" for Go work). Deterministic delegations with a skill/skill_absent_reason do not hit the cascade and succeed without either. Also: the task tool cannot spawn workers from this depth ("Subagent depth limit reached (1)"), so Swarm worktree parallelism is unavailable — sequential uf_delegate_command is the fallback for [P]-marked tasks, and any half-created worktree/comms reservation must be cleaned up and released before falling back.
