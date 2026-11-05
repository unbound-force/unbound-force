<!-- capability: run-plan-command
     change: delegation-timeout-budgets
     originating_issue: 930
     rationale: adds the delegation timeout recovery protocol to
     the run-plan orchestrator; the capability's founding
     unnumbered requirements live in the unarchived founding
     change openspec/changes/run-plan-command/ and are not
     modified here (the capability is not yet in
     openspec/specs/, so this delta uses ADDED)
-->

Requirement numbering: this change allocates RPC-FR-001 for the
requirement added below. RPC-FR-002 through RPC-FR-010 are
reserved for future run-plan-command changes. The capability's
founding requirements are unnumbered and live in the unarchived
founding change `openspec/changes/run-plan-command/`; the sibling
capability `delegation-timeouts` (delta in this same change)
allocates DT-FR-001 through DT-FR-004 and reserves DT-FR-005
through DT-FR-012. Requirement ids MUST stay inside the
allocated range.

## ADDED Requirements

### Requirement: Delegation Timeout Recovery

**ID: RPC-FR-001** This requirement MUST be satisfied.

The orchestrator MUST delegate `/uf.unleash` with
`timeout: 14400000` and MUST NOT pass `idle_timeout` (the
`uf_delegate_command` default applies). `/uf.finale` keeps
`timeout: 600000` and MUST set `idle_timeout: 300000` so its
commit-gate wedge ends as an idle timeout with
`pending_question: true` before the wall clock.

On a `DELEGATE_FAILED` envelope carrying a `timeout` object, the
orchestrator MUST branch on its fields:

- `pending_question: true`: gate wedge; apply gate-timeout
  recovery.
- `kind: "idle"` without a pending question: re-delegate the same
  command once as a resume (same `command_path`, no routing
  context, the prompt carrying the description, issue
  reference(s), and `resume.last_checkpoint_step`); when that step
  is null the re-run is effectively a fresh start; a second idle
  timeout on the same item MUST route through the blocked-item
  protocol.
- `kind: "wall_clock"`: route through the blocked-item protocol,
  relaying `child_session_id`, `child_active`, and
  `resume.last_checkpoint_step`.
- `kind: "budget"`: treated as `wall_clock` (not expected at the
  orchestrator, which has no budgeted parent).

When `resume.last_checkpoint_step` is non-null, the orchestrator
MUST NOT restart the item fresh on `main`.

#### Scenario: Provider stall resumes

- **GIVEN** a `/uf.unleash` delegation that fails with
  `timeout.kind: "idle"`, `pending_question: false`, and
  `resume.last_checkpoint_step: "tasks"`
- **WHEN** the orchestrator recovers
- **THEN** it re-delegates `/uf.unleash` once on the feature
  branch with the resume step in the prompt

#### Scenario: Repeated stall blocks the item

- **GIVEN** an item whose resume re-delegation also fails with
  `timeout.kind: "idle"`
- **WHEN** the orchestrator recovers
- **THEN** it marks the item blocked and asks the pause-relay
  question instead of re-delegating

#### Scenario: Gate wedge

- **GIVEN** a delegation that fails with
  `timeout.pending_question: true`
- **WHEN** the orchestrator recovers
- **THEN** it applies gate-timeout recovery and does not
  re-delegate before the gate is cleared

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
