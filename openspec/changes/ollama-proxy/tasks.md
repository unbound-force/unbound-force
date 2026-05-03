## 1. Extract PID File Package

- [x] 1.1 Create `internal/pidfile/pidfile.go` — move
  `WritePID`, `ReadPID`, `IsAlive`, `CleanupStale`,
  `RemovePID`, `PIDInfo` from `internal/gateway/pid.go`
- [x] 1.2 Create `internal/pidfile/signal_unix.go` —
  move `signalZero` from `internal/gateway/signal_unix.go`
- [x] 1.3 Create `internal/pidfile/pidfile_test.go` —
  move PID-related tests from gateway
- [x] 1.4 Update `internal/gateway/pid.go`: remove all
  functions, replace with imports from `internal/pidfile/`
  (thin wrappers or direct usage)
- [x] 1.5 Update `internal/gateway/gateway.go`: change
  all PID calls to `pidfile.*`
- [x] 1.6 Update `internal/gateway/gateway_test.go`:
  change PID references to `pidfile.*`
- [x] 1.7 Verify gateway tests pass:
  `go test -race -count=1 ./internal/gateway/`
- [x] 1.8 Verify build: `go build ./...`

## 2. Extract TokenManager to `internal/auth/`

- [x] 2.1 Create `internal/auth/token.go` with
  `TokenManager` struct encapsulating: token string +
  `sync.RWMutex`, tokenExpiry, background refresh loop,
  proactive refresh with `TryLock()` dedup, atomic
  invalidation on failure
- [x] 2.2 Add `NewTokenManager(opts TokenManagerOpts)`
  constructor with fields: `RefreshFn`, `Lifetime`,
  `ProactiveWindow`, `Interval`, `ExecCmd`
- [x] 2.3 Add `Start(ctx)` method: initial token
  acquisition + background refresh goroutine
- [x] 2.4 Add `Stop()` method: cancel context
- [x] 2.5 Add `Token() (string, error)` method: return
  valid token or error, with proactive refresh
- [x] 2.6 Create `internal/auth/vertex.go` with exported
  `RefreshVertexToken(execCmd) (string, error)`
- [x] 2.7 Create `internal/auth/bedrock.go` with exported
  `RefreshBedrockCredentials(execCmd) (...)`,
  `ParseEnvExport`, `ParseAWSCredentialsJSON`
- [x] 2.8 Create `internal/auth/token_test.go` — tests
  for TokenManager lifecycle, proactive refresh,
  atomic invalidation on failure, context cancellation
- [x] 2.9 Create `internal/auth/vertex_test.go` — tests
  for RefreshVertexToken (success, failure, re-auth)
- [x] 2.10 Create `internal/auth/bedrock_test.go` — tests
  for RefreshBedrockCredentials and parse helpers
- [x] 2.11 Update `internal/gateway/provider.go`:
  replace inline token lifecycle in `VertexProvider`
  and `BedrockProvider` with `auth.TokenManager`
  instances. Update `VertexProvider.Start()` (2 calls),
  `BedrockProvider.Start()` (1 call)
- [x] 2.12 Update `internal/gateway/refresh.go`: remove
  `refreshVertexToken`, `refreshLoop`, `refreshMinute`,
  `refreshBedrockCredentials`, `parseEnvExport`,
  `parseAWSCredentialsJSON`. Keep SigV4 signing only.
- [x] 2.13 Update `internal/gateway/gateway_test.go`:
  update token refresh test references
- [x] 2.14 Count test functions before and after
  extraction: combined count across `internal/auth/`
  + `internal/gateway/` MUST equal or exceed
  pre-extraction count
- [x] 2.15 Verify all gateway tests pass:
  `go test -race -count=1 ./internal/gateway/`
- [x] 2.16 Verify build: `go build ./...`

## 3. Ollama Proxy Core

- [x] 3.1 Create `internal/ollamaproxy/proxy.go` with
  `Options` struct: `Port`, `EmbedModel`, `GatewayURL`,
  `ProjectDir`, `Detach`, `Stdout`, `Stderr` config
  fields + injectable deps (`LookPath`, `ExecCmd`,
  `ExecStart`, `Getenv`, `HTTPGet`, `FindProcess`,
  `ListenAndServe`, `HTTPClient`)
- [x] 3.2 Implement `defaults()` method — fill zero
  values (port 11434, embed model `text-embedding-005`,
  gateway URL `http://localhost:53147`)
- [x] 3.3 Implement `validateGatewayURL()`: validate
  scheme http/https, host is loopback (127.0.0.1, ::1,
  localhost), port valid, path empty or `/`. Reject
  non-loopback with clear error.
