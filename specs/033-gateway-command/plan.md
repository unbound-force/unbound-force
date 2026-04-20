# Implementation Plan: LLM Gateway Command

**Branch**: `033-gateway-command` | **Date**: 2026-04-20 | **Spec**: `specs/033-gateway-command/spec.md`
**Input**: Feature specification from `specs/033-gateway-command/spec.md`

## Summary

Add a `uf gateway` command — a minimal reverse proxy
that runs on the host machine and serves the Anthropic
Messages API on port 53147. The gateway auto-detects
the cloud provider (Anthropic, Vertex AI, Bedrock)
from environment variables, injects host-side
credentials into upstream requests, and auto-refreshes
OAuth tokens for Vertex AI. The sandbox auto-starts
the gateway when a provider is detected, passing the
gateway URL to the container and eliminating credential
file mounts. The implementation uses Go stdlib
`net/http/httputil.ReverseProxy` for the proxy core
and the Strategy pattern (matching
`internal/sandbox/backend.go`) for provider
abstraction.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: `net/http`, `net/http/httputil` (stdlib), `github.com/spf13/cobra` (CLI), `github.com/charmbracelet/log` (logging)
**Storage**: PID file at `.uf/gateway.pid` (plain text)
**Testing**: Standard library `testing` package, `t.TempDir()` for filesystem tests, injected dependencies for all external calls
**Target Platform**: macOS (darwin/arm64, darwin/amd64), Linux (linux/amd64, linux/arm64)
**Project Type**: CLI (extension to existing `uf` binary)
**Performance Goals**: < 50ms proxy latency overhead (SC-005)
**Constraints**: < 5MB binary size increase (SC-006), stdlib HTTP only (no external HTTP client deps), no AWS SDK
**Scale/Scope**: Single-user local proxy, one provider at a time, ~4 new source files + 2 modified files

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Research Check

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. Autonomous Collaboration | ✅ PASS | The gateway operates independently — it is a standalone process that produces a machine-parseable health endpoint (JSON). The sandbox detects the gateway via health check, not runtime coupling. When the gateway is unavailable, the sandbox falls back to credential mounts (no blocking dependency). |
| II. Composability First | ✅ PASS | The gateway is independently usable (`uf gateway` works without the sandbox). The sandbox auto-detects the gateway and activates enhanced functionality when it's available (composable). The gateway does not require the sandbox as a prerequisite. |
| III. Observable Quality | ✅ PASS | The health endpoint returns JSON with provider, port, PID, and uptime. `uf gateway status` provides human-readable output. Error responses include structured JSON error bodies. |
| IV. Testability | ✅ PASS | All external dependencies (exec, env, HTTP, filesystem) are injected via function fields on `Options`, following the established pattern from `internal/sandbox/sandbox.go`. The `Provider` interface enables mock providers in tests. PID file operations use injected `ReadFile`/`WriteFile`. |

**Gate result**: PASS — proceed to Phase 0 research.

### Post-Design Re-Check

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. Autonomous Collaboration | ✅ PASS | The gateway produces a JSON health response (artifact). The sandbox consumes it via HTTP GET (artifact-based communication). No synchronous inter-hero coupling. The gateway does not depend on any other hero. |
| II. Composability First | ✅ PASS | Three independent usage modes: (1) `uf gateway` standalone, (2) `uf sandbox start` auto-starts gateway, (3) sandbox without gateway (fallback). Each mode works independently. The `Provider` interface is an extension point for future providers. |
| III. Observable Quality | ✅ PASS | Health endpoint is JSON (machine-parseable). Status command provides human-readable output. Error responses are structured. All output includes enough context for debugging (provider name, port, PID). |
| IV. Testability | ✅ PASS | Coverage strategy defined below. All functions are testable in isolation via dependency injection. No global state. No external service requirements for unit tests. Provider implementations are mockable via the interface. |

**Gate result**: PASS — proceed to implementation.

### Coverage Strategy

