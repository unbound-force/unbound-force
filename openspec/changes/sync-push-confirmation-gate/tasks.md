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

## 1. Go Backend: Add dry-run support

- [x] 1.1 Add `DryRun bool` field to `Syncer` struct in `internal/sync/sync.go`. When `DryRun` is true, `Push()` MUST collect pending actions (CREATE or UPDATE per item) and write structured preview output to `s.out` without calling `s.runner.Run()`. Output format per FR-003: one line per item (`CREATE`/`UPDATE`, item ID, title, optional issue number), followed by a summary line with totals.
- [x] 1.2 Add `--dry-run` boolean flag to the `sync-push` cobra command in `cmd/mutimind/main.go`. Wire the flag to set `syncer.DryRun = true` before calling `syncer.Push()`.

## 2. Go Backend: Tests

- [ ] 2.1 [P] Enhance `StubGHRunner` in `internal/sync/sync_test.go` with a `Calls int` field (incremented on each `Run()` invocation) to support zero-call assertions. Then add table-driven tests for dry-run mode: (a) dry-run with items to create, (b) dry-run with items to update, (c) dry-run with mixed create/update, (d) dry-run with single item filter, (e) dry-run with empty backlog, (f) dry-run with non-existent item ID (error path). Each test case MUST assert: (1) `StubGHRunner.Calls == 0`, (2) output matches expected FR-003 format (use `strings.Contains` for field presence), (3) summary line totals are correct. Use table-driven test structure per TC-006.
- [ ] 2.2 [P] Enhance `stubGHRunner` in `cmd/mutimind/main_test.go` with a `calls int` field. Add tests for the `--dry-run` flag: (a) `sync-push --dry-run` with items present -- assert output contains `CREATE`/`UPDATE` lines and summary line per FR-003, assert stub call count is zero; (b) `sync-push --dry-run <item-id>` with specific item -- assert only that item appears in output.

## 3. Command File: Add confirmation gate

- [ ] 3.1 Update `.opencode/commands/muti-mind.sync-push.md` to add the confirmation flow per FR-002: (1) invoke Go backend with `--dry-run` to get preview, (2) present preview to user, (3) if items are pending, use AskUserQuestion with options `["Yes -- sync to GitHub", "No -- abort"]`, (4) on confirmation invoke Go backend without `--dry-run`, (5) on abort inform user sync was cancelled, (6) if no items pending, skip confirmation and inform user.

## 4. Scaffold Registry

- [ ] 4.1 Verify that `.opencode/commands/muti-mind.sync-push.md` is referenced in the scaffold test registry at `internal/scaffold/scaffold_test.go`. The file path does not change in this spec, so no registry update is needed. Run `go test -race -count=1 ./internal/scaffold/...` to confirm the registry still passes.

## 5. Verification

- [ ] 5.1 Run `go test -race -count=1 ./internal/sync/... ./cmd/mutimind/...` to verify all tests pass.
- [ ] 5.2 Run `go vet ./...` and `golangci-lint run` to verify no lint violations.
- [ ] 5.3 Verify constitution alignment: confirm the change satisfies Testability (dry-run testable with GHRunner stub, no external calls), Observable Quality (structured preview output), and Security by Default (confirmation gate prevents unintended external actions).
<!-- spec-review: passed -->
