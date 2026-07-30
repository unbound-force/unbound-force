## ADDED Requirements

### Requirement: Push confirmation gate (FR-001)

The `/finale` command MUST present an AskUserQuestion
confirmation gate immediately before executing `git push`
in Step 4 (Push to Remote). The gate MUST offer two
options: "Push to remote" and "Abort -- do not push".

If the user selects "Abort -- do not push", the command
MUST stop and report that the push was aborted. The
commit from Step 3 is preserved locally.

If the user selects "Push to remote", the command MUST
proceed with the `git push` execution as currently
defined.

#### Scenario: User confirms push

- **GIVEN** the user has completed Step 3 (commit
  approved and created)
- **WHEN** Step 4 presents the AskUserQuestion gate
- **AND** the user selects "Push to remote"
- **THEN** the command executes `git push` (or
  `git push -u origin <branch>` if no upstream is set)

#### Scenario: User aborts push

- **GIVEN** the user has completed Step 3 (commit
  approved and created)
- **WHEN** Step 4 presents the AskUserQuestion gate
- **AND** the user selects "Abort -- do not push"
- **THEN** the command stops with a message indicating
  the push was aborted
- **AND** the local commit from Step 3 is preserved

#### Scenario: Push gate after upstream detection

- **GIVEN** the command has checked for an upstream
  branch via `git rev-parse --abbrev-ref @{upstream}`
- **WHEN** the AskUserQuestion gate is presented
- **THEN** the gate appears after the upstream check
  but before the actual push command

## MODIFIED Requirements

### Requirement: Step 4 push execution (FR-002)

The `git push` commands in Step 4 MUST only execute
after the user has selected "Push to remote" from the
AskUserQuestion gate. Previously, push executed
immediately after upstream detection with no user
confirmation.

Previously: "If no upstream: `git push -u origin
<branch>`. If upstream exists: `git push`." (executed
unconditionally)

## REMOVED Requirements

None.
<!-- scaffolded by uf vdev -->
