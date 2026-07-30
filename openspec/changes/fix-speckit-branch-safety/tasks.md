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

## 1. Relocate Branch Safety to Pre-conditions

- [ ] 1.1 In `.opencode/skills/speckit-workflow/SKILL.md`,
  insert a new "## Pre-conditions" section immediately before
  "## Reading tasks.md" (line 30). Populate it with the
  branch safety content currently at lines 116-128:
  - **CRITICAL**: All work MUST be committed and pushed on
    the current feature branch before any branch switch
  - After completing all phases, commit and push before
    suggesting PR creation, merging, or switching to main
  - Before creating a new feature branch, check
    `git status --short` for uncommitted changes; if any
    exist, STOP and ask the user
  - Never silently switch branches with a dirty working tree

- [ ] 1.2 Remove the original "## Branch Safety" section
  (lines 114-129) from the same file to eliminate duplication.

## 2. Verification

- [ ] 2.1 Verify the "Pre-conditions" section appears before
  "Reading tasks.md" in the final file.
- [ ] 2.2 Verify no duplicate branch safety content exists
  (the old "Branch Safety" section is fully removed).
- [ ] 2.3 Verify all other sections remain unchanged in
  content and ordering.
