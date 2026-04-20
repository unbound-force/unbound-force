# Tasks: LLM Gateway Command

**Input**: Design documents from `specs/033-gateway-command/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/gateway-api.md

**Organization**: Tasks are grouped by implementation phase (per plan.md) and tagged with user stories. Tests are included per the coverage strategy in plan.md.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Gateway Core — PID File & Options (US1, US5)

**Purpose**: Build the foundational types (Options, PIDInfo, HealthResponse), PID file management, and the gateway struct with health endpoint. This phase delivers the skeleton that all other phases build on.

**Depends on**: Nothing (standalone)

### PID File Management

- [x] T001 [P] [US5] Create `internal/gateway/pid.go` — define `PIDInfo` struct (PID, Port, Provider, Started fields) and `WritePID(path string, info PIDInfo) error` function that writes PID file atomically (write to temp file, rename) with `key=value` metadata lines per data-model.md format
- [x] T002 [P] [US5] Implement `ReadPID(path string) (*PIDInfo, error)` in `internal/gateway/pid.go` — parse PID from line 1, parse `port=`, `provider=`, `started=` metadata from subsequent lines, ignore unknown keys for forward compatibility
- [x] T003 [US5] Implement `IsAlive(pid int, findProcess func(int) (*os.Process, error)) bool` in `internal/gateway/pid.go` — use injected `findProcess` to look up PID, send signal 0 to check liveness, return false if process not found or signal fails
- [x] T004 [US5] Implement `RemovePID(path string) error` in `internal/gateway/pid.go` — remove PID file, return nil if file does not exist (idempotent)
- [x] T005 [US5] Implement `CleanupStale(path string, findProcess func(int) (*os.Process, error)) error` in `internal/gateway/pid.go` — read PID file, check if process is alive, remove PID file if stale, return nil if no PID file exists

### Gateway Options & Types

- [x] T006 [P] [US1] Create `internal/gateway/gateway.go` — define `Options` struct per data-model.md (Port, ProviderName, Detach, ProjectDir, Stdout, Stderr, and all injectable function fields: LookPath, ExecCmd, Getenv, HTTPGet, ReadFile, WriteFile, FindProcess, ListenAndServe), define `defaults()` method that fills zero-value fields with production implementations following the pattern from `internal/sandbox/sandbox.go`
- [x] T007 [P] [US1] Define `HealthResponse` struct in `internal/gateway/gateway.go` per data-model.md (Status, Provider, Port, PID, UptimeSeconds with JSON tags), add `DefaultPort` constant (53147) and `GatewayChildEnv` sentinel constant (`_UF_GATEWAY_CHILD`)

### Health Endpoint & HTTP Routing

- [x] T008 [US1] Implement `newMux(provider Provider, port int, startTime time.Time) http.Handler` in `internal/gateway/gateway.go` — create `http.ServeMux` with routes: `GET /health` returns `HealthResponse` JSON (FR-006), `POST /v1/messages` and `POST /v1/messages/count_tokens` forward to `httputil.ReverseProxy` with provider's `PrepareRequest` as Director (FR-001), all other paths return 405 with JSON error body per contracts/gateway-api.md
- [x] T009 [US1] Implement ReverseProxy Director function in `internal/gateway/gateway.go` — call `provider.PrepareRequest(req)`, preserve `anthropic-beta`, `anthropic-version`, `X-Claude-Code-Session-Id` headers (FR-002, FR-015), strip inbound `Authorization` and `x-api-key` headers before provider injection (FR-013), set `ErrorHandler` to return 502 JSON error on upstream failure

### Gateway Lifecycle

- [x] T010 [US1] Implement `Start(opts Options) error` in `internal/gateway/gateway.go` — call `defaults()`, detect provider (or use `ProviderName` override per FR-009), call `provider.Start(ctx)`, clean up stale PID files, check for port conflicts (return clear error per edge case), write PID file, call `ListenAndServe`, handle graceful shutdown on SIGINT/SIGTERM (stop provider, remove PID file)
- [x] T011 [US5] Implement `detach(opts Options) (int, error)` in `internal/gateway/gateway.go` — re-exec the binary with `_UF_GATEWAY_CHILD=1` sentinel env var per research.md R5, set `SysProcAttr.Setsid` for terminal detach, wait for health endpoint with exponential backoff, return child PID (FR-007)
- [x] T012 [US5] Implement `Stop(opts Options) error` in `internal/gateway/gateway.go` — read PID file, check if process is alive, send SIGTERM, wait briefly for process exit, remove PID file, print "Gateway stopped." (FR-008), print "No gateway running." if no PID file or process not alive
- [x] T013 [US5] Implement `Status(opts Options) error` in `internal/gateway/gateway.go` — read PID file, check if process is alive, if alive: query health endpoint and display provider, port, PID, uptime in human-readable format per data-model.md, if not alive: clean up stale PID file and print "No gateway running." (FR-008)

**Checkpoint**: PID file management and gateway lifecycle are complete. `Start()` can be called with a mock provider. Health endpoint serves JSON. Background mode works via re-exec.

---

## Phase 2: Provider Implementations (US1, US3, US4)

**Purpose**: Implement the three provider strategies (Anthropic, Vertex AI, Bedrock) with credential injection, URL rewriting, and token refresh. The `Provider` interface and `DetectProvider` function enable auto-detection (FR-003) and explicit override (FR-009).

**Depends on**: Phase 1 (gateway core must exist for `Provider` interface usage)

### Provider Interface & Detection

- [x] T014 [US1] Create `internal/gateway/provider.go` — define `Provider` interface per data-model.md and contracts/gateway-api.md (Name, PrepareRequest, Start, Stop methods), implement `DetectProvider(getenv func(string) string, execCmd func(string, ...string) ([]byte, error)) (Provider, error)` with priority order: Vertex (CLAUDE_CODE_USE_VERTEX=1 + ANTHROPIC_VERTEX_PROJECT_ID) → Bedrock (CLAUDE_CODE_USE_BEDROCK=1) → Anthropic (ANTHROPIC_API_KEY) → error listing supported providers (FR-003)
- [x] T015 [US4] Implement `NewProviderByName(name string, getenv func(string) string, execCmd func(string, ...string) ([]byte, error)) (Provider, error)` in `internal/gateway/provider.go` — create provider by explicit name ("anthropic", "vertex", "bedrock"), return error listing valid names for invalid input (FR-009)

### Anthropic Provider

- [x] T016 [P] [US1] Implement `AnthropicProvider` struct and methods in `internal/gateway/provider.go` — `Name()` returns "anthropic", `Start()` reads `ANTHROPIC_API_KEY` from env (error if empty), `PrepareRequest()` sets URL to `https://api.anthropic.com` + request path, adds `x-api-key` header, `Stop()` is no-op

