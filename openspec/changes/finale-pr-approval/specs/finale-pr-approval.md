## ADDED Requirements

### Requirement: PR content confirmation

`/finale` Step 5 MUST show the proposed PR title and body
to the user for approval before calling `gh pr create`.
The confirmation prompt MUST offer three options:
approve, edit, or provide their own content.

#### Scenario: User approves generated PR content

- **GIVEN** `/finale` has generated a PR title and body
  from commit history
- **WHEN** the user is shown the proposed PR content
- **THEN** the command displays the title and body in a
  formatted block and asks "Approve, edit, or provide
  your own?"
- **AND** if the user approves, the PR is created with
  the generated title and body

#### Scenario: User edits the PR content

- **GIVEN** `/finale` has shown the proposed PR content
- **WHEN** the user chooses to edit
- **THEN** the command accepts the user's modified title
  and/or body
- **AND** the PR is created with the user's edits

#### Scenario: User provides their own PR content

- **GIVEN** `/finale` has shown the proposed PR content
- **WHEN** the user provides their own title and body
- **THEN** the PR is created with the user-provided
  content instead of the generated content

### Requirement: PR creation guardrail

The Guardrails section of `/finale` MUST include the rule:
"NEVER create a PR without user approval of the title
and body."

#### Scenario: Guardrail consistency

- **GIVEN** `/finale` has a guardrail "NEVER commit
  without user approval of the message"
- **WHEN** the PR creation guardrail is added
- **THEN** both guardrails exist as symmetric protections
  for the two user-facing content generation steps

## MODIFIED Requirements

### Requirement: Step 5 substep ordering

Step 5 substeps are reordered to accommodate the
confirmation step. Previously: (a) fork detection,
(b) generate title, (c) generate body, (d) create PR,
(e) report URL. New ordering: (a) fork detection,
(b) generate title, (c) generate body, **(d) confirm
with user**, (e) create PR, (f) report URL.

Previously: Steps b through d had no user interaction
between content generation and PR creation.

## REMOVED Requirements

None.
