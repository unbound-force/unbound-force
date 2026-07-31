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

## 1. Add pre-condition block to SKILL.md

- [x] 1.1 Insert a `**Pre-condition**` block in `.opencode/skills/openspec-apply-change/SKILL.md` immediately after the `**Steps**` heading (line 16) and before Step 1 (line 18). The block text MUST be: `**Pre-condition**: Before any step, verify: NEVER switch branches or suggest archiving with uncommitted changes. Run `git status --short` if branch state is uncertain.`

## 2. Verify

- [x] 2.1 Verify the pre-condition block appears before Step 1 in the file and the existing guardrail at line 212 remains unchanged
- [x] 2.2 Verify constitution alignment: confirm the change is N/A for all four org principles (no artifact formats, composability, observable quality, or testability impact)
<!-- spec-review: passed -->
<!-- code-review: passed -->
