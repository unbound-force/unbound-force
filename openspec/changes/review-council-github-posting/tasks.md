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

  All tasks modify the same file
  (.opencode/commands/review-council.md) so no tasks
  are parallel-eligible. The scaffold copy sync (3.1)
  depends on all prior tasks completing first.
-->

## 1. PR Detection and Protocol 2 Unlock

- [x] 1.1 Add PR detection logic to `/review-council` Code
  Review Mode. After Step 6 (Final Report), add Step 7
  preamble that detects an open PR via
  `gh pr view --json number,headRefName,baseRefName`.
  Parse `$ARGUMENTS` for an explicit PR number (integer
  after mode keyword). Validate PR number as positive
  integer (digits only, range 1-999999); reject
  non-numeric or out-of-range values with informational
  error. If no PR found and no number provided, skip
  Step 7 with informational note. If `gh`
  unavailable/unauthenticated, skip with informational
  note.
  **File**: `.opencode/commands/review-council.md`

- [x] 1.2 Modify Phase 1c (review-context skill invocation)
  to conditionally run Protocol 2 when an explicit PR
  number is provided via `$ARGUMENTS`. Fetch PR body via
  `gh pr view <N> --json body --jq '.body'`. Pass
  linked issue context to Guard persona prompt in Step 2.
  When no explicit PR number is available, skip Protocol 2
  as before. For auto-detected PRs (Step 7), linked issue
  context is included only in the posted review body.
  **File**: `.opencode/commands/review-council.md`

## 2. Review State and Pre-Posting Checks

- [x] 2.1 Add review state fetching (transplant from
  `/review-pr` Step 7.5). Fetch existing reviews, inline
  comments, and current user login via `gh api`. Apply
  same error handling: skip sub-step on 403/404/timeout.
  Cap existing comments at 3000 characters.
  **File**: `.opencode/commands/review-council.md`

- [x] 2.2 Add pre-posting checks (transplant from
  `/review-pr` Step 11a). Implement duplicate review
  detection using fetched review state. For APPROVE
  verdicts only: fetch branch protection settings, warn
  on stale review dismissal, warn on CODEOWNER
  requirements.
  **File**: `.opencode/commands/review-council.md`

## 3. Finding Aggregation and Posting

- [x] 3.1 Add multi-persona finding aggregation. Structure
  review body with council summary header (verdict,
  reviewer list, iteration count) and per-persona
  sections. Summarize LOW findings (count only). Include
  MEDIUM+ findings with full text. Handle consolidated
  cross-persona findings under primary persona with
  attribution. Add provenance disclosure footer.
  **File**: `.opencode/commands/review-council.md`

- [x] 3.2 Add inline comment preparation with severity-first
  round-robin allocation. Collect file-specific findings
  from all personas, sort by severity, round-robin across
  personas within same tier, cap at 15. Use GitHub
  suggestion block format for concrete single-file fixes.
  Overflow to review body summary.
  **File**: `.opencode/commands/review-council.md`

- [x] 3.3 Add verdict mapping and human confirmation.
  Map council verdict to GitHub API event type (APPROVE →
  `APPROVE`, REQUEST CHANGES → `REQUEST_CHANGES`,
  APPROVE WITH ADVISORIES → `COMMENT` with explanatory
  note). Show prepared review content via AskUserQuestion
  with options: post, skip, edit, change verdict. NEVER
  post without explicit confirmation.
  **File**: `.opencode/commands/review-council.md`

- [x] 3.4 Add posting mechanism. Write JSON payload to temp
  file, post via `gh api repos/{owner}/{repo}/pulls/<N>/reviews
  --method POST --input <file>`. Implement graceful
  degradation: on 403/422 fall back to COMMENT event
  type; on second failure inform user of permission issue.
  Remove temp file after posting.
  **File**: `.opencode/commands/review-council.md`

## 4. Sync and Verification

- [x] 4.1 Copy updated `.opencode/commands/review-council.md`
  to `internal/scaffold/assets/opencode/commands/review-council.md`
  to maintain scaffold sync. Verify byte-identical with
  diff.
  **File**: `internal/scaffold/assets/opencode/commands/review-council.md`

- [x] 4.2 Verify constitution alignment: confirm the
  implemented Step 7 maintains opt-in behavior
  (Composability, Principle II), includes per-persona
  provenance in posted reviews (Observable Quality,
  Principle III), uses JSON-to-tempfile for shell
  injection prevention (Security, Principle V), and
  preserves local-only behavior when no PR exists.

## 5. Documentation

- [x] 5.1 [P] Update `CHANGELOG.md` with entry under
  Unreleased documenting the new GitHub review posting
  capability, Protocol 2 unlock, and pre-posting checks.
  Reference `Fixes: #314`.
  **File**: `CHANGELOG.md`

- [x] 5.2 [P] File a documentation issue in
  `unbound-force/website` for the `/review-council`
  GitHub posting capability. This is a user-facing change
  that requires documentation updates per the
  cross-repo documentation gate.
  **Note**: Token lacks permission to create issues in
  `unbound-force/website`. Issue must be filed manually
  before PR merge.
<!-- spec-review: passed -->
<!-- code-review: passed -->
