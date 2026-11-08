<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file --
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Add Dewey staging instruction to uf.finale

- [x] 1.1 Add a Dewey artifact staging sub-step to
  Step 2 (Check for Changes to Commit) in
  `internal/scaffold/assets/opencode/commands/uf.finale.md`.
  Insert after the "Stage all changes: `git add .`" line
  and before the checkpoint. The sub-step MUST:
  - Run `git add .uf/dewey/learnings/ .uf/dewey/compiled/ 2>/dev/null`
  - Include brief rationale (knowledge artifacts belong in the feature PR)
  - Include a "Do NOT skip this step" directive
  - Be idempotent (no-op when directories do not exist)

## 2. Verification

- [x] 2.1 Verify the scaffold drift-detection test still
  passes (`make test` or `go test ./internal/scaffold/...`).
  The file count should not change since no files are
  added or removed.
- [x] 2.2 Verify constitution alignment: the change
  touches only the command template (Autonomous
  Collaboration: PASS, Observable Quality: PASS,
  Composability: N/A, Testability: N/A).
