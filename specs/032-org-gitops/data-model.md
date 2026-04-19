# Data Model: GitHub Org GitOps

**Branch**: `032-org-gitops` | **Date**: 2026-04-18
**Spec**: [spec.md](spec.md)

## Overview

This spec produces no Go source code and no runtime data
structures. The "data model" is the set of YAML
configuration files that define the desired state of the
GitHub organization and its repositories. These files
are the single source of truth for org governance.

---

## Entity 1: Org Config (Peribolos YAML)

**Location**: `.github` repo → `org/config.yaml`
**Format**: YAML (Peribolos schema)
**Managed by**: Peribolos CLI

### Schema

```yaml
orgs:
  <org-name>:
    # Org-level settings (all optional — omitted fields
    # are not managed)
    default_repository_permission: read | write | admin | none
    members_can_create_repositories: boolean
    has_organization_projects: boolean
    has_repository_projects: boolean
    company: string
    email: string
    name: string
    description: string
    location: string
    billing_email: string  # typically omitted (sensitive)

    # Member roster
    admins:
      - <username>    # org admin role
    members:
      - <username>    # org member role

    # Team definitions
    teams:
      <team-slug>:
        description: string
        privacy: closed | secret
        previously:
          - <old-team-name>  # rename support
        maintainers:
          - <username>       # team admin
        members:
          - <username>       # team member
        repos:
          <repo-name>: read | write | admin | maintain | triage
```

### Concrete Instance (Unbound Force)

```yaml
orgs:
  unbound-force:
    default_repository_permission: read
    members_can_create_repositories: true
    has_organization_projects: true
    has_repository_projects: true

    admins:
      - jflowers
      - marcusburghardt

    members:
      - gvauter
      - hbraswelrh
      - jpower432
      - tyraziel

    teams:
      overlords:
        description: ""
        privacy: closed
        maintainers:
          - jflowers
          - marcusburghardt
        members:
          - gvauter
          - jpower432
          - hbraswelrh
        repos:
          unbound-force: write
          gaze: write
          website: write
          dewey: write
          replicator: write
          homebrew-tap: write
          containerfile: write
```

### Validation Rules

- Every username in `admins` and `members` MUST be a
  valid GitHub user
- A user MUST NOT appear in both `admins` and `members`
  (Peribolos enforces this)
- Team `maintainers` and `members` MUST be org members
  or admins
- Repo names under `teams.*.repos` MUST exist in the org
- `privacy` MUST be `closed` or `secret`

---

## Entity 2: Org-Wide Repo Settings

**Location**: `.github` repo → `settings.yml`
**Format**: YAML (Repository Settings App schema)
**Managed by**: Repository Settings App (GitHub App)

### Schema

```yaml
repository:
  delete_branch_on_merge: boolean
  allow_squash_merge: boolean
  allow_merge_commit: boolean
  allow_rebase_merge: boolean
  enable_automated_security_fixes: boolean
  enable_vulnerability_alerts: boolean

labels:
  - name: string
    color: string (hex, no #)
    description: string

branches:
  - name: string
    protection:
      required_pull_request_reviews:
        required_approving_review_count: integer (1-6)
        dismiss_stale_reviews: boolean
        require_code_owner_reviews: boolean
        dismissal_restrictions:
          users: [string]
          teams: [string]
      required_status_checks:
        strict: boolean
        contexts: [string]
      enforce_admins: boolean
      restrictions: null | object
      required_linear_history: boolean
```

### Concrete Instance (Org Defaults)

```yaml
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
        require_code_owner_reviews: true
      required_status_checks:
        strict: true
        contexts: []
      enforce_admins: false
      restrictions: null
```

### Validation Rules

- Label `color` MUST be a valid 6-character hex string
  (no `#` prefix)
- `required_approving_review_count` MUST be 1-6
- All four top-level branch protection fields MUST be
  present (even if `null`) — the Settings App requires
  this
- `enforce_admins: false` allows admin emergency bypass
  (per FR-005)
- `contexts: []` in org defaults means no required
  status checks by default — repos override this

---

## Entity 3: Per-Repo Settings Override

**Location**: `<repo>/.github/settings.yml`
**Format**: YAML (Repository Settings App schema)
**Managed by**: Repository Settings App (GitHub App)

### Schema

```yaml
_extends: .github

# Override only what differs from org defaults
repository:
  # repo-specific settings (optional)

labels:
  # additional repo-specific labels (merged by name)

branches:
  - name: main
    protection:
      required_status_checks:
        strict: true
        contexts:
          - <job-name>  # repo-specific CI job names
```

### Concrete Instances

