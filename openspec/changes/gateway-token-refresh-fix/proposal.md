## Why

The `uf gateway` Vertex AI provider acquires an OAuth
token at startup via `gcloud auth application-default
print-access-token` and refreshes every 50 minutes.
When a refresh fails (expired ADC credentials, network
change, gcloud error), the gateway continues forwarding
requests with the stale, expired token. Vertex AI
rejects these with `ACCESS_TOKEN_TYPE_UNSUPPORTED` (401
UNAUTHENTICATED).

The user sees a cryptic Google error with no indication
that the gateway's token has expired or that
re-authentication is needed. Two compounding problems
make diagnosis difficult:

1. **Stale token forwarding**: `PrepareRequest` only
   checks for an empty token, not an expired one. After
   a failed refresh, the old (expired) token remains
   non-empty, so the check passes and the bad token is
   sent to Vertex.

2. **Silent log loss**: In detached mode (the default
   when started via sandbox), child process stdout/stderr
   are set to `nil`. All `charmbracelet/log` output --
   including "vertex token refresh failed" errors -- is
   discarded. The user has no way to see that refresh
   failed.

This is a pattern previously seen in the global-region
silent fallback (PR #101, learning `gateway-2`), where
a silent failure produced confusing 401 errors. The fix
there was to return an explicit error rather than
silently misrouting. This change applies the same
principle to token lifecycle management.

## What Changes

### Token Expiry Tracking and Stale Token Invalidation

Add a `tokenExpiry` timestamp to `VertexProvider` (and
analogous `credExpiry` to `BedrockProvider`). On each
successful token acquisition, set expiry to
`now + 55 minutes` (5-minute safety margin before the
60-minute gcloud token lifetime). In `PrepareRequest`,
check expiry before forwarding -- if the token has
expired, return a clear error message instead of
forwarding a stale credential.

When a refresh fails, clear the stored token so that
`PrepareRequest` returns the "Re-authenticate" error
immediately rather than continuing to forward a token
that Vertex will reject.

### Gateway Log File in Detached Mode

In the `detach()` function, redirect child process
stdout/stderr to `.uf/gateway.log` instead of
discarding them. This makes refresh failures, upstream
errors, and provider diagnostics visible for
troubleshooting.

### Proactive Token Refresh on Expiry Detection

When `PrepareRequest` detects that the current token
is within 5 minutes of expiry, attempt a synchronous
token refresh before the request proceeds. This
reduces the window where requests fail between the
50-minute refresh cycle and actual token expiry.

## Capabilities

### New Capabilities
- `token-expiry-tracking`: Gateway tracks token
  acquisition time and rejects requests with a clear
  error when the token has expired, instead of
  forwarding stale credentials.
- `gateway-log-file`: Detached gateway writes logs to
  `.uf/gateway.log`, making refresh failures and
  provider errors visible for troubleshooting.
- `proactive-refresh`: `PrepareRequest` attempts a
  synchronous token refresh when the token is near
  expiry, reducing the failure window.

### Modified Capabilities
- `VertexProvider.PrepareRequest`: Now checks token
  expiry and attempts proactive refresh before
  forwarding requests.
- `BedrockProvider.PrepareRequest`: Analogous expiry
  check for AWS session credentials.
- `detach()`: Redirects child stdout/stderr to log
  file instead of discarding.
- `uf gateway status`: Displays log file path when
  gateway is running.

### Removed Capabilities
- None.

## Impact

### Files Modified

| File | Change |
|------|--------|
| `internal/gateway/provider.go` | Add `tokenExpiry`/`credExpiry` fields, `tokenRefreshing`/`credRefreshing` mutexes, expiry checks and `tryProactiveRefresh()` in `PrepareRequest`, clear token on refresh failure, named constants |
| `internal/gateway/refresh.go` | Add `refreshVertexTokenWithTimeout` helper for context-bounded refresh |
| `internal/gateway/gateway.go` | Redirect detached child to `.uf/gateway.log`, display log path in status |
| `internal/gateway/gateway_test.go` | 15 test functions for expired-token rejection, proactive refresh, concurrency dedup, log file, foreground negative case |

### Behavioral Changes

- Requests that would previously silently fail with a
  Vertex 401 now fail fast with a gateway-level error
  message: "vertex AI token expired. Re-authenticate:
  gcloud auth application-default login"
- The `.uf/gateway.log` file is created when gateway
  runs in detached mode (new file in `.uf/`)
- Bedrock provider gains the same expiry protection
  (session tokens also expire)

### No Breaking Changes

- Foreground gateway behavior unchanged (logs still go
  to stderr)
- Token refresh interval (50 minutes) unchanged
- Gateway API surface (routes, response format)
  unchanged
- Sandbox integration unchanged

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

This change is internal to the gateway package. It does
not affect inter-hero artifact formats, communication
protocols, or artifact envelope schemas. The gateway
remains a transparent proxy -- heroes never interact
with it directly.

### II. Composability First

**Assessment**: PASS

The gateway remains independently usable (`uf gateway
start`) and the sandbox integration is unchanged. No
new mandatory dependencies are introduced. The log file
is created only in detached mode and does not affect
standalone gateway operation.

### III. Observable Quality

**Assessment**: PASS

This change directly improves observability:
- Expired-token errors are surfaced as structured JSON
  error responses (matching the existing
  `writeJSONError` format) instead of opaque Vertex 401s
- Gateway logs are preserved in `.uf/gateway.log`
  instead of being discarded
- `uf gateway status` displays the log file path

### IV. Testability

**Assessment**: PASS

All changes follow the existing injectable-function
pattern (`ExecCmd`, `Getenv`). Token expiry is
testable by injecting a known expiry time. Log file
creation is testable via the existing `ExecStart`
injection point. No external services required for
testing.
