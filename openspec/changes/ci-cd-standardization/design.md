## Context

unbound-force has 3 CI/CD workflows. The complytime
organization's `org-infra` repository provides 13 reusable
workflows, of which 9 are applicable to non-container
projects. complyctl (the reference consumer) uses all 9.
unbound-force uses only 1 (`reusable_crapload_analysis`).

The existing `test.yml` is a monolithic workflow that
combines build, vet, lint, test, vulnerability scanning,
and coverage ratchets into a single job. This violates
separation of concerns and makes failures harder to
diagnose. The vulnerability check (`govulncheck || true`)
is non-blocking, providing false confidence.

The `ci_crapload.yml` is the only workflow that follows
the org-infra standard (SHA-pinned, SPDX header, proper
permissions). It serves as the local reference pattern.

## Goals / Non-Goals

### Goals
- Consume 4 additional org-infra reusable workflows
  (CI, security, scheduled, dependencies) using the
  established consumer pattern
- Refactor `test.yml` into `ci_local.yml` focused on
  build + test + coverage (what reusable workflows
  cannot cover)
- Standardize all workflows with permissions,
  concurrency, SPDX headers, and SHA-pinned references
- Add MegaLinter and commitlint configuration files
- Complete the dependabot pipeline (review + auto-approve)

### Non-Goals
- Container image workflows (no Containerfile exists)
- SonarCloud/SonarQube integration (requires org setup)
- Compliance scanning (complyctl-specific workflow)
- CRAP Load main-push metrics publishing (separate change)
- Custom MegaLinter rules or extensive linter tuning
- Pre-commit hook configuration

## Decisions

### D1: Follow org-infra consumer pattern exactly

All new consumer workflows follow the structure
established by `ci_crapload.yml` and by org-infra's own
consumer workflows:

```yaml
# SPDX-License-Identifier: Apache-2.0
#
# <Workflow Name>
# ===============
# <Purpose description>

name: <Name>

on:
  <triggers>

permissions:
  <restrictive-top-level>

jobs:
  <job>:
    name: <Display Name>
    permissions:
      <job-level-escalation>
    uses: complytime/org-infra/...@<sha> # <version>
```

**Rationale**: Consistency with the org standard. The
`ci_crapload.yml` already follows this pattern in this
repo, and org-infra's own consumer workflows demonstrate
the canonical form. SHA pinning (not `@main`) ensures
reproducible builds.

### D2: SHA-pin to v0.1.0

All reusable workflow references pin to SHA
`baf5b2e21e61581b4a3a129795286e8592e6afbb` with comment
`# v0.1.0`. This matches what `ci_crapload.yml` already
uses and what org-infra uses in its own consumer workflows.

**Rationale**: `@main` references are non-deterministic.
A breaking change to a reusable workflow would silently
break CI. SHA pinning provides the same supply chain
security as SHA-pinned actions.

### D3: MegaLinter linter selection

Start with a focused set of 4 linters:

| Linter | Purpose |
|---|---|
| `GO_GOLANGCI_LINT` | Go linting (uses `.golangci.yml`) |
| `ACTION_ACTIONLINT` | GitHub Actions workflow validation |
| `BASH_SHELLCHECK` | Bash script quality |
| `REPOSITORY_GITLEAKS` | Secret detection |

Markdown and YAML linters are commented out initially.
The repo has 100+ Markdown files (specs, agents,
commands) and many YAML configs. Enabling these linters
without tuning would produce excessive noise.

**Rationale**: Match complyctl's approach (they also
have `MARKDOWN_MARKDOWNLINT` and `YAML_YAMLLINT`
commented out). Start clean, enable incrementally.

### D4: Rename test.yml to ci_local.yml

The refactored workflow keeps only what reusable
workflows cannot provide:

| Kept | Reason |
|---|---|
| `go build ./...` | Project-specific build verification |
| `go test -race -count=1 -coverprofile=coverage.out ./...` | Project-specific test execution |
| Coverage ratchet enforcement | Project-specific thresholds |