### Vertex AI Provider

- [x] T017 [P] [US1] Implement `VertexProvider` struct in `internal/gateway/provider.go` — fields: projectID, region (default "us-east5" from CLOUD_ML_REGION), token (sync.RWMutex-protected), cancel context.CancelFunc, execCmd injection
- [x] T018 [US1] Implement `VertexProvider.Start(ctx)` in `internal/gateway/provider.go` — call `gcloud auth application-default print-access-token` via injected `execCmd` to get initial token, return clear error if gcloud fails (US3 scenario 2), start refresh goroutine
- [x] T019 [US1] Implement `VertexProvider.PrepareRequest(req)` in `internal/gateway/provider.go` — set URL to Vertex rawPredict endpoint (`https://{region}-aiplatform.googleapis.com/v1/projects/{projectID}/locations/{region}/publishers/anthropic/models/{model}:rawPredict`), extract model from request body (default `claude-sonnet-4-20250514`), add `Authorization: Bearer` header with current token (read under RLock)
- [x] T020 [US1] Implement `VertexProvider.Stop()` in `internal/gateway/provider.go` — cancel refresh goroutine context

### Bedrock Provider

- [x] T021 [P] [US1] Implement `BedrockProvider` struct in `internal/gateway/provider.go` — fields: region (from AWS_REGION or AWS_DEFAULT_REGION), accessKey, secretKey, sessionToken (sync.RWMutex-protected), cancel context.CancelFunc, execCmd injection
- [x] T022 [US1] Implement `BedrockProvider.Start(ctx)` in `internal/gateway/provider.go` — call `aws configure export-credentials --format env` via injected `execCmd` to get initial credentials, parse JSON response for AccessKeyId, SecretAccessKey, SessionToken, return clear error if aws CLI fails, start refresh goroutine
- [x] T023 [US1] Implement `BedrockProvider.PrepareRequest(req)` in `internal/gateway/provider.go` — set URL to Bedrock invoke endpoint (`https://bedrock-runtime.{region}.amazonaws.com/model/{model}/invoke`), extract model from request body, sign request with SigV4 using current credentials (read under RLock)
- [x] T024 [US1] Implement `BedrockProvider.Stop()` in `internal/gateway/provider.go` — cancel refresh goroutine context

