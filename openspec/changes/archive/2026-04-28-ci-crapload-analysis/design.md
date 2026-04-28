## Context

The complytime organization maintains a reusable GitHub Actions
workflow for CRAP (Change Risk Anti-Patterns) analysis at
`complytime/org-infra/.github/workflows/
reusable_crapload_analysis.yml`. This workflow:

1. Detects changed Go packages via `git diff` against the PR base
2. Runs `go test` with coverage on changed packages only
3. Runs `gaze report --format=json` to produce CRAP scores
4. Compares current scores against a committed baseline file
5. Generates a Markdown PR comment body with summary, regressions,
   improvements, and new functions
6. Uploads the comment body as a GitHub Actions artifact
7. Fails the job if regressions or threshold violations are found

The consumer workflow (`ci_crapload.yml`) calls this reusable
workflow and handles PR comment posting in a separate job.

The reusable workflow is in a public repository, so cross-org
calls from `unbound-force` work without special configuration.

## Goals / Non-Goals

### Goals

- Call the complytime reusable CRAP workflow from unbound-force
  PRs with a pinned SHA reference
- Configure dependabot to auto-update the pinned SHA
- Generate and commit a baseline from the current codebase
- Add SPDX license headers to new workflow files
- Make the CRAP check blocking (fail CI on regressions)

### Non-Goals

- Modifying the reusable workflow (owned by complytime/org-infra)
- Changing the existing test.yml workflow logic
- Auto-merging dependabot PRs (manual review)
- Customizing Gaze version or CRAP thresholds (defaults are fine)
- Adding SPDX headers to existing workflow files (separate change)

## Decisions

### D1: Pin reusable workflow to SHA, not branch

**Decision**: Reference the reusable workflow as
`complytime/org-infra/...@9205a3ac6b76b75dbe6e22b2f0f330bc8edbeb38`
instead of `@main`.

**Rationale**: Pinning to SHA provides reproducible builds and
protects against upstream breaking changes. Dependabot's
`github-actions` ecosystem updates reusable workflow references
the same way it updates action pins -- opening PRs when new
commits appear on the default branch. This gives us controlled
updates with review.

**Alternative considered**: `@main` (what complyctl uses) --
simpler but means any upstream push immediately affects our CI.

### D2: Two-job pattern (reusable + comment poster)

**Decision**: Keep the same two-job architecture as complyctl:
Job 1 calls the reusable workflow (permissions: contents read),
Job 2 downloads the artifact and posts a PR comment (permissions:
pull-requests write).

**Rationale**: The reusable workflow cannot receive
`pull-requests: write` permission (it only declares
`contents: read`). Splitting the comment posting into a separate
job with its own permissions is the standard GitHub Actions
pattern for this. The `!cancelled()` condition on Job 2 ensures
comments are posted even when the analysis job fails
(regression detected).

### D3: Dependabot covers github-actions and gomod

**Decision**: Configure both `github-actions` (daily) and `gomod`
(weekly, 10 PR limit) ecosystems.

**Rationale**: The `github-actions` ecosystem is required for the
primary goal (pinned SHA management). Adding `gomod` was an
explicit user decision to also track Go dependency updates.
Daily cadence for actions matches complyctl's pattern. Weekly
for gomod avoids PR noise while staying current.

### D4: Baseline generated from current main

**Decision**: Generate `.gaze/baseline.json` by running Gaze
against all packages on the current main branch and committing
the result.

**Rationale**: The baseline represents the "known good" state.
All future PRs are compared against it. Without a baseline, the
workflow passes with a warning but provides no regression data.
Generating from main ensures the first PR after this change
gets meaningful comparison data.

### D5: Default thresholds

**Decision**: Accept the reusable workflow's defaults:
CRAP threshold 15, new function threshold 30.

**Rationale**: These are reasonable defaults. The CRAP threshold
of 15 matches academic literature. The new function threshold of
30 is lenient for initial adoption. Both can be overridden via
workflow inputs if needed later.

### D6: SPDX license headers on new files

**Decision**: Add `# SPDX-License-Identifier: Apache-2.0` to
new workflow files, consistent with complyctl's pattern.

**Rationale**: User decision for consistency with the source
pattern. Existing workflows do not have SPDX headers and are
not modified by this change.

## Risks / Trade-offs

### R1: Upstream workflow changes

**Risk**: The complytime reusable workflow may change in ways
that break our consumer workflow (artifact name changes, output
key renames).

**Mitigation**: SHA pinning means changes only arrive when we
accept a dependabot PR. Review the diff before merging.

### R2: Gaze version drift

**Risk**: The reusable workflow installs `gaze@latest` by
default. Gaze breaking changes could cause unexpected failures.

**Mitigation**: We can override via the `gaze-version` input
to pin a specific Gaze version. Not needed initially -- Gaze
follows semver and breaking changes require a major bump.

### R3: Baseline staleness

**Risk**: The baseline is a point-in-time snapshot. As the
codebase evolves, functions are renamed or moved, and baseline
entries become orphaned while new functions have no baseline
entry (treated as "new" with the 30-threshold ceiling).

**Mitigation**: Periodic baseline regeneration. The reusable
workflow handles both cases gracefully: orphaned baseline
entries are silently ignored, new functions are reported with
their threshold status.

### R4: Dependabot PR volume

**Risk**: Daily `github-actions` checks across 3 workflow files
plus weekly `gomod` checks could generate significant PR volume.

**Mitigation**: The 10-PR limit on `gomod` caps the worst case.
Action updates are typically infrequent (major actions release
monthly). PRs can be batched or auto-merged with additional
tooling if volume becomes a problem.
