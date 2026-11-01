---
tag: openspec-workflow
author: jay-flowers
category: gotcha
created_at: 2026-10-31T21:32:21Z
identity: openspec-workflow-20261031T213221-jay-flowers
tier: draft
---

OpenSpec tasks.md checkbox state is consumed by /uf.unleash resume detection — unchecked boxes drive re-implementation planning. When implementation front-loads a later task's enumerated deliverable (e.g., comprehensive unit tests landing as per-task verification under an earlier task's label), the later task's enumeration and checkbox must be reconciled immediately: a stale unchecked box whose enumerated list is already satisfied misleads a resumed implementer into redoing work, while a stale enumeration hides what actually remains.
