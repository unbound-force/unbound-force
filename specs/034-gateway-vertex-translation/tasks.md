# Tasks: Gateway Vertex Translation

**Input**: Design documents from `specs/034-gateway-vertex-translation/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/gateway-vertex-api.md, quickstart.md

**Organization**: Tasks are grouped by implementation phase, mapped to user stories by priority. Each phase has a checkpoint gate.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Verification)

**Purpose**: Verify prerequisites and confirm existing code state before making changes. No code modifications.

- [x] T001 Verify `internal/gateway/provider.go` exists with `VertexProvider`, `extractModelFromBody()`, and `PrepareRequest()` from Spec 033
- [x] T002 [P] Verify `internal/gateway/gateway.go` exists with `newMux()`, `ReverseProxy`, and `writeJSONError()` from Spec 033
- [x] T003 [P] Verify `internal/sandbox/config.go` has `gatewaySkippedKeys` map containing `CLAUDE_CODE_USE_VERTEX`, `ANTHROPIC_VERTEX_PROJECT_ID`, and `VERTEX_LOCATION` (confirms FR-011 is already satisfied)
- [x] T004 [P] Verify `internal/gateway/gateway_test.go` exists and all existing tests pass: `go test -race -count=1 ./internal/gateway/...`
- [x] T005 [P] Verify `make check` passes (build, test, vet, lint) to establish clean baseline

**Checkpoint**: Existing codebase is verified. All Spec 033 gateway functionality is intact. Proceed to foundational work.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Create the core body transformation function and SSE filter infrastructure that ALL user stories depend on. No integration yet — these are standalone, testable units.

**⚠️ CRITICAL**: No user story integration (Phase 3+) can begin until this phase is complete.

### Body Transformation Function

- [x] T006 [US1] Create `transformVertexBody(req *http.Request) (model string, stream bool, err error)` in `internal/gateway/provider.go` — reads body into `map[string]any`, extracts `model` (default `claude-sonnet-4-20250514`), extracts `stream` flag (default `false`), deletes `model` from map, injects `anthropic_version: "vertex-2023-10-16"` if absent (FR-002), preserves existing `anthropic_version` (FR-003), re-encodes body, updates `req.ContentLength` (FR-014). Returns extracted model and stream flag. Per research.md R7.
- [x] T007 [US1] Add test `TestTransformVertexBody_RemovesModel` in `internal/gateway/gateway_test.go` — verify `model` field is removed from body and returned as first return value
- [x] T008 [P] [US1] Add test `TestTransformVertexBody_InjectsAnthropicVersion` in `internal/gateway/gateway_test.go` — verify `anthropic_version: "vertex-2023-10-16"` is injected when absent
- [x] T009 [P] [US1] Add test `TestTransformVertexBody_PreservesExistingAnthropicVersion` in `internal/gateway/gateway_test.go` — verify existing `anthropic_version` value is not overwritten (FR-003)
- [x] T010 [P] [US2] Add test `TestTransformVertexBody_DetectsStreamTrue` in `internal/gateway/gateway_test.go` — verify `stream` return value is `true` when body contains `"stream": true`
- [x] T011 [P] [US2] Add test `TestTransformVertexBody_DetectsStreamFalse` in `internal/gateway/gateway_test.go` — verify `stream` return value is `false` when body contains `"stream": false` or field is absent
- [x] T012 [P] [US1] Add test `TestTransformVertexBody_MalformedJSON` in `internal/gateway/gateway_test.go` — verify malformed JSON returns default model, `stream=false`, and forwards original body unchanged
- [x] T013 [P] [US1] Add test `TestTransformVertexBody_NilBody` in `internal/gateway/gateway_test.go` — verify nil body returns default model, `stream=false`, no error
- [x] T014 [P] [US1] Add test `TestTransformVertexBody_EmptyBody` in `internal/gateway/gateway_test.go` — verify empty body returns default model, `stream=false`
- [x] T015 [US1] Add test `TestTransformVertexBody_UpdatesContentLength` in `internal/gateway/gateway_test.go` — verify `req.ContentLength` matches the re-encoded body length after transformation (FR-014)
- [x] T016 [US1] Add test `TestTransformVertexBody_PreservesOtherFields` in `internal/gateway/gateway_test.go` — verify `messages`, `max_tokens`, `temperature`, `tools`, `system` fields pass through unchanged

### SSE Filter Infrastructure

- [x] T017 [US3] Create `internal/gateway/sse.go` with `sseFilterReader` struct (fields: `source io.ReadCloser`, `scanner *bufio.Scanner`, `buf bytes.Buffer`, `filtered map[string]bool`, `done bool`), `newSSEFilterReader(source io.ReadCloser, filtered map[string]bool) *sseFilterReader` constructor (sets scanner buffer to 1MB per research.md R1), `Read(p []byte) (int, error)` method (accumulates lines until blank line event boundary, drops events matching filtered set, forwards others), `Close() error` method (delegates to source.Close)
- [x] T018 [US3] Create `vertexSSEFilter() func(*http.Response) error` in `internal/gateway/sse.go` — returns a `ModifyResponse` function that wraps `resp.Body` in `sseFilterReader` when `Content-Type` starts with `text/event-stream`, sets filtered types to `{"vertex_event": true, "ping": true}`, removes `Content-Length` header and sets `resp.ContentLength = -1` (chunked encoding). Non-streaming responses pass through unchanged (FR-009).
- [x] T019 [US3] Add test `TestSSEFilterReader_DropsVertexEvent` in `internal/gateway/gateway_test.go` — construct SSE stream with `event: vertex_event\ndata: {...}\n\n`, verify it is dropped
- [x] T020 [P] [US3] Add test `TestSSEFilterReader_DropsPing` in `internal/gateway/gateway_test.go` — construct SSE stream with `event: ping\ndata: \n\n`, verify it is dropped (FR-008)
- [x] T021 [P] [US3] Add test `TestSSEFilterReader_ForwardsStandardEvents` in `internal/gateway/gateway_test.go` — verify `message_start`, `content_block_delta`, `content_block_stop`, `message_delta`, `message_stop` events pass through unchanged (SC-002)
- [x] T022 [P] [US3] Add test `TestSSEFilterReader_MixedEvents` in `internal/gateway/gateway_test.go` — construct stream with interleaved standard and filtered events, verify only standard events remain in output
- [x] T023 [P] [US3] Add test `TestSSEFilterReader_EmptyStream` in `internal/gateway/gateway_test.go` — verify empty reader returns `io.EOF` immediately
- [x] T024 [US3] Add test `TestSSEFilterReader_Close` in `internal/gateway/gateway_test.go` — verify `Close()` delegates to the underlying source's `Close()`
- [x] T025 [US3] Add test `TestVertexSSEFilter_NonStreamingPassthrough` in `internal/gateway/gateway_test.go` — verify `vertexSSEFilter()` does not wrap body when `Content-Type` is `application/json` (FR-009)
- [x] T026 [US3] Add test `TestVertexSSEFilter_StreamingWraps` in `internal/gateway/gateway_test.go` — verify `vertexSSEFilter()` wraps body when `Content-Type` is `text/event-stream` and removes `Content-Length` header

**Checkpoint**: `transformVertexBody` and `sseFilterReader` are implemented and fully tested in isolation. Run `go test -race -count=1 ./internal/gateway/...` — all tests must pass. Proceed to user story integration.

---

## Phase 3: US1 + US2 — Request Translation (Priority: P1) 🎯 MVP

**Goal**: Integrate `transformVertexBody` into `VertexProvider.PrepareRequest()` so the gateway translates standard Anthropic requests to Vertex rawPredict format. Select `rawPredict` vs `streamRawPredict` based on the `stream` field.

**Independent Test**: Send a standard Anthropic `/v1/messages` POST to the gateway running in Vertex mode. Verify the upstream request has `model` removed from the body, `anthropic_version` injected, headers stripped, and the URL targets the correct rawPredict/streamRawPredict endpoint.

### Implementation

- [x] T027 [US1] Replace `extractModelFromBody(req)` call with `transformVertexBody(req)` in `VertexProvider.PrepareRequest()` in `internal/gateway/provider.go` — use returned `model` for URL path, use returned `stream` for endpoint selection. Keep `extractModelFromBody()` unchanged for Anthropic and Bedrock providers.
- [x] T028 [US2] Update URL path construction in `VertexProvider.PrepareRequest()` in `internal/gateway/provider.go` — use `:streamRawPredict` when `stream=true`, `:rawPredict` when `stream=false` or for `count_tokens` path (FR-005, FR-006). Detect `count_tokens` by checking if the original `req.URL.Path` contains `count_tokens`.
- [x] T029 [US1] Add `anthropic-beta` and `anthropic-version` header stripping to `VertexProvider.PrepareRequest()` in `internal/gateway/provider.go` — call `req.Header.Del("anthropic-beta")` and `req.Header.Del("anthropic-version")` after body transformation (FR-004)
- [x] T030 [US1] Update existing test `TestVertexProvider_PrepareRequest` in `internal/gateway/gateway_test.go` — verify `model` is removed from body, `anthropic_version` is injected, `anthropic-beta` and `anthropic-version` headers are stripped, URL contains `rawPredict`
- [x] T031 [US2] Add test `TestVertexProvider_PrepareRequest_StreamingEndpoint` in `internal/gateway/gateway_test.go` — send request with `"stream": true` in body, verify URL path ends with `:streamRawPredict`
- [x] T032 [US2] Add test `TestVertexProvider_PrepareRequest_NonStreamingEndpoint` in `internal/gateway/gateway_test.go` — send request with `"stream": false` in body, verify URL path ends with `:rawPredict`
- [x] T033 [US2] Add test `TestVertexProvider_PrepareRequest_CountTokensAlwaysRawPredict` in `internal/gateway/gateway_test.go` — send request to `/v1/messages/count_tokens` with `"stream": true`, verify URL path still ends with `:rawPredict` (count_tokens never streams)
- [x] T034 [US1] Add test `TestVertexProvider_PrepareRequest_HeaderStripping` in `internal/gateway/gateway_test.go` — send request with `anthropic-beta` and `anthropic-version` headers, verify both are removed after PrepareRequest (FR-004)
- [x] T035 [US1] Add test `TestVertexProvider_PrepareRequest_PreservesOtherHeaders` in `internal/gateway/gateway_test.go` — verify `Content-Type` and `X-Claude-Code-Session-Id` headers are preserved after PrepareRequest

### End-to-End Proxy Tests

- [x] T036 [US1] Add test `TestNewMux_VertexProxyTranslation` in `internal/gateway/gateway_test.go` — create mock upstream server, create `VertexProvider` with mock token, create `newMux`, send standard Anthropic request, verify upstream receives transformed body (no `model`, has `anthropic_version`) and stripped headers
- [x] T037 [US2] Add test `TestNewMux_VertexStreamingEndpoint` in `internal/gateway/gateway_test.go` — send request with `"stream": true` through the proxy, verify the upstream URL path contains `streamRawPredict`

### Backward Compatibility

- [x] T038 [US1] Add test `TestAnthropicProvider_NoBodyTransformation` in `internal/gateway/gateway_test.go` — verify Anthropic provider does NOT call `transformVertexBody`, body passes through unchanged, `anthropic-beta` and `anthropic-version` headers are preserved (FR-010, SC-007)
- [x] T039 [US1] Add test `TestBedrockProvider_NoBodyTransformation` in `internal/gateway/gateway_test.go` — verify Bedrock provider still uses `extractModelFromBody()`, body passes through unchanged (FR-010)

**Checkpoint**: Request translation is complete. `uf gateway` with Vertex credentials translates requests correctly. Run `go test -race -count=1 ./internal/gateway/...` — all tests must pass. Run `make check` for full validation.

---

## Phase 4: US3 — SSE Response Filtering (Priority: P1)

**Goal**: Integrate the `sseFilterReader` into the gateway's reverse proxy via `ModifyResponse` so `vertex_event` and `ping` SSE events are dropped from Vertex streaming responses.

**Independent Test**: Construct a mock SSE stream containing `message_start`, `content_block_delta`, `vertex_event`, `ping`, and `message_stop` events. Pass it through the gateway. Verify `vertex_event` and `ping` are dropped and all standard events pass through unchanged.

### Implementation

- [x] T040 [US3] Set `proxy.ModifyResponse = vertexSSEFilter()` in `newMux()` in `internal/gateway/gateway.go` — add conditional: only set when `provider.Name() == "vertex"`. Place after proxy creation, before route registration.
- [x] T041 [US3] Update catch-all route message in `newMux()` in `internal/gateway/gateway.go` — add `/v1/models` to the supported endpoints list in the 405 error message (preparation for Phase 6)

### Tests

- [x] T042 [US3] Add test `TestNewMux_VertexSSEFiltering` in `internal/gateway/gateway_test.go` — create mock upstream that returns `text/event-stream` with `vertex_event` and standard events, create Vertex provider mux, verify `vertex_event` events are dropped from the response body
- [x] T043 [US3] Add test `TestNewMux_VertexNonStreamingNoFilter` in `internal/gateway/gateway_test.go` — create mock upstream that returns `application/json`, verify response passes through unchanged (FR-009)
- [x] T044 [US3] Add test `TestNewMux_AnthropicNoSSEFilter` in `internal/gateway/gateway_test.go` — create Anthropic provider mux with SSE upstream, verify `vertex_event` events are NOT filtered (FR-010 — filter is Vertex-only)
- [x] T045 [US3] Add test `TestNewMux_VertexErrorResponseNoFilter` in `internal/gateway/gateway_test.go` — create mock upstream returning 400 error, verify error response passes through unchanged

**Checkpoint**: SSE response filtering is complete. Vertex streaming responses have `vertex_event` and `ping` events dropped. Run `go test -race -count=1 ./internal/gateway/...` — all tests must pass.

---

## Phase 5: US4 — Sandbox Provider Isolation (Priority: P2)

**Goal**: Verify and ensure the sandbox container does not receive Vertex-specific environment variables when the gateway is active. The `gatewaySkippedKeys` map in `internal/sandbox/config.go` already covers this (confirmed in T003), so this phase is primarily verification with targeted tests.

**Independent Test**: Start a sandbox with Vertex credentials on the host. Verify the container environment contains only `ANTHROPIC_BASE_URL` and `ANTHROPIC_AUTH_TOKEN=gateway`. Verify no `CLAUDE_CODE_USE_VERTEX`, `ANTHROPIC_VERTEX_PROJECT_ID`, or `VERTEX_LOCATION` vars are present.

### Verification Tests

- [x] T046 [US4] Add test `TestForwardedEnvVars_GatewayActiveSkipsVertexVars` in `internal/sandbox/sandbox_test.go` — call `forwardedEnvVars(opts, true)` with Vertex env vars set, verify `CLAUDE_CODE_USE_VERTEX`, `ANTHROPIC_VERTEX_PROJECT_ID`, `VERTEX_LOCATION`, and `GOOGLE_CLOUD_PROJECT` are NOT in the output (FR-011)
- [x] T047 [P] [US4] Add test `TestForwardedEnvVars_GatewayInactiveForwardsVertexVars` in `internal/sandbox/sandbox_test.go` — call `forwardedEnvVars(opts, false)` with Vertex env vars set, verify they ARE forwarded (backward compatible, FR-012)
- [x] T048 [P] [US4] Add test `TestGatewayEnvVars_ContainsBaseURLAndToken` in `internal/sandbox/sandbox_test.go` — call `gatewayEnvVars(53147)`, verify output contains `ANTHROPIC_BASE_URL=http://host.containers.internal:53147` and `ANTHROPIC_AUTH_TOKEN=gateway`
- [x] T049 [US4] Add test `TestBuildRunArgs_GatewayActiveNoCredentialMounts` in `internal/sandbox/sandbox_test.go` — call `buildRunArgs(opts, platform, true, 53147)`, verify no Google Cloud credential mounts are present and gateway env vars are included