| Removed | Replacement |
|---|---|
| `go vet ./...` | MegaLinter (golangci-lint includes govet) |
| `golangci-lint run` | MegaLinter (`GO_GOLANGCI_LINT`) |
| `govulncheck ./... \|\| true` | `ci_security.yml` (OSV + Trivy, blocking) |

**Rationale**: Separation of concerns. Local checks
(build, test, coverage) are inherently project-specific.
Linting and security scanning are org-standardized
concerns that benefit from the reusable workflow pattern.

### D5: Concurrency groups

All PR-triggered workflows get cancel-in-progress
concurrency:

```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true
```

Release workflow gets non-cancelling concurrency:

```yaml
concurrency:
  group: release-${{ github.ref }}
  cancel-in-progress: false
```

**Rationale**: Pushing multiple commits to a PR branch
triggers N parallel runs. Cancel-in-progress ensures
only the latest commit is tested, saving runner minutes.
Release workflows must never be cancelled mid-execution.

### D6: Dependabot auto-approval criteria

Matches the org standard from complyctl:
- Risk level is not `high` (non-major updates)
- Dependency review passes (no known vulnerabilities)
- Release age >= 24 hours (avoids supply chain attacks)
- Release age is known (not `-1`)

All four criteria must be met. If any fails, the PR
requires manual review.

**Rationale**: The 24-hour release age window is a
supply chain defense. Newly-published malicious packages
are typically detected within hours. Requiring the
package to have been published for at least a day
significantly reduces risk.

### D7: Security workflow replaces govulncheck

The existing `govulncheck || true` step is replaced by
three complementary scanners:

| Scanner | Coverage |
|---|---|
| OSV-Scanner | Known CVE database (Go, GitHub Actions) |
| Trivy source | Secrets, misconfigs, IaC issues |
| OpenSSF Scorecards | Supply chain security posture |

All three upload SARIF to GitHub's Security tab,
providing a unified security dashboard.

**Rationale**: `govulncheck` only covers Go stdlib and
direct dependency vulnerabilities. OSV-Scanner covers
a broader CVE database. Trivy adds secret and
misconfiguration scanning. Scorecards track the project's
overall supply chain security posture. The `|| true`
on govulncheck means it was never actually blocking --
these replacements are properly blocking.

## Risks / Trade-offs

### R1: MegaLinter may find pre-existing issues

MegaLinter running golangci-lint may surface findings
that the inline `golangci-lint run` was already
producing (and being ignored or not noticed). The
existing `.golangci.yml` config will be respected, so
the linter set and rules are identical.

**Mitigation**: `VALIDATE_ALL_CODEBASE` is `true` only
on push to main. PR runs are incremental (changed files
only). Pre-existing issues in unchanged files will not
block PRs.

### R2: Branch protection update requires coordination

Renaming the `test.yml` job changes the required status
check context from `"Build & Test"` to the new name.
This must be updated in `.github/settings.yml`
simultaneously, or PRs will be blocked by a check that
no longer exists.

**Mitigation**: Update `settings.yml` in the same
commit as the workflow rename.

### R3: Reusable workflow version lag

SHA-pinning to `v0.1.0` means the project won't
automatically receive updates to org-infra reusable
workflows. When org-infra releases a new version, the
SHA references must be manually updated.

**Mitigation**: Dependabot is not configured to update
reusable workflow references (only action SHAs). This
is acceptable for v1 -- manual updates via a dedicated
PR when org-infra releases. A future improvement could
add automation for this.

### R4: Commitlint may reject existing PR title patterns

If contributors are not using conventional commit
format for PR titles, commitlint will fail. The
constitution already mandates conventional commits,
so this enforces an existing rule rather than
introducing a new one. Dependabot PRs are exempt
(the reusable workflow skips validation for
`dependabot[bot]`).

**Mitigation**: `subject-case` rule is disabled
(level 0), matching complyctl's config. Only the
type prefix (`feat:`, `fix:`, etc.) is validated.
