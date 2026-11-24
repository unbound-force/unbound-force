## ADDED Requirements

### Requirement: Head Root Preparation

The `uf-workflow` plugin MUST provide a `review_head_root` tool
whose `prepare` action returns a directory whose `HEAD` is the
given `head_sha`. The tool MUST NOT switch, stash, reset, or
otherwise modify the user's checkout. Git and gh MUST be invoked
with argument arrays, never through a shell.

#### Scenario: Current checkout already at the head
- **GIVEN** `HEAD` equals `head_sha` and no tracked file is
  modified
- **WHEN** `prepare` is called
- **THEN** it MUST return the plugin directory as `root` with
  `method: "checkout"` and `created: false`, and MUST NOT create a
  worktree

#### Scenario: Checkout on another commit
- **GIVEN** `HEAD` differs from `head_sha` and the head commit is
  present locally
- **WHEN** `prepare` is called
- **THEN** it MUST add a detached worktree at `head_sha` under the
  tool-owned temp prefix and return it with `method: "worktree"`
  and `created: true`
- **AND** the user's branch, index, and working tree MUST be
  unchanged

#### Scenario: Dirty checkout at the head
- **GIVEN** `HEAD` equals `head_sha` but a tracked file is modified
- **WHEN** `prepare` is called
- **THEN** it MUST create a worktree rather than reuse the checkout

#### Scenario: Head commit missing locally
- **GIVEN** the head commit object is absent
- **WHEN** `prepare` is called
- **THEN** it MUST fetch `refs/pull/<N>/head` from the remote that
  names the PR's base repository and verify the fetched SHA before
  creating the worktree

#### Scenario: PR pushed after the review froze its input
- **GIVEN** the fetched `refs/pull/<N>/head` differs from
  `head_sha`
- **WHEN** `prepare` is called
- **THEN** it MUST fail with code `HEAD_MOVED` and create no
  worktree

#### Scenario: Head cannot be obtained
- **GIVEN** no remote matches the base repository, or the fetch
  fails
- **WHEN** `prepare` is called
- **THEN** it MUST fail with code `HEAD_UNAVAILABLE`, and a fetch
  error MUST be marked retryable

#### Scenario: Worktree creation fails
- **GIVEN** `git worktree add` exits non-zero
- **WHEN** `prepare` is called
- **THEN** it MUST fail with code `WORKTREE_FAILED`, include git's
  message, and leave no temp directory behind

### Requirement: Head Root Release

The `release` action MUST remove only worktrees that `prepare`
created in the same process.

#### Scenario: Releasing a created worktree
- **GIVEN** `prepare` returned a `created: true` root
- **WHEN** `release` is called with that root
- **THEN** the worktree and its temp parent MUST be removed and the
  root MUST no longer be accepted by either tool

#### Scenario: Releasing the reused checkout
- **GIVEN** `prepare` returned `method: "checkout"`
- **WHEN** `release` is called with that root
- **THEN** it MUST succeed without removing anything

#### Scenario: Releasing an unknown path
- **GIVEN** a path `prepare` did not return
- **WHEN** `release` is called with it
- **THEN** it MUST fail with code `INVALID_ROOT` and delete
  nothing

#### Scenario: Stale worktrees from a crashed run
- **GIVEN** a tool-owned worktree older than 24 hours is
  registered
- **WHEN** `prepare` is called
- **THEN** that worktree SHOULD be removed before any new one is
  created

### Requirement: Declared Checks Execution Root

`uf_preflight_declared_checks` MUST accept an optional `root`.
When `root` is given, checks MUST execute with that working
directory, and diff scope and the soft-gate baseline MUST derive
from that root's `HEAD`. Configuration and check executables MUST
still resolve from the plugin directory.

#### Scenario: No root given
- **GIVEN** `root` is omitted
- **WHEN** the tool runs
- **THEN** its behavior and output MUST be identical to the
  behavior before this change

#### Scenario: Root is a prepared head worktree
- **GIVEN** `root` is a live `review_head_root` worktree
- **WHEN** the tool runs in `soft-gate` mode and a required check
  fails
- **THEN** the check MUST have run in `root`, and its baseline
  MUST use the merge-base of `root`'s `HEAD`

#### Scenario: Root not issued by the tool
- **GIVEN** `root` is neither the plugin directory nor a live
  prepared root
- **WHEN** the tool runs
- **THEN** it MUST fail with code `INVALID_ROOT` and run no check

### Requirement: PR-Mode Reviews Gate the Frozen Head

When `/uf.review-pr` or `/uf.review-council` reviews an explicit or
resolved PR, every local gate MUST run against the prepared head
root. This covers pre-flight detected tools, declared checks, and
the baseline. Gaze MUST run only when that root is the reused
checkout. The report MUST state the commit and method
the gates used.

#### Scenario: review-pr from main
- **GIVEN** the user is on `main` and runs `/uf.review-pr <N>`
- **WHEN** Step 3.7 runs local tools
- **THEN** they MUST run in the root `prepare` returned for
  `head_sha`, and the report MUST show `Gates ran at: <head_sha>`

#### Scenario: review-council with a PR number
- **GIVEN** the user is on an unrelated branch and runs
  `/uf.review-council <N>`
- **WHEN** Phases 1a and 1b run
- **THEN** pre-flight MUST target the prepared root, and a failure
  that exists only on the user's branch MUST NOT stop the council

#### Scenario: Gaze with a worktree root
- **GIVEN** `prepare` returned `method: "worktree"`
- **WHEN** Phase 1b runs
- **THEN** `gaze-reporter` MUST NOT run, and the report MUST note
  that Gaze was skipped because the head is in a worktree

#### Scenario: Gaze with the reused checkout
- **GIVEN** `prepare` returned `method: "checkout"` and `gaze` is
  installed
- **WHEN** Phase 1b runs
- **THEN** `gaze-reporter` MUST run as it does today

#### Scenario: review-council without a PR
- **GIVEN** `/uf.review-council` runs with local or uncommitted
  input
- **WHEN** Phases 1a and 1b run
- **THEN** they MUST run in the current checkout and MUST NOT
  call `review_head_root`

#### Scenario: Release on every exit
- **GIVEN** `prepare` succeeded
- **WHEN** the command completes, STOPs, or finalizes
  `INCONCLUSIVE`
- **THEN** it MUST call `release` with the prepared root before
  finalization

### Requirement: Fail Closed When the Head Is Unavailable

If `prepare` fails, a PR-mode review MUST NOT run local gates in
any other directory.

#### Scenario: review-pr prepare failure
- **GIVEN** `prepare` returns any failure code
- **WHEN** `/uf.review-pr` handles it
- **THEN** it MUST STOP with the code and its remediation and post
  no review

#### Scenario: review-council prepare failure
- **GIVEN** `prepare` returns any failure code
- **WHEN** `/uf.review-council` handles it
- **THEN** it MUST record coverage `FAIL` and an `INCONCLUSIVE`
  verdict through the shared finalization protocol and dispatch no
  reviewer children
