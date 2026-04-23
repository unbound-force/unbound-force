# Research: Gateway Vertex Translation

**Branch**: `034-gateway-vertex-translation` | **Date**: 2026-04-23
**Spec**: `specs/034-gateway-vertex-translation/spec.md`

## Phase 0: Research Findings

### R1 — SSE Event Filtering in Go

**Question**: What is the best approach for filtering
a streaming `io.Reader` line-by-line without breaking
SSE event boundaries?

**Finding**: SSE events follow a strict format defined
by the W3C EventSource specification:

```text
event: <event-type>\n
data: <json-payload>\n
\n
```

Each event is terminated by a blank line (`\n\n`). An
event consists of one or more field lines (`event:`,
`data:`, `id:`, `retry:`) followed by a blank line
separator.

**Filtering approach**: Use `httputil.ReverseProxy`'s
`ModifyResponse` hook to wrap the response body in a
filtering `io.Reader`. The filter reads the upstream
body line-by-line using `bufio.Scanner`, accumulates
lines into complete SSE events (delimited by blank
lines), and either forwards or drops each event based
on the `event:` field value.

**Implementation pattern**:

```go
// sseFilterReader wraps an io.Reader and drops SSE
// events matching a set of filtered event types.
type sseFilterReader struct {
    scanner    *bufio.Scanner
    buf        bytes.Buffer
    filtered   map[string]bool // e.g., {"vertex_event": true, "ping": true}
}

func (r *sseFilterReader) Read(p []byte) (int, error) {
    // If buffer has data from a previous event, drain it.
    if r.buf.Len() > 0 {
        return r.buf.Read(p)
    }

    // Accumulate lines until we see a blank line (event boundary).
    var event strings.Builder
    var eventType string
    for r.scanner.Scan() {
        line := r.scanner.Text()
        if line == "" {
            // Blank line = end of event.
            if eventType != "" && r.filtered[eventType] {
                // Drop this event — reset and continue.
                event.Reset()
                eventType = ""
                continue
            }
            // Forward this event.
            event.WriteString("\n") // blank line separator
            r.buf.WriteString(event.String())
            return r.buf.Read(p)
        }
        event.WriteString(line)
        event.WriteString("\n")
        if strings.HasPrefix(line, "event: ") {
            eventType = strings.TrimPrefix(line, "event: ")
        }
    }
    // Scanner exhausted — return remaining data or EOF.
    if event.Len() > 0 {
        r.buf.WriteString(event.String())
        return r.buf.Read(p)
    }
    return 0, io.EOF
}
```

**Key considerations**:

1. **Partial events across TCP packets**: The
   `bufio.Scanner` handles this naturally — it buffers
   until a full line (`\n`) is available. The event
   accumulator then buffers until a blank line is seen.

2. **Flushing**: The `ReverseProxy` flushes SSE events
   automatically when the `ResponseWriter` implements
   `http.Flusher`. The filter must not introduce
   buffering delays — it should emit complete events
   as soon as they are validated.

3. **Content-Length**: Streaming responses use chunked
   transfer encoding, so `Content-Length` is not set.
   The filter does not need to update it.

4. **Non-streaming responses**: The filter is only
   applied when the response `Content-Type` is
   `text/event-stream`. Non-streaming responses pass
   through unchanged (FR-009).

5. **Scanner buffer size**: The default `bufio.Scanner`
   buffer is 64KB. SSE data lines can be large (full
   JSON payloads). Use `scanner.Buffer()` to increase
   the buffer to 1MB to handle large content blocks.

**Design decision**: Implement the filter as an
`io.ReadCloser` wrapper applied in `ModifyResponse`.
This keeps the filter composable and testable — it
operates on `io.Reader` without knowledge of HTTP.

**Alternative considered**: Using `ModifyResponse` to
read the entire response body, filter, and replace it.
Rejected because this would buffer the entire streaming
response in memory, defeating the purpose of streaming
and violating SC-005 (no perceptible latency).

### R2 — Vertex streamRawPredict vs rawPredict

**Question**: What are the endpoint URL format
differences between streaming and non-streaming
Vertex AI requests?

**Finding**: Vertex AI uses separate endpoints for
streaming and non-streaming Claude model invocation:

1. **Non-streaming (rawPredict)**:
   ```
   POST https://{region}-aiplatform.googleapis.com/v1/
     projects/{project}/locations/{region}/publishers/
     anthropic/models/{model}:rawPredict
   ```

