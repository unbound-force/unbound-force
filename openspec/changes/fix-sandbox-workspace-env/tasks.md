## 1. Fix

- [x] 1.1 In `buildRunArgs()` in `internal/sandbox/config.go`, add `-e WORKSPACE=/workspace/<filepath.Base(opts.ProjectDir)>` when `useParentMount(opts)` is true. Place it next to the existing `--workdir` block.

## 2. Tests

- [x] 2.1 Update `TestBuildRunArgs_Isolated` in `internal/sandbox/sandbox_test.go` — verify args contain `WORKSPACE=/workspace/test-project`
- [x] 2.2 Update `TestBuildRunArgs_Direct` in `internal/sandbox/sandbox_test.go` — verify args contain `WORKSPACE=/workspace/test-project`
- [x] 2.3 Add `TestBuildRunArgs_NoParentNoWorkspace` in `internal/sandbox/sandbox_test.go` — verify args do NOT contain `WORKSPACE=` when `NoParent` is true
- [x] 2.4 Add `TestBuildRunArgs_RootFallbackNoWorkspace` in `internal/sandbox/sandbox_test.go` — verify args do NOT contain `WORKSPACE=` when `ProjectDir=/myproject` (root fallback)

## 3. Validation

- [x] 3.1 Run `go test -race -count=1 ./internal/sandbox/...` — all tests pass
- [x] 3.2 Run `go build ./...` — build succeeds
- [x] 3.3 Run `golangci-lint run` — no lint errors

<!-- spec-review: passed -->
<!-- code-review: passed -->
