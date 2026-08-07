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

  Fixes: https://github.com/unbound-force/unbound-force/issues/419
-->

## 1. Fix cleanupRenamedCommands

Modify `cleanupRenamedCommands()` in
`internal/scaffold/scaffold.go` to accept an
`io.Writer` and log warnings for files that cannot
be removed.

- [x] 1.1 Update `cleanupRenamedCommands` signature
  to accept `io.Writer` as the second parameter.
  Update the call site at line 249 of `Run()` to
  pass `opts.Stdout`. When `os.Remove` fails for a
  file that `os.Stat` confirmed exists, write a
  warning to the writer in the format:
  `"  ⚠  could not remove %s: %v\n"`. Use
  `fmt.Fprintf` (not `charmbracelet/log`) to match
  the existing `warnLegacyReviewerFiles` pattern.
  Files: `internal/scaffold/scaffold.go`

## 2. Add warnStaleCommandRefs

Add a new function to scan agent files for references
to old command names and warn the user.

- [x] 2.1 Create `warnStaleCommandRefs(w io.Writer,
  targetDir string)` in `internal/scaffold/scaffold.go`.
  The function MUST: (a) glob
  `.opencode/agents/*.md` in targetDir, (b) read each
  file and scan for old command name patterns derived
  from `renamedCommands` keys (strip directory and
  `.md` extension, prefix with `/`), (c) for each
  match, record the agent file name, the stale
  reference, and the correct replacement (derived
  from the corresponding value in `renamedCommands`),
  (d) if any matches found, print a warning block
  listing affected files and replacements. Call this
  function from `Run()` after `cleanupRenamedCommands`,
  before `printSummary`.
  Files: `internal/scaffold/scaffold.go`

## 3. Add test coverage

- [x] 3.1 Add `TestCleanupRenamedCommands` to
  `internal/scaffold/scaffold_test.go`. Test cases:
  (a) happy path -- create old-name files in
  `t.TempDir()`, run cleanup, verify files removed
  and returned paths match mapped output paths
  (`.opencode/commands/...`); (b) no-op -- empty dir,
  verify empty return and no warnings; (c) partial
  failure -- make one file read-only (skip on Windows),
  verify remaining files cleaned and warning written
  to buffer. Use `bytes.Buffer` for output capture.
  Files: `internal/scaffold/scaffold_test.go`

- [x] 3.2 Add `TestWarnStaleCommandRefs` to
  `internal/scaffold/scaffold_test.go`. Test cases:
  (a) agent file with stale ref `/review-council` --
  verify warning output contains file name, stale ref,
  and replacement `/uf.review-council`; (b) agent file
  with only current refs -- verify no warning; (c) no
  agent files -- verify no warning and no error. Use
  `t.TempDir()` and `bytes.Buffer`.
  Files: `internal/scaffold/scaffold_test.go`

## 4. Verification

- [x] 4.1 Run `go test -race -count=1 ./internal/scaffold/...`
  and verify all tests pass including the new ones.
  Run `go vet ./...` to check for issues.
  Verify constitution alignment: Observable Quality
  (errors are now logged), Testability (new tests use
  isolated temp dirs and buffer injection).
<!-- spec-review: passed -->
<!-- code-review: passed -->
