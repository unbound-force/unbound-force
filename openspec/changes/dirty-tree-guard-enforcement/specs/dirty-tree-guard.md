## ADDED Requirements

### Requirement: AskUserQuestion enforcement for dirty-tree guard

When uncommitted changes are detected during the
dirty-tree check (step 3a), the agent MUST use the
AskUserQuestion tool to confirm the user's intent before
proceeding to `git checkout -b`. The agent MUST NOT
proceed based on prose-only reasoning.

The AskUserQuestion tool MUST be called with the
following options:
1. "Stash changes and continue"
2. "Abort -- keep changes as-is"

#### Scenario: Uncommitted changes detected, user stashes

- **GIVEN** `git status --short` returns one or more lines
  of output indicating uncommitted changes
- **WHEN** the agent reaches the dirty-tree guard in
  step 3a
- **THEN** the agent MUST display the `git status --short`
  output, then call the AskUserQuestion tool with the
  two options above. If the user selects "Stash changes
  and continue", the agent MUST run `git stash` before
  proceeding to `git checkout -b`.

#### Scenario: Uncommitted changes detected, user aborts

- **GIVEN** `git status --short` returns one or more lines
  of output indicating uncommitted changes
- **WHEN** the agent reaches the dirty-tree guard in
  step 3a and the user selects "Abort -- keep changes
  as-is"
- **THEN** the agent MUST stop execution and report that
  the user should handle their uncommitted work before
  retrying. The agent MUST NOT proceed to
  `git checkout -b`.

#### Scenario: Clean working tree

- **GIVEN** `git status --short` returns no output
- **WHEN** the agent reaches the dirty-tree guard in
  step 3a
- **THEN** the agent MUST proceed directly to the branch
  check (step 3b) without calling AskUserQuestion.

## MODIFIED Requirements

### Requirement: Dirty-tree guard in SKILL.md and opsx-propose.md

Previously: The dirty-tree guard described the
confirmation requirement in prose only ("STOP and ask
the user for confirmation before switching branches").

Now: The dirty-tree guard MUST specify the exact
AskUserQuestion tool call with option labels. The
replacement text SHALL be identical in both
`.opencode/skills/openspec-propose/SKILL.md` (lines
48-64) and `.opencode/commands/opsx-propose.md` (lines
41-57).

## REMOVED Requirements

None.
