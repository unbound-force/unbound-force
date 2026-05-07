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

## 1. Add PR content confirmation to /finale

- [x] 1.1 In `.opencode/commands/finale.md` Step 5,
  insert a new substep (d) between body generation
  (current c) and PR creation (current d). The substep
  MUST show the proposed PR title and body in a
  formatted block and ask "Approve, edit, or provide
  your own?" — mirroring the Step 3c commit message
  confirmation pattern. Renumber subsequent substeps
  (d→e create, e→f report URL).

## 2. Add PR creation guardrail

- [x] 2.1 In `.opencode/commands/finale.md` Guardrails
  section, add "NEVER create a PR without user approval
  of the title and body" after the existing "NEVER
  commit without user approval of the message" line.

## 3. Sync scaffold asset

- [x] 3.1 Copy `.opencode/commands/finale.md` to
  `internal/scaffold/assets/opencode/commands/finale.md`
  to keep the scaffold asset in sync with the live
  command.

## 4. Documentation

- [x] 4.1 [P] Add CHANGELOG.md entry for
  `opsx/finale-pr-approval` describing the change.
- [x] 4.2 [P] Verify `TestEmbeddedAssets_MatchSource`
  passes (scaffold drift test confirms live and asset
  copies match).

## 5. Verification

- [x] 5.1 Run `make check` (build, test, vet) to
  confirm no regressions.

<!-- spec-review: passed -->
<!-- code-review: passed -->