### Token Refresh

- [x] T025 [P] [US3] Create `internal/gateway/refresh.go` — implement `refreshLoop(ctx context.Context, interval time.Duration, refreshFn func() error)` generic refresh goroutine with ticker, context cancellation, and error logging via `charmbracelet/log` (FR-005)
- [x] T026 [US3] Implement `refreshVertexToken(execCmd func(string, ...string) ([]byte, error)) (string, error)` in `internal/gateway/refresh.go` — call `gcloud auth application-default print-access-token`, return token string or error with clear message suggesting re-authentication (US3 scenario 2)
- [x] T027 [US3] Implement `refreshBedrockCredentials(execCmd func(string, ...string) ([]byte, error)) (accessKey, secretKey, sessionToken string, err error)` in `internal/gateway/refresh.go` — call `aws configure export-credentials --format env`, parse JSON response, return credentials or error

### SigV4 Signing

- [x] T028 [US1] Implement `signV4(req *http.Request, region, service, accessKey, secretKey, sessionToken string) error` in `internal/gateway/refresh.go` — minimal SigV4 implementation using `crypto/hmac` + `crypto/sha256` per research.md R3 (~200 lines), set `Authorization`, `X-Amz-Date`, and optionally `X-Amz-Security-Token` headers

**Checkpoint**: All three providers are implemented. `DetectProvider` auto-detects from env vars. Token refresh runs in background goroutines. SigV4 signing works for Bedrock. The gateway can now proxy requests to any provider.

---

## Phase 3: CLI Commands (US1, US4, US5)

**Purpose**: Wire the gateway package into Cobra commands. Register `uf gateway`, `uf gateway stop`, and `uf gateway status` in the CLI.

**Depends on**: Phase 1 + Phase 2 (gateway package must be functional)

- [x] T029 [US1] Create `cmd/unbound-force/gateway.go` — implement `newGatewayCmd() *cobra.Command` returning the `uf gateway` parent command with `--port` (int, default 53147), `--provider` (string, default ""), and `--detach` (bool) flags, `RunE` calls `gateway.Start()` with options populated from flags and `os.Getwd()` for ProjectDir
- [x] T030 [US5] Implement `newGatewayStopCmd() *cobra.Command` in `cmd/unbound-force/gateway.go` — `uf gateway stop` subcommand, `RunE` calls `gateway.Stop()` with ProjectDir from cwd
- [x] T031 [US5] Implement `newGatewayStatusCmd() *cobra.Command` in `cmd/unbound-force/gateway.go` — `uf gateway status` subcommand, `RunE` calls `gateway.Status()` with ProjectDir from cwd
- [x] T032 [US1] Register gateway command in `cmd/unbound-force/main.go` — add `root.AddCommand(newGatewayCmd())` and import `gateway` package

**Checkpoint**: `uf gateway`, `uf gateway stop`, and `uf gateway status` are available in the CLI. Help output shows all flags and subcommands.

---

## Phase 4: Sandbox Integration (US2)

**Purpose**: Modify the sandbox to auto-start the gateway when a cloud provider is detected, pass the gateway URL to the container, and skip credential mounts when the gateway is active.

**Depends on**: Phase 1 + Phase 2 + Phase 3 (gateway must be fully functional)

### Gateway Detection & Auto-Start

- [x] T033 [US2] Implement `gatewayHealthCheck(httpGet func(string) (int, error), port int) bool` in `internal/sandbox/sandbox.go` — HTTP GET to `http://localhost:{port}/health`, return true if status 200, false otherwise (used to detect already-running gateway)
- [x] T034 [US2] Implement `autoStartGateway(opts Options) (int, bool, error)` in `internal/sandbox/sandbox.go` — detect provider from env vars (same priority as gateway: Vertex → Bedrock → Anthropic), if provider detected: check if gateway already running via health check (reuse existing if so), if not running: start `uf gateway --detach` via `ExecCmd`, wait for health endpoint with exponential backoff, return (port, true, nil), if no provider detected: return (0, false, nil) for fallback (FR-010, FR-012)

