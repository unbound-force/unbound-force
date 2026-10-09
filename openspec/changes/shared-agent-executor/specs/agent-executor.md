## ADDED Requirements

### Requirement: Shared Agent Executor Function

The `.opencode/lib/agent-executor.ts` module MUST export an
`executeAgentSession()` function that manages the full child
session lifecycle: abort controller setup, child session
creation, prompt delivery with model/variant, response
parsing (text extraction, usage accumulation, model mismatch
detection), timeout enforcement, cancellation propagation,
and cleanup.

The function MUST NOT own input validation, manifest
checking, model resolution, host introspection, or prompt
file reading. Each consumer MUST validate its own inputs
and resolve models before calling the executor.

#### Scenario: Successful agent session execution

- **GIVEN** a valid `ExecutorDependencies` with `client`,
  `sessionID`, `directory`, `parentAbort`, and
  `timeoutMilliseconds`
- **WHEN** `executeAgentSession()` is called with
  `agent`, `promptText`, `model`, `variant`, and
  `read_only`
- **THEN** a child session MUST be created via
  `client.session.create()`, the prompt MUST be delivered
  via `client.session.prompt()`, the response MUST be
  parsed into an `InvokeAgentResult` with `status`,
  `text`, `usage`, `error`, and `provenance` fields

#### Scenario: Parent abort propagation

- **GIVEN** a running child session created by the executor
- **WHEN** the parent abort signal fires
- **THEN** the executor MUST propagate the abort to the
  child session via `client.session.abort()`, set
  `abortCause` to `"cancelled"`, and return a result with
  `status: "cancelled"`

#### Scenario: Timeout enforcement

- **GIVEN** `timeoutMilliseconds` is set to 30000
- **WHEN** the child session does not complete within 30
  seconds
- **THEN** the executor MUST abort the child session,
  set `abortCause` to `"timeout"`, and return a result
  with `status: "failed"` and an error describing the
  timeout

#### Scenario: Cleanup on all exit paths

- **GIVEN** a child session has been created
- **WHEN** the executor exits (success, failure, timeout,
  or cancellation)
- **THEN** the abort listener MUST be removed from the
  parent signal, the timeout timer MUST be cleared, and
  all pending cleanup operations MUST be awaited

### Requirement: Executor Type Exports

The `agent-executor.ts` module MUST export the following
types: `ModelIdentity`, `InvocationTokens`,
`InvocationUsage`, `InvocationError`,
`InvocationProvenance`, `InvokeAgentResult`.

These types MUST be structurally identical to the types
currently defined in `invoke-agent/index.ts`.

#### Scenario: Type compatibility with existing consumers

- **GIVEN** code importing `InvokeAgentResult` from
  `invoke-agent/index.ts`
- **WHEN** the import is changed to
  `../lib/agent-executor.js`
- **THEN** the code MUST compile without type errors

### Requirement: Executor Utility Function Exports

The `agent-executor.ts` module MUST export the following
functions: `errorText()`, `sanitizeInvocationError()`,
`responseFailure()`, `directModelIdentity()`,
`reportedModel()`, `failedResult()`, `invocationUsage()`,
`responseText()`, `validateTimeout()`.

All functions MUST preserve their current signatures and
behavior.

#### Scenario: sanitizeInvocationError strips credentials

- **GIVEN** an error message containing
  `Bearer eyJhbGciOiJIUzI...`
- **WHEN** `sanitizeInvocationError()` is called
- **THEN** the output MUST replace the token with
  `Bearer [REDACTED]` and MUST be bounded to 4096 bytes

### Requirement: ExecutorDependencies Interface

The executor MUST accept dependencies via an
`ExecutorDependencies` interface containing:
- `client`: The plugin client for session management
- `sessionID`: Parent session identifier
- `directory`: Working directory for the child session
- `parentAbort`: Parent `AbortSignal` for propagation
- `timeoutMilliseconds`: Maximum execution time

All dependencies MUST be injected, not captured from
module scope.

#### Scenario: Dependency injection for testing

- **GIVEN** a mock `client` implementing session create,
  prompt, abort, and message methods
- **WHEN** `executeAgentSession()` is called with the mock
- **THEN** the function MUST use only the injected
  dependencies and produce a testable result

## MODIFIED Requirements

(None — this is a new module)

## REMOVED Requirements

(None)