**Checkpoint**: Sandbox provider isolation is verified. Run `go test -race -count=1 ./internal/sandbox/...` — all tests must pass.

---

## Phase 6: US5 — Synthetic Model Catalog (Priority: P3)

**Goal**: Update the `/v1/models` endpoint to return a catalog of Claude models available on Vertex AI with capabilities metadata. Add `/v1/models/{model_id}` for single-model lookup.

**Independent Test**: Send `GET /v1/models` to the gateway. Verify the response contains at least 9 models with `capabilities` metadata. Send `GET /v1/models/unknown-model` and verify 404.

### Implementation

- [x] T050 [US5] Define `syntheticModel` struct and `modelCapabilities` struct in `internal/gateway/gateway.go` — fields per data-model.md: `ID`, `Type`, `DisplayName`, `CreatedAt`, `MaxInputTokens`, `MaxTokens`, `Capabilities` (with `Vision`, `ExtendedThinking`, `PDFInput` booleans)
- [x] T051 [US5] Define `knownModels` slice (`var knownModels = []syntheticModel{...}`) in `internal/gateway/gateway.go` — populate with 9 models from research.md R3: claude-opus-4-7-20250416, claude-sonnet-4-6-20250217, claude-opus-4-6-20250205, claude-opus-4-5-20241124, claude-sonnet-4-5-20241022, claude-opus-4-1-20250414, claude-haiku-4-5-20241022, claude-opus-4-20250514, claude-sonnet-4-20250514. All have vision=true, pdf_input=true. All have extended_thinking=true except Haiku 4.5.
- [x] T052 [US5] Add `GET /v1/models` handler in `newMux()` in `internal/gateway/gateway.go` — returns JSON `{"data": [...], "has_more": false, "first_id": "...", "last_id": "..."}` per contracts/gateway-vertex-api.md
- [x] T053 [US5] Add `GET /v1/models/{model_id}` handler in `newMux()` in `internal/gateway/gateway.go` — returns single model JSON on match, 404 JSON error `{"error": {"type": "not_found", "message": "Model '...' not found"}}` on miss. Use `strings.TrimPrefix(req.URL.Path, "/v1/models/")` to extract model ID.

