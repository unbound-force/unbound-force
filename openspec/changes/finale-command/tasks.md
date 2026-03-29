## 1. Create the slash command

- [x] 1.1 Create `.opencode/command/finale.md` with
  the full command definition including: branch safety
  gate, auto-stage with secrets warning, commit message
  generation and approval, push, PR creation (or reuse
  existing), CI check watching, rebase merge, return to
  main, and completion summary.

- [x] 1.2 Copy `.opencode/command/finale.md` to
  `internal/scaffold/assets/opencode/command/finale.md`
  so it is included in the scaffold embed.

## 2. Update scaffold tests

- [x] 2.1 Update the expected file count in
  `cmd/unbound-force/main_test.go` from 50 to 51 to
  account for the new `finale.md` command file.

## 3. Verification

- [x] 3.1 Run `go build ./...` to verify the build
  succeeds with the new embedded asset.

- [x] 3.2 Run `go test -race -count=1 ./...` to verify
  all tests pass including the updated file count
  assertion and scaffold drift detection.

- [x] 3.3 Verify constitution alignment: Composability
  (command works standalone, no hero dependencies) and
  Testability (file count test updated).
