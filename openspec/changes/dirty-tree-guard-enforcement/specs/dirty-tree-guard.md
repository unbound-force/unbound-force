## ADDED Requirements

### Requirement: Explicit AskUserQuestion on dirty tree

When uncommitted changes are detected by `git status --short`,
the agent MUST call the AskUserQuestion tool with the following
options before proceeding to `git checkout -b`:

- "Stash changes and continue"
- "Abort -- keep changes as-is"

The agent MUST NOT proceed to branch creation without receiving
a response from the AskUserQuestion tool.

#### Scenario: Dirty tree with uncommitted changes

- **GIVEN** the working tree has uncommitted changes (staged,
  unstaged, or untracked files visible in `git status --short`
  output)
- **WHEN** the agent reaches the branch creation step
- **THEN** the agent MUST first display the output of
  `git status --short` to the user, then call AskUserQuestion
  with two options: "Stash changes and continue" and "Abort --
  keep changes as-is". The display MUST occur before the
  AskUserQuestion call so the user can make an informed decision

#### Scenario: User selects stash and continue

- **GIVEN** the AskUserQuestion is presented with dirty tree
- **WHEN** the user selects "Stash changes and continue"
- **THEN** the agent MUST run `git stash push -m
  "openspec-propose: auto-stash before branch switch"`.
  If the stash command succeeds, the agent MUST print
  "Changes stashed. Run `git stash pop` to restore them
  when ready." and proceed to `git checkout -b`.
  If the stash command fails (non-zero exit), the agent
  MUST stop immediately, display the error output, and
  MUST NOT proceed to branch creation

#### Scenario: User selects abort

- **GIVEN** the AskUserQuestion is presented with dirty tree
- **WHEN** the user selects "Abort -- keep changes as-is"
- **THEN** the agent MUST stop immediately and MUST NOT
  create the branch or any artifacts

#### Scenario: Clean working tree

- **GIVEN** the working tree has no uncommitted changes
  (`git status --short` produces no output)
- **WHEN** the agent reaches the branch creation step
- **THEN** the agent MUST proceed directly to the branch
  check (part b) without prompting the user

#### Scenario: Stash operation fails

- **GIVEN** the user selects "Stash changes and continue"
- **WHEN** `git stash push` exits with a non-zero status
- **THEN** the agent MUST stop immediately, display the
  error output, and MUST NOT proceed to branch creation

#### Scenario: Exception clause preserved for explicit
  change requests

- **GIVEN** the user explicitly requested a new change in
  the same message (e.g., `/opsx-propose fix-typos`)
- **WHEN** the working tree has uncommitted changes
- **THEN** the AskUserQuestion MUST still be presented --
  the exception does not bypass the guard, it only
  clarifies that explicit requests do not exempt the user
  from confirmation

## MODIFIED Requirements

### Requirement: Dirty-tree guard confirmation mechanism

The dirty-tree guard MUST use the AskUserQuestion tool with
explicit options instead of prose-only instructions to "ask
the user for confirmation."

Previously: "STOP and ask the user for confirmation before
switching branches. Show what uncommitted changes exist and
warn that switching branches with a dirty working tree may
cause changes to be applied to the wrong branch."

The new text replaces the prose with a concrete tool call
specification including exact option labels and behavior for
each selection.

## REMOVED Requirements

(None)
