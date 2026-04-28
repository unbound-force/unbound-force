## 1. Token Expiry Tracking (VertexProvider)

- [x] 1.1 Add `tokenExpiry time.Time` field to
  `VertexProvider` struct in `provider.go`
- [x] 1.2 Add `tokenRefreshing sync.Mutex` field to
  `VertexProvider` for proactive refresh
  deduplication
- [x] 1.3 Extract named constants:
  `vertexTokenLifetime = 55 * time.Minute` and
  `proactiveRefreshWindow = 5 * time.Minute`
- [x] 1.4 In `VertexProvider.Start()`, set
  `p.tokenExpiry = time.Now().Add(vertexTokenLifetime)`
  after successful initial token acquisition
  (`provider.go:213-215`)
- [x] 1.5 In the refresh closure
  (`provider.go:221-232`), set `p.tokenExpiry` on
  successful refresh
- [x] 1.6 In the refresh closure, on failure: clear
  `p.token = ""` and set `p.tokenExpiry` to zero
  value atomically under a single `p.tokenMu.Lock()`
  acquisition, after logging the error

## 2. Token Expiry Check in PrepareRequest (Vertex)

- [x] 2.1 In `VertexProvider.PrepareRequest()`, check
  empty token FIRST (existing check), then check
  expiry: if `!p.tokenExpiry.IsZero() &&
  time.Now().After(p.tokenExpiry)`, return error:
  "vertex AI token expired. Re-authenticate:
  gcloud auth application-default login"
- [x] 2.2 Add proactive refresh: if token expires
  within `proactiveRefreshWindow`
  (`time.Now().Add(proactiveRefreshWindow)
  .After(p.tokenExpiry)`), call
  `p.tryProactiveRefresh()`
- [x] 2.3 Implement `tryProactiveRefresh()` on
  `VertexProvider`:
  - Use `p.tokenRefreshing.TryLock()` to deduplicate
    concurrent attempts. If lock not acquired, return
    immediately (another goroutine is refreshing).
  - Create `context.WithTimeout(context.Background(),
    5*time.Second)` for the subprocess call.
  - Call `refreshVertexToken(p.execCmd)`. On success:
    acquire `p.tokenMu` write lock, update both
    `p.token` and `p.tokenExpiry` atomically, release
    `p.tokenMu`.
  - On failure (including timeout): log warning,
    release `tokenRefreshing`, do NOT clear the token
    (it may still be valid).
  - Always: `defer p.tokenRefreshing.Unlock()`

## 3. Bedrock Credential Expiry Parity

- [x] 3.1 Add `credExpiry time.Time` field and
  `credRefreshing sync.Mutex` to `BedrockProvider`
- [x] 3.2 Extract named constant:
  `bedrockCredLifetime = 50 * time.Minute`
- [x] 3.3 In `BedrockProvider.Start()`, set
  `p.credExpiry = time.Now().Add(bedrockCredLifetime)`
  after successful credential acquisition
- [x] 3.4 In the refresh closure
  (`provider.go:357-370`), set `p.credExpiry` on
  successful refresh
- [x] 3.5 In the refresh closure, on failure: clear
  `p.accessKey`, `p.secretKey`, `p.sessionToken` to
  empty strings AND set `p.credExpiry` to zero value,
  all atomically under a single `p.credMu.Lock()`
  acquisition, after logging the error
- [x] 3.6 Add expiry check at the start of
  `BedrockProvider.PrepareRequest()`: check empty
  credentials FIRST (existing check), then if
  credentials have expired, return error: "bedrock
  credentials expired. Re-authenticate: aws sso login"
- [x] 3.7 Add proactive refresh to
  `BedrockProvider.PrepareRequest()` using the same
  `TryLock` + 5-second timeout pattern as Vertex

## 4. Gateway Log File for Detached Mode

- [x] 4.1 In `detach()` (`gateway.go:539`), after
  creating the child `exec.Cmd`, open
  `.uf/gateway.log` with `os.OpenFile` using
  `O_CREATE|O_WRONLY|O_TRUNC` and `0600` permissions
  (owner-only since log contains auth diagnostics)
- [x] 4.2 Set `cmd.Stdout` and `cmd.Stderr` to the
  opened log file (instead of `nil`)
- [x] 4.3 Add the log file path to the success
  message: "Gateway started (PID %d) on port %d.
  Logs: .uf/gateway.log"
- [x] 4.4 Ensure `.uf/` directory exists before
  opening the log file (it should already exist from
  PID file creation, but guard defensively)

## 5. Status Command Log Path Display

- [x] 5.1 In `Status()` (`gateway.go:643`), after
  confirming the gateway is alive, check if
  `.uf/gateway.log` exists
- [x] 5.2 If the log file exists, add a "Log:" line
  to the status output showing the file path

## 6. Tests

