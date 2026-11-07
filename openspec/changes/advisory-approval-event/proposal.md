## Why

`/uf.review-pr` posts every `APPROVE WITH ADVISORIES` verdict as a
`COMMENT`. That mapping exists so unresolved HIGH or CRITICAL
advisories cannot unblock a merge, but it also catches verdicts
whose advisories are only MEDIUM or LOW.

Now that a machine account posts reviews, those PRs end up with no
approval even though the council found nothing that needs a human.
Of the first 12 PRs it reviewed, 5 were advisory verdicts posted as
comments and 1 was approved. A PR left in that state cannot merge,
and each round of fixes starts another full review.

## What Changes

- `/uf.review-pr` posts `APPROVE WITH ADVISORIES` as `APPROVE` when
  no finding is CRITICAL or HIGH.
- A verdict with any CRITICAL or HIGH finding, or a finding of
  unknown severity, still posts as `COMMENT`, with a note saying
  those findings need human judgment before merge.
- Advisories stay in the review as inline comments in both cases.
- The confirmation gate picks its options from the mapped event, so
  a mapped approval uses the approve option set.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `review-pr-posting`: verdict to GitHub event mapping.

## Impact

- `.opencode/commands/uf.review-pr.md` and its scaffold copy
  `internal/scaffold/assets/opencode/commands/uf.review-pr.md`.
- `/uf.review-council` is unchanged: its advisory verdict applies
  to spec review, where this problem does not arise.
- Anyone running `/uf.review-pr` on another person's PR can now post
  an approval for a MEDIUM or LOW advisory verdict, after the
  existing confirmation gate.
