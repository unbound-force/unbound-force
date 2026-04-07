## 1. Implementation

- [x] 1.1 Add `gitignoreBlock` string constant to
  `internal/scaffold/scaffold.go` containing the full
  UF ignore block (marker comment + runtime patterns +
  legacy directories)
- [x] 1.2 Add `ensureGitignore()` function to
  `internal/scaffold/scaffold.go` that reads `.gitignore`,
  checks for the marker, and appends the block if missing.
  Creates the file if it does not exist. Returns a
  `subToolResult` result.
- [x] 1.3 Call `ensureGitignore()` from `Run()` after the
  legacy file warning but before sub-tool delegation.
  Result prepended to subResults for scaffold summary.

## 2. Tests

- [x] 2.1 Add test `TestEnsureGitignore_FreshDir` —
  no `.gitignore` exists, function creates it with
  the UF block including marker
- [x] 2.2 Add test `TestEnsureGitignore_ExistingNoBlock` —
  `.gitignore` exists with custom content, function
  appends UF block, existing content preserved
- [x] 2.3 Add test `TestEnsureGitignore_ExistingWithBlock` —
  `.gitignore` already has marker, function skips,
  file unchanged (idempotent)
- [x] 2.4 Add test `TestEnsureGitignore_Idempotent` —
  call twice, verify `.gitignore` content is identical
  after both calls (block not duplicated)

## 3. Documentation

- [x] 3.1 Add Recent Changes entry to `AGENTS.md`

## 4. Verification

- [x] 4.1 Run `go test -race -count=1 ./internal/scaffold/...`
- [x] 4.2 Run `go build ./...`
