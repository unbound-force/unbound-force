# per-task-delegation-budgets — Technical Design

## Context

`/uf.unleash` implementation phase fans out each `tasks.md` phase as a set
of prompt-only `uf_delegate_command` delegations (`checkpoint_step: "implement"`,
`phase: N`, no `command_path`). Each delegation is bounded by the executor's
single wall-clock timeout (`MAX_TIMEOUT_MS = 1_800_000 ms` in
`.opencode/lib/run-timeout.ts` and `agent-executor.ts`) with no idle watchdog
and no resume payload.

Issue #950 lifted the ceiling to 4 hours and added `timeout` + `resume`
envelopes only for `command_path` delegations. Prompt-only per-task delegations
did not inherit them. Evidence from issue #1080 — child
`ses_e24e418a4ffeMlFj7pkRMd0V0A` (dewey#203), 2026-11-14 17:30–21:29 UTC — shows
4 of 11 per-task delegations hit the 30-minute wall clock (2 while
`child_active: true`) and lost ~1 h 50 m re-doing work before the outer 4-hour
budget was exhausted. The current tool layer already has the building blocks
(`run-timeout.ts` derivation per #849, `delegation-budget.ts` registry +
`probeSubtreeActivity` per #930/#950, executor `timeout` envelope per #930)
but they are not wired to this call site.

This change wires those blocks with no new concepts, preserves every
gatekeeping invariant (the 1_800_000 ms `TimeoutSchema` bound, hierarchical
clamping, `budget_exhausted` non-retryable), and keeps the fix inside the
tool layer so command prose remains a thin branching layer.

## Goals / Non-Goals

### Goals

- Replace the flat 30-minute per-task default with a deterministic,
  work-derived wall clock bounded by `MAX_TIMEOUT_MS` unless explicitly lifted;
  provenance (inputs + clamp) is observable per RD-FR-009.
- Make idle detection the primary stall signal for per-task children running
  under a budgeted `/uf.unleash` parent; wall-clock remains an absolute
  backstop; hierarchical `budget_exhausted` semantics are unchanged.
- Return a lightweight, advisory `resume` hint (files modified + tasks completed)
  for prompt-only delegations carrying phase context, reusing already-observable
  state so retries can re-anchor without a blank-context restart.
- Surface remaining parent budget at each implement-phase checkpoint so the
  orchestrator can plan a resume boundary before the outer budget expires.

### Non-Goals

- Changing the `TimeoutSchema` / `validateTimeout` 1_800_000 ms bound, any
  coverage / CRAP / review gate, or the canonical `.uf/review-matrix.yaml`
  limit values (gatekeeping).
- Adding a new checkpoint type or a semantic merge of timed-out worktree
  edits (resume hint is advisory; caller decides).
- Introducing a new global idle default outside the existing budget registry;
  tasks inherit idle semantics only via a budgeted ancestor.
- Rewriting the per-step orchestrator or the review-dispatch derivation
  (both unchanged; review-dispatch contention and tier logic is reused as a
  helper only).

## Decisions

### D1: Per-task derivation variant in `run-timeout.ts`

A pure function `derivePerTaskTimeoutMs(signals)` is added alongside
`derivePerRunTimeoutMs` from #849. Inputs are phase size (task count in the
phase), files touched (from `parse_tasks` metadata or the phase's `files`
list), and model throughput tier (`Tier.TermMs` in `run-timeout.ts`). The
function is deterministic (no I/O, no clock), monotonic in work size, and
applies contention only when the implement phase fans out multiple concurrent
children (reused contention factor 0.15, capped at 1.5 from #849).

The derived value is clamped to `min(configuredCeilingMs, 1_800_000)` with
floor-at-base precedence: when the ceiling is below the resolved base window,
the base window wins. An explicit `timeout` argument still overrides
derivation entirely. Derivation + clamp outcome is threaded as timeout
provenance on the executor result, matching RD-FR-009.

Alternative rejected: asking the LLM to estimate duration — non-deterministic,
violates Deterministic Execution (VI), and repeats the #849 lesson.

### D2: Idle inheritance via the delegation-budget registry

Per-task delegations run inside a command-path `/uf.unleash` child that already
registers a 4-hour deadline and, when `idle_timeout` is set, an idle watchdog.
The executor's existing `resolveInheritedDeadline` + `clampToBudget` +
`probeSubtreeActivity` path handles nesting: before creating a per-task child,
the executor resolves the parent deadline from the `globalThis` symbol-keyed
registry (direct hit on `deps.sessionID`, else walk `parentID` up to 8 levels).
The per-task child's effective wall clock is `min(requested, deadline - now -
60s reserve)`; below 30 s remainder it fails fast with `budget_exhausted`
(non-retryable).

When the parent carries an `idle_timeout` (default 600_000 ms preserved from
#930), the idle watchdog is inherited by the per-task child via the same
registry entry. A child with `child_active: true` (fingerprint change within
the last 2 minutes per D4 in #950) or a running tool part (per
`idle-timeout-aborts-children-waiting-on`) is not aborted by wall-clock while
progressing. After `idle_timeout` of no subtree change with at least one
confirming successful probe, the executor aborts with `idle_timeout`
(retryable). Failed probes neither reset nor advance the clock. Wall-clock
remains the absolute backstop; `budget_exhausted` clamping is unchanged.

Alternative rejected: adding a new per-task idle default outside the registry
— would create a second source of truth and change top-level behavior;
requiring every per-task caller to pass `idle_timeout` — breaks the thin
`/uf.unleash.md` prose and duplicates registry logic.

### D3: Lightweight resume hint for prompt-only delegations

When a per-task `uf_delegate_command` is invoked with
`checkpoint_step: "implement"` and `phase`, the executor captures advisory
state on timeout:

- `files_modified`: files touched by the child (from the worktree
  cherry-pick diff or executor's existing file tracking — no new file-system
  watcher).
- `tasks_completed`: `tasks.md` checkbox deltas since delegation start
  (`parse_tasks` before/after or the staged diff already emitted by the
  executor).

Both are already observable from the existing merge path; the hint is assembled
by diffing that state, not by semantic analysis. The hint is returned inside
the existing `DELEGATE_FAILED` `timeout.resume` envelope (or a new optional
`sibling` field on it) only for this call site; other prompt-only callers are
unaffected. An unreadable state yields `resume: null` (fail-closed, as in D4 of
#950). Command prose MUST NOT retry from blank when `child_active: true` or a
resume hint is present — it must relay files/tasks to the retry prompt.

Alternative rejected: full semantic merge of timed-out worktree — complex,
error-prone, and outside Tier 1 scope; returning raw session transcripts —
leaks message content and violates Security by Default (V).

### D4: Outer-budget visibility at phase checkpoints

`/uf.unleash` already emits a `checkpoint` with `phase` payload and a
human-readable checkpoint line per `implement` phase. This change amends that
rendering to include remaining parent budget when a budgeted ancestor exists
(`resolveInheritedDeadline` returns non-null): `remaining_ms` in the machine
`phase` payload and `remaining budget: ~Xm` in the human line. No new
checkpoint type or state file is introduced. When no budgeted ancestor exists,
no budget line is emitted (top-level behavior unchanged).

This is the only change to `uf.unleash.md` prose beyond the per-task timeout
derivation call site and the resume-hint relay branch. Actual prose deltas
are ~+20–40 lines; the PR records the final diff.

## Risks / Trade-offs

- **Under-derived timeouts still hit wall-clock** — mitigated by D2: a working
  child is not killed while `child_active`; the derived floor never goes below
  the base window, and explicit `timeout` overrides remain.
- **Over-derived timeouts near outer budget** — hierarchical clamping (60 s
  reserve, 30 s minimum) turns excess into `budget_exhausted` before spawning;
  the phase-checkpoint budget line (D4) lets the orchestrator plan a resume
  boundary proactively.
- **Resume hint staleness** — files/tasks are advisory snapshots at abort
  time; the retry prompt re-anchors from them but does not auto-merge edits.
  This is intentional — avoids semantic conflicts while eliminating blank
  re-derivation.
- **Probe cost** — bounded by the existing 32-session cap and 30 s poll
  interval from #930; per-task children only add depth 1 under the same
  parent, so no new amplification.
- **Constitution alignment** — derivation is pure and Composable (II, VI);
  provenance + budget line are machine-parseable artifacts (I, III); no
  secrets or new deps (V); all new branches are testable with injected fakes
  (IV). See proposal.md Constitution Alignment for per-principle PASS.
