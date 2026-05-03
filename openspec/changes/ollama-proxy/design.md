## Context

The `uf gateway` (Spec 033) established the pattern for
local API proxies: Options struct with injectable deps,
Start/Stop/Status lifecycle, PID file management,
detach mode via re-exec, background token refresh, and
health endpoint. The ollama-proxy follows this pattern
exactly, with a different API surface (Ollama vs
Anthropic) and a different upstream (Vertex AI embedding
API + gateway Messages API).

The gateway's token refresh functions
(`refreshVertexToken`, `refreshLoop`) are currently
unexported in `internal/gateway/`. Both the gateway and
ollama-proxy need Vertex AI OAuth — extracting to a
shared package eliminates duplication.

## Goals / Non-Goals

### Goals
- Implement Ollama-compatible `/api/embed`,
  `/api/generate`, `/api/tags` endpoints
- Reuse the gateway lifecycle pattern (Options,
  Start/Stop/Status, PID, detach)
- Extract shared auth to `internal/auth/`
- Add unified config section for the proxy
- Default to port 11434 for zero-config Dewey usage

### Non-Goals
- Streaming support (`stream: false` is hardcoded in
  Dewey — no streaming needed)
- Multi-provider embedding (Vertex AI only in v1)
- Ollama endpoints beyond what Dewey uses
  (`/api/chat`, `/api/pull`, `/api/show`, etc.)
- Modifying Dewey's code or configuration
- Dewey constitution amendment (deferred to a separate
  Dewey-side issue)

## Decisions

### D1: Package name `internal/ollamaproxy/`

Go does not allow hyphens in package names. The package
is `internal/ollamaproxy/` (not `internal/ollama-proxy/`
as the issue suggested). The command file remains
`cmd/unbound-force/ollama_proxy.go` (underscore per Go
file naming).

### D2: Extract `TokenManager` to `internal/auth/`

Create `internal/auth/` with a `TokenManager` struct
that encapsulates the full Vertex AI OAuth lifecycle:

- Token storage with `sync.RWMutex` protection
- Expiry tracking (`tokenExpiry time.Time`)
- Background refresh via `RefreshLoop` (generic
  ticker-based goroutine, cancellable via context)
- Proactive refresh within 5 minutes of expiry using
  `sync.Mutex.TryLock()` deduplication
- Atomic token invalidation on background refresh
  failure (clear token + expiry under write lock)
- `RefreshVertexToken(execCmd) (string, error)` as
  the credential acquisition function

Also extract `RefreshBedrockCredentials` and helpers
(`parseEnvExport`, `parseAWSCredentialsJSON`) to
`internal/auth/`. Keep SigV4 signing in
`internal/gateway/` (it's HTTP request signing, not
generic auth).

The gateway and ollama-proxy both instantiate a
`TokenManager` instead of implementing their own
token lifecycle. The `RefreshLoop` interval is a
parameter (no mutable package-level variable — the
existing `refreshMinute` var is removed).

Rationale: The stale-token bug (`learning/gateway-3`)
was a real production issue. Extracting the full
`TokenManager` guarantees the proxy inherits all
defensive patterns without re-implementation risk.

### D3: Vertex AI embedding direct call (not reverse proxy)

Unlike the gateway (which uses `httputil.ReverseProxy`
to forward opaque request bodies), the ollama-proxy
must translate between two different API formats. The
embed handler constructs a new `http.Request` to the
Vertex AI endpoint, marshals the Vertex request body,
and unmarshals the response — no reverse proxy.

Auth injection follows the same pattern: set
`Authorization: Bearer <token>` header from the
refreshed OAuth token.

### D4: `/api/generate` delegates to gateway

The generate handler translates the Ollama generate
request to an Anthropic Messages request and POSTs it
to `http://localhost:53147/v1/messages` (the gateway).
The gateway handles Vertex AI credential injection and
Anthropic-to-Vertex translation.

This avoids duplicating the Anthropic/Vertex translation
logic. If the gateway is not running, the handler
returns a clear Ollama-format error.

The gateway health is checked at proxy startup with
an informational warning (not a hard failure) — the
proxy can still serve `/api/embed` without the gateway.

### D5: Model name mapping table

A `map[string]string` maps Ollama model names to cloud
model names:

```go
var defaultModelMap = map[string]string{
    "granite-embedding:30m":                "text-embedding-005",
    "granite-embedding-small-english-r2":   "text-embedding-005",
    "llama3.2:3b":                          "claude-sonnet-4-20250514",
}
```

The map is hardcoded in v1 (configuration deferred to
a future change). Unknown model names are rejected with
an Ollama-format error — passthrough is not allowed
because model names are interpolated into Vertex AI
URLs, and unvalidated strings could contain path
traversal characters.

