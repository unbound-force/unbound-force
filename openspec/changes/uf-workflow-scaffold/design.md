## Context

Phase 2 of the Token Reduction epic (#659) replaces
prose-described deterministic procedures with plugin tools
in a `uf-workflow` plugin. This design covers the scaffold
— the foundational structure that all subsequent tool
extractions will build on.

Existing plugins (`invoke-agent`, `review-dispatch`)
establish the patterns: `@opencode-ai/plugin` framework,
`server()` entry point, Zod schemas, shared types in
`.opencode/lib/`, and test files in `.opencode/test/`.

## Goals / Non-Goals

### Goals
- Establish the `uf-workflow` plugin entry point following
  existing plugin patterns
- Define the standard result envelope types shared by all
  future uf-workflow tools
- Register the plugin in `opencode.json`
- Create a canonical copy under `internal/scaffold/assets/`
  for `uf init` distribution
- Provide a test scaffold with envelope type validation
  and plugin smoke test
- Leverage existing Go drift detection
  (`TestEmbeddedAssets_MatchSource`) — no new drift test
  code needed

### Non-Goals
- Implementing any actual workflow tools (Phase 2 follow-ups)
- Creating Zod runtime validators for the envelope (types
  are sufficient for this scaffold; tools will add Zod
  schemas for their specific I/O)
- Modifying existing plugins or shared libraries
- Adding new npm dependencies beyond what
  `.opencode/package.json` already declares

## Decisions

### D1: Result envelope as TypeScript types, not Zod schemas

The `ToolSuccess<T>`, `ToolFailure`, and `ToolResult<T>`
types are defined as TypeScript type aliases in
`.opencode/lib/uf-workflow-types.ts`. Zod schemas are not
needed at the envelope level because:

- The envelope is an output shape, not an input validation
  boundary. Tools validate their inputs with Zod; the
  envelope types constrain what they return.
- Individual tools will define Zod schemas for their
  specific `data` payloads. The envelope wrapper is
  applied by a helper function, not parsed from untrusted
  input.
- TypeScript's structural typing catches envelope
  violations at compile time.

This aligns with **Observable Quality** — the envelope
produces machine-parseable JSON with structured status and
error codes.

### D2: Helper functions for envelope construction

A `success<T>(data?: T)` and `failure(message, options?)`
helper pair in the types library ensures consistent
envelope construction. Tools call `success({ ... })` or
`failure("msg", { retryable: true, code: "NOT_FOUND" })`
rather than constructing raw objects.

### D3: PluginModule export with empty tool registry

The plugin entry point exports a named const satisfying
`PluginModule` from `@opencode-ai/plugin`, with `id` and
async `server` properties. The `server` function returns
an object with an empty `tool` registry. A TODO comment
marks where tools will be registered. This follows the
pattern established by `invoke-agent` and
`review-dispatch` — building the skeleton first, then
filling in tools via separate PRs (each traced to a
specific Phase 2 issue).

### D4: Dual-copy with existing drift detection

The canonical copy lives at
`internal/scaffold/assets/opencode/plugins/uf-workflow/index.ts`.
The existing `TestEmbeddedAssets_MatchSource` test in
`internal/scaffold/scaffold_test.go` automatically detects
drift between `internal/scaffold/assets/` and the repo
root — no new drift test code is required. The test
compares embedded assets byte-for-byte against their
source files.

This maintains **Composability First** — the scaffold is
distributable via `uf init` without manual synchronization.

### D5: Test file structure

`uf-workflow.test.ts` follows existing patterns
(vitest, `describe`/`it`, `expect`). Tests cover:

1. **Envelope type validation** — `success()` and
   `failure()` helpers return correctly shaped objects
   with expected fields
2. **Plugin smoke test** — the default export satisfies
   `PluginModule` with `id` string and async `server`
   that resolves to the expected tool registry structure

Tests run under `make plugin-test` with no additional
configuration needed.

## Risks / Trade-offs

### R1: Types-only envelope may drift from runtime shape

**Mitigation**: The helper functions (`success`, `failure`)
are the only way to construct envelopes, and tests validate
their output shapes. When Zod schemas are needed for
specific tool data payloads, they compose with the
envelope helpers rather than replacing them.

### R2: Empty plugin has no functional tools yet

**Accepted**: This is intentional — the scaffold is a
foundation. Registering an empty plugin in `opencode.json`
means it loads on startup (negligible cost) but proves
the registration path works before tools are added.

### R3: Canonical copy adds maintenance surface

**Mitigated by existing infrastructure**: The Go drift
detection test (`TestEmbeddedAssets_MatchSource`) already
runs in CI. Any drift between the working copy and the
canonical copy fails the build immediately. No new test
code or CI changes needed.
