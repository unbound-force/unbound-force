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

## 1. Add pre-condition block

- [x] 1.1 In `.opencode/skills/openspec-apply-change/SKILL.md`, insert a Pre-condition block immediately after the `**Steps**` heading (line 16) and before Step 1 (line 18). The block text MUST be:

```
**Pre-condition**: Before any step, verify: NEVER switch
branches or suggest archiving with uncommitted changes.
Run `git status --short` if branch state is uncertain.
```

## 2. Verify

- [x] 2.1 Verify the pre-condition block appears after the `**Steps**` heading and before Step 1 (`1. **Select the change**`)
- [x] 2.2 Verify the existing guardrail bullet at the end of the file (line 212) is preserved unchanged
- [x] 2.3 Verify no other content in the file was modified
<!-- spec-review: passed -->
