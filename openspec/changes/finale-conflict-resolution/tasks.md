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

## 1. Add sub-agent option to step 6b

All tasks in this group modify the same file
(`.opencode/commands/finale.md`), so no parallel markers.

- [x] 1.1 Add option 5 text to the conflict recovery
  options menu in step 6b. Insert after option 4:
  `> 5. Spawn sub-agent to resolve conflicts
  (AI-assisted)`. Update the AskUserQuestion prompt
  to include the new option.
  **File**: `.opencode/commands/finale.md`

- [x] 1.2 Add the "Option 5" section after the existing
  "Option 4 -- Continue anyway" section in step 6b.
  Document the merge-first flow: fetch, merge (creating
  conflict markers), identify conflicting files via
  `git diff --name-only --diff-filter=U`, then spawn
  the sub-agent.
  **File**: `.opencode/commands/finale.md`

- [x] 1.3 Document the sub-agent prompt construction.
  Specify that the Task tool is called with
  `subagent_type: cobalt-crush-dev` and a prompt
  containing: the list of conflicting files, the target
  branch name, instructions to resolve conflict markers
  (`<<<<<<<`, `=======`, `>>>>>>>`), instructions to
  stage resolved files, and a directive to report
  per-file success/failure.
  **File**: `.opencode/commands/finale.md`

## 2. Add user approval gate and failure handling

Continues modifying the same file sequentially.

- [x] 2.1 Document the user approval gate after sub-agent
  completion. After the sub-agent returns, run
  `git diff --cached` and show the diff to the user.
  Present options: approve, request edits, or abort.
  On approve: complete merge commit and push. On abort:
  `git merge --abort` and return to options menu.
  **File**: `.opencode/commands/finale.md`

- [x] 2.2 Document failure handling. If the sub-agent
  reports unresolved files: report which were resolved
  and which remain, run `git merge --abort`, and return
  to the conflict recovery options menu. If the sub-agent
  fails completely: report the failure, abort, and return
  to options.
  **File**: `.opencode/commands/finale.md`

- [x] 2.3 Document post-resolution CI check polling.
  After successful resolution, push, and merge commit,
  use the same bash polling loop as options 1-2 to wait
  for CI checks.
  **File**: `.opencode/commands/finale.md`

## 3. Verification

- [x] 3.1 Verify that options 1-4 text and numbering are
  unchanged. Read the modified finale.md and confirm the
  existing options are preserved verbatim.
  **File**: `.opencode/commands/finale.md`

- [x] 3.2 Verify constitution alignment. Confirm the
  change maintains Autonomous Collaboration (sub-agent
  communicates through artifacts, not runtime coupling)
  and Observable Quality (resolution diff is shown to
  user before commit). Ref: proposal.md constitution
  assessment.
<!-- spec-review: passed -->
<!-- code-review: passed -->