### Gateway-Aware Container Configuration

- [x] T035 [US2] Implement `gatewayEnvVars(port int) []string` in `internal/sandbox/config.go` — return `-e ANTHROPIC_BASE_URL=http://host.containers.internal:{port}` and `-e ANTHROPIC_AUTH_TOKEN=gateway` flag pairs (FR-011)
- [x] T036 [US2] Modify `forwardedEnvVars(opts Options)` in `internal/sandbox/config.go` — add `gatewayActive bool` parameter, when gateway is active: skip `ANTHROPIC_API_KEY`, `ANTHROPIC_VERTEX_PROJECT_ID`, `CLAUDE_CODE_USE_VERTEX` from forwarded keys (FR-011), keep `OLLAMA_HOST`, `OPENAI_API_KEY`, `GEMINI_API_KEY`, `OPENROUTER_API_KEY` unchanged
- [x] T037 [US2] Modify `googleCloudCredentialMounts(opts Options, platform PlatformConfig)` in `internal/sandbox/config.go` — add `gatewayActive bool` parameter, when gateway is active: skip all gcloud credential mounts (no service account file mount, no gcloud dir mount) since the gateway handles authentication (FR-011)
- [x] T038 [US2] Modify `buildRunArgs(opts Options, platform PlatformConfig)` in `internal/sandbox/config.go` — add `gatewayActive bool` and `gatewayPort int` parameters, when gateway is active: call `gatewayEnvVars(port)` instead of credential mounts and provider API key forwarding, pass updated `gatewayActive` to `forwardedEnvVars` and `googleCloudCredentialMounts`

### Sandbox Start/Stop Integration

- [x] T039 [US2] Modify `Start()` in `internal/sandbox/sandbox.go` — before container start: call `autoStartGateway(opts)`, pass `gatewayActive` and `gatewayPort` to `buildRunArgs`, log "Gateway active on port {port} — credentials proxied" when gateway is used, log "Gateway not available — using credential mounts" when falling back (FR-012)
- [x] T040 [US2] Verify backward compatibility — when no provider env vars are set, `Start()` must produce identical `buildRunArgs` output as before this change (no gateway env vars, credential mounts preserved, all API keys forwarded)

**Checkpoint**: `uf sandbox start` auto-starts the gateway when a provider is detected. Container receives `ANTHROPIC_BASE_URL` instead of credential mounts. Fallback to credential mounts works when no provider is detected.

---

## Phase 5: Tests (All US)

**Purpose**: Comprehensive test coverage for all new and modified code per the coverage strategy in plan.md.

**Depends on**: Phase 1 + Phase 2 + Phase 3 + Phase 4

### PID File Tests

- [x] T041 [P] [US5] Test `WritePID` and `ReadPID` round-trip in `internal/gateway/gateway_test.go` — write PIDInfo with all fields, read back, verify all fields match, use `t.TempDir()`
- [x] T042 [P] [US5] Test `ReadPID` with malformed file in `internal/gateway/gateway_test.go` — non-numeric PID, missing metadata lines, empty file, verify error handling
- [x] T043 [P] [US5] Test `IsAlive` with injected `FindProcess` in `internal/gateway/gateway_test.go` — mock process found (alive), mock process not found (dead), mock signal error
- [x] T044 [P] [US5] Test `RemovePID` idempotency in `internal/gateway/gateway_test.go` — remove existing file, remove non-existent file (no error)
- [x] T045 [P] [US5] Test `CleanupStale` in `internal/gateway/gateway_test.go` — stale PID file (process dead) is removed, active PID file (process alive) is preserved, no PID file returns nil

### Provider Detection Tests

