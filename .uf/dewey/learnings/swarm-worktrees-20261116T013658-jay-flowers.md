---
tag: swarm-worktrees
author: jay-flowers
category: gotcha
created_at: 2026-11-16T01:36:58Z
identity: swarm-worktrees-20261116T013658-jay-flowers
tier: draft
---

Swarm worktree parallelism in /uf.unleash phase 3 forks workers from HEAD, and uncommitted working-tree changes are invisible to them. Because implementation work stays uncommitted until /uf.finale (the pipeline's commit step), [P] tasks that assert against just-written production code — test-migration tasks are the canonical case — cannot run in worktrees: they would write tests against the stale HEAD surface and fail. Check git status for uncommitted modifications to the tasks' dependency files before creating worktrees; when the dependencies are dirty, fall back to sequential uf_delegate_command execution for the [P] tasks. This cost a sequential phase 3 in the compaction-storm-detector-sums-compactio run (branch opsx/compaction-storm-detector-sums-compactio, 2026-11-15/16) and is the correct trade — a few extra minutes beats structurally broken parallel workers.
