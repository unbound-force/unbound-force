# Feature Specification: Gateway Vertex Translation

**Feature Branch**: `034-gateway-vertex-translation`
**Created**: 2026-04-23
**Status**: Draft
**Input**: User description: "Transform the uf gateway
from a pure auth-injecting reverse proxy into a
translation-mode proxy for Vertex AI. Accept standard
Anthropic API requests from OpenCode and translate
them to/from Vertex rawPredict format, including
request body rewriting, streaming endpoint detection,
and SSE response filtering."

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Request Body Translation (Priority: P1)

A developer runs `uf sandbox start` on a machine with
Vertex AI credentials. OpenCode inside the sandbox
sends a standard Anthropic `/v1/messages` request
(with `model` in the body, `anthropic-beta` and
`anthropic-version` headers). The gateway translates
this into a Vertex-compatible request: removes `model`
from the body, injects `anthropic_version:
"vertex-2023-10-16"`, strips SDK-injected headers,
and forwards to the correct Vertex rawPredict
endpoint.

**Why this priority**: Without request translation,
Vertex rejects every request from the sandbox with
`anthropic_version: Field required`. This blocks all
sandbox usage on Vertex AI.

**Independent Test**: Send a standard Anthropic
`/v1/messages` POST to the gateway running in Vertex
mode. Verify the upstream request has `model` removed
from the body, `anthropic_version` injected, and
headers stripped. Verify the URL targets the correct
rawPredict endpoint with the model in the path.

**Acceptance Scenarios**:

1. **Given** the gateway is running with Vertex
   provider, **When** OpenCode sends `POST
   /v1/messages` with `{"model":
   "claude-sonnet-4-20250514", "messages": [...]}`,
   **Then** the gateway forwards to
   `https://{region}-aiplatform.googleapis.com/v1/
   projects/{project}/locations/{region}/publishers/
   anthropic/models/claude-sonnet-4-20250514:rawPredict`
   with `model` removed from the body and
   `anthropic_version: "vertex-2023-10-16"` injected.

2. **Given** the request body already contains
   `anthropic_version`, **When** the gateway
   translates the request, **Then** the existing
   `anthropic_version` value is preserved (not
   overwritten).

3. **Given** the request contains `anthropic-beta`
   and `anthropic-version` HTTP headers, **When** the
   gateway translates the request, **Then** both
   headers are removed before forwarding to Vertex.

---

### User Story 2 — Streaming Endpoint Detection (Priority: P1)

When OpenCode sends a streaming request (`"stream":
true` in the body), the gateway uses Vertex's
`streamRawPredict` endpoint instead of `rawPredict`.
This is required because Vertex uses separate
endpoints for streaming and non-streaming requests.

**Why this priority**: OpenCode always streams chat
completions. Using the wrong endpoint causes Vertex
to return non-streaming responses or errors.

**Independent Test**: Send a request with `"stream":
true` in the body and verify the upstream URL uses
`streamRawPredict`. Send a request without `stream`
or with `"stream": false` and verify `rawPredict` is
used.

**Acceptance Scenarios**:

1. **Given** the request body contains `"stream":
   true`, **When** the gateway builds the upstream
   URL, **Then** the URL path ends with
   `:streamRawPredict`.

2. **Given** the request body contains `"stream":
   false` or omits the `stream` field, **When** the
   gateway builds the upstream URL, **Then** the URL
   path ends with `:rawPredict`.

3. **Given** the request body is malformed or empty,
   **When** the gateway builds the upstream URL,
   **Then** the gateway falls back to `rawPredict`
   (non-streaming default).

---

### User Story 3 — SSE Response Filtering (Priority: P1)

When Vertex returns streaming responses, it includes
`vertex_event` SSE events that the standard Anthropic
API does not produce. OpenCode's Anthropic provider
cannot parse these events and fails with a type
validation error. The gateway filters out
`vertex_event` events from the SSE stream before
forwarding to OpenCode.

**Why this priority**: Without response filtering,
every streaming response fails with
`Type validation failed: ... "type": "vertex_event"`.
This blocks all chat usage in the sandbox.

**Independent Test**: Construct a mock SSE stream
containing `message_start`, `content_block_delta`,
`vertex_event`, and `message_stop` events. Pass it
through the filter. Verify `vertex_event` events are
dropped and all others pass through unchanged.

**Acceptance Scenarios**:

1. **Given** a Vertex streaming response contains an
   SSE event with `event: vertex_event`, **When** the
   gateway forwards the response to OpenCode, **Then**
   the `vertex_event` event and its `data:` line are
   dropped from the stream.

2. **Given** a Vertex streaming response contains
   standard Anthropic events (`message_start`,
   `content_block_delta`, `content_block_stop`,
   `message_delta`, `message_stop`), **When** the
   gateway forwards the response, **Then** all
   standard events pass through unchanged.

