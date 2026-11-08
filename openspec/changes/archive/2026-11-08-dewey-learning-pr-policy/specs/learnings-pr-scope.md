## ADDED Requirements

### Requirement: Learnings Are Feature PR Scope

Dewey learnings (`.uf/dewey/learnings/*.md` and `.uf/dewey/compiled/*.md`) produced during a change's workflow (e.g., `/uf.unleash` Step 9 retrospective via `dewey_store_learning`) MUST be included in the feature PR that generated them. These files are intentionally part of the change, not unrelated scope noise.

#### Scenario: Retrospective learnings in feature PR
- **GIVEN** a feature branch has completed implementation and review
- **WHEN** the `/uf.unleash` Step 9 retrospective produces learnings stored via `dewey_store_learning`
- **THEN** the resulting `.uf/dewey/learnings/*.md` and `.uf/dewey/compiled/*.md` files MUST be committed on the feature branch and included in the feature PR

#### Scenario: Reviewer evaluates learnings relevance
- **GIVEN** a feature PR includes Dewey learning files
- **WHEN** a reviewer assesses the PR scope
- **THEN** the learning files MUST be treated as in-scope documentation artifacts related to the active change

### Requirement: Commit Scope Rule Clarification

The Commit scope behavioral rule in `AGENTS.md` MUST explicitly state that Dewey learnings produced during the change's workflow are considered directly related to the active spec or change.

#### Scenario: Agent stages files for commit
- **GIVEN** an agent is preparing a commit on a feature branch
- **WHEN** `.uf/dewey/learnings/*.md` files exist that were produced during the change's workflow
- **THEN** the agent MUST include these files in the commit as in-scope artifacts

### Requirement: Documentation Gate Satisfaction

Learnings produced by the retrospective MUST be recognized as satisfying the Documentation gate. They are intentional documentation artifacts that capture session knowledge, not orphaned or aspirational content.

#### Scenario: Documentation gate assessment
- **GIVEN** a task is being marked complete
- **WHEN** the Documentation gate is assessed
- **THEN** retrospective learnings committed in the same PR SHOULD be noted as satisfying the documentation impact assessment for knowledge-capture purposes

## MODIFIED Requirements

### Requirement: Commit Scope Behavioral Rule

**Commit scope**: Only commit files directly related to the active spec or change. Dewey learnings (`.uf/dewey/learnings/*.md` and `.uf/dewey/compiled/*.md`) produced during the change's workflow are directly related to the active change and MUST be included in the feature PR. Tooling scaffolds (`uf init`, convention pack updates, command directory renames, schema template updates) MUST be committed on a separate branch (e.g., `chore/uf.init-sync`), not mixed into feature branches. Never use `git add -A` or `git add .` on feature branches — stage files explicitly.

(note: "Previously: **Commit scope**: Only commit files directly related to the active spec or change. Tooling scaffolds (`uf init`, convention pack updates, command directory renames, schema template updates) MUST be committed on a separate branch (e.g., `chore/uf.init-sync`), not mixed into feature branches. Never use `git add -A` or `git add .` on feature branches — stage files explicitly.")

## REMOVED Requirements

None.
<!-- scaffolded by uf vdev -->