- [x] T046 [P] [US1] Test `DetectProvider` priority order in `internal/gateway/gateway_test.go` — Vertex detected when CLAUDE_CODE_USE_VERTEX=1 + ANTHROPIC_VERTEX_PROJECT_ID set, Bedrock detected when CLAUDE_CODE_USE_BEDROCK=1 set, Anthropic detected when only ANTHROPIC_API_KEY set, error when no vars set
- [x] T047 [P] [US4] Test `NewProviderByName` in `internal/gateway/gateway_test.go` — valid names ("anthropic", "vertex", "bedrock") return correct provider type, invalid name returns error listing valid names
- [x] T048 [P] [US1] Test `DetectProvider` precedence in `internal/gateway/gateway_test.go` — when both ANTHROPIC_API_KEY and CLAUDE_CODE_USE_VERTEX=1 are set, Vertex is selected (Vertex has higher priority)

### Anthropic Provider Tests

- [x] T049 [P] [US1] Test `AnthropicProvider.PrepareRequest` in `internal/gateway/gateway_test.go` — verify URL rewritten to `https://api.anthropic.com/v1/messages`, verify `x-api-key` header set, verify `anthropic-beta` and `anthropic-version` headers preserved, verify inbound `Authorization` header stripped
- [x] T050 [P] [US1] Test `AnthropicProvider.Start` in `internal/gateway/gateway_test.go` — success when ANTHROPIC_API_KEY set, error when ANTHROPIC_API_KEY empty

### Vertex Provider Tests

- [x] T051 [P] [US1] Test `VertexProvider.PrepareRequest` in `internal/gateway/gateway_test.go` — verify URL rewritten to Vertex rawPredict endpoint with correct project/region/model, verify `Authorization: Bearer` header set with current token
- [x] T052 [P] [US1] Test `VertexProvider.Start` with mock `execCmd` in `internal/gateway/gateway_test.go` — success when gcloud returns token, error when gcloud fails (verify error message suggests re-authentication)
- [x] T053 [P] [US3] Test `VertexProvider` token refresh in `internal/gateway/gateway_test.go` — verify `refreshLoop` calls refresh function on ticker, verify token is updated under mutex, verify context cancellation stops the loop

### Bedrock Provider Tests

- [x] T054 [P] [US1] Test `BedrockProvider.PrepareRequest` in `internal/gateway/gateway_test.go` — verify URL rewritten to Bedrock invoke endpoint, verify SigV4 signature present in `Authorization` header, verify `X-Amz-Date` header set
- [x] T055 [P] [US1] Test `BedrockProvider.Start` with mock `execCmd` in `internal/gateway/gateway_test.go` — success when aws CLI returns credentials JSON, error when aws CLI fails
- [x] T056 [P] [US3] Test `BedrockProvider` credential refresh in `internal/gateway/gateway_test.go` — verify refresh goroutine updates credentials, verify context cancellation stops the loop

### SigV4 Signing Tests

- [x] T057 [P] [US1] Test `signV4` with known test vectors in `internal/gateway/gateway_test.go` — use AWS-published SigV4 test vectors to verify signature correctness, verify `X-Amz-Security-Token` header set when session token present, verify header absent when session token empty

### Gateway Core Tests

- [x] T058 [P] [US1] Test `newMux` health endpoint in `internal/gateway/gateway_test.go` — `GET /health` returns 200 with correct JSON fields (status, provider, port, pid, uptime_seconds), verify Content-Type is `application/json`
- [x] T059 [P] [US1] Test `newMux` unsupported endpoint in `internal/gateway/gateway_test.go` — `GET /v1/completions` returns 405 with JSON error body listing supported endpoints
- [x] T060 [P] [US1] Test `newMux` proxy routing in `internal/gateway/gateway_test.go` — `POST /v1/messages` calls provider's `PrepareRequest` and forwards to upstream (use `httptest.NewServer` as mock upstream)
- [x] T061 [P] [US1] Test `newMux` proxy routing for count_tokens in `internal/gateway/gateway_test.go` — `POST /v1/messages/count_tokens` routes correctly to upstream
- [x] T062 [P] [US1] Test upstream error forwarding in `internal/gateway/gateway_test.go` — mock upstream returns 429 with error body, verify gateway forwards status code and body as-is (FR-014)

### Gateway Lifecycle Tests

