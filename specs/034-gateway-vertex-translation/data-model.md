# Data Model: Gateway Vertex Translation

**Branch**: `034-gateway-vertex-translation` | **Date**: 2026-04-23
**Spec**: `specs/034-gateway-vertex-translation/spec.md`

## Entities

### Modified: VertexProvider

The existing `VertexProvider` struct (from Spec 033)
is extended with request body transformation and
streaming endpoint detection. No new fields are added
to the struct — the changes are in `PrepareRequest()`.

```go
// VertexProvider forwards requests to the Vertex AI
// rawPredict/streamRawPredict endpoint with an OAuth
// bearer token and request body translation.
//
// Extended by Spec 034 to:
//   - Remove `model` from request body (FR-001)
//   - Inject `anthropic_version` (FR-002, FR-003)
//   - Strip anthropic-beta/anthropic-version headers (FR-004)
//   - Select rawPredict vs streamRawPredict (FR-005, FR-006)
//   - Filter SSE response events (FR-007, FR-008)
type VertexProvider struct {
    projectID string // from ANTHROPIC_VERTEX_PROJECT_ID
    region    string // from CLOUD_ML_REGION, default "us-east5"
    token     string // current OAuth token
    tokenMu   sync.RWMutex
    cancel    context.CancelFunc
    execCmd   func(string, ...string) ([]byte, error)
    getenv    func(string) string
}
```

### New: transformVertexBody

A new function that performs request body
transformation for Vertex. Returns the extracted model
name and stream flag.

```go
// transformVertexBody reads the request body, removes
// the `model` field, injects `anthropic_version` if
// absent, and replaces the body. Returns the extracted
// model name and stream flag.
//
// Uses map[string]any to preserve all unknown fields
// without requiring a struct definition for the full
// Anthropic Messages API body (per research.md R7).
func transformVertexBody(req *http.Request) (model string, stream bool, err error) {
    defaultModel := "claude-sonnet-4-20250514"

    if req.Body == nil {
        return defaultModel, false, nil
    }

    body, err := io.ReadAll(req.Body)
    if err != nil {
        return defaultModel, false, nil
    }

    var payload map[string]any
    if err := json.Unmarshal(body, &payload); err != nil {
        // Malformed JSON — forward unchanged (edge case).
        req.Body = io.NopCloser(bytes.NewReader(body))
        return defaultModel, false, nil
    }

    // Extract model (FR-001).
    if m, ok := payload["model"].(string); ok && m != "" {
        model = m
    } else {
        model = defaultModel
    }
    delete(payload, "model")

    // Extract stream flag (FR-005, FR-006).
    if s, ok := payload["stream"].(bool); ok {
        stream = s
    }

    // Inject anthropic_version if absent (FR-002, FR-003).
    if _, ok := payload["anthropic_version"]; !ok {
        payload["anthropic_version"] = "vertex-2023-10-16"
    }

    // Re-encode.
    newBody, err := json.Marshal(payload)
    if err != nil {
        // Marshal failed — forward original body.
        req.Body = io.NopCloser(bytes.NewReader(body))
        return model, stream, nil
    }

    req.Body = io.NopCloser(bytes.NewReader(newBody))
    req.ContentLength = int64(len(newBody))
    return model, stream, nil
}
```

### New: sseFilterReader

An `io.ReadCloser` that wraps an upstream response
body and drops SSE events matching filtered event
types.

```go
// sseFilterReader wraps an io.ReadCloser and drops
// SSE events whose `event:` field matches a filtered
// type. Used to remove `vertex_event` and `ping`
// events from Vertex streaming responses (FR-007,
// FR-008).
//
// Design decision: Implemented as an io.ReadCloser
// wrapper applied in ModifyResponse, keeping the
// filter composable and testable — it operates on
// io.Reader without knowledge of HTTP (per
// research.md R1).
type sseFilterReader struct {
    source     io.ReadCloser
    scanner    *bufio.Scanner
    buf        bytes.Buffer
    filtered   map[string]bool
    done       bool
}

// newSSEFilterReader creates a filter that drops
// events with the given event types.
func newSSEFilterReader(source io.ReadCloser, filtered map[string]bool) *sseFilterReader {
    s := bufio.NewScanner(source)
    // Increase buffer to 1MB to handle large SSE
    // data lines (content blocks can be large).
    s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
    return &sseFilterReader{
        source:   source,
        scanner:  s,
        filtered: filtered,
    }
}

func (r *sseFilterReader) Read(p []byte) (int, error)
func (r *sseFilterReader) Close() error
```

### New: vertexSSEFilter (ModifyResponse hook)

A function that wraps the response body in an
`sseFilterReader` when the response is a streaming
SSE response from Vertex.

```go
// vertexSSEFilter returns a ModifyResponse function
// for httputil.ReverseProxy that filters Vertex-
// specific SSE events from streaming responses.
//
// Only applied when Content-Type is
// text/event-stream (FR-009 — non-streaming
// responses pass through unchanged).
func vertexSSEFilter() func(*http.Response) error {
    return func(resp *http.Response) error {
        ct := resp.Header.Get("Content-Type")
        if !strings.HasPrefix(ct, "text/event-stream") {
            return nil // Non-streaming — pass through.
        }

        filtered := map[string]bool{
            "vertex_event": true,
            "ping":         true,
        }
        resp.Body = newSSEFilterReader(resp.Body, filtered)
        // Remove Content-Length since filtering changes
        // the body size (chunked encoding handles this).
        resp.Header.Del("Content-Length")
        resp.ContentLength = -1
        return nil
    }
}
```

