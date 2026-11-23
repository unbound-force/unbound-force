---
tag: uf-unleash
author: jay-flowers
category: pattern
created_at: 2026-11-16T01:49:33Z
identity: uf-unleash-20261116T014933-jay-flowers
tier: draft
---

When /uf.unleash reaches [P] (parallel-eligible) tasks while all feature work is still uncommitted, Swarm worktree isolation cannot be used: replicator_forge_worktree_create checks out HEAD, which predates the working-tree changes, so workers would code against a stale base missing the tool under test, and cherry-pick merges back would semantically conflict with the uncommitted modifications to the very files the workers extend. The safe pattern is the protocol's sequential fallback through uf_delegate_command on the main tree, announced with the true reason (uncommitted base makes worktree isolation unsafe) rather than the not-installed message, executing [P] tasks in listed order after the non-[P] tasks per the tasks.md convention.
