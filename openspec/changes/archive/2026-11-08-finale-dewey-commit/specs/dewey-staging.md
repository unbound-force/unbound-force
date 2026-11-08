## ADDED Requirements

### Requirement: Dewey artifact staging

The `/uf.finale` command MUST explicitly stage Dewey
knowledge artifacts (`.uf/dewey/learnings/` and
`.uf/dewey/compiled/`) during Step 2 (Check for Changes
to Commit), after the general `git add .` step.

The staging command MUST suppress stderr so it is
idempotent when the directories do not exist.

The instruction MUST include a brief rationale
explaining why Dewey artifacts belong in the commit.

The instruction MUST include a directive that agents
SHALL NOT skip this step.

#### Scenario: Dewey files exist in working tree

- **GIVEN** the working tree contains new or modified
  files under `.uf/dewey/learnings/` or
  `.uf/dewey/compiled/`
- **WHEN** the agent executes `/uf.finale` Step 2
- **THEN** those Dewey files MUST be staged for commit

#### Scenario: No Dewey files exist

- **GIVEN** the working tree does not contain any files
  under `.uf/dewey/learnings/` or `.uf/dewey/compiled/`
- **WHEN** the agent executes `/uf.finale` Step 2
- **THEN** the Dewey staging sub-step MUST complete
  without error (no-op)

#### Scenario: Dewey files already staged by git add .

- **GIVEN** `git add .` has already staged Dewey files
- **WHEN** the explicit Dewey staging command runs
- **THEN** the command MUST be idempotent (no duplicate
  staging, no error)

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
