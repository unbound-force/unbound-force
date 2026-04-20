# Research: LLM Gateway Command

**Branch**: `033-gateway-command` | **Date**: 2026-04-20
**Spec**: `specs/033-gateway-command/spec.md`

## Phase 0: Research Findings

### R1 — `net/http/httputil.ReverseProxy` Behavior

**Question**: How does Go's stdlib `ReverseProxy`
handle request/response forwarding, and what hooks
does it expose for credential injection?

**Finding**: `httputil.ReverseProxy` is the stdlib
reverse proxy. Key properties:

- **Director function**: Called before each request is
  forwarded. Receives the outbound `*http.Request` and
  can modify headers, URL, host, etc. This is the hook
  for credential injection (add `Authorization` header,
  `x-api-key` header, or AWS SigV4 signature).

- **ModifyResponse function**: Called after the upstream
  response is received but before it's written to the
  client. Can be used for logging or response
  transformation (we don't need transformation — FR-014
  requires forwarding errors as-is).

- **ErrorHandler function**: Called when the upstream
  connection fails. Returns a custom error response to
  the client.

- **Transport field**: Accepts a custom `http.RoundTripper`
  for the upstream connection. Default is
  `http.DefaultTransport`. Can be used for TLS
  configuration or request-level middleware.

- **Streaming**: `ReverseProxy` handles streaming
  responses (SSE, chunked transfer) natively. It copies
  the response body using `io.Copy` and flushes if the
  `ResponseWriter` implements `http.Flusher`. This is
  critical for Claude's streaming API responses.

- **Hop-by-hop headers**: `ReverseProxy` strips
  hop-by-hop headers (`Connection`, `Keep-Alive`,
  `Transfer-Encoding`, etc.) per RFC 2616 §13.5.1.
  Application headers like `anthropic-beta`,
  `anthropic-version`, and `X-Claude-Code-Session-Id`
  are preserved automatically.

- **X-Forwarded-For**: `ReverseProxy` appends the
  client's IP to `X-Forwarded-For` by default. This is
  fine for our use case (localhost only).

**Design implication**: Use `Director` for credential
injection. Each provider implements a `Director`-
compatible function that modifies the outbound request.
The `ReverseProxy` handles streaming, header forwarding,
and error propagation natively.

**Code pattern**:

```go
proxy := &httputil.ReverseProxy{
    Director: func(req *http.Request) {
        provider.PrepareRequest(req)
    },
}
```

### R2 — Vertex AI OAuth Token Refresh

**Question**: How do Vertex AI OAuth tokens work, and
how should the gateway refresh them?

**Finding**: Vertex AI uses Google Cloud's Application
Default Credentials (ADC) chain. The token lifecycle:

1. **Initial acquisition**: Call
   `google.FindDefaultCredentials()` or shell out to
   `gcloud auth application-default print-access-token`
   to get an OAuth2 access token.

2. **Token format**: Bearer token, typically valid for
   3600 seconds (1 hour). The response includes an
   `expires_in` field or the token source tracks expiry
   internally.

3. **Refresh strategy**: Two approaches:
   - **Go SDK approach**: Use
     `golang.org/x/oauth2/google.FindDefaultCredentials`
     which returns a `TokenSource` that auto-refreshes.
     However, this adds a dependency on `golang.org/x/oauth2`
     and `cloud.google.com/go/auth`.
   - **Shell-out approach**: Call
     `gcloud auth application-default print-access-token`
     periodically. Simpler, no new dependencies, but
     requires `gcloud` CLI on the host.

4. **FR-005 requirement**: Refresh at least 10 minutes
   before expiry. With 1-hour tokens, refresh at the
   50-minute mark.

**Design decision**: Use the **shell-out approach** for
v1. Rationale:
- No new Go module dependencies (per project convention
  of minimal dependencies — see `go.mod`)
- `gcloud` is already a prerequisite for Vertex AI users
- Simpler implementation: goroutine with ticker, calls
  `gcloud auth application-default print-access-token`
- Token stored in a `sync.RWMutex`-protected field
- If `gcloud` fails, the gateway logs the error and
  returns a clear message to the client (US3 scenario 2)

**Refresh goroutine pattern**:

```go
func (p *VertexProvider) refreshLoop(ctx context.Context) {
    ticker := time.NewTicker(50 * time.Minute)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            p.refreshToken()
        }
    }
}
```

**Future enhancement**: If the shell-out approach proves
unreliable (e.g., `gcloud` not in PATH in some
environments), a follow-up spec can add the Go SDK
approach as a fallback. The `Provider` interface
abstracts this — swapping the implementation requires
no changes to the gateway core.

### R3 — AWS SigV4 Signing for Bedrock

**Question**: How does AWS Bedrock authentication work,
and can we sign requests without the AWS SDK?

**Finding**: Bedrock uses AWS Signature Version 4
(SigV4) for request authentication. The signing process:

1. **Credential chain**: AWS credentials are resolved
   from environment variables (`AWS_ACCESS_KEY_ID`,
   `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`),
   shared credentials file (`~/.aws/credentials`),
   or IAM role (EC2/ECS metadata).

