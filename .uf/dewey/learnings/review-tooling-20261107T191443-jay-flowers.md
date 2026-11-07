---
tag: review-tooling
author: jay-flowers
category: pattern
created_at: 2026-11-07T19:14:43Z
identity: review-tooling-20261107T191443-jay-flowers
tier: draft
---

When scoping a retroactive spec-record change from enumerated review findings, re-scan the full final changeset for other instances of the same contradiction class before finalizing — the findings enumeration can lag commits added later in the same PR. PR 948's retroactive-push-skill-contracts change was scoped from the round-2 findings' five contradicted deltas and remediated all five correctly, but the OSV dependency bump riding the same branch introduced a sixth contradiction (the SC-FR-003 zod pin) that no round-2 finding had enumerated, so it escaped the remedy. The check is mechanical: for every value this changeset pins, configures, or deletes, grep the unarchived openspec/changes/ stack for that value and annotate or amend every hit. Symptom pattern: a remediation change that is internally complete against its findings list yet incomplete against its own diff.
