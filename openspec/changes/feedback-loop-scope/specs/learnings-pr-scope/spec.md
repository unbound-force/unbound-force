<!-- capability: learnings-pr-scope
     change: feedback-loop-scope
     supersedes: dewey-learning-pr-policy, finale-dewey-commit (archived without merging specs)
     requirement ids: LPS-FR-001 allocated here -->

## ADDED Requirements

### Requirement: Learnings Stay Off PR Branches

**ID: LPS-FR-001** This requirement MUST be satisfied.

Dewey learnings under `.uf/dewey/learnings/` MUST NOT be committed on
a feature or PR branch. They MUST reach the default branch through a
separate `chore/dewey-learnings-*` pull request. `uf_finale_stage`
MUST NOT stage them, MUST unstage any already staged, and MUST
succeed when the directory is absent or untracked.

#### Scenario: Finale staging

- **GIVEN** untracked learnings and changed source files
- **WHEN** `uf_finale_stage` runs
- **THEN** the source files MUST be staged
- **AND** no file under `.uf/dewey/learnings/` MUST be staged

#### Scenario: No learnings directory

- **GIVEN** a repository with no `.uf/dewey/learnings/` directory
- **WHEN** `uf_finale_stage` runs
- **THEN** it MUST succeed

#### Scenario: Feedback commit

- **GIVEN** `/uf.address-feedback` stored lessons during assessment
- **WHEN** it commits fixes
- **THEN** the commits MUST NOT include `.uf/dewey/learnings/`

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
