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

## 1. Scaffold Engine — Logic

- [x] 1.1 Add `hasRuleContent(path string) bool` helper to
  `internal/scaffold/scaffold.go`. The function reads the file
  at `path`, finds the last occurrence of the sentinel string
  `<!-- Add project-specific rules below this line -->`, and
  returns `true` iff there is any non-whitespace content after
  it. Returns `true` on any I/O error (fail-open). Place the
  helper near `collectDeployedPacks`.
- [x] 1.2 Update `collectDeployedPacks(lang string) []string`
  signature to `collectDeployedPacks(lang, root string)
  []string`. When `root == ""`, behaviour is unchanged (all
  packs returned). When `root != ""`, filter each `*-custom.md`
  entry through `hasRuleContent(filepath.Join(root,
  ".opencode/uf/packs/", name))`.
- [x] 1.3 Update `buildCLAUDEmdBlock(lang string) string`
  signature to `buildCLAUDEmdBlock(lang, root string) string`.
  Pass `root` to `collectDeployedPacks`.
- [x] 1.4 Update the one call site of `buildCLAUDEmdBlock` in
  `ensureCLAUDEmd` (`scaffold.go:1093`) to pass
  `opts.TargetDir` as the `root` argument.

## 2. Scaffold Engine — Tests

- [x] 2.1 Add unit tests for `hasRuleContent` in
  `internal/scaffold/scaffold_test.go` (or a new
  `scaffold_claude_test.go` if the file is too large). Use
  `t.TempDir()`. Cover: empty stub (returns false), stub with
  content after sentinel (returns true), no sentinel present
  (returns true), file does not exist (returns true).
- [x] 2.2 Add unit tests for `collectDeployedPacks` with a
  non-empty root. Cover: all custom packs empty → custom packs
  absent from result; one custom pack populated → only that
  one included; non-custom packs always present.
- [x] 2.3 Add an integration-style test for
  `ensureCLAUDEmd` (using the existing dependency-injection
  pattern with `opts.ReadFile`/`opts.WriteFile`) that asserts
  empty custom packs are omitted from the generated block.
  Verify the existing test for non-empty root `""` behaviour
  still passes (backward compat).

## 3. CLAUDE.md — Remove Empty Imports

- [x] 3.1 Edit `CLAUDE.md` in the repo root: remove the three
  `@` import lines for empty custom packs:
  - `@.opencode/uf/packs/default-custom.md`
  - `@.opencode/uf/packs/content-custom.md`
  - `@.opencode/uf/packs/go-custom.md`

## 4. Verification

- [x] 4.1 Run `go build ./...` — must succeed with no errors.
- [x] 4.2 Run `go test -race -count=1 ./internal/scaffold/...`
  — all tests must pass.
- [x] 4.3 Run `go vet ./...` — no issues.
- [x] 4.4 Confirm `CLAUDE.md` no longer contains any
  `*-custom.md` import lines after the change.
- [x] 4.5 Verify constitution alignment: confirm no
  exported API was changed and no external dependencies were
  added.
