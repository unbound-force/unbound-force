# Contract: Gateway Vertex Translation API

**Version**: 1.0.0 | **Date**: 2026-04-23
**Extends**: `specs/033-gateway-command/contracts/gateway-api.md`

## Overview

This contract extends the Gateway API contract (Spec
033) with Vertex AI-specific request body translation,
streaming endpoint detection, and SSE response
filtering. The gateway accepts standard Anthropic
Messages API requests and transparently translates
them to/from Vertex rawPredict format.

## Request Translation (Vertex Provider Only)

### Body Transformation

When the active provider is `vertex`, the gateway
transforms the request body before forwarding:

**Removed fields**:
- `model` — Vertex uses the URL path for model
  selection. The `model` value is extracted and used
  to build the endpoint URL.

**Injected fields**:
- `anthropic_version: "vertex-2023-10-16"` — Required
  by Vertex rawPredict. Only injected if the field is
  not already present in the request body.

**Preserved fields**:
- All other fields (`messages`, `max_tokens`,
  `temperature`, `tools`, `system`, `metadata`,
  `stream`, `thinking`, etc.) pass through unchanged.

**Example — Input (from OpenCode)**:

```json
{
  "model": "claude-sonnet-4-20250514",
  "max_tokens": 4096,
  "messages": [
    {"role": "user", "content": "Hello"}
  ],
  "stream": true
}
```

**Example — Output (to Vertex)**:

```json
{
  "anthropic_version": "vertex-2023-10-16",
  "max_tokens": 4096,
  "messages": [
    {"role": "user", "content": "Hello"}
  ],
  "stream": true
}
```

### Header Stripping

When the active provider is `vertex`, the gateway
strips the following HTTP headers before forwarding:

| Header | Reason |
|--------|--------|
| `anthropic-beta` | Vertex rawPredict does not accept this header |
| `anthropic-version` | Vertex uses `anthropic_version` in the body instead |

**Preserved headers** (forwarded to Vertex):
- `Content-Type`
- `X-Claude-Code-Session-Id` (if present)

**Replaced headers** (same as Spec 033):
- `Authorization` — replaced with `Bearer <oauth-token>`
- `x-api-key` — removed (Vertex uses Bearer auth)

### Endpoint Selection

| Condition | Endpoint |
|-----------|----------|
| `"stream": true` in body | `.../{model}:streamRawPredict` |
| `"stream": false` in body | `.../{model}:rawPredict` |
| `stream` absent from body | `.../{model}:rawPredict` |
| Body is malformed/empty | `.../{model}:rawPredict` |
| `count_tokens` path | `.../{model}:rawPredict` (always) |

**URL format**:

```
https://{region}-aiplatform.googleapis.com/v1/
  projects/{project}/locations/{region}/publishers/
  anthropic/models/{model}:{action}
```

Where `{action}` is `rawPredict` or `streamRawPredict`.

### Content-Length Update

After modifying the request body, the gateway MUST
update `req.ContentLength` to match the new body
length (FR-014). Failure to do so causes Vertex to
reject the request with a length mismatch error.

## Response Filtering (Vertex Provider Only)

### SSE Event Filtering

When the response `Content-Type` is
`text/event-stream` (streaming response), the gateway
filters the following SSE event types before
forwarding to the client:

| Event Type | Action | Reason |
|------------|--------|--------|
| `vertex_event` | **Dropped** | Vertex-specific metadata; OpenCode cannot parse it |
| `ping` | **Dropped** | Vertex keepalive; not part of Anthropic SSE protocol |
| `message_start` | Forwarded | Standard Anthropic event |
| `content_block_start` | Forwarded | Standard Anthropic event |
| `content_block_delta` | Forwarded | Standard Anthropic event |
| `content_block_stop` | Forwarded | Standard Anthropic event |
| `message_delta` | Forwarded | Standard Anthropic event |
| `message_stop` | Forwarded | Standard Anthropic event |

### Non-Streaming Responses

When the response `Content-Type` is NOT
`text/event-stream`, no filtering is applied. The
response passes through unchanged (FR-009).

### Error Responses

Vertex error responses (non-200 status codes) are
forwarded unchanged. No filtering or transformation
is applied to error responses.

## Synthetic Model Catalog

### `GET /v1/models`

Returns a list of Claude models available on Vertex AI.

**Response** (200 OK):

```json
{
  "data": [
    {
      "id": "claude-opus-4-7-20250416",
      "type": "model",
      "display_name": "Claude Opus 4.7",
      "created_at": "2025-04-16T00:00:00Z",
      "max_input_tokens": 200000,
      "max_tokens": 32768,
      "capabilities": {
        "vision": true,
        "extended_thinking": true,
        "pdf_input": true
      }
    }
  ],
  "has_more": false,
  "first_id": "claude-opus-4-7-20250416",
  "last_id": "claude-sonnet-4-20250514"
}
```

### `GET /v1/models/{model_id}`

Returns a single model's details.

**Response** (200 OK): Single model object (same
format as list entries).

**Response** (404 Not Found):

```json
{
  "error": {
    "type": "not_found",
    "message": "Model 'unknown-model' not found"
  }
}
```

## Provider Isolation

### No Translation for Anthropic Provider

When the active provider is `anthropic` (direct API),
no request or response translation occurs (FR-010).
The gateway behaves exactly as in Spec 033:
- Request body forwarded as-is
- `anthropic-beta` and `anthropic-version` headers
  preserved
- Response forwarded as-is (no SSE filtering)

### Sandbox Environment

When the gateway is active, the sandbox container
receives only:
- `ANTHROPIC_BASE_URL=http://host.containers.internal:{port}`
- `ANTHROPIC_AUTH_TOKEN=gateway`

The following variables are NOT forwarded (FR-011):
- `CLAUDE_CODE_USE_VERTEX`
- `ANTHROPIC_VERTEX_PROJECT_ID`
- `VERTEX_LOCATION`
- `GOOGLE_CLOUD_PROJECT`
- `ANTHROPIC_API_KEY`

This is already implemented by `gatewaySkippedKeys`
in `internal/sandbox/config.go` (Spec 033). No
changes needed.

## Behavioral Guarantees

1. **Latency**: Request body transformation adds
   < 5ms overhead (SC-004). SSE filtering adds no
   perceptible latency — events are forwarded within
   one streaming chunk (SC-005).

2. **Correctness**: All standard Anthropic SSE event
   types pass through unchanged (SC-002). Only
   `vertex_event` and `ping` are dropped (SC-003).

3. **Backward compatibility**: The Anthropic direct
   provider path is fully functional with no
   behavioral changes (SC-007).

4. **Partial events**: The SSE filter correctly
   handles events that span multiple `Read()` calls
   by buffering until a complete event boundary
   (`\n\n`) is detected (FR-015).
