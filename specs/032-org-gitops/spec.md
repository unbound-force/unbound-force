# Feature Specification: GitHub Org GitOps

**Feature Branch**: `032-org-gitops`
**Created**: 2026-04-18
**Status**: Draft
**Input**: User description: "Plan org and repo control with uwu-tools/peribolos and Repository Settings App"

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Seed Current Org State (Priority: P1)

An org admin wants to capture the current Unbound Force
GitHub organization configuration (members, teams, team
memberships, org settings) as a YAML file so that the
organization has a single source of truth for its
structure.

**Why this priority**: Without an accurate seed of the
current state, any subsequent GitOps changes risk
drift, accidental member removal, or team
misconfiguration. The seed is the foundation for
everything else.

**Independent Test**: Run the Peribolos `--dump` command
against the live org and verify the output YAML
matches the current org state (6 members, 1 team,
2 admins, 4 members, team membership).

**Acceptance Scenarios**:

1. **Given** the Unbound Force org exists with 6 members
   and 1 team, **When** the admin runs the seed command,
   **Then** a YAML file is produced that accurately lists
   all members (jflowers, marcusburghardt as admins;
   gvauter, jpower432, hbraswelrh, tyraziel as members),
   the overlords team (5 members, closed privacy, write
   permission on all repos), and org settings
   (default_repository_permission: read).

2. **Given** the seed YAML exists, **When** the admin
   runs Peribolos in dry-run mode against the live org,
   **Then** zero changes are proposed (confirming the
   seed accurately reflects reality).

---

### User Story 2 — Repo Settings via Pull Request (Priority: P1)

An org admin wants to define shared repository settings
(merge strategy, branch deletion policy, default labels)
in a central config file so that all 7 repos inherit
consistent defaults, and changes are reviewed via pull
request before taking effect.

**Why this priority**: Repository settings are currently
unmanaged — each repo has ad-hoc defaults with no
branch protection, inconsistent labels, and no
auto-delete on merge. This is the highest-impact
improvement for day-to-day development.

**Independent Test**: Install the Repository Settings
App on one test repo, push a `settings.yml`, and verify
the repo settings are updated to match the YAML.

**Acceptance Scenarios**:

1. **Given** a `.github` repo exists with a
   `settings.yml` defining org-wide defaults, **When**
   the Repository Settings App is installed on the org,
   **Then** all repos that inherit from `.github` adopt
   the default settings (delete_branch_on_merge: true,
   standard label set, branch protection on main).

2. **Given** a repo has a local `.github/settings.yml`
   with overrides, **When** a change is pushed to the
   repo's default branch, **Then** the repo's settings
   reflect the merged result of org defaults plus local
   overrides.

3. **Given** a contributor opens a PR that modifies
   `settings.yml`, **When** the PR is reviewed and
   merged, **Then** the settings are applied
   automatically and the change is auditable in git
   history.

---

### User Story 3 — Org Membership Changes via PR (Priority: P2)

An org admin wants to add, remove, or change the role of
an org member by editing a YAML file and opening a pull
request, so that membership changes are reviewed,
auditable, and reversible.

**Why this priority**: Membership changes are infrequent
(the org has 6 members) but high-risk — accidental
removal or role demotion can lock people out. PR-based
changes add a review gate.

**Independent Test**: Edit the members YAML to add a
test account, run Peribolos in dry-run mode, verify the
proposed change, then run with --confirm and verify the
new member received an invitation.

**Acceptance Scenarios**:

1. **Given** the org config YAML lists 6 members,
   **When** an admin adds a 7th member to the YAML and
   the CI workflow runs Peribolos with --confirm,
   **Then** the new member receives a GitHub org
   invitation.

2. **Given** a member is listed in the YAML, **When** an
   admin removes them and the workflow runs, **Then**
   the member is removed from the org.

3. **Given** a member's role changes from member to
   admin in the YAML, **When** the workflow runs,
   **Then** their org role is updated accordingly.

4. **Given** Peribolos is configured with
   --maximum-removal-delta=0.25, **When** a YAML change
   would remove more than 25% of members, **Then** the
   workflow fails with a safety error and no members are
   removed.

---

### User Story 4 — Team Management via PR (Priority: P2)

An org admin wants to create, modify, or delete teams
and manage team membership through YAML configuration,
so that team structure is version-controlled alongside
org membership.

**Why this priority**: The org currently has one team
(overlords) but will add more as the project grows.
Having team management in the same config file as
membership ensures consistency.

**Independent Test**: Add a new team definition to the
YAML, run Peribolos in dry-run mode, verify the
proposed team creation, then run with --confirm and
verify the team exists on GitHub.

**Acceptance Scenarios**:

