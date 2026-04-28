## 1. Configuration Files

- [x] 1.1 Create `.mega-linter.yml` at repository root
  with `ENABLE_LINTERS` set to `GO_GOLANGCI_LINT`,
  `ACTION_ACTIONLINT`, `BASH_SHELLCHECK`,
  `REPOSITORY_GITLEAKS`. Exclude `vendor/` directory.
  Comment out `MARKDOWN_MARKDOWNLINT` and
  `YAML_YAMLLINT` for future enablement.
- [x] 1.2 Create `commitlint.config.js` at repository
  root extending `@commitlint/config-conventional` with
  `subject-case` rule disabled (level 0).

## 2. New Consumer Workflows

- [x] 2.1 Create `.github/workflows/ci_checks.yml`
  consuming `reusable_ci.yml@baf5b2e...# v0.1.0`.
  Triggers: push to main, PR to main. Top-level
  permissions: `contents: read`, `issues: none`,
  `pull-requests: none`. Job-level: `contents: read`,
  `issues: read`. Include SPDX header, comment block,
  concurrency group with cancel-in-progress.
- [x] 2.2 Create `.github/workflows/ci_security.yml`
  consuming `reusable_vuln_scan.yml@baf5b2e...# v0.1.0`
  (with `enable_trivy_source: true`) and
  `reusable_security.yml@baf5b2e...# v0.1.0`.
  Triggers: push to main, PR to main. Top-level
  permissions: restrictive (all none/read). Job-level:
  escalate per-job for SARIF upload and OIDC. Include
  SPDX header, comment block, concurrency group.
- [x] 2.3 Create `.github/workflows/ci_scheduled.yml`
  consuming `reusable_scheduled.yml@baf5b2e...# v0.1.0`.
  Trigger: schedule daily at midnight UTC
  (`cron: '0 0 * * *'`). Top-level permissions:
  restrictive. Job-level: `contents: read`,
  `actions: read`, `security-events: write`,
  `id-token: write`. Include SPDX header, comment block.
- [x] 2.4 Create `.github/workflows/ci_dependencies.yml`
  consuming `reusable_deps_reviewer.yml@baf5b2e...# v0.1.0`
  and `reusable_dependabot_reviewer.yml@baf5b2e...# v0.1.0`.
  Triggers: push to main, PR to main. Include 4 jobs:
  general dep review, dependabot review, conditional PR
  comment (dependabot[bot] only), conditional auto-approve
  (dependabot[bot] only, all 4 criteria). Set
  `MIN_RELEASE_AGE_HOURS: 24` env var. Include SPDX
  header, comment block, concurrency group. Use
  `peter-evans/create-or-update-comment@e8674b075228eee787fea43ef493e45ece1004c9`
  (v5.0.0) for structured comment and
  `actions/github-script@3a2844b7e9c422d3c10d287c895573f7108da1b3`
  (v9.0.0) for auto-approve. Auto-approve uses
  `github.rest.pulls.createReview` with `event: 'APPROVE'`
  (review approval, not merge). Auto-approve job requires
  `pull-requests: write` at job level.

## 3. Refactor Existing Workflows

- [x] 3.1 Rename `.github/workflows/test.yml` to
  `.github/workflows/ci_local.yml`. Update workflow
  `name:` field to `Local CI`. Update job name.
- [x] 3.2 Remove `go vet ./...` step from ci_local.yml
  (covered by MegaLinter via golangci-lint govet linter).
- [x] 3.3 Remove golangci-lint install and run steps
  from ci_local.yml (covered by MegaLinter
  `GO_GOLANGCI_LINT`).
- [x] 3.4 Remove govulncheck install and run steps from
  ci_local.yml (replaced by ci_security.yml with
  OSV-Scanner + Trivy).
- [x] 3.5 Add SPDX header and descriptive comment block
  to ci_local.yml.
- [x] 3.6 Add explicit `permissions:` block to
  ci_local.yml (`contents: read` top-level).
- [x] 3.7 Add concurrency group to ci_local.yml
  (`${{ github.workflow }}-${{ github.ref }}`,
  `cancel-in-progress: true`).

## 4. Standardize Existing Workflows

- [x] 4.1 Add SPDX header and descriptive comment block
  to `release.yml`.
- [x] 4.2 Add top-level `permissions: {}` (empty/none)
  to `release.yml` above the jobs section. Add explicit
  job-level permissions: `release` job needs
  `contents: write`; `sign-macos` job needs
  `contents: write`.
- [x] 4.3 Add concurrency group to `release.yml`
  (`release-${{ github.ref }}`,
  `cancel-in-progress: false`).
- [x] 4.4 Add concurrency group to `ci_crapload.yml`
  (`${{ github.workflow }}-${{ github.ref }}`,
  `cancel-in-progress: true`).

## 5. Branch Protection Update

- [x] 5.1 Update `.github/settings.yml`
  `required_status_checks.contexts` to use the new job
  name from ci_local.yml instead of `"Build & Test"`.

## 6. Documentation Update

- [x] 6.1 Update AGENTS.md "Build & Test Commands"
  section to document the new CI workflow structure
  (ci_local for build+test, ci_checks for linting,
  ci_security for vulnerability scanning,
  ci_dependencies for dependency review,
  ci_scheduled for daily scans).
- [x] 6.2 Update AGENTS.md "Gatekeeping Value Protection"
  item 4 to replace `govulncheck` with `OSV-Scanner` and
  `Trivy` as the protected vulnerability scanning tools.
  This gate change is authorized by human decision (the
  existing `govulncheck || true` was non-blocking and
  never served as a functional gate).

## 7. Verification

- [x] 7.1 Verify all workflow files have SPDX headers.
- [x] 7.2 Verify all reusable workflow references are
  SHA-pinned to `baf5b2e...# v0.1.0`.
- [x] 7.3 Verify all workflow files have explicit
  `permissions:` blocks.
- [x] 7.4 Verify all PR-triggered workflows have
  concurrency groups with cancel-in-progress.
- [x] 7.5 Verify ci_local.yml no longer contains lint
  or vuln-check steps.
- [x] 7.6 Verify `.mega-linter.yml` excludes `vendor/`.
- [x] 7.7 Verify `commitlint.config.js` extends
  `@commitlint/config-conventional`.
- [x] 7.8 Verify constitution alignment: Composability
  (each workflow independently deployable), Observable
  Quality (SARIF uploads, structured outputs), Testability
  (test flags `-race -count=1` retained in ci_local.yml).
<!-- spec-review: passed -->
<!-- code-review: passed -->
