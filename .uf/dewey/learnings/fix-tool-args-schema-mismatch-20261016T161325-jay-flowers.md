---
tag: fix-tool-args-schema-mismatch
author: jay-flowers
category: pattern
created_at: 2026-10-16T16:13:25Z
identity: fix-tool-args-schema-mismatch-20261016T161325-jay-flowers
tier: draft
---

The forge worktree parallel flow for [P]-marked tasks worked reliably across Phases 1-2 of fix-tool-args-schema-mismatch (four workers, zero conflicts): create an epic with subtasks via org_create_epic to get cell IDs, reserve target file paths via comms_reserve, create per-task worktrees with forge_worktree_create at the current commit, then dispatch Task-tool workers (subagent_type worker) with fully self-contained prompts — the generic forge_spawn_subtask shells are too vague, so embed the worktree absolute path, the exact task spec, and strict rules (work only in the worktree, never write the main directory, no npm/vitest because worktrees lack node_modules, no comms/forge MCP calls, commit with a prescribed conventional message). Workers commit on swarm/<cell-id> branches; merge back one at a time with forge_worktree_merge (cherry-picks onto the currently checked-out feature branch — verify with git branch --show-current), grep for conflict markers after each merge, clean up per worktree, release reservations, and mark tasks [x] in tasks.md immediately. Batch parallel workers at four or fewer and wait for each batch before spawning the next.
