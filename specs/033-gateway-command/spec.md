# Feature Specification: LLM Gateway Command

**Feature Branch**: `033-gateway-command`
**Created**: 2026-04-20
**Status**: Draft
**Input**: User description: "Add uf gateway command — minimal LLM reverse proxy for sandbox credential isolation"

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Gateway Starts with Auto-Detected Provider (Priority: P1)

A developer runs `uf gateway` on their host machine.
The gateway reads existing environment variables to
determine which cloud provider to use, starts a local
proxy on port 53147, and serves the Anthropic Messages
API. The developer can then point any tool (inside or
outside a container) at `http://localhost:53147` to
make LLM requests without managing credentials in the
client.

**Why this priority**: This is the core capability.
Without the gateway running and serving requests,
nothing else works.

**Independent Test**: Set `ANTHROPIC_API_KEY` in the
shell, run `uf gateway`, then `curl` the health
endpoint and send a test `/v1/messages` request.

**Acceptance Scenarios**:

1. **Given** `ANTHROPIC_API_KEY` is set in the
   environment, **When** the developer runs
   `uf gateway`, **Then** the gateway starts on port
   53147, auto-detects direct Anthropic mode, and
   serves `GET /health` returning the provider name
   and port.

2. **Given** `CLAUDE_CODE_USE_VERTEX=1` and
   `ANTHROPIC_VERTEX_PROJECT_ID` are set, **When** the
   developer runs `uf gateway`, **Then** the gateway
   auto-detects Vertex AI mode, obtains an OAuth token
   from the host's credential chain, and forwards
   requests to the Vertex rawPredict endpoint.

3. **Given** `CLAUDE_CODE_USE_BEDROCK=1` and
   `AWS_REGION` are set, **When** the developer runs
   `uf gateway`, **Then** the gateway auto-detects
   Bedrock mode and forwards requests using the host's
   AWS credential chain.

4. **Given** no recognized provider environment
   variables are set, **When** the developer runs
   `uf gateway`, **Then** the gateway exits with a
   clear error listing the supported providers and
   which environment variables to set.

---

### User Story 2 — Sandbox Auto-Starts Gateway (Priority: P1)

A developer runs `uf sandbox start` and the sandbox
automatically starts a gateway on the host (if not
already running), passes the gateway URL into the
container as `ANTHROPIC_BASE_URL`, and stops mounting
credential files into the container. The developer
never needs to know the gateway exists — the sandbox
"just works" with cloud providers.

