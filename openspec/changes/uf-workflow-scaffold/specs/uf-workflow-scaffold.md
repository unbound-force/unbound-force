## ADDED Requirements

### FR-001: Plugin Entry Point

The `uf-workflow` plugin MUST provide an entry point at
`.opencode/plugins/uf-workflow/index.ts` that exports a
named const satisfying `PluginModule` from
`@opencode-ai/plugin`, with `id` and async `server`
properties. The const MUST be the default export.

The entry point MUST import from `@opencode-ai/plugin`
following the patterns established by existing plugins
(`invoke-agent`, `review-dispatch`). Additional imports
(e.g., `zod`) SHOULD be added only when tools are
registered and require schema validation.

The `server` function MUST return an object with a
`tool` property containing a tools registry that MAY be
empty at scaffold time and SHALL be populated by
subsequent Phase 2 tool extraction PRs.

#### Scenario: Plugin loads successfully

- **GIVEN** the `uf-workflow` plugin exists at
  `.opencode/plugins/uf-workflow/index.ts`
- **WHEN** OpenCode starts and processes the `plugin`
  array in `opencode.json`
- **THEN** the plugin loads without errors, the default
  export satisfies `PluginModule`, and its `server`
  function executes without throwing

#### Scenario: Empty tool registry

- **GIVEN** the plugin scaffold has no tools registered
- **WHEN** a consumer queries available tools
- **THEN** the plugin reports zero tools without error

---

### FR-002: Standard Result Envelope Types

The shared types library at
`.opencode/lib/uf-workflow-types.ts` MUST export the
following types:

- `ToolSuccess<T>` — `{ status: "ok", data?: T }`
- `ToolFailure` — `{ status: "error", message: string,
  retryable: boolean, code?: string }`
- `ToolResult<T>` — `ToolSuccess<T> | ToolFailure`

The library MUST export helper functions:

- `success<T>(data?: T): ToolSuccess<T>` — constructs a
  success envelope
- `failure(message: string, options?: { retryable?: boolean,
  code?: string }): ToolFailure` — constructs a failure
  envelope; `retryable` MUST default to `false`

Implementations SHOULD NOT include internal file paths,
stack traces, or system details in the `message` field.
Error messages SHOULD be user-actionable and free of
information that could leak internal implementation
details.

#### Scenario: Success envelope without data

- **GIVEN** a side-effect tool completes successfully
- **WHEN** the tool calls `success()`
- **THEN** the result is `{ status: "ok" }` with no
  `data` property

#### Scenario: Success envelope with data

- **GIVEN** a query tool resolves a value
- **WHEN** the tool calls `success({ family: "claude" })`
- **THEN** the result is
  `{ status: "ok", data: { family: "claude" } }`

#### Scenario: Failure envelope with defaults

- **GIVEN** a tool encounters an error
- **WHEN** the tool calls `failure("file not found")`
- **THEN** the result is `{ status: "error",
  message: "file not found", retryable: false }` with
  no `code` property

#### Scenario: Failure envelope with all fields

- **GIVEN** a tool encounters a retryable error
- **WHEN** the tool calls
  `failure("rate limited", { retryable: true, code: "RATE_LIMIT" })`
- **THEN** the result is `{ status: "error",
  message: "rate limited", retryable: true,
  code: "RATE_LIMIT" }`

---

### FR-003: Plugin Registration

The `opencode.json` file MUST include the `uf-workflow`
plugin path in its `plugin` array.

The plugin path MUST follow the format used by existing
plugins (relative path from project root).

#### Scenario: Plugin registered in opencode.json

- **GIVEN** `opencode.json` contains a `plugin` array
- **WHEN** the uf-workflow scaffold is applied
- **THEN** the array includes
  `./.opencode/plugins/uf-workflow/index.ts`

---

### FR-004: Canonical Scaffold Copy

A byte-identical copy of the plugin entry point MUST exist
at `internal/scaffold/assets/opencode/plugins/uf-workflow/index.ts`
for distribution via `uf init`.

The existing `TestEmbeddedAssets_MatchSource` test in
`internal/scaffold/scaffold_test.go` SHALL detect drift
between the working copy and the canonical copy without
requiring new test code.

#### Scenario: Drift detection catches divergence

- **GIVEN** the canonical copy exists under
  `internal/scaffold/assets/`
- **WHEN** a developer modifies the working copy at
  `.opencode/plugins/uf-workflow/index.ts` without
  updating the canonical copy
- **THEN** `TestEmbeddedAssets_MatchSource` fails with
  a drift error message

#### Scenario: Copies are synchronized

- **GIVEN** both copies exist
- **WHEN** `go test ./internal/scaffold/...` runs
- **THEN** drift detection passes (byte-identical)

---

### FR-005: Test Scaffold

A test file MUST exist at
`.opencode/test/uf-workflow.test.ts` using vitest.

The test file MUST include:

1. Envelope helper tests verifying `success()` and
   `failure()` return correctly shaped objects
2. A plugin smoke test verifying the default export
   satisfies `PluginModule` (has `id` and `server`)

The tests MUST pass under `make plugin-test` without
network access or external services.

#### Scenario: Envelope helpers produce correct shapes

- **GIVEN** the uf-workflow-types library is imported
- **WHEN** tests call `success()`, `success(data)`,
  `failure(msg)`, and `failure(msg, opts)`
- **THEN** each result matches the specified type shape

#### Scenario: Plugin smoke test

- **GIVEN** the uf-workflow plugin module is importable
- **WHEN** the default export is inspected
- **THEN** it is a `PluginModule` object with `id`
  string and async `server` function that resolves
  without throwing

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
