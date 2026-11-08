---
tag: openspec-governance
author: jay-flowers
category: gotcha
created_at: 2026-11-07T19:55:30Z
identity: openspec-governance-20261107T195530-jay-flowers
tier: draft
---

A supersession record that marks contradicted spans across sibling OpenSpec changes must census every file of each contradicted change — including secondary spec deltas — before claiming complete coverage. In PR 948's retroactive-push-skill-contracts, the record annotated delegation-skill-enforcement in three of its four files carrying the legacy `.opencode/skill/` root (unleash spec, design.md, tasks.md) and annotated the identical illustrative absence-reason example in spawn-subtask-classification-params' determinism-policy delta, but missed delegation-skill-enforcement/specs/determinism-policy/spec.md, which carries the same example and materializes into the live spec at archive — leaving design.md's "the inline annotations mark every contradicted span" mitigation claim factually incomplete. Verification pattern: grep the retired token across openspec/changes/*/specs/ (every spec delta of live changes, not just each change's primary capability folder) and diff the match set against the annotation set before writing a completeness claim; a change touching multiple capabilities (unleash + determinism-policy) has multiple spec deltas that all materialize at archive.
