## Why

On PR #957, `/uf.review-pr` and `/uf.address-feedback` ran 5 full
rounds each and the PR collected 35 commits and about 60 inline
findings. Four things kept the loop going:

- `/uf.address-feedback` fixed every finding, LOW included, and each
  fix earned another full review that found more LOW items.
- Every re-review ran the full Divisor council, even when the push
  only answered the previous round's findings.
- `/uf.address-feedback` also ran `/uf.review-council` before each
  push, so every fix was reviewed by two councils.
- 10 of the 35 commits only added Dewey learnings. Each was another
  push for the next review to cover, and reviewers spent findings on
  the learnings themselves.

## What Changes

- `/uf.address-feedback` changes code only for CRITICAL and HIGH
  findings. MEDIUM and LOW findings are filed as issues or answered
  with a reply, and their threads are resolved. A finding with no
  severity label, such as most human comments, counts as HIGH.
- `/uf.review-pr` gains a re-review evaluator, the new
  `resolve_review_mode` plugin tool. When the current user
  already reviewed an earlier commit and the push since then is
  small, contained, and touches no sensitive path, it runs a light
  check instead of the council: it confirms the earlier CRITICAL and
  HIGH findings are fixed, reads only the new commits, and can
  approve. A new CRITICAL or HIGH problem escalates to the council.
- `/uf.address-feedback` no longer runs `/uf.review-council` before
  pushing. It runs the build, test, and lint commands instead, and
  the re-review is the one council for the cycle.
- Dewey learnings stay off feature and PR branches. `uf_finale_stage`
  no longer stages `.uf/dewey/learnings/`, and the learnings reach
  `main` through a separate `chore/dewey-learnings-*` PR, which
  gh-hud already opens once a day per checkout.

## Capabilities

### New Capabilities

- `address-feedback-scope`: which findings change code.
- `review-pr-re-review`: full council or light check on a re-review.
- `learnings-pr-scope`: where Dewey learnings are committed.

### Modified Capabilities

None in `openspec/specs`. `learnings-pr-scope` reverses the rule set
by `dewey-learning-pr-policy` and `finale-dewey-commit`, which this
change moves to the archive without merging their specs.

## Impact

- `.opencode/commands/uf.address-feedback.md`, `uf.review-pr.md`,
  `uf.finale.md`, `uf.unleash.md`, and their scaffold copies.
- `.opencode/plugins/uf-finale/index.ts`,
  `.opencode/plugins/review-dispatch/index.ts`, and their scaffold
  copies.
- `docs/usage.md`, `docs/architecture.md`, and
  `schemas/feedback-triage/README.md`.
- `AGENTS.md` Commit scope and Documentation gate rules.
- Option-label pins in `internal/scaffold/scaffold_test.go` and the
  `uf_finale_stage` tests.
- Repositories that refresh their scaffold pick up the new commands
  and staging behaviour. Without gh-hud, learnings need a learnings
  PR opened by hand.