**unbound-force** (`.github/settings.yml`):
```yaml
_extends: .github

labels:
  - name: hero
    color: 6f42c1
    description: Hero-related work
  - name: phase-1
    color: 0e8a16
    description: Phase 1 foundation work
  - name: phase-2
    color: fbca04
    description: Phase 2 cross-cutting work

branches:
  - name: main
    protection:
      required_status_checks:
        strict: true
        contexts:
          - "Build & Test"
```

**gaze** (`.github/settings.yml`):
```yaml
_extends: .github

branches:
  - name: main
    protection:
      required_status_checks:
        strict: true
        contexts:
          - "Unit + Integration Tests (Go 1.24)"
          - "Unit + Integration Tests (Go 1.25)"
          - "e2e"
```

**dewey** (`.github/settings.yml`):
```yaml
_extends: .github

branches:
  - name: main
    protection:
      required_status_checks:
        strict: true
        contexts:
          - "build-and-test"
```

**replicator** (`.github/settings.yml`):
```yaml
_extends: .github

branches:
  - name: main
    protection:
      required_status_checks:
        strict: true
        contexts:
          - "test"
```

**website** (`.github/settings.yml`):
```yaml
_extends: .github

labels:
  - name: blog
    color: 1d76db
    description: Blog content
  - name: docs
    color: 0075ca
    description: Documentation updates
  - name: tutorial
    color: bfdadc
    description: Tutorial content

branches:
  - name: main
    protection:
      required_status_checks: null
```

**homebrew-tap** (`.github/settings.yml`):
```yaml
_extends: .github

branches:
  - name: main
    protection:
      required_status_checks: null
```

**containerfile** (`.github/settings.yml`):
```yaml
_extends: .github

branches:
  - name: main
    protection:
      required_status_checks: null
```

### Status Check Name Resolution

The `contexts` array uses the **job display name** (the
`name:` field in the workflow YAML), not the job key.
When a job uses a matrix strategy, GitHub creates
separate check runs with the matrix values appended.

| Repo | Job Key | Display Name (contexts value) |
|------|---------|-------------------------------|
| unbound-force | `test` | `Build & Test` |
| gaze | `unit-and-integration` | `Unit + Integration Tests (Go 1.24)`, `Unit + Integration Tests (Go 1.25)` |
| gaze | `e2e` | `e2e` (no custom name) |
| dewey | `build-and-test` | `build-and-test` (no custom name) |
| replicator | `test` | `test` (no custom name) |

**Note**: The unbound-force `test` job has `name: Build & Test`,
so the required status check context must use the
display name, not the job key.

**Note**: The gaze `unit-and-integration` job uses a
matrix strategy with Go versions 1.24 and 1.25. Each
matrix combination creates a separate status check.
Both must be listed as required contexts.

---

## Entity 4: Sync Workflow

**Location**: `.github` repo → `.github/workflows/peribolos-sync.yml`
**Format**: GitHub Actions YAML
**Managed by**: GitHub Actions

### Schema

```yaml
name: string
on:
  push:
    branches: [main]
    paths: [org/**]
  workflow_dispatch: {}

permissions:
  contents: read

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - checkout
      - generate GitHub App token
      - install peribolos
      - run peribolos --confirm
```

See quickstart.md for the complete workflow file.

---

## Entity 5: CODEOWNERS

**Location**: `.github` repo → `CODEOWNERS`
**Format**: GitHub CODEOWNERS syntax
**Managed by**: Git (enforced by branch protection)

### Schema

```
# Restrict sensitive config to org admins
settings.yml    @jflowers @marcusburghardt
org/            @jflowers @marcusburghardt
.github/        @jflowers @marcusburghardt
```

### Validation Rules

- Usernames MUST be valid GitHub users with org
  membership
- Paths MUST use CODEOWNERS glob syntax
- CODEOWNERS is only enforced when branch protection
  has `require_code_owner_reviews: true`

---

## Entity Relationships

```text
┌─────────────────────────────────────────────┐
│              .github repo                    │
│                                              │
│  org/config.yaml ──── Peribolos YAML         │
│       │                   │                  │
│       │              (applied by             │
│       │               sync workflow)         │
│       │                   │                  │
│  .github/workflows/ ─── peribolos-sync.yml   │
│                                              │
│  settings.yml ──── Org-wide defaults         │
│       │                   │                  │
│       │              (inherited by           │
│       │               per-repo overrides)    │
│       │                   │                  │
│  CODEOWNERS ──── Access control              │
│  profile/README.md ── Org profile            │
└─────────────────────────────────────────────┘
         │
         │ _extends: .github
         ▼
┌─────────────────────────────────────────────┐
│         Per-repo .github/settings.yml        │
│                                              │
│  - Inherits org defaults                     │
│  - Overrides status checks per repo          │
│  - Adds repo-specific labels                 │
└─────────────────────────────────────────────┘
```
<!-- scaffolded by uf vdev -->
