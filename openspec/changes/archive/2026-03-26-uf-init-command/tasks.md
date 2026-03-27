## 1. Write the Command File

- [x] 1.1 Create `.opencode/command/uf-init.md` with the full `/uf-init` command definition. The command instructs the LLM to: (1) check prerequisites (`.opencode/` exists, 4 OpenSpec skill files exist, 3 OpenSpec command files exist), (2) apply branch enforcement to 7 target files (skills: propose/apply/archive, commands: opsx-propose/opsx-apply/opsx-archive), (3) apply Dewey context queries to 5 target files (skills: propose/apply/explore, commands: opsx-propose/opsx-apply), (4) apply 3-tier Dewey degradation to 3 skill files (propose/apply/explore), (5) report results with emoji status indicators. Each customization includes: what to check for (idempotency), what to insert, where to insert it (described in prose for LLM reasoning), and the rationale. Error handling for missing files includes fix suggestions referencing `uf setup` and `uf init`. The command notes it should be re-run after OpenSpec CLI updates.

## 2. Scaffold Asset

- [x] 2.1 Copy `.opencode/command/uf-init.md` to `internal/scaffold/assets/opencode/command/uf-init.md` (the scaffold asset copy that gets embedded in the binary)

## 3. Update Tests

- [x] 3.1 Add `"opencode/command/uf-init.md"` to the `expectedAssetPaths` list in `internal/scaffold/scaffold_test.go`
- [x] 3.2 Update the file count assertion in `cmd/unbound-force/main_test.go` from the current count to current+1

## 4. Verify

- [x] 4.1 Run `go build ./...` to verify the scaffold asset embeds correctly
- [x] 4.2 Run `go test -race -count=1 ./internal/scaffold/...` to verify scaffold tests pass
- [x] 4.3 Run `go test -race -count=1 ./...` to verify full test suite passes
