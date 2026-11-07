---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-11-07T19:14:28Z
identity: openspec-20261107T191428-jay-flowers
tier: draft
---

When templating superseded-by annotations across sibling OpenSpec deltas, carry the archive-order sentence ("That change MUST archive after this change") only into deltas whose requirement text the superseding change actually restates, and mirror every ordering assertion into the change's central constraint lists (proposal What Changes, design decisions, delta header comments). A boilerplate ordering claim pasted onto an illustrative-example annotation creates a constraint recorded in exactly one place — the opposite of the redundant-recording standard the superseding change sets for itself, and a consistency gap future archivers cannot reconcile from the central lists. Discovered reviewing retroactive-push-skill-contracts (PR 948): the spawn-subtask-classification-params annotation asserted an ordering constraint absent from all four central surfaces.