All model names (both mapped and unknown) MUST be
validated against a safe character set (alphanumeric,
hyphens, colons, periods, underscores) before URL
construction. Names containing `/`, `\`, `%`, or
control characters MUST be rejected.

### D6: Default port 11434

The proxy defaults to port 11434 (same as Ollama) for
zero-config Dewey usage. Users who run both real Ollama
and the proxy must use `--port` to avoid conflict. This
is intentional — the proxy replaces Ollama for GPU-less
users.

### D7: PID file via `internal/pidfile/`

Extract PID file management functions (`WritePID`,
`ReadPID`, `IsAlive`, `CleanupStale`, `RemovePID`,
`PIDInfo`) from `internal/gateway/pid.go` to a new
shared `internal/pidfile/` package. Both gateway and
ollama-proxy import from there.

This avoids `ollamaproxy → gateway` import dependency.
The `PIDInfo.Provider` field is repurposed as a
generic service identifier (e.g., `"vertex"` for
gateway, `"vertex-embedding"` for ollama-proxy).

PID file at `.uf/ollama-proxy.pid`.

### D8: Detach mode via re-exec

Same pattern as gateway: re-exec the binary with
`_UF_OLLAMA_PROXY_CHILD=1` sentinel env var. Child
redirects stdout/stderr to `.uf/ollama-proxy.log`
(created with `0o600` permissions — owner-only
read/write, consistent with gateway log file).
Parent polls `GET /health` with exponential backoff.

### D8a: Security constraints

The proxy MUST NOT include OAuth tokens, API keys, or
credential material in any:
- Error response returned to the Ollama client
- Log message (at any level)
- Health endpoint response
- `.uf/ollama-proxy.log` file content

If the Vertex AI error response echoes the
Authorization header, the proxy MUST redact it before
relaying to the client.

### D8b: GatewayURL validation

`GatewayURL` MUST be validated at startup:
1. Scheme MUST be `http` or `https` only
2. Host MUST resolve to a loopback address
   (`127.0.0.1`, `::1`, `localhost`)
3. Port MUST be valid
4. Path MUST be empty or `/`

Non-loopback URLs are rejected with a clear error.
The gateway is always local — forwarding prompt
content to arbitrary remote endpoints is an SSRF
vector.

### D9: Vertex AI embedding endpoint construction

The Vertex AI embedding predict URL is:
```
https://{REGION}-aiplatform.googleapis.com/v1/projects/{PROJECT}/locations/{REGION}/publishers/google/models/{MODEL}:predict
```

Region and project ID are read from the same env vars
as the gateway (`CLOUD_ML_REGION`, `VERTEX_LOCATION`,
`ANTHROPIC_VERTEX_PROJECT_ID`, `GOOGLE_CLOUD_PROJECT`).
Region resolution follows the same chain as the gateway
provider (D9 in Spec 034 design).

## Test Strategy

All tests are **unit tests** using `httptest.NewServer`
for upstream mocking and injectable `ExecCmd`/`HTTPClient`
for subprocess and HTTP calls.

- **Embed handler**: Mock Vertex AI embedding endpoint
  via `httptest`. Verify Ollama request/response
  translation, batch handling, OAuth header injection.
- **Generate handler**: Mock gateway Messages endpoint
  via `httptest`. Verify Ollama-to-Anthropic translation,
  error propagation when gateway unavailable.
- **Tags handler**: Verify synthetic model list matches
  configured model names.
- **Auth extraction**: Existing gateway token refresh
  tests move to `internal/auth/` — coverage preserved.
- **Lifecycle**: All lifecycle tests MUST use
  `t.TempDir()` for `ProjectDir`, inject
  `ListenAndServe` to avoid real port binding, and
  inject `FindProcess` to avoid real process signaling.
  Follow the `testOpts()` pattern from gateway tests.
- **Coverage targets**:
  - `internal/auth/`: >= existing gateway coverage for
    moved functions (measure before extraction)
  - `internal/pidfile/`: >= existing gateway coverage
    for moved functions
  - `internal/ollamaproxy/`: >= 80% line coverage
    overall, >= 90% for handlers
  - All tests are unit-level using `httptest` mocks —
    no integration tests for v1

## Risks / Trade-offs

### R1: Embedding dimension mismatch

Vertex `text-embedding-005` produces 768-dim vectors vs
granite's 256-dim (per `internal/config/config.go`
`Dimensions: 256`). Dewey stores dimension per model
in the embeddings table key `(block_uuid, model_id)`.
Users must run `dewey reindex` once when switching.
The proxy logs a startup warning about this requirement.

### R5: PID file race condition

PID file operations have a TOCTOU race window between
checking existence and writing (same as gateway).
Acceptable for a single-user developer tool.

### R2: Port conflict with real Ollama

Defaulting to 11434 means real Ollama and the proxy
cannot run simultaneously. Intentional — `--port` flag
available for users who need both.

### R3: Gateway dependency for generation

`/api/generate` requires `uf gateway` running. If
gateway is down, compilation and curation fail but
semantic search (the P1 feature) continues working.
The proxy warns at startup if gateway is unreachable.

### R4: gcloud CLI dependency

Same dependency as the gateway. Mitigated by the same
pattern: `uf doctor` checks for gcloud, `uf setup`
provides install guidance.
<!-- scaffolded by uf vdev -->
