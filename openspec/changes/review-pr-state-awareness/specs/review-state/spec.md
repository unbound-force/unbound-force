## ADDED Requirements

### Requirement: Fetch Existing Reviews

`/review-pr` MUST fetch existing PR reviews and inline
comments via `gh api` after Step 7 (Load Convention
Packs) and before Step 8 (AI Review).

#### Scenario: Reviews exist on the PR
- **GIVEN** a PR with 2 existing reviews and 5 inline
  comments
- **WHEN** `/review-pr` reaches the review state fetch
  step
- **THEN** all reviews and comments are fetched and
  structured as context for Step 8

#### Scenario: No existing reviews
- **GIVEN** a PR with no prior reviews
- **WHEN** `/review-pr` reaches the review state fetch
  step
- **THEN** the step completes with empty context and
  proceeds to Step 8

#### Scenario: API call fails
- **GIVEN** `gh api` returns 403 or times out
- **WHEN** `/review-pr` attempts to fetch review state
- **THEN** the error is logged, the step is skipped, and
  Step 8 proceeds without existing review context

### Requirement: Expanded Metadata Fetch

Step 2 MUST include `reviewDecision` and `reviewRequests`
in the `gh pr view --json` field list.

#### Scenario: PR with pending review requests
- **GIVEN** a PR with 2 pending review requests
- **WHEN** Step 2 fetches metadata
- **THEN** `reviewDecision` and `reviewRequests` are
  included in the metadata output

### Requirement: Duplicate Finding Suppression

Step 8 SHOULD annotate findings that overlap with
existing inline comments on the same file and line range
as "previously raised by @user".

#### Scenario: Existing comment on same line
- **GIVEN** an existing review comment on `foo.go:42`
  mentioning "missing error check"
- **WHEN** the AI review generates a finding for
  `foo.go:42` about error handling
- **THEN** the finding is annotated as "previously raised
  by @user" rather than presented as new

### Requirement: Token Budget for Existing Comments

Existing review comments passed to Step 8 MUST be capped
at 3000 characters total. Oldest comments SHOULD be
truncated first.

#### Scenario: Large comment volume
- **GIVEN** a PR with 50 inline comments totaling 8000
  characters
- **WHEN** review state is prepared for Step 8
- **THEN** only the most recent comments fitting within
  3000 characters are included

### Requirement: Stale Review Warning

When posting an APPROVE verdict in Step 11, `/review-pr`
SHOULD check if `dismiss_stale_reviews` is enabled on the
base branch and display a warning if so.

#### Scenario: APPROVE with stale dismissal enabled
- **GIVEN** a repo with `dismiss_stale_reviews: true`
- **WHEN** the user confirms posting an APPROVE review
- **THEN** a warning is displayed: "This repo dismisses
  stale reviews. If the author pushes any new commits
  after this APPROVE, it will be automatically
  invalidated."

#### Scenario: Branch protection API unavailable
- **GIVEN** the token lacks branch protection read access
- **WHEN** the stale review check is attempted
- **THEN** the check is silently skipped and the APPROVE
  proceeds without warning

### Requirement: Duplicate Review Detection

Before posting a review in Step 11, `/review-pr` MUST
check if a review from the same account already exists.

#### Scenario: Same verdict exists
- **GIVEN** the current account has a prior APPROVE on
  the PR
- **WHEN** the new verdict is also APPROVE
- **THEN** the user is asked: "You already have an
  APPROVE review on this PR. Post a new one?"

#### Scenario: Different verdict exists
- **GIVEN** the current account has a prior
  REQUEST_CHANGES on the PR
- **WHEN** the new verdict is APPROVE
- **THEN** the user is asked: "You have a prior
  REQUEST_CHANGES review. Post a new APPROVE? This will
  override the previous verdict."

### Requirement: CODEOWNER Awareness

When posting APPROVE with `require_code_owner_reviews:
true`, `/review-pr` SHOULD warn if the posting account
may not satisfy the CODEOWNER requirement.

#### Scenario: CODEOWNERS file exists
- **GIVEN** a repo with `require_code_owner_reviews: true`
  and a CODEOWNERS file
- **WHEN** posting APPROVE
- **THEN** a warning is displayed: "This repo requires
  code owner reviews. This APPROVE may not satisfy branch
  protection if this account is not listed in CODEOWNERS."

### Requirement: Dependabot Auto-Approval Idempotency

The `ci_dependencies.yml` auto-approval step MUST check
for existing REQUEST_CHANGES reviews from non-bot users
before creating an APPROVE review.

#### Scenario: Human REQUEST_CHANGES exists
- **GIVEN** a dependabot PR with a REQUEST_CHANGES review
  from a human user
- **WHEN** the auto-approval step runs
- **THEN** auto-approval is skipped with a log message

#### Scenario: No blocking reviews
- **GIVEN** a dependabot PR with no blocking reviews
- **WHEN** the auto-approval step runs
- **THEN** auto-approval proceeds normally

## MODIFIED Requirements

### Requirement: Step 2 Metadata Fields

Previously: `gh pr view <N> --json title,body,files,
additions,deletions,baseRefName,headRefName,labels,
milestone,commits`

Updated to include: `reviewDecision,reviewRequests`

## REMOVED Requirements

None
