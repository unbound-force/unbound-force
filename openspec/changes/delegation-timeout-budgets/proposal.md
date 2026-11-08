## Why

`/uf.run-pipeline-plan` delegates a whole `/uf.unleash` run per
work item through `uf_delegate_command` with `timeout: 1800000`,
the executor maximum (`MAX_TIMEOUT_MS`,
`.opencode/lib/run-timeout.ts`). Inside that child, `/uf.unleash`
makes its own `uf_delegate_command` and `dispatch_review_runs`
calls, each of which may also run for up to 30 minutes. The inner
budgets are as large as the outer cap, so the outer timeout aborts
the child partway through nested work, and gate-timeout recovery
re-delegates into the same wall (issue #930).

Evidence from opencode session history (Oct 19 - Nov 2):

- `DELEGATE_FAILED: timeout: invoke_agent run timed out` 275 times
  across 76 sessions; the timeout share rose from about 5% of
  delegations (Oct 22-27) to 35-40% per day (Oct 29 - Nov 2).
- `uf.unleash.md` command-path delegations at `timeout: 1800000`
  timed out 184 of 270 times; one orchestrator lost about 13 hours
  to 26 repeated 30-minute attempts.
- About half of the timed-out children were still active in their
  last two minutes; the other half had stalled on a single model
  turn. The plugin returned the same retryable timeout in both
  cases, so the orchestrator could not tell them apart.

Three defects combine: budgets are flat rather than hierarchical,
the only limit is wall-clock (no notion of progress), and a
timeout result carries no resume state.

## What Changes

1. **Hierarchical budgets.** The shared executor
   (`executeAgentSession`) records each child session's absolute
   deadline in a process-wide budget registry. A nested delegation
   resolves its parent's deadline (registry hit, or a bounded walk
   up the session `parentID` chain) and clamps its own wall-clock
   timeout to the remaining budget minus a reserve. A nested call
   with no usable budget left fails fast (`budget_exhausted`,
   non-retryable) instead of being killed mid-work. Every executor
   caller — `uf_delegate_command`, `invoke_agent`,
   `dispatch_review_runs` — inherits the clamp with no caller
   change.
2. **Idle (no-progress) timeout.** `uf_delegate_command` gains an
   `idle_timeout` (default 10 minutes). The executor polls the
   child's session subtree (the child plus its descendant
   sessions) and aborts only when no message or part changed for
   the idle window. Activity in a nested grandchild counts as
   progress for the parent, so a working child is never aborted
   by the idle clock.
3. **Lifted cap for command-path delegations.** Delegations that
   carry `command_path` (commands that own durable checkpoints)
   accept `timeout` up to 4 hours (`14400000`). Prompt-only
   delegations keep the 30-minute ceiling. The idle timeout is the
   primary stall detector; the wall clock becomes a backstop.
4. **Resume-carrying timeout results.** Timeout failure envelopes
   carry a `timeout` object (`kind`: `wall_clock` | `idle` |
   `budget`, effective and requested limits, elapsed time,
   `child_active`, `pending_question`) and, for command-path
   delegations, a `resume` object (state key, last checkpoint
   step, completed steps) read from the progress state store.
5. **Recovery prose.** `/uf.run-pipeline-plan` delegates
   `/uf.unleash` with `timeout: 14400000`, and its timeout
   recovery branches on `timeout.kind`: a pending question is a
   gate wedge (existing gate-timeout recovery); an idle stall gets
   one resume re-delegation that re-anchors from durable state; a
   wall-clock exhaustion goes to the blocked-item protocol. Fresh
   restarts of an item that has a recorded checkpoint are
   forbidden.

## Capabilities

### New Capabilities

- `delegation-timeouts`: hierarchical budgets, idle timeout,
  command-path ceiling, and resume-carrying timeout results for
  the shared agent executor and `uf_delegate_command`.

### Modified Capabilities

- `run-plan-command`: `/uf.run-pipeline-plan` delegation timeout
  and timeout-recovery protocol.

### Removed Capabilities

- None.

## Impact

- `.opencode/lib/delegation-budget.ts` (new): budget registry,
  ancestry resolution, clamp, subtree activity probe.
- `.opencode/lib/agent-executor.ts`: budget clamp, deadline
  registration, idle watchdog, timeout details on results.
- `.opencode/plugins/uf-workflow/index.ts`: `uf_delegate_command`
  schema (`timeout` ceiling by delegation kind, `idle_timeout`),
  failure envelope (`timeout`, `resume`). Edits stay inside the
  `uf_delegate_command` section.
- `.opencode/commands/uf.run-pipeline-plan.md`: delegation timeout
  and recovery prose.
- Byte-identical mirrors under `internal/scaffold/assets/opencode/`.
- Tests: `.opencode/test/delegation-budget.test.ts` (new),
  executor and delegate tests.
- Documentation: user-facing reference for `idle_timeout`, the
  timeout ceilings, the timeout envelope fields, and the
  recovery protocol tracked as unbound-force/website#381
  (AGENTS.md documentation gate; constitution Cross-Repo
  Documentation rule). Supersedes the "wedged at the gate
  until timeout" assessment in #838, whose wedge now ends
  as an idle timeout with `pending_question: true`.
- No change to `invoke_agent`'s or `dispatch_review_runs`' input
  bounds, to `run-timeout.ts` derivation (its record type gains
  optional executor-outcome audit fields), or to any quality
  gate (coverage floor, CRAP, review limits).

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.4.0).

### I. Autonomous Collaboration

**Assessment**: PASS

Timeout results become self-describing artifacts: the envelope
states which limit fired, whether the child was active, and the
last durable checkpoint, so the orchestrator acts on published
data instead of inferring from a bare error string.

### II. Composability First

**Assessment**: PASS

The budget lives in the shared executor, so every tool that runs a
child session composes correctly when nested, with no caller-side
coordination. Top-level calls (no budgeted ancestor) behave as
before.

### III. Observable Quality

**Assessment**: PASS

Each timeout result records the effective and requested limits,
whether the budget clamp applied, elapsed time, and the activity
verdict, making the 35-40% timeout rate diagnosable per call.

### IV. Testability

**Assessment**: PASS

The registry, clamp, and fingerprint are pure or injectable; the
idle watchdog takes an injected poll interval and clock; probe
APIs are reached through the injected OpenCode client, so every
path is covered with fakes.

### V. Security by Default

**Assessment**: PASS

No new dependencies or trust boundaries. The activity probe reads
only session metadata the plugin already has client access to and
never surfaces message content; the resume object carries step
names and the branch-derived state key, not command output.

### VI. Deterministic Execution

**Assessment**: PASS

Timeout classification and resume state are computed by the tool
and returned as structured fields; the command's recovery prose
branches on those fields rather than asking the agent to inspect
session history or reconstruct progress.
