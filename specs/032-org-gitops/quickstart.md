# Quickstart: GitHub Org GitOps

**Branch**: `032-org-gitops` | **Date**: 2026-04-18
**Spec**: [spec.md](spec.md)

## Prerequisites

1. **GitHub org admin access** on `unbound-force`
2. **Go toolchain** (for installing Peribolos locally)
3. **GitHub CLI** (`gh`) authenticated with org admin
   permissions

---

## Step 1: Register the GitHub App

1. Navigate to:
   `github.com/organizations/unbound-force/settings/apps`

2. Click **New GitHub App** with these settings:
   - **Name**: `unbound-force-peribolos`
   - **Homepage URL**: `https://github.com/unbound-force`
   - **Webhook**: Deactivate (uncheck "Active")
   - **Permissions**:
     - Organization Members: Read & Write
     - Organization Administration: Read & Write
   - **Where can this app be installed**: Only on this
     account

3. After creation:
   - Note the **App ID** from the app settings page
   - Click **Generate a private key** (downloads `.pem`)
   - Click **Install App** → install on `unbound-force`

4. Store secrets (after `.github` repo is created):
   ```bash
   gh secret set APP_ID --repo unbound-force/.github
   gh secret set APP_PRIVATE_KEY < path/to/private-key.pem --repo unbound-force/.github
   ```

---

## Step 2: Create the `.github` Repo

```bash
gh repo create unbound-force/.github \
  --public \
  --description "Organization-wide GitHub configuration" \
  --clone
cd .github
```

---

## Step 3: Seed Org Config with Peribolos

```bash
# Install Peribolos
go install github.com/uwu-tools/peribolos@latest

# Dump current org state
peribolos --dump unbound-force \
  --github-token-path <(gh auth token) \
  > org/config.yaml

# Review and clean up the dump:
# - Remove billing_email if present
# - Verify member list matches expectations
# - Verify team structure
```

---

## Step 4: Create Org-Wide Settings

Create `settings.yml` in the repo root:

```yaml
# Org-wide repository defaults
# Applied by the Repository Settings App (github.com/apps/settings)
# Per-repo overrides use _extends: .github

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
  - name: documentation
    color: 0075ca
    description: Improvements or additions to documentation
  - name: duplicate
    color: cfd3d7
    description: This issue or pull request already exists
  - name: enhancement
    color: a2eeef
    description: New feature or request
  - name: good first issue
    color: 7057ff
    description: Good for newcomers
  - name: help wanted
    color: 008672
    description: Extra attention is needed
  - name: invalid
    color: e4e669
    description: This doesn't seem right
  - name: question
    color: d876e3
    description: Further information is requested
  - name: wontfix
    color: ffffff
    description: This will not be worked on

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

---

## Step 5: Create the Peribolos Sync Workflow

Create `.github/workflows/peribolos-sync.yml`:

```yaml
name: Org Sync

on:
  push:
    branches: [main]
    paths:
      - 'org/**'
  workflow_dispatch:

permissions:
  contents: read

jobs:
  sync:
    name: Sync Org Config
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683  # v4.2.2

      - name: Generate GitHub App Token
        id: app-token
        uses: actions/create-github-app-token@v1
        with:
          app-id: ${{ secrets.APP_ID }}
          private-key: ${{ secrets.APP_PRIVATE_KEY }}
          owner: unbound-force

      - name: Set up Go
        uses: actions/setup-go@40f1582b2485089dde7abd97c1529aa768e1baff  # v5.6.0
        with:
          go-version: '1.24'

      - name: Install Peribolos
        run: go install github.com/uwu-tools/peribolos@latest

      - name: Run Peribolos
        env:
          GITHUB_TOKEN: ${{ steps.app-token.outputs.token }}
        run: |
          peribolos \
            --config-path org/config.yaml \
            --required-admins=jflowers \
            --min-admins=2 \
            --maximum-removal-delta=0.25 \
            --confirm
```

**Note**: The `actions/create-github-app-token` action
should be SHA-pinned before final implementation. The
exact SHA will be resolved during implementation.

---

## Step 6: Create CODEOWNERS

Create `CODEOWNERS` in the repo root:

```
# Org config and settings require admin review
settings.yml    @jflowers @marcusburghardt
org/            @jflowers @marcusburghardt
.github/        @jflowers @marcusburghardt
```

---

## Step 7: Create Org Profile

Create `profile/README.md`:

```markdown
# Unbound Force

AI agent personas and roles for a software agent swarm,
themed as a superhero team. Each hero is a specialized
tool — tester, reviewer, developer, product owner,
manager — that works independently or as part of a
coordinated swarm.

## Heroes

| Hero | Role | Repo |
|------|------|------|
| Cobalt-Crush | Developer | [unbound-force](https://github.com/unbound-force/unbound-force) |
| Gaze | Tester | [gaze](https://github.com/unbound-force/gaze) |
| The Divisor | Reviewer | [unbound-force](https://github.com/unbound-force/unbound-force) |
| Muti-Mind | Product Owner | [unbound-force](https://github.com/unbound-force/unbound-force) |
| Mx F | Manager | [unbound-force](https://github.com/unbound-force/unbound-force) |

## Tools

| Tool | Purpose | Repo |
|------|---------|------|
| Dewey | Knowledge graph | [dewey](https://github.com/unbound-force/dewey) |
| Replicator | Agent orchestration | [replicator](https://github.com/unbound-force/replicator) |

## Getting Started

```bash
brew install unbound-force/tap/unbound-force
uf init
uf setup
```
```

---

## Step 8: Install Repository Settings App

1. Navigate to: `github.com/apps/settings`
2. Click **Install**
3. Select the `unbound-force` organization
4. Grant access to **All repositories**

---

## Step 9: Create Per-Repo Settings Overrides

For each repo that needs status checks or extra labels,
create `.github/settings.yml` with `_extends: .github`
and the repo-specific overrides. See data-model.md for
the concrete YAML for each repo.

---

## Step 10: Verify

```bash
# Verify Peribolos seed accuracy (dry-run, expect zero changes)
peribolos \
  --config-path org/config.yaml \
  --github-token-path <(gh auth token) \
  --required-admins=jflowers \
  --min-admins=2

# Verify branch protection is applied
for repo in unbound-force gaze website dewey replicator homebrew-tap containerfile; do
  echo "=== $repo ==="
  gh api repos/unbound-force/$repo/branches/main/protection \
    --jq '{required_reviews: .required_pull_request_reviews.required_approving_review_count, enforce_admins: .enforce_admins.enabled}' \
    2>&1
done

# Verify labels
gh api repos/unbound-force/unbound-force/labels \
  --jq '.[].name' | sort
```
<!-- scaffolded by uf vdev -->
