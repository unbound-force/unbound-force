## ADDED Requirements

### Requirement: Dependency Review Pipeline

All pull requests targeting the main branch MUST be
evaluated by the org-infra dependency review workflow.
When the PR author is `dependabot[bot]`, the workflow
MUST additionally run the dependabot reviewer workflow,
post a structured review comment, and conditionally
auto-approve the PR.

Auto-approval submits a GitHub PR review with `APPROVE`
disposition via `github.rest.pulls.createReview`. It
MUST NOT merge the PR directly. The auto-approve job
MUST have `pull-requests: write` at job level.

Auto-approval MUST require all four criteria:
- Risk level is not `high`
- Dependency review conclusion is `success`
- Release age is known (not `-1`)
- Release age is >= 24 hours

#### Scenario: Dependabot PR with safe update

- **GIVEN** a pull request authored by `dependabot[bot]`
  targeting main with a minor version bump
- **WHEN** the dependency review passes and the release
  age exceeds 24 hours
- **THEN** a structured review table is posted as a PR
  comment and the PR is auto-approved

#### Scenario: Dependabot PR with high-risk update

- **GIVEN** a pull request authored by `dependabot[bot]`
  targeting main with a major version bump
- **WHEN** the dependabot reviewer classifies risk as
  `high`
- **THEN** a structured review table is posted with
  "Manual review required" and the PR is NOT
  auto-approved

#### Scenario: Non-dependabot PR with dependency changes

- **GIVEN** a pull request authored by a human
  contributor that modifies `go.mod`
- **WHEN** the workflow runs
- **THEN** only the general dependency review executes;
  the dependabot-specific jobs (comment, auto-approve)
  are skipped

### Requirement: Security Scanning Pipeline

Push to main and pull requests targeting main MUST
trigger vulnerability scanning and supply chain security
assessment. The workflow MUST include:
- OSV-Scanner for known CVE detection in dependencies
- Trivy source scan for secrets and misconfigurations
- OpenSSF Scorecards for supply chain security posture

All scanners MUST upload SARIF results to GitHub's
Security tab. Security findings MUST block the pipeline
(no `|| true` or `continue-on-error` on scanning steps).

#### Scenario: Vulnerability detected in dependency

- **GIVEN** a pull request that introduces a dependency
  with a known CVE
- **WHEN** the security workflow runs
- **THEN** OSV-Scanner detects the vulnerability, uploads
  SARIF, and the workflow fails

#### Scenario: Secret detected in source

- **GIVEN** a pull request that adds a file containing a
  pattern matching a known secret format
- **WHEN** the Trivy source scan runs
- **THEN** the secret is flagged and the workflow fails

### Requirement: Scheduled Vulnerability Monitoring

A scheduled workflow MUST run daily at midnight UTC to
detect newly-disclosed vulnerabilities in existing
dependencies. The workflow MUST include OSV-Scanner and
OpenSSF Scorecards.

#### Scenario: New CVE disclosed after merge

- **GIVEN** a dependency that was clean when merged
- **WHEN** a new CVE is disclosed for that dependency
  and the daily scheduled scan runs
- **THEN** the OSV-Scanner detects the vulnerability
  and uploads SARIF to the Security tab

### Requirement: Multi-Language Linting

Push to main and pull requests targeting main MUST
trigger MegaLinter via the org-infra reusable CI
workflow. MegaLinter MUST lint at minimum:
- Go source files (via golangci-lint)
- GitHub Actions workflow files (via actionlint)
- Bash scripts (via shellcheck)
- Repository secrets (via gitleaks)

The `.mega-linter.yml` configuration file MUST exist at
the repository root. MegaLinter MUST respect the
existing `.golangci.yml` configuration for Go linting.

#### Scenario: Invalid GitHub Actions workflow syntax

- **GIVEN** a pull request that modifies a workflow file
  with invalid syntax
- **WHEN** the CI checks workflow runs
- **THEN** actionlint detects the issue and the workflow
  fails

#### Scenario: Go lint violation in changed file

- **GIVEN** a pull request that introduces an unused
  variable in a Go file
- **WHEN** MegaLinter runs golangci-lint
- **THEN** the unused variable is detected and the
  workflow fails

### Requirement: PR Title Validation

Pull requests authored by humans (not `dependabot[bot]`)
MUST have their title validated against the conventional
commit format. A `commitlint.config.js` file MUST exist
at the repository root extending
`@commitlint/config-conventional`.

#### Scenario: Non-conventional PR title

- **GIVEN** a pull request with title "Update the thing"
- **WHEN** the CI checks workflow runs commitlint
- **THEN** the validation fails with an error indicating
  the expected format

#### Scenario: Dependabot PR title

- **GIVEN** a pull request authored by `dependabot[bot]`
- **WHEN** the CI checks workflow runs
- **THEN** commitlint validation is skipped

### Requirement: Workflow Standards

All CI/CD workflow files MUST include:
- SPDX license header (`# SPDX-License-Identifier: Apache-2.0`)
- Explicit `permissions:` block at the top level with
  restrictive defaults
- Job-level permission escalation only where needed
- Concurrency group (`${{ github.workflow }}-${{ github.ref }}`)
  with `cancel-in-progress: true` for PR-triggered workflows
- SHA-pinned references for all reusable workflows and
  actions, with version comment

#### Scenario: Multiple pushes to PR branch

- **GIVEN** a pull request with a running CI workflow
- **WHEN** a new commit is pushed to the same PR branch
- **THEN** the in-progress workflow run is cancelled and
  a new run starts for the latest commit

## MODIFIED Requirements

### Requirement: Test Workflow Scope

Previously: `test.yml` contained build, vet, lint,
govulncheck, test, and coverage ratchet enforcement in
a single monolithic job named "Build & Test".

The workflow MUST be renamed to `ci_local.yml` and MUST
contain only project-specific checks that cannot be
delegated to org-infra reusable workflows:
- `go build ./...`
- `go test -race -count=1 -coverprofile=coverage.out ./...`
- Coverage ratchet enforcement

The workflow MUST NOT include linting steps (delegated
to ci_checks via MegaLinter) or vulnerability scanning
(delegated to ci_security).

### Requirement: Branch Protection Status Checks

Previously: `.github/settings.yml` required status check
context was `"Build & Test"`.

The required status check context MUST be updated to
match the renamed job name in `ci_local.yml`.

## REMOVED Requirements

### Requirement: Non-blocking Vulnerability Scanning

The `govulncheck ./... || true` step is removed. Non-blocking
vulnerability scanning provides false confidence. It is
replaced by properly blocking OSV-Scanner and Trivy scans
in `ci_security.yml`.

### Requirement: Inline Lint Steps in Test Workflow

The `go vet ./...` and `golangci-lint run` steps are removed
from the test workflow. Go linting is now handled by
MegaLinter in `ci_checks.yml`, which uses the same
`.golangci.yml` configuration (and golangci-lint includes
govet as an enabled linter).