- [x] 3.4 Implement `Start(opts)`: validate config
  (including GatewayURL), validate gcloud in PATH,
  cleanup stale PID, check already-running (probe
  health for `"service": "uf-ollama-proxy"` to
  distinguish from real Ollama), init
  `auth.TokenManager`, check gateway health (warn if
  unavailable), log startup warning about `dewey
  reindex` if switching from local Ollama, register
  HTTP routes, write PID via `pidfile.WritePID`,
  start server with graceful shutdown
- [x] 3.5 Implement `Stop(opts)`: read PID via
  `pidfile.ReadPID`, check alive via `pidfile.IsAlive`,
  send SIGTERM, remove PID via `pidfile.RemovePID`
- [x] 3.6 Implement `Status(opts)`: read PID, check
  alive, display port/model/uptime/gateway status
- [x] 3.7 Implement detach mode: re-exec with
  `_UF_OLLAMA_PROXY_CHILD=1`, child redirects to
  `.uf/ollama-proxy.log` (0o600 perms), parent polls
  health
- [x] 3.8 Implement `GET /health` endpoint returning
  JSON with `service`, `status`, `port`, `embed_model`,
  `gateway_available`
- [x] 3.9 Add model name mapping table
  (`defaultModelMap`) with granite-embedding:30m,
  granite-embedding-small-english-r2, llama3.2:3b
- [x] 3.10 Add `mapModelName(name string) (string, bool)`
  — returns (cloud name, true) or ("", false) for
  unknown
