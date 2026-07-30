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

  NOTE: All tasks modify the same file
  (.opencode/commands/triage-issue.md), so no tasks
  are parallel-eligible.
-->

## 1. Update Label Application Policy Text

- [ ] 1.1 In `.opencode/commands/triage-issue.md` section
  4.2 (line 243), replace the automatic-application policy
  text. Change "Labels are applied **automatically without
  user confirmation**, with one exception: the `duplicate`
  label requires user confirmation because it carries
  implicit 'close' semantics." to text stating that ALL
  label mutations require AskUserQuestion confirmation
  before execution.

## 2. Add Label Creation Confirmation Gate

- [ ] 2.1 In `.opencode/commands/triage-issue.md` section
  4.2, before the `gh label create` command block
  (line ~262), add an AskUserQuestion gate. The gate MUST
  display the label name and description being created.
  Options: `["Yes -- create label", "No -- skip"]`. If
  the user selects "No -- skip", skip label creation,
  record the skip in `actions_taken`, and skip label
  application (since the label does not exist).

## 3. Add Label Application Confirmation Gate

- [ ] 3.1 In `.opencode/commands/triage-issue.md` section
  4.2, before the `gh issue edit --add-label` command
  block (line ~270), add an AskUserQuestion gate. The
  gate MUST display the label name and issue number.
  Options: `["Yes -- apply label", "No -- skip"]`. If
  the user selects "No -- skip", skip label application,
  record the skip in `actions_taken`, and continue with
  comment posting.

## 4. Preserve Duplicate Label Context

- [ ] 4.1 In `.opencode/commands/triage-issue.md` section
  4.2, update the duplicate-specific text (lines 273-276)
  to serve as additional context displayed alongside the
  general label application gate rather than being the
  sole confirmation gate. The duplicate-specific messaging
  about close semantics MUST be preserved but presented
  as supplementary information, not a separate gate.

## 5. Verification

- [ ] 5.1 Read the modified `triage-issue.md` and verify
  that every `gh label create` call is preceded by an
  AskUserQuestion gate.
- [ ] 5.2 Read the modified `triage-issue.md` and verify
  that every `gh issue edit --add-label` call is preceded
  by an AskUserQuestion gate.
- [ ] 5.3 Verify the re-run check (line ~257) still
  bypasses both gates when the label is already applied.
- [ ] 5.4 Verify the duplicate-specific close-semantics
  messaging is preserved.
- [ ] 5.5 Verify constitution alignment: the change
  advances Principle V (Security by Default) by closing
  the T1 weakness for all label mutations.
