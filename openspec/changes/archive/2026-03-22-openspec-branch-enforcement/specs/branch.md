## ADDED Requirements

### Requirement: opsx-branch-creation

The `/opsx-propose` command MUST create and checkout a
git branch named `opsx/<change-name>` after creating
the change directory.

#### Scenario: propose creates branch

- **GIVEN** the developer is on `main`
- **WHEN** `/opsx-propose fix-auth-bug` is run
- **THEN** branch `opsx/fix-auth-bug` is created and
  checked out, and the change directory exists at
  `openspec/changes/fix-auth-bug/`

#### Scenario: propose on existing opsx branch

- **GIVEN** the developer is on `opsx/other-change`
- **WHEN** `/opsx-propose new-thing` is run
- **THEN** the command errors with "Already on branch
  opsx/other-change -- finish or archive that change
  first"

### Requirement: opsx-branch-validation

The `/opsx-apply` command MUST validate that the
current branch is `opsx/<change-name>` before
implementing tasks.

#### Scenario: apply on correct branch

- **GIVEN** the developer is on `opsx/fix-auth-bug`
- **WHEN** `/opsx-apply fix-auth-bug` is run
- **THEN** implementation proceeds normally

#### Scenario: apply on wrong branch

- **GIVEN** the developer is on `main`
- **WHEN** `/opsx-apply fix-auth-bug` is run
- **THEN** the command errors with "Must be on branch
  opsx/fix-auth-bug. Run: git checkout
  opsx/fix-auth-bug"

### Requirement: opsx-branch-cleanup

The `/opsx-archive` command MUST checkout `main`
after archiving a change.

#### Scenario: archive returns to main

- **GIVEN** the developer is on `opsx/fix-auth-bug`
  and all tasks are complete
- **WHEN** `/opsx-archive fix-auth-bug` is run
- **THEN** the change is archived and the developer
  is on `main`

### Requirement: cobalt-crush-opsx-branch-check

The `/cobalt-crush` command's OpenSpec detection path
MUST validate the current branch matches the detected
change.

#### Scenario: cobalt-crush with matching branch

- **GIVEN** the developer is on `opsx/fix-auth-bug`
  and `openspec/changes/fix-auth-bug/tasks.md` exists
- **WHEN** `/cobalt-crush` is run without arguments
- **THEN** implementation proceeds for the fix-auth-bug
  change

#### Scenario: cobalt-crush with mismatched branch

- **GIVEN** the developer is on `main` and
  `openspec/changes/fix-auth-bug/tasks.md` exists
- **WHEN** `/cobalt-crush` is run without arguments
- **THEN** the command errors with a hint to checkout
  the correct branch

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
