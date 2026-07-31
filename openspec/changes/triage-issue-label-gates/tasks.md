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

  All tasks in groups 1-5 modify the same file:
  .opencode/commands/triage-issue.md
  Therefore no tasks are parallel-eligible.

  Line numbers reference the file as of the current
  main branch. Earlier tasks will shift line numbers --
  use content anchors (section headers, code blocks)
  rather than line numbers when locating insertion points.
-->

## 1. Update label application documentation

- [x] 1.1 Replace line 243 text "Labels are applied
  **automatically without user confirmation**, with one
  exception: the `duplicate` label requires user
  confirmation because it carries implicit 'close'
  semantics." with "All labels require user confirmation
  before application. The `duplicate` label includes an
  additional warning about implicit close semantics."
  File: `.opencode/commands/triage-issue.md`

## 2. Add label creation confirmation gate

- [x] 2.1 Insert an `AskUserQuestion` gate before the
  `gh label create` command (line 262). After the label
  existence check (line 259) and before `gh label create`,
  add: "Use the **AskUserQuestion tool** with options
  `['Yes -- create label', 'No -- skip']`. Include the
  label name and description in the prompt. If the user
  selects 'No -- skip', skip both label creation and
  label application, and continue to Phase 4.3."
  File: `.opencode/commands/triage-issue.md`

## 3. Add label application confirmation gate

- [x] 3.1 Insert an `AskUserQuestion` gate before the
  `gh issue edit --add-label` command (line 270). After
  the label existence/creation step and before label
  application, add: "Use the **AskUserQuestion tool**
  with options `['Yes -- apply label', 'No -- skip']`.
  Include the label name in the prompt. If the user
  selects 'No -- skip', skip label application and
  continue to Phase 4.3."
  File: `.opencode/commands/triage-issue.md`

## 4. Update duplicate label gate

- [x] 4.1 Merge the existing duplicate-specific gate
  (lines 273-276) into the general label application
  gate. For the `duplicate` category, the application
  gate prompt MUST use options
  `["Yes -- apply duplicate label", "No -- skip"]` and
  MUST include the close-semantics warning: "the
  `duplicate` label signals the issue should be closed."
  Remove the separate duplicate-only gate block to avoid
  double-prompting.
  File: `.opencode/commands/triage-issue.md`

## 5. Update display template

- [x] 5.1 Update the "Proposed Actions" display in
  Phase 4.1 (line 235). Change
  `Label: <label> (auto-apply / requires confirmation)`
  to `Label: <label> (requires confirmation)` since all
  labels now require confirmation.
  File: `.opencode/commands/triage-issue.md`

## 6. Verification

- [x] 6.1 Verify that the updated Phase 4.2 contains
  `AskUserQuestion` gates before every `gh label create`
  and every `gh issue edit --add-label` invocation. No
  label mutation path should bypass user confirmation.

- [x] 6.2 Verify that the `duplicate` label retains its
  close-semantics warning in the prompt text.

- [x] 6.3 Verify that the re-run check (the block
  containing "label already present" skip) is preserved
  and fires before the new gates.

- [x] 6.4 Verify constitution alignment: the change
  strengthens Security by Default (Principle V) and
  maintains Gatekeeping Integrity (adds gates, does not
  weaken existing ones). No other principles are affected.

## 7. Sync scaffold asset

- [x] 7.1 Copy the modified `.opencode/commands/triage-issue.md`
  to `internal/scaffold/assets/opencode/commands/triage-issue.md`
  so the scaffold asset remains byte-identical.
  Run `go test -race -count=1 ./internal/scaffold/...`
  to verify drift detection passes.

<!-- spec-review: passed -->
<!-- code-review: passed -->
