## ADDED Requirements

### Requirement: Dirty-tree AskUserQuestion gate

When a dirty working tree is detected during Step 3a,
the agent MUST present the user with an `AskUserQuestion`
call offering exactly two options:

1. "Stash changes and continue"
2. "Abort — keep changes as-is"

If the user selects "Stash changes and continue", the
agent MUST run `git stash` before proceeding with branch
creation.

If the user selects "Abort", the agent MUST stop
execution and report that the user needs to handle
uncommitted changes manually.

The agent MUST NOT silently switch branches when
uncommitted changes are present.

#### Scenario: Dirty tree with staged changes

- **GIVEN** the working tree has staged changes
- **WHEN** the agent reaches Step 3a (branch creation)
- **THEN** the agent MUST present AskUserQuestion with
  the two options before proceeding

#### Scenario: Dirty tree — user selects stash

- **GIVEN** the working tree has uncommitted changes
- **WHEN** the user selects "Stash changes and continue"
- **THEN** the agent MUST run `git stash` and then
  proceed to create the branch

#### Scenario: Dirty tree — user selects abort

- **GIVEN** the working tree has uncommitted changes
- **WHEN** the user selects "Abort — keep changes as-is"
- **THEN** the agent MUST stop execution and report that
  uncommitted changes need to be handled first

#### Scenario: Clean working tree

- **GIVEN** the working tree has no uncommitted changes
- **WHEN** the agent reaches Step 3a
- **THEN** the agent MUST proceed directly to the branch
  check (Step 3b) without presenting AskUserQuestion

### Requirement: STOP HERE preamble placement

A bolded preamble MUST appear immediately after the
`**Steps**` heading and before Step 1. The preamble
MUST state:

- This command creates spec artifacts only
- It MUST NOT implement code, commit, push, create PRs,
  or invoke implementation commands
- After artifacts are complete, STOP and prompt the user

The existing STOP HERE block after Step 6 SHOULD be
retained as reinforcement.

#### Scenario: Preamble precedes all workflow steps

- **GIVEN** an agent loads the opsx-propose command or
  openspec-propose skill
- **WHEN** it begins processing the Steps section
- **THEN** the first content after `**Steps**` MUST be
  the preamble stating the artifacts-only prohibition

#### Scenario: Existing STOP HERE retained

- **GIVEN** the preamble has been added before Step 1
- **WHEN** the agent processes the full instruction file
- **THEN** the STOP HERE block after Step 6 MUST still
  be present as reinforcement

## MODIFIED Requirements

### Requirement: Dirty-tree guard prose

Previously: The dirty-tree guard in Step 3a described
the check in narrative prose ("STOP and ask the user
for confirmation") without specifying the tool or
options to use.

Modified: The guard now MUST specify `AskUserQuestion`
with two concrete options. The prose description of
behavior is retained but augmented with the explicit
tool call specification.

## REMOVED Requirements

(none)
