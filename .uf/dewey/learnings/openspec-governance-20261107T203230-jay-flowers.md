---
tag: openspec-governance
author: jay-flowers
category: pattern
created_at: 2026-11-07T20:32:30Z
identity: openspec-governance-20261107T203230-jay-flowers
tier: draft
---

A retroactive supersession change's sweep enumeration is a scope claim reviewers can falsify against two references: the change's own Goal text, and every other contract-moving change in the same PR. Detection heuristic: extract the Goal's quantifier domain (e.g. "live delta text contradicting shipped behavior"), grep the unarchived openspec/changes/ stack for every value the same PR moved (dependency version strings, whole-file hashes, rollback baselines), and diff the hits against the annotated set; any hit outside the enumeration is an unannotated contradicting delta. Instance (PR 948, retroactive-push-skill-contracts): the change annotated five behavioral push/skill deltas while the same PR's zod 4.1.8→4.6.6 security bump left clarification-marker-superset SC-FR-003 MUST-pinning the old version and rollback-baseline lock SHA, unannotated — surfaced only as post-merge review feedback. Remedy matching in-repo precedent: a SUPERSEDED-BY blockquote on the contradicted requirement (review-council-multi-model-fanout pattern) referencing the bump commit and CHANGELOG entry, plus an enumeration update in design/proposal/tasks so the record's claims match its execution.