**Why this priority**: This is the primary use case
that motivated the feature. It solves the credential
isolation problem (issue #108) and eliminates UID
mismatch issues with `~/.config/gcloud/` mounts.

**Independent Test**: Configure Vertex AI credentials
on the host, run `uf sandbox start`, then verify
OpenCode inside the container can make LLM requests
without any credential files mounted.

**Acceptance Scenarios**:

1. **Given** a developer has Vertex AI credentials on
   the host and no gateway is running, **When** they
   run `uf sandbox start`, **Then** the sandbox starts
   `uf gateway --detach` automatically, waits for the
   health endpoint, and passes
   `ANTHROPIC_BASE_URL=http://host.containers.internal:53147`
   and `ANTHROPIC_AUTH_TOKEN=gateway` to the container.

2. **Given** a gateway is already running on port
   53147, **When** the developer runs
   `uf sandbox start`, **Then** the sandbox detects the
   existing gateway via health check and reuses it
   without starting a second instance.

3. **Given** the gateway cannot auto-detect a provider
   (no env vars set), **When** the developer runs
   `uf sandbox start`, **Then** the sandbox falls back
   to the existing credential mount behavior (backward
   compatible) and logs a message suggesting
   `uf gateway` for easier credential management.

4. **Given** the sandbox is using the gateway, **When**
   the developer inspects the container's environment,
   **Then** no credential files are mounted and no
   provider API keys are present in the container's
   environment variables.

---

### User Story 3 — Automatic Token Refresh (Priority: P2)

A developer using Google Vertex AI starts the gateway
and works for several hours. The gateway automatically
refreshes OAuth tokens before they expire, so the
developer never experiences authentication failures
during long sessions.

**Why this priority**: Token expiry during a coding
session causes confusing errors. This story ensures
the gateway handles credential lifecycle transparently.

**Independent Test**: Start the gateway in Vertex
mode, wait for the token refresh interval to pass,
then make a request and verify it succeeds.

**Acceptance Scenarios**:

1. **Given** the gateway is running in Vertex AI mode,
   **When** the current OAuth token is within 10
   minutes of expiry, **Then** the gateway obtains a
   fresh token from the host's credential chain before
   the next request.

2. **Given** the gateway cannot refresh the token
   (e.g., `gcloud` not authenticated), **When** a
   request arrives, **Then** the gateway returns a
   clear error to the client indicating the
   credential refresh failed and suggesting the
   developer re-authenticate on the host.

---

### User Story 4 — Explicit Provider Override (Priority: P2)

A developer wants to override the auto-detected
provider or use a non-default port. They run
`uf gateway --provider anthropic --port 9000` to
start the gateway with explicit configuration.

**Why this priority**: Auto-detection covers the common
case, but developers need escape hatches for
non-standard setups or testing.

**Independent Test**: Run `uf gateway --provider
anthropic --port 9000`, then verify the health
endpoint at `localhost:9000` returns the correct
provider.

**Acceptance Scenarios**:

1. **Given** the developer specifies
   `--provider anthropic`, **When** the gateway starts,
   **Then** it uses direct Anthropic mode regardless of
   other environment variables.

2. **Given** the developer specifies `--port 9000`,
   **When** the gateway starts, **Then** it listens on
   port 9000 instead of the default 53147.

3. **Given** the developer specifies an invalid
   provider name, **When** the gateway starts, **Then**
   it exits with an error listing valid provider names.

---

### User Story 5 — Background Mode and Lifecycle (Priority: P3)

A developer starts the gateway in the background with
`uf gateway --detach` and later stops it with
`uf gateway stop`. The gateway manages its own PID
file for lifecycle tracking.

**Why this priority**: Background mode is needed for
the sandbox auto-start integration (US2) and for
developers who want the gateway running persistently.

**Independent Test**: Run `uf gateway --detach`, verify
the health endpoint responds, then run
`uf gateway stop` and verify the process is gone.

**Acceptance Scenarios**:

1. **Given** the developer runs `uf gateway --detach`,
   **When** the gateway starts, **Then** it forks to
   the background, writes a PID file, and the
   developer's terminal returns immediately.

2. **Given** a gateway is running in the background,
   **When** the developer runs `uf gateway stop`,
   **Then** the process is terminated and the PID file
   is removed.

3. **Given** no gateway is running, **When** the
   developer runs `uf gateway stop`, **Then** the
   command exits gracefully with a message indicating
   no gateway was found.

4. **Given** a gateway is running, **When** the
   developer runs `uf gateway status`, **Then** the
   command displays the provider, port, PID, and
   uptime.

---

### Edge Cases

- What happens when port 53147 is already in use by
  another process? The gateway exits with an error
  identifying the port conflict and suggesting
  `--port` to use an alternative.
- What happens when the gateway receives a request
  for an unsupported endpoint (e.g., `/v1/completions`
  instead of `/v1/messages`)? The gateway returns a
  405 Method Not Allowed with a message listing
  supported endpoints.
- What happens when the upstream provider returns an
  error (rate limit, server error)? The gateway
  forwards the error response as-is to the client,
  preserving the status code and error body.
- What happens when the gateway process crashes while
  in background mode? The PID file becomes stale. On
  the next `uf gateway` or `uf sandbox start`, the
  health check fails, the stale PID file is cleaned
  up, and a new gateway is started.
- What happens when the developer's machine goes to
  sleep and wakes up? The gateway process continues
  running. If OAuth tokens expired during sleep, the
  next request triggers a refresh before forwarding.
- What happens when the gateway is started but the
  upstream provider is unreachable? The gateway starts
  successfully (it's a proxy, not a provider). The
  first request to the provider will fail with a
  connection error forwarded to the client.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The `uf gateway` command MUST start a
  local reverse proxy that serves the Anthropic
  Messages API (`/v1/messages`,
  `/v1/messages/count_tokens`) on port 53147 by
  default.

- **FR-002**: The gateway MUST forward
  `anthropic-beta` and `anthropic-version` request
  headers to the upstream provider, as required by the
  Claude Code LLM gateway specification.

- **FR-003**: The gateway MUST auto-detect the cloud
  provider from environment variables:
  `CLAUDE_CODE_USE_VERTEX=1` +
  `ANTHROPIC_VERTEX_PROJECT_ID` → Vertex AI mode;
  `CLAUDE_CODE_USE_BEDROCK=1` → Bedrock mode;
  `ANTHROPIC_API_KEY` present → direct Anthropic mode.

- **FR-004**: The gateway MUST inject host-side
  credentials into upstream requests: API key header
  for Anthropic, OAuth bearer token for Vertex AI, and
  AWS signature for Bedrock.

- **FR-005**: The gateway MUST automatically refresh
  expiring credentials before they expire. For Vertex
  AI OAuth tokens, refresh MUST occur at least 10
  minutes before expiry.

- **FR-006**: The gateway MUST serve a health endpoint
  at `GET /health` returning the provider name, port,
  and status in a machine-parseable format.

- **FR-007**: The `--detach` flag MUST start the
  gateway as a background process and write a PID file
  to `.uf/gateway.pid`.

- **FR-008**: The `uf gateway stop` subcommand MUST
  terminate a running background gateway and remove its
  PID file. The `uf gateway status` subcommand MUST
  display the running gateway's provider, port, PID,
  and uptime.

- **FR-009**: The `--provider` flag MUST override
  auto-detection. The `--port` flag MUST override the
  default port. Both flags are optional.

- **FR-010**: The `uf sandbox start` command MUST
  auto-start the gateway (via `uf gateway --detach`)
  when a cloud provider is detected and no gateway is
  already running.

- **FR-011**: When the gateway is active, `uf sandbox
  start` MUST pass
  `ANTHROPIC_BASE_URL=http://host.containers.internal:53147`
  and `ANTHROPIC_AUTH_TOKEN=gateway` to the container,
  and MUST NOT mount credential files or forward
  provider API keys into the container.

- **FR-012**: When the gateway cannot auto-detect a
  provider, `uf sandbox start` MUST fall back to the
  existing credential mount behavior for backward
  compatibility.

- **FR-013**: The gateway MUST NOT require inbound
  authentication from clients, since it listens only on
  localhost.

- **FR-014**: The gateway MUST forward upstream error
  responses (status codes, error bodies) to the client
  without modification.

- **FR-015**: The gateway MUST include the
  `X-Claude-Code-Session-Id` header in forwarded
  requests if present in the inbound request (per
  Claude Code LLM gateway specification).

### Key Entities

- **Gateway Process**: The running reverse proxy
  instance, identified by PID file at
  `.uf/gateway.pid`. Has a provider, port, and
  credential state.

- **Provider**: The upstream cloud service
  (Anthropic, Vertex AI, Bedrock). Determines the
  credential injection strategy and upstream URL
  format.

- **Credential State**: The current authentication
  material (API key, OAuth token, AWS credentials).
  For token-based providers, includes expiry time
  and refresh logic.

## Dependencies & Assumptions

- **Spec 028 (Sandbox Command)**: The gateway
  integrates with the existing sandbox `Start()`
  function. The sandbox must be modified to detect
  and auto-start the gateway.

- **Spec 029 (Sandbox CDE Lifecycle)**: The CDE
  backend does not use the gateway — CDE credentials
  are injected via Kubernetes secrets. The gateway is
  for Podman backend only.

- **Issue #108 (gcloud credential strategy)**: This
  spec supersedes the bind-mount approach. After
  implementation, issue #108 can be closed.

- **OpenSpec change `sandbox-gcloud-dir-mount`**: The
  `~/.config/gcloud/` directory mount implemented in
  that change becomes the fallback behavior when no
  gateway is running.

- **Claude Code LLM gateway specification**: The
  gateway must comply with the requirements at
  https://code.claude.com/docs/en/llm-gateway —
  specifically the Anthropic Messages API format,
  required header forwarding, and session ID header.

- **Assumption**: The developer has valid credentials
  on the host machine (e.g., `ANTHROPIC_API_KEY`,
  `gcloud auth application-default login`, or AWS
  credentials configured).

- **Assumption**: The `host.containers.internal`
  hostname resolves correctly from within Podman
  containers (established in Spec 028).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can start the gateway and
  make an LLM request through it within 30 seconds of
  running `uf gateway`, as measured from command
  invocation to first successful response.

- **SC-002**: The sandbox starts with cloud provider
  access without any credential files mounted in the
  container, verified by inspecting the container's
  mount list and environment variables.

- **SC-003**: The gateway handles token refresh
  transparently — a developer working for 4+ hours
  with Vertex AI experiences zero authentication
  failures caused by token expiry.

- **SC-004**: When the gateway is unavailable or
  cannot auto-detect a provider, the sandbox falls
  back to the existing credential mount behavior with
  zero user-visible errors (backward compatible).

- **SC-005**: The gateway adds less than 50ms of
  latency to each request, as measured by comparing
  direct provider requests to gateway-proxied requests.

- **SC-006**: The gateway binary footprint is under
  5MB (the `uf` binary size increase), maintaining the
  project's "small footprint" principle.
<!-- scaffolded by uf vdev -->