- [x] T063 [P] [US1] Test `Start` with mock `ListenAndServe` in `internal/gateway/gateway_test.go` — verify provider detection, PID file written, server started on correct port
- [x] T064 [P] [US5] Test `detach` with mock `ExecCmd` in `internal/gateway/gateway_test.go` — verify re-exec includes `_UF_GATEWAY_CHILD=1` env var, verify `--detach` flag removed from child args
- [x] T065 [P] [US5] Test `Stop` in `internal/gateway/gateway_test.go` — PID file exists and process alive: process terminated and PID file removed, PID file exists but process dead: stale PID file removed with message, no PID file: prints "No gateway running."
- [x] T066 [P] [US5] Test `Status` in `internal/gateway/gateway_test.go` — gateway running: displays provider, port, PID, uptime, gateway not running: prints "No gateway running."
- [x] T067 [P] [US4] Test `Start` with `--port` override in `internal/gateway/gateway_test.go` — verify server listens on custom port, PID file records custom port
- [x] T068 [P] [US4] Test `Start` with `--provider` override in `internal/gateway/gateway_test.go` — verify explicit provider used regardless of env vars (FR-009)
- [x] T069 [P] [US1] Test `Start` port conflict in `internal/gateway/gateway_test.go` — mock `ListenAndServe` returns "address already in use" error, verify error message suggests `--port`

### Sandbox Integration Tests

- [x] T070 [P] [US2] Test `gatewayHealthCheck` in `internal/sandbox/sandbox_test.go` — mock HTTP returns 200: returns true, mock HTTP returns error: returns false
- [x] T071 [P] [US2] Test `autoStartGateway` with provider detected in `internal/sandbox/sandbox_test.go` — mock env vars for Anthropic, mock ExecCmd for `uf gateway --detach`, mock health check success, verify returns (port, true, nil)
- [x] T072 [P] [US2] Test `autoStartGateway` with no provider in `internal/sandbox/sandbox_test.go` — no provider env vars set, verify returns (0, false, nil) for fallback
- [x] T073 [P] [US2] Test `autoStartGateway` with existing gateway in `internal/sandbox/sandbox_test.go` — mock health check returns 200 (gateway already running), verify ExecCmd NOT called (reuses existing gateway)
- [x] T074 [P] [US2] Test `gatewayEnvVars` in `internal/sandbox/sandbox_test.go` — verify returns correct `-e ANTHROPIC_BASE_URL=http://host.containers.internal:{port}` and `-e ANTHROPIC_AUTH_TOKEN=gateway` pairs
- [x] T075 [P] [US2] Test `forwardedEnvVars` with gateway active in `internal/sandbox/sandbox_test.go` — verify ANTHROPIC_API_KEY, ANTHROPIC_VERTEX_PROJECT_ID, CLAUDE_CODE_USE_VERTEX are skipped, verify OLLAMA_HOST, OPENAI_API_KEY, GEMINI_API_KEY still forwarded
- [x] T076 [P] [US2] Test `forwardedEnvVars` with gateway inactive in `internal/sandbox/sandbox_test.go` — verify all keys forwarded (backward compatible, identical to pre-gateway behavior)
- [x] T077 [P] [US2] Test `googleCloudCredentialMounts` with gateway active in `internal/sandbox/sandbox_test.go` — verify no gcloud mounts returned when gateway is active
- [x] T078 [P] [US2] Test `googleCloudCredentialMounts` with gateway inactive in `internal/sandbox/sandbox_test.go` — verify gcloud mounts returned as before (backward compatible)
- [x] T079 [P] [US2] Test `buildRunArgs` with gateway active in `internal/sandbox/sandbox_test.go` — verify ANTHROPIC_BASE_URL and ANTHROPIC_AUTH_TOKEN present, verify no credential mounts, verify no ANTHROPIC_API_KEY forwarding
- [x] T080 [P] [US2] Test `buildRunArgs` with gateway inactive in `internal/sandbox/sandbox_test.go` — verify identical output to pre-gateway behavior (backward compatibility regression test)
- [x] T081 [P] [US2] Test `Start()` auto-starts gateway in `internal/sandbox/sandbox_test.go` — mock provider env vars, mock ExecCmd, verify gateway started before container, verify container receives gateway env vars

**Checkpoint**: All tests pass. Coverage targets met per plan.md (gateway core 90%+, providers 85%+, PID file 95%+, sandbox integration 85%+). No regressions in existing sandbox tests.

