## MODIFIED Requirements

### Requirement: /finale Workflow Steps

The `/finale` command MUST execute steps 1 through 6
and step 8, skipping step 7 (Merge PR). After CI
checks pass, the command MUST return to main without
merging the PR.

Previously: The command executed all 9 steps including
step 7 (`gh pr merge --rebase --delete-branch`).

#### Scenario: PR stays open after finale

- **GIVEN** a feature branch with committed changes
- **WHEN** the user runs `/finale`
- **THEN** the command commits, pushes, creates a PR,
  watches CI checks, returns to main
- **AND** the PR remains open (not merged)

#### Scenario: Summary reports ready for review

- **GIVEN** `/finale` completes successfully
- **WHEN** the summary is displayed
- **THEN** the summary states the PR is "ready for
  review" (not "merged via rebase")
- **AND** includes a next step: "Request reviewers,
  then merge after approval"

### Requirement: /finale Guardrails

The guardrails section MUST NOT reference merge
behavior. The guardrails "NEVER merge with failing
checks" and "ALWAYS use rebase merge" MUST be removed.

Previously: Both guardrails were present because
`/finale` performed the merge.

## REMOVED Requirements

### Requirement: Automatic PR Merge

The `/finale` command no longer merges PRs. Step 7
(Merge PR) is removed entirely. Users merge via
GitHub UI or `gh pr merge` after reviewer approval.

## ADDED Requirements

None.