| Layer | Target | Approach |
|-------|--------|----------|
| Unit (gateway core) | 90%+ | Test `Start()`, `Stop()`, `Status()` with injected deps. Mock HTTP server for health endpoint. Mock `ExecCmd` for background process. |
| Unit (providers) | 85%+ | Test `PrepareRequest()` for each provider with mock requests. Test `Start()` with mock `ExecCmd` (gcloud, aws). Test token refresh with mock ticker. |
| Unit (PID file) | 95%+ | Test `WritePID`, `ReadPID`, `IsAlive`, `RemovePID`, stale detection with `t.TempDir()`. |
| Unit (sandbox integration) | 85%+ | Test gateway auto-start in `Start()`. Test `buildRunArgs()` with gateway active vs. inactive. Test fallback behavior. |
| Contract | N/A | Health endpoint response format validated by unit tests. Provider interface contract enforced by type system. |
| Integration | Manual | Quickstart verification steps (see `quickstart.md`). |

**Ratchet**: Coverage must not decrease from the
current baseline for `internal/sandbox/`. New
`internal/gateway/` package must meet the targets
above.

## Project Structure

### Documentation (this feature)

```text
specs/033-gateway-command/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0: research findings
├── data-model.md        # Phase 1: entity definitions
├── quickstart.md        # Phase 1: verification steps
├── contracts/
│   └── gateway-api.md   # Gateway HTTP API contract
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
internal/gateway/
├── gateway.go           # Gateway struct, Start(), Serve(),
│                        # health endpoint, provider detection,
│                        # background mode (detach/re-exec)
├── provider.go          # Provider interface + 3 implementations
│                        # (AnthropicProvider, VertexProvider,
│                        # BedrockProvider)
├── refresh.go           # Token refresh goroutine for Vertex/
│                        # Bedrock, SigV4 signing helper
├── pid.go               # PID file management (write, read,
│                        # alive check, stale cleanup)
└── gateway_test.go      # Tests for all gateway functions

cmd/unbound-force/
├── main.go              # Modified: add newGatewayCmd()
└── gateway.go           # New: Cobra command (gateway,
                         # gateway stop, gateway status)

internal/sandbox/
├── sandbox.go           # Modified: gateway auto-start in
│                        # Start(), gateway stop in Stop()
├── config.go            # Modified: gateway-aware buildRunArgs(),
│                        # gateway-aware forwardedEnvVars()
└── sandbox_test.go      # Modified: new tests for gateway
                         # integration paths
```

**Structure Decision**: New `internal/gateway/` package
follows the established pattern of domain packages
under `internal/` (matching `internal/sandbox/`,
`internal/doctor/`, `internal/setup/`). The gateway
package is self-contained with its own `Options` struct
and injectable dependencies. The sandbox integration
modifies existing files in `internal/sandbox/` to add
gateway detection and auto-start logic.

## Implementation Phases

### Phase 1: Gateway Core (US1, US5)

Build the gateway process, provider detection, health
endpoint, PID file management, and background mode.
This phase delivers a working `uf gateway` command
that can start, serve health checks, run in the
background, and be stopped.

**Files**: `internal/gateway/gateway.go`,
`internal/gateway/pid.go`,
`cmd/unbound-force/gateway.go`,
`cmd/unbound-force/main.go`

**Depends on**: Nothing (standalone)

### Phase 2: Provider Implementations (US1, US3, US4)

Implement the three provider strategies (Anthropic,
Vertex, Bedrock) with credential injection, URL
rewriting, and token refresh.

**Files**: `internal/gateway/provider.go`,
`internal/gateway/refresh.go`

**Depends on**: Phase 1 (gateway core must exist)

### Phase 3: Sandbox Integration (US2)

Modify the sandbox to auto-start the gateway, pass
the gateway URL to the container, and skip credential
mounts when the gateway is active.

**Files**: `internal/sandbox/sandbox.go`,
`internal/sandbox/config.go`

**Depends on**: Phase 1 + Phase 2 (gateway must be
functional)

