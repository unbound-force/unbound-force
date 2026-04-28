## Context

The `uf gateway` Vertex AI provider (Spec 033) acquires
OAuth tokens via `gcloud auth application-default
print-access-token` and refreshes every 50 minutes. Two
defects cause silent authentication failures:

1. When token refresh fails, the stale expired token
   remains in `p.token` and is forwarded to Vertex,
   producing a 401 `ACCESS_TOKEN_TYPE_UNSUPPORTED`.
2. In detached mode, child process stdout/stderr are
   `nil`, so refresh failure logs are discarded.

This follows the same "silent failure" anti-pattern
that was fixed in the global-region fallback (PR #101,
learning `gateway-2`). The fix applies the same
principle: fail explicitly with actionable guidance
rather than silently forwarding bad data.

## Goals / Non-Goals

### Goals
- Expired tokens produce a clear, actionable error
  at the gateway level (not a cryptic Vertex 401)
- Failed token refreshes invalidate the stored token
  immediately, preventing stale token forwarding
- Detached gateway logs are preserved in a log file
  for post-hoc troubleshooting
- Proactive refresh reduces the failure window when
  tokens are near expiry
- Bedrock provider gets analogous expiry protection
- All new behavior is testable via existing injection
  points (Principle IV)

### Non-Goals
- Automatic re-authentication (running `gcloud auth`
  interactively is out of scope -- the gateway is
  headless)
- Log rotation or log management (`.uf/gateway.log`
  is overwritten on each gateway start; log rotation
  is a future concern)
- Token caching across gateway restarts (tokens are
  short-lived; re-acquiring on startup is correct)
- Monitoring or alerting integration (the error
  response is sufficient for v1)

## Decisions

### D1: Token Expiry Tracking via Timestamp Field

Add `tokenExpiry time.Time` to `VertexProvider` and
`credExpiry time.Time` to `BedrockProvider`. Set on
each successful token acquisition to
`time.Now().Add(55 * time.Minute)` (5-minute safety
margin before the 60-minute gcloud token lifetime).

**Rationale**: Simpler than parsing token JWTs to
extract `exp` claims. The 55-minute window is
conservative -- gcloud tokens are valid for 60 minutes
from issuance. The safety margin prevents edge cases
where a token expires during an in-flight request.

**Testability**: Time-based expiry is testable by
setting `tokenExpiry` directly in test fixtures or
by injecting a clock function. The existing
injectable `ExecCmd` pattern supports this without
new injection points.

### D2: Clear Token on Refresh Failure

When `refreshLoop`'s closure fails to acquire a new
token, set `p.token = ""` (Vertex) or clear
`accessKey`/`secretKey` (Bedrock). This causes
`PrepareRequest` to return the "Re-authenticate"
error on the next request instead of forwarding a
stale credential.

**Rationale**: The current behavior -- logging the
error and continuing with the old token -- was
designed for transient failures (e.g., a momentary
gcloud CLI hang). But for ADC credential expiry,
the old token is permanently invalid. Clearing it
forces an explicit error that tells the user what
to do.

**Trade-off**: If a refresh fails transiently but
the old token is still valid (within its 60-minute
window), clearing it causes an unnecessary error.
However, this window is small (between the 50-minute
refresh and 60-minute expiry), and the error message
is actionable. The user can restart the gateway to
immediately re-acquire a token.

### D3: Proactive Refresh in PrepareRequest

When `PrepareRequest` detects that the token is
within 5 minutes of expiry, attempt a synchronous
refresh before proceeding. If the synchronous refresh
succeeds, use the new token. If it fails, fall through
to the expiry check (which will return an error if
the token has actually expired).

**Rationale**: The 50-minute refresh interval means
there's a 10-minute window (50-60 minutes) where the
token is valid but won't be refreshed until the next
tick. Proactive refresh closes this window for active
sessions.

**Concurrency**: Use `sync.Mutex.TryLock()` on a
dedicated `tokenRefreshing` mutex to deduplicate
concurrent proactive refresh attempts. The locking
protocol is:

1. `tryProactiveRefresh()` calls
   `tokenRefreshing.TryLock()`. If it cannot acquire
   the lock (another goroutine is already refreshing),
   it returns immediately -- the request proceeds with
   the current (still valid) token.
