## Context

The `invoke-agent` plugin (`invoke_agent` tool) currently
embeds ~152 lines of session management logic: abort
controller setup, child session create/prompt, response
parsing (text extraction, usage accumulation), timeout
orchestration, and cleanup. This logic is not reusable by
other plugins.

The `review-dispatch` plugin needs to invoke child sessions
for the new `dispatch_agent_run` tool. Without extraction,
it would duplicate the entire session lifecycle — violating
DRY and creating maintenance divergence.

Additionally, `invoke_agent` only accepts inline `prompt`
strings. Review-council workflows build ~94KB prompt files
on disk, then the parent agent must read each file into
context and reproduce it verbatim in each tool call
(5 x 94KB = 470KB inline text), causing context exhaustion
and frequent failures.

## Goals / Non-Goals

### Goals
- Extract reusable session management to a shared library
  that both `invoke-agent` and `review-dispatch` can use
- Add `promptFile` to `invoke_agent` for disk-based prompt
  delivery, eliminating context bloat
- Create `dispatch_agent_run` tool that resolves models
  deterministically from the review matrix, removing the
  LLM from the model-resolution data path
- Maintain full backward compatibility — existing callers
  of `invoke_agent` with inline `prompt` work unchanged
- Preserve import compatibility via re-exports from
  `invoke-agent/index.ts`

### Non-Goals
- Changing the dispatch-advisor skill to use
  `dispatch_agent_run` (tracked separately)
- Modifying `plan_review_dispatch` or
  `finalize_review_dispatch` behavior
- Changing the review matrix schema or reviewer manifest
  format
- Adding batched/parallel invocation to the executor

## Decisions

### D1: Shared executor in `.opencode/lib/agent-executor.ts`

Extract session lifecycle management into a single
`executeAgentSession()` function. This function owns:
abort controller setup, child session create, prompt
with model/variant, response parsing (text, usage,
mismatch check), timeout, cancellation, and cleanup.

It does NOT own: input validation, manifest checking,
model resolution, host introspection, prompt file reading.

Each consumer validates inputs and resolves models using
its own policies, then delegates session execution to the
shared function. This separation aligns with Composability
First — the executor is independently usable without
requiring review-dispatch or invoke-agent internals.

### D2: Types and functions that move to the library

**Types**: `ModelIdentity`, `InvocationTokens`,
`InvocationUsage`, `InvocationError`,
`InvocationProvenance`, `InvokeAgentResult`

**Functions**: `errorText()`, `sanitizeInvocationError()`,
`responseFailure()`, `directModelIdentity()`,
`reportedModel()`, `failedResult()`, `invocationUsage()`,
`responseText()`, `validateTimeout()`

**Re-exports**: `invoke-agent/index.ts` re-exports all
moved types and `sanitizeInvocationError` so existing test
imports (`import { ... } from "./invoke-agent/index.js"`)
remain stable without requiring import path changes.

### D3: `promptFile` as mutually exclusive alternative

Add an optional `promptFile: string` parameter (max 1024
chars) to `invoke_agent`'s input schema. Exactly one of
`prompt` or `promptFile` is required — enforced via
`.superRefine()`. When `promptFile` is provided, the
plugin reads the file using the injected `deps.readText()`
dependency at invocation time. Everything downstream uses
the resolved prompt text — no other changes needed.

This aligns with Observable Quality — the prompt file is
a traceable artifact on disk, readable by any consumer.

### D4: `dispatch_agent_run` tool registration

Register the new tool in the `review-dispatch` plugin's
`server()` method. The plugin already captures `readText`
and `parseYaml` from `input.client`; it must additionally
capture `client` itself for session management.

The tool accepts `agent` + `prompt`/`promptFile` + `tier`
(or explicit `model`). It reads the review matrix using
existing `loadPolicies()` and `profilePair()` to resolve
the full model slug including `@suffix` deterministically.
Then it calls `executeAgentSession()` directly.

### D5: Manifest validation in `dispatch_agent_run`

