## Context

The `/finale` command is a Markdown slash command at
`.opencode/command/finale.md` with a scaffold copy at
`internal/scaffold/assets/opencode/command/finale.md`.
It defines a 9-step workflow. Step 7 calls
`gh pr merge <number> --rebase --delete-branch`.

## Goals / Non-Goals

### Goals
- PRs remain open after `/finale` for human review
- User returns to main after CI passes (ready to work
  on other things)
- Summary clearly states the PR is open, not merged

### Non-Goals
- Adding a separate `/merge` command (users can run
  `gh pr merge` directly or merge via GitHub UI)
- Changing CI check behavior (step 6 stays)
- Changing commit or push behavior (steps 1-5 stay)

## Decisions

**D1: Remove step 7 entirely, keep steps 6 and 8**

Step 6 (watch CI) provides immediate feedback on
whether the PR is green before the author walks away.
Step 8 (return to main) lets the author start other
work. Only step 7 (merge) is removed because that's
what should wait for reviewer approval.

**D2: Update summary to say "ready for review"**

The step 9 summary changes from reporting "merged via
rebase" to "CI passed, ready for review" with a next
step prompt to request reviewers.

**D3: Remove merge-related guardrails**

The guardrails "NEVER merge with failing checks" and
"ALWAYS use rebase merge" become irrelevant since
`/finale` no longer merges. Remove them to avoid
confusion.

## Risks / Trade-offs

- **Trade-off**: Users must merge manually after review.
  This adds one extra step but enables proper code
  review. The `gh pr merge --rebase --delete-branch`
  command is documented in the summary for convenience.