### Tests

- [x] T054 [US5] Add test `TestNewMux_ModelsList` in `internal/gateway/gateway_test.go` — send `GET /v1/models`, verify 200 status, JSON response contains `data` array with at least 9 entries, each with `id`, `type`, `display_name`, `capabilities`
- [x] T055 [P] [US5] Add test `TestNewMux_ModelsSingleFound` in `internal/gateway/gateway_test.go` — send `GET /v1/models/claude-sonnet-4-20250514`, verify 200 status and correct model returned
- [x] T056 [P] [US5] Add test `TestNewMux_ModelsSingleNotFound` in `internal/gateway/gateway_test.go` — send `GET /v1/models/unknown-model`, verify 404 status and error JSON
- [x] T057 [US5] Add test `TestNewMux_ModelsCapabilities` in `internal/gateway/gateway_test.go` — verify Haiku 4.5 has `extended_thinking: false` while Opus 4.7 has `extended_thinking: true`

**Checkpoint**: Model catalog is complete. Run `go test -race -count=1 ./internal/gateway/...` — all tests must pass.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Scaffold asset synchronization, documentation updates, and final validation.

### Scaffold Asset Sync

- [x] T058 Synchronize `internal/scaffold/assets/opencode/agents/cobalt-crush-dev.md` with live `.opencode/agents/cobalt-crush-dev.md` — ensure scaffold copy matches the live file (standard drift prevention)

