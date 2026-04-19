# Research: GitHub Org GitOps

**Branch**: `032-org-gitops` | **Date**: 2026-04-18
**Spec**: [spec.md](spec.md)

## 1. uwu-tools/peribolos

### Overview

Peribolos is a standalone Go binary (Apache 2.0) forked
from the Kubernetes `prow/cmd/peribolos` tool. It
manages GitHub organization settings, member rosters,
and team definitions declaratively via YAML.

- **Repository**: github.com/uwu-tools/peribolos
- **Latest tag**: v0.1.0
- **Language**: Go
- **License**: Apache 2.0
- **Stars**: 24 (niche but actively maintained)
- **Origin**: Refactored from kubernetes/test-infra's
  peribolos; used at scale by the Kubernetes project

### Installation

```bash
go install github.com/uwu-tools/peribolos@latest
```

No pre-built binaries are published (no GitHub Releases
with assets). Installation requires a Go toolchain.
The workflow will use `go install` in the CI runner.

### Key Commands

| Command | Purpose |
|---------|---------|
| `peribolos --dump <org>` | Seed current org state to YAML |
| `peribolos --config-path <file>` | Dry-run: show proposed changes |
| `peribolos --config-path <file> --confirm` | Apply changes to GitHub |

### YAML Format

The config uses a top-level `orgs` key with nested org
names. Each org contains:

```yaml
orgs:
  unbound-force:
    # Org settings
    default_repository_permission: read
    members_can_create_repositories: true
    has_organization_projects: true
    has_repository_projects: true

    # Member roster
    admins:
      - jflowers
      - marcusburghardt
    members:
      - gvauter
      - hbraswelrh
      - jpower432
      - tyraziel

    # Team definitions
    teams:
      overlords:
        description: ""
        privacy: closed
        members:
          - gvauter
          - jpower432
          - hbraswelrh
        maintainers:
          - jflowers
          - marcusburghardt
        repos:
          unbound-force: write
          gaze: write
          website: write
          dewey: write
          replicator: write
          homebrew-tap: write
          containerfile: write
```

**Key behaviors**:
- Fields missing from config are NOT managed (safe
  partial configs)
- `--dump` outputs the current state for seeding
- Teams use `members` (regular) and `maintainers`
  (team admins) distinction
- Repo permissions are set per-team under `repos:`

### Safety Guards

| Flag | Default | Purpose |
|------|---------|---------|
| `--required-admins` | empty | Users who MUST be admins |
| `--min-admins` | 5 | Minimum admin count |
| `--maximum-removal-delta` | 0.25 | Max % of members to remove |
| `--require-self` | true | Bot must be admin |
| `--confirm` | false | No mutations without this |

For our org (6 members, 2 admins), we set:
- `--required-admins=jflowers` (protect primary admin)
- `--min-admins=2` (match current admin count)
- `--maximum-removal-delta=0.25` (prevent mass removal)

### Authentication

Peribolos accepts authentication via:
1. `--github-token-path` — path to a file containing a
   token
2. Environment variable `GITHUB_TOKEN`

For CI, we generate a short-lived installation token
from a registered GitHub App using the
`actions/create-github-app-token` action.

### Resolved Unknown: Peribolos + GitHub App Tokens

**Question**: Can Peribolos use GitHub App installation
tokens (short-lived, scoped) instead of PATs?

**Answer**: Yes. Peribolos accepts any valid GitHub token
via `--github-token-path` or `GITHUB_TOKEN`. GitHub App
installation tokens are standard OAuth tokens that work
with all GitHub API endpoints. The workflow generates
a token using `actions/create-github-app-token` and
passes it to Peribolos.

**Required App permissions**:
- Organization Members: Read & Write
- Organization Administration: Read & Write

---

## 2. Repository Settings App

### Overview

The Repository Settings App is a hosted GitHub App
(Probot-based) that syncs repository settings from a
`.github/settings.yml` file to GitHub. It is installed
at the org level — zero infrastructure needed.

- **App URL**: github.com/apps/settings
- **Source**: github.com/repository-settings/app
- **Maintained by**: repository-settings org
- **Powered by**: Probot framework, hosted on Vercel