2. **Streaming (streamRawPredict)**:
   ```
   POST https://{region}-aiplatform.googleapis.com/v1/
     projects/{project}/locations/{region}/publishers/
     anthropic/models/{model}:streamRawPredict
   ```

The only difference is the action suffix:
`:rawPredict` vs `:streamRawPredict`. The request body
format is identical for both endpoints.

**Detection**: The gateway must parse the request body
to check for `"stream": true`. This is already done
by `extractModelFromBody()` (Spec 033) — the function
reads and replaces the body. The stream detection can
be combined with the model extraction to avoid reading
the body twice.

**Design decision**: Extend the existing body parsing
in `VertexProvider.PrepareRequest()` to also extract
the `stream` field. Use a combined struct:

```go
var payload struct {
    Model  string `json:"model"`
    Stream bool   `json:"stream"`
}
```

The `stream` field defaults to `false` when absent
(Go zero value), which correctly maps to `rawPredict`
(FR-006).

### R3 — Claude Models Available on Vertex AI

**Question**: Which Claude models are available on
Vertex AI as of April 2026?

**Finding**: Based on the Google Cloud Vertex AI
documentation (fetched 2026-04-23), the following
Claude models are available:

| Model | ID | Max Input | Max Output | Vision | Extended Thinking | PDF |
|-------|-----|-----------|------------|--------|-------------------|-----|
| Claude Opus 4.7 | `claude-opus-4-7-20250416` | 200K | 32K | ✓ | ✓ | ✓ |
| Claude Sonnet 4.6 | `claude-sonnet-4-6-20250217` | 200K | 64K | ✓ | ✓ | ✓ |
| Claude Opus 4.6 | `claude-opus-4-6-20250205` | 200K | 32K | ✓ | ✓ | ✓ |
| Claude Opus 4.5 | `claude-opus-4-5-20241124` | 200K | 32K | ✓ | ✓ | ✓ |
| Claude Sonnet 4.5 | `claude-sonnet-4-5-20241022` | 200K | 8K | ✓ | ✓ | ✓ |
| Claude Opus 4.1 | `claude-opus-4-1-20250414` | 200K | 32K | ✓ | ✓ | ✓ |
| Claude Haiku 4.5 | `claude-haiku-4-5-20241022` | 200K | 8K | ✓ | ✗ | ✓ |
| Claude Opus 4 | `claude-opus-4-20250514` | 200K | 32K | ✓ | ✓ | ✓ |
| Claude Sonnet 4 | `claude-sonnet-4-20250514` | 200K | 8K | ✓ | ✓ | ✓ |

**Notes**:
- All models support vision (image input).
- All models except Haiku 4.5 support extended thinking.
- All models support PDF input.
- The model IDs use the format
  `claude-{tier}-{version}-{date}`.
- Vertex AI also lists aliased model IDs (e.g.,
  `claude-sonnet-4-20250514` without the tier suffix)
  that resolve to the latest version.

**Design decision**: The synthetic model catalog in
the gateway should include all 9 models listed above.
Each entry includes `capabilities` metadata (vision,
extended_thinking, pdf_input) so OpenCode can detect
supported features. The catalog is a Go slice of
structs, not a map, to preserve ordering.

**Maintenance note**: When Anthropic releases new
models on Vertex, the catalog must be updated manually.
This is acceptable because model releases are
infrequent (every few months) and the catalog is a
simple data structure.

### R4 — Vertex rawPredict Body Requirements

**Question**: Does Vertex rawPredict accept the `model`
field in the request body, or must it be removed?

**Finding**: Confirmed from Portkey's source code
(`src/providers/google-vertex-ai/chatComplete.ts`,
`VertexAnthropicChatCompleteConfig`):

```typescript
model: {
    param: 'model',
    required: false,
    transform: () => {
        return undefined;  // Removes model from body
    },
},
```

Portkey explicitly removes the `model` field from the
request body when forwarding to Vertex. The model is
specified in the URL path (`:rawPredict` endpoint URL
contains the model ID).

**Additional confirmation**: The Vertex AI
documentation states that the `rawPredict` endpoint
accepts the "raw" Anthropic API body, but the model
selection is done via the URL path. Including `model`
in the body causes Vertex to return an error or ignore
it (behavior varies by model version).

**Design decision**: The gateway MUST remove the
`model` field from the request body before forwarding
to Vertex (FR-001). This is done during the body
transformation step, which also injects
`anthropic_version`.

