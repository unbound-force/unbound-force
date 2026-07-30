## ADDED Requirements

_None._

## MODIFIED Requirements

### Requirement: Dirty-tree guard enforcement in speckit.specify

The dirty-tree guard in `speckit.specify.md` (step 2, lines 42-50)
MUST use an explicit `AskUserQuestion` tool call to enforce user
confirmation when uncommitted changes are detected, rather than
relying on prose-only "STOP and ask" instructions.

Previously: The guard instructed the agent in prose to "STOP and
ask the user for confirmation" with no tool-call specification.
Under context compression, agents could skip the prose instruction
and proceed to `git checkout -b` without confirmation.

#### FR-001: AskUserQuestion enforcement

When `git status --short` reports uncommitted changes, the agent
MUST invoke the `AskUserQuestion` tool with the following options
before proceeding to branch creation:

- "Stash changes and continue"
- "Abort -- keep changes as-is"

The agent MUST NOT proceed to `git checkout -b` until the user
has responded to the `AskUserQuestion` prompt.

#### FR-002: Display uncommitted files

The agent MUST display the list of uncommitted files (output of
`git status --short`) to the user as context within the
`AskUserQuestion` prompt, so the user can make an informed
decision.

#### FR-003: Stash handling

If the user selects "Stash changes and continue", the agent
MUST run `git stash` before proceeding to `git checkout -b`.
If `git stash` exits with a non-zero status, the agent MUST
stop execution and report the stash failure to the user rather
than proceeding to branch creation.

#### FR-004: Abort handling

If the user selects "Abort -- keep changes as-is", the agent
MUST stop execution and report that the operation was aborted
due to uncommitted changes.

#### FR-005: Exception clause preservation

The exception for explicit user commands ("user explicitly said
to create a new spec in the same message") MUST still require
the `AskUserQuestion` tool call when uncommitted changes are
detected. The exception means the agent need not proactively
warn about dirty trees in general, but when changes ARE detected,
confirmation is still REQUIRED.

#### Scenario: Uncommitted changes detected before branch creation

- **GIVEN** the user runs `/speckit.specify` with a feature
  description
- **AND** `git status --short` reports one or more uncommitted
  changes
- **WHEN** the agent reaches the branch creation step
- **THEN** the agent MUST invoke `AskUserQuestion` with options
  "Stash changes and continue" and "Abort -- keep changes as-is"
- **AND** the agent MUST display the list of uncommitted files
- **AND** the agent MUST NOT proceed to `git checkout -b` until
  the user responds

#### Scenario: User chooses to stash and continue

- **GIVEN** the agent has presented the `AskUserQuestion` prompt
  for uncommitted changes
- **WHEN** the user selects "Stash changes and continue"
- **THEN** the agent MUST run `git stash`
- **AND** the agent MUST proceed to `git checkout -b`

#### Scenario: User chooses to abort

- **GIVEN** the agent has presented the `AskUserQuestion` prompt
  for uncommitted changes
- **WHEN** the user selects "Abort -- keep changes as-is"
- **THEN** the agent MUST stop execution
- **AND** the agent MUST report that the operation was aborted

#### Scenario: git stash fails

- **GIVEN** the user selected "Stash changes and continue"
- **WHEN** `git stash` exits with a non-zero status
- **THEN** the agent MUST stop execution
- **AND** the agent MUST report the stash failure to the user
- **AND** the agent MUST NOT proceed to `git checkout -b`

#### Scenario: Clean working tree

- **GIVEN** the user runs `/speckit.specify` with a feature
  description
- **AND** `git status --short` reports no uncommitted changes
- **WHEN** the agent reaches the branch creation step
- **THEN** the agent MUST proceed directly to branch creation
  without invoking `AskUserQuestion`

#### Scenario: Explicit command with uncommitted changes

- **GIVEN** the user explicitly typed `/speckit.specify add auth`
  in the same message
- **AND** `git status --short` reports uncommitted changes
- **WHEN** the agent reaches the branch creation step
- **THEN** the agent MUST still invoke `AskUserQuestion` with
  the stash/abort options
- **AND** the agent MUST NOT silently switch branches

## REMOVED Requirements

_None._
