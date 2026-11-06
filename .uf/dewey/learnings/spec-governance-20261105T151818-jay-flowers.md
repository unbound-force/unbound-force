---
tag: spec-governance
author: jay-flowers
category: gotcha
created_at: 2026-11-05T15:18:18Z
identity: spec-governance-20261105T151818-jay-flowers
tier: draft
---

OpenSpec cross-change contradiction remedy mechanics (PR 957): when a new change falsifies requirement text that lives in an unarchived sibling change (not yet merged into openspec/specs/), the reconciliation belongs in the sibling change's unarchived delta text (edit it before it archives) or a coordination note recorded in both changes — a MODIFIED delta in the new change cannot validate, because OpenSpec MODIFIED requirements must match canonical spec text that does not exist yet. Detection half (already captured in learning spec-governance-20261104T224935-jay-flowers): grep the openspec/changes tree for the old mechanism's text before rewriting a shared Make target, command, or config mechanism. Remedy half: route the fix through the unarchived delta that owns the contradicted text, and record the cross-change dependency in the new change's design.md so archive ordering cannot silently resurrect the contradiction.
