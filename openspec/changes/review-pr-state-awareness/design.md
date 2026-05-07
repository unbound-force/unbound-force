## Context

`/review-pr` currently has 11 steps. Review state
awareness requires inserting a new step between Step 7
(Load Convention Packs) and Step 8 (AI Review), plus
modifications to Steps 2 and 11. The dependabot workflow
change is isolated to `ci_dependencies.yml`.

The current step flow:

```
0. Prerequisites
1. Resolve PR Number
2. Fetch PR Metadata
3. Fetch CI Check Results
4. Run Local Tools
5. Fetch Diff
6. Locate Spec + Linked Issues
7. Load Convention Packs
8. AI Review
9. Output Format
10. Fix-Branch Offer
11. Verdict Posting
```

## Goals / Non-Goals

### Goals
- Fetch existing reviews and inline comments before AI
  analysis to prevent duplicate findings
- Warn about stale review dismissal when posting APPROVE
- Detect and warn about duplicate reviews from same
  account
- Warn when APPROVE may not satisfy CODEOWNER
  requirements
- Make dependabot auto-approval idempotent (skip when
  REQUEST_CHANGES exists)
- Document GitHub review lifecycle for agent awareness

### Non-Goals
- Automatic review dismissal or deletion
- Cross-repo review aggregation
- Review assignment or routing logic
- Automated re-review after new commits
- Changes to `/review-council` (separate issue #177)

## Decisions

1. **New Step 7.5 (not renumbering)**: Insert the review
   state fetch as "Step 7.5" in the prose, positioned
   between Steps 7 and 8. This avoids renumbering all
   subsequent steps which would create a large diff and
   risk breaking references in AGENTS.md and other docs.

2. **Token budget for existing comments**: Cap existing
   review comments at 3000 characters total when passing
   to Step 8. Truncate oldest comments first. This
   prevents token bloat on PRs with extensive prior
   discussion.

3. **Deduplication strategy**: Use file path + line
   number matching for inline comments, and body text
   similarity for top-level review comments. Matching
   findings are annotated as "previously raised by
   @user" rather than fully suppressed — the AI may have
   additional context.

4. **Branch protection API availability**: The
   `dismiss_stale_reviews` check requires the branch
   protection API. If the API returns 404 (no branch
   protection) or 403 (insufficient permissions), skip
   the stale review warning silently. This is a
   nice-to-have warning, not a gate.

5. **Dependabot idempotency**: Check for
   REQUEST_CHANGES reviews from non-bot users before
   auto-approving. This is a conservative check — if a
   human has explicitly requested changes, the bot
   should not override that intent.

## Risks / Trade-offs

- **API rate limiting**: Adding 2-3 additional `gh api`
  calls per review increases API usage. Risk is low —
  these are lightweight reads, not writes.
- **Stale data**: Review state is fetched at review
  start. If reviews are posted/dismissed during the
  review, the data may be stale. Acceptable — the
  review is a point-in-time analysis.
- **Token budget**: Existing comments consume tokens
  from the AI review context window. The 3000-character
  cap mitigates this but may miss relevant context on
  heavily-discussed PRs.
- **Permission variance**: Not all tokens have access to
  branch protection API. Graceful degradation handles
  this but means stale review warnings may not appear
  for all users.
