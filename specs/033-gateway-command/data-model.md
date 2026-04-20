# Data Model: LLM Gateway Command

**Branch**: `033-gateway-command` | **Date**: 2026-04-20
**Spec**: `specs/033-gateway-command/spec.md`

## Entities

### Gateway Options

The `Options` struct configures gateway operations.
Follows the established pattern from
`internal/sandbox/sandbox.go` — all external
dependencies are injected as function fields for
testability (Constitution Principle IV).

```go
// Options configures gateway operations. All external
// dependencies are injected as function fields for
// testability per Constitution Principle IV.
type Options struct {
    // Port is the local port to listen on.
    // Default: 53147.
    Port int

    // ProviderName overrides auto-detection.
    // Valid values: "anthropic", "vertex", "bedrock".
    // Default: "" (auto-detect from env vars).
    ProviderName string

    // Detach starts the gateway in the background.
    Detach bool

    // ProjectDir is the project directory (for PID
    // file location at <ProjectDir>/.uf/gateway.pid).
    ProjectDir string

    // Stdout is the writer for user-facing output.
    Stdout io.Writer

    // Stderr is the writer for progress/status messages.
    Stderr io.Writer

    // --- Injectable dependencies ---

    // LookPath finds a binary in PATH.
    LookPath func(string) (string, error)

    // ExecCmd runs a command and returns combined output.
    ExecCmd func(name string, args ...string) ([]byte, error)

    // Getenv reads an environment variable.
    Getenv func(string) string

    // HTTPGet performs an HTTP GET and returns status code.
    // Used for health check polling.
    HTTPGet func(url string) (int, error)

    // ReadFile reads a file's contents.
    ReadFile func(string) ([]byte, error)

    // WriteFile writes data to a file with permissions.
    WriteFile func(string, []byte, os.FileMode) error

    // FindProcess looks up a process by PID.
    FindProcess func(int) (*os.Process, error)

    // Signal sends a signal to a process. Used for
    // liveness checks (Signal(0)) and stop (SIGTERM).
    // Injected for testability.

    // ListenAndServe starts the HTTP server. Injected
    // for testability — tests can provide a no-op or
    // channel-based implementation.
    ListenAndServe func(addr string, handler http.Handler) error
}
```

### Provider Interface

The `Provider` interface abstracts cloud provider
differences. Follows the Strategy pattern established
in `internal/sandbox/backend.go`.

```go
// Provider abstracts the upstream cloud provider's
// authentication and URL rewriting strategy.
//
// Design decision: Strategy pattern per SOLID
// Open/Closed Principle. Adding a new provider
// (e.g., OpenAI-compatible) requires only a new
// implementation, not modification of the gateway
// core.
type Provider interface {
    // Name returns the provider identifier
    // ("anthropic", "vertex", "bedrock").
    Name() string

    // PrepareRequest modifies the outbound request
    // before it is forwarded to the upstream provider.
    // Responsibilities:
    //   - Set the upstream URL (scheme, host, path)
    //   - Inject authentication headers
    //   - Transform the request body if needed
    //     (e.g., Bedrock body format)
    PrepareRequest(req *http.Request) error

    // Start initializes the provider (e.g., acquire
    // initial OAuth token for Vertex). Called once
    // at gateway startup. Returns error if credentials
    // are not available.
    Start(ctx context.Context) error

    // Stop cleans up provider resources (e.g., stop
    // token refresh goroutine). Called on gateway
    // shutdown.
    Stop()
}
```

### Provider Implementations

#### AnthropicProvider

```go
type AnthropicProvider struct {
    apiKey string // from ANTHROPIC_API_KEY
}
```

- `PrepareRequest`: Sets `req.URL` to
  `https://api.anthropic.com/v1/messages` (or
  `count_tokens`), adds `x-api-key` header.
- `Start`: Reads `ANTHROPIC_API_KEY` from env. Returns
  error if empty.
- `Stop`: No-op (no refresh needed).

#### VertexProvider

```go
type VertexProvider struct {
    projectID string // from ANTHROPIC_VERTEX_PROJECT_ID
    region    string // from CLOUD_ML_REGION, default "us-east5"
    token     string // current OAuth token
    tokenMu   sync.RWMutex
    cancel    context.CancelFunc // stops refresh goroutine
    execCmd   func(string, ...string) ([]byte, error)
}
```

- `PrepareRequest`: Sets `req.URL` to the Vertex
  rawPredict endpoint, adds `Authorization: Bearer`
  header with current token.
- `Start`: Calls `gcloud auth application-default
  print-access-token` to get initial token. Starts
  refresh goroutine (50-minute interval per R2).
- `Stop`: Cancels refresh goroutine context.

#### BedrockProvider

```go
type BedrockProvider struct {
    region      string // from AWS_REGION or AWS_DEFAULT_REGION
    accessKey   string
    secretKey   string
    sessionToken string
    tokenMu     sync.RWMutex
    cancel      context.CancelFunc
    execCmd     func(string, ...string) ([]byte, error)
}
```

- `PrepareRequest`: Sets `req.URL` to the Bedrock
  invoke endpoint, signs the request with SigV4.
