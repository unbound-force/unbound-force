---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-11-02T21:14:06Z
identity: openspec-20261102T211406-jay-flowers
tier: draft
---

Concurrent MODIFIED-requirement hazard in OpenSpec: when two unarchived changes under openspec/changes/ both carry a `### Requirement: FR-NNN` delta for the same capability requirement, each restates the COMPLETE requirement — so each silently carries the other's target clause at its pre-modification value. Whichever archives second overwrites the first's modification wholesale (e.g. fix-uf-finale-ci-poll-gh-schema and review-dispatch-diff-manifest both modify FR-006 Pre-computed fields: one restates exceeds_threshold at origin semantics while the other preserves the always-true ci_pending rule the first repeals). Before writing a MODIFIED delta, scan sibling unarchived changes for the same requirement id and either rebase the restated text onto the sibling's modification or document archive ordering in the delta header — as fix-uf-finale-ci-poll-gh-schema already does for its finale-token-reduction sibling.
