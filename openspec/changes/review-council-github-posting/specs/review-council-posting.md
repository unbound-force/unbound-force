## ADDED Requirements

### Requirement: PR Detection

The `/review-council` command MUST detect whether the current
branch has an open pull request on GitHub. Detection SHALL use
`gh pr view --json number,headRefName,baseRefName`. The user
MAY provide an explicit PR number as an argument to override
auto-detection. The PR number argument MUST be validated as a
positive integer (digits only, range 1-999999). Non-numeric or
out-of-range values MUST be rejected with an informational
error message.

#### Scenario: Auto-detect PR for current branch

- **GIVEN** the current branch has an open PR on GitHub
- **WHEN** `/review-council` completes the fix loop (Step 6)
- **THEN** the command detects the PR number and offers
  GitHub review posting

#### Scenario: Explicit PR number argument

- **GIVEN** the user runs `/review-council code 42`
- **WHEN** the command processes arguments
- **THEN** PR number 42 is used for posting, overriding
  auto-detection

#### Scenario: Invalid PR number argument

- **GIVEN** the user runs `/review-council code abc`
- **WHEN** the command processes arguments
- **THEN** the argument is rejected with an informational
  error message
- **AND** posting is skipped

#### Scenario: No PR exists

- **GIVEN** the current branch has no open PR on GitHub
- **WHEN** `/review-council` completes the fix loop
- **THEN** Step 7 is skipped with an informational note
- **AND** the command behaves identically to current behavior

#### Scenario: `gh` CLI unavailable

- **GIVEN** the `gh` CLI is not installed or not authenticated
- **WHEN** `/review-council` attempts PR detection
- **THEN** Step 7 is skipped with an informational note
- **AND** the command behaves identically to current behavior

### Requirement: Review State Fetching

When a PR is detected, the command MUST fetch existing review
state before posting. This SHALL include: existing reviews,
inline comments, and current authenticated user login.

#### Scenario: Fetch existing reviews

- **GIVEN** a PR is detected with number N
- **WHEN** review state is fetched
- **THEN** existing reviews are retrieved via `gh api`
- **AND** each review's user, state, body, and commit ID are
  recorded

#### Scenario: API error during fetch

- **GIVEN** a PR is detected but `gh api` returns 403 or 404
- **WHEN** review state fetching fails
- **THEN** the sub-step is skipped silently
- **AND** posting proceeds with reduced deduplication accuracy

### Requirement: Multi-Persona Finding Aggregation

The command MUST aggregate findings from all Divisor personas
into a single GitHub review body with per-persona sections.
The review body MUST include: council verdict, reviewer list,
iteration count, and per-persona findings with severity levels.

#### Scenario: Multiple personas with findings

- **GIVEN** Guard has 2 findings and Adversary has 3 findings
- **WHEN** the review body is aggregated
- **THEN** the body contains a "Guard" section with 2 findings
- **AND** the body contains an "Adversary" section with 3
  findings
- **AND** each finding includes its severity level

#### Scenario: Persona with no findings

- **GIVEN** Architect returned APPROVE with no findings
- **WHEN** the review body is aggregated
- **THEN** the Architect section reads "No findings."

#### Scenario: Cross-persona consolidated findings

- **GIVEN** a finding was consolidated across Guard and SRE
- **WHEN** the review body is aggregated
- **THEN** the finding appears under the primary persona
- **AND** the attribution notes both contributing personas

### Requirement: Inline Comment Allocation

File-specific findings MUST be posted as inline comments,
capped at 15. Allocation MUST follow severity-first
round-robin: sort by severity (CRITICAL > HIGH > MEDIUM >
LOW), then round-robin across personas within the same
severity tier.

#### Scenario: Under cap

- **GIVEN** 10 total file-specific findings across all personas
- **WHEN** inline comments are prepared
- **THEN** all 10 are included as inline comments

#### Scenario: Over cap

- **GIVEN** 20 total file-specific findings across all personas
- **WHEN** inline comments are prepared
- **THEN** 15 are posted as inline comments
- **AND** remaining 5 are summarized in the review body
- **AND** the 15 selected prioritize CRITICAL over HIGH

#### Scenario: Fair distribution across personas

- **GIVEN** Adversary has 10 HIGH findings and Guard has
  10 HIGH findings
- **WHEN** inline comments are prepared with 15 slots
- **THEN** allocation round-robins: Adversary gets 8, Guard
  gets 7 (or similar balanced split)

### Requirement: Verdict Mapping

The council verdict MUST be mapped to GitHub API event types:
APPROVE to `APPROVE`, REQUEST CHANGES to `REQUEST_CHANGES`,
APPROVE WITH ADVISORIES to `COMMENT`.

#### Scenario: Unanimous approval

- **GIVEN** all personas returned APPROVE
- **WHEN** the verdict is mapped
- **THEN** the GitHub event type is `APPROVE`

#### Scenario: Any persona dissented

- **GIVEN** at least one persona returned REQUEST CHANGES
- **WHEN** the verdict is mapped
- **THEN** the GitHub event type is `REQUEST_CHANGES`

#### Scenario: Approval with advisories

- **GIVEN** the council returned APPROVE WITH ADVISORIES
- **WHEN** the verdict is mapped
- **THEN** the GitHub event type is `COMMENT`
- **AND** the review body includes a note explaining the
  mapping

