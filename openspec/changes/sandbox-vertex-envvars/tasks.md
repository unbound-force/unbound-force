## 1. Implementation

- [x] 1.1 Add `ANTHROPIC_VERTEX_PROJECT_ID` and
  `CLAUDE_CODE_USE_VERTEX` to the `forwardedAPIKeys`
  slice in `internal/sandbox/config.go`

## 2. Tests

- [x] 2.1 Update `TestForwardedEnvVars` in
  `internal/sandbox/sandbox_test.go` to verify the
  new vars are forwarded when set

## 3. Verification

- [x] 3.1 Run `go build ./...` and
  `go test -race -count=1 ./internal/sandbox/...`
