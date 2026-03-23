## 1. Add Scaffold Asset

- [x] 1.1 Copy `.opencode/command/cobalt-crush.md` to `internal/scaffold/assets/opencode/command/cobalt-crush.md`
- [x] 1.2 Verify `go build ./...` succeeds (scaffold asset embeds correctly)

## 2. Update Test Assertion

- [x] 2.1 Update the file count assertion in `cmd/unbound-force/main_test.go` from the current count to current+1
- [x] 2.2 Run `go test -race -count=1 ./cmd/unbound-force/...` to verify the count assertion passes

## 3. Verify

- [x] 3.1 Run `go test -race -count=1 ./internal/scaffold/...` to verify scaffold tests pass (including regression tests for stale unbound/graphthulhu references)
- [x] 3.2 Run `go test -race -count=1 ./...` to verify full test suite passes
- [x] 3.3 Verify constitution alignment: Composability (command works independently) and Testability (regression tests cover the new file) per proposal assessment
