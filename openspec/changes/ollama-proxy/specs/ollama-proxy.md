## ADDED Requirements

### Requirement: Embed endpoint

`POST /api/embed` MUST accept Ollama embed requests and
return Ollama embed responses, translating to the
Vertex AI embedding predict API.

Ollama request: `{"model": "<name>", "input": ["text"]}`
Ollama response: `{"model": "<name>", "embeddings": [[float...]]}`

Vertex request: `{"instances": [{"content": "text"}]}`
Vertex response: `{"predictions": [{"embeddings": {"values": [float...]}}]}`

The handler MUST:
- Map the Ollama model name to a Vertex model via the
  model name mapping table
- Transform `input[]` to `instances[].content`
- Transform `predictions[].embeddings.values` to
  `embeddings[]`
- Handle batch (multiple inputs produce multiple
  embedding vectors)
- Inject `Authorization: Bearer <token>` from the
  refreshed OAuth token
- Return Ollama-format JSON errors on failure

#### Scenario: Single embedding

- **GIVEN** the proxy is running and Vertex AI OAuth
  is configured
- **WHEN** `POST /api/embed` receives
  `{"model": "granite-embedding:30m", "input": ["hello"]}`
- **THEN** it MUST call Vertex AI with
  `{"instances": [{"content": "hello"}]}` and return
  `{"model": "granite-embedding:30m", "embeddings": [[0.1, ...]]}`

#### Scenario: Batch embedding

- **GIVEN** the proxy is running
- **WHEN** `POST /api/embed` receives
  `{"model": "granite-embedding:30m", "input": ["a", "b"]}`
- **THEN** it MUST call Vertex AI with
  `{"instances": [{"content": "a"}, {"content": "b"}]}`
  and return two embedding vectors

#### Scenario: Unknown model name

- **GIVEN** an unknown model name not in the mapping
- **WHEN** `POST /api/embed` receives the request
- **THEN** it MUST return an Ollama-format error
  rejecting the unknown model name

#### Scenario: Model name with path traversal

- **GIVEN** a model name containing `/`, `\`, `%`,
  or control characters
- **WHEN** `POST /api/embed` receives the request
- **THEN** it MUST reject it with an Ollama-format
  error before constructing the Vertex URL

#### Scenario: Vertex API error

- **GIVEN** Vertex AI returns a non-200 status
- **WHEN** the embed handler processes the response
- **THEN** it MUST return an Ollama-format error
  response with the upstream error details, with any
  OAuth tokens redacted from the error body

### Requirement: Generate endpoint

`POST /api/generate` MUST accept Ollama generate
requests and return Ollama generate responses,
translating to the Anthropic Messages API via the
gateway.

Ollama request: `{"model": "<name>", "prompt": "...", "stream": false}`
Ollama response: `{"model": "<name>", "response": "..."}`

Anthropic request: `{"model": "<claude-model>", "max_tokens": 4096, "messages": [{"role": "user", "content": "..."}]}`
Anthropic response: `{"content": [{"text": "..."}]}`

The handler MUST:
- Map the Ollama model name to a Claude model via the
  model name mapping table
- Wrap the `prompt` string into Anthropic Messages
  format
- Extract `content[0].text` from the response into
  the `response` field
- POST to the gateway URL (default
  `http://localhost:53147/v1/messages`)
- Return an Ollama-format error when the gateway is
  unreachable

#### Scenario: Successful generation

- **GIVEN** the proxy is running and `uf gateway` is
  running
- **WHEN** `POST /api/generate` receives
  `{"model": "llama3.2:3b", "prompt": "Hello", "stream": false}`
- **THEN** it MUST POST to the gateway with
  `{"model": "claude-sonnet-4-20250514", ...}` and return
  `{"model": "llama3.2:3b", "response": "..."}`

#### Scenario: Gateway unavailable

- **GIVEN** the gateway is not running
- **WHEN** `POST /api/generate` receives a request
- **THEN** it MUST return an Ollama-format error
  containing "gateway not available" and instructions
  to run `uf gateway`

### Requirement: Tags endpoint

`GET /api/tags` MUST return a synthetic model list
containing the configured Ollama model names.

#### Scenario: Model availability check

- **GIVEN** the proxy is configured with model mappings
  for `granite-embedding:30m` and `llama3.2:3b`
- **WHEN** `GET /api/tags` is called
- **THEN** the response MUST include both model names
  in the `models` array

### Requirement: Proxy lifecycle

The proxy MUST follow the `uf gateway` lifecycle
pattern:

- `Start(opts)`: Validate config, start token refresh,
  check gateway health (warn if unavailable), write
  PID file, start HTTP server
- `Stop(opts)`: Read PID, send SIGTERM, remove PID
- `Status(opts)`: Read PID, check alive, display info

PID file at `.uf/ollama-proxy.pid`. Log file at
`.uf/ollama-proxy.log` (detach mode, `0o600`
permissions).

The proxy MUST NOT include OAuth tokens, API keys, or
credential material in any error response, log message,
health endpoint output, or log file content.

`GatewayURL` MUST be validated at startup: scheme
`http`/`https` only, host MUST be loopback
(`127.0.0.1`, `::1`, `localhost`). Non-loopback URLs
MUST be rejected with a clear error.

#### Scenario: Non-loopback gateway URL

