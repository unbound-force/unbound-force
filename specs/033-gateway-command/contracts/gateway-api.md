# Contract: Gateway API

**Version**: 1.0.0 | **Date**: 2026-04-20

## Overview

The LLM Gateway exposes an HTTP API on localhost that
proxies Anthropic Messages API requests to the
configured upstream cloud provider. This contract
defines the endpoints, request/response formats, and
error behavior.

## Endpoints

### `GET /health`

Health check endpoint. Returns the gateway's current
status.

**Request**: No body, no required headers.

**Response** (200 OK):

```json
{
  "status": "ok",
  "provider": "anthropic",
  "port": 53147,
  "pid": 42195,
  "uptime_seconds": 3600
}
```

**Error** (503 Service Unavailable): Returned if the
gateway is shutting down.

### `POST /v1/messages`

Proxied to the upstream provider's messages endpoint.

**Request**: Standard Anthropic Messages API request
body. The gateway forwards the body as-is (Anthropic
and Vertex) or transforms it (Bedrock body format).

**Required headers forwarded** (per Claude Code LLM
gateway spec):
- `anthropic-beta`
- `anthropic-version`
- `X-Claude-Code-Session-Id` (if present)
- `Content-Type`

**Ignored headers** (replaced by gateway):
- `Authorization` — replaced with provider credentials
- `x-api-key` — replaced with provider credentials

**Response**: Upstream response forwarded as-is,
including status code, headers, and body. Streaming
responses (SSE) are forwarded in real-time.

### `POST /v1/messages/count_tokens`

Proxied to the upstream provider's token counting
endpoint. Same header forwarding rules as `/v1/messages`.

### All Other Paths

**Response** (405 Method Not Allowed):

```json
{
  "error": {
    "type": "not_found",
    "message": "Unsupported endpoint. Supported: /v1/messages, /v1/messages/count_tokens, /health"
  }
}
```

## Authentication

### Inbound (Client → Gateway)

No authentication required (FR-013). The gateway
listens only on localhost (`127.0.0.1` / `::1`). The
`ANTHROPIC_AUTH_TOKEN=gateway` value set by the sandbox
is a placeholder — the gateway ignores it.

### Outbound (Gateway → Provider)

| Provider | Auth Method | Header |
|----------|------------|--------|
| Anthropic | API key | `x-api-key: <ANTHROPIC_API_KEY>` |
| Vertex AI | OAuth bearer | `Authorization: Bearer <token>` |
| Bedrock | AWS SigV4 | `Authorization: AWS4-HMAC-SHA256 ...` |

## Error Handling

### Upstream Errors

The gateway forwards upstream error responses as-is
(FR-014). Status codes, error bodies, and headers are
preserved. The gateway does NOT retry failed requests.

### Gateway Errors

| Condition | Status | Body |
|-----------|--------|------|
| Unsupported endpoint | 405 | `{"error": {"type": "not_found", ...}}` |
| Token refresh failed | 502 | `{"error": {"type": "auth_error", "message": "..."}}` |
| Upstream unreachable | 502 | `{"error": {"type": "upstream_error", "message": "..."}}` |
| Gateway shutting down | 503 | `{"error": {"type": "unavailable", ...}}` |

## Behavioral Guarantees

1. **Streaming**: The gateway flushes SSE events as
   they arrive from the upstream. No buffering.

2. **Latency**: The gateway adds < 50ms overhead per
   request (SC-005). The proxy is a thin pass-through.

3. **Concurrency**: The gateway handles concurrent
   requests. The `httputil.ReverseProxy` is safe for
   concurrent use. Token state is protected by
   `sync.RWMutex`.

4. **Idempotency**: `uf gateway stop` is idempotent.
   `uf gateway` detects an already-running instance.

## Provider Interface Contract

```go
type Provider interface {
    Name() string
    PrepareRequest(req *http.Request) error
    Start(ctx context.Context) error
    Stop()
}
```

### `Name() string`

Returns the provider identifier: `"anthropic"`,
`"vertex"`, or `"bedrock"`.

### `PrepareRequest(req *http.Request) error`

Called by the `ReverseProxy.Director` function before
each request is forwarded. Must:
- Set `req.URL` to the upstream endpoint
- Set `req.Host` to the upstream host
- Inject authentication headers
- Preserve `anthropic-beta`, `anthropic-version`, and
  `X-Claude-Code-Session-Id` headers
- Return error if credentials are unavailable (e.g.,
  token expired and refresh failed)

### `Start(ctx context.Context) error`

Called once at gateway startup. Must:
- Validate that required credentials are available
- Acquire initial tokens (Vertex, Bedrock)
- Start background refresh goroutines (if applicable)
- Return error if credentials cannot be obtained

### `Stop()`

Called on gateway shutdown. Must:
- Cancel background goroutines
- Release resources
- Be idempotent