- [x] 3.11 Add `validateModelName(name string) error` —
  reject names with `/`, `\`, `%`, or control chars.
  Allow alphanumeric, hyphens, colons, periods,
  underscores.
- [x] 3.12 Add `redactToken(body string) string` —
  strip Bearer tokens from error response bodies

## 4. Embed Handler

- [x] 4.1 Create `internal/ollamaproxy/embed.go` with
  `handleEmbed(w, r)` HTTP handler
- [x] 4.2 Parse Ollama embed request body
- [x] 4.3 Validate model name via `validateModelName()`
- [x] 4.4 Map model name via `mapModelName()` — reject
  unknown models with Ollama-format error
- [x] 4.5 Construct Vertex AI predict URL from region,
  project ID, and mapped model name (using
  `filepath.Join`-safe URL construction)
- [x] 4.6 Build Vertex request body with batch support
- [x] 4.7 Get token via `TokenManager.Token()` and
  inject `Authorization: Bearer` header
- [x] 4.8 Call Vertex AI endpoint via `HTTPClient`
- [x] 4.9 Parse Vertex response, extract embeddings
- [x] 4.10 Build Ollama response
- [x] 4.11 Handle errors: redact tokens via
  `redactToken()`, return Ollama-format errors
- [x] 4.12 Enforce max request body size (10MB default)
- [x] 4.13 Add test: `TestHandleEmbed_SingleInput` —
  mock Vertex via httptest, assert exact embedding
  values match, assert Authorization header is
  `Bearer <mock-token>`, assert Vertex request body
  contains correct instances
- [x] 4.14 Add test: `TestHandleEmbed_BatchInput` —
  assert count matches, vectors in correct order
- [x] 4.15 Add test: `TestHandleEmbed_VertexError` —
  assert error propagation, token redaction in
  Ollama-format error
- [x] 4.16 Add test: `TestHandleEmbed_UnknownModel` —
  assert rejection (not passthrough)
- [x] 4.17 Add test: `TestHandleEmbed_ModelPathTraversal`
  — assert rejection for `../../evil` model name
- [x] 4.18 Add test: `TestHandleEmbed_OversizedBody` —
  assert HTTP 413 for body > 10MB
- [x] 4.19 Add test: `TestValidateModelName` — unit
  test for validation function
- [x] 4.20 Add test: `TestMapModelName_Known` — unit
  test returning mapped name
- [x] 4.21 Add test: `TestMapModelName_Unknown` — unit
  test returning false

## 5. Generate Handler

- [x] 5.1 Create `internal/ollamaproxy/generate.go` with
  `handleGenerate(w, r)` HTTP handler
- [x] 5.2 Parse Ollama generate request
- [x] 5.3 Validate and map model name
- [x] 5.4 Build Anthropic Messages request with
  `max_tokens: 4096`
- [x] 5.5 POST to gateway URL via `HTTPClient`
- [x] 5.6 Parse Anthropic response: handle empty
  `content` array gracefully (no panic)
- [x] 5.7 Build Ollama response
- [x] 5.8 Handle gateway unavailable: return
  Ollama-format error with `uf gateway` instructions
- [x] 5.9 Handle gateway error responses (non-200):
  return Ollama-format error with gateway error details
- [x] 5.10 Enforce max request body size (10MB)
- [x] 5.11 Add test: `TestHandleGenerate_Success` —
  mock gateway via httptest, assert model mapping
- [x] 5.12 Add test: `TestHandleGenerate_GatewayDown` —
  assert error includes `uf gateway`
- [x] 5.13 Add test: `TestHandleGenerate_GatewayError` —
  assert error propagation for HTTP 500
- [x] 5.14 Add test: `TestHandleGenerate_EmptyContent` —
  assert graceful error for empty content array
- [x] 5.15 Add test: `TestHandleGenerate_MaxTokensSet` —
  assert 4096 in Anthropic request body

## 6. Tags Handler

- [x] 6.1 Create `internal/ollamaproxy/tags.go` with
  `handleTags(w, r)` HTTP handler
- [x] 6.2 Return synthetic model list from configured
  model names in Ollama tags format
- [x] 6.3 Add test: `TestHandleTags_ReturnsModels` —
  verify all mapped model names appear

## 7. Configuration

- [x] 7.1 Add `OllamaProxyConfig` struct to
  `internal/config/config.go` with `Port int`,
  `EmbedModel string`, `GatewayURL string`
- [x] 7.2 Add `OllamaProxy OllamaProxyConfig` field to
  `Config` struct
- [x] 7.3 Add merge logic in `merge()`
- [x] 7.4 Add env var overrides in `applyEnvOverrides()`:
  `UF_OLLAMA_PROXY_PORT`, `UF_OLLAMA_PROXY_EMBED_MODEL`,
  `UF_OLLAMA_PROXY_GATEWAY_URL`
- [x] 7.5 Add ollama-proxy section to config template
- [x] 7.6 Add test: `TestOllamaProxyConfig_Defaults`
- [x] 7.7 Add test: `TestOllamaProxyConfig_EnvOverrides`
- [x] 7.8 Add test: `TestOllamaProxyConfig_Merge`

## 8. Cobra Command

- [x] 8.1 Create `cmd/unbound-force/ollama_proxy.go`
  with `newOllamaProxyCmd()` — flags: `--port`,
  `--embed-model`, `--gateway-url`, `--detach`
- [x] 8.2 Implement `runOllamaProxy()`: construct
  `ollamaproxy.Options` from flags + config, call Start
- [x] 8.3 Add `stop` subcommand
- [x] 8.4 Add `status` subcommand
- [x] 8.5 Register in `cmd/unbound-force/main.go`
- [x] 8.6 Add test: `TestOllamaProxyCmd_Registered`
- [x] 8.7 Update `cmd/unbound-force/main_test.go` file
  count assertion if applicable

## 9. Proxy Lifecycle Tests

All lifecycle tests MUST use `t.TempDir()` for
`ProjectDir`, inject `ListenAndServe` to avoid real
port binding, and inject `FindProcess` to avoid real
process signaling (follow `testOpts()` from gateway).

- [x] 9.1 Add test: `TestStart_DefaultPort`
- [x] 9.2 Add test: `TestStart_CustomPort`
- [x] 9.3 Add test: `TestStart_GatewayWarning` — proxy
  starts even when gateway unreachable
- [x] 9.4 Add test: `TestStart_AlreadyRunning`
- [x] 9.5 Add test: `TestStart_GcloudMissing` — error
  with install instructions
- [x] 9.6 Add test: `TestStart_NonLoopbackGatewayURL` —
  error rejecting non-loopback
- [x] 9.7 Add test: `TestStop_Running`
- [x] 9.8 Add test: `TestStatus_Running`
- [x] 9.9 Add test: `TestHealthEndpoint_ResponseFields`
  — assert all JSON fields including `service`
- [x] 9.10 Add test: `TestStart_TokenRefreshFailure` —
  assert re-auth error (not stale token forwarding)

## 10. Documentation

- [x] 10.1 Update AGENTS.md project structure: add
  `internal/ollamaproxy/`, `internal/auth/`,
  `internal/pidfile/`, `cmd/.../ollama_proxy.go`
- [x] 10.2 Update AGENTS.md "Recent Changes"
- [x] 10.3 File GitHub issue in `unbound-force/dewey`
  for constitution amendment: "Local-Only Processing"
  → "local-only by default, cloud providers opt-in"
- [x] 10.4 File GitHub issue in `unbound-force/website`
  for documentation: `uf ollama-proxy` command,
  GPU-less developer workflow

## 11. Verification

- [x] 11.1 Run `go build ./...`
- [x] 11.2 Run `go test -race -count=1 ./...`
- [x] 11.3 Verify constitution alignment
<!-- spec-review: passed -->
<!-- code-review: passed -->
<!-- scaffolded by uf vdev -->
