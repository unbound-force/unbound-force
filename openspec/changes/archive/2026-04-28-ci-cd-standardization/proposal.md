## Why

unbound-force currently has 3 CI/CD workflows (`test.yml`,
`release.yml`, `ci_crapload.yml`) while the complytime
organization's standard -- established in `org-infra` and
adopted by `complyctl` -- provides 13 reusable workflows
(9 applicable to non-container projects).
unbound-force consumes only 1 of them (CRAP load analysis).

Key gaps identified through analysis of `complyctl` and
`org-infra`:

- **Dependabot is configured but has no review pipeline.**
  PRs are created but receive no automated risk assessment,
  structured comments, or conditional auto-approval.
- **Security scanning is non-blocking.** `govulncheck` runs
  with `|| true`, meaning vulnerabilities never fail the
  build. No OSV-Scanner, no Trivy, no OpenSSF Scorecards.
- **No scheduled vulnerability monitoring.** Newly-disclosed
  CVEs after merge are never detected.
- **No multi-language linting.** YAML, Markdown, Bash, and
  GitHub Actions files are not linted. Go linting is embedded
  in a monolithic test workflow rather than using the org's
  MegaLinter integration.
- **Existing workflows lack standard hygiene:** missing
  permissions blocks, no concurrency controls, missing SPDX
  headers, non-deterministic tool versions.

## What Changes

Add 4 new consumer workflows that delegate to org-infra
reusable workflows, refactor the existing test workflow, and
standardize all workflows to match the org-infra pattern.

## Capabilities

### New Capabilities
- `ci_dependencies`: Dependency review on all PRs +
  dependabot-specific risk assessment, structured PR
  comments, and conditional auto-approval (non-major,
  24h+ release age, no vulnerabilities)
- `ci_security`: OSV-Scanner + Trivy source scan + OpenSSF
  Scorecards on push/PR to main, with SARIF upload to
  GitHub Security tab
- `ci_scheduled`: Daily OSV-Scanner + OpenSSF Scorecards
  for ongoing vulnerability monitoring
- `ci_checks`: MegaLinter (Go, GitHub Actions, Bash,
  Gitleaks) + conventional commit PR title validation via
  reusable CI workflow

### Modified Capabilities
- `test.yml` → `ci_local.yml`: Renamed and stripped of
  lint steps (moved to ci_checks) and govulncheck (replaced
  by ci_security). Retains build, test with `-race -count=1`,
  and coverage ratchet enforcement. Adds permissions,
  concurrency, SPDX header.
- `release.yml`: Adds SPDX header, concurrency group, and
  explicit top-level permissions block
- `ci_crapload.yml`: Adds concurrency group
- `.github/settings.yml`: Updates required_status_checks
  to reflect renamed job name

### Removed Capabilities
- `govulncheck || true` step: Removed from test workflow.
  Non-blocking vulnerability scanning is replaced by
  properly blocking OSV-Scanner + Trivy in ci_security.
  **Gatekeeping note**: AGENTS.md §Gatekeeping Value
  Protection lists `govulncheck` as a protected CI flag.
  However, the existing `govulncheck || true` was never
  a functional gate (the `|| true` ensured it always
  passed). This removal is authorized by human decision
  during the explore session. AGENTS.md item 4 will be
  updated to replace `govulncheck` with `OSV-Scanner`
  and `Trivy` as the protected vulnerability scanning
  tools.
- Inline golangci-lint step: Removed from test workflow.
  Go linting is handled by MegaLinter in ci_checks.

## Impact

**Files created:**
- `.github/workflows/ci_checks.yml`
- `.github/workflows/ci_security.yml`
- `.github/workflows/ci_scheduled.yml`
- `.github/workflows/ci_dependencies.yml`
- `.mega-linter.yml`
- `commitlint.config.js`

**Files modified:**
- `.github/workflows/test.yml` → renamed to `ci_local.yml`
- `.github/workflows/release.yml`
- `.github/workflows/ci_crapload.yml`
- `.github/settings.yml`
- `AGENTS.md` (Gatekeeping Value Protection item 4)

**External dependencies:**
- All new workflows consume reusable workflows from
  `complytime/org-infra` SHA-pinned to `v0.1.0`
  (`@baf5b2e21e61581b4a3a129795286e8592e6afbb`)
- MegaLinter requires `.mega-linter.yml` configuration
- Commitlint requires `commitlint.config.js` and
  `@commitlint/config-conventional` (installed at CI
  runtime by the reusable workflow)

**Branch protection:**
- `.github/settings.yml` required status check context
  changes from `"Build & Test"` to the new job name in
  `ci_local.yml`

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change adds CI/CD infrastructure workflows. It does
not affect inter-hero artifact communication, artifact
envelope formats, or hero-to-hero data exchange. All
artifacts produced (SARIF reports, scorecard results,
dependency review verdicts) are consumed by GitHub's own
infrastructure (Security tab, PR checks), not by heroes.

### II. Composability First

**Assessment**: PASS

Each new workflow is independently deployable and
removable. No workflow depends on another workflow being
present. The reusable workflow references are to an
external org-infra repository -- removing any single
consumer workflow does not affect the others. The
configuration files (`.mega-linter.yml`,
`commitlint.config.js`) are only consumed by their
respective workflows.

### III. Observable Quality

**Assessment**: PASS

This change directly strengthens observable quality:
- Security scan results are uploaded as SARIF to
  GitHub's machine-parseable Security tab
- OpenSSF Scorecards produce structured, comparable
  metrics across runs
- Dependency review produces structured outputs with
  risk classification
- CRAP analysis already produces machine-parseable JSON
- Coverage ratchets provide automated, reproducible
  quality evidence

### IV. Testability

**Assessment**: PASS

The refactored `ci_local.yml` retains `-race -count=1`
flags and coverage ratchet enforcement. No test
infrastructure is removed. The change does not affect
the project's test isolation properties. CI workflow
files themselves are validated by actionlint (via
MegaLinter). This change produces no new Go code, so
the traditional coverage strategy (unit/integration/e2e
with specific targets) is N/A. Verification is handled
through automated workflow syntax validation (actionlint)
and the manual verification checklist in tasks.md.