1. **Given** the org config defines the overlords team
   with 5 members, **When** Peribolos runs, **Then**
   the team exists with the correct members and settings
   (closed privacy) and write permission on all 7 repos.

2. **Given** a new team is added to the YAML, **When**
   Peribolos runs, **Then** the team is created on
   GitHub with the specified members and privacy.

3. **Given** a team has repo permission assignments in
   the YAML, **When** Peribolos runs, **Then** the team
   has the correct permission level on each listed repo.

---

### User Story 5 — Branch Protection Across Repos (Priority: P2)

An org admin wants to enforce branch protection on the
main branch of all code repositories so that all changes
require pull request review before merging, consistent
with the project's existing code review requirements.

**Why this priority**: Currently zero repos have branch
protection. The project's AGENTS.md and constitution
already require code review, but nothing enforces it at
the GitHub level. This closes that gap.

**Independent Test**: Push a `settings.yml` with branch
protection config for main, verify that direct pushes
to main are rejected and PRs require at least 1
approving review.

**Acceptance Scenarios**:

1. **Given** branch protection is configured for main,
   **When** a contributor attempts to push directly to
   main, **Then** the push is rejected.

2. **Given** branch protection requires 1 approving
   review, **When** a PR is opened with no reviews,
   **Then** the merge button is disabled.

3. **Given** enforce_admins is false, **When** an org
   admin pushes directly to main, **Then** the push
   succeeds (emergency bypass).

4. **Given** required status checks are configured for a
   repo, **When** a PR is opened and CI fails, **Then**
   the merge button is disabled until CI passes.

---

### User Story 6 — CI-Driven Org Sync (Priority: P3)

An org admin wants org configuration changes to be
applied automatically when the config YAML is merged
to the `.github` repo's default branch, so that
manual intervention is not required after PR approval.

**Why this priority**: Automation is valuable but not
essential for v1 — the admin can also run Peribolos
locally. This story adds convenience and ensures the
config is always in sync.

**Independent Test**: Push a change to the org YAML on
the default branch, verify the GitHub Actions workflow
runs Peribolos and applies the change.

**Acceptance Scenarios**:

1. **Given** a PR modifying `org/` files is merged to
   main, **When** the sync workflow triggers, **Then**
   Peribolos runs with --confirm and applies the
   changes.

2. **Given** the workflow runs, **When** Peribolos
   encounters no changes needed, **Then** the workflow
   succeeds with a "no changes" summary.

3. **Given** the workflow fails (e.g., API rate limit),
   **When** an admin triggers the workflow manually via
   workflow_dispatch, **Then** the workflow retries
   successfully.

---

### Edge Cases

- What happens when a member listed in the YAML has
  a pending invitation that hasn't been accepted yet?
  Peribolos tracks pending invitations and does not
  re-invite.
- What happens when the GitHub App token generation
  fails or the app lacks sufficient permissions? The
  workflow fails with a clear error message identifying
  the missing permission or invalid credentials.
