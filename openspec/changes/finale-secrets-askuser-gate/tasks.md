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

## 1. Replace prose confirmation with AskUserQuestion gate

All edits target `.opencode/commands/finale.md`.
No [P] markers -- single file.

- [x] 1.1 Replace lines 79-81 (the prose confirmation
  "Ask for confirmation. If the user declines, stop and
  let them handle it manually.") with an explicit
  AskUserQuestion tool call:

  ```
  Use the **AskUserQuestion tool** with options:
    ["Yes -- stage all files including flagged ones",
     "No -- stop and let me handle it manually"]

  If the user selects "No", STOP. Do not run `git add .`.
  ```

  Preserve the surrounding warning message (lines 70-78)
  and the `git add .` line (line 83) unchanged.

**Checkpoint**: Read the modified file. Verify the
secrets check section (lines ~63-83) references the
**AskUserQuestion tool** with two options. Verify no
bare "Ask for confirmation" prose remains.

## 2. Sync scaffold asset

- [x] 2.1 Copy `.opencode/commands/finale.md` to
  `internal/scaffold/assets/opencode/commands/finale.md`
  (byte-identical).

**Checkpoint**: Run `go test ./internal/scaffold/... -count=1`.
`TestEmbeddedAssets_MatchSource` MUST pass.

## 3. Verification

- [x] 3.1 Run `go test -race -count=1 ./...` to verify
  no test regressions.

- [x] 3.2 Verify constitution alignment: confirm no
  artifact formats, hero interfaces, or machine-parseable
  outputs were modified. Only instruction text in
  markdown files was changed.
<!-- spec-review: passed -->
<!-- code-review: passed -->
<!-- scaffolded by uf vdev -->
