# Implementation Plan: GitHub Org GitOps

**Branch**: `032-org-gitops` | **Date**: 2026-04-18 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/032-org-gitops/spec.md`

## Summary

Set up GitOps management for the `unbound-force` GitHub
organization using two complementary tools: **Peribolos**
(uwu-tools/peribolos) for org-level management (members,
teams, org settings) and the **Repository Settings App**
(github.com/apps/settings) for repo-level settings
(merge strategy, labels, branch protection). The
implementation creates a new `.github` repo as the
single source of truth, seeds the current org state,
configures branch protection across all 7 repos, and
sets up CI-driven sync via GitHub Actions.

This spec involves **no Go source code changes** to the
unbound-force binary. All deliverables are YAML
configuration files, Markdown documents, and a GitHub
Actions workflow deployed to GitHub repositories.

## Technical Context

**Language/Version**: YAML (GitHub Actions, Peribolos config, Settings App config), Markdown
**Primary Dependencies**: uwu-tools/peribolos (Go binary, Apache 2.0), Repository Settings App (hosted GitHub App, Probot-based)
**Storage**: Git repositories (`.github` org repo + per-repo `.github/settings.yml`)
**Testing**: Manual verification via Peribolos dry-run, GitHub API queries, branch protection checks
**Target Platform**: GitHub (github.com/unbound-force organization)
**Project Type**: Infrastructure configuration (GitOps)
**Performance Goals**: N/A (configuration, not runtime)
**Constraints**: GitHub Free plan (no org-level rulesets); legacy branch protection API only; 6 members, 7 repos
**Scale/Scope**: 1 org, 7 repos, 6 members, 1 team

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Design Check

| Principle | Verdict | Rationale |
|-----------|---------|-----------|
| I. Autonomous Collaboration | **PASS** | This spec creates configuration files in Git repos — pure artifact-based collaboration. No runtime coupling between tools. Peribolos and the Settings App operate independently; each reads its own config file and applies changes without requiring the other. |
| II. Composability First | **PASS** | Each tool is independently installable and useful alone. Peribolos manages org membership without the Settings App. The Settings App manages repo settings without Peribolos. Combining them provides additive value (full org + repo coverage) without mandatory dependencies. |
| III. Observable Quality | **PASS** | All configuration is version-controlled in Git, providing full audit trail. Peribolos dry-run produces machine-readable diff output. The Settings App applies changes on push with GitHub's audit log. Branch protection status is queryable via the GitHub API. |
| IV. Testability | **PASS** | Peribolos supports `--confirm=false` dry-run mode for testing without side effects. The seed can be validated by running dry-run and expecting zero changes. Branch protection can be verified via API queries. No external services are required for validation beyond GitHub itself (which is the target platform). |

**Gate result**: PASS — all four principles satisfied.

### Post-Design Check

| Principle | Verdict | Rationale |
|-----------|---------|-----------|
| I. Autonomous Collaboration | **PASS** | The `.github` repo is a self-describing artifact store. Config files contain all metadata needed for tools to operate. No inter-tool communication required. |
| II. Composability First | **PASS** | Peribolos can be removed without affecting the Settings App, and vice versa. Per-repo `settings.yml` overrides work independently of org defaults (they just lose inheritance). |
| III. Observable Quality | **PASS** | Every change goes through a PR with git history. Peribolos workflow logs show exactly what was changed. GitHub API provides programmatic verification of applied settings. |
| IV. Testability | **PASS** | Coverage strategy is defined below. Validation is via Peribolos dry-run (zero-diff check) and GitHub API queries (branch protection, labels, settings). No Go code means no unit test coverage requirement — validation is operational. |

**Gate result**: PASS — all four principles satisfied post-design.

### Coverage Strategy

This spec produces no Go source code, so traditional
unit/integration/e2e test coverage metrics do not apply.
Instead, validation follows an operational verification
strategy:

| Verification Type | What It Checks | How |
|-------------------|---------------|-----|
| Seed accuracy | Peribolos YAML matches live org | `peribolos --config-path org/config.yaml` (dry-run, expect zero changes) |
| Branch protection | All repos have protection on main | `gh api repos/{owner}/{repo}/branches/main/protection` for each repo |
| Label consistency | All repos have 9 standard labels | `gh api repos/{owner}/{repo}/labels` for each repo |
| Settings inheritance | Per-repo overrides merge correctly | Push override, verify via API |
| Workflow execution | Peribolos sync runs on push | Merge a change, verify workflow completes |
| Safety guards | Mass removal is blocked | Modify YAML to remove >25% members, verify workflow fails |

This is not a coverage regression from the current state
(there is no existing GitOps configuration to regress
from). The verification strategy is appropriate for
infrastructure-as-code where the "tests" are operational
checks against the live system.

## Project Structure

### Documentation (this feature)

```text
specs/032-org-gitops/
├── plan.md              # This file
├── research.md          # Phase 0: tool research, YAML formats, auth
├── data-model.md        # Phase 1: config entities, concrete YAML
├── quickstart.md        # Phase 1: step-by-step setup guide
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Deliverables (deployed to GitHub repos)

