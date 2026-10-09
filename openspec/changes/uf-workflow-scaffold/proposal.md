## Why

Issue #682 (parent: #659 Phase 2 — Foundation) requires a
plugin scaffold that all subsequent `uf-workflow` tools will
build on. Phase 2 replaces prose-described deterministic
procedures with plugin tools that return structured results.
Before any tool can be extracted, the scaffold must exist:
entry point, result envelope types, `opencode.json`
registration, canonical copy for `uf init`, and a test
harness.

Without this scaffold, every Phase 2 tool extraction would
independently reinvent registration, error shapes, and test
patterns — creating the same duplication the epic exists to
eliminate.

## What Changes

1. **Plugin entry point** —
   `.opencode/plugins/uf-workflow/index.ts` with a named
   const satisfying `PluginModule` (with `id` and async
   `server` properties), empty tool registry, and standard
   imports.

2. **Canonical scaffold copy** —
   `internal/scaffold/assets/opencode/plugins/uf-workflow/index.ts`
   for `uf init` distribution with drift detection.

3. **Shared result envelope types** —
   `.opencode/lib/uf-workflow-types.ts` exporting
   `ToolSuccess<T>`, `ToolFailure`, `ToolResult<T>` with
   status, retryable flag, and error code.

4. **Plugin registration** — Add the plugin path to the
   `plugin` array in `opencode.json`.

5. **Test scaffold** —
   `.opencode/test/uf-workflow.test.ts` with envelope type
   tests and a plugin smoke test.

## Capabilities

### New Capabilities
- `uf-workflow plugin`: Empty plugin scaffold with
  `PluginModule`-satisfying export ready for tool registration.
- `ToolResult<T> envelope`: Typed result shape for all future
  uf-workflow tools — `ToolSuccess<T>` and `ToolFailure`
  with status ok/error, retryable flag, error code.
- `uf-workflow-types.ts`: Shared type library importable by
  plugin tools and test code.
- `uf-workflow.test.ts`: Test scaffold with envelope type
  validation and plugin smoke test.

### Modified Capabilities
- `opencode.json`: Plugin array gains a third entry for the
  uf-workflow plugin.
- `internal/scaffold/assets/`: Gains canonical copy of the
  plugin for `uf init` scaffolding and drift detection.

### Removed Capabilities
- None.

## Impact

- **New files**: 5 (plugin index, plugin canonical copy,
  types lib, types lib canonical copy, test file).
- **Modified files**: 1 (`opencode.json` plugin array).
- **Risk**: Low — additive only, no existing behavior changes.
- **Dependencies**: `@opencode-ai/plugin`, `zod` (already in
  `.opencode/package.json`), `vitest` (dev, already present).
- **CI**: `make plugin-test` must pass with the new test file
  included.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The result envelope is a self-describing artifact format.
Every tool response carries status, error codes, and
retryable flags — consumers interpret results without
consulting the producing tool. This directly implements
the constitution's requirement for self-describing outputs
with enough metadata for any consumer to interpret them.

### II. Composability First

**Assessment**: PASS

The plugin is independently registerable via
`opencode.json` with no mandatory dependencies on other
plugins. The shared types library is an opt-in import, not
a runtime coupling. The scaffold uses the same
`@opencode-ai/plugin` framework as existing plugins,
maintaining the independently-installable pattern.

### III. Observable Quality

**Assessment**: PASS

The `ToolResult<T>` envelope produces machine-parseable
JSON output with structured error codes and status fields.
Output verbosity rules enforce minimal responses for
side-effect success and actionable detail for failures —
directly supporting the observable quality principle.

### IV. Testability

**Assessment**: PASS

The test scaffold validates envelope types and plugin
registration in isolation without external services. The
dual-copy structure includes drift detection tests to
verify canonical and working copies stay synchronized.
All tests run under `make plugin-test` with no network
or LLM dependencies.

### V. Security by Default

**Assessment**: PASS

No new dependencies introduced. Existing packages reused
(`@opencode-ai/plugin`, `zod`, `vitest`). No external
input boundaries at scaffold time — the envelope types
constrain tool output shapes, not untrusted input. Future
tools adding input parsing will need to revisit this
assessment for injection and sanitization concerns.
