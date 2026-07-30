## ADDED Requirements

### Requirement: Pre-condition block for branch safety

The `openspec-apply-change/SKILL.md` file MUST include a
Pre-condition block immediately after the "Steps" heading and
before Step 1 that states:

> **Pre-condition**: Before any step, verify: NEVER switch
> branches or suggest archiving with uncommitted changes. Run
> `git status --short` if branch state is uncertain.

This pre-condition MUST appear before any numbered workflow step
so that agents encounter it before executing the workflow.

#### Scenario: Agent encounters pre-condition before workflow

- **GIVEN** an agent loads `openspec-apply-change/SKILL.md`
- **WHEN** the agent parses the "Steps" section
- **THEN** the pre-condition block MUST appear before Step 1

#### Scenario: Pre-condition instructs branch state verification

- **GIVEN** an agent is about to execute the apply-change workflow
- **WHEN** branch state is uncertain
- **THEN** the agent MUST run `git status --short` before proceeding

## MODIFIED Requirements

### Requirement: Guardrail bullet retention

The existing guardrail bullet at the end of the file ("NEVER switch
branches or suggest archiving with uncommitted changes") SHOULD be
retained for completeness of the Guardrails checklist.

Previously: The guardrail bullet was the sole location of this rule,
placed after the full workflow.

## REMOVED Requirements

None.
