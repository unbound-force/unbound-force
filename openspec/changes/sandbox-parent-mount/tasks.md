## 1. Options & CLI

- [x] 1.1 Add `NoParent bool` field to `Options` struct in `internal/sandbox/sandbox.go`
- [x] 1.2 Add `--no-parent` flag to the `start` subcommand in `cmd/unbound-force/sandbox.go`, wired to `opts.NoParent`

## 2. Volume Mount

- [x] 2.1 Update `buildVolumeMounts()` in `internal/sandbox/config.go` — when `opts.NoParent` is false (default), mount `filepath.Dir(opts.ProjectDir)` at `/workspace` instead of `opts.ProjectDir`. When `opts.NoParent` is true, use current behavior (project-only mount).
- [x] 2.2 Add root directory fallback in `buildVolumeMounts()` — if `filepath.Dir(opts.ProjectDir)` returns `/`, fall back to project-only mount regardless of `NoParent` setting. Log at debug level.
- [x] 2.3 Preserve `:ro` suffix for isolated mode and `,Z` suffix for SELinux on the parent mount (FR-043).

## 3. Working Directory

- [x] 3.1 Update `buildRunArgs()` in `internal/sandbox/config.go` — when parent mount is active (not `NoParent` and not root fallback), add `--workdir /workspace/<filepath.Base(opts.ProjectDir)>` to the podman args. Insert before the image name argument.

## 4. Tests

- [x] 4.1 Add `TestBuildVolumeMounts_ParentMount` in `internal/sandbox/sandbox_test.go` — verify parent directory is mounted at `/workspace` and `--workdir` is set to `/workspace/<basename>`
- [x] 4.2 Add `TestBuildVolumeMounts_NoParentFlag` in `internal/sandbox/sandbox_test.go` — verify `--no-parent` uses project-only mount with no `--workdir`
- [x] 4.3 Add `TestBuildVolumeMounts_RootFallback` in `internal/sandbox/sandbox_test.go` — verify `ProjectDir=/myproject` falls back to project-only mount
- [x] 4.4 Add `TestBuildVolumeMounts_ParentMountIsolated` in `internal/sandbox/sandbox_test.go` — verify `:ro` is applied to parent mount in isolated mode
- [x] 4.5 Add `TestBuildVolumeMounts_ParentMountSELinux` in `internal/sandbox/sandbox_test.go` — verify `,Z` is applied to parent mount when SELinux is true
- [x] 4.6 Update existing `TestBuildRunArgs_*` tests to account for new `--workdir` argument in output

## 5. Validation

- [x] 5.1 Run `go test -race -count=1 ./internal/sandbox/...` — all tests pass
- [x] 5.2 Run `go build ./...` — build succeeds
- [x] 5.3 Run `golangci-lint run` — no lint errors
- [x] 5.4 Verify constitution alignment: Testability (all new logic testable via `Options` injection, no external services required)

<!-- spec-review: passed -->
<!-- code-review: passed -->
