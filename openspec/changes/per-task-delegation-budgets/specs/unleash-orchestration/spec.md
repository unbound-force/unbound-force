<!-- capability: unleash-orchestration
     change: per-task-delegation-budgets
     originating_issue: 1080 -->

## ADDED Requirements

### Requirement: Unleash Implementation Delegation Site

**ID: UO-FR-010** This requirement MUST be satisfied.

`/uf.unleash` implementation-phase per-task delegations MUST derive their
`timeout` via PTB-FR-001 rather than the flat default, MUST enable
idle-aware stall detection via PTB-FR-002 under a budgeted parent, MUST thread
the advisory `resume` hint per PTB-FR-003 and must not retry from blank when
`child_active: true` or a hint is present is, and MUST report remaining parent
budget at each implement-phase checkpoint per PTB-FR-004. No new checkpoint
type is introduced; the change stays inside the existing
`uf_delegate_command` call site and the `checkpoint` `phase` payload.

#### Scenario: Per-task delegation derives timeout

- **GIVEN** an implement phase with a task count and file list
- **WHEN** `/uf.unleash` fans out the per-task `uf_delegate_command` calls
- **THEN** each call's timeout is the work-derived value per PTB-FR-001

#### Scenario: Per-task delegation relays resume on retry

- **GIVEN** a per-task timeout failure carrying `resume.files_modified` and
  `resume.tasks_completed`
- **WHEN** `/uf.unleash` retries the task
- **THEN** the retry prompt includes the files and tasks from the hint
- **AND** the retry is not issued from a blank context

#### Scenario: Phase checkpoint reports remaining budget

- **GIVEN** a budgeted `/uf.unleash` parent
- **WHEN** an implement-phase checkpoint is emitted
- **THEN** the checkpoint includes remaining budget per PTB-FR-004
