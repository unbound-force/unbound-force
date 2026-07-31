<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Update `.opencode/commands/opsx-propose.md`

- [x] 1.1 Insert a bolded preamble block immediately after the `**Steps**` line and before Step 1. The preamble states: this command creates spec artifacts only, MUST NOT implement code / commit / push / create PRs / invoke implementation commands, and MUST STOP and prompt the user after artifacts are complete. Retain the existing STOP HERE block after Step 6 as reinforcement.
- [x] 1.2 Replace the prose-only dirty-tree guard in Step 3a with an explicit `AskUserQuestion` specification. The guard MUST present exactly two options: "Stash changes and continue" (agent runs `git stash --include-untracked` then proceeds; if stash fails, stop and report) and "Abort — keep changes as-is" (agent stops and reports). Keep the prose description of what dirty-tree detection means, but augment it with the explicit tool call. Retain the exception note about explicit user requests still requiring confirmation.

## 2. Update `.opencode/skills/openspec-propose/SKILL.md`

- [x] 2.1 Insert the same bolded preamble block immediately after the `**Steps**` line and before Step 1. Match the preamble content from task 1.1. Retain the existing STOP HERE block after Step 6 as reinforcement.
- [x] 2.2 Apply the same AskUserQuestion replacement to the Step 3a dirty-tree guard. Match the structure and options from task 1.2. Preserve the existing `/opsx:propose` invocation syntax in the exception example (vs `/opsx-propose` in the command file) — these reflect how each file is invoked in its respective context.

## 3. Verify parity and constitution alignment

- [x] 3.1 Verify both files have identical guard logic and preamble content (accounting for frontmatter differences and the known `/opsx-propose` vs `/opsx:propose` syntax divergence in the exception example). Run `diff` on the guard and preamble sections. Confirm the preamble appears before Step 1 in both files, the AskUserQuestion specification is identical in both, and the existing STOP HERE + Guardrails sections are retained. Confirm no scaffold assets exist for these files (they are managed by `openspec init`, not `uf init`).
- [x] 3.2 Verify constitution alignment: the change aligns with Observable Quality (deterministic agent behavior via structured tool calls) and Security by Default (explicit user consent before branch operations). No other principles are impacted.
<!-- spec-review: passed -->
<!-- code-review: passed -->