### Documentation

- [x] T059 [P] Update `AGENTS.md` Recent Changes section — add entry for `034-gateway-vertex-translation` summarizing: extended `uf gateway` Vertex provider with request body translation (model removal, anthropic_version injection, header stripping), streaming endpoint detection (rawPredict vs streamRawPredict), SSE response filtering (vertex_event and ping event dropping), synthetic model catalog with capabilities, sandbox provider isolation verification. List new file `internal/gateway/sse.go`, modified files `provider.go`, `gateway.go`, `gateway_test.go`. List user story count, task count.
- [x] T060 [P] Update `AGENTS.md` Active Technologies section — no new dependencies (all stdlib). Verify existing `net/http`, `net/http/httputil` entries cover this spec.

### Final Validation

- [x] T061 Run `go test -race -count=1 ./internal/gateway/...` — all gateway tests pass
- [x] T062 [P] Run `go test -race -count=1 ./internal/sandbox/...` — all sandbox tests pass
- [x] T063 Run `make check` — full build, test, vet, lint passes
- [x] T064 Run CI-equivalent checks from `.github/workflows/` — verify all CI commands pass locally (CI Parity Gate per AGENTS.md)
- [x] T065 Run quickstart.md verification steps 1-3 (build, test, model catalog) — confirm spec success criteria SC-001 through SC-007

