<!-- capability: delegation-timeouts
     change: idle-timeout-aborts-children-waiting-on
     originating_issue: 1117
     rationale: make idle detection tool-aware: running tool parts count as
                activity, idle defers until tool timeout+grace, abort envelope
                carries running_tool discriminator; complements compaction-storm
                fix (DT-FR-002) which excluded synthetic markers
     dependencies: delegation-timeout-budgets (DT-FR-001..004)
-->

Requirement numbering: this change allocates DT-FR-013 through DT-FR-015.
DT-FR-005 through DT-FR-012 remain reserved by prior changes. The sibling
capability `run-plan-command` is unchanged by this change. Requirement ids
MUST stay inside the allocated range.

## ADDED Requirements

### Requirement: Running Tool Counts as Activity

**ID: DT-FR-013** This requirement MUST be satisfied.

The idle probe's progress fingerprint MUST include every tool-call part whose
`state.status` is `"running"` anywhere in the delegated subtree. The
fingerprint MUST mix the set of `(partId, status, tool)` triples for those
parts into the hash — all three fields, with `tool` set to the part's
tool name (empty string when the part carries none) — sorted
lexicographically by `partId` so the hash is deterministic. A change in
the running set — completion, a new part, or a tool swap on the same
`partId` — MUST be observed as progress. Activity in any descendant session
MUST count, using the same depth-first, newest-child-first walk as DT-FR-002.
Synthetic `opencode-compaction-*` markers MUST remain excluded from the
fingerprint (DT-FR-002).

When at least one tool part is `running`, the subtree MUST be considered
active for the purpose of `last_progress_ms_ago` even if the message-hash
component is unchanged — within a DT-FR-014 deferral bound. Growth of
`state.metadata` (e.g., `bash` `output`) SHOULD also count as progress where
observable, but the `running` set alone is sufficient to defer the idle
clock, subject to that bound.

#### Scenario: Long bash with no new messages stays alive

- **GIVEN** a delegation with a 10-minute idle timeout
- **AND** a descendant session whose latest message contains a `bash` part with `state.status: running` for 15 minutes and no new messages are created
- **WHEN** 15 minutes pass
- **THEN** the delegation is not aborted with `kind: idle`
- **AND** `last_progress_ms_ago` does not exceed the running interval

#### Scenario: Running tool set change is progress

- **GIVEN** a subtree with one `bash` part running
- **WHEN** that part completes and a new `bash` part starts (partId changes)
- **THEN** the fingerprint changes and the idle clock resets

#### Scenario: Quiet bash still defers idle

- **GIVEN** a 15-minute `bash` invocation whose `state.metadata.output` does not grow for 10 minutes
- **WHEN** the idle window elapses while `state.status` remains `running`
- **THEN** the delegation is not aborted

### Requirement: Idle Defers to Tool Timeout

**ID: DT-FR-014** This requirement MUST be satisfied.

While any tool part in the subtree has `state.status === "running"`, the idle
watcher MUST NOT abort the delegation with `kind: idle` before the running
part's deferral bound has elapsed. The deferral bound is:

- When the running part exposes an observable tool deadline: that
  deadline's remaining time plus a grace period. The grace period MUST be
  a fixed constant and MUST NOT exceed 60,000 ms.
- When no running part exposes an observable tool deadline: a fixed
  deferral cap `IDLE_DEFERRAL_CAP_MS` measured from the last fingerprint
  change. The cap MUST be between 5 and 15 minutes (300,000-900,000 ms).

If multiple tools are running, the deferral bound MUST cover the earliest
remaining deadline — or the cap when no deadline is observable — because
the first completion changes the fingerprint and resets both clocks. Any
fingerprint change (DT-FR-013) resets the deferral accumulation; with no
fingerprint change, the idle abort MUST NOT be held later than
`idle_timeout` plus the deferral bound.

The tool timeout is model-handled (the model receives the tool error); the
idle timeout is model-stall-handled. Wall-clock (`kind: wall_clock`,
including budget-clamped runs) remains the absolute backstop for tools
that never time out at the tool layer.

