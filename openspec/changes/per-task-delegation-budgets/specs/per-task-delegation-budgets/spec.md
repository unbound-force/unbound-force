<!-- capability: per-task-delegation-budgets
     change: per-task-delegation-budgets
     originating_issue: 1080 -->

## ADDED Requirements

### Requirement: Per-Task Timeout Derivation

**ID: PTB-FR-001** This requirement MUST be satisfied.

`/uf.unleash` implementation-phase per-task delegations (`uf_delegate_command`
with `checkpoint_step: "implement"` and `phase`) MUST NOT use the flat
`1_800_000 ms` default as their effective wall clock when no explicit
`timeout` argument is provided. The tool layer MUST derive the wall-clock
budget deterministically from the task's work signals via a pure function with
no I/O, no clock, and no LLM judgment; identical inputs MUST produce an
identical timeout.

The derivation MUST compose, in order:

1. a base window resolved per `run-timeout.ts` throughput tiers
   (`Tier.TermMs` — lightweight / standard / heavy; zero when no tier signal
   is available), bounded below by the resolved base;
2. a phase-size term monotonically non-decreasing in the task count of the
   enclosing `tasks.md` phase (from `parse_tasks` phase header);
3. a files-touched term monotonically non-decreasing in the file count the
   phase declares (phase `files` list or `tasks.md` metadata), carried by the
   prompt-size analogue from RD-FR-028 when a prompt payload is present;
4. a contention multiplier strictly greater than 1, applied only when the
   implementation phase fans out more than one concurrent per-task delegation
   (configured `max_parallel_runs` exceeds 1), capped at `1.5`.

The derived value MUST be clamped to `min(configuredCeilingMs, 1_800_000)` per
the existing `TimeoutSchema` / `validateTimeout` bound. When the configured
ceiling is below the resolved base window, the base floor MUST take
precedence: the effective ceiling is the base window. An explicit `timeout`
argument MUST override derivation entirely. Derivation inputs, the unclamped
derived value, the ceiling, and the clamp outcome MUST be recorded as timeout
provenance on the executor result per RD-FR-009 / RD-FR-030 (source, base,
phase size, file count, prompt bytes when present, tier, contention
multiplier, `max_parallel_runs`, `effective_ms`, `clamped`).

#### Scenario: Flat default no longer applies for per-task delegations

- **GIVEN** an `/uf.unleash` implementation delegation for a phase task with
  no explicit `timeout`
- **WHEN** the per-task timeout is resolved
- **THEN** the effective wall clock is the derived value from work signals,
  not the `1_800_000 ms` flat default
- **AND** the derivation provenance is present on the result

#### Scenario: Larger phases scale the window

- **GIVEN** two identical per-task delegations except that one belongs to a
  phase with materially more tasks or files
- **WHEN** per-task timeouts are derived for both
- **THEN** the larger phase's effective timeout is greater than or equal to
  the smaller phase's

#### Scenario: Parallel contention multiplies the window

- **GIVEN** an implement phase whose `max_parallel_runs` exceeds 1
- **WHEN** the per-task timeout is derived
- **THEN** the effective timeout is greater than or equal to the same
  derivation with `max_parallel_runs` of 1
- **AND** the unclamped derivation is strictly greater than the
  single-run case

#### Scenario: Derived value clamped at the ceiling

- **GIVEN** work signals large enough that the unclamped derivation exceeds
  the configurable ceiling
- **WHEN** the per-task timeout is derived
- **THEN** the effective timeout equals the ceiling
- **AND** the ceiling does not exceed `1_800_000 ms`
- **AND** provenance records `clamped: true`

#### Scenario: Explicit timeout bypasses derivation

- **GIVEN** any combination of work signals
- **WHEN** the caller provides an explicit `timeout` argument
- **THEN** the explicit value is the effective timeout
- **AND** no derivation term modifies it
- **AND** provenance source is `explicit`

#### Scenario: Determinism

- **GIVEN** the same base window, phase size, file count, tier, concurrency,
  and ceiling
- **WHEN** the derivation runs twice
- **THEN** both invocations produce the identical effective timeout

### Requirement: Idle Detection as Primary Stall Signal under Command-Path Parent

**ID: PTB-FR-002** This requirement MUST be satisfied.

