# Quickstart: LLM Gateway Command

**Branch**: `033-gateway-command` | **Date**: 2026-04-20

## Verification Steps

### 1. Build

```bash
make build
# or: go build ./...
```

Verify no compilation errors. The `uf` binary should
include the new `gateway` command.

### 2. Unit Tests

```bash
go test -race -count=1 ./internal/gateway/...
go test -race -count=1 ./cmd/unbound-force/...
go test -race -count=1 ./internal/sandbox/...
```

All tests must pass. Run the full suite to verify no
regressions in sandbox behavior:

```bash
make test
# or: go test -race -count=1 ./...
```

### 3. Lint

```bash
golangci-lint run
```

No new lint findings.

### 4. CLI Smoke Test — Help Output

```bash
uf gateway --help
uf gateway stop --help
uf gateway status --help
```

Verify:
- `uf gateway` shows usage with `--port`, `--provider`,
  `--detach` flags
- `uf gateway stop` shows usage
- `uf gateway status` shows usage

### 5. Provider Auto-Detection (Anthropic)

```bash
export ANTHROPIC_API_KEY=sk-test-key
uf gateway
```

Expected output:
```
Gateway started on port 53147 (provider: anthropic)
```

In another terminal:
```bash
curl -s http://localhost:53147/health | jq .
```

Expected:
```json
{
  "status": "ok",
  "provider": "anthropic",
  "port": 53147,
  "pid": <PID>,
  "uptime_seconds": <N>
}
```

Stop with Ctrl+C. Verify PID file is removed:
```bash
ls .uf/gateway.pid  # should not exist
```

### 6. Background Mode

```bash
uf gateway --detach
```

Expected output:
```
Gateway started on port 53147 (provider: anthropic, PID: <N>)
```

Verify:
```bash
uf gateway status
# Should show provider, port, PID, uptime

uf gateway stop
# Should show "Gateway stopped."

uf gateway status
# Should show "No gateway running."
```

### 7. Sandbox Integration

```bash
# With ANTHROPIC_API_KEY set:
uf sandbox start --detach

# Check container environment:
podman exec uf-sandbox env | grep ANTHROPIC_BASE_URL
# Expected: ANTHROPIC_BASE_URL=http://host.containers.internal:53147

podman exec uf-sandbox env | grep ANTHROPIC_AUTH_TOKEN
# Expected: ANTHROPIC_AUTH_TOKEN=gateway

# Verify no credential mounts:
podman inspect uf-sandbox --format '{{.Mounts}}' | grep gcloud
# Expected: no output (gcloud dir not mounted)

# Clean up:
uf sandbox stop
uf gateway stop
```

### 8. Fallback Behavior

```bash
# Unset all provider env vars:
unset ANTHROPIC_API_KEY
unset CLAUDE_CODE_USE_VERTEX
unset CLAUDE_CODE_USE_BEDROCK

uf sandbox start --detach
# Should start without gateway (fallback to credential mounts)
# Should log: "Gateway not available — using credential mounts"

uf sandbox stop
```

### 9. Port Conflict

```bash
# Start a process on port 53147:
python3 -c "import http.server; http.server.HTTPServer(('', 53147), http.server.SimpleHTTPRequestHandler).serve_forever()" &
PY_PID=$!

uf gateway
# Expected error: "port 53147 already in use, use --port to specify an alternative"

kill $PY_PID
```

### 10. Custom Port and Provider

```bash
export ANTHROPIC_API_KEY=sk-test-key
uf gateway --provider anthropic --port 9000
```

Verify:
```bash
curl -s http://localhost:9000/health | jq .provider
# Expected: "anthropic"
```

### 11. No Provider Detected

```bash
unset ANTHROPIC_API_KEY
unset CLAUDE_CODE_USE_VERTEX
unset CLAUDE_CODE_USE_BEDROCK

uf gateway
# Expected error listing supported providers and env vars
```

### 12. Unsupported Endpoint

```bash
export ANTHROPIC_API_KEY=sk-test-key
uf gateway --detach

curl -s -o /dev/null -w "%{http_code}" \
  http://localhost:53147/v1/completions
# Expected: 405

uf gateway stop
```