3. **Given** a non-streaming response (not
   `text/event-stream`), **When** the gateway
   forwards the response, **Then** no filtering is
   applied and the response passes through unchanged.

4. **Given** a Vertex streaming response contains a
   `ping` event, **When** the gateway forwards the
   response, **Then** the `ping` event is dropped
   (matching Portkey behavior).

---

### User Story 4 — Sandbox Provider Isolation (Priority: P2)

The sandbox container should not need any
Vertex-specific environment variables. OpenCode
inside the container sees only `ANTHROPIC_BASE_URL`
and `ANTHROPIC_API_KEY=gateway`, configuring itself
as a standard Anthropic provider. The gateway handles
all provider-specific translation transparently.

**Why this priority**: Provider isolation simplifies
the sandbox configuration and prevents env var
leakage. It also means the same sandbox image works
identically regardless of whether the host uses
Vertex, Bedrock, or direct Anthropic.

**Independent Test**: Start a sandbox with Vertex
credentials on the host. Verify the container
environment contains only `ANTHROPIC_BASE_URL` and
`ANTHROPIC_API_KEY=gateway`. Verify no
`CLAUDE_CODE_USE_VERTEX`, `ANTHROPIC_VERTEX_PROJECT_ID`,
or `VERTEX_LOCATION` vars are present.

**Acceptance Scenarios**:

1. **Given** the host has `CLAUDE_CODE_USE_VERTEX=1`
   and `ANTHROPIC_VERTEX_PROJECT_ID` set, **When**
   `uf sandbox start` launches the container with
   gateway active, **Then** the container environment
   contains `ANTHROPIC_BASE_URL` and
   `ANTHROPIC_API_KEY=gateway` but does NOT contain
   `CLAUDE_CODE_USE_VERTEX`,
   `ANTHROPIC_VERTEX_PROJECT_ID`, or
   `VERTEX_LOCATION`.

2. **Given** the gateway is NOT active (no provider
   detected), **When** `uf sandbox start` launches
   the container, **Then** host API keys are
   forwarded as before (backward compatible).

---

### User Story 5 — Synthetic Model Catalog (Priority: P3)

The gateway's `/v1/models` endpoint returns a list of
Claude models available on Vertex AI. Since Vertex
has no model listing API, the gateway maintains a
hardcoded catalog matching the models Google
documents as available on Vertex. Each model entry
includes capabilities (vision, extended thinking, PDF
input) so OpenCode can correctly detect supported
features.

**Why this priority**: The model list enables
OpenCode's model picker to show available models
inside the sandbox. Lower priority because OpenCode
can function with a default model even without the
list.

**Independent Test**: Send `GET /v1/models` to the
gateway. Verify the response contains the current
Vertex-available Claude models with correct
capabilities.

**Acceptance Scenarios**:

1. **Given** the gateway is running, **When** OpenCode
   sends `GET /v1/models`, **Then** the response
   contains at least 9 models matching the Vertex AI
   Claude model catalog (Opus 4.7, Sonnet 4.6, Opus
   4.6, Opus 4.5, Sonnet 4.5, Opus 4.1, Haiku 4.5,
   Opus 4, Sonnet 4).

2. **Given** the model list response, **When**
   OpenCode parses a model entry, **Then** each entry
   includes `id`, `type`, `display_name`,
   `created_at`, `max_input_tokens`, `max_tokens`,
   and `capabilities` (vision, extended_thinking,
   pdf_input).

3. **Given** a request for a specific model (`GET
   /v1/models/{model_id}`), **When** the model ID
   matches a catalog entry, **Then** the gateway
   returns that model's details. **When** the model
   ID does not match, **Then** the gateway returns a
   404 error.

---

### Edge Cases

- What happens when the request body cannot be parsed
  as JSON? The gateway falls back to default model
  and forwards the original body unchanged.
- What happens when a single SSE event spans multiple
  TCP packets / Read() calls? The SSE filter must
  buffer partial events correctly.
- What happens when the Vertex upstream returns a
  non-200 error? The gateway forwards the error
  response unchanged (no filtering needed for error
  responses).
- What happens when `count_tokens` is called? It
  should use `rawPredict` (not streaming), and the
  request body should still be transformed.
- What happens when the Anthropic direct provider is
  used (not Vertex)? No translation occurs — requests
  and responses pass through unchanged. Translation
  is Vertex-provider-specific.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The gateway MUST remove the `model`
  field from the request body when forwarding to
  Vertex rawPredict (Vertex uses the URL path for
  model selection).
- **FR-002**: The gateway MUST inject
  `anthropic_version: "vertex-2023-10-16"` into the
  request body when the field is not already present.