When a per-task `uf_delegate_command` child runs inside a budgeted
command-path `/uf.unleash` parent, its wall clock MUST be hierarchically
clamped to the parent deadline via the existing `delegation-budget.ts` registry
(`globalThis` symbol-keyed registry, `resolveInheritedDeadline` with at most 8
parent hops, `clampToBudget` with 60 s reserve and 30 s minimum, fail-fast
`budget_exhausted` non-retryable when no budget remains — per #930/#950).

When that parent carries an `idle_timeout` (default `600_000 ms` preserved from
#930), the idle watchdog MUST be inherited: the executor MUST poll
`probeSubtreeActivity` (DFS over `session.children`, newest child first, at most
32 sessions, latest message per session) on the injectable poll interval. A
child that is progressing MUST NOT be aborted by wall-clock while progressing:
progress is defined as `child_active: true` (fingerprint change within the
last 2 minutes) or a running tool part per
`idle-timeout-aborts-children-waiting-on` (a delegated child or an `execute`
part mid-solve/merge). After `idle_timeout` of no observable subtree change
with at least one confirming successful probe, the executor MUST abort with
`idle_timeout` (retryable). Failed probes MUST neither reset nor advance the
idle clock. Wall-clock remains an absolute backstop; `budget_exhausted` clamping
is unchanged.

No new global idle default is introduced outside the registry; top-level
callers without a budgeted ancestor behave as before.

#### Scenario: Active child survives wall clock

- **GIVEN** a per-task child with `child_active: true` and 29 s since last
  progress under a command-path parent with idle
- **WHEN** its derived wall clock expires before the child finishes
- **THEN** the executor does not abort while the child remains active
- **AND** abort occurs only if idle is later detected or wall-clock remains
  after activity stops

#### Scenario: Stalled child aborted by idle

- **GIVEN** a per-task child with no fingerprint change and no running tool
  for longer than `idle_timeout`, with one confirming successful probe
- **WHEN** the poll evaluates
- **THEN** the executor aborts with `idle_timeout`
- **AND** the result carries `resume` per PTB-FR-003

#### Scenario: Running tool counts as progress

- **GIVEN** a per-task child where the subtree contains a delegated child or
  `execute` part mid-solve
- **WHEN** the idle poll evaluates with no message progress
- **THEN** the poll reports `child_active: true`
- **AND** the idle clock does not advance

#### Scenario: Budget exhausted remains non-retryable

- **GIVEN** a per-task child whose clamped wall clock exhausts the outer
  4-hour parent budget
- **WHEN** the executor resolves the inherited deadline
- **THEN** the failure is `budget_exhausted`
- **AND** the command prose MUST NOT retry it

### Requirement: Lightweight Resume Hint for Prompt-Only Delegations

**ID: PTB-FR-003** This requirement MUST be satisfied.

When a per-task `uf_delegate_command` is invoked with
`checkpoint_step: "implement"` and `phase` and terminates due to timeout
(`wall_clock`, `idle_timeout`, or `budget_exhausted`), the failure envelope
MUST include an optional advisory `resume` object alongside the existing
`timeout` envelope. The hint MUST contain:

- `files_modified`: repository-relative file paths touched by the child (from
  the worktree cherry-pick diff or the executor's already-observable session
  file tracking; no new filesystem watcher is introduced);
- `tasks_completed`: task identifiers marked complete since delegation start
  (`tasks.md` checkbox deltas via `parse_tasks` before/after or the staged
  diff already emitted by the executor).

When state is unreadable, the executor MUST return `resume: null` rather than
synthesizing content (fail-closed, per D4 from #950). The hint is advisory: it
does not perform a semantic merge; the retry caller decides how to re-anchor.

Command prose (`/uf.unleash`) MUST NOT issue a retry from a blank context
when `child_active: true` at timeout or when a resume hint is present. It
MUST relay `files_modified` and `tasks_completed` into the retry prompt.

Prompt-only delegations without `checkpoint_step`/`phase` are unaffected;
other callers ignore the optional field.

#### Scenario: Timeout with progress returns files and tasks

- **GIVEN** a per-task delegation that touched `internal/foo/bar.go` and
  marked tasks `1.1` and `1.2` complete before timing out with
  `child_active: true`
- **WHEN** the executor returns the timeout failure
- **THEN** `resume.files_modified` includes `internal/foo/bar.go`
- **AND** `resume.tasks_completed` includes `1.1` and `1.2`

#### Scenario: Timeout with no progress returns empty hint

- **GIVEN** a per-task delegation that timed out with `child_active: false`
  and produced no file or task deltas
- **WHEN** the executor returns the timeout failure
- **THEN** `resume` is present with empty `files_modified` and
  `tasks_completed` arrays
- **AND** no message content is leaked

#### Scenario: Unreadable state yields null resume

- **GIVEN** a per-task delegation whose file or `tasks.md` state cannot be
  read at abort time
- **WHEN** the executor assembles the resume hint
- **THEN** `resume` is `null`
- **AND** the timeout failure still carries the `timeout` envelope

#### Scenario: Retries must not start blank when hint is present

- **GIVEN** a per-task timeout whose result carries `child_active: true` or a
  non-null `resume` hint
- **WHEN** `/uf.unleash` handles the failure
- **THEN** it MUST NOT retry the task from a blank prompt without the hint
- **AND** the retry prompt includes the hint's files and tasks

### Requirement: Outer-Budget Visibility at Phase Checkpoints

**ID: PTB-FR-004** This requirement MUST be satisfied.

`/uf.unleash` MUST report its remaining parent budget at each implement-phase
checkpoint when a budgeted ancestor exists. The existing `checkpoint` tool's
`phase` payload and its human-readable checkpoint line are amended: the machine
payload MUST include `remaining_ms` (milliseconds until the resolved parent
deadline) and the human line MUST include a `remaining budget: ~Xms` clause.
No new checkpoint type or state file is introduced; when no budgeted ancestor
exists (top-level invocation or unbudgeted parent), no budget line is emitted.

The remaining-budget value MUST be derived from `resolveInheritedDeadline` on
the current session; probes do not affect it. Wall-clock and hierarchical
clamping remain authoritative — the budget line is informational for the
orchestrator to plan a resume boundary before being cut off mid-phase.

#### Scenario: Budgeted parent reports remaining at checkpoint

- **GIVEN** an `/uf.unleash` command-path parent with a 4-hour deadline and
  a per-task child that just completed a phase-1 checkpoint
- **WHEN** the implementation phase emits the checkpoint
- **THEN** the checkpoint `phase` payload includes `remaining_ms`
- **AND** the human-readable line includes remaining budget

#### Scenario: Unbudgeted invocation omits budget line

- **GIVEN** an `/uf.unleash` invocation without a budgeted ancestor
- **WHEN** an implement-phase checkpoint is emitted
- **THEN** no `remaining_ms` or remaining-budget prose is present

#### Scenario: Reporting does not change execution

- **GIVEN** any checkpoint with or without a budget line
- **WHEN** the checkpoint is recorded
- **THEN** task scheduling and budget enforcement behave identically with or
  without the reporting