`dispatch_agent_run` MUST validate the agent against the
reviewer manifest (`.uf/reviewer-capabilities.yaml`), same
as `invoke_agent`. Defense in depth: the dispatch plan
validated the agent, but the tool is independently callable
and must enforce its own invariants. The file read cost is
negligible compared to the child session cost.

### D6: Dual-copy sync maintained

Every file change applies to both:
- `.opencode/lib/` and `.opencode/plugins/` (scaffolded)
- `internal/scaffold/assets/opencode/lib/` and
  `internal/scaffold/assets/opencode/plugins/` (canonical)

Existing drift detection tests enforce parity. No special
handling needed beyond keeping both copies in sync.

### D7: `dispatch_agent_run` as tool name

Use `dispatch_agent_run` — it clearly conveys "dispatch
policy resolves the model, then runs the agent." The
`dispatch_` prefix signals that model resolution happens
inside the tool, distinguishing it from `invoke_agent`
where the caller must supply the model.

### D8: `tier` parameter name (not `profile`)

Use `tier` to match the existing `TierSchema` enum used
throughout the codebase (`"lightweight" | "standard" |
"heavy"`). While the matrix config calls them `profiles`,
the dispatch pipeline consistently calls them `tiers` in
its public API. Consistency with existing callers wins.

### D9: Agent definition reinforcement via active file read

When an orchestrator invokes a Divisor child agent, the agent
definition IS delivered as a system prompt (via `body.agent` in
`executeAgentSession()`). However, the orchestrator-constructed
user prompt is comprehensive enough that child agents follow it
exclusively — ignoring critical system prompt sections like
Step 0 (Prior Learnings), Source Documents, and Convention Pack
markers.

The fix adds a single bullet to each orchestrator's mandatory
child prompt ingredients: instruct the child agent to read its
own agent definition file at `.opencode/agents/{agent}.md` as
Step 0. This reinforces the system prompt content through
active retrieval — the child agent processes the definition as
part of its task flow, not as passive background context.

This approach was chosen over two alternatives:

1. **"Follow your system prompt"** — too vague; LLMs
   de-prioritize system prompt when detailed user prompt
   exists.
2. **Inline agent definition content** — 197-285 lines per
   agent × 5 agents = ~1,250 extra lines; risks exceeding
   the 128 KiB prompt limit and creates maintenance coupling
   between orchestrator and agent definitions.

Active file read adds one tool call per child run (negligible
cost vs. the review session itself) and self-heals as agent
definitions evolve.

## Risks / Trade-offs

### R1: Re-export maintenance

Re-exports from `invoke-agent/index.ts` add a maintenance
obligation — if the shared library renames a type, both
locations need updating. Mitigated by: TypeScript compiler
catches broken re-exports at build time, and the type names
are stable (established in the multi-model-fanout change).

### R2: Two invocation paths

After this change, agents can invoke child sessions via
either `invoke_agent` or `dispatch_agent_run`. The
dispatch-advisor skill update (tracked separately) will
guide agents to the correct tool. During the transition,
both paths work correctly — `invoke_agent` with explicit
model is equivalent to `dispatch_agent_run` when the caller
gets the slug right.

### R3: `promptFile` path traversal

`promptFile` accepts an absolute path and reads from disk
via `deps.readText()`. The path is bounded to 1024 chars.
The file is read by the plugin process (same privilege as
the agent), not elevated. The agent already has filesystem
access, so `promptFile` does not expand the attack surface.

### R4: Plugin `client` capture scope

The review-dispatch plugin's `server()` method must
capture `input.client` (not just `readText`/`parseYaml`)
for `dispatch_agent_run` to create child sessions. This
is the same pattern used by `invoke-agent`. The captured
client reference is used only within tool handlers, not
stored globally.

### R5: Agent definition read may fail

If the child agent cannot read its definition file (file
missing, permission denied), the reinforcement instruction
degrades gracefully — the child still has its system prompt
(delivered via `body.agent`). The instruction uses SHOULD
language for executing Step 0 and Source Documents, so a
failed read does not hard-fail the review run. The agent
definition files are scaffolded assets that MUST exist in a
properly initialized workspace.