**Checkpoint**: All tests pass, documentation is updated, scaffold assets are synced. Implementation is complete.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — verification only, can start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 — creates `transformVertexBody` and `sseFilterReader` that ALL subsequent phases depend on
- **Phase 3 (US1+US2)**: Depends on Phase 2 — integrates `transformVertexBody` into `VertexProvider.PrepareRequest()`
- **Phase 4 (US3)**: Depends on Phase 2 — integrates `sseFilterReader` into `newMux()` via `ModifyResponse`
- **Phase 5 (US4)**: Independent of Phases 3-4 — can run after Phase 2 (only touches `internal/sandbox/`)
- **Phase 6 (US5)**: Independent of Phases 3-5 — can run after Phase 2 (only adds handlers to `gateway.go`)
- **Phase 7 (Polish)**: Depends on all previous phases

### Parallel Opportunities

After Phase 2 completes, the following can proceed in parallel:

```text
Phase 2 complete
  ├── Phase 3 (US1+US2) — touches provider.go, gateway_test.go
  ├── Phase 4 (US3)     — touches gateway.go, gateway_test.go
  ├── Phase 5 (US4)     — touches sandbox_test.go only (different package)
  └── Phase 6 (US5)     — touches gateway.go, gateway_test.go
```

**Note**: Phases 3, 4, and 6 all modify `gateway_test.go` and some modify `gateway.go`, so they MUST be sequential if executed by a single worker. Phase 5 touches a different package (`internal/sandbox/`) and CAN run in parallel with any other phase.

