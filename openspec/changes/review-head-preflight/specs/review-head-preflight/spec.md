## ADDED Requirements

### Requirement: Head Root Preparation

The `uf-workflow` plugin MUST provide a `review_head_root` tool
whose `prepare` action returns a directory whose `HEAD` is the
given `head_sha`, together with the PR's changed files
(`base_sha...head_sha`) and the config commit. The tool MUST NOT
switch, stash, reset, or modify the user's branch, index, or
working tree. Git and gh MUST be invoked with argument arrays,
never through a shell, and path arguments MUST follow `--`.

#### Scenario: Current checkout already at the head
- **GIVEN** `HEAD` equals `head_sha` and
  `git status --porcelain --untracked-files=normal` is empty
- **WHEN** `prepare` is called
- **THEN** it MUST return the plugin directory as `root` with
  `method: "checkout"` and `created: false`, MUST NOT create a
  worktree, and MUST NOT apply the trust gate

#### Scenario: Untracked file at the head
- **GIVEN** `HEAD` equals `head_sha` and an untracked, non-ignored
  file exists
- **WHEN** `prepare` is called
- **THEN** it MUST NOT reuse the checkout

#### Scenario: Trusted PR on another commit
- **GIVEN** `HEAD` differs from `head_sha`, and the PR's head repo
  equals its base repo or its author association is OWNER,
  MEMBER, or COLLABORATOR
- **WHEN** `prepare` is called
- **THEN** it MUST add a detached worktree at `head_sha` under
  `.uf/review-head/<id>` and return it with `method: "worktree"`
- **AND** the user's branch, index, and working tree MUST be
  unchanged

#### Scenario: Untrusted fork PR
- **GIVEN** the head repo differs from the base repo and the author
  association is none of OWNER, MEMBER, or COLLABORATOR
- **WHEN** `prepare` is called on a checkout not at the head
- **THEN** it MUST fail with code `UNTRUSTED_HEAD`, fetch nothing,
  and create no worktree

#### Scenario: Commit missing locally
- **GIVEN** `head_sha` or `base_sha` is absent from the object store
- **WHEN** `prepare` is called
- **THEN** it MUST fetch it from the remote whose URL names the base
  repository's host and `owner/repo`, case-insensitively, then
  verify presence with `git cat-file -e`
- **AND** it MUST NOT read `FETCH_HEAD`

#### Scenario: Head force-pushed away
- **GIVEN** the fetch succeeds but `head_sha` is still absent
- **WHEN** `prepare` is called
- **THEN** it MUST fail with code `HEAD_MOVED` and create no
  worktree

#### Scenario: Head cannot be obtained
- **GIVEN** no remote matches the base repository, or the fetch or
  PR lookup fails
- **WHEN** `prepare` is called
- **THEN** it MUST fail with code `HEAD_UNAVAILABLE`, and fetch and
  lookup errors MUST be marked retryable

#### Scenario: Worktree creation fails
- **GIVEN** `git worktree add` exits non-zero
- **WHEN** `prepare` is called
- **THEN** it MUST fail with code `WORKTREE_FAILED`, include git's
  message, and leave no directory behind

#### Scenario: Worktree directory stays out of status
- **GIVEN** `prepare` created a worktree
- **WHEN** `git status --porcelain --untracked-files=normal` runs in
  the user's checkout
- **THEN** `.uf/review-head/` MUST NOT appear

#### Scenario: Stale worktrees from a crashed run
- **GIVEN** a `.uf/review-head/<id>` worktree older than 24 hours,
  one younger than 24 hours, and a user worktree elsewhere
- **WHEN** `prepare` is called
- **THEN** only the older review-head worktree MUST be removed

### Requirement: Head Root Validation and Release

A root MUST be accepted only when it is the plugin directory, or
exactly `<plugin dir>/.uf/review-head/<id>` with an 8-character
`[a-z0-9]` id that `git worktree list` reports. `release` MUST
remove only such worktrees.

#### Scenario: Releasing a created worktree
- **GIVEN** `prepare` returned a `created: true` root
- **WHEN** `release` is called with that root
- **THEN** the worktree and its directory MUST be removed
- **AND** a second `release` of it MUST fail with `INVALID_ROOT`

#### Scenario: Release from a restarted process
- **GIVEN** `prepare` ran in an earlier plugin process
- **WHEN** `release` is called with its root
- **THEN** it MUST succeed

#### Scenario: Releasing the reused checkout
- **GIVEN** `prepare` returned `method: "checkout"`
- **WHEN** `release` is called with that root
- **THEN** it MUST succeed without removing anything

