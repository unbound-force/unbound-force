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

## 1. Protect `/uf.init` instruction body

All tasks in this group modify `.opencode/commands/uf.init.md`.

- [x] 1.1 Add an idempotency check for the `<protect>` block: before
  inserting, verify that no `<protect>` tag exists between the
  `## Instructions` heading and the `### Step 0` heading. If present,
  report `⊘ uf.init.md: protect block already present (skipped)` and
  skip insertion.
- [x] 1.2 Insert a `<protect>` opening tag immediately after the
  `## Instructions` heading (line 24 of the current file).
- [x] 1.3 Insert a `</protect>` closing tag after the Step 12 content
  (after the `### Legacy Directory Cleanup` section and its
  `#### Sub-task B` content) and before the `### Post-Write Verification`
  heading.
- [x] 1.4 Verify the `<protect>` block is correctly placed by re-reading
  the file and confirming the opening tag is after `## Instructions`
  and the closing tag is before `### Post-Write Verification`.

## 2. Add `<protect>` blocks to Step 5 creation templates

Update the Step 5 insertion templates so that files created by Step 5
(`speckit.analyze.md`, `speckit.checklist.md`, `speckit.clarify.md`,
`speckit.taskstoissues.md`) include a `<protect>` block wrapping their
core instruction body.

- [x] 2.1 Update the `speckit.analyze.md` template in Step 5 to include
  a `<protect>` block after the title heading and before the
  `## Guardrails` section.
- [x] 2.2 Update the `speckit.checklist.md` template in Step 5 to include
  a `<protect>` block after the title heading and before the
  `## Guardrails` section.
- [x] 2.3 Update the `speckit.clarify.md` template in Step 5 to include
  a `<protect>` block after the title heading and before the
  `## Guardrails` section.
- [x] 2.4 Update the `speckit.taskstoissues.md` template in Step 5 to
  include a `<protect>` block after the title heading and before the
  `## Guardrails` section.

## 3. Update insertion logic for protected target files

Update Steps 2-4, 6, 8, and 10 so that when inserting content into a
target file that already contains a `<protect>` block, the inserted
content is placed inside the existing `<protect>` block (between the
`<protect>` and `</protect>` tags). When the target file does NOT
contain a `<protect>` block, the insertion behavior is unchanged.

- [x] 3.1 Update Step 2 (Branch Enforcement) insertion instructions to
  check for an existing `<protect>` block in the target file and place
  inserted content inside it when present.
- [x] 3.2 Update Step 3 (Dewey Context) insertion instructions to check
  for an existing `<protect>` block in the target file and place
  inserted content inside it when present.
- [x] 3.3 Update Step 4 (3-Tier Dewey Degradation) insertion instructions
  to check for an existing `<protect>` block in the target file and
  place inserted content inside it when present.
- [x] 3.4 Update Step 6 (Speckit Command Guardrails) insertion instructions
  to check for an existing `<protect>` block in the target file and
  place inserted content inside it when present.
- [x] 3.5 Update Step 8 (OpenSpec Command Guardrails) insertion instructions
  to check for an existing `<protect>` block in the target file and
  place inserted content inside it when present.
- [x] 3.6 Update Step 10 (STOP HERE Blocks) insertion instructions to
  check for an existing `<protect>` block in the target file and place
  inserted content inside it when present.

## 4. Verification

- [x] 4.1 Run `/uf.init` on a test target repository and verify that the
  `<protect>` block is correctly added to `.opencode/commands/uf.init.md`.
- [x] 4.2 Run `/uf.init` a second time on the same target repository and
  verify idempotency: no duplicate `<protect>` blocks, and the command
  reports `⊘ uf.init.md: protect block already present (skipped)`.
- [x] 4.3 Verify that speckit command files created by Step 5 contain
  `<protect>` blocks.
- [x] 4.4 Verify constitution alignment: confirm that the change passes
  all four principles (I. Autonomous Collaboration, II. Composability
  First, III. Observable Quality, IV. Testability) as assessed in the
  proposal.
<!-- spec-review: passed -->
<!-- code-review: passed -->
<!-- scaffolded by uf vdev -->
