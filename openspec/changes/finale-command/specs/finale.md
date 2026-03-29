## ADDED Requirements

### Requirement: finale-branch-gate

The `/finale` command MUST refuse to run when the
current branch is `main`. It MUST display an error
message instructing the user to switch to a feature
branch.

#### Scenario: refuse on main
- **GIVEN** the user is on the `main` branch
- **WHEN** they run `/finale`
- **THEN** the command displays an error and stops
  without making any changes

### Requirement: finale-auto-stage

The `/finale` command MUST stage all uncommitted
changes using `git add .` before committing.

#### Scenario: stage all changes
- **GIVEN** the user has unstaged changes on a
  feature branch
- **WHEN** they run `/finale`
- **THEN** all changes are staged before the commit

### Requirement: finale-secrets-warning

The `/finale` command SHOULD warn the user and ask
for confirmation before staging files that likely
contain secrets (`.env`, `credentials.json`, etc.).

#### Scenario: warn about secrets
- **GIVEN** the working tree contains a `.env` file
- **WHEN** `/finale` is about to stage changes
- **THEN** the command warns about the file and asks
  for confirmation before proceeding

### Requirement: finale-commit-message

The `/finale` command MUST generate a conventional
commit message by analyzing the staged diff. It MUST
show the proposed message to the user for approval
before committing. If `$ARGUMENTS` are provided, they
SHOULD be used as a hint for the message.

#### Scenario: generate and confirm commit message
- **GIVEN** staged changes exist
- **WHEN** the commit message is generated
- **THEN** the user is shown the proposed message and
  asked to approve, edit, or provide their own

#### Scenario: use arguments as hint
- **GIVEN** the user runs `/finale fix the typo`
- **WHEN** the commit message is generated
- **THEN** the arguments are used as context for the
  generated message

### Requirement: finale-push

The `/finale` command MUST push the current branch to
the remote. If no upstream is set, it MUST use
`git push -u origin <branch>`.

#### Scenario: push with no upstream
- **GIVEN** the branch has never been pushed
- **WHEN** `/finale` pushes
- **THEN** it sets the upstream with `-u origin`

### Requirement: finale-pr-create

The `/finale` command MUST create a pull request if
one does not already exist for the current branch.
The PR title SHOULD be derived from the commit
messages on the branch. If a PR already exists, the
command MUST use the existing PR.

#### Scenario: create new PR
- **GIVEN** no PR exists for the current branch
- **WHEN** `/finale` reaches the PR step
- **THEN** it creates a PR with a title derived from
  the branch commits and a summary body

#### Scenario: use existing PR
- **GIVEN** a PR already exists for the current branch
- **WHEN** `/finale` reaches the PR step
- **THEN** it uses the existing PR number and skips
  creation

### Requirement: finale-watch-checks

The `/finale` command MUST watch CI checks until they
complete. If checks fail, it MUST stop and report the
failure with details. It MUST NOT proceed to merge
when checks have failed.

#### Scenario: checks pass
- **GIVEN** the PR has been created or found
- **WHEN** CI checks complete successfully
- **THEN** the command proceeds to merge

#### Scenario: checks fail
- **GIVEN** the PR has been created or found
- **WHEN** CI checks fail
- **THEN** the command stops, reports the failure, and
  asks the user how to proceed

### Requirement: finale-merge

The `/finale` command MUST merge the PR using rebase
strategy (`gh pr merge --rebase --delete-branch`).
Other merge strategies (squash, merge commit) MUST
NOT be used.

#### Scenario: rebase merge
- **GIVEN** CI checks have passed
- **WHEN** the command merges
- **THEN** it uses `--rebase --delete-branch`

### Requirement: finale-return-to-main

The `/finale` command MUST switch to `main` and pull
after a successful merge.

#### Scenario: return to main
- **GIVEN** the PR has been merged
- **WHEN** the merge completes
- **THEN** the command switches to `main` and runs
  `git pull`

### Requirement: finale-summary

The `/finale` command MUST display a completion summary
including: branch name, commit message, PR number and
URL, check status, and current branch state.

#### Scenario: success summary
- **GIVEN** all steps completed successfully
- **WHEN** the summary is displayed
- **THEN** it shows branch, commit, PR number, and
  confirms the user is on `main` and up to date

### Requirement: finale-scaffold

The `/finale` command file MUST be included in the
scaffold assets at
`internal/scaffold/assets/opencode/command/finale.md`
and the expected scaffold file count test MUST be
updated.

#### Scenario: scaffold includes finale
- **GIVEN** a user runs `uf init` in a new repo
- **WHEN** scaffold assets are deployed
- **THEN** `.opencode/command/finale.md` is created

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