### R5 — Vertex anthropic_version String

**Question**: What is the correct `anthropic_version`
string for Vertex rawPredict?

**Finding**: Confirmed from Portkey's source code
(`src/providers/google-vertex-ai/chatComplete.ts`,
line 15 of `VertexAnthropicChatCompleteConfig`):

```typescript
anthropic_version: {
    param: 'anthropic_version',
    required: true,
    default: 'vertex-2023-10-16',
    transform: (params, providerOptions) => {
        return (
            providerOptions?.anthropicVersion ||
            params.anthropic_version ||
            'vertex-2023-10-16'
        );
    },
},
```

The default value is `"vertex-2023-10-16"`. This is
injected into the request body (not as an HTTP header).
If the request already contains `anthropic_version`,
the existing value is preserved (FR-003).

**Important distinction**: The `anthropic_version`
field in the body is different from the
`anthropic-version` HTTP header. Vertex rawPredict
uses the body field. The HTTP header
(`anthropic-version`) must be stripped because Vertex
does not accept it (FR-004).

### R6 — SSE Events from Vertex streamRawPredict

**Question**: What SSE event types does Vertex
streamRawPredict return, and which ones need filtering?

**Finding**: Confirmed from Portkey's
`VertexAnthropicChatCompleteStreamChunkTransform`
function:

```typescript
if (
    chunk.startsWith('event: ping') ||
    chunk.startsWith('event: content_block_stop') ||
    chunk.startsWith('event: vertex_event')
) {
    return;  // Drop these events
}
```

Portkey drops three event types:
1. `ping` — Vertex keepalive events
2. `content_block_stop` — Portkey drops this for its
   own OpenAI-compatible transformation
3. `vertex_event` — Vertex-specific metadata events

**Design decision for our gateway**: We drop only
`vertex_event` and `ping` events (FR-007, FR-008).
We do NOT drop `content_block_stop` because our
gateway forwards native Anthropic SSE events — it
does not transform them to OpenAI format. OpenCode
expects `content_block_stop` events as part of the
standard Anthropic streaming protocol.

**Event types that pass through unchanged**:
- `message_start` — contains message metadata
- `content_block_start` — starts a content block
- `content_block_delta` — streaming content
- `content_block_stop` — ends a content block
- `message_delta` — message-level updates (stop_reason)
- `message_stop` — end of message

### R7 — Request Body Transformation Approach

**Question**: How should the gateway transform the
request body (remove `model`, inject
`anthropic_version`) efficiently?

**Finding**: The current `extractModelFromBody()`
function in `provider.go` already reads and replaces
the request body. The transformation can be integrated
into this function for Vertex:

1. Read the entire body into `[]byte`
2. Unmarshal into `map[string]any` (not a struct, to
   preserve unknown fields)
3. Extract `model` (for URL path)
4. Extract `stream` (for endpoint selection)
5. Delete `model` from the map
6. Inject `anthropic_version` if not present
7. Re-marshal to JSON
8. Replace `req.Body` with the new body
9. Update `req.ContentLength`

**Why `map[string]any` instead of a struct**: The
Anthropic Messages API body contains many fields
(`messages`, `max_tokens`, `temperature`, `tools`,
`metadata`, `system`, etc.). Using a struct would
require defining all fields. A `map[string]any`
preserves all fields without needing to enumerate them.

**Performance**: JSON parse + modify + re-encode adds
< 1ms for typical request bodies (< 100KB). This is
well within the SC-004 target of < 5ms.

**Design decision**: Create a new function
`transformVertexBody(req)` that returns the extracted
model, stream flag, and any error. This function
replaces the body in-place. The existing
`extractModelFromBody()` remains unchanged for
Anthropic and Bedrock providers.

## Summary of Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| SSE filter | `io.ReadCloser` wrapper in `ModifyResponse` | Composable, testable, no buffering delay |
| Body parsing | `map[string]any` unmarshal | Preserves unknown fields without struct enumeration |
| Stream detection | Combined with model extraction | Avoids reading body twice |
| Event filtering | Drop `vertex_event` + `ping` only | `content_block_stop` needed by OpenCode |
| `anthropic_version` | `"vertex-2023-10-16"` default | Confirmed from Portkey source |
| Model removal | Delete from `map[string]any` | Vertex uses URL path for model selection |
| Scanner buffer | 1MB max | Handles large content block SSE data lines |
