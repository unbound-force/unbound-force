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

## 1. Add workdir and WORKSPACE to persistent args

- [x] 1.1 In `internal/sandbox/podman.go`, add
  `--workdir` and `-e WORKSPACE=` to
  `buildPersistentRunArgs()`. Insert after the UID
  mapping args (line ~311) and before the image
  argument. Use `filepath.Base(opts.ProjectDir)` to
  derive the project basename. Pattern:
  ```go
  projectSubdir := fmt.Sprintf("/workspace/%s",
      filepath.Base(opts.ProjectDir))
  args = append(args, "--workdir", projectSubdir)
  args = append(args, "-e",
      fmt.Sprintf("WORKSPACE=%s", projectSubdir))
  ```

## 2. Add test coverage

- [x] 2.1 In `internal/sandbox/sandbox_test.go`, update
  `TestBuildPersistentRunArgs_IncludesUserNS` (or add a
  new test `TestBuildPersistentRunArgs_SetsWorkdir`) to
  verify:
  - `--workdir /workspace/test-project` is present
  - `WORKSPACE=/workspace/test-project` is present in
    the args (as `-e WORKSPACE=...`)
  Derive the expected basename from the `testOpts()`
  project directory.

## 3. Verification

- [x] 3.1 Run `go test -race -count=1 ./internal/sandbox/...`
  and verify all tests pass.
- [x] 3.2 Run `make check` (lint + test + build) and
  verify clean.
- [x] 3.3 Verify constitution alignment: the fix uses
  dependency injection via `Options.ProjectDir`
  (Principle IV), produces observable env vars
  (Principle III), and requires no inter-hero changes
  (Principles I and II).

<!-- spec-review: passed -->
<!-- code-review: passed -->
