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

## 1. Replace dirty-tree guard in both files

- [ ] 1.1 [P] Update `.opencode/skills/openspec-propose/SKILL.md`
  lines 51-64: replace prose-only dirty-tree guard with
  explicit AskUserQuestion tool call. The new text MUST:
  (a) show `git status --short` output to the user,
  (b) call AskUserQuestion with options
  "Stash changes and continue" and
  "Abort -- keep changes as-is",
  (c) run `git stash` if user selects stash,
  (d) stop execution if user selects abort.
  Keep the surrounding guard structure (step 3a header,
  step 3b branch check) unchanged.

- [ ] 1.2 [P] Update `.opencode/commands/opsx-propose.md`
  lines 44-57: apply the identical replacement text as
  task 1.1. The dirty-tree guard section in both files
  MUST be identical per design decision D2.

## 2. Verification

- [ ] 2.1 Verify both files have identical dirty-tree
  guard text by diffing the relevant sections. Confirm
  the AskUserQuestion tool call, option labels, stash
  behavior, and abort behavior are all present.

- [ ] 2.2 Verify constitution alignment: confirm the
  change does not introduce runtime coupling between
  heroes (Principle I), does not add mandatory
  dependencies (Principle II), and maintains observable
  gate behavior via explicit tool calls (Principle III).
