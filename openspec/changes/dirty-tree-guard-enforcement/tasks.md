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

- [x] 1.1 [P] Update `.opencode/skills/openspec-propose/SKILL.md`
  (the dirty-tree guard section under step 3, starting at
  `a. **Dirty working tree check**`): Replace the prose-only
  dirty-tree confirmation with an explicit AskUserQuestion tool
  call. The new text MUST:
  - First display `git status --short` output to the user
    (before prompting, so the user can make an informed decision)
  - Then call AskUserQuestion with options: "Stash changes and
    continue", "Abort -- keep changes as-is"
  - On "Stash changes and continue": run `git stash push -m
    "openspec-propose: auto-stash before branch switch"`. If
    stash succeeds, print "Changes stashed. Run `git stash pop`
    to restore them when ready." and proceed to branch creation.
    If stash fails (non-zero exit), stop immediately, display
    the error, and do not create the branch
  - On "Abort": stop immediately, do not create the branch
  - Preserve the existing exception clause about explicit
    change requests still requiring confirmation

- [x] 1.2 [P] Update `.opencode/commands/opsx-propose.md`
  (the dirty-tree guard section under step 3, starting at
  `a. **Dirty working tree check**`): Apply the identical
  replacement text as task 1.1. The guard logic MUST be
  identical in both files to prevent drift.

## 2. Verification

- [x] 2.1 Verify both files have identical dirty-tree guard
  text by running `diff` on the extracted guard sections from
  both files and confirming zero differences. Specifically
  confirm:
  - `git status --short` display instruction appears before
    the AskUserQuestion call
  - AskUserQuestion tool call with correct option labels
  - Stash command with correct message string
  - Stash recovery message: "Changes stashed. Run
    `git stash pop` to restore them when ready."
  - Stash failure handling (stop on non-zero exit)
  - Abort behavior (stop immediately)
  - Exception clause preserved (explicit change requests
    still require confirmation)

- [x] 2.2 Verify constitution alignment: confirm the change
  supports Principle V (Security by Default) by ensuring
  the guard is now a tool-enforced gate rather than
  prose-only instruction. No other principles are affected
  (N/A for I-IV per proposal).

## 3. Documentation

- [x] 3.1 Add a CHANGELOG entry under the Unreleased section
  noting the dirty-tree guard hardening in the
  openspec-propose skill and command. Reference issues #350
  and #353.

<!-- spec-review: passed -->
<!-- code-review: passed -->
