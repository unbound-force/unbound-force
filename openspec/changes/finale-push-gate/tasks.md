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

## 1. Add push confirmation gate to /finale

All edits target `.opencode/commands/finale.md`.
No [P] markers -- single file, sequential edits.

- [ ] 1.1 Add AskUserQuestion gate in Step 4
  (lines 156-165). After the upstream detection check
  (line 160) and before the push commands (lines 163-164),
  insert: Use the **AskUserQuestion tool** with options
  `["Push to remote", "Abort -- do not push"]`.
  If the user selects "Abort -- do not push", report
  that the push was aborted and **STOP**. If the user
  selects "Push to remote", proceed with the existing
  push logic.

**Checkpoint**: Read the modified file. Verify the
AskUserQuestion gate appears between the upstream
detection and the `git push` commands. Verify both
options are present with action-descriptive labels.

## 2. Sync scaffold asset

- [ ] 2.1 Copy `.opencode/commands/finale.md` to
  `internal/scaffold/assets/opencode/commands/finale.md`
  (byte-identical).

**Checkpoint**: Run `go test ./internal/scaffold/...
-count=1`. `TestEmbeddedAssets_MatchSource` MUST pass.

## 3. Verification

- [ ] 3.1 Run `go test -race -count=1 ./...` to verify
  no test regressions.

- [ ] 3.2 Verify constitution alignment: confirm no
  artifact formats, hero interfaces, or machine-parseable
  outputs were modified. Only instruction text in
  markdown files was changed.
<!-- scaffolded by uf vdev -->