### Recommended Sequential Order (Single Worker)

```text
Phase 1 → Phase 2 → Phase 3 → Phase 4 → Phase 6 → Phase 5 → Phase 7
```

Phase 5 is placed after Phase 6 because it is lower risk (verification-only) and can serve as a natural break point.

### Within Each Phase

- Tests marked [P] within a phase can be written in parallel (they test independent behaviors)
- Implementation tasks within a phase are sequential unless marked [P]
- Each phase ends with a checkpoint — run tests before proceeding

---

## Task-to-FR Traceability

| FR | Description | Tasks |
|----|-------------|-------|
| FR-001 | Remove `model` from body | T006, T007, T027, T030 |
| FR-002 | Inject `anthropic_version` | T006, T008, T030 |
| FR-003 | Preserve existing `anthropic_version` | T006, T009 |
| FR-004 | Strip `anthropic-beta`/`anthropic-version` headers | T029, T034 |
| FR-005 | Use `streamRawPredict` for streaming | T028, T031 |
| FR-006 | Use `rawPredict` for non-streaming | T028, T032, T033 |
| FR-007 | Filter `vertex_event` SSE events | T017, T019, T042 |
| FR-008 | Filter `ping` SSE events | T017, T020 |
| FR-009 | No filtering for non-streaming | T018, T025, T043 |
| FR-010 | No translation for Anthropic provider | T038, T039, T044 |
| FR-011 | No Vertex env vars in container | T003, T046 |
| FR-012 | Only gateway env vars in container | T048 |
| FR-013 | Model catalog with capabilities | T050-T053, T054-T057 |
| FR-014 | Update Content-Length after body mod | T006, T015 |
| FR-015 | Handle partial SSE events | T017 (scanner buffering) |

## Task-to-SC Traceability

| SC | Description | Verification Tasks |
|----|-------------|-------------------|
| SC-001 | Sandbox chat works with Vertex | T065 |
| SC-002 | Standard SSE events pass through | T021, T042 |
| SC-003 | vertex_event/ping dropped | T019, T020, T042 |
| SC-004 | < 5ms body transformation | T015 (ContentLength validates fast path) |
| SC-005 | No perceptible SSE latency | T017 (streaming design), T065 |
| SC-006 | No Vertex vars in container | T046, T048 |
| SC-007 | Anthropic path unchanged | T038, T039, T044 |

<!-- spec-review: passed -->
<!-- code-review: passed -->
