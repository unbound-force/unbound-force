## ADDED Requirements

### Requirement: Archive confirmation gate

The `openspec-archive-change` skill MUST include an
`AskUserQuestion` gate between step 5 (commit and push)
and step 6 (perform the archive) that confirms the user
is ready to proceed with archiving.

The gate MUST present two options:
- "Changes committed and pushed -- proceed to archive"
- "Abort -- need to commit first"

If the user selects the abort option, the skill MUST stop
the archive workflow and inform the user to commit their
changes before retrying. The skill MUST NOT proceed to
step 6 or step 7.

#### Scenario: User confirms archive after clean commit

- **GIVEN** step 5 has completed and all changes are
  committed and pushed
- **WHEN** the `AskUserQuestion` gate is presented
- **AND** the user selects "Changes committed and pushed
  -- proceed to archive"
- **THEN** the skill proceeds to step 6 (perform the
  archive)

#### Scenario: User aborts archive to commit first

- **GIVEN** step 5 has completed
- **WHEN** the `AskUserQuestion` gate is presented
- **AND** the user selects "Abort -- need to commit first"
- **THEN** the skill stops the archive workflow
- **AND** informs the user to commit changes before
  retrying

### Requirement: Branch switch confirmation gate

The `openspec-archive-change` skill MUST include an
`AskUserQuestion` gate before executing `git checkout main`
in step 7 (return to main branch).

The gate MUST present two options:
- "Return to main"
- "Stay on branch"

If the user selects "Stay on branch", the skill MUST skip
the `git checkout main` command and note in the summary
that the user remained on the `opsx/<name>` branch.

#### Scenario: User confirms return to main

- **GIVEN** the archive commit has been pushed in step 7
- **WHEN** the `AskUserQuestion` gate is presented before
  `git checkout main`
- **AND** the user selects "Return to main"
- **THEN** the skill executes `git checkout main`
- **AND** continues to step 8 (display summary)

#### Scenario: User stays on branch

- **GIVEN** the archive commit has been pushed in step 7
- **WHEN** the `AskUserQuestion` gate is presented before
  `git checkout main`
- **AND** the user selects "Stay on branch"
- **THEN** the skill skips `git checkout main`
- **AND** notes in the step 8 summary that the user
  remained on the `opsx/<name>` branch

## MODIFIED Requirements

No existing requirements are modified. The existing
step 5 commit/push logic and step 6 archive logic remain
unchanged. The gates are additive insertions at step
boundaries.

## REMOVED Requirements

No requirements are removed.