2. **SigV4 signing**: Each request must be signed with
   the AWS access key, secret key, region, and service
   name (`bedrock-runtime`). The signature is placed in
   the `Authorization` header.

3. **Implementation options**:
   - **AWS SDK for Go v2**: `github.com/aws/aws-sdk-go-v2`
     provides `credentials.NewDefaultCredentials()` and
     `signer.SignHTTP()`. Adds ~5MB to the binary.
   - **Shell-out to `aws`**: Call
     `aws sts get-caller-identity` to verify credentials,
     then manually construct SigV4 signatures. Complex
     and error-prone.
   - **Minimal SigV4 implementation**: Implement the
     SigV4 algorithm directly (~200 lines). The algorithm
     is well-documented (AWS docs) and stable. No
     external dependencies.

4. **Credential refresh**: AWS session tokens from
   `aws sso login` or `aws sts assume-role` expire
   (typically 1-12 hours). The credential chain
   re-reads from the environment/files on each call,
   so refresh is automatic if the developer runs
   `aws sso login` again.

**Design decision**: Use **shell-out to `aws configure
export-credentials`** for v1. This command outputs
JSON with `AccessKeyId`, `SecretAccessKey`,
`SessionToken`, and `Expiration`. Rationale:
- No AWS SDK dependency (keeps binary small per SC-006)
- `aws` CLI is already a prerequisite for Bedrock users
- Credential refresh is handled by re-calling the
  command before expiry
- SigV4 signing can use a minimal stdlib implementation
  (~200 lines of `crypto/hmac` + `crypto/sha256`)

