---
tag: determinism
author: jay-flowers
category: decision
created_at: 2026-11-15T00:34:11Z
identity: determinism-20261115T003411-jay-flowers
tier: draft
---

Spec-review DE-008 calibration (from review-dispatch-a-reviewer-with-empty-ou, iteration 2): OpenSpec tasks.md artifacts in this repo do not pre-state per-task task_classification — including delegation-skill-enforcement, the change that introduced DE-008 — because the contract's carrier is the delegation call itself: uf.unleash's implement template classifies each task before delegation and passes task_classification plus skill or skill_absent_reason on every uf_delegate_command / replicator_forge_spawn_subtask call, enforced fail-closed and recorded as the audit source. The determinism spec-mode delegation check therefore fires only when a change adds or edits delegating instruction text or explicitly plans delegation calls, not for a standard implement-phase task list. Adjunct practice: for deterministic tasks with no repo tool (e.g. byte-identical scaffold-mirror copies — tool-inventory --match "mirror scaffold assets" returns empty), the spec-review owner lookup produces exactly the substantive skill_absent_reason evidence (what was searched, what was found) the delegation contract will demand; hand it to the author as an advisory.
