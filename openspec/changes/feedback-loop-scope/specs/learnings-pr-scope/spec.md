## ADDED Requirements

### Requirement: Learnings Stay Off PR Branches

Dewey learnings under `.uf/dewey/learnings/` MUST NOT be committed on
a feature or PR branch. They MUST reach the default branch through a
separate `chore/dewey-learnings-*` pull request.

#### Scenario: Finale staging

- **GIVEN** untracked learnings and changed source files
- **WHEN** `uf_finale_stage` runs
- **THEN** the source files MUST be staged
- **AND** no file under `.uf/dewey/learnings/` MUST be staged

#### Scenario: Feedback commit

- **GIVEN** `/uf.address-feedback` stored lessons during assessment
- **WHEN** it commits fixes
- **THEN** the commits MUST NOT include `.uf/dewey/learnings/`