### Requirement: Pre-Posting Checks

Before posting, the command MUST perform: duplicate review
detection, stale review dismissal warning (APPROVE only),
and CODEOWNER requirement warning (APPROVE only).

#### Scenario: Duplicate review exists with same verdict

- **GIVEN** the current user already has an APPROVE review
- **AND** the council verdict is APPROVE
- **WHEN** pre-posting checks run
- **THEN** the user is warned and asked to confirm overwrite

#### Scenario: Stale review dismissal enabled

- **GIVEN** the base branch has `dismiss_stale_reviews` enabled
- **AND** the council verdict is APPROVE
- **WHEN** pre-posting checks run
- **THEN** the user is warned that the APPROVE will be
  invalidated if new commits are pushed

#### Scenario: CODEOWNER review required

- **GIVEN** the repo requires code owner reviews
- **AND** the council verdict is APPROVE
- **WHEN** pre-posting checks run
- **THEN** the user is warned that this APPROVE may not
  satisfy branch protection

### Requirement: Human Confirmation

The command MUST require explicit human confirmation via the
AskUserQuestion tool before posting any review. The command
MUST NOT post reviews without confirmation.

#### Scenario: User confirms posting

- **GIVEN** the user is shown the prepared review content
- **WHEN** the user selects "Yes -- post review"
- **THEN** the review is posted to GitHub

#### Scenario: User declines posting

- **GIVEN** the user is shown the prepared review content
- **WHEN** the user selects "No -- skip posting"
- **THEN** no review is posted
- **AND** the terminal report remains the only output

#### Scenario: User edits before posting

- **GIVEN** the user is shown the prepared review content
- **WHEN** the user selects "Edit comments first"
- **THEN** the user can modify comments before re-confirming

### Requirement: Graceful Degradation

The posting step MUST degrade gracefully when `gh` is
unavailable, unauthenticated, or API calls fail.

#### Scenario: Posting returns 403

- **GIVEN** `gh api` returns HTTP 403 (self-review prohibition)
- **WHEN** posting fails
- **THEN** the command falls back to `COMMENT` event type
- **AND** the review body notes the original verdict

#### Scenario: Fallback also fails

- **GIVEN** the fallback `COMMENT` posting also fails
- **WHEN** the second attempt fails
- **THEN** the user is informed their token lacks permissions
- **AND** the terminal report remains available

#### Scenario: Rate limited

- **GIVEN** `gh api` returns HTTP 429 (rate limited)
- **WHEN** posting or state fetching is attempted
- **THEN** the command skips the affected step gracefully
- **AND** the user is informed of the rate limit

#### Scenario: `gh` version too old

- **GIVEN** `gh` is installed but does not support `--input`
- **WHEN** posting fails with an unrecognized flag error
- **THEN** the user is informed to upgrade `gh` to >= 2.0
- **AND** the terminal report remains available

### Requirement: Protocol 2 Conditional Execution

When an explicit PR number is provided as an argument, the
review-context skill's Protocol 2 (Issue Linking) MUST be
executed during Phase 1c using the PR body. When a PR is
auto-detected (no explicit argument), Protocol 2 context is
included in the posted review body but does not enrich Step 2
delegation. When no PR is detected, Protocol 2 MUST be
skipped as before.

#### Scenario: Explicit PR number, linked issues found

- **GIVEN** the user provides PR number as argument
- **AND** the PR body contains "Fixes #314"
- **WHEN** Protocol 2 runs during Phase 1c
- **THEN** issue #314 is fetched and acceptance criteria
  extracted
- **AND** the criteria are passed to the Guard persona in
  Step 2

#### Scenario: Auto-detected PR, linked issues found

- **GIVEN** the PR is auto-detected in Step 7
- **AND** the PR body contains "Fixes #314"
- **WHEN** the review body is assembled for posting
- **THEN** linked issue context is included in the posted
  review body as additional context

#### Scenario: No PR detected

- **GIVEN** no PR is detected for the current branch
- **WHEN** Phase 1c runs
- **THEN** Protocol 2 is skipped
- **AND** behavior is identical to current implementation

### Requirement: Provenance Disclosure

Every posted review MUST include the disclosure:
`_This review was generated by /review-council (AI-assisted)._`

#### Scenario: Review body footer

- **GIVEN** a review is being posted
- **WHEN** the review body is assembled
- **THEN** the footer includes the disclosure text

### Requirement: Scaffold Sync

The scaffold copy at
`internal/scaffold/assets/opencode/commands/review-council.md`
MUST be updated identically to the live copy at
`.opencode/commands/review-council.md`.

#### Scenario: Drift detection

- **GIVEN** the live copy has been updated with Step 7
- **WHEN** the scaffold copy is not updated
- **THEN** the drift detection test MUST fail

## MODIFIED Requirements

### Requirement: Phase 1c Review-Context Skill Consumption

Previously: Protocol 2 is always skipped with the note "This
protocol requires a PR body; `/review-council` is a local
pre-PR command with no PR body to parse."

Now: Protocol 2 is conditionally executed. When an explicit PR
number is provided via `$ARGUMENTS`, Protocol 2 runs during
Phase 1c and enriches the Guard persona's Step 2 delegation.
When the PR is auto-detected in Step 7, Protocol 2 context is
included in the posted review body only. When no PR is
available, Protocol 2 is skipped as before.

## REMOVED Requirements

None.