**Alternative considered**: Using `aws-sdk-go-v2` was
rejected because it would add ~5MB to the binary
(violating SC-006's 5MB budget) and introduce a large
transitive dependency tree. The minimal SigV4
implementation is well-understood and testable.

### R4 — PID File Management

**Question**: What are the best practices for PID file
management in Go CLI tools?

**Finding**: PID file conventions:

1. **Location**: `.uf/gateway.pid` (consistent with
   `.uf/` convention from Spec 025). The `.uf/`
   directory is already in `.gitignore`.

2. **Format**: Plain text file containing the PID as a
   decimal integer, followed by a newline. Optionally
   include metadata (port, provider) on subsequent
   lines for `uf gateway status`.

3. **Write**: Use `os.WriteFile` with mode `0644`.
   Write atomically (write to temp file, rename) to
   prevent partial reads.

4. **Read**: Parse the PID, then verify the process is
   alive with `os.FindProcess(pid)` + `process.Signal(0)`.
   On Unix, `Signal(0)` returns nil if the process
   exists and the caller has permission to signal it.

5. **Stale detection**: If the PID file exists but the
   process is not alive, the PID file is stale. Clean
   it up and proceed (per spec edge case: "stale PID
   file is cleaned up").

6. **Cleanup**: Remove the PID file on graceful
   shutdown (SIGINT, SIGTERM). Use `signal.NotifyContext`
   to handle signals and clean up.

7. **Race condition**: Two `uf gateway` invocations
   could race to write the PID file. Mitigate with
   `os.OpenFile` with `O_CREATE|O_EXCL` (fails if file
   exists). Check health endpoint first to detect
   already-running gateway.

**Design pattern**:

```go
type PIDFile struct {
    Path     string
    PID      int
    Port     int
    Provider string
}

func WritePID(path string, pid, port int, provider string) error
func ReadPID(path string) (*PIDFile, error)
func IsAlive(pid int) bool
func CleanupStale(path string) error
```

### R5 — Background Process Forking in Go

**Question**: How should `--detach` work in Go? Can we
use `syscall.Fork`?

**Finding**: Go does not support `syscall.Fork` safely
because the Go runtime is multi-threaded (goroutines,
GC, scheduler). Forking a multi-threaded process leads
to undefined behavior.

**Recommended approach**: Re-exec the binary with a
sentinel environment variable:

1. When `--detach` is specified, the parent process:
   - Sets a sentinel env var (e.g.,
     `_UF_GATEWAY_CHILD=1`)
   - Calls `os/exec.Command` with the same binary and
     args (minus `--detach`)
   - Sets `cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}`
     to detach from the terminal session (Unix only)
   - Starts the child process
   - Waits briefly for the health endpoint
   - Prints the PID and exits

2. When the child process starts:
   - Detects `_UF_GATEWAY_CHILD=1`
   - Runs the gateway server (blocking)
   - Writes the PID file
   - Handles SIGINT/SIGTERM for graceful shutdown

**This is the same pattern used by**:
- Docker daemon (`dockerd`)
- Caddy server (`caddy start`)
- Many Go CLI tools with background mode

**Code pattern**:

```go
func detach(args []string) (int, error) {
    exe, _ := os.Executable()
    cmd := exec.Command(exe, args...)
    cmd.Env = append(os.Environ(), "_UF_GATEWAY_CHILD=1")
    cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
    if err := cmd.Start(); err != nil {
        return 0, err
    }
    return cmd.Process.Pid, nil
}
```

**Testability**: The `ExecCmd` injection pattern on
`Options` (established in Spec 028) allows tests to
mock the re-exec. The sentinel env var check is a
simple `os.Getenv` call that can be injected via
`Getenv` on `Options`.

### R6 — Anthropic Messages API Endpoint Format

**Question**: What is the exact URL format for each
provider's upstream endpoint?

**Finding**: Based on the Claude Code LLM gateway
specification:

1. **Direct Anthropic**:
   - Base URL: `https://api.anthropic.com`
   - Endpoints: `/v1/messages`, `/v1/messages/count_tokens`
   - Auth: `x-api-key: <ANTHROPIC_API_KEY>` header
   - Required headers forwarded: `anthropic-beta`,
     `anthropic-version`

2. **Vertex AI**:
   - Base URL:
     `https://{CLOUD_ML_REGION}-aiplatform.googleapis.com/v1/projects/{PROJECT_ID}/locations/{REGION}/publishers/anthropic/models/{MODEL}:rawPredict`
   - The gateway receives requests in Anthropic Messages
     format and must rewrite the URL to the Vertex
     rawPredict endpoint.
   - Auth: `Authorization: Bearer <oauth-token>` header
   - Required headers forwarded: `anthropic-beta`,
     `anthropic-version`
   - Model extracted from request body or defaulted to
     `claude-sonnet-4-20250514`

3. **Bedrock**:
   - Base URL:
     `https://bedrock-runtime.{REGION}.amazonaws.com/model/{MODEL}/invoke`
   - Auth: AWS SigV4 signature in `Authorization` header
   - The `anthropic_beta` and `anthropic_version` fields
     are in the request body (not headers) for Bedrock
   - Model extracted from request body or defaulted

**Design implication**: The `Provider` interface needs a
`PrepareRequest(*http.Request)` method that handles URL
rewriting, header injection, and body transformation
specific to each provider. The gateway core routes
`/v1/messages` and `/v1/messages/count_tokens` to the
appropriate provider.

### R7 — Container-to-Host Communication

**Question**: How does the container reach the gateway
running on the host?

**Finding**: Podman provides `host.containers.internal`
as a hostname that resolves to the host machine's IP
from within a container. This is already established in
Spec 028 (used for Ollama at
`host.containers.internal:11434`).

- The gateway listens on `0.0.0.0:53147` (all
  interfaces) so it's reachable from both localhost and
  the container network.
- The sandbox passes
  `ANTHROPIC_BASE_URL=http://host.containers.internal:53147`
  to the container.
- Claude Code inside the container reads
  `ANTHROPIC_BASE_URL` and sends requests to the
  gateway instead of directly to the provider.

**Port 53147**: Chosen to avoid conflicts with common
development ports (3000, 4096, 8080, etc.). The number
is arbitrary but fixed by default, with `--port` as an
override.

### R8 — `ANTHROPIC_AUTH_TOKEN` for Gateway Auth

**Question**: How does Claude Code authenticate with
the gateway?

**Finding**: Per the Claude Code LLM gateway
specification, Claude Code sends authentication via
the `Authorization` header (from `ANTHROPIC_AUTH_TOKEN`)
or the `x-api-key` header (from `ANTHROPIC_API_KEY`).

For the gateway:
- The gateway does NOT require authentication from
  clients (FR-013) since it listens only on localhost.
- The sandbox sets `ANTHROPIC_AUTH_TOKEN=gateway` as a
  placeholder value so Claude Code includes an
  `Authorization` header (some Claude Code versions
  require a non-empty auth token).
- The gateway ignores the inbound `Authorization` and
  `x-api-key` headers — it injects its own credentials
  for the upstream provider.

### R9 — Graceful Shutdown and Signal Handling

**Question**: How should the gateway handle shutdown
signals?

**Finding**: Use `signal.NotifyContext` (Go 1.16+) to
create a context that cancels on SIGINT or SIGTERM:

```go
ctx, stop := signal.NotifyContext(context.Background(),
    syscall.SIGINT, syscall.SIGTERM)
defer stop()
```

On cancellation:
1. Stop the token refresh goroutine (via context)
2. Call `http.Server.Shutdown(ctx)` for graceful
   HTTP shutdown (drains in-flight requests)
3. Remove the PID file
4. Exit cleanly

The `http.Server.Shutdown` method waits for active
connections to complete, which is important for
long-running streaming responses.

## Summary of Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Proxy engine | `httputil.ReverseProxy` | Stdlib, handles streaming, minimal code |
| Vertex auth | Shell out to `gcloud` | No new dependencies, `gcloud` already required |
| Bedrock auth | Shell out to `aws` + minimal SigV4 | No AWS SDK, keeps binary small |
| Background mode | Re-exec with sentinel env var | Go doesn't support safe `fork()` |
| PID file format | Text with PID + metadata | Simple, human-readable, easy to parse |
| Token refresh | Background goroutine with ticker | Non-blocking, context-cancellable |
| Provider interface | Strategy pattern | Matches `internal/sandbox/backend.go` pattern |