### Phase 4: Tests and Polish (All US)

Comprehensive test coverage for all new and modified
code. Edge case handling, error messages, and
documentation updates.

**Files**: `internal/gateway/gateway_test.go`,
`internal/sandbox/sandbox_test.go`

**Depends on**: Phase 1 + Phase 2 + Phase 3

## Key Design Decisions

### D1: Stdlib-Only HTTP

**Decision**: Use `net/http/httputil.ReverseProxy` for
the proxy core. No external HTTP client libraries.

**Rationale**: The stdlib `ReverseProxy` handles
streaming (SSE), header forwarding, and error
propagation natively. Adding an external HTTP library
would increase the binary size and dependency surface
without meaningful benefit. The `Director` function
provides the hook needed for credential injection.

**Alternatives rejected**: `caddy` (too heavy),
`traefik` (too heavy), custom HTTP client (unnecessary
when stdlib handles streaming).

### D2: Shell-Out for Cloud Credentials

**Decision**: Use `gcloud auth application-default
print-access-token` for Vertex AI tokens and
`aws configure export-credentials` for Bedrock
credentials. No Google Cloud SDK or AWS SDK Go
dependencies.

**Rationale**: Keeps the binary small (SC-006: < 5MB
increase). The cloud CLIs are already prerequisites
for developers using these providers. The `Provider`
interface abstracts the credential source — a future
spec can add SDK-based providers without changing the
gateway core.

**Alternatives rejected**: `golang.org/x/oauth2/google`
(adds ~2MB transitive deps), `aws-sdk-go-v2` (adds
~5MB transitive deps).

### D3: Re-Exec for Background Mode

**Decision**: Use `os/exec.Command` to re-launch the
binary with a sentinel env var (`_UF_GATEWAY_CHILD=1`)
for `--detach` mode. Not `syscall.Fork`.

**Rationale**: Go's runtime is multi-threaded; forking
is unsafe. The re-exec pattern is used by Docker,
Caddy, and other Go tools. The sentinel env var is
testable via the injected `Getenv` function.

### D4: Strategy Pattern for Providers

**Decision**: Use the `Provider` interface with three
implementations, matching the `Backend` interface
pattern in `internal/sandbox/backend.go`.

**Rationale**: SOLID Open/Closed Principle. Adding a
new provider (e.g., OpenAI-compatible) requires only a
new `Provider` implementation, not modification of the
gateway core. The pattern is already established in the
codebase.

### D5: Gateway-Aware Sandbox (Conditional)

**Decision**: The sandbox detects the gateway via
health check before starting the container. If the
gateway is available, it passes `ANTHROPIC_BASE_URL`
and skips credential mounts. If unavailable, it falls
back to the existing behavior.

**Rationale**: Backward compatibility (FR-012). The
gateway is an enhancement, not a requirement. Existing
sandbox users who don't use the gateway see no behavior
change.

### D6: Localhost-Only, No Auth

**Decision**: The gateway listens on `127.0.0.1` only
and does not require inbound authentication (FR-013).

**Rationale**: The gateway is a local development tool.
Binding to localhost prevents network exposure. Adding
authentication would complicate the setup without
security benefit (the threat model is local-only).

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| `gcloud` token refresh fails silently | Medium | High | Clear error message to client (US3 scenario 2). Log refresh failures. |
| Port 53147 conflict with another tool | Low | Medium | `--port` override flag. Clear error message identifying the conflict. |
| Stale PID file after crash | Medium | Low | Health check + stale cleanup on next start (spec edge case). |
| Streaming response buffering | Low | High | `httputil.ReverseProxy` handles streaming natively. Verify with SSE test. |
| Binary size exceeds 5MB budget | Low | Medium | No external HTTP/cloud SDK deps. Monitor with `go build` size check. |
| SigV4 signing implementation bugs | Medium | Medium | Well-documented algorithm. Test against known AWS test vectors. |

## Complexity Tracking

> No constitution violations to justify.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
