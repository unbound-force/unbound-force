## ADDED Requirements

### Requirement: Label Creation Confirmation Gate

The `/triage-issue` command MUST use the AskUserQuestion tool
to obtain user confirmation before executing `gh label create`
for any label. The prompt MUST include the label name and
description. Options MUST be
`["Yes -- create label", "No -- skip"]`.

If the user selects "No -- skip", the command MUST skip label
creation, record the skip in `actions_taken`, and continue
with remaining triage actions.

#### Scenario: New label requires creation

- **GIVEN** the triage classification maps to a label that
  does not exist in the repository
- **WHEN** the command reaches the label existence check
- **THEN** the command MUST prompt the user with
  AskUserQuestion before executing `gh label create`

#### Scenario: User declines label creation

- **GIVEN** a label does not exist and the user is prompted
- **WHEN** the user selects "No -- skip"
- **THEN** the command MUST NOT execute `gh label create`,
  MUST record the skip in `actions_taken`, and MUST skip
  label application (since the label does not exist)

### Requirement: Label Application Confirmation Gate

The `/triage-issue` command MUST use the AskUserQuestion tool
to obtain user confirmation before executing
`gh issue edit --add-label` for any label. The prompt MUST
include the label name and issue number. Options MUST be
`["Yes -- apply label", "No -- skip"]`.

If the user selects "No -- skip", the command MUST skip label
application, record the skip in `actions_taken`, and continue
with remaining triage actions.

#### Scenario: Label exists and is ready to apply

- **GIVEN** the triage classification maps to a label that
  exists in the repository (or was just created)
- **WHEN** the command reaches the label application step
- **THEN** the command MUST prompt the user with
  AskUserQuestion before executing `gh issue edit --add-label`

#### Scenario: User declines label application

- **GIVEN** a label exists and the user is prompted
- **WHEN** the user selects "No -- skip"
- **THEN** the command MUST NOT execute
  `gh issue edit --add-label`, MUST record the skip in
  `actions_taken`, and MUST continue with comment posting

#### Scenario: Label already applied (re-run)

- **GIVEN** the target label is already applied to the issue
  (detected in Phase 1.2)
- **WHEN** the command reaches label application
- **THEN** the command MUST skip both the AskUserQuestion
  prompt and the `gh issue edit --add-label` call, and note
  "label already present"

## MODIFIED Requirements

### Requirement: Label Application Policy

The `/triage-issue` command MUST require AskUserQuestion
confirmation before ALL label mutations, not only the
`duplicate` label.

Previously: "Labels are applied automatically without user
confirmation, with one exception: the `duplicate` label
requires user confirmation because it carries implicit
'close' semantics."

### Requirement: Duplicate Label Additional Context

The duplicate-specific AskUserQuestion messaging (informing
the user that the label signals the issue should be closed)
MUST be preserved as additional context displayed before or
alongside the general label application gate.

Previously: The duplicate gate was the only confirmation gate.
Now it provides supplementary context on top of the general
gate that applies to all labels.

## REMOVED Requirements

None.