- What happens when the Repository Settings App and
  branch protection rules conflict? The Settings App
  uses the legacy branch protection API; if repo-level
  rulesets are also configured, the most restrictive
  combination applies (per GitHub's rule layering).
- What happens when a repo does not have a local
  `.github/settings.yml`? It inherits all defaults from
  the `.github` org repo's `settings.yml`.
- What happens when an admin removes themselves from
  the YAML? Peribolos's --required-admins flag prevents
  this if configured; otherwise, the safety delta check
  provides a backstop.
- What happens when a repo has custom labels not in
  the org defaults? The Repository Settings App only
  manages labels listed in the YAML; unlisted labels
  are not deleted (additive by default).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The `.github` repo MUST contain a
  Peribolos-compatible YAML file defining org settings,
  member roster (admins and members), and team
  definitions (name, privacy, members, maintainers,
  repo permissions).

- **FR-002**: The Peribolos config MUST be seeded from
  the current org state using the `--dump` command so
  that the initial YAML accurately reflects reality.

- **FR-003**: The `.github` repo MUST contain a
  `settings.yml` file defining org-wide default
  repository settings (merge strategy, branch deletion,
  vulnerability alerts, default labels, branch
  protection rules).

- **FR-004**: Per-repo `settings.yml` overrides MUST
  use the `_extends: .github` directive to inherit org
  defaults and override only repo-specific settings.

- **FR-005**: Branch protection on main MUST be
  configured with: required pull request reviews
  (1 approving review minimum), dismiss stale reviews
  on new commits, require code owner reviews (to
  enforce CODEOWNERS restrictions per FR-012), and
  enforce_admins set to false (allowing admin emergency
  bypass).

- **FR-006**: Required status checks MUST be configured
  per repo based on each repo's CI workflow job names:
  `test` (unbound-force, replicator), `unit-and-integration`
  and `e2e` (gaze), `build` (website),
  `build-and-test` (dewey).

- **FR-007**: The `delete_branch_on_merge` setting MUST
  be set to true in the org-wide defaults so that
  feature branches are cleaned up after merge.

- **FR-008**: The standard label set MUST be defined in
  the org-wide `settings.yml` and inherited by all
  repos: bug, documentation, duplicate, enhancement,
  good first issue, help wanted, invalid, question,
  wontfix.

- **FR-009**: Repos with additional labels (unbound-force:
  hero, phase-1, phase-2, spec-*; website: blog, docs,
  tutorial) MUST define those labels in their local
  `settings.yml`.

- **FR-010**: A GitHub Actions workflow MUST run
  Peribolos on push to the `.github` repo's default
  branch when files under `org/` are modified, and on
  manual workflow_dispatch trigger.

- **FR-011**: Peribolos MUST be configured with safety
  guards: --required-admins=jflowers,
  --min-admins=2, --maximum-removal-delta=0.25.

- **FR-012**: A CODEOWNERS file in the `.github` repo
  MUST restrict changes to `settings.yml`, `org/`, and
  `.github/workflows/` to org admins only.

- **FR-013**: The `.github` repo MUST include an org
  profile README (displayed on
  github.com/unbound-force) describing the project.

- **FR-014**: Repos without CI workflows (homebrew-tap)
  or with non-PR-triggered CI (containerfile) SHOULD
  have branch protection without required status checks.

### Key Entities

- **Org Config**: The Peribolos YAML file defining org
  settings, member roster, and team structure. Lives in
  the `.github` repo under `org/`.

- **Repo Settings**: The Repository Settings App YAML
  file defining repo-level settings, branch protection,
  labels, and collaborators. The org default lives in
  the `.github` repo as `settings.yml`; per-repo
  overrides live in each repo's `.github/settings.yml`.

- **Sync Workflow**: The GitHub Actions workflow that
  runs Peribolos to apply org config changes. Lives in
  the `.github` repo under `.github/workflows/`.

## Dependencies & Assumptions

- **uwu-tools/peribolos**: Standalone Go binary (Apache
  2.0), installable via `go install`. Supports --dump
  for seeding, --confirm for applying, safety guards
  for preventing accidental damage. Active as of April
  2026.

- **Repository Settings App**: Hosted GitHub App
  (github.com/apps/settings) by probot/settings.
  Installed on the org — zero infrastructure needed.
  Supports settings inheritance via `_extends` from
  the `.github` repo.

- **GitHub Free plan**: The org is on the Free plan.
  Org-level rulesets are NOT available (require Team
  plan). Repo-level rulesets are available but the
  Repository Settings App uses the legacy branch
  protection API instead, which works on Free.

- **GitHub App for authentication**: A registered GitHub
  App with Organization Members (read/write) and
  Organization Administration (read/write) permissions
  is used for Peribolos authentication. The workflow
  generates short-lived installation tokens per run
  using the app's private key (stored as repository
  secrets: APP_ID and APP_PRIVATE_KEY). This avoids
  dependency on personal accounts and provides automatic
  token rotation.

- **Existing org state**: 6 members (2 admins, 4
  members), 1 team (overlords, 5 members), 7 repos (all
  public, Apache-2.0 or MIT licensed, main branch, no
  branch protection, no rulesets).

## Clarifications

### Session 2026-04-18

- Q: Token type for Peribolos workflow — PAT, fine-grained PAT, or GitHub App? → A: GitHub App (register a custom app, generate short-lived installation tokens per workflow run).
- Q: Should the overlords team have explicit repo permission assignments? → A: Yes, assign write permission on all 7 repos.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All 7 repositories have branch protection
  enabled on main within one day of implementation, as
  verified by the GitHub API returning protection rules
  for each repo's main branch.

- **SC-002**: Org membership changes (add/remove/role
  change) can be completed through a PR workflow in
  under 10 minutes from PR creation to member
  invitation, without requiring manual GitHub UI
  interaction.

- **SC-003**: The Peribolos dry-run produces zero
  proposed changes when run against the live org
  immediately after initial seed, confirming 100%
  accuracy of the seed YAML.

- **SC-004**: All 7 repos share a consistent baseline of
  9 standard labels, verified by comparing each repo's
  label list against the org default after Settings App
  deployment.

- **SC-005**: Direct pushes to main are blocked on all
  code repos (unbound-force, gaze, website, dewey,
  replicator) for non-admin users, verified by
  attempting a direct push with a member-role account.

- **SC-006**: The `.github` repo serves as the single
  source of truth — any manual change to org settings,
  membership, or repo settings that diverges from the
  YAML is detected and corrected on the next Peribolos
  or Settings App sync cycle.
<!-- scaffolded by uf vdev -->
