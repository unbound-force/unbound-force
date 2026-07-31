## ADDED Requirements

### Requirement: Branch-safety pre-condition

The `openspec-apply-change` skill MUST include a pre-condition
block immediately after the "Steps" heading and before Step 1
that states:

> **Pre-condition**: Before any step, verify: NEVER switch
> branches or suggest archiving with uncommitted changes. Run
> `git status --short` if branch state is uncertain.

This pre-condition MUST be encountered by agents before any
workflow step executes.

#### Scenario: Agent processes skill file sequentially

- **GIVEN** an agent loads `openspec-apply-change/SKILL.md`
- **WHEN** the agent begins processing the "Steps" section
- **THEN** the agent encounters the branch-safety pre-condition
  BEFORE reading Step 1 ("Select the change")

#### Scenario: Agent has uncommitted changes and reaches
completion

- **GIVEN** an agent is executing the apply-change workflow
- **AND** there are uncommitted changes in the working tree
- **WHEN** the agent reaches Step 8 (completion)
- **THEN** the agent MUST NOT suggest switching branches or
  archiving because the pre-condition was processed before
  Step 1

## MODIFIED Requirements

### Requirement: Guardrails section branch-safety rule

The existing guardrail "NEVER switch branches or suggest
archiving with uncommitted changes" at the end of the file
SHALL remain unchanged.

Previously: This was the only location of the branch-safety
constraint. Now it serves as a secondary reinforcement of the
pre-condition.

## REMOVED Requirements

None.
