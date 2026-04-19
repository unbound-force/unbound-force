## ADDED Requirements

### Requirement: Canonical Command Ordering

All implementation handoff prompts MUST list commands
in priority order: `/unleash`, `/cobalt-crush`,
`/opsx-apply` (or `/speckit.implement` where
applicable).

#### Scenario: Agent recommends implementation after proposal

- **GIVEN** a spec-phase command completes artifact
  creation
- **WHEN** the agent presents the "next steps" prompt
- **THEN** `/unleash` MUST be listed first, with a note
  that it is recommended for multi-task changes

#### Scenario: Cobalt-Crush fallback with no context

- **GIVEN** `/cobalt-crush` is invoked with no active
  workflow detected
- **WHEN** the agent presents the fallback menu
- **THEN** `/unleash` MUST appear as the first option
  before `/speckit.implement` and `/opsx-apply`

### Requirement: Speckit Workflow Entry Point

The `speckit-workflow` skill MUST include an Entry Point
section that identifies `/unleash` as the primary
command for triggering autonomous pipeline execution.

#### Scenario: Swarm coordinator reads speckit-workflow skill

- **GIVEN** a Swarm coordinator loads the
  `speckit-workflow` skill
- **WHEN** the coordinator looks for how to start
  execution
- **THEN** the skill MUST reference `/unleash` as the
  entry point command

## MODIFIED Requirements

### Requirement: Guardrails NEVER-Run Lists

All spec-phase command guardrails that list commands
agents MUST NOT run now include `/unleash` alongside
`/opsx-apply` and `/cobalt-crush`.

Previously: Guardrails listed only `/opsx-apply` and
`/cobalt-crush` in "NEVER run" lists, omitting
`/unleash`.

#### Scenario: Agent reads guardrails during opsx-propose

- **GIVEN** an agent is executing `/opsx-propose`
- **WHEN** it reads the Guardrails section
- **THEN** the NEVER-run list MUST include `/unleash`,
  `/opsx-apply`, and `/cobalt-crush`

### Requirement: uf-init Template Consistency

The guardrails template injected by `/uf-init` into
`opsx-propose.md` MUST match the corrected text with
`/unleash` included.

Previously: The template in `uf-init.md` Step 8 listed
only `/opsx-apply` and `/cobalt-crush`.

#### Scenario: Fresh uf-init injects guardrails

- **GIVEN** a user runs `/uf-init` on a repo
- **WHEN** Step 8 injects guardrails into
  `opsx-propose.md`
- **THEN** the injected text MUST list `/unleash` in
  both the NEVER-run list and the user prompt

## REMOVED Requirements

None.
