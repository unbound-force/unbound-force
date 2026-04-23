# Quickstart: Gateway Vertex Translation

**Branch**: `034-gateway-vertex-translation` | **Date**: 2026-04-23
**Spec**: `specs/034-gateway-vertex-translation/spec.md`

## Prerequisites

- Go 1.24+ installed
- `gcloud` CLI installed and authenticated
  (`gcloud auth application-default login`)
- Vertex AI API enabled on your GCP project
- Environment variables set:
  ```bash
  export CLAUDE_CODE_USE_VERTEX=1
  export ANTHROPIC_VERTEX_PROJECT_ID=your-project-id
  export CLOUD_ML_REGION=us-east5  # optional, defaults to us-east5
  ```

## Verification Steps

### 1. Build and Run Tests

```bash
# Build the project
make build

# Run all tests (including new gateway translation tests)
make test

# Run only gateway tests
go test -race -count=1 ./internal/gateway/...
```

### 2. Verify Request Body Transformation

Start the gateway and send a test request:

```bash
# Start the gateway (requires Vertex credentials)
uf gateway --port 53147

# In another terminal, send a standard Anthropic request:
curl -s http://localhost:53147/v1/messages \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -H "anthropic-beta: messages-2024-01-01" \
  -d '{
    "model": "claude-sonnet-4-20250514",
    "max_tokens": 100,
    "messages": [{"role": "user", "content": "Say hello"}]
  }'
```

**Expected**: The gateway translates the request:
- Removes `model` from the body
- Injects `anthropic_version: "vertex-2023-10-16"`
- Strips `anthropic-version` and `anthropic-beta` headers
- Forwards to Vertex rawPredict endpoint
- Returns a valid Claude response

### 3. Verify Streaming with SSE Filtering

```bash
curl -s -N http://localhost:53147/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4-20250514",
    "max_tokens": 100,
    "stream": true,
    "messages": [{"role": "user", "content": "Say hello"}]
  }'
```

**Expected**: The gateway:
- Uses `streamRawPredict` endpoint (not `rawPredict`)
- Filters out `vertex_event` and `ping` SSE events
- Forwards standard Anthropic events unchanged
  (`message_start`, `content_block_delta`, etc.)

### 4. Verify Model Catalog

```bash
# List all models
curl -s http://localhost:53147/v1/models | python3 -m json.tool

# Get a specific model
curl -s http://localhost:53147/v1/models/claude-sonnet-4-20250514 | python3 -m json.tool

# Verify 404 for unknown model
curl -s http://localhost:53147/v1/models/unknown-model
```

**Expected**:
- `/v1/models` returns at least 9 models with
  capabilities metadata
- `/v1/models/{id}` returns the model details
- Unknown model returns 404

### 5. Verify Sandbox Integration

```bash
# Start sandbox with Vertex credentials on host
uf sandbox start

# Inside the sandbox, verify environment:
# - ANTHROPIC_BASE_URL should be set
# - ANTHROPIC_AUTH_TOKEN should be "gateway"
# - CLAUDE_CODE_USE_VERTEX should NOT be set
# - ANTHROPIC_VERTEX_PROJECT_ID should NOT be set
```

**Expected**: OpenCode inside the sandbox sees only
`ANTHROPIC_BASE_URL` and `ANTHROPIC_AUTH_TOKEN=gateway`.
No Vertex-specific variables are present.

### 6. Verify Backward Compatibility

```bash
# Set Anthropic direct credentials (not Vertex)
export ANTHROPIC_API_KEY=sk-ant-your-key
unset CLAUDE_CODE_USE_VERTEX
unset ANTHROPIC_VERTEX_PROJECT_ID

# Start gateway
uf gateway --port 53147

# Send a request — should work without translation
curl -s http://localhost:53147/v1/messages \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude-sonnet-4-20250514",
    "max_tokens": 100,
    "messages": [{"role": "user", "content": "Say hello"}]
  }'
```

**Expected**: No request body transformation, no
header stripping, no SSE filtering. The request and
response pass through unchanged (FR-010).

## Success Criteria Checklist

| SC | Description | How to Verify |
|----|-------------|---------------|
| SC-001 | Sandbox chat works with Vertex | Step 5 |
| SC-002 | Standard SSE events pass through | Step 3 |
| SC-003 | vertex_event/ping dropped | Step 3 (no type validation errors) |
| SC-004 | < 5ms body transformation | `go test -bench` on `transformVertexBody` |
| SC-005 | No perceptible SSE latency | Step 3 (streaming feels instant) |
| SC-006 | No Vertex vars in container | Step 5 |
| SC-007 | Anthropic path unchanged | Step 6 |
