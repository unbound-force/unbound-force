## ADDED Requirements

### Requirement: Re-review Evaluator

`/uf.review-pr` MUST choose between the full council and a light
check when the current user already reviewed an earlier commit of
the PR. It MUST choose the light check only when the reviewed commit
is an ancestor of the head, the delta is at most 200 changed lines,
no file was added, removed, or renamed, no sensitive path changed,
the prior review raised no CRITICAL finding, and CI shows no
PR-caused failure.

#### Scenario: First review

- **GIVEN** no earlier review by the current user
- **WHEN** `/uf.review-pr` runs
- **THEN** it MUST run the full council

#### Scenario: Rebased branch

- **GIVEN** the reviewed commit is no longer in the branch
- **WHEN** `/uf.review-pr` runs
- **THEN** it MUST run the full council

#### Scenario: Small fix push

- **GIVEN** a push of 40 lines to two already-reviewed files
- **WHEN** `/uf.review-pr` runs
- **THEN** it MUST run the light check and dispatch no child agents

### Requirement: Light Check Verdict

The light check MUST approve only when every CRITICAL and HIGH
finding of the prior review is resolved, local tools pass, and the
new commits introduce no CRITICAL or HIGH problem. It MUST request
changes when a prior CRITICAL or HIGH finding is still open, and
MUST escalate to the full council when it finds a new CRITICAL or
HIGH problem.

#### Scenario: Prior HIGH finding still open

- **GIVEN** a light check where a prior HIGH finding is unchanged
- **THEN** the verdict MUST be REQUEST CHANGES

#### Scenario: New HIGH problem

- **GIVEN** a light check that finds a new HIGH problem in the delta
- **THEN** the command MUST run the full council before posting