Stalling between tool calls — where no part has `status: running` anywhere in
the subtree — MUST still be subject to the idle window (e.g., the `read`
non-existent tool / `dispatch_review_runs` never-returned evidence in #1117).

#### Scenario: True stall still aborts

- **GIVEN** a delegation with a 10-minute idle timeout
- **AND** a descendant model turn that made no tool calls for 10 minutes and no `running` parts exist
- **WHEN** the idle window elapses
- **THEN** the child is aborted with `kind: idle`, retryable, and `running_tool` is null

#### Scenario: Model waiting on nested dispatch

- **GIVEN** a `divisor-testing` subagent whose `dispatch_review_runs` tool never returned and whose probe shows no `running` parts for 10 minutes
- **WHEN** the idle window elapses
- **THEN** the delegation is aborted with `kind: idle`

#### Scenario: Tool timeout is not an idle abort

- **GIVEN** a running `bash` part whose tool timeout fires at 10 minutes
- **WHEN** the tool returns its timeout error
- **THEN** the delegation is not aborted with `kind: idle` at that moment; the model is given the grace period to handle the error

#### Scenario: Stuck running part aborts at the deferral cap

- **GIVEN** a delegation with a 10-minute idle timeout and a running `bash` part that exposes no observable tool deadline and never completes
- **WHEN** the idle window plus `IDLE_DEFERRAL_CAP_MS` elapse since the last fingerprint change
- **THEN** the delegation is aborted with `kind: idle` and `running_tool: "bash"`

### Requirement: Abort Envelope Carries Running Tool

**ID: DT-FR-015** This requirement MUST be satisfied.

When a run ends by any timeout limit, the `timeout` object (DT-FR-004) MUST
carry an additional field `running_tool` (type: `string | null`) for every
`kind`. Its value is the tool name (e.g., `"bash"`) of the `running` part
with the lexicographically smallest `partId` at abort time, or `null` when
no part was running — a deterministic selection rule, reproducible for a
given subtree state (Constitution III).

The tool name originates from model-controlled message parts. Before it is
threaded into any envelope or log line, it MUST be validated against
`^[A-Za-z0-9._:-]{1,128}$`. A value failing validation MUST be reported as
`null` and MUST NOT be logged verbatim (Constitution V: untrusted input is
never echoed). `uf_delegate_command` MUST thread `running_tool` into its
`DELEGATE_FAILED` envelope and, for command-path delegations, into `resume`
where applicable. Existing fields (`kind`, `effective_ms`, `requested_ms`,
`budget_clamped`, `idle_timeout_ms`, `elapsed_ms`, `last_progress_ms_ago`,
`child_active`, `pending_question`) MUST remain unchanged.

#### Scenario: Abort while bash running reports running_tool

- **GIVEN** a subtree where `bash` is `running` at wall-clock expiry
- **WHEN** the delegation ends by timeout
- **THEN** the envelope carries `timeout.running_tool: "bash"`

#### Scenario: Abort with no running tool reports null

- **GIVEN** a true stall with no `running` parts
- **WHEN** the idle timeout aborts the delegation
- **THEN** the envelope carries `timeout.running_tool: null`

#### Scenario: Adversarial tool name is not threaded verbatim

- **GIVEN** a running part whose `tool` value fails the `^[A-Za-z0-9._:-]{1,128}$` validation
- **WHEN** the delegation ends by timeout
- **THEN** the envelope carries `timeout.running_tool: null`
- **AND** the raw `tool` value is not logged verbatim

## MODIFIED Requirements

### Requirement: Idle Timeout

**ID: DT-FR-002** This requirement MUST be satisfied.

Previously: When an idle timeout is configured, the executor MUST poll the
child's session subtree for activity and MUST abort the child with error code
`idle_timeout` (retryable) only after no activity change has been observed for
the idle window and at least one probe has succeeded since the last observed
change. A failed probe MUST NOT reset or advance the idle clock. Activity in
any descendant session MUST count as progress for the child, including growth
of a running tool's streamed `state.metadata` (for example bash output).

New: When an idle timeout is configured, the executor MUST poll the child's
session subtree for activity and MUST abort the child with error code
`idle_timeout` (retryable) only after no activity change has been observed for
the idle window, at least one probe has succeeded since the last observed
change, **and no tool-call part in the subtree has `state.status ===
"running"` within an unexpired DT-FR-014 deferral bound**. A failed probe
MUST NOT reset or advance the idle clock. Activity in any
descendant session MUST count as progress for the child, including **the
presence of any `running` tool part (DT-FR-013)** and growth of a running
tool's streamed `state.metadata`. The probe MUST walk the subtree depth-first,
newest child first. A probe that completes after the child's prompt returned
MUST NOT abort the child or change the run's result.

#### Scenario: Active child survives the idle window

- **GIVEN** a delegation with a 10-minute idle timeout
- **AND** a grandchild session that posts new parts every minute
- **WHEN** 15 minutes pass
- **THEN** the delegation is not aborted

#### Scenario: Stalled child aborted early

- **GIVEN** a delegation with a 4-hour wall clock and a 10-minute idle timeout
- **WHEN** no session in the subtree changes for 10 minutes
- **THEN** the child is aborted with `idle_timeout`

#### Scenario: Long command still printing

- **GIVEN** a delegation with a 10-minute idle timeout
- **AND** a child running a 15-minute bash command whose `state.metadata.output` grows every minute
- **WHEN** 15 minutes pass
- **THEN** the delegation is not aborted

#### Scenario: Probe unavailable

- **GIVEN** a host whose activity probe always fails
- **WHEN** the idle window elapses
- **THEN** the child is not aborted by the idle clock

### Requirement: Resume-Carrying Timeout Results

**ID: DT-FR-004** This requirement MUST be satisfied.

Previously: When a run ends by a limit, the executor result MUST carry a
`timeout` object with `kind` (`wall_clock`, `idle`, `budget`), `effective_ms`,
`requested_ms`, `budget_clamped`, `idle_timeout_ms`, `elapsed_ms`,
`last_progress_ms_ago`, `child_active`, and `pending_question`.

New: When a run ends by a limit, the executor result MUST carry a `timeout`
object with `kind` (`wall_clock`, `idle`, `budget`, `storm`), `effective_ms`,
`requested_ms`, `budget_clamped`, `idle_timeout_ms`, `elapsed_ms`,
`last_progress_ms_ago`, `child_active`, `pending_question`, **and
`running_tool` (string | null, DT-FR-015)**. `uf_delegate_command` MUST thread
that object into its `DELEGATE_FAILED` envelope and, for command-path
delegations, MUST add `resume` with the state `command`, `key`,
`last_checkpoint_step`, and `completed_steps` read from the progress state store
after the child stops; an unreadable state MUST yield `resume: null`.

#### Scenario: Idle timeout of a checkpointed command

- **GIVEN** a `/uf.unleash` command-path delegation whose child recorded `plan` as its current step
- **WHEN** the child is aborted by the idle timeout
- **THEN** the envelope carries `timeout.kind: "idle"`
- **AND** `resume.last_checkpoint_step: "plan"`
- **AND** `child_session_id`

#### Scenario: Question-gate wedge

- **GIVEN** a delegated child whose latest part is a running `question` tool call
- **WHEN** the idle timeout fires
- **THEN** the envelope carries `timeout.pending_question: true`

## REMOVED Requirements

None.
