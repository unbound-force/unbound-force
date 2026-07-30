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

## 1. Add confirmation gates to SKILL.md

All tasks modify the same file
(`.opencode/skills/openspec-archive-change/SKILL.md`),
so no [P] markers -- these run sequentially.

- [ ] 1.1 Add `AskUserQuestion` gate between step 5 and
  step 6. Insert after line 89 (end of the CRITICAL prose
  warning) and before step 6 heading. The gate presents
  two options: "Changes committed and pushed -- proceed to
  archive" and "Abort -- need to commit first". If the user
  selects abort, stop the workflow and inform the user to
  commit before retrying.

- [ ] 1.2 Add `AskUserQuestion` gate before `git checkout
  main` in step 7. Insert before line 118 (`Then switch
  branches:`). The gate presents two options: "Return to
  main" and "Stay on branch". If the user selects "Stay on
  branch", skip `git checkout main` and note in the step 8
  summary that the user remained on the `opsx/<name>`
  branch.

- [ ] 1.3 Remove the prose-only CRITICAL warning (lines
  85-89) since it is now superseded by the structural
  `AskUserQuestion` gate. Replace with a brief note
  referencing the gate.

## 2. Verification

- [ ] 2.1 Verify the modified SKILL.md has correct
  markdown structure: all steps numbered correctly,
  `AskUserQuestion` blocks use the exact option text from
  the spec, abort/stay behaviors are clearly documented.

- [ ] 2.2 Verify constitution alignment: confirm the
  change does not affect Autonomous Collaboration,
  Composability, Observable Quality, or Testability
  principles. Confirm Security by Default alignment
  (structural gates replacing prose-only warnings).