- **FR-003**: The gateway MUST preserve an existing
  `anthropic_version` field in the request body
  (no overwrite).
- **FR-004**: The gateway MUST strip `anthropic-beta`
  and `anthropic-version` HTTP headers before
  forwarding to Vertex (Vertex rawPredict does not
  accept these headers).
- **FR-005**: The gateway MUST use the
  `streamRawPredict` endpoint when the request body
  contains `"stream": true`.
- **FR-006**: The gateway MUST use the `rawPredict`
  endpoint when the request body does not contain
  `"stream": true` (including when `stream` is absent
  or false).
- **FR-007**: The gateway MUST filter SSE events of
  type `vertex_event` from streaming responses before
  forwarding to the client.
- **FR-008**: The gateway MUST filter SSE events of
  type `ping` from streaming responses (matching
  Portkey behavior).
- **FR-009**: The gateway MUST NOT filter
  non-streaming responses (pass through unchanged).
- **FR-010**: The gateway MUST NOT apply request or
  response translation when the active provider is
  Anthropic (direct API). Translation is
  Vertex-provider-specific.
- **FR-011**: The sandbox container MUST NOT receive
  Vertex-specific environment variables
  (`CLAUDE_CODE_USE_VERTEX`,
  `ANTHROPIC_VERTEX_PROJECT_ID`, `VERTEX_LOCATION`)
  when the gateway is active.
- **FR-012**: The sandbox container MUST receive only
  `ANTHROPIC_BASE_URL` and `ANTHROPIC_API_KEY=gateway`
  when the gateway is active (regardless of host
  provider).
- **FR-013**: The `/v1/models` endpoint MUST return a
  model catalog matching the Claude models available
  on Vertex AI, with `capabilities` metadata (vision,
  extended_thinking, pdf_input).
- **FR-014**: The gateway MUST update
  `Content-Length` after modifying the request body
  to prevent upstream length mismatch errors.
- **FR-015**: The SSE filter MUST handle events that
  span multiple `Read()` calls (partial event
  buffering).

### Key Entities

- **SSE Event**: A server-sent event consisting of an
  `event:` line and a `data:` line, separated by
  `\n\n`. The filter operates on complete events.
- **Vertex rawPredict**: The Vertex AI endpoint for
  non-streaming Claude model invocation. URL format:
  `.../models/{model}:rawPredict`.
- **Vertex streamRawPredict**: The Vertex AI endpoint
  for streaming Claude model invocation. URL format:
  `.../models/{model}:streamRawPredict`.
- **Synthetic Model**: A hardcoded model entry in the
  gateway's `/v1/models` response, since Vertex has
  no model listing API.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can run `uf sandbox start`
  with Vertex AI credentials on the host and
  successfully chat with Claude models inside the
  sandbox without any provider-specific configuration
  in the container.
- **SC-002**: All standard Anthropic SSE event types
  (`message_start`, `content_block_delta`,
  `content_block_stop`, `message_delta`,
  `message_stop`) pass through the gateway unchanged.
- **SC-003**: `vertex_event` and `ping` SSE events
  are dropped from the stream before reaching
  OpenCode, preventing type validation errors.
- **SC-004**: The gateway adds less than 5ms overhead
  for request body transformation (JSON parse,
  modify, re-encode).
- **SC-005**: The SSE response filter does not
  introduce perceptible latency — events are
  forwarded within one streaming chunk of being
  received.
- **SC-006**: The sandbox container environment
  contains zero Vertex-specific variables when the
  gateway is active.
- **SC-007**: The Anthropic direct provider path
  remains fully functional with no behavioral
  changes (backward compatible).

## Dependencies

- **Spec 033** (Gateway Command): This spec extends
  the gateway implemented in Spec 033. All existing
  gateway functionality (provider detection, token
  refresh, PID management, health endpoint) is
  preserved.

## Assumptions

- OpenCode always sends streaming requests
  (`"stream": true`) for chat completions. This was
  confirmed by investigating OpenCode's architecture
  (Vercel AI SDK default behavior) and its
  `chunkTimeout` config option.
- Vertex AI has no REST API for listing available
  Claude models. The synthetic model list must be
  maintained manually when Anthropic releases new
  models on Vertex.
- The `vertex-2023-10-16` version string is the
  correct `anthropic_version` value for Vertex
  rawPredict. This was confirmed from Portkey's
  source code
  (`src/providers/google-vertex-ai/messages.ts`).
- The Vertex `streamRawPredict` endpoint returns
  standard Anthropic SSE events plus
  Vertex-specific events (`vertex_event`, `ping`).
  Dropping the Vertex-specific events is sufficient
  — no other response transformation is needed.
- The Bedrock provider is out of scope for this spec.
  If Bedrock needs similar translation in the future,
  it will be a separate spec.
