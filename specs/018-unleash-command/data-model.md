# Data Model: Unleash Command

## Pipeline Steps

The `/unleash` pipeline has 8 sequential steps. Each
step has a detection condition for resumability.

| Step | Name | Detection (Done When) | Resume Target |
|------|------|-----------------------|---------------|
| 1 | Clarify | No `[NEEDS CLARIFICATION]` in spec.md AND (Clarifications section exists OR plan.md exists) | Plan |
| 2 | Plan | plan.md exists in feature dir | Tasks |
| 3 | Tasks | tasks.md exists in feature dir | Spec Review |
| 4 | Spec Review | `<!-- spec-review: passed -->` marker exists in tasks.md | Implement |
| 5 | Implement | All tasks `[x]` in tasks.md | Code Review |
| 6 | Code Review | CI build+test commands pass (derived from `.github/workflows/`) | Retrospective |
| 7 | Retrospective | Always runs (idempotent) | Demo |
| 8 | Demo | Terminal state | N/A |

## Exit Points

| Exit | Trigger | Severity | Action |
|------|---------|----------|--------|
| Clarify Exit | Dewey can't answer 1+ questions | Expected | Present questions, suggest re-run |
| Spec Review Exit | HIGH/CRITICAL findings | Blocking | Present findings, suggest /speckit.clarify |
| Worker Failure | Parallel worker crashes | Error | Stop all workers, present error |
| Merge Conflict | Auto-resolution fails | Error | Present conflict details |
| Build Checkpoint | Build or test fails | Error | Present failure, exit |
| Code Review Exit | 3 iterations exhausted | Blocking | Present persistent findings |

## Swarm Worker Lifecycle

For each `[P]`-marked task in a phase:

```
1. swarm_worktree_create(project, task_id, start_commit)
   → worktree path

2. swarm_spawn_subtask(bead_id, epic_id, title, files)
   → worker executes in worktree

3. Worker completes:
   - Commits changes in worktree
   - Marks task [x] in tasks.md
   - Reports via swarm_complete()

4. swarm_worktree_merge(project, task_id)
   → cherry-picks commits to main branch
   → auto-resolves conflicts if possible

5. swarm_worktree_cleanup(project, task_id)
   → removes worktree
```

## Parallel Execution Model

Within each phase:

```text
Phase N:
  ├── Sequential tasks (no [P]) → run first, in order
  │   T001 → T002 → T003
  │
  └── Parallel tasks ([P]) → run after sequential
      T004 [P] ──┐
      T005 [P] ──┤── concurrent workers
      T006 [P] ──┤   (each in own worktree)
      T007 [P] ──┘
                  │
          merge all worktrees
                  │
          phase checkpoint (build + test)
                  │
          next phase
```

## Graceful Degradation

| Tool | Available | Degraded |
|------|-----------|----------|
| Dewey | Semantic search for clarify answers | All questions go to human |
| Gaze | Quality analysis in code review | Code review skips Phase 1b |
| Swarm worktrees | Parallel `[P]` execution | Sequential execution |
| Swarm workers | Parallel task spawning | Single Cobalt-Crush agent |
| Hivemind | Retrospective learning storage | Retrospective skipped |
| SwarmMail | File reservation for workers | Workers proceed without locks |