### How It Works

1. Install the app on the org (or specific repos)
2. Create `.github/settings.yml` in a repo
3. On push to the default branch, the app reads the
   YAML and applies settings via the GitHub API
4. Changes are auditable in git history

### Settings YAML Format

```yaml
# .github/settings.yml (org-wide defaults in .github repo)
repository:
  delete_branch_on_merge: true
  allow_squash_merge: true
  allow_merge_commit: true
  allow_rebase_merge: true
  enable_automated_security_fixes: true
  enable_vulnerability_alerts: true

labels:
  - name: bug
    color: d73a4a
    description: Something isn't working
  - name: enhancement
    color: a2eeef
    description: New feature or request
  # ... more labels

branches:
  - name: main
    protection:
      required_pull_request_reviews:
        required_approving_review_count: 1
        dismiss_stale_reviews: true
        require_code_owner_reviews: false
      required_status_checks:
        strict: true
        contexts: []
      enforce_admins: false
      restrictions: null
```

### Inheritance via `_extends`

Per-repo `.github/settings.yml` files can inherit from
the org-wide `.github` repo:

```yaml
_extends: .github
# Override only what differs
branches:
  - name: main
    protection:
      required_status_checks:
        strict: true
        contexts:
          - "unit-and-integration"
          - "e2e"
```

Individual settings in `labels`, `teams`, and `branches`
arrays are merged by `name` with the base repo. This
means a per-repo override only needs to specify the
branches or labels that differ from the org default.

### Branch Protection API

The Settings App uses the **legacy branch protection
API** (`PUT /repos/{owner}/{repo}/branches/{branch}/protection`),
NOT the newer repository rulesets API. This is important
because:

- Legacy branch protection works on **GitHub Free**
- Repository rulesets also work on Free (repo-level)
  but org-level rulesets require Team plan
- The Settings App does not support rulesets

### Security Warning

The Settings App escalates anyone with `push` access to
effective admin, since they can modify `settings.yml`.
Mitigation: use CODEOWNERS to restrict who can approve
changes to `.github/settings.yml`.

### Resolved Unknown: Settings App + Free Plan

**Question**: Does the Repository Settings App work on
GitHub Free organizations?

**Answer**: Yes. The app uses the legacy branch
protection API which is available on all plans. The
only limitation is that org-level rulesets (a different
feature) require the Team plan — but the Settings App
does not use rulesets.

---

## 3. GitHub App Registration

### Resolved Unknown: How to Register a Custom GitHub App

**Process**:
1. Go to github.com/organizations/unbound-force/settings/apps
2. Click "New GitHub App"
3. Configure:
   - Name: `unbound-force-peribolos` (or similar)
   - Homepage URL: `https://github.com/unbound-force`
   - Webhook: Deactivate (not needed for CI-only use)
   - Permissions:
     - Organization Members: Read & Write
     - Organization Administration: Read & Write
   - Where can this app be installed: Only on this
     account
4. Generate a private key (downloads `.pem` file)
5. Note the App ID from the app settings page
6. Install the app on the `unbound-force` organization
7. Store secrets in the `.github` repo:
   - `APP_ID`: The numeric app ID
   - `APP_PRIVATE_KEY`: The `.pem` file contents

### Token Generation in CI

Use `actions/create-github-app-token` (official GitHub
action) to generate short-lived installation tokens:

```yaml
- name: Generate token
  id: app-token
  uses: actions/create-github-app-token@v1
  with:
    app-id: ${{ secrets.APP_ID }}
    private-key: ${{ secrets.APP_PRIVATE_KEY }}
    owner: unbound-force

- name: Run Peribolos
  env:
    GITHUB_TOKEN: ${{ steps.app-token.outputs.token }}
  run: peribolos --config-path org/config.yaml --confirm
```

**Benefits over PAT**:
- Tokens expire after 1 hour (automatic rotation)
- Scoped to specific permissions (least privilege)
- Not tied to a personal account (survives member
  departure)
- Audit log shows app identity, not personal identity

---

## 4. Current Org State (Verified)

### Members (6 total)