### Modified: newMux (gateway.go)

The `newMux` function is extended to set
`ModifyResponse` on the `ReverseProxy` when the
provider is Vertex. This is the integration point
for the SSE filter.

```go
// In newMux(), after creating the proxy:
if provider.Name() == "vertex" {
    proxy.ModifyResponse = vertexSSEFilter()
}
```

### New: Synthetic Model Catalog

A hardcoded catalog of Claude models available on
Vertex AI, used by the `/v1/models` endpoint.

```go
// syntheticModel represents a Claude model available
// on Vertex AI. Used for the /v1/models endpoint
// since Vertex has no model listing API.
type syntheticModel struct {
    ID              string `json:"id"`
    Type            string `json:"type"`
    DisplayName     string `json:"display_name"`
    CreatedAt       string `json:"created_at"`
    MaxInputTokens  int    `json:"max_input_tokens"`
    MaxTokens       int    `json:"max_tokens"`
    Capabilities    modelCapabilities `json:"capabilities"`
}

type modelCapabilities struct {
    Vision           bool `json:"vision"`
    ExtendedThinking bool `json:"extended_thinking"`
    PDFInput         bool `json:"pdf_input"`
}

// knownModels is the catalog of Claude models
// available on Vertex AI. Updated when Anthropic
// releases new models on Vertex.
var knownModels = []syntheticModel{
    {
        ID:             "claude-opus-4-7-20250416",
        Type:           "model",
        DisplayName:    "Claude Opus 4.7",
        CreatedAt:      "2025-04-16T00:00:00Z",
        MaxInputTokens: 200000,
        MaxTokens:      32768,
        Capabilities:   modelCapabilities{Vision: true, ExtendedThinking: true, PDFInput: true},
    },
    {
        ID:             "claude-sonnet-4-6-20250217",
        Type:           "model",
        DisplayName:    "Claude Sonnet 4.6",
        CreatedAt:      "2025-02-17T00:00:00Z",
        MaxInputTokens: 200000,
        MaxTokens:      64000,
        Capabilities:   modelCapabilities{Vision: true, ExtendedThinking: true, PDFInput: true},
    },
    // ... 7 more entries (see spec US5 for full list)
}
```

### Modified: Sandbox Environment (config.go)

The existing `gatewaySkippedKeys` map already
prevents Vertex-specific env vars from leaking into
the container (FR-011). No changes needed — the
current implementation already satisfies FR-011 and
FR-012.

```go
// Already implemented in Spec 033:
var gatewaySkippedKeys = map[string]bool{
    "ANTHROPIC_API_KEY":            true,
    "ANTHROPIC_VERTEX_PROJECT_ID":  true,
    "CLAUDE_CODE_USE_VERTEX":       true,
    "GOOGLE_CLOUD_PROJECT":         true,
    "VERTEX_LOCATION":              true,
}
```

**Verification**: FR-011 requires that
`CLAUDE_CODE_USE_VERTEX`,
`ANTHROPIC_VERTEX_PROJECT_ID`, and `VERTEX_LOCATION`
are not forwarded. All three are already in
`gatewaySkippedKeys`. No code changes needed for
US4.

## Entity Relationships

```text
Gateway (1) ──uses──> VertexProvider (1)
  │                      │
  │                      ├── transformVertexBody()
  │                      │     ├── removes model
  │                      │     ├── injects anthropic_version
  │                      │     └── detects stream flag
  │                      │
  │                      └── PrepareRequest()
  │                            ├── builds rawPredict/streamRawPredict URL
  │                            ├── strips anthropic-beta/version headers
  │                            └── injects Authorization: Bearer
  │
  ├── ReverseProxy.ModifyResponse
  │     └── vertexSSEFilter()
  │           └── sseFilterReader
  │                 ├── drops vertex_event
  │                 └── drops ping
  │
  └── /v1/models handler
        └── knownModels catalog
```

## File Ownership

| File | Change | Rationale |
|------|--------|-----------|
| `internal/gateway/provider.go` | Modified | Add `transformVertexBody()`, update `VertexProvider.PrepareRequest()` |
| `internal/gateway/gateway.go` | Modified | Add `ModifyResponse` to proxy for Vertex, add `/v1/models` handler, update model catalog |
| `internal/gateway/sse.go` | New | `sseFilterReader`, `newSSEFilterReader`, `vertexSSEFilter()` |
| `internal/gateway/gateway_test.go` | Modified | Add tests for body transformation, SSE filtering, streaming endpoint, model catalog |
| `internal/sandbox/config.go` | Unchanged | `gatewaySkippedKeys` already covers FR-011 |
| `internal/sandbox/sandbox.go` | Unchanged | `autoStartGateway()` already handles gateway lifecycle |

## Coverage Strategy

Per Constitution Principle IV (Testability):

- **Unit tests**: All new functions
  (`transformVertexBody`, `sseFilterReader.Read`,
  `vertexSSEFilter`, model catalog handlers) tested
  in isolation with injected dependencies.
- **Contract tests**: HTTP handler tests using
  `httptest.NewServer` to verify end-to-end request
  transformation and SSE filtering through the proxy.
- **Edge case tests**: Malformed JSON body, empty body,
  missing `model` field, partial SSE events, non-
  streaming responses, `anthropic_version` already
  present.
- **Regression tests**: Verify Anthropic provider path
  remains unchanged (FR-010 backward compatibility).
- **Coverage target**: ≥ 80% line coverage for new code
  in `provider.go`, `sse.go`, and `gateway.go`.