#### Scenario: Path variants are refused
- **GIVEN** a live root spelled with a trailing slash, a `..`
  segment, or a symlink, or a path like
  `.uf/review-head/../../x`
- **WHEN** `release` is called with it
- **THEN** it MUST fail with code `INVALID_ROOT` and delete nothing

#### Scenario: Git remove fails
- **GIVEN** `git worktree remove` exits non-zero for a valid root
- **WHEN** `release` is called
- **THEN** the directory MUST still be removed and the call MUST
  succeed

### Requirement: Declared Checks Execution Root

`uf_preflight_declared_checks` MUST accept optional `root` and
`base_sha`. With `root`, checks MUST execute there. With
`base_sha`, diff scope and the soft-gate merge-base MUST derive
from `git merge-base <base_sha> HEAD` in the execution root.
Configuration and check executables MUST resolve from the plugin
directory.

#### Scenario: No root or base given
- **GIVEN** `root` and `base_sha` are omitted
- **WHEN** the tool runs
- **THEN** its output MUST be identical to the output before this
  change

#### Scenario: Root is a prepared head worktree
- **GIVEN** `root` is a valid review-head worktree and `base_sha`
  is the PR base
- **WHEN** the tool runs in `soft-gate` mode and a required check
  fails
- **THEN** the check MUST have run in `root`, scoped to the PR's
  changes, and its baseline MUST use that merge-base

#### Scenario: Root not valid
- **GIVEN** `root` fails validation
- **WHEN** the tool runs
- **THEN** it MUST fail with code `INVALID_ROOT` and run no check

### Requirement: PR-Mode Reviews Gate the Frozen Head

When `/uf.review-pr` or `/uf.review-council` reviews a PR, every
pre-flight phase MUST run against the prepared root:
- workflow parsing
- tool detection
- diff scope, taken from `changed_files`
- execution
- declared checks
- the baseline

Gaze MUST run only when the root is the reused checkout. The
report MUST state the head commit, the method, and the config
commit, or why local gates did not run.

#### Scenario: review-pr from main
- **GIVEN** the user is on `main` and the PR changes only `*.go`
  files
- **WHEN** `/uf.review-pr <N>` runs Step 3.7 and CI has no matching
  Go test check
- **THEN** `go test` MUST NOT be marked "SKIP — no in-scope files",
  it MUST run in the prepared root, and the report MUST show
  `Gates ran at: <head_sha>`

#### Scenario: review-council with a PR number
- **GIVEN** the user is on an unrelated branch with a failing test,
  and runs `/uf.review-council <N>` for a trusted PR
- **WHEN** Phase 1a runs
- **THEN** pre-flight MUST target the prepared root, and the user's
  failing test MUST NOT stop the council

#### Scenario: Gaze with a worktree root
- **GIVEN** `prepare` returned `method: "worktree"`
- **WHEN** Phase 1b runs
- **THEN** `gaze-reporter` MUST NOT run, and the report MUST note
  why

#### Scenario: review-council without a PR
- **GIVEN** `/uf.review-council` runs with local or uncommitted
  input
- **WHEN** Phases 1a and 1b run
- **THEN** they MUST run in the current checkout and MUST NOT call
  `review_head_root`

#### Scenario: Release once, non-fatally
- **GIVEN** `prepare` succeeded
- **WHEN** the command reaches its final finalization or any STOP
- **THEN** it MUST call `release` exactly once
- **AND** a release failure MUST be reported as a warning and MUST
  NOT change the verdict

### Requirement: Local Gates Never Run on the Wrong Tree

If `prepare` fails, a PR-mode review MUST NOT run local gates in
any other directory.

#### Scenario: review-pr prepare failure
- **GIVEN** `prepare` returns any failure code
- **WHEN** `/uf.review-pr` handles it
- **THEN** it MUST record `Local gates: not run (<code>)` and
  continue with CI evidence only

#### Scenario: review-council untrusted head
- **GIVEN** `prepare` returns `UNTRUSTED_HEAD`
- **WHEN** `/uf.review-council` handles it
- **THEN** it MUST record `Local gates: not run (UNTRUSTED_HEAD)`,
  skip Phases 1a and 1b, and continue to Phase 1c

#### Scenario: review-council prepare error
- **GIVEN** `prepare` returns `HEAD_MOVED`, `HEAD_UNAVAILABLE`, or
  `WORKTREE_FAILED`
- **WHEN** `/uf.review-council` handles it
- **THEN** it MUST record coverage `FAIL` and an `INCONCLUSIVE`
  verdict through the shared finalization protocol and dispatch no
  reviewer children
