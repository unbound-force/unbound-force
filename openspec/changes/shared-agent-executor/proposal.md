## Why

Two related problems emerge when agents invoke child sessions
for review runs:

1. **Model slug stripping**: Agents strip `@default` (and
   other `@suffix` version snapshots) from model slugs like
   `google-vertex-anthropic/claude-opus-4-6@default` when
   constructing `invoke_agent` calls. The review-dispatch
   pipeline preserves `@default` faithfully — the LLM is
   the weak link, misreading plan output or stripping the
   suffix when constructing tool calls. All 4 runs failed
   in observed sessions; manual correction to include
   `@default` succeeded on all 4.

2. **Prompt construction bloat**: `invoke_agent` only accepts
   `prompt` as an inline string — no file reference option.
   For review-council runs, agents build ~94KB prompt files
   on disk, then must read and reproduce the content verbatim
   in each tool call (5 x 94KB = 470KB inline text). This
   causes analysis paralysis (3+ minutes of thinking,
   multiple compression cycles) and frequently fails when
   the agent cannot fit prompt content into tool call
   parameters.

Both problems stem from the LLM sitting in the data path:
the LLM sits between correct model slugs and the invocation
tool, and must read prompt content into context then
reproduce it in tool calls.

## What Changes

1. **Extract shared agent executor** to
   `.opencode/lib/agent-executor.ts` — reusable session
   management (abort controller, child session create,
   prompt delivery, response parsing, timeout, cleanup)
   currently embedded in `invoke-agent/index.ts`.

2. **Add `promptFile` to `invoke_agent`** — an optional
   parameter that reads prompt content from disk at
   invocation time. Mutually exclusive with `prompt`.
   Path bounded to 1024 characters; content capped at
   128 KiB (matching inline `prompt` limit). Path
   resolution and file access delegated to
   `deps.readText()` at the same privilege level as the
   calling agent — no path traversal escalation beyond
   existing capabilities. Eliminates 470KB of inline
   text from the parent agent's context across 5 review
   runs.

3. **New `dispatch_agent_run` tool** in the review-dispatch
   plugin — accepts `agent` + `prompt`/`promptFile` + `tier`
   (or explicit model). Reads the review matrix to resolve
   the exact model slug deterministically. Calls
   `executeAgentSession()` directly. No LLM in the
   model-resolution or prompt-delivery loop.

4. **Dual-copy sync** — all changes applied to both
   scaffolded (`.opencode/`) and canonical
   (`internal/scaffold/assets/opencode/`) locations.
   Existing drift detection tests enforce parity.

## Capabilities

### New Capabilities
- `executeAgentSession()`: Shared function in
  `.opencode/lib/agent-executor.ts` for reusable child
  session lifecycle management
- `promptFile` on `invoke_agent`: Disk-based prompt delivery
  as alternative to inline `prompt` parameter
- `dispatch_agent_run` tool: Deterministic model resolution
  + disk-based prompt delivery in one tool call, registered
  in the review-dispatch plugin

### Modified Capabilities
- `invoke_agent`: Refactored to delegate session management
  to shared executor; gains `promptFile` parameter; all
  existing behavior preserved via re-exports

### Removed Capabilities
- None — all changes are additive and backward-compatible

## Impact

- **`.opencode/lib/`**: New `agent-executor.ts` with types
  and functions extracted from `invoke-agent/index.ts`
- **`.opencode/plugins/invoke-agent/`**: Refactored to use
  shared executor; shrinks from ~175 to ~50 lines of
  session logic; re-exports moved types for import
  compatibility
- **`.opencode/plugins/review-dispatch/`**: Extended with
  `dispatch_agent_run` tool; gains `client` capture in
  `server()` method
- **`internal/scaffold/assets/opencode/`**: Canonical copies
  of all above changes
- **`.opencode/test/`**: New and updated test files for
  `promptFile`, `dispatch_agent_run`, and integration
- **Dispatch advisor skill**: Will need a follow-up update
  to instruct agents to call `dispatch_agent_run` instead
  of `invoke_agent` for review runs (tracked separately)

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

This change strengthens artifact-based communication by
removing the LLM from the model-resolution data path. The
shared executor produces the same structured
`InvokeAgentResult` artifacts with full provenance metadata.
File-based prompt delivery (`promptFile`) is still
artifact-based — the prompt file is a well-known artifact
that the tool reads directly, eliminating a lossy
intermediary.

### II. Composability First

**Assessment**: PASS

The shared `agent-executor.ts` library is independently
usable by both the `invoke-agent` and `review-dispatch`
plugins. `invoke_agent` continues to work standalone with
inline `prompt`. `dispatch_agent_run` adds a new capability
without modifying existing tool contracts. All changes are
additive — no mandatory dependencies introduced.

### III. Observable Quality

**Assessment**: PASS

The `dispatch_agent_run` tool returns the same structured
`InvokeAgentResult` with full provenance: requested model,
resolved parent model, reported child model, variant,
usage, and error details. Model resolution is deterministic
and auditable through the review matrix rather than opaque
LLM interpretation.

### IV. Testability

**Assessment**: PASS

All components are testable in isolation:
- `executeAgentSession()` accepts injected dependencies
  (`client`, `sessionID`, `directory`, `parentAbort`,
  `timeoutMilliseconds`)
- `promptFile` reading uses the existing `deps.readText()`
  dependency injection
- `dispatch_agent_run` uses the existing `loadPolicies()`
  and `profilePair()` test infrastructure
- New unit tests for executor, `promptFile` validation,
  tier resolution, and session lifecycle

**Coverage strategy**: Unit tests target the new
`agent-executor.ts` exports, `promptFile` path and content
validation (mutual exclusivity, path length, content size,
read failure), and `dispatch_agent_run` handler (tier
resolution, explicit model bypass, default fallback, unknown
agent rejection, provenance metadata). Integration coverage
via existing `make plugin-test` coverage gate.

### V. Security by Default

**Assessment**: PASS

The change introduces two new input surfaces — both secured
by existing mechanisms:
- **`promptFile` path validation**: Bounded to 1024
  characters. Content size capped at 128 KiB (matching
  inline `prompt` limit). Path resolution and file access
  delegated to `deps.readText()`, which operates at the
  same privilege level as the calling agent — no path
  traversal escalation.
- **Review matrix validation**: `dispatch_agent_run` parses
  the review matrix through the existing `loadPolicies()`
  pipeline with schema validation. Agent names validated
  against `parseReviewerManifest()`.
- **No new external dependencies**: All functionality uses
  existing `@opencode-ai/plugin` SDK primitives.
- **Error sanitization**: `sanitizeInvocationError()`
  credential stripping preserved in the shared executor —
  all error responses pass through the same redaction path.
