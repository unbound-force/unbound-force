## ADDED Requirements

### Requirement: Label creation confirmation gate

Before executing `gh label create`, the command MUST use the
**AskUserQuestion tool** with options
`["Yes -- create label", "No -- skip"]` to obtain user
confirmation. The prompt MUST include the label name and
description being created.

If the user selects "No -- skip", the command MUST skip both
label creation and label application, and MUST continue to
Phase 4.3.

#### Scenario: Label does not exist and user confirms creation

- **GIVEN** the triage classification resolves to category `bug`
- **AND** the `bug` label does not exist in the repository
- **WHEN** the command reaches label creation
- **THEN** the command MUST prompt with `AskUserQuestion`
  `["Yes -- create label", "No -- skip"]`
- **AND** if the user selects "Yes -- create label", the
  command MUST execute `gh label create "bug" ...`

#### Scenario: Label does not exist and user declines creation

- **GIVEN** the triage classification resolves to category `bug`
- **AND** the `bug` label does not exist in the repository
- **WHEN** the user selects "No -- skip" at the creation prompt
- **THEN** the command MUST NOT execute `gh label create`
- **AND** the command MUST NOT execute `gh issue edit --add-label`
- **AND** the command MUST record the skip in `actions_taken`
  (`labels_applied` MUST NOT include the skipped label)
- **AND** the command MUST continue to Phase 4.3

#### Scenario: Label creation fails after user confirms

- **GIVEN** the triage classification resolves to category `bug`
- **AND** the `bug` label does not exist in the repository
- **AND** the user selects "Yes -- create label"
- **WHEN** `gh label create` fails (e.g., insufficient permissions)
- **THEN** the command MUST record the failure in `actions_taken`
  (`label_creation_failed: true`)
- **AND** the command MUST skip the label application gate
- **AND** the command MUST continue to Phase 4.3

#### Scenario: Label already exists

- **GIVEN** the triage classification resolves to category `bug`
- **AND** the `bug` label already exists in the repository
- **WHEN** the command reaches label creation
- **THEN** the command MUST skip the creation gate entirely
- **AND** the command MUST proceed to the label application gate

### Requirement: Label application confirmation gate

Before executing `gh issue edit --add-label`, the command MUST
use the **AskUserQuestion tool** with options
`["Yes -- apply label", "No -- skip"]` to obtain user
confirmation. The prompt MUST include the label name being
applied.

If the user selects "No -- skip", the command MUST skip label
application and MUST continue to Phase 4.3.

#### Scenario: User confirms label application

- **GIVEN** the label exists in the repository (either
  pre-existing or just created)
- **AND** the label is not already applied to the issue
- **WHEN** the command reaches label application
- **THEN** the command MUST prompt with `AskUserQuestion`
  `["Yes -- apply label", "No -- skip"]`
- **AND** if the user selects "Yes -- apply label", the
  command MUST execute `gh issue edit <ISSUE_NUMBER> --add-label`

#### Scenario: User declines label application

- **GIVEN** the label exists in the repository
- **AND** the label is not already applied to the issue
- **WHEN** the user selects "No -- skip" at the application
  prompt
- **THEN** the command MUST NOT execute
  `gh issue edit --add-label`
- **AND** the command MUST record the skip in `actions_taken`
  (`labels_applied` MUST NOT include the skipped label)
- **AND** the command MUST continue to Phase 4.3

#### Scenario: Label already applied to issue

- **GIVEN** the target label is already applied to the issue
  (detected by the re-run check in Phase 4.2)
- **WHEN** the command reaches label handling
- **THEN** the command MUST skip the creation gate (label
  necessarily exists)
- **AND** the command MUST skip the application gate
- **AND** the command MUST note "label already present"

## MODIFIED Requirements

### Requirement: Label application documentation (line 243)

The command MUST state that all labels require user confirmation
before application.

Previously: "Labels are applied automatically without user
confirmation, with one exception: the `duplicate` label requires
user confirmation because it carries implicit 'close' semantics."

New text: "All labels require user confirmation before
application. The `duplicate` label includes an additional warning
about implicit close semantics."

### Requirement: Duplicate label gate (lines 273-276)

The existing separate `duplicate` label gate MUST be merged into
the general label application gate. The close-semantics warning
from the existing gate MUST be preserved in the unified prompt.

Previously: The duplicate gate was the only confirmation gate in
Phase 4.2 and existed as a separate block (lines 273-276).

New behavior: The separate duplicate-specific gate block is
removed. For `duplicate` labels, the general application gate
uses specialized prompt text that includes the close-semantics
warning. This avoids double-prompting while preserving the
warning.

#### Scenario: Duplicate label application

- **GIVEN** the triage classification resolves to `duplicate`
- **AND** the `duplicate` label is not already applied
- **WHEN** the command reaches label application
- **THEN** the command MUST prompt with `AskUserQuestion` with
  options `["Yes -- apply duplicate label", "No -- skip"]`
- **AND** the prompt MUST inform the user that the `duplicate`
  label signals the issue should be closed

#### Scenario: Duplicate label does not exist and needs creation

- **GIVEN** the triage classification resolves to `duplicate`
- **AND** the `duplicate` label does not exist in the repository
- **WHEN** the command reaches label creation
- **THEN** the creation gate MUST use the generic prompt
  `["Yes -- create label", "No -- skip"]`
- **AND** if created, the application gate MUST use the
  specialized duplicate prompt with close-semantics warning

## REMOVED Requirements

### Requirement: Automatic label application

The behavior of applying labels automatically without user
confirmation (line 243) is removed for all label categories.

Reason: Org policy from issue #346 audit mandates
`AskUserQuestion` for all irreversible external actions. Label
mutations are irreversible external actions.
