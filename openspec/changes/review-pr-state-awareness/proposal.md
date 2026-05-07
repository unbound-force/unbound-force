## Why

`/review-pr` has no awareness of GitHub's review state
model. It operates in a vacuum — generating findings
without checking what's already been said, posting reviews
without understanding the current review state, and
offering APPROVEs without warning about stale dismissal.
This leads to duplicate findings, redundant work, stale
APPROVE surprises, ghost review requests, duplicate review
events, CODEOWNER blindness, and dependabot auto-approval
gaps.

Fixes #165.

## What Changes

Add GitHub review state awareness to `/review-pr` by:

1. Fetching existing reviews and inline comments before
   AI analysis (new Step 7.5)
2. Expanding Step 2 metadata to include `reviewDecision`
   and `reviewRequests`
3. Adding stale review warnings when posting APPROVE with
   `dismiss_stale_reviews` enabled
4. Adding duplicate review detection before posting
5. Adding CODEOWNER awareness when posting APPROVE
6. Adding idempotency check to dependabot auto-approval
   in `ci_dependencies.yml`
7. Documenting GitHub review lifecycle in AGENTS.md

## Capabilities

### New Capabilities
- `review-state-fetch`: Fetches existing PR reviews and
  inline comments via `gh api` before AI analysis
- `stale-review-warning`: Warns when posting APPROVE with
  `dismiss_stale_reviews` enabled
- `duplicate-review-detection`: Detects and warns about
  existing reviews from the same account
- `codeowner-awareness`: Warns when APPROVE may not
  satisfy branch protection CODEOWNER requirements
- `dependabot-idempotency`: Checks for blocking
  REQUEST_CHANGES before dependabot auto-approval

### Modified Capabilities
- `metadata-fetch`: Step 2 expanded to include
  `reviewDecision` and `reviewRequests` fields
- `ai-review`: Step 8 receives existing comments as
  context to suppress/annotate duplicate findings
- `verdict-posting`: Step 11 enhanced with stale review,
  duplicate, and CODEOWNER warnings

### Removed Capabilities
- None

## Impact

- `.opencode/commands/review-pr.md` — new step, expanded
  metadata, enhanced verdict posting
- `internal/scaffold/assets/opencode/command/review-pr.md`
  — scaffold copy sync
- `.github/workflows/ci_dependencies.yml` — idempotency
  check in auto-approval step
- `AGENTS.md` — GitHub review lifecycle documentation

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

Review state data is fetched via `gh api` and passed as
structured context to the AI review step. No runtime
coupling — the review state is consumed as data input,
not a live API dependency during analysis.

### II. Composability First

**Assessment**: PASS

All new capabilities degrade gracefully. If `gh api`
calls fail for review state, the command proceeds
without deduplication (current behavior). CODEOWNER
and stale review warnings are additive — they do not
block the review flow.

### III. Observable Quality

**Assessment**: PASS

Existing review data is surfaced in the output format
with structured annotations ("previously raised by
@user"). Warnings are machine-readable with consistent
prefix markers.

### IV. Testability

**Assessment**: N/A

No Go code changes. Markdown + YAML only. The CI
workflow JavaScript change is testable via the existing
GitHub Actions test infrastructure.
