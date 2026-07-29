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

## 1. Replace prose confirmation with AskUserQuestion

- [x] 1.1 [P] In `.opencode/commands/finale.md`, replace
  lines 79-80 (the prose "Ask for confirmation..." text)
  with an explicit **AskUserQuestion tool** call. Use
  options: "Yes -- stage all files including flagged
  ones" and "No -- stop and let me handle it manually".
  Add a conditional: if the user selects "No", STOP
  immediately and do not run `git add .`.
  **File**: `.opencode/commands/finale.md`

- [x] 1.2 [P] Apply the identical change to the scaffold
  asset at
  `internal/scaffold/assets/opencode/commands/finale.md`.
  The file MUST be byte-identical to the command file
  after both edits.
  **File**: `internal/scaffold/assets/opencode/commands/finale.md`

## 2. Verification

- [ ] 2.1 Run `make test` to verify the scaffold drift
  detection test (`TestEmbeddedAssets_MatchSource`) passes
  with both files updated.

- [ ] 2.2 Verify constitution alignment: confirm no
  artifact formats, hero interfaces, or machine-parseable
  outputs were modified (Autonomous Collaboration,
  Composability First, Observable Quality -- all N/A).
  Confirm the scaffold drift test covers Testability
  (PASS).
<!-- spec-review: passed -->
<!-- scaffolded by uf vdev -->
