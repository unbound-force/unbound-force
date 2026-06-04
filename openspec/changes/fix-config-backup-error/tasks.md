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

## 1. Fix backup write error handling

- [ ] 1.1 In `internal/config/init.go` at line 80, replace
  `_ = opts.WriteFile(backupPath, existing, 0o644)` with a
  checked call that returns `fmt.Errorf("write backup
  config: %w", err)` on failure. Add a comment citing CS-006
  and explaining why aborting is the correct behavior.

## 2. Add regression test

- [ ] 2.1 In `internal/config/init_test.go`, add import `"fmt"`
  if not present.
- [ ] 2.2 [P] Add test `TestInitFile_BackupWriteFailureAbortsUpdate`
  that: (a) creates a config missing the `gateway` section to
  trigger an update path; (b) injects a `WriteFile` stub via
  `InitOptions.WriteFile` that returns an error for `.bak`
  paths; (c) asserts `InitFile` returns a non-nil error
  containing `"write backup config"`; (d) asserts the original
  config file content is unchanged.

## 3. Verify

- [ ] 3.1 Run `go test -race -count=1 ./internal/config/...`
  and confirm all tests pass including the new regression test.
- [ ] 3.2 Run `go vet ./internal/config/...` and confirm no
  issues.
