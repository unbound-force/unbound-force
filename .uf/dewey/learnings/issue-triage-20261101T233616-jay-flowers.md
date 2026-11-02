---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:36:16Z
identity: issue-triage-20261101T233616-jay-flowers
tier: draft
---

Documentation-gate issues filed for contract-carrier migrations must audit related open documentation issues for stale carrier descriptions. In unbound-force issue #834 (docs gate for the spawn-subtask-classification-params change), the change moved the Swarm-path delegation contract carrier from subtask-description prose to fail-closed spawn-tool parameters on replicator_forge_spawn_subtask (task_classification/skill/skill_absent_reason, SKILL_REFERENCE_REQUIRED/SKILL_NOT_FOUND), while the related website issue unbound-force/website#317 still described the Swarm path in description-carrier terms. The gate issue correctly flagged #317's stale bullet with a note to document the spawn tool's parameter parity (replicator#143) when #317 is worked, rather than filing a duplicate website issue. Triage pattern: when a change migrates a contract's carrier (prose → tool parameters), check related open issues for descriptions of the old carrier and record the needed correction inside the gate issue to prevent documentation drift without duplication.
