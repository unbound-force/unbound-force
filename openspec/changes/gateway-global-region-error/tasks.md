## 1. Change newVertexProvider signature and add validation

- [x] 1.1 Change `newVertexProvider()` return type from
  `*VertexProvider` to `(*VertexProvider, error)` in
  `internal/gateway/provider.go`
- [x] 1.2 Replace the `if region == "" || region == "global"`
  block (line 182-184) with two separate checks:
  (a) `if region == "global"` returns an error with
  actionable message including the rejected value,
  reason (rawPredict requires specific region),
  recommended fix (ANTHROPIC_VERTEX_REGION), and
  example regions; (b) `if region == ""` falls back
  to `"us-east5"` (preserving existing default)
- [x] 1.3 Update the existing comment block (lines 168-174)
  to reflect the new error behavior instead of
  "rejected" language

## 2. Propagate error through callers

- [x] 2.1 Update `DetectProvider()` (line 66) to handle
  the error from `newVertexProvider()` -- change from
  `return newVertexProvider(getenv, execCmd), nil` to
  a two-value receive with error propagation
- [x] 2.2 Update `NewProviderByName()` (line 99) to handle
  the error from `newVertexProvider()` -- change from
  `return newVertexProvider(getenv, execCmd), nil` to
  a two-value receive with error propagation

## 3. Tests

- [x] 3.1 Add `TestNewVertexProvider_GlobalRegionError` --
  verify that `newVertexProvider()` returns a non-nil
  error when all region env vars resolve to `"global"`
  (VERTEX_LOCATION=global, CLOUD_ML_REGION=global,
  ANTHROPIC_VERTEX_REGION unset); verify the error
  message contains `"global"` and
  `"ANTHROPIC_VERTEX_REGION"`
- [x] 3.2 Add
  `TestNewVertexProvider_CloudMLRegionGlobalAlone` --
  verify that when only `CLOUD_ML_REGION=global` is set
  (VERTEX_LOCATION and ANTHROPIC_VERTEX_REGION unset),
  `newVertexProvider()` returns a non-nil error
- [x] 3.3 Add
  `TestNewVertexProvider_GlobalOverriddenBySpecificRegion`
  -- verify that when `VERTEX_LOCATION=global` but
  `ANTHROPIC_VERTEX_REGION=us-east5`,
  `newVertexProvider()` returns a valid provider with
  region `"us-east5"` and nil error
- [x] 3.4 Add `TestNewVertexProvider_EmptyRegionDefault` --
  verify that when no region env vars are set,
  `newVertexProvider()` returns a valid provider with
  region `"us-east5"` and nil error (existing default
  behavior preserved)
- [x] 3.5 Add `TestDetectProvider_GlobalRegionError` --
  verify that `DetectProvider()` returns the error
  from `newVertexProvider()` when
  `CLAUDE_CODE_USE_VERTEX=1`,
  `ANTHROPIC_VERTEX_PROJECT_ID=proj`, and
  `VERTEX_LOCATION=global`
- [x] 3.6 Add `TestNewProviderByName_VertexGlobalRegionError`
  -- verify that `NewProviderByName("vertex", ...)`
  returns the error from `newVertexProvider()` when
  `VERTEX_LOCATION=global`
- [x] 3.7 Update any existing tests that call
  `newVertexProvider()` to handle the new
  `(*VertexProvider, error)` return type

## 4. Documentation and verification

- [x] 4.1 Update AGENTS.md Recent Changes section with a
  summary of this change
- [x] 4.2 Verify constitution alignment: Observable
  Quality (PASS -- immediate error replaces delayed
  401), Testability (PASS -- error return testable
  via injected getenv)
- [x] 4.3 Run `go test -race -count=1 ./internal/gateway/...`
  and verify all tests pass
- [x] 4.4 Run `go vet ./...` and `golangci-lint run` and
  verify no new findings
<!-- spec-review: passed -->
<!-- code-review: passed -->
