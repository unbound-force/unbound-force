# Implementation Plan: Gateway Vertex Translation

**Branch**: `034-gateway-vertex-translation` | **Date**: 2026-04-23 | **Spec**: `specs/034-gateway-vertex-translation/spec.md`
**Input**: Feature specification from `specs/034-gateway-vertex-translation/spec.md`

## Summary

Transform the `uf gateway` from a pure auth-injecting
reverse proxy into a translation-mode proxy for Vertex
AI. The gateway accepts standard Anthropic API requests
from OpenCode and translates them to/from Vertex
rawPredict format: removing `model` from the request
body, injecting `anthropic_version: "vertex-2023-10-16"`,
selecting `rawPredict` vs `streamRawPredict` based on
the `stream` field, stripping SDK-injected headers, and
filtering `vertex_event`/`ping` SSE events from
streaming responses. The implementation extends the
existing `VertexProvider` (Spec 033) with body
transformation in `PrepareRequest()` and adds an SSE
filter via `httputil.ReverseProxy.ModifyResponse`. A
synthetic model catalog provides `/v1/models` for
OpenCode's model picker.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: `net/http`, `net/http/httputil` (stdlib), `github.com/spf13/cobra` (CLI), `github.com/charmbracelet/log` (logging) — all existing, no new dependencies
**Storage**: N/A (no persistent state)
**Testing**: Standard library `testing` package, `net/http/httptest` for HTTP handler tests
**Target Platform**: macOS (primary), Linux (secondary)
**Project Type**: CLI (extends existing `uf gateway` command)
**Performance Goals**: < 5ms request body transformation (SC-004), no perceptible SSE latency (SC-005)
**Constraints**: No new Go module dependencies, backward compatible with Anthropic direct provider
**Scale/Scope**: Extends `internal/gateway/` package (~400 lines new code, ~100 lines modified)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Design Check

| Principle | Status | Evidence |
|-----------|--------|----------|
| **I. Autonomous Collaboration** | ✅ PASS | The gateway operates independently — it translates requests without requiring synchronous interaction with other heroes. No new inter-hero artifacts are introduced. |
| **II. Composability First** | ✅ PASS | Translation is Vertex-provider-specific. The Anthropic direct provider path is unchanged (FR-010). The gateway remains independently usable with any provider. No mandatory dependencies on other heroes. |
| **III. Observable Quality** | ✅ PASS | The `/health` endpoint (Spec 033) already returns machine-parseable JSON. The `/v1/models` endpoint returns structured JSON with capabilities metadata. Error responses use the Anthropic JSON error format. |
| **IV. Testability** | ✅ PASS | All new functions are testable in isolation: `transformVertexBody()` operates on `*http.Request` (injectable), `sseFilterReader` operates on `io.Reader` (no HTTP dependency), model catalog handlers use `httptest.NewRecorder`. Coverage strategy defined in data-model.md. |

### Post-Design Check

| Principle | Status | Evidence |
|-----------|--------|----------|
| **I. Autonomous Collaboration** | ✅ PASS | No changes to artifact formats or inter-hero communication. |
| **II. Composability First** | ✅ PASS | SSE filter is applied only when provider is Vertex. Body transformation is encapsulated in `transformVertexBody()`. No changes to the `Provider` interface. |
| **III. Observable Quality** | ✅ PASS | Model catalog includes `capabilities` metadata for machine consumption. |
| **IV. Testability** | ✅ PASS | `sseFilterReader` is a pure `io.Reader` wrapper — testable with `strings.NewReader` input. `transformVertexBody` is testable with `httptest.NewRequest`. No external services required. Coverage target: ≥ 80% for new code. |

## Project Structure

### Documentation (this feature)

```text
specs/034-gateway-vertex-translation/
├── plan.md              # This file
├── research.md          # Phase 0: SSE filtering, Vertex endpoints, model catalog
├── data-model.md        # Key entities and coverage strategy
├── quickstart.md        # Verification steps
├── contracts/
│   └── gateway-vertex-api.md  # Translation behavior contract
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
internal/gateway/
├── gateway.go           # Modified: ModifyResponse for Vertex, /v1/models handler
├── provider.go          # Modified: transformVertexBody(), updated PrepareRequest()
├── sse.go               # New: sseFilterReader, vertexSSEFilter()
├── refresh.go           # Unchanged
├── pid.go               # Unchanged
├── signal_unix.go       # Unchanged
└── gateway_test.go      # Modified: new tests for translation + SSE filtering

cmd/unbound-force/
├── gateway.go           # Unchanged (no new CLI flags)
└── main.go              # Unchanged

internal/sandbox/
├── sandbox.go           # Unchanged (autoStartGateway already works)
└── config.go            # Unchanged (gatewaySkippedKeys already covers FR-011)
```

**Structure Decision**: All changes are within the
existing `internal/gateway/` package. One new file
(`sse.go`) is added for the SSE filter, following the
established pattern of separating concerns into
focused files (e.g., `refresh.go` for token refresh,
`pid.go` for PID management). No new packages or
directories are needed.

## Implementation Approach

### Phase 1: Request Body Transformation (US1, US2)

Modify `VertexProvider.PrepareRequest()` to call a new
`transformVertexBody()` function that:

1. Reads the request body into `map[string]any`
2. Extracts `model` (for URL path) and `stream` (for
   endpoint selection)
3. Deletes `model` from the map
4. Injects `anthropic_version: "vertex-2023-10-16"` if
   absent
5. Re-encodes the body and updates `ContentLength`
6. Strips `anthropic-beta` and `anthropic-version`
   HTTP headers

The existing `extractModelFromBody()` function is
replaced by `transformVertexBody()` for the Vertex
provider. Anthropic and Bedrock providers continue
using `extractModelFromBody()` unchanged.

**Endpoint selection**: The `stream` flag determines
whether the URL path ends with `:rawPredict` or
`:streamRawPredict`. The `count_tokens` path always
uses `:rawPredict`.

### Phase 2: SSE Response Filtering (US3)

Create `internal/gateway/sse.go` with:

1. `sseFilterReader` — an `io.ReadCloser` wrapper that
   reads SSE events line-by-line, accumulates complete
   events (delimited by blank lines), and drops events
   whose `event:` field matches a filtered set.

2. `vertexSSEFilter()` — returns a `ModifyResponse`
   function that wraps the response body in an
   `sseFilterReader` when `Content-Type` is
   `text/event-stream`.

Integration: In `newMux()`, set
`proxy.ModifyResponse = vertexSSEFilter()` when the
provider is Vertex.

### Phase 3: Model Catalog (US5)

Add `/v1/models` and `/v1/models/{model_id}` handlers
to `newMux()`. The model catalog is a `[]syntheticModel`
slice with 9 entries matching the Vertex AI Claude
model catalog (per research.md R3).

**Note**: The `opsx/gateway-provider-config` change
already added synthetic model endpoints. This phase
verifies the catalog includes `capabilities` metadata
and updates the model list to match the current Vertex
offering.

### Phase 4: Tests and Verification

Add tests for:
- `transformVertexBody` — model extraction, stream
  detection, anthropic_version injection, preservation,
  malformed JSON, empty body
- `sseFilterReader` — vertex_event dropped, ping
  dropped, standard events forwarded, non-streaming
  passthrough, partial events
- End-to-end proxy tests — Vertex request translation
  through `newMux` with mock upstream
- Model catalog — `/v1/models` list, single model,
  404 for unknown
- Backward compatibility — Anthropic provider
  unchanged

## Complexity Tracking

No constitution violations to justify. The
implementation adds ~400 lines of new code and ~100
lines of modifications, well within the scope of a
single package extension.