```text
# New repo: unbound-force/.github
.github/
├── org/
│   └── config.yaml              # Peribolos org config (seeded)
├── settings.yml                 # Org-wide repo defaults
├── .github/
│   └── workflows/
│       └── peribolos-sync.yml   # CI workflow for org sync
├── CODEOWNERS                   # Access control for config files
└── profile/
    └── README.md                # Org profile (github.com/unbound-force)

# Per-repo overrides (in each repo)
<repo>/.github/settings.yml     # _extends: .github + repo-specific overrides
```

**Structure Decision**: No source code directories. All
deliverables are configuration files deployed to GitHub
repositories. The `.github` repo is the central config
store; per-repo `settings.yml` files provide overrides.

## Implementation Phases

### Phase A: Foundation (`.github` repo + GitHub App)

**Prerequisite**: Org admin access

1. Register the GitHub App (`unbound-force-peribolos`)
   with Organization Members and Administration
   permissions
2. Create the `.github` repo
3. Store APP_ID and APP_PRIVATE_KEY as repo secrets
4. Install the Repository Settings App on the org

### Phase B: Org Config (Peribolos)

**Prerequisite**: Phase A complete

1. Install Peribolos locally
2. Run `peribolos --dump` to seed current org state
3. Clean up the dump (remove sensitive fields, verify
   accuracy)
4. Add safety guard flags to the config
5. Validate with dry-run (expect zero changes)
6. Create the Peribolos sync workflow
7. Create CODEOWNERS

### Phase C: Repo Settings (Settings App)

**Prerequisite**: Phase A complete (Settings App installed)

1. Create org-wide `settings.yml` in `.github` repo
2. Create per-repo `settings.yml` overrides for repos
   with status checks or extra labels
3. Verify settings are applied via GitHub API

### Phase D: Org Profile + Verification

**Prerequisite**: Phases B and C complete

1. Create org profile README
2. Push all config to `.github` repo
3. Run full verification suite (dry-run, API checks,
   label audit)

## Dependency Map

```text
Phase A: Foundation
  ├── Register GitHub App (manual)
  ├── Create .github repo (manual/gh CLI)
  ├── Store secrets (gh CLI)
  └── Install Settings App (manual)
       │
       ├──────────────────────┐
       ▼                      ▼
Phase B: Org Config     Phase C: Repo Settings
  ├── Seed with --dump    ├── Org-wide settings.yml
  ├── Clean up YAML       ├── Per-repo overrides
  ├── Validate dry-run    └── Verify via API
  ├── Sync workflow
  └── CODEOWNERS
       │                      │
       └──────────┬───────────┘
                  ▼
Phase D: Profile + Verification
  ├── Org profile README
  ├── Push to .github
  └── Full verification
```

Phases B and C can proceed in parallel after Phase A.
Phase D depends on both B and C.

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Peribolos `--dump` output doesn't match expected format | Blocks seeding | Research confirmed format; manual cleanup step included |
| Settings App doesn't apply branch protection on Free | Blocks US5 | Research confirmed legacy API works on Free |
| GitHub App token lacks required permissions | Blocks CI sync | Permissions explicitly listed; test with dry-run first |
| Website repo has no PR-triggered CI | Status checks don't block PRs | Set `required_status_checks: null` for website; document gap |
| Mass member removal via typo | Org disruption | `--maximum-removal-delta=0.25` safety guard |
| Settings App escalation (push = admin) | Security risk | CODEOWNERS restricts `settings.yml` changes to admins |

## Complexity Tracking

> No constitution violations to justify.

| Aspect | Complexity | Notes |
|--------|-----------|-------|
| Tool count | 2 (Peribolos + Settings App) | Each handles a distinct domain (org vs repo); no overlap |
| Repos affected | 8 (1 new + 7 existing) | `.github` repo is new; 7 existing repos get `settings.yml` |
| Manual steps | 4 (app registration, app install, repo creation, secret storage) | Cannot be fully automated; documented in quickstart |
| YAML files | ~10 total | 1 Peribolos config + 1 org settings + 7 per-repo overrides + 1 workflow |
<!-- scaffolded by uf vdev -->
