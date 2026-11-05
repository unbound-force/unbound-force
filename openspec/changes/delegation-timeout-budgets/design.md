## Context

Every child session in the review and workflow plugins runs
through `executeAgentSession` (`.opencode/lib/agent-executor.ts`).
It creates a parented child session, sends one blocking
`session.prompt`, and aborts the child when a single wall-clock
`setTimeout` fires. Callers pass that timeout: `invoke_agent`
(bounded 30 minutes), `dispatch_review_runs` (derived per run by
`run-timeout.ts`, ceiling 30 minutes), and `uf_delegate_command`
(bounded 30 minutes, default 10).

`/uf.run-pipeline-plan` delegates a whole `/uf.unleash` at the
30-minute maximum. The child makes nested executor calls of its
own, each with up to 30 minutes, so the outer timer fires while
nested work is still running (issue #930). The executor has no
knowledge of its position in the session tree, no notion of
progress, and returns the same `timeout` error whether the child
was busy or stalled.

## Goals / Non-Goals

### Goals

- A nested executor call never outlives its parent's budget.
- A child that is making progress is not aborted for slowness
  alone; a stalled child is detected well before the wall clock.
- Whole-command delegations that own durable checkpoints can run
  long enough to finish.
- A timeout result tells the caller which limit fired, whether the
  child was active, and where durable progress stands.
- `/uf.run-pipeline-plan` resumes timed-out items instead of
  restarting them.

### Non-Goals

- Changing `run-timeout.ts` derivation or the `invoke_agent` and
  `dispatch_review_runs` input bounds.
- Step-wise driving of `/uf.unleash` from the orchestrator (D3).
- Surfacing question-tool prompts from parented children (a host
  limitation; this change only detects the wedge faster).

## Decisions

### D1: Hierarchical budgets via a process-wide deadline registry

The executor records `childSessionID -> absolute deadline` in a
registry when it creates a child, and removes the entry when the
run ends. Before creating a child, it resolves the inherited
deadline for its own parent session (`deps.sessionID`):

1. Registry hit on `deps.sessionID` (the common case: a tool
   called directly inside an executor-created child).
2. Otherwise, walk `session.get(...).parentID` up to 8 levels and
   use the first registered ancestor (covers sessions the host's
   own task tool created under a budgeted child).
3. No registered ancestor, or any lookup error: no inherited
   budget (top-level behavior is unchanged).

With an inherited deadline, the effective wall clock is
`min(requested, deadline - now - 60s reserve)`. The reserve gives
the parent time to receive and act on the nested failure. When the
remaining budget is below 30 seconds after the reserve, the call
fails before spawning with `budget_exhausted` (non-retryable: the
parent should return, not retry into the same deadline).

The registry is stored on `globalThis` under
`Symbol.for("unbound-force.delegation-budget")`. The plugins
(`invoke-agent`, `review-dispatch`, `uf-workflow`) are separate
modules that import the shared library; a symbol-keyed global
guarantees one registry even if the host loads the library more
than once. This is deliberate process-wide state, the only way for
a nested tool call in a different plugin to see its ancestor's
deadline; it holds only session ids and numbers and entries are
removed in `finally`.

Rejected: passing the budget through prompts or tool arguments
(the child agent would have to relay it faithfully — not
deterministic); making the parent cap the sum of child caps (the
number of nested calls is unknown in advance).

### D2: Idle timeout distinct from the wall clock

`ExecutorDependencies` gains an optional `idleTimeoutMilliseconds`.
When set, the executor polls the child's session subtree every 30
seconds (injectable) and computes an activity fingerprint:

- Depth-first over `session.children`, starting at the child,
  newest child first, bounded to 32 sessions. Depth-first matters:
  a long-lived child accumulates finished direct children
  (classifier and reviewer sessions), and a breadth-first walk
  would spend the cap on them before reaching the active
  grandchild under the newest one.
- For each session, the latest message (`session.messages` with
  `limit: 1`): message id, completion time, error presence, part
  count, and the last part's id, type, tool state status, tool
  title, and the combined length of its text, `state.output`, and
  serialized `state.metadata`.

A changed fingerprint is progress. Streaming tokens grow the text
length; tool calls start and finish; a running tool streams
through its metadata (bash keeps live output in
`state.metadata.output`, since the running state has no `output`
field), so a long test run that keeps printing is progress; nested
sessions appear. A
parent blocked on a nested `uf_delegate_command` is credited with
its grandchild's progress, which is how the parent's limit
"accounts for" its children. When no change has been observed for
the idle window, the executor aborts with `idle_timeout`.

A failed probe neither resets nor advances the idle clock; an
idle abort additionally requires at least one successful probe
since the last observed change, so an unreachable API never kills
a working child. Clients without the probe endpoints (test fakes)
never trip the idle path. A probe still in flight when the child's
prompt returns is ignored (a `settled` flag set on return and in
`finally`), so it cannot abort a finished child or relabel its
result.

A running non-question tool does not pause the idle clock. Pausing
would also hide the permission-ask wedge (a parented child blocked
on a permission prompt sits in a running tool with no output) for
the full wall clock; a silent tool longer than the idle window is
instead handled by the per-call `idle_timeout` override.

The probe also reports `pending_question`: the latest part of some
session in the subtree is a running `question` tool call. That is
the question-gate wedge `/uf.run-pipeline-plan` documents, now
detectable in one idle window instead of a full wall clock.

`uf_delegate_command` enables the watchdog for every call, default
`idle_timeout` 600,000 ms (10 minutes), range 60,000 - 1,800,000.
Session history showed stalls of 2-10+ minutes on a single model
turn; 10 minutes tolerates long quiet tool runs (race tests) while
cutting a stall far sooner than a multi-hour wall clock.
`invoke_agent` and `dispatch_review_runs` do not enable it (their
runs are single review turns already bounded by derived limits).

### D3: Lifted cap for command-path delegations, not step-wise driving

`uf_delegate_command` accepts `timeout` up to 14,400,000 ms
(4 hours) when `command_path` is present; prompt-only delegations
keep the 1,800,000 ms ceiling, enforced in the schema's
`superRefine`. The executor itself has no ceiling of its own;
input bounds stay at each tool boundary.

Command files such as `/uf.unleash` record durable checkpoints and
resume from them, so a long-running command-path child loses
nothing if it is eventually stopped. With D2 detecting stalls, the
wall clock only needs to be a generous backstop.

Rejected: having the orchestrator drive `/uf.unleash` one step at a
time. It would duplicate the command's own step machine in
orchestrator prose (a determinism-gate violation) and reintroduce
the per-step relay the run-pipeline-plan rewrite removed.

### D4: Timeout results carry classification and resume state

`InvokeAgentResult` gains an optional `timeout` object when a run
ends by a limit:

| Field | Meaning |
|---|---|
| `kind` | `wall_clock`, `idle`, or `budget` |
| `effective_ms` | Wall clock actually applied |
| `requested_ms` | Wall clock the caller asked for |
| `budget_clamped` | Effective limit was reduced by D1 |
| `idle_timeout_ms` | Idle window, or null when disabled |
| `elapsed_ms` | Time from start to abort |
| `last_progress_ms_ago` | Time since last observed change, or null |
| `child_active` | Progress seen in the last 2 minutes; null when never probed |
| `pending_question` | A question tool call was pending at last probe |

The executor error codes are `timeout` (wall clock, retryable),
`idle_timeout` (retryable), and `budget_exhausted`
(non-retryable). `uf_delegate_command` threads `timeout` into its
`DELEGATE_FAILED` envelope beside `child_session_id`, and for
command-path delegations adds `resume`:

| Field | Meaning |
|---|---|
| `command` | State command name (command file basename) |
| `key` | Progress state key (sanitized current branch) |
| `last_checkpoint_step` | `current_step`, or null with no state |
| `completed_steps` | Recorded steps, empty with no state |

`resume` is read through the same `readState` path `get_progress`
uses, after the child stops, so it reflects the child's last
durable write. A read failure yields `resume: null` rather than
masking the timeout.

### D5: Recovery protocol in `/uf.run-pipeline-plan`

Delegation 1 sets `timeout: 14400000` and omits `idle_timeout`.
On `DELEGATE_FAILED` with a `timeout` object:

- `pending_question: true` -> gate wedge: existing gate-timeout
  recovery (human clears the gate, or bounded direct drive).
- `kind: "idle"` without a pending question -> provider stall:
  re-delegate once as a resume (the mid-pipeline re-run class:
  same `command_path`, no routing context, prompt carries the
  description, issue reference(s), and the `resume` step); a
  second idle timeout on the same item goes to the blocked-item
  protocol.
- `kind: "wall_clock"` -> blocked-item protocol, relaying
  `child_session_id`, `child_active`, and
  `resume.last_checkpoint_step`.

When `resume.last_checkpoint_step` is non-null, the orchestrator
MUST NOT restart the item fresh on `main`. Delegation 2
(`/uf.finale`) keeps `timeout: 600000`; its designed commit-gate
wedge now usually ends as an idle timeout with
`pending_question: true`, which the same branch handles. Because
the idle window must be shorter than the wall clock to fire,
Delegation 2 sets `idle_timeout: 300000` (5 minutes): long enough
for staging and commit-message preparation, and it ends the gate
wait with `pending_question: true` before the 10-minute wall
clock. `kind: "budget"` cannot occur at the orchestrator (it has
no budgeted parent) and is treated as `wall_clock`. An idle
timeout whose `resume.last_checkpoint_step` is null resumes from
nothing, so that re-run is effectively a fresh start and its
prompt re-carries the description and issue reference(s).

## Risks / Trade-offs

- **False idle aborts** for a tool that runs silently longer than
  the idle window. Mitigation: 10-minute default, per-call
  override up to 30 minutes, and abort only after a confirming
  probe.
- **Probe cost**: two requests per probed session (`messages` and
  `children`), so up to 64 requests per child every 30 seconds.
  Bounded by the 32-session cap and only for
  `uf_delegate_command` children.
- **Clamped review-run provenance** (resolved): the run file now
  records the executor outcome alongside the derived provenance —
  `executor_effective_ms` and `executor_budget_clamped` are copied
  from the executor result's timeout details, so a clamped run no
  longer overstates its applied limit (`timeout.effective_ms`
  remains the derived pre-clamp value for provenance).
  `dispatch_status` prefers `executor_effective_ms` when present;
  the consolidate dispatchRuns mapping omits both fields
  (run-file audit only).
- **`messages` ordering**: the probe relies on `limit` returning
  the most recent messages (OpenCode 1.19 behavior). If a host
  returned the oldest instead, the fingerprint would only change
  through new descendant sessions and could idle-abort a busy
  child; the per-call `idle_timeout` override is the escape hatch.
- **Clamped review runs** inside a nearly exhausted parent fail
  fast instead of running; that is intended — the parent's result
  then carries resume state for a clean resume.
- **Registry leaks** are bounded: entries are removed in
  `finally`, and a host crash discards the process anyway.
