<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Shared Types Library

- [x] 1.1 [P] Create `.opencode/lib/uf-workflow-types.ts`
  with `ToolSuccess<T>`, `ToolFailure`, `ToolResult<T>`
  type aliases and `success<T>(data?)` / `failure(message,
  options?)` helper functions. `retryable` defaults to
  `false`. Export all types and helpers.

## 2. Plugin Entry Point

- [x] 2.1 Create `.opencode/plugins/uf-workflow/index.ts`
  with a named const satisfying `PluginModule` (with `id`
  and async `server` properties), empty `tool` registry,
  and imports from `@opencode-ai/plugin` and `zod`.
  Import envelope types from
  `../../lib/uf-workflow-types.js`. Export the const as
  default and re-export envelope helpers for consumer
  convenience.

## 3. Registration and Distribution

- [x] 3.1 [P] Add
  `./.opencode/plugins/uf-workflow/index.ts` to the
  `plugin` array in `opencode.json`.
- [x] 3.2 [P] Copy the plugin entry point to
  `internal/scaffold/assets/opencode/plugins/uf-workflow/index.ts`
  (byte-identical canonical copy for `uf init`).
- [x] 3.3 [P] Copy the types library to
  `internal/scaffold/assets/opencode/lib/uf-workflow-types.ts`
  (byte-identical canonical copy for `uf init`).
- [x] 3.4 [P] Update `expectedAssetPaths` in
  `internal/scaffold/scaffold_test.go` to include the
  new plugin and types library paths.

## 4. Test Scaffold

- [x] 4.1 Create `.opencode/test/uf-workflow.test.ts`
  with vitest. Include:
  - Envelope helper tests: `success()` returns
    `{ status: "ok" }`, `success(data)` includes `data`,
    `failure(msg)` returns `{ status: "error",
    message, retryable: false }`, `failure(msg,
    { retryable: true, code: "X" })` includes all fields.
  - Plugin smoke test: default export satisfies
    `PluginModule` with `id` and async `server` that
    resolves without throwing.

## 5. Verification

- [x] 5.1 Run `make plugin-test` — all tests pass
  including new `uf-workflow.test.ts`.
- [x] 5.2 Run `go test ./internal/scaffold/...` — drift
  detection passes (canonical copy matches working copy).
- [x] 5.3 Verify constitution alignment:
  - Autonomous Collaboration: envelope is a self-describing
    artifact format with status, error codes, retryable.
  - Composability First: plugin independently registerable,
    no mandatory dependencies on other plugins.
  - Observable Quality: machine-parseable JSON output with
    structured error codes.
  - Testability: all tests run in isolation without
    network or LLM dependencies.

<!-- spec-review: passed -->
<!-- code-review: passed -->