- `Start`: Calls `aws configure export-credentials
  --format env` to get initial credentials. Starts
  refresh goroutine.
- `Stop`: Cancels refresh goroutine context.

### PID File Format

Plain text file at `.uf/gateway.pid`. Contains
structured metadata for `uf gateway status`:

```text
<PID>
port=<PORT>
provider=<PROVIDER_NAME>
started=<RFC3339_TIMESTAMP>
```

Example:

```text
42195
port=53147
provider=vertex
started=2026-04-20T14:30:00Z
```

**Parsing**: Line 1 is always the PID (decimal integer).
Subsequent lines are `key=value` metadata pairs. Unknown
keys are ignored for forward compatibility.

```go
// PIDInfo represents the contents of the PID file.
type PIDInfo struct {
    PID      int
    Port     int
    Provider string
    Started  time.Time
}

func WritePID(path string, info PIDInfo) error
func ReadPID(path string) (*PIDInfo, error)
func IsAlive(pid int) bool
func RemovePID(path string) error
```

### Health Response Format

`GET /health` returns JSON (per Constitution Principle
III: Observable Quality — machine-parseable output):

```json
{
  "status": "ok",
  "provider": "vertex",
  "port": 53147,
  "pid": 42195,
  "uptime_seconds": 3600
}
```

```go
// HealthResponse is the JSON payload for GET /health.
type HealthResponse struct {
    Status        string `json:"status"`
    Provider      string `json:"provider"`
    Port          int    `json:"port"`
    PID           int    `json:"pid"`
    UptimeSeconds int64  `json:"uptime_seconds"`
}
```

### Gateway Status Display

`uf gateway status` output format (human-readable):

```text
Gateway Status
  Provider:  vertex
  Port:      53147
  PID:       42195
  Uptime:    1h 23m
```

When no gateway is running:

```text
No gateway running.
```

### Provider Detection Priority

Auto-detection follows a priority order (FR-003):

```text
1. CLAUDE_CODE_USE_VERTEX=1 + ANTHROPIC_VERTEX_PROJECT_ID → Vertex
2. CLAUDE_CODE_USE_BEDROCK=1                              → Bedrock
3. ANTHROPIC_API_KEY present                              → Anthropic
4. None matched                                           → Error
```

Vertex is checked first because a developer may have
both `ANTHROPIC_API_KEY` and Vertex env vars set (the
API key might be for a different tool). The explicit
`CLAUDE_CODE_USE_VERTEX=1` flag indicates intent.

### Sandbox Integration Points

Modified fields on `sandbox.Options` (none — the
gateway integration uses existing fields):

- `Getenv`: Used to check for gateway health endpoint
- `ExecCmd`: Used to start `uf gateway --detach`
- `HTTPGet`: Used to poll gateway health

Modified behavior in `sandbox.Start()`:

```text
Before container start:
  1. Check if gateway is already running (health check)
  2. If not, detect provider from env vars
  3. If provider detected, start gateway with --detach
  4. Wait for gateway health endpoint
  5. Set ANTHROPIC_BASE_URL and ANTHROPIC_AUTH_TOKEN
     in container env
  6. Skip credential mounts and API key forwarding

If gateway unavailable:
  Fall back to existing credential mount behavior
  (backward compatible per FR-012)
```

Modified behavior in `buildRunArgs()`:

```text
When gateway is active:
  - Add: -e ANTHROPIC_BASE_URL=http://host.containers.internal:<port>
  - Add: -e ANTHROPIC_AUTH_TOKEN=gateway
  - Skip: credential file mounts (gcloud dir, service account)
  - Skip: ANTHROPIC_API_KEY forwarding
  - Skip: CLAUDE_CODE_USE_VERTEX forwarding
  - Keep: OLLAMA_HOST (unrelated to gateway)
  - Keep: OPENAI_API_KEY, GEMINI_API_KEY, etc. (not proxied)
```

## Entity Relationships

```text
Gateway (1) ──uses──> Provider (1)
  │                      │
  ├── Options            ├── AnthropicProvider
  ├── PIDInfo            ├── VertexProvider
  └── HealthResponse     └── BedrockProvider

Sandbox.Start() ──auto-starts──> Gateway (--detach)
  │
  ├── detects provider from env
  ├── checks gateway health
  └── passes ANTHROPIC_BASE_URL to container
```

## File Ownership

| File | Owner | Rationale |
|------|-------|-----------|
| `internal/gateway/gateway.go` | New | Gateway core |
| `internal/gateway/provider.go` | New | Provider interface + implementations |
| `internal/gateway/refresh.go` | New | Token refresh goroutine |
| `internal/gateway/pid.go` | New | PID file management |
| `internal/gateway/gateway_test.go` | New | Tests for all gateway functions |
| `cmd/unbound-force/gateway.go` | New | Cobra command registration |
| `cmd/unbound-force/main.go` | Modified | Add `newGatewayCmd()` |
| `internal/sandbox/sandbox.go` | Modified | Gateway auto-start in `Start()` |
| `internal/sandbox/config.go` | Modified | Gateway-aware `buildRunArgs()` |
