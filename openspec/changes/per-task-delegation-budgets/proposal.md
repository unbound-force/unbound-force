# per-task-delegation-budgets — Change Proposal

## Why

`/uf.unleash` implementation phase delegates each task through `uf_delegate_command`
**prompt-only** (no `command_path`). #950 lifted the wall-clock ceiling to 4 h
(`14_400_000 ms`) and added resume envelopes **only for command-path**
delegations. Prompt-only per-task delegations therefore still use the
fixed `1_800_000 ms` (30-minute) limit from `.opencode/lib/run-timeout.ts` /
`agent-executor.ts` and carry no `resume` information.

Evidence from issue #1080 — `/uf.unleash` child `ses_e24e418a4ffeMlFj7pkRMd0V0A`
(dewey#203), 2026-11-14 17:30–21:29 UTC, dev build containing #950:

| Task | Result | `child_active` | `last_progress_ms_ago` | Effective budget | Notes |
|---|---|---|---|---|---|
| 1.2 wire helper into ingestion | `wall_clock` 30 min | `true` | 29 s | 30 min fixed | retry 12 min |
| 2.1 git resolver | `wall_clock` 30 min | `false` | 239 s | 30 min fixed | retry 20 min |
| 2.3 batch ls-tree fast path | `wall_clock` 30 min | `false` | 329 s | 30 min fixed | retry 16 min |
| 3.1 promote rewire | `wall_clock` 16 min (`budget_clamped: true`) | `true` | 59 s | clamped to outer 4 h remainder | outer budget exhausted |

The other 7 task delegations succeeded in 6–20 min. Retries succeeded only
because the timed-out attempt's edits remained on disk; each retry started
with a fresh context and re-derived work with no information about where the
previous attempt stopped. About 1 h 50 m of the outer 4-hour budget was lost
to re-doing timed-out work and was the primary reason the outer delegation
was cut off mid-phase-3.

Root causes:

1. **Flat budget** — `DEFAULT_TIMEOUT_MILLISECONDS = 1_800_000` applies to
   every per-task delegation regardless of phase size, files touched, or
   observed throughput. The review-dispatch fix #849 derives per-run timeouts
   from work signals; per-task delegations do not.
2. **No progress-aware stall detection** — two of the four were `child_active:
   true` when killed. Wall-clock is the only limit; idle detection
   (`idle_timeout`, `probeSubtreeActivity` from #930/#950) is available in
   `delegation-budget.ts` but not wired as the primary limit for this call
   site.
3. **No resume hint** — prompt-only delegations carry no durable checkpoint to
   re-anchor a retry. The on-disk edits and `tasks.md` checkbox state that
   retried tasks exploited are not returned as structured `resume` data.
4. **Outer budget invisible at phase boundaries** — `/uf.unleash` does not report
   remaining budget at phase checkpoints, so the orchestrator cannot plan a
   resume boundary before the parent budget expires.

This change is scoped to Tier 1 (autonomous execution reliability) under
milestone v0.22.0 (#1060). Docs/blog/tutorial and OpenCode v2 migration remain
out of scope and tracked separately.

## What Changes

Four coordinated fixes in the tool layer (issue #1080 proposed directions,
hardened to preserve #930/#950/#849 invariants):

1. **Derive the per-task budget from the task.** Replace the fixed 30-minute
   default for `/uf.unleash` implementation delegations with a deterministic
   derivation analogous to `run-timeout.ts` (`#849`): inputs are phase size
   (task count in the phase), files touched (from `tasks.md` / `parse_tasks`
   metadata or the phase's `files` list), and model throughput tier
   (`Tier.TermMs` in `run-timeout.ts`). The formula MUST remain bounded by the
   existing `MAX_TIMEOUT_MS = 1_800_000 ms` schema bound unless the delegation
   is explicitly lifted (see 2). The derivation and its clamp outcome are
   recorded as timeout provenance on the run result, matching RD-FR-009.

2. **Idle detection as the primary stall signal for tasks under a command-path
   parent.** Implementation delegations run inside a command-path `/uf.unleash`
   child that already has a 4-hour budget and an `idle_timeout`. Idle polling
   (`probeSubtreeActivity`) is wired so a working child (`child_active: true`
   or running tool parts per `idle-timeout-aborts-children-waiting-on`) is not
   aborted by wall-clock while it is making progress. Wall-clock remains an
   absolute backstop, and hierarchical clamping (`budget_exhausted`,
   non-retryable) from #930/#950 is unchanged. Where the existing `idle_timeout`
   default (600 000 ms) is preserved, tasks inherit it via the parent budget
   registry; no new global idle default is introduced outside that registry.

3. **Lightweight resume hint for prompt-only delegations carrying
   `checkpoint_step`/`phase`.** When a per-task delegation is invoked with
   `checkpoint_step: "implement"` and `phase`, its timeout result includes a
   `resume` object with (a) files modified by the child (from the
   worktree/cherry-pick diff or session file tracking already emitted by the
   executor), and (b) tasks marked complete (`tasks.md` checkbox deltas) since
   the delegation started. The resume hint is advisory — it does not attempt a
   semantic merge; the retry caller decides how to re-anchor. Retries MUST NOT
   be issued from a blank context when `child_active: true` or when a resume
   hint is present.

4. **Outer-budget visibility at phase checkpoints.** `/uf.unleash` reports its
   remaining parent budget (when a budgeted ancestor exists) at each
   implement-phase checkpoint via the existing `checkpoint` tool's `phase`
   payload or its human-readable checkpoint line, so the orchestrator can plan
   a resume boundary instead of being cut off mid-phase. No new checkpoint
   type is introduced; this amends the phase-checkpoint rendering.

## Capabilities

### New Capabilities

- `per-task-delegation-budgets`: work-derived wall-clock for `/uf.unleash`
  per-task `uf_delegate_command` calls, idle-aware stall detection inherited
  from the parent budget, and a lightweight resume hint (`files_modified` +
  `tasks_completed`) for prompt-only delegations that carry phase context.

### Modified Capabilities

- `unleash-orchestration`: `/uf.unleash` implementation phase delegation call
  site — derives its timeout per task, wires idle semantics, and surfaces
  remaining budget at phase checkpoints.
- `delegation-timeouts`: `probeSubtreeActivity` / `idle_timeout` applicability
  extended to cover per-task children of a budgeted `/uf.unleash` parent;
  budget derivation gains a per-task variant sharing `run-timeout.ts`
  throughput tiers.
- `agent-executor`: timeout result envelope gains optional per-task provenance
  fields and the lightweight `resume` hint for the affected call site; the
  `TimeoutSchema` 1_800_000 ms bound is preserved.

### Removed Capabilities

- None. The flat 30-minute default stops being the effective default for the
  `/uf.unleash` per-task call site; prompt-only delegations elsewhere retain
  the existing ceiling and fail-closed semantics.

## Impact

- `.opencode/commands/uf.unleash.md` — per-task delegation call site: derive
  timeout, attach `checkpoint_step`/`phase` resume hint, and emit remaining
  budget at phase checkpoints. Instruction-text delta expected ~+20–40 lines;
  actual delta recorded in the PR.
- `.opencode/lib/run-timeout.ts` — per-task derivation variant (or shared
  helper) reusing `TIER_TERM_MS` / contention logic from #849; pure function,
  no I/O.
- `.opencode/lib/delegation-budget.ts` — probe/idle applicability for
  per-task children; optional `resume` hint assembly from staged-file and
  `tasks.md` deltas already observable by the executor.
- `.opencode/lib/agent-executor.ts` — timeout provenance + lightweight resume
  on the result envelope; hierarchical clamp unchanged.
- `.opencode/plugins/uf-workflow/index.ts` — `uf_delegate_command`
  failure envelope fields for the per-task resume hint; edits stay inside the
  `uf_delegate_command` section as in #930.
- `internal/scaffold/assets/opencode/` and `internal/scaffold/assets/uf/` —
  byte-identical mirrors of every modified canonical asset, plus drift-table
  entries as in #849/#930.
- `.opencode/test/` — unit tests for derivation (formula, clamp, provenance),
  idle inheritance, and resume-hint assembly via injected filesystem fakes.
- `docs/` / website — user-facing reference for derived per-task timeouts
  and the resume hint tracked as `unbound-force/website#` per the
  Cross-Repo Documentation rule; blog/tutorial assessed N/A (internal
  reliability fix, no new operator workflow).
- No change to `invoke_agent` or `dispatch_review_runs` input bounds, to any
  `run-timeout.ts` 1_800_000 ms ceiling, or to any quality gate (coverage
  floor, CRAP, review limits). Gatekeeping note per constitution:
  no gate is weakened, bypassed, or disabled.

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.4.0, 2026-10-14;
six active principles). The `openspec instructions` template still lists four
principles; this assessment covers all six.

### I. Autonomous Collaboration

**Assessment**: PASS

Timeout derivation and the lightweight resume hint are deterministic tool-layer
additions. Heroes continue to collaborate through self-describing artifacts
(run results, checkpoint state, `tasks.md` deltas); no synchronous coupling or
ad-hoc exchange format is introduced, and the resume hint travels inside the
existing timeout-result envelope.

### II. Composability First

**Assessment**: PASS

The derivation helper remains a pure standalone module usable without the
plugin. `/uf.unleash` callers without a budgeted parent behave as before;
tasks outside `/uf.unleash` keep the existing 30-minute ceiling. No new
mandatory dependency is added, and additive value appears only when the
delegation budget registry has an ancestor to clamp against.

### III. Observable Quality

**Assessment**: PASS

Every per-task delegation records effective timeout, derivation inputs (phase
size, files touched, tier, contention multiplier), clamp outcome, idle
verdict, and — on timeout — the lightweight resume hint, all in
machine-parseable JSON with provenance. Remaining-budget reporting at phase
checkpoints makes the 4-hour budget diagnosable per phase without scraping
timestamps.

### IV. Testability

**Assessment**: PASS

The derivation function is pure (no I/O, no clock, no network) and
unit-testable across formula, clamping, and edge cases. Budget clamping and
idle inheritance are testable through the established injectable patterns in
`.opencode/test/` and `internal/scaffold` fakes. No external service is
required to verify any requirement.

### V. Security by Default

**Assessment**: PASS

No secrets, authentication, or injection surface is touched. Existing input
validation (`TimeoutSchema` 1_800_000 ms bound, `validateTimeout`) is
preserved; the optional resume hint carries file paths and task IDs only, not
command output or message content. No new dependencies are introduced.

### VI. Deterministic Execution

**Assessment**: PASS

Per-task budgets are derived by a deterministic pure function, not by an
orchestrating LLM re-reasoning the limit. Idle vs. wall-clock classification
and the resume hint are computed by the tool layer and returned as structured
fields; command prose branches on those fields rather than asking the agent
to inspect session history or reconstruct progress.
