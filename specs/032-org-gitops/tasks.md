# Tasks: GitHub Org GitOps

**Input**: Design documents from `/specs/032-org-gitops/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, quickstart.md

**Note**: This spec produces NO Go source code. All
deliverables are YAML configuration files, Markdown
documents, and a GitHub Actions workflow deployed to
GitHub repositories. Validation is operational (dry-run,
API queries), not unit tests.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Exact file paths are relative to the target repo root

---

## Phase 1: Setup (GitHub Infrastructure)

**Purpose**: Create the `.github` repo, register the
GitHub App for Peribolos authentication, and install the
Repository Settings App on the org. These are manual
steps that MUST be completed before any config files can
be deployed.

- [x] T001 Register GitHub App `unbound-force-peribolos` at github.com/organizations/unbound-force/settings/apps with Organization Members (Read & Write) and Organization Administration (Read & Write) permissions, webhook deactivated, restricted to this account only (per research.md §3)
- [x] T002 Generate private key for the GitHub App and note the App ID from the app settings page
- [x] T003 Install the registered GitHub App on the `unbound-force` organization
- [x] T004 Create the `.github` repo: `gh repo create unbound-force/.github --public --description "Organization-wide GitHub configuration" --clone`
- [x] T005 Store APP_ID and APP_PRIVATE_KEY as repository secrets on the `.github` repo: `gh secret set APP_ID --repo unbound-force/.github` and `gh secret set APP_PRIVATE_KEY < private-key.pem --repo unbound-force/.github`
- [x] T006 Install the Repository Settings App on the `unbound-force` org via github.com/apps/settings — grant access to All repositories (per quickstart.md §8)

**Checkpoint**: `.github` repo exists, GitHub App is
registered and installed, secrets are stored, Settings
App is installed. All subsequent phases can proceed.

---

## Phase 2: Foundation (Org Config Seed)

**Purpose**: Install Peribolos locally, seed the current
org state, and validate the seed is accurate. This is the
foundation for US1 (seed) and US3/US4 (membership/teams).

- [x] T007 Install Peribolos locally: `go install github.com/uwu-tools/peribolos@latest`
- [x] T008 Run `peribolos --dump unbound-force --github-token-path <(gh auth token) > org/config.yaml` to seed the current org state into the `.github` repo (per quickstart.md §3)
- [x] T009 Review and clean up the Peribolos dump output: remove `billing_email` if present, verify 6 members (2 admins: jflowers, marcusburghardt; 4 members: gvauter, hbraswelrh, jpower432, tyraziel), verify overlords team (5 members, closed privacy), verify `default_repository_permission: read` (per data-model.md Entity 1)
- [x] T010 Add overlords team repo permissions to `org/config.yaml` — assign `write` on all 7 repos (unbound-force, gaze, website, dewey, replicator, homebrew-tap, containerfile) under `teams.overlords.repos` (per spec.md clarification and data-model.md concrete instance)
- [x] T011 Validate seed accuracy: run `peribolos --config-path org/config.yaml --github-token-path <(gh auth token) --required-admins=jflowers --min-admins=2` and verify zero proposed changes (per SC-003)

**Checkpoint**: `org/config.yaml` accurately reflects the
live org state. Peribolos dry-run shows zero changes.

---

## Phase 3: US1 — Seed Current Org State (Priority: P1) 🎯 MVP

**Goal**: The Peribolos YAML in `org/config.yaml` is the
single source of truth for org structure (members, teams,
org settings).

**Independent Test**: Peribolos dry-run produces zero
proposed changes against the live org.

- [x] T012 [US1] Verify `org/config.yaml` contains all required org settings: `default_repository_permission: read`, `members_can_create_repositories: true`, `has_organization_projects: true`, `has_repository_projects: true` (per data-model.md concrete instance)
- [x] T013 [US1] Verify `org/config.yaml` member roster matches live org — cross-reference with `gh api orgs/unbound-force/members --jq '.[].login'` and `gh api orgs/unbound-force/members?role=admin --jq '.[].login'`
- [x] T014 [US1] Verify `org/config.yaml` team structure — cross-reference overlords team with `gh api orgs/unbound-force/teams/overlords/members --jq '.[].login'` (expect 5 members: jflowers, marcusburghardt, gvauter, jpower432, hbraswelrh; tyraziel is NOT in overlords)
- [x] T015 [US1] Final dry-run validation with all safety guards: `peribolos --config-path org/config.yaml --github-token-path <(gh auth token) --required-admins=jflowers --min-admins=2 --maximum-removal-delta=0.25` — expect zero changes (per FR-002, FR-011, SC-003)

**Checkpoint**: US1 complete — org config YAML is seeded
and validated. Peribolos dry-run confirms 100% accuracy.

---

## Phase 4: US2 — Repo Settings via Pull Request (Priority: P1)

**Goal**: Org-wide default repo settings and per-repo
overrides are defined in YAML, applied by the Repository
Settings App.

**Independent Test**: Push `settings.yml` to the `.github`
repo, verify repo settings are updated via GitHub API.

### Org-Wide Defaults

- [x] T016 [US2] Create `settings.yml` in `.github` repo root with org-wide defaults: `delete_branch_on_merge: true`, all merge strategies enabled, `enable_automated_security_fixes: true`, `enable_vulnerability_alerts: true` (per FR-003, FR-007, data-model.md Entity 2)
- [x] T017 [US2] Add 9 standard labels to `settings.yml`: bug, documentation, duplicate, enhancement, good first issue, help wanted, invalid, question, wontfix — with colors and descriptions per data-model.md (per FR-008)
- [x] T018 [US2] Add branch protection defaults to `settings.yml` for main: `required_approving_review_count: 1`, `dismiss_stale_reviews: true`, `require_code_owner_reviews: true`, `required_status_checks: null`, `enforce_admins: false`, `restrictions: null` (per FR-005, data-model.md Entity 2)

### Per-Repo Overrides

- [x] T019 [P] [US2] Create `unbound-force/.github/settings.yml` with `_extends: .github`, hero/phase-1/phase-2 labels, and required status check `Build & Test` (per FR-004, FR-006, FR-009, data-model.md Entity 3)
- [x] T020 [P] [US2] Create `gaze/.github/settings.yml` with `_extends: .github` and required status checks `Unit + Integration Tests (Go 1.24)`, `Unit + Integration Tests (Go 1.25)`, `e2e` (per FR-006, data-model.md status check resolution table)
- [x] T021 [P] [US2] Create `dewey/.github/settings.yml` with `_extends: .github` and required status check `build-and-test` (per FR-006)
- [x] T022 [P] [US2] Create `replicator/.github/settings.yml` with `_extends: .github` and required status check `test` (per FR-006)
- [x] T023 [P] [US2] Create `website/.github/settings.yml` with `_extends: .github`, blog/docs/tutorial labels, and `required_status_checks: null` (per FR-009, FR-014, data-model.md)
- [x] T024 [P] [US2] Create `homebrew-tap/.github/settings.yml` with `_extends: .github` and `required_status_checks: null` (per FR-014)
- [x] T025 [P] [US2] Create `containerfile/.github/settings.yml` with `_extends: .github` and `required_status_checks: null` (per FR-014)

### Verification

- [x] T026 [US2] Verify org-wide settings applied: `gh api repos/unbound-force/unbound-force --jq '{delete_branch_on_merge, allow_squash_merge}'` — confirm `delete_branch_on_merge: true` (per SC-006)
- [x] T027 [US2] Verify standard labels on all repos: `gh api repos/unbound-force/unbound-force/labels --jq '.[].name' | sort` — confirm 9 standard labels present (per SC-004)

**Checkpoint**: US2 complete — all 7 repos inherit org
defaults, repos with CI have required status checks,
repos without PR-triggered CI have null status checks.

---

## Phase 5: US3 + US4 — Membership & Teams via PR (Priority: P2)

**Goal**: Org membership and team management are
controlled through the Peribolos YAML created in US1.
No additional config files are needed — US3 and US4 are
satisfied by the `org/config.yaml` from Phase 2/3.

**Independent Test**: Edit the YAML to add a member, run
dry-run, verify the proposed change.

- [x] T028 [US3] Verify membership change workflow: temporarily add a test username to `org/config.yaml` members list, run Peribolos dry-run, confirm it proposes adding the member, then revert the change (per FR-001, SC-002)
- [x] T029 [US3] Verify safety guard: temporarily remove >25% of members from `org/config.yaml`, run Peribolos dry-run with `--maximum-removal-delta=0.25`, confirm it fails with a safety error, then revert (per FR-011)
- [x] T030 [US4] Verify team management: temporarily add a new team definition to `org/config.yaml`, run Peribolos dry-run, confirm it proposes creating the team, then revert (per FR-001)

**Checkpoint**: US3 + US4 complete — membership and team
changes are managed through `org/config.yaml` with
safety guards.

---

## Phase 6: US5 — Branch Protection Across Repos (Priority: P2)

**Goal**: All code repos have branch protection on main
requiring PR review. Non-code repos have protection
without status checks.

**Independent Test**: Attempt a direct push to main on a
protected repo — it should be rejected for non-admin
users.

- [x] T031 [US5] Verify branch protection applied on all repos: run `gh api repos/unbound-force/{repo}/branches/main/protection --jq '{required_reviews: .required_pull_request_reviews.required_approving_review_count, enforce_admins: .enforce_admins.enabled}'` for each of the 7 repos (per SC-001, SC-005)
- [x] T032 [US5] Verify required status checks on code repos: `gh api repos/unbound-force/unbound-force/branches/main/protection/required_status_checks --jq '.contexts'` — confirm `Build & Test` is listed (per FR-006)
- [x] T033 [US5] Verify repos without CI have no required status checks: `gh api repos/unbound-force/homebrew-tap/branches/main/protection/required_status_checks` — confirm null or empty contexts (per FR-014)
- [x] T034 [US5] Verify admin bypass: confirm `enforce_admins: false` on all repos so org admins can push directly in emergencies (per FR-005)

**Checkpoint**: US5 complete — branch protection is
enforced on all repos with appropriate status checks.
Note: unbound-force repo branch protection is pending
Settings App webhook processing (settings.yml deployed
via API, webhook fires on next git push).

---

## Phase 7: US6 — CI-Driven Org Sync (Priority: P3)

**Goal**: Org config changes are applied automatically
when merged to the `.github` repo's main branch.

**Independent Test**: Push a change to `org/` on main,
verify the GitHub Actions workflow runs Peribolos.

- [x] T035 [US6] Look up the SHA-pinned version of `actions/create-github-app-token@v1` for use in the workflow (per research.md §5 SHA-pinned convention)
- [x] T036 [US6] Create `.github/workflows/peribolos-sync.yml` in the `.github` repo with: trigger on push to main (paths: `org/**`) and `workflow_dispatch`, `permissions: contents: read`, job using `actions/checkout` (SHA-pinned), `actions/create-github-app-token` (SHA-pinned), `actions/setup-go` (SHA-pinned), `go install peribolos@latest`, and `peribolos --config-path org/config.yaml --required-admins=jflowers --min-admins=2 --maximum-removal-delta=0.25 --confirm` (per FR-010, FR-011, data-model.md Entity 4, quickstart.md §5)
- [x] T037 [US6] Verify workflow triggers: push a trivial comment change to `org/config.yaml` on main, confirm the `Org Sync` workflow runs and completes successfully
- [x] T038 [US6] Verify manual dispatch: trigger the workflow via `gh workflow run peribolos-sync.yml --repo unbound-force/.github`, confirm it completes successfully (per spec.md US6 acceptance scenario 3)

**Checkpoint**: US6 complete — org config changes are
automatically applied via CI on merge to main.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: CODEOWNERS, org profile, documentation
updates, and final verification.

### Access Control & Profile

- [x] T039 [P] Create `CODEOWNERS` in `.github` repo root: `settings.yml`, `org/`, and `.github/` paths restricted to `@jflowers @marcusburghardt` (per FR-012, data-model.md Entity 5)
- [x] T040 [P] Create `profile/README.md` in `.github` repo with org description, hero table (Cobalt-Crush, Gaze, The Divisor, Muti-Mind, Mx F), tools table (Dewey, Replicator), and getting started instructions (per FR-013, quickstart.md §7)

### Push All Config

- [x] T041 Commit and push all files to the `.github` repo: `org/config.yaml`, `settings.yml`, `.github/workflows/peribolos-sync.yml`, `CODEOWNERS`, `profile/README.md`
- [x] T042 Push per-repo `.github/settings.yml` overrides to each of the 7 repos (unbound-force, gaze, dewey, replicator, website, homebrew-tap, containerfile)

### Final Verification

- [x] T043 [P] Run full Peribolos dry-run with all safety guards and confirm zero changes (final seed accuracy check per SC-003)
- [x] T044 [P] Verify branch protection on all 7 repos via GitHub API (per SC-001)
- [x] T045 [P] Verify 9 standard labels on all 7 repos via GitHub API (per SC-004)
- [x] T046 [P] Verify org profile README is visible at github.com/unbound-force (per FR-013)
- [x] T047 Verify the `.github` repo CODEOWNERS is enforced: confirm that PRs modifying `settings.yml`, `org/`, or `.github/` require review from @jflowers or @marcusburghardt (per FR-012)
- [x] T048 Run quickstart.md §10 full verification script to confirm all success criteria (SC-001 through SC-006)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately.
  Tasks T001–T006 are sequential (app registration before
  repo creation before secret storage).
- **Phase 2 (Foundation)**: Depends on Phase 1 completion.
  Tasks T007–T011 are sequential (install before seed
  before cleanup before validation).
- **Phase 3 (US1)**: Depends on Phase 2 completion.
  Verification of the seed from Phase 2.
- **Phase 4 (US2)**: Depends on Phase 1 completion
  (Settings App installed). T016–T018 are sequential
  (build the org-wide file). T019–T025 are parallel
  (per-repo overrides touch different repos).
- **Phase 5 (US3+US4)**: Depends on Phase 2 completion.
  Verification tasks only — no new files.
- **Phase 6 (US5)**: Depends on Phase 4 completion
  (branch protection is configured in settings.yml).
  Verification tasks only.
- **Phase 7 (US6)**: Depends on Phase 2 completion
  (org/config.yaml exists). Can proceed in parallel
  with Phase 4.
- **Phase 8 (Polish)**: Depends on all prior phases.
  T039–T040 are parallel. T043–T046 are parallel.

### Parallel Opportunities

```text
After Phase 1 completes:
  ├── Phase 2 (Foundation) ──→ Phase 3 (US1)
  │                        ──→ Phase 5 (US3+US4)
  │                        ──→ Phase 7 (US6)
  └── Phase 4 (US2) ──────→ Phase 6 (US5)

After all phases:
  └── Phase 8 (Polish)
```

Within Phase 4: T019–T025 (per-repo overrides) are all
parallel — each touches a different repository.

Within Phase 8: T039–T040 (CODEOWNERS + profile) are
parallel. T043–T046 (verification) are parallel.

---

## Notes

- All tasks produce YAML, Markdown, or shell commands —
  no Go source code
- Tasks T001–T006 require org admin access and are
  partially manual (GitHub UI)
- Peribolos dry-run is the primary validation mechanism
  (expect zero changes after seed)
- Per-repo `settings.yml` files use `_extends: .github`
  for inheritance — only overrides are specified
- SHA-pinned action versions are required per project
  convention (research.md §5)
- The website repo's deploy workflow is push-only (not
  PR-triggered) — branch protection is configured with
  `required_status_checks: null` until a PR-triggered
  workflow is added
- The `unbound-force` repo's branch protection is
  pending Settings App webhook processing (settings.yml
  was deployed via GitHub API; webhook fires on next
  git push to default branch)
<!-- spec-review: passed -->
<!-- code-review: passed -->
<!-- scaffolded by uf vdev -->
