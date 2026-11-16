<!-- capability: agent-executor
     change: per-task-delegation-budgets
     originating_issue: 1080 -->

## ADDED Requirements

### Requirement: Per-Task Timeout Provenance and Resume Envelope

**ID: AE-FR-010** This requirement MUST be satisfied.

The `agent-executor.ts` timeout result envelope gains optional per-task fields
without changing the existing `MAX_TIMEOUT_MS = 1_800_000 ms`
`TimeoutSchema` bound or `validateTimeout` semantics. When the caller is a
per-task `uf_delegate_command` (`checkpoint_step: "implement"` + `phase`),
the envelope MUST carry (a) derivation provenance per PTB-FR-001 and (b) an
advisory `resume` hint per PTB-FR-003 (`files_modified`, `tasks_completed`,
or `null` when unreadable). Other callers are unaffected. Hierarchical
`budget_exhausted` clamping is unchanged per PTB-FR-002.

#### Scenario: Per-task timeout carries provenance

- **GIVEN** a per-task delegation with no explicit timeout
- **WHEN** the executor produces a result (success or timeout)
- **THEN** the result's timeout provenance fields are present per PTB-FR-001

#### Scenario: Per-task timeout carries resume on failure

- **GIVEN** a per-task delegation with `checkpoint_step: "implement"` that
  times out
- **WHEN** the executor returns the failure
- **THEN** the optional `resume` field is present per PTB-FR-003

#### Scenario: Timeout bound preserved

- **GIVEN** any per-task derivation
- **WHEN** the timeout is validated
- **THEN** the existing `1_800_000 ms` ceiling is enforced
- **AND** no gate is weakened
