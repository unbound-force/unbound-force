## Why

Dewey (the semantic knowledge layer) uses Ollama for two
capabilities: embedding via `POST /api/embed` (semantic
search, similarity, knowledge linking) and LLM
generation via `POST /api/generate` (knowledge
compilation and curation). Both require GPU-accelerated
models running locally. Team members without GPUs cannot
use Dewey's semantic search or compilation features —
the `--no-embeddings` flag degrades to keyword-only
search, losing the primary value proposition.

Rather than adding cloud provider abstractions to Dewey
itself (which would violate its "Local-Only Processing"
constitution principle), we add a transparent proxy in
the `uf` binary that presents an Ollama-compatible API
backed by Vertex AI. Dewey talks to `localhost:11434`
unchanged — it cannot tell whether it's hitting real
Ollama or the proxy.

This is consistent with the `uf gateway` pattern
(Spec 033): a local proxy that handles credential
injection and API translation, keeping cloud provider
logic out of downstream tools.

## What Changes

### New `uf ollama-proxy` command

Add a new top-level command with the same lifecycle
pattern as `uf gateway`: start (default), stop, status,
`--detach` mode, PID file management, health endpoint.

### Ollama API translation layer

Implement three Ollama endpoints:

1. `POST /api/embed` — translate to Vertex AI
   embedding API (`text-embedding-005`), handle OAuth
   token refresh via `gcloud` CLI
2. `POST /api/generate` — translate to Anthropic
   Messages API via `uf gateway` (which handles Vertex
   AI credential injection)
3. `GET /api/tags` — return synthetic model list so
   Dewey's `checkModelAvailable()` passes

### Shared auth extraction

Export the gateway's Vertex AI token refresh functions
(`refreshVertexToken`, `refreshLoop`) to a new
`internal/auth` package so both gateway and
ollama-proxy can reuse them without duplication.

### Configuration

Add `OllamaProxyConfig` section to the unified config
(`internal/config/`) with port, embedding model, and
gateway URL fields. Environment variable overrides
following the `UF_OLLAMA_PROXY_*` pattern.

## Capabilities

### New Capabilities
- `uf ollama-proxy`: Ollama-compatible HTTP API on
  port 11434 backed by Vertex AI embedding and
  Anthropic Messages API via gateway.
- `uf ollama-proxy stop`: Stop a running proxy.
- `uf ollama-proxy status`: Show proxy status.
- `internal/ollamaproxy/`: Translation layer for
  `/api/embed`, `/api/generate`, `/api/tags`.
- `internal/auth/`: Shared Vertex AI OAuth token
  refresh (extracted from gateway).
- `OllamaProxyConfig` in unified config.

### Modified Capabilities
- `internal/gateway/`: Token refresh functions moved
  to `internal/auth/`, gateway imports from there.
- `internal/config/`: New `OllamaProxy` section added.

### Removed Capabilities
- None.

## Impact

### Files Affected

| Area | Changes |
|------|---------|
| `internal/ollamaproxy/` | NEW: proxy.go, embed.go, generate.go, tags.go, proxy_test.go |
| `internal/auth/` | NEW: vertex.go, refresh.go, vertex_test.go (extracted from gateway) |
| `internal/gateway/` | Import token refresh from `internal/auth/` instead of local functions |
| `internal/config/config.go` | Add `OllamaProxyConfig` section |
| `internal/config/template.go` | Add ollama-proxy config template |
| `internal/config/config_test.go` | Add ollama-proxy config tests |
| `cmd/unbound-force/ollama_proxy.go` | NEW: Cobra command |
| `cmd/unbound-force/main.go` | Register ollama-proxy command |

### External Dependencies
- `gcloud` CLI for OAuth token refresh (same as gateway)
- `uf gateway` for `/api/generate` translation (optional — embedding works independently)
- No new Go module dependencies (stdlib HTTP + JSON)

## Constitution Alignment

### I. Autonomous Collaboration

**Assessment**: PASS — the proxy runs independently.
Dewey communicates with it via standard HTTP. No
runtime coupling between the proxy and Dewey beyond
the Ollama API contract.

### II. Composability First

**Assessment**: PASS — the proxy is independently
installable and usable. It benefits any tool that
speaks the Ollama API, not just Dewey. The `/api/embed`
endpoint works without the gateway. The `/api/generate`
endpoint gracefully degrades when the gateway is
unavailable.

### III. Observable Quality

**Assessment**: PASS — health endpoint at `GET /health`
returns JSON status. PID file management for lifecycle
observability. `uf ollama-proxy status` shows running
state, port, and uptime.

### IV. Testability

**Assessment**: PASS — all HTTP handlers testable via
`httptest`. Token refresh uses injected `ExecCmd`.
Vertex AI calls use injected `HTTPClient`. No external
services required for unit tests.
<!-- scaffolded by uf vdev -->
