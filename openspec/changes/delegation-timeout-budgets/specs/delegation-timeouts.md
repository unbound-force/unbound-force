## ADDED Requirements

### Requirement: Hierarchical Delegation Budgets

The shared agent executor MUST record each child session's
absolute deadline in a process-wide budget registry for the
duration of the run and MUST remove it when the run ends.

Before creating a child session, the executor MUST resolve the
inherited deadline of its parent session: a registry entry for the
parent session, else the first registered ancestor found by
walking session `parentID` links up to 8 levels. Lookup failures
MUST be treated as "no inherited deadline".

With an inherited deadline, the effective wall-clock timeout MUST
be the lesser of the requested timeout and the remaining budget
minus a 60-second reserve. When less than 30 seconds remain after
the reserve, the executor MUST fail without creating a child
session, with error code `budget_exhausted`, non-retryable.

#### Scenario: Nested call clamped to parent budget

- **GIVEN** a child session registered with 20 minutes remaining
- **WHEN** a tool inside that child requests a 30-minute run
- **THEN** the nested run's effective timeout is 19 minutes
- **AND** its timeout details report `budget_clamped: true`

#### Scenario: Parent budget exhausted

- **GIVEN** a child session registered with 70 seconds remaining
- **WHEN** a tool inside that child starts a nested run
- **THEN** no child session is created
- **AND** the result fails with `budget_exhausted`, not retryable

#### Scenario: Top-level call unchanged

- **GIVEN** a session with no registered ancestor
- **WHEN** a run is requested with a 10-minute timeout
- **THEN** the effective timeout is 10 minutes

### Requirement: Idle Timeout

When an idle timeout is configured, the executor MUST poll the
child's session subtree for activity and MUST abort the child with
error code `idle_timeout` (retryable) only after no activity change
has been observed for the idle window and at least one probe has
succeeded since the last observed change. A failed probe MUST NOT
reset or advance the idle clock. Activity in any descendant session
MUST count as progress for the child.

`uf_delegate_command` MUST accept `idle_timeout` (60,000 -
1,800,000 ms) and MUST default it to 600,000 ms.

#### Scenario: Active child survives the idle window

- **GIVEN** a delegation with a 10-minute idle timeout
- **AND** a grandchild session that posts new parts every minute
- **WHEN** 15 minutes pass
- **THEN** the delegation is not aborted

#### Scenario: Stalled child aborted early

- **GIVEN** a delegation with a 4-hour wall clock and a 10-minute
  idle timeout
- **WHEN** no session in the subtree changes for 10 minutes
- **THEN** the child is aborted with `idle_timeout`

#### Scenario: Probe unavailable

- **GIVEN** a host whose activity probe always fails
- **WHEN** the idle window elapses
- **THEN** the child is not aborted by the idle clock

### Requirement: Command-Path Timeout Ceiling

`uf_delegate_command` MUST accept `timeout` up to 14,400,000 ms
when `command_path` is supplied and MUST reject a `timeout` above
1,800,000 ms when it is not.

#### Scenario: Prompt-only delegation over the cap

- **GIVEN** a prompt-only delegation
- **WHEN** it requests `timeout: 3600000`
- **THEN** validation fails before any child session is created

### Requirement: Resume-Carrying Timeout Results

When a run ends by a limit, the executor result MUST carry a
`timeout` object with `kind` (`wall_clock`, `idle`, `budget`),
`effective_ms`, `requested_ms`, `budget_clamped`,
`idle_timeout_ms`, `elapsed_ms`, `last_progress_ms_ago`,
`child_active`, and `pending_question`.

`uf_delegate_command` MUST thread that object into its
`DELEGATE_FAILED` envelope and, for command-path delegations, MUST
add `resume` with the state `command`, `key`,
`last_checkpoint_step`, and `completed_steps` read from the
progress state store after the child stops; an unreadable state
MUST yield `resume: null`.

#### Scenario: Idle timeout of a checkpointed command

- **GIVEN** a `/uf.unleash` command-path delegation whose child
  recorded `plan` as its current step
- **WHEN** the child is aborted by the idle timeout
- **THEN** the envelope carries `timeout.kind: "idle"`
- **AND** `resume.last_checkpoint_step: "plan"`
- **AND** `child_session_id`

#### Scenario: Question-gate wedge

- **GIVEN** a delegated child whose latest part is a running
  `question` tool call
- **WHEN** the idle timeout fires
- **THEN** the envelope carries `timeout.pending_question: true`
