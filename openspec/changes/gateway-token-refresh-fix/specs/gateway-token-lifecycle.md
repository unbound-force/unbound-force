## ADDED Requirements

### FR-001: Token Expiry Tracking

The gateway MUST track the acquisition time of each
OAuth token (Vertex) or session credential (Bedrock).
A token MUST be considered expired after 55 minutes
from acquisition (5-minute safety margin before the
60-minute gcloud token lifetime). The expiry duration
SHOULD be extracted into a named constant
(`vertexTokenLifetime`). The proactive refresh window
(5 minutes) SHOULD also be a named constant
(`proactiveRefreshWindow`).

#### Scenario: Token used within validity window
- **GIVEN** a Vertex gateway with a token acquired
  3 minutes ago
- **WHEN** a `POST /v1/messages` request arrives
- **THEN** the request is forwarded with the valid
  token

#### Scenario: Token used after expiry
- **GIVEN** a Vertex gateway with a token acquired
  56 minutes ago
- **WHEN** a `POST /v1/messages` request arrives
- **THEN** the gateway returns a 502 JSON error
  (via `writeJSONError` with `http.StatusBadGateway`
  and error type `"auth_error"`, consistent with
  existing `ErrorHandler` behavior) with message
  containing "vertex AI token expired" and
  "gcloud auth application-default login"
- **AND** the request is NOT forwarded to Vertex

### FR-002: Stale Token Invalidation

When a token refresh attempt fails, the gateway
MUST clear the stored token and set the token expiry
to zero value atomically under `tokenMu` write lock.
Subsequent requests MUST receive a gateway-level
error rather than being forwarded with the stale
credential.

#### Scenario: Refresh failure clears token
- **GIVEN** a Vertex gateway with a valid token
- **WHEN** the 50-minute refresh fires
- **AND** `gcloud auth application-default
  print-access-token` fails
- **THEN** the stored token is set to empty string
  AND tokenExpiry is set to zero value, both under
  a single `tokenMu` write lock acquisition
- **AND** the next request returns a 502 JSON error
  with "Re-authenticate" guidance

#### Scenario: Refresh failure for Bedrock
- **GIVEN** a Bedrock gateway with valid credentials
- **WHEN** the 50-minute refresh fires
- **AND** `aws configure export-credentials` fails
- **THEN** the stored access key, secret key, and
  session token are cleared AND credExpiry is set to
  zero value, all under a single `credMu` write lock
  acquisition
- **AND** the next request returns a 502 JSON error
  with "Re-authenticate" guidance

#### Scenario: Token cleared but expiry not yet
  zeroed is impossible
- **GIVEN** the refresh closure implementation
- **WHEN** the refresh fails
- **THEN** token clear and expiry zero are performed
  atomically under `tokenMu` write lock (single
  critical section)
- **AND** `PrepareRequest` checks empty-token
  condition BEFORE the expiry condition, providing
  defense in depth

### FR-003: Proactive Token Refresh

When `PrepareRequest` detects that the current token
is within 5 minutes of expiry, the gateway SHOULD
attempt a synchronous token refresh before forwarding
the request. The proactive refresh MUST be
deduplicated across concurrent requests using
`sync.Mutex.TryLock()` on a dedicated
`tokenRefreshing` mutex.

#### Scenario: Proactive refresh succeeds
- **GIVEN** a Vertex gateway with a token that
  expires in 3 minutes
- **WHEN** a `POST /v1/messages` request arrives
- **THEN** a synchronous token refresh is attempted
  with a 5-second timeout
- **AND** the new token is used for the request
- **AND** the token expiry is reset to 55 minutes
  from now

#### Scenario: Proactive refresh fails but token
  still valid
- **GIVEN** a Vertex gateway with a token that
  expires in 3 minutes
- **WHEN** a `POST /v1/messages` request arrives
- **AND** the synchronous refresh fails
- **THEN** the request proceeds with the current
  (still valid) token