---

## Phase 6: Polish & Documentation (All US)

**Purpose**: Documentation updates, edge case handling, and final verification.

**Depends on**: Phase 1 + Phase 2 + Phase 3 + Phase 4 + Phase 5

- [x] T082 [US1] Update `AGENTS.md` — add `internal/gateway/` to Project Structure tree, add `cmd/unbound-force/gateway.go` entry, update Active Technologies with `net/http/httputil` (stdlib), add Recent Changes entry for Spec 033
- [x] T083 [P] Update `cmd/unbound-force/main.go` import comment — ensure gateway import is documented in the import block
- [x] T084 [P] Verify `go build ./...` succeeds with no compilation errors
- [x] T085 [P] Verify `go test -race -count=1 ./...` passes all tests including new and existing
- [x] T086 [P] Verify `golangci-lint run` produces no new findings
- [x] T087 Run quickstart.md verification steps 1-4 (build, unit tests, lint, CLI smoke test) to confirm end-to-end functionality

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Gateway Core)**: No dependencies — can start immediately
- **Phase 2 (Providers)**: Depends on Phase 1 (Provider interface used by gateway core)
- **Phase 3 (CLI Commands)**: Depends on Phase 1 + Phase 2 (gateway package must be functional)
- **Phase 4 (Sandbox Integration)**: Depends on Phase 1 + Phase 2 + Phase 3 (gateway must be fully operational)
- **Phase 5 (Tests)**: Depends on Phase 1 + Phase 2 + Phase 3 + Phase 4 (all code must exist)
- **Phase 6 (Polish)**: Depends on all previous phases

### Within-Phase Parallelism

- **Phase 1**: T001, T002 can run in parallel (independent PID functions). T006, T007 can run in parallel with T001/T002 (different concerns in same file). T003-T005 depend on T001/T002. T008-T013 are sequential (build on each other).
- **Phase 2**: T016, T017, T021 can run in parallel (independent provider structs). T025 can run in parallel with providers (independent file). T014 must complete before T015. T018-T020 are sequential within Vertex. T022-T024 are sequential within Bedrock.
- **Phase 3**: T029-T031 can run in parallel (independent Cobra commands). T032 depends on T029.
- **Phase 4**: T033-T034 must complete before T039. T035-T038 can run in parallel. T039 depends on T033-T038. T040 depends on T039.
- **Phase 5**: All test tasks marked [P] can run in parallel (they test independent functions).
- **Phase 6**: T084, T085, T086 can run in parallel. T087 depends on all previous.

### User Story Dependencies

- **US1 (Gateway Starts)**: Phase 1 + Phase 2 + Phase 3 — no dependencies on other stories
- **US2 (Sandbox Auto-Start)**: Phase 4 — depends on US1 being complete
- **US3 (Token Refresh)**: Phase 2 (T025-T027) — can be implemented alongside US1
- **US4 (Provider Override)**: Phase 2 (T015) + Phase 3 (T029 flags) — can be implemented alongside US1
- **US5 (Background Mode)**: Phase 1 (T001-T005, T011-T013) + Phase 3 (T030-T031) — can be implemented alongside US1

---

## Task Summary

| Phase | Tasks | Files |
|-------|-------|-------|
| Phase 1: Gateway Core | T001–T013 (13 tasks) | `internal/gateway/pid.go`, `internal/gateway/gateway.go` |
| Phase 2: Providers | T014–T028 (15 tasks) | `internal/gateway/provider.go`, `internal/gateway/refresh.go` |
| Phase 3: CLI Commands | T029–T032 (4 tasks) | `cmd/unbound-force/gateway.go`, `cmd/unbound-force/main.go` |
| Phase 4: Sandbox Integration | T033–T040 (8 tasks) | `internal/sandbox/sandbox.go`, `internal/sandbox/config.go` |
| Phase 5: Tests | T041–T081 (41 tasks) | `internal/gateway/gateway_test.go`, `internal/sandbox/sandbox_test.go` |
| Phase 6: Polish | T082–T087 (6 tasks) | `AGENTS.md`, build/test/lint verification |
| **Total** | **87 tasks** | |
<!-- scaffolded by uf vdev -->
<!-- spec-review: passed -->
<!-- code-review: passed -->