2. If `TryLock()` succeeds, call
   `refreshVertexToken()` with a 5-second context
   timeout to prevent hung `gcloud` subprocesses from
   blocking requests.
3. On success, acquire `tokenMu` write lock, update
   both `token` and `tokenExpiry` atomically, release
   `tokenMu`, then release `tokenRefreshing`.
4. On failure (including timeout), release
   `tokenRefreshing` without modifying the token or
   expiry (the token may still be valid).

The background refresh closure uses `tokenMu` only
(not `tokenRefreshing`). The two mutexes have
non-overlapping responsibilities: `tokenRefreshing`
serializes proactive refresh attempts,
`tokenMu` protects the token/expiry fields. The
background closure and proactive refresh both acquire
`tokenMu` for writes but never hold both mutexes
simultaneously, preventing deadlock.

**Timeout**: The synchronous refresh uses a 5-second
timeout via `context.WithTimeout`. If `gcloud` does
not complete within 5 seconds, the refresh is
abandoned and the request proceeds with the current
token. This prevents a hung subprocess from cascading
into request timeouts.

### D4: Log File for Detached Mode

Redirect child process stdout/stderr to
`.uf/gateway.log` (opened with `O_CREATE|O_TRUNC`)
in the `detach()` function. The file is truncated on
each gateway start to prevent unbounded growth.

**Rationale**: The current `nil` redirect discards
all diagnostic output. A log file preserves it for
troubleshooting without cluttering the parent
terminal. `.uf/` is the established location for
gateway runtime state (PID file is already there).

**Observable Quality (Principle III)**: This directly
improves observability by making provider diagnostics
accessible.

**Testability**: Log file creation is testable by
injecting `ExecStart` to capture the `exec.Cmd`
configuration and verifying `cmd.Stdout`/`cmd.Stderr`
are set to a non-nil `*os.File` pointing to
`.uf/gateway.log`. The existing `ExecStart` injection
point requires no changes.

**Status code rationale**: Expired-token errors use
HTTP 502 Bad Gateway via the existing `writeJSONError`
path with error type `"auth_error"`. This is
consistent with the existing `ErrorHandler` in
`newMux()` (`gateway.go:213`) which returns 502 for
all `PrepareRequest` failures. 502 is semantically
correct because the gateway is acting as a proxy and
cannot authenticate with the upstream provider.

### D5: Status Command Shows Log Path

When `uf gateway status` reports a running gateway,
include the log file path if the file exists. This
guides users to diagnostics when troubleshooting.

### D6: Bedrock Credential Expiry Parity

Apply the same expiry tracking pattern to
`BedrockProvider`. AWS session credentials expire in
1-12 hours depending on the source. Use
`now + 50 minutes` as the conservative expiry (matching
the refresh interval) since the actual expiry is not
reliably parseable from `aws configure export-
credentials` output.

## Risks / Trade-offs

### R1: Clearing Token on Transient Failure

Clearing the token on any refresh failure (D2) means
a transient gcloud CLI error causes immediate request
failure even if the old token is still valid. This is
acceptable because:
- The failure window is small (50-60 minutes)
- The error message is actionable
- The alternative (silent 401s) is worse

### R2: Synchronous Refresh Latency

Proactive refresh in `PrepareRequest` (D3) adds
latency to a request that triggers it (~1-2 seconds
for a `gcloud` subprocess). This is acceptable because:
- It only fires once per token lifecycle (near expiry)
- The alternative is a failed request that must be
  retried manually
- The refresh is deduplicated across concurrent
  requests

### R3: Log File Growth

The log file is truncated on each start (D4), so
growth is bounded by a single gateway session. For
long-running sessions (days), the log could grow
large. Log rotation is deferred as a non-goal for
this change.

### R4: Hardcoded Expiry Duration

The 55-minute expiry (D1) assumes gcloud tokens are
valid for 60 minutes. If Google changes the token
lifetime, the safety margin may be wrong. This is
low risk -- the 60-minute lifetime has been stable
for years, and the proactive refresh (D3) provides
additional resilience.