- **GIVEN** `GatewayURL` is `http://192.168.1.1:53147`
- **WHEN** the proxy starts
- **THEN** it MUST return an error refusing to forward
  requests to a non-loopback host

#### Scenario: Token redaction in error response

- **GIVEN** Vertex AI returns an error that echoes
  the Authorization header
- **WHEN** the embed handler formats the Ollama error
- **THEN** the OAuth token MUST be redacted from the
  error body

#### Scenario: Start and serve

- **GIVEN** Vertex AI OAuth is configured (`gcloud`
  authenticated, project ID set)
- **WHEN** `uf ollama-proxy` is called
- **THEN** the proxy MUST start on port 11434 with
  a health endpoint at `GET /health`

#### Scenario: Detach mode

- **GIVEN** `--detach` flag is set
- **WHEN** `uf ollama-proxy --detach` is called
- **THEN** the proxy MUST start in the background,
  write PID file, redirect logs to
  `.uf/ollama-proxy.log`, and the parent process MUST
  poll health and exit once healthy

#### Scenario: Already running

- **GIVEN** a proxy instance is already running on
  the configured port
- **WHEN** `uf ollama-proxy` is called
- **THEN** it MUST report the existing instance and
  exit without error

### Requirement: TokenManager in `internal/auth/`

A `TokenManager` struct MUST be extracted to
`internal/auth/` encapsulating the full Vertex AI
OAuth lifecycle:
- Token storage with `sync.RWMutex`
- Expiry tracking
- Background refresh loop (context-cancellable)
- Proactive refresh within 5 min of expiry with
  `TryLock()` deduplication
- Atomic token invalidation on background failure
- `RefreshVertexToken(execCmd) (string, error)`
- `RefreshBedrockCredentials` and helpers

The gateway MUST instantiate `TokenManager` instead
of reimplementing token lifecycle. Behavior MUST be
unchanged.

#### Scenario: Gateway still works after extraction

- **GIVEN** the gateway's token management has been
  replaced with `auth.TokenManager`
- **WHEN** all existing gateway tests are run
- **THEN** they MUST pass without modification
  (except import paths)

#### Scenario: Background refresh failure

- **GIVEN** a background token refresh fails
- **WHEN** the next embed request arrives
- **THEN** the proxy MUST return a clear
  re-authentication error (not forward a stale token)

#### Scenario: gcloud not installed

- **GIVEN** `gcloud` is not in PATH
- **WHEN** the proxy starts
- **THEN** it MUST return an error with install
  instructions

### Requirement: PID file in `internal/pidfile/`

PID file management functions (`WritePID`, `ReadPID`,
`IsAlive`, `CleanupStale`, `RemovePID`, `PIDInfo`)
MUST be extracted from `internal/gateway/pid.go` to
`internal/pidfile/`. Both gateway and ollama-proxy
MUST import from `internal/pidfile/`.

#### Scenario: Gateway PID still works after extraction

- **GIVEN** PID functions have been moved to
  `internal/pidfile/`
- **WHEN** all existing gateway tests are run
- **THEN** they MUST pass without modification
  (except import paths)

### Requirement: Configuration

`OllamaProxyConfig` MUST be added to the unified
config with fields:

- `Port int` (default 11434)
- `EmbedModel string` (default "text-embedding-005")
- `GatewayURL string` (default "http://localhost:53147")

Environment variable overrides:
- `UF_OLLAMA_PROXY_PORT`
- `UF_OLLAMA_PROXY_EMBED_MODEL`
- `UF_OLLAMA_PROXY_GATEWAY_URL`

#### Scenario: Config defaults

- **GIVEN** no config file and no env vars
- **WHEN** the proxy starts
- **THEN** it MUST use port 11434, embed model
  `text-embedding-005`, gateway URL
  `http://localhost:53147`

### Requirement: Health endpoint

`GET /health` MUST return JSON status:
```json
{
  "service": "uf-ollama-proxy",
  "status": "ok",
  "port": 11434,
  "embed_model": "text-embedding-005",
  "gateway_available": true
}
```

The `service` field identifies this as the proxy (not
real Ollama) for `uf doctor` and `uf ollama-proxy
status` to distinguish. The `gateway_available` field
reflects whether the gateway health check succeeded at
the most recent probe.

### Requirement: Cobra command

`uf ollama-proxy` MUST be registered as a top-level
command with:
- Default action: start the proxy
- `stop` subcommand
- `status` subcommand
- Flags: `--port`, `--embed-model`, `--gateway-url`,
  `--detach`

## MODIFIED Requirements

### Requirement: Gateway token management extraction

`internal/gateway/` token lifecycle (token storage,
expiry tracking, refresh loop, proactive refresh,
atomic invalidation) MUST be replaced with
`auth.TokenManager`. `internal/gateway/pid.go`
functions MUST be replaced with `pidfile.*` imports.

Previously: token lifecycle was implemented inline in
`VertexProvider` and `BedrockProvider`. PID functions
were in `internal/gateway/pid.go`.

### Requirement: Unified config

`Config` struct MUST include a new `OllamaProxy`
field of type `OllamaProxyConfig`. The `merge()`,
`applyEnvOverrides()`, and config template MUST be
updated.

Previously: no ollama-proxy section existed.

## REMOVED Requirements

None.
<!-- scaffolded by uf vdev -->
