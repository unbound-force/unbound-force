## ADDED Requirements

### Requirement: CRAP Load CI Check

The repository MUST include a GitHub Actions workflow that runs
CRAP analysis on pull requests targeting the main branch. The
workflow MUST call the reusable CRAP analysis workflow from
`complytime/org-infra` and MUST fail when per-function CRAP
score regressions or new function threshold violations are
detected. Two thresholds apply: a per-function regression
comparison (current score vs baseline value) and a new
function ceiling (default 30) for functions with no
baseline entry.

#### Scenario: PR with CRAP regression

- **GIVEN** a pull request that modifies a Go function
- **WHEN** the modified function's CRAP score exceeds its
  baseline value
- **THEN** the CI check MUST fail and a PR comment MUST be
  posted showing the regression details

#### Scenario: PR with new high-CRAP function

- **GIVEN** a pull request that adds a new Go function
- **WHEN** the new function's CRAP score exceeds the new
  function threshold (default 30)
- **THEN** the CI check MUST fail and a PR comment MUST
  report the function as a threshold violation

#### Scenario: PR with no Go changes

- **GIVEN** a pull request that modifies only non-Go files
- **WHEN** the CRAP analysis runs
- **THEN** the check MUST pass with a comment indicating
  no Go code changes were detected

#### Scenario: PR with CRAP improvement

- **GIVEN** a pull request that improves a function's CRAP
  score (lower than baseline)
- **WHEN** the CRAP analysis compares against baseline
- **THEN** the check MUST pass and the PR comment MUST
  list the function under improvements

### Requirement: Reusable Workflow Pinning

The consumer workflow MUST reference the reusable workflow
using a pinned commit SHA, not a branch name. The SHA MUST
be managed by dependabot for automated updates.

#### Scenario: Dependabot updates reusable workflow SHA

- **GIVEN** a new commit is pushed to the default branch of
  `complytime/org-infra`
- **WHEN** dependabot runs its daily check
- **THEN** dependabot SHOULD open a PR updating the pinned
  SHA in `ci_crapload.yml`

### Requirement: CRAP Score Baseline

The repository MUST maintain a `.gaze/baseline.json` file
containing per-function CRAP scores for all Go functions.
The baseline MUST be committed to version control and used
by the CI workflow for regression detection.

#### Scenario: Baseline file present

- **GIVEN** `.gaze/baseline.json` exists and contains valid
  CRAP score data
- **WHEN** the CI workflow runs CRAP analysis
- **THEN** the workflow MUST compare current scores against
  the baseline and report regressions and improvements

#### Scenario: Baseline file absent

- **GIVEN** `.gaze/baseline.json` does not exist
- **WHEN** the CI workflow runs CRAP analysis
- **THEN** the workflow MUST pass with a warning and skip
  per-function comparison

### Requirement: Dependabot Configuration

The repository MUST include a `.github/dependabot.yml` that
configures automated version updates for the `github-actions`
ecosystem (daily) and the `gomod` ecosystem (weekly).

#### Scenario: Action SHA becomes stale

- **GIVEN** a GitHub Action used in any workflow has a newer
  version available
- **WHEN** dependabot runs its daily check
- **THEN** dependabot SHOULD open a PR updating the pinned
  SHA with a `ci:` commit prefix

#### Scenario: Go module update available

- **GIVEN** a Go module in `go.mod` has a newer version
- **WHEN** dependabot runs its weekly check
- **THEN** dependabot SHOULD open a PR (up to 10 open)
  with a `chore:` commit prefix
