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

## 1. Extract shared agent executor library

- [x] 1.1 Create `.opencode/lib/agent-executor.ts` with
  types (`ModelIdentity`, `InvocationTokens`,
  `InvocationUsage`, `InvocationError`,
  `InvocationProvenance`, `InvokeAgentResult`) and utility
  functions (`errorText`, `sanitizeInvocationError`,
  `responseFailure`, `directModelIdentity`, `reportedModel`,
  `failedResult`, `invocationUsage`, `responseText`,
  `validateTimeout`) extracted from
  `.opencode/plugins/invoke-agent/index.ts`. Add the
  `ExecutorDependencies` interface and
  `executeAgentSession()` function encapsulating the full
  child session lifecycle (abort controller, create, prompt,
  response parsing, timeout, cancellation, cleanup).
- [x] 1.2 Refactor `.opencode/plugins/invoke-agent/index.ts`
  to import types and functions from
  `../../lib/agent-executor.js`. Remove moved code. Add
  re-exports of all moved types and
  `sanitizeInvocationError` so existing imports remain
  stable. Retain `InvokeAgentDependencies`,
  `createInvokeAgentTool()`, input schema, host model
  resolution, and manifest validation.
- [x] 1.3 Verify existing `invoke_agent` behavior is
  preserved: run `make plugin-test` and confirm all
  existing invoke-agent tests pass without modification.

## 2. Add promptFile parameter to invoke_agent

- [x] 2.1 Update the `InvokeAgentInputSchema` in
  `.opencode/plugins/invoke-agent/index.ts`: make `prompt`
  optional, add `promptFile` (string, max 1024 chars), add
  `superRefine` enforcing mutual exclusivity (exactly one
  of `prompt` or `promptFile` required). Update the
  `invokeAgent()` function to resolve `promptFile` via
  `deps.readText()` when present, handling read failures
  with sanitized error responses.
- [x] 2.2 [P] Add unit tests in `.opencode/test/` for
  `promptFile` validation: mutual exclusivity enforcement,
  path length limit, successful file read, read failure
  handling, and backward compatibility with inline
  `prompt`.

## 3. Create dispatch_agent_run tool

- [x] 3.1 Update `.opencode/plugins/review-dispatch/index.ts`
  `server()` method to capture `input.client` alongside
  existing `readText` and `parseYaml` captures.
- [x] 3.2 Register the `dispatch_agent_run` tool in the
  review-dispatch plugin. Implement the input schema:
  `agent` (required, same regex), `prompt`/`promptFile`
  (mutually exclusive), `tier` (optional, enum), `model`
  (optional, mutually exclusive with `tier`), `variant`
  (optional), `read_only` (optional, default true),
  `timeout` (optional). Default to `"standard"` tier when
  neither `tier` nor `model` is provided.
- [x] 3.3 Implement tier-based model resolution: load the
  review matrix via `loadPolicies()`, resolve tier to model
  slug via `profilePair()`, preserving `@suffix`. Validate
  agent against reviewer manifest via
  `parseReviewerManifest()`. Resolve `promptFile` if
  present. Call `executeAgentSession()` and return the
  `InvokeAgentResult` with full provenance.
- [x] 3.4 [P] Add unit tests in `.opencode/test/` for
  `dispatch_agent_run`: tier-based model resolution,
  explicit model bypass, default tier fallback,
  tier/model mutual exclusivity, unknown agent rejection,
  missing review matrix handling, and provenance metadata
  correctness.

## 4. Dual-copy sync

- [x] 4.1 [P] Copy `.opencode/lib/agent-executor.ts` to
  `internal/scaffold/assets/opencode/lib/agent-executor.ts`.
- [x] 4.2 [P] Copy updated
  `.opencode/plugins/invoke-agent/index.ts` to
  `internal/scaffold/assets/opencode/plugins/invoke-agent/index.ts`.
- [x] 4.3 [P] Copy updated
  `.opencode/plugins/review-dispatch/index.ts` to
  `internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts`.
- [x] 4.4 Run drift detection tests (`make test`) to
  confirm scaffolded and canonical copies are in sync.

## 5. Verification

- [x] 5.1 Run `make plugin-test` — all TypeScript plugin
  tests pass (unit, integration, smoke, coverage).
- [x] 5.2 Run `make test` — all Go tests pass including
  drift detection.
- [x] 5.3 Run `make lint` — no lint errors.
- [x] 5.4 Verify constitution alignment:
  I. Autonomous Collaboration — shared executor
  communicates via structured `InvokeAgentResult`
  artifacts, not runtime coupling.
  II. Composability First — `agent-executor.ts` is
  independently usable by any plugin without requiring
  `invoke-agent` or `review-dispatch` internals.
  III. Observable Quality — `dispatch_agent_run` returns
  full provenance metadata; model resolution is
  deterministic and auditable.
  IV. Testability — all components accept injected
  dependencies and are testable in isolation.
  V. Security by Default — `promptFile` path validation
  bounded to 1024 chars with 128 KiB content limit;
  `dispatch_agent_run` uses existing `loadPolicies()`
  validation; `sanitizeInvocationError` credential
  stripping preserved in shared executor.

## 6. Agent definition reinforcement in orchestrator prompts

- [x] 6.1 Edit `uf.review-council.md` (canonical at
  `internal/scaffold/assets/opencode/commands/`) to add an
  agent definition reinforcement bullet to the mandatory child
  prompt ingredients (after the existing list). The instruction
  MUST tell the child agent to read its own agent definition
  file at `.opencode/agents/{agent}.md` as Step 0 before
  conducting the review, executing Prior Learnings queries and
  loading Source Documents defined therein.
- [x] 6.2 [P] Apply the same agent definition reinforcement
  bullet to `uf.review-pr.md` (canonical).
- [x] 6.3 [P] Apply the same agent definition reinforcement
  bullet to `uf.triage-issue.md` (canonical).
- [x] 6.4 [P] Apply the same agent definition reinforcement
  bullet to `uf.address-feedback.md` (canonical).
- [x] 6.5 Dual-copy sync: copy all four updated canonical
  command files to their scaffolded locations under
  `.opencode/commands/`.
- [x] 6.6 Run `make test` to verify scaffold drift detection
  passes and all Go tests pass.

<!-- spec-review: passed -->
<!-- code-review: passed -->
