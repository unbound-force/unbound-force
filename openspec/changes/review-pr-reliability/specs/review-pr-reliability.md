## ADDED Requirements

### FR-001: Argument-First Parsing Gate

The `/review-pr` command MUST parse the PR number
from the user's message before any tool calls are
executed. If a PR number is provided, `PR_NUMBER`
MUST be set immediately and all auto-detection
commands MUST be skipped.

#### Scenario: PR number provided as argument
- **GIVEN** the user invokes `/review-pr 139`
- **WHEN** the command begins execution
- **THEN** `PR_NUMBER` is set to `139` before Step 0
- **AND** Step 1 is skipped entirely
- **AND** `gh pr view --json number`,
  `git branch --show-current`, and any other
  auto-detection commands are NOT executed

#### Scenario: No PR number provided
- **GIVEN** the user invokes `/review-pr` without an
  argument
- **WHEN** the command begins execution
- **THEN** Step 1 runs auto-detection via
  `gh pr view --json number --jq '.number'`
- **AND** if no open PR exists, the command STOPs
  with an error

### FR-002: Execution Mode Check

The command MUST verify that the agent can execute
local tools (build, test, lint) during Step 0
Prerequisites. If the agent is in a read-only or
plan-only mode, the command MUST STOP with an
actionable message before any metadata or diff
fetching occurs.

#### Scenario: Agent in plan/read-only mode
- **GIVEN** the agent is in plan mode or read-only
  mode
- **WHEN** `/review-pr 42` is invoked
- **THEN** the mode check detects the restriction
- **AND** the command STOPs with a message directing
  the user to switch to a full/auto mode
- **AND** no PR metadata, CI checks, or diff are
  fetched (tokens saved)

#### Scenario: Agent in full execution mode
- **GIVEN** the agent can execute commands normally
- **WHEN** `/review-pr 42` is invoked
- **THEN** the mode check passes silently
- **AND** execution continues to Step 1

### FR-003: CI Coverage Matrix

The command MUST build and display a visible CI
coverage matrix before deciding which local tools to
run. The matrix MUST map each detected local tool to
its CI equivalent and show the skip/run decision.

#### Scenario: CI covers all checks
- **GIVEN** CI checks for test, lint, and build all
  report PASS
- **WHEN** Step 4 begins
- **THEN** a coverage matrix is displayed showing
  each tool mapped to its CI check with "No" in the
  "Run locally?" column
- **AND** no local tools are executed

#### Scenario: CI has no checks
- **GIVEN** `gh pr checks` returns no checks at all
- **WHEN** Step 4 begins
- **THEN** a coverage matrix is displayed showing
  all detected local tools with "NONE" in the CI
  column and "Yes" in the "Run locally?" column
- **AND** all detected local tools are executed

#### Scenario: CI partially covers checks
- **GIVEN** CI has a passing test check but no lint
  check
- **WHEN** Step 4 begins
- **THEN** the matrix shows `go test` mapped to CI
  PASS (skip) and `golangci-lint` mapped to NONE
  (run)
- **AND** only `golangci-lint` is executed locally

### FR-004: Save-and-Navigate Diff Handling

For diffs exceeding 500 lines, the command MUST save
the full diff to a temporary file and navigate it
using file boundary detection and offset-based
reading. The command MUST NOT attempt file-filter
syntax on `gh pr diff`.

#### Scenario: Large diff processing
- **GIVEN** a PR with a diff exceeding 500 lines
- **WHEN** Step 5 fetches the diff
- **THEN** the diff is saved to a temp file
- **AND** file boundaries are found via
  `grep -n '^diff --git'`
- **AND** specific file sections are read via
  offset/limit

#### Scenario: Nonexistent syntax prevented
- **GIVEN** a PR with any diff size
- **WHEN** the agent needs to read a specific file's
  diff
- **THEN** `gh pr diff <N> -- <path>` is NOT executed
- **AND** `git show <remote>/<branch>:<path>` is NOT
  executed
- **AND** `git fetch <remote> <branch>` is NOT
  executed

### FR-005: GitHub API for PR File Contents

When the agent needs full file contents from the PR
branch (not just the diff), it MUST use the GitHub
API via `gh api` instead of git operations. Git
operations (`git show`, `git fetch`) MUST NOT be
used for PR branch access.

#### Scenario: Reading a full file from PR branch
- **GIVEN** the agent needs the complete content of
  a file that exists on the PR branch but not on the
  base branch
- **WHEN** the agent accesses the file
- **THEN** it uses `gh api repos/{owner}/{repo}/
  contents/<path>?ref=<headRefName>`
- **AND** it does NOT use `git show` or `git fetch`

### FR-006: PR-Introduced Spec Detection

Step 6 MUST check the PR's changed file list (from
Step 2 metadata) for spec artifacts when they are
not found on the local filesystem. Spec content
SHOULD be read from the saved diff file.

#### Scenario: Spec introduced by the PR
- **GIVEN** the PR adds files under
  `openspec/changes/<branch>/proposal.md`
- **AND** the file does not exist on the base branch
- **WHEN** Step 6 searches for specs
- **THEN** the spec is discovered in the PR's changed
  file list
- **AND** the spec content is read from the saved
  diff file
- **AND** the spec is used for alignment review

#### Scenario: Spec exists on base branch
- **GIVEN** the PR's spec artifacts already exist on
  the base branch
- **WHEN** Step 6 searches for specs
- **THEN** the spec is found on the local filesystem
  (existing behavior)
- **AND** no changed-file-list search is needed

## MODIFIED Requirements

### FR-007: Step 0 Prerequisites Extended

Previously: Step 0 verified only `gh` CLI
availability and authentication.

Now: Step 0 MUST additionally verify execution mode
capability (FR-002). The mode check MUST occur after
`gh auth status` and before any PR-specific commands.

### FR-008: Step 5 Large Diff Handling Rewritten

Previously: Step 5 instructed "process file-by-file
instead of loading the entire diff" without
specifying the mechanism.

Now: Step 5 MUST specify the save-and-navigate
technique (FR-004) and include an explicit DO NOT
list of failing commands.

### FR-009: Step 6 Spec Detection Extended

Previously: Step 6 checked only the local filesystem
for spec artifacts.

Now: Step 6 MUST also check the PR's changed file
list (FR-006) when the filesystem check finds
nothing.

## REMOVED Requirements

None.