- [x] 6.1 Add
  `TestVertexPrepareRequest_StaleTokenRegression`
  regression test (TC-006): construct a
  `VertexProvider` with a non-empty token and a past
  `tokenExpiry`, call `PrepareRequest`, verify it
  returns an error containing "token expired". This
  test reproduces the original failure (stale token
  forwarded without error) and passes only with the
  fix.
- [x] 6.2 Add `TestVertexPrepareRequest_ValidToken`
  test: set `tokenExpiry` to a future time, verify
  `PrepareRequest` succeeds
- [x] 6.3 Add
  `TestVertexPrepareRequest_ProactiveRefresh` test:
  set `tokenExpiry` to 3 minutes from now, inject
  `ExecCmd` that returns a new token, verify the
  token is refreshed and expiry is updated
- [x] 6.4 Add
  `TestVertexPrepareRequest_ProactiveRefreshFails`
  test: set `tokenExpiry` to 3 minutes from now,
  inject `ExecCmd` that fails, verify the request
  proceeds with the current (still valid) token and
  token is NOT cleared
- [x] 6.5 Add
  `TestVertexPrepareRequest_ProactiveRefreshTimeout`
  test: set `tokenExpiry` to 3 minutes from now,
  inject `ExecCmd` that blocks for 10 seconds, verify
  the request proceeds after 5 seconds with the
  current token
- [x] 6.6 Add
  `TestVertexRefreshFailure_ClearsToken` test:
  simulate a refresh failure in the refresh closure,
  verify `p.token` is empty and `p.tokenExpiry` is
  zero afterward
- [x] 6.7 Add
  `TestVertexPrepareRequest_ConcurrentProactiveRefresh`
  test: use `sync.WaitGroup` barrier to launch 5
  goroutines calling `PrepareRequest` simultaneously
  with near-expiry token, inject `ExecCmd` with
  `atomic.Int32` counter, verify exactly 1 invocation.
  Run under `-race`.
- [x] 6.8 Add
  `TestBedrockPrepareRequest_ExpiredCredentials`
  test: set `credExpiry` to a past time, verify
  `PrepareRequest` returns error containing
  "credentials expired"
- [x] 6.9 Add
  `TestBedrockRefreshFailure_ClearsCredentials`
  test: simulate refresh failure, verify accessKey,
  secretKey, sessionToken are empty and credExpiry is
  zero
- [x] 6.10 Add
  `TestBedrockPrepareRequest_ProactiveRefresh` test:
  set `credExpiry` to 3 minutes from now, inject
  `ExecCmd` that returns new credentials, verify
  credentials refreshed and expiry updated
- [x] 6.11 Add
  `TestBedrockPrepareRequest_ProactiveRefreshFails`
  test: set `credExpiry` to 3 minutes from now,
  inject `ExecCmd` that fails, verify request
  proceeds with current credentials
- [x] 6.12 Add `TestDetach_CreatesLogFile` test:
  in the injected `ExecStart`, assert that
  `cmd.Stdout` is a non-nil `*os.File` and that the
  file path ends with `gateway.log`. Verify the file
  was created in `t.TempDir()/.uf/` with `0600`
  permissions.
- [x] 6.13 Add `TestStatus_ShowsLogPath` test: create
  `.uf/gateway.log`, verify `Status()` output
  includes the log path
- [x] 6.14 Add `TestStart_ForegroundNoLogFile` test:
  start gateway in foreground mode (child path),
  verify `.uf/gateway.log` does not exist
- [x] 6.15 Update existing test call sites that
  construct `VertexProvider` or `BedrockProvider`
  directly to set `tokenExpiry`/`credExpiry` to a
  future time. Specifically update:
  `TestVertexProvider_PrepareRequest`,
  `TestVertexProvider_PrepareRequest_StreamingEndpoint`,
  `TestVertexProvider_PrepareRequest_NonStreamingEndpoint`,
  `TestVertexProvider_PrepareRequest_CountTokensAlwaysRawPredict`,
  `TestVertexProvider_PrepareRequest_HeaderStripping`,
  `TestVertexProvider_PrepareRequest_PreservesOtherHeaders`,
  `TestBedrockProvider_PrepareRequest`

## 7. Documentation and Verification

- [x] 7.1 Update AGENTS.md Recent Changes section
  with a summary of this change
- [x] 7.2 Verify `.uf/gateway.log` is covered by
  existing `.gitignore` patterns: run
  `git check-ignore .uf/gateway.log` and confirm
  it returns a match (covered by `*.log` on line 88)
- [x] 7.3 Assess Website Documentation Gate: this
  change modifies internal error messages and adds a
  log file -- no new CLI commands, flags, or workflows.
  Exempt per AGENTS.md criteria.
- [x] 7.4 Run `make check` (build, test, vet, lint)
  and verify all checks pass
- [x] 7.5 Verify constitution alignment:
  - Principle III (Observable Quality): confirm
    expired-token errors are JSON-formatted via
    `writeJSONError` with 502 status code
  - Principle IV (Testability): confirm all new
    behavior is testable via injected dependencies
    without external services
<!-- spec-review: passed -->
<!-- code-review: passed -->