- **AND** the token is NOT cleared (it has not
  expired yet)

#### Scenario: Proactive refresh subprocess hangs
- **GIVEN** a Vertex gateway with a token that
  expires in 3 minutes
- **WHEN** a `POST /v1/messages` request arrives
- **AND** the `gcloud` subprocess does not complete
  within 5 seconds
- **THEN** the proactive refresh is abandoned
- **AND** the request proceeds with the current
  (still valid) token

#### Scenario: Concurrent proactive refreshes
  deduplicated
- **GIVEN** a Vertex gateway with a token that
  expires in 3 minutes
- **WHEN** 5 concurrent requests arrive
- **THEN** only 1 `gcloud` subprocess is spawned
  (via `tokenRefreshing.TryLock()` -- non-winning
  goroutines return immediately)
- **AND** all 5 requests use either the old valid
  token or the refreshed token

#### Scenario: Background refresh concurrent with
  proactive refresh
- **GIVEN** a Vertex gateway with a near-expiry token
- **WHEN** the background refresh closure fires
  simultaneously with a proactive refresh
- **THEN** both use `tokenMu` for token writes but
  neither holds both `tokenMu` and
  `tokenRefreshing` simultaneously
- **AND** no deadlock occurs
- **AND** the token is set to whichever refresh
  completed last (both are valid)

#### Scenario: Token expires during proactive refresh
- **GIVEN** a Vertex gateway with a token that
  expires in 1 second
- **WHEN** a proactive refresh is attempted but
  takes longer than 1 second
- **AND** the refresh fails
- **THEN** the expiry check catches the now-expired
  token
- **AND** the request returns a 502 JSON error with
  "token expired" guidance

### FR-004: Gateway Log File

When the gateway starts in detached mode, the child
process MUST redirect stdout and stderr to
`.uf/gateway.log`. The log file MUST be truncated
on each gateway start. The log file MUST NOT contain
raw token values.

#### Scenario: Detached gateway creates log file
- **GIVEN** no running gateway
- **WHEN** `uf gateway start --detach` is executed
- **THEN** `.uf/gateway.log` is created (or
  truncated if it exists)
- **AND** all `charmbracelet/log` output from the
  child process is written to the file

#### Scenario: Foreground gateway does not create
  log file
- **GIVEN** no running gateway
- **WHEN** `uf gateway start` is executed (no
  `--detach`)
- **THEN** logs continue to go to stderr as before
- **AND** no `.uf/gateway.log` is created

### FR-005: Status Shows Log Path

When a gateway is running in detached mode and
`.uf/gateway.log` exists, `uf gateway status` SHOULD
display the log file path.

#### Scenario: Status with log file
- **GIVEN** a running detached gateway
- **AND** `.uf/gateway.log` exists
- **WHEN** `uf gateway status` is executed
- **THEN** the output includes a "Log:" line with
  the path to `.uf/gateway.log`

## MODIFIED Requirements

### FR-006: Vertex Token Refresh Failure Handling

Previously: When `refreshVertexToken()` failed, the
error was logged and the old token remained in use.

Now: When `refreshVertexToken()` fails, the error
MUST be logged AND the stored token and expiry MUST
be cleared atomically under `tokenMu` write lock.
The refresh closure in `VertexProvider.Start()` MUST
set `p.token = ""` and `p.tokenExpiry = time.Time{}`
in a single critical section.

### FR-007: Bedrock Credential Refresh Failure
  Handling

Previously: When `refreshBedrockCredentials()` failed,
the error was logged and the old credentials remained
in use.

Now: When `refreshBedrockCredentials()` fails, the
error MUST be logged AND the stored credentials and
expiry MUST be cleared atomically under `credMu` write
lock (accessKey, secretKey, sessionToken set to empty
strings, credExpiry set to zero value).

## REMOVED Requirements

None.
