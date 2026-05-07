<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Expand Step 2 Metadata Fetch

- [x] 1.1 Add `reviewDecision,reviewRequests` to the
  `gh pr view --json` field list in Step 2 of
  `.opencode/commands/review-pr.md`

## 2. Add Step 7.5: Fetch Existing Reviews and Comments

- [x] 2.1 Add new "Step 7.5: Fetch Existing Review State"
  section between Step 7 and Step 8 in
  `.opencode/commands/review-pr.md`. Include `gh api`
  commands for fetching reviews and inline comments,
  token budget (3000 char cap), error handling
  (graceful degradation on 403/timeout).

## 3. Update Step 8 for Duplicate Finding Awareness

- [x] 3.1 Update Step 8 (AI Review) introduction in
  `.opencode/commands/review-pr.md` to instruct the
  AI to use existing review data from Step 7.5 for
  deduplication: annotate overlapping findings as
  "previously raised by @user", reference prior
  discussions, acknowledge resolved threads.

## 4. Update Step 11 Verdict Posting Warnings

- [x] 4.1 Add stale review warning to Step 11 in
  `.opencode/commands/review-pr.md`. Before posting
  APPROVE, check `dismiss_stale_reviews` via
  `gh api repos/{owner}/{repo}/branches/<base>/protection`.
  Display warning if enabled. Silently skip on 403/404.

- [x] 4.2 Add duplicate review detection to Step 11 in
  `.opencode/commands/review-pr.md`. Before posting,
  check if a review from the same account exists (from
  Step 7.5 data). Prompt user with context-aware
  message based on same/different verdict.

- [x] 4.3 Add CODEOWNER awareness to Step 11 in
  `.opencode/commands/review-pr.md`. When posting
  APPROVE with `require_code_owner_reviews: true`,
  check for CODEOWNERS file and warn if posting
  account may not satisfy requirement.

## 5. Sync Scaffold Asset

- [x] 5.1 Copy the updated `.opencode/commands/review-pr.md`
  to `internal/scaffold/assets/opencode/command/review-pr.md`
  to keep the scaffold copy in sync.

## 6. Update Dependabot Auto-Approval

- [x] 6.1 [P] Update `ci_dependencies.yml`
  `approve_dependabot_prs` job to check for existing
  REQUEST_CHANGES reviews from non-bot users before
  creating APPROVE review.

## 7. Document GitHub Review Lifecycle

- [x] 7.1 [P] Add "GitHub Review Lifecycle" subsection
  to the "PR Review" section of AGENTS.md covering:
  `dismiss_stale_reviews` behavior, the four review
  states, `requestedReviewers` vs submitted reviews,
  CODEOWNER requirements, review request lifecycle.

## 8. Verify Constitution Alignment

- [x] 8.1 Verify all changes maintain constitution
  alignment as documented in the proposal.
