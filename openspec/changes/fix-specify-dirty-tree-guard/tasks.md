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

## 1. Update dirty-tree guard in speckit.specify.md

- [x] 1.1 Replace prose-only dirty-tree guard (lines 42-50) in
  `.opencode/commands/speckit.specify.md` with an explicit
  `AskUserQuestion` tool-call instruction. The replacement text
  MUST:
  - Instruct the agent to run `git status --short`
  - If uncommitted changes exist, invoke `AskUserQuestion` with
    options: "Stash changes and continue" / "Abort -- keep
    changes as-is"
  - Display the list of uncommitted files as context
  - If user selects stash: run `git stash` then proceed
  - If `git stash` fails (non-zero exit): stop execution and
    report the failure (do NOT proceed to branch creation)
  - If user selects abort: stop execution and report
  - Preserve the exception clause but clarify that confirmation
    is still required even for explicit commands
  - File: `.opencode/commands/speckit.specify.md`

## 2. Verification

- [x] 2.1 Review the updated `speckit.specify.md` to confirm:
  - The `AskUserQuestion` tool call is explicitly specified
  - Both options ("Stash changes and continue", "Abort -- keep
    changes as-is") are present
  - The uncommitted files list display instruction is included
  - The stash and abort handling instructions are clear,
    including `git stash` failure handling
  - The exception clause still requires confirmation
  - No other sections of the file were inadvertently modified
  - File: `.opencode/commands/speckit.specify.md`

- [x] 2.2 Verify constitution alignment: confirm the change is
  N/A for principles I-IV and PASS for principle V (Security by
  Default) as documented in the proposal.

<!-- spec-review: passed -->