| Username | Role | Verified |
|----------|------|----------|
| jflowers | admin | ✓ (API) |
| marcusburghardt | admin | ✓ (API) |
| gvauter | member | ✓ (API) |
| hbraswelrh | member | ✓ (API) |
| jpower432 | member | ✓ (API) |
| tyraziel | member | ✓ (API) |

### Teams (1 total)

| Team | Privacy | Members | Repos |
|------|---------|---------|-------|
| overlords | closed | jflowers, marcusburghardt, gvauter, jpower432, hbraswelrh (5) | All 7 repos (write) |

Note: `tyraziel` is NOT in the overlords team (only 5
of 6 members). The spec states "5 members" which is
confirmed.

### Org Settings (Verified via API)

| Setting | Value |
|---------|-------|
| Plan | Free |
| default_repository_permission | read |
| members_can_create_repositories | true |
| two_factor_requirement_enabled | false |

### Repositories (7 total)

| Repo | Default Branch | CI Workflows | PR-Triggered Jobs |
|------|---------------|-------------|-------------------|
| unbound-force | main | test.yml, release.yml | `test` (Build & Test) |
| gaze | main | test.yml, mega-linter.yml, release.yml | `unit-and-integration`, `e2e` |
| website | main | deploy-gh-pages.yml | `build` (push-only, not PR) |
| dewey | main | ci.yml, mega-linter.yml, release.yml | `build-and-test` |
| replicator | main | ci.yml, release.yml | `test` |
| homebrew-tap | main | (none) | (none) |
| containerfile | main | build-base.yml, build-push.yml | (none — schedule/dispatch only) |

### Required Status Checks per Repo

Based on CI workflow analysis:

| Repo | Required Checks | Notes |
|------|----------------|-------|
| unbound-force | `test` | Single job: build, vet, lint, test, vulncheck |
| gaze | `unit-and-integration`, `e2e` | Matrix build (Go 1.24, 1.25) + e2e |
| dewey | `build-and-test` | Single job |
| replicator | `test` | Single job |
| website | `build` | Deploy workflow; build job runs on push to main |
| homebrew-tap | (none) | No CI workflows |
| containerfile | (none) | Schedule/dispatch only, not PR-triggered |

**Important**: The `website` deploy workflow triggers on
push to main only, NOT on pull requests. The `build` job
name exists but is not PR-triggered. Branch protection
for website should either: (a) not require status checks,
or (b) add a separate PR-triggered CI workflow. The spec
says `build` — we'll configure it but note it may not
block PRs until a PR-triggered workflow is added.

---

## 5. SHA-Pinned Actions Convention

The project convention (per AGENTS.md and existing
workflows) requires all GitHub Actions to use SHA-pinned
versions, not tags. Examples from existing workflows:

```yaml
# Correct (SHA-pinned with version comment)
uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683  # v4.2.2
uses: actions/setup-go@40f1582b2485089dde7abd97c1529aa768e1baff  # v5.6.0

# Incorrect (tag-based)
uses: actions/checkout@v4
```

The Peribolos sync workflow MUST follow this convention.

---

## 6. Risk Assessment

### Low Risk
- Peribolos dry-run mode prevents accidental changes
- Safety guards (delta, required admins) provide
  backstops
- Repository Settings App is additive for labels
  (unlisted labels are not deleted)

### Medium Risk
- **Settings App security escalation**: Anyone with push
  access can modify `settings.yml` and effectively gain
  admin. Mitigated by CODEOWNERS requiring admin review.
- **Website CI gap**: The website repo's deploy workflow
  is not PR-triggered. Branch protection with required
  status checks may not block PRs as intended until a
  PR-triggered workflow is added.

### Decisions Made
- **GitHub App over PAT**: Short-lived tokens, not tied
  to personal accounts, automatic rotation, audit trail.
- **Legacy branch protection over rulesets**: Required
  by the Settings App; works on Free plan.
- **Separate tools for org vs repo**: Peribolos for org
  (members, teams, settings), Settings App for repo
  (settings, labels, branch protection). Clear
  separation of concerns.
<!-- scaffolded by uf vdev -->
