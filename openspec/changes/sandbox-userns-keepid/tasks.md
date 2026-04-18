## 1. Implementation

- [x] 1.1 In `internal/sandbox/config.go`, add
  `--userns=keep-id` to `buildRunArgs()` when
  `platform.OS == "linux"`. Place after resource
  limits, before the image and command arguments.

## 2. Tests

- [x] 2.1 Add `TestBuildRunArgs_UsernsKeepId` to
  `internal/sandbox/sandbox_test.go` — verify
  `--userns=keep-id` is present in args when
  platform OS is `linux`
- [x] 2.2 Add `TestBuildRunArgs_UsernsNotOnMac` to
  verify `--userns=keep-id` is NOT present when
  platform OS is `darwin`

## 3. Verification

- [x] 3.1 Run `go build ./...` and
  `go test -race -count=1 ./internal/sandbox/...`

<!-- spec-review: passed -->
