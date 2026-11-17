---
tag: uf-unleash
author: jay-flowers
category: gotcha
created_at: 2026-11-15T04:32:32Z
identity: uf-unleash-20261115T043232-jay-flowers
tier: draft
---

Swarm worktree parallelism for [P] implement tasks is unsound when the phase's implementation exists only as uncommitted working-tree changes: replicator_forge_worktree_create checks out HEAD, so parallel children in the worktree cannot see the uncommitted code they are supposed to test or extend — task 2.2's dispatch tests would have run against a tree without the phase-1 implementation. The correct fallback in /uf.unleash is sequential uf_delegate_command delegations in the shared working tree (children share the working directory, so uncommitted state is visible), announced honestly as a routing decision rather than claiming worktrees were unavailable. Worktree parallelism only becomes sound after the base implementation is committed.
