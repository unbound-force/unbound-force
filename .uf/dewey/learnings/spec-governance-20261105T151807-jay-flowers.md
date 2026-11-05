---
tag: spec-governance
author: jay-flowers
category: pattern
created_at: 2026-11-05T15:18:07Z
identity: spec-governance-20261105T151807-jay-flowers
tier: draft
---

A gatekeeping-rule exception for changing a CI gate's install or enforcement discipline is only satisfied when the documented rationale amends the capability that owns the gate's requirement text — not merely the capability that owns the changed command. PR 957 (stop-shipping-plugin-lock) documented its lockless-install rationale (design D2, Risks) in the scaffold capability while the ci capability's unarchived requirement — which names `plugin-test` as the coverage-gate enforcement path, specifies "locked `npm ci`" as its composition, and carries the anti-weakening clause — kept asserting the old discipline. On archive, the mainline spec would brand the shipped Makefile a MUST violation even though a valid rationale existed, filed under the wrong capability. Extends the prior spec-governance learning (grep unarchived deltas for the old mechanism's text): detection is step one; step two is filing the reconciliation in the owning capability (a delta or coordination note in the change that carries the requirement), in the same PR, while both changes are still unarchived and mutable. Check: when a change touches a mechanism named in another capability's requirement text (enforcement paths, pinned commands, gate compositions), identify the owning change and amend or coordinate with it before merge — fixing it after archive requires a new governance change.
