## ADDED Requirements

### FR-001: soft-gate Execution Policy

The pre-flight skill MUST support a `soft-gate`
execution policy in addition to the existing `hard-gate`
and `ci-aware` policies. The consuming command specifies
which mode to use.

In `soft-gate` mode, the skill MUST:
1. Run ALL detected and available tools (same as
   `hard-gate`). Do NOT stop on first failure during
   execution.
2. After execution, establish a baseline for the default
   branch to classify each failure.
3. Classify each failing tool as `branch-caused`,
   `pre-existing`, or `unknown`.
4. Gate only on `branch-caused` or `unknown` failures.
   `pre-existing` failures MUST NOT gate.

#### Scenario: Branch-caused failure blocks council

- **GIVEN** a branch with a lint error not present on
  `main`
- **WHEN** the pre-flight skill runs in `soft-gate`
  mode
- **THEN** the lint failure is classified as
  `branch-caused`
- **AND** the verdict is FAIL with
  `branch-caused failures: [golangci-lint]`
- **AND** the consuming command MUST NOT proceed past
  the pre-flight gate

#### Scenario: Pre-existing failure does not block

- **GIVEN** a branch where `go test` fails, and the
  same test also fails on `main`
- **WHEN** the pre-flight skill runs in `soft-gate`
  mode
- **THEN** the test failure is classified as
  `pre-existing`
- **AND** the verdict is PASS (no branch-caused
  failures)
- **AND** the pre-existing failure is reported as an
  informational finding
- **AND** the consuming command MAY proceed past the
  pre-flight gate

#### Scenario: Mixed failures

- **GIVEN** a branch where `go test` fails
  (pre-existing on `main`) and `golangci-lint` fails
  (not on `main`)
- **WHEN** the pre-flight skill runs in `soft-gate`
  mode
- **THEN** `go test` is classified as `pre-existing`
- **AND** `golangci-lint` is classified as
  `branch-caused`
- **AND** the verdict is FAIL (branch-caused failures
  exist)
- **AND** both failures are listed in the verdict with
  their classification

#### Scenario: All tools pass

- **GIVEN** a branch where all tools pass
- **WHEN** the pre-flight skill runs in `soft-gate`
  mode
- **THEN** no baseline establishment is needed
- **AND** the verdict is PASS
- **AND** the behavior is identical to `hard-gate`
  all-pass

### FR-002: Baseline Establishment (CI API Tier)

When running in `soft-gate` mode, the skill MUST first
detect the repository's default branch (via
`git symbolic-ref refs/remotes/origin/HEAD`, falling
back to checking for `main` then `master`). The skill
MUST NOT hardcode `main`. If the default branch cannot
be detected, the skill MUST fall through to the
conservative fallback (all failures classified as
`unknown`).

The skill MUST then attempt to establish a baseline
via the GitHub CI API:
1. Check `gh` CLI availability via `which gh`.
2. If available, query the latest check-run results for
   the default branch:
   ```bash
   gh api repos/{owner}/{repo}/commits/${DEFAULT_BRANCH}/check-runs \
     --jq '.check_runs[] | {name, conclusion}'
   ```
3. Use `--arg` for any dynamic values to prevent
   injection (consistent with `/review-pr` Step 3a).
4. Map CI check names to local tool names using the
   existing coverage matrix logic from Phase 3.

If `gh` is not available, or the API call returns no
data, or the API call fails, the skill MUST fall through
to FR-003 (worktree baseline).

#### Scenario: CI API provides baseline

- **GIVEN** `gh` is installed and authenticated
- **AND** the default branch has CI check results
- **WHEN** a tool fails on the branch
- **THEN** the skill queries the CI API for the matching
  check on the default branch
- **AND** classifies the failure based on the CI API
  result

#### Scenario: gh CLI not available

- **GIVEN** `gh` is not installed (not in PATH)
- **WHEN** a tool fails on the branch
- **THEN** the skill skips the CI API tier silently
- **AND** proceeds to FR-003 (worktree baseline)

### FR-003: Baseline Establishment (Worktree Tier)

When the CI API tier is unavailable or returns no data,
the skill MUST fall back to establishing a baseline via
a temporary git worktree.

The skill MUST:
1. Create a temporary detached worktree of the default
   branch (as detected in FR-002):
   ```bash
   git worktree add /tmp/preflight-baseline-<SHORT_SHA> \
     ${DEFAULT_BRANCH} --detach
   ```
   where `<SHORT_SHA>` is the first 8 characters of the
   default branch's HEAD commit SHA.
2. Run ONLY the tools that failed on the branch in the
   worktree directory. Tools that passed on the branch
   MUST NOT be run against the baseline.
3. Compare exit codes: if a tool also fails in the
   worktree, classify as `pre-existing`; if it passes,
   classify as `branch-caused`.
4. Clean up the worktree after classification:
   ```bash
   git worktree remove \
     /tmp/preflight-baseline-<SHORT_SHA> --force
   ```

If worktree creation fails, the skill MUST classify all
failures as `unknown` and treat them as branch-caused
(conservative fallback).

#### Scenario: Worktree baseline for pre-existing failure

- **GIVEN** `gh` is not available
- **AND** `go test` fails on the branch
- **WHEN** the skill creates a worktree of the default
  branch and runs `go test` there
- **AND** `go test` also fails in the worktree
- **THEN** the failure is classified as `pre-existing`

#### Scenario: Worktree baseline for branch-caused failure

- **GIVEN** `gh` is not available
- **AND** `golangci-lint` fails on the branch
- **WHEN** the skill creates a worktree of the default
  branch and runs `golangci-lint` there
- **AND** `golangci-lint` passes in the worktree
- **THEN** the failure is classified as `branch-caused`

#### Scenario: Worktree creation fails

- **GIVEN** `gh` is not available
- **AND** worktree creation fails (disk space, dirty
  state, etc.)
- **WHEN** a tool fails on the branch
- **THEN** the failure is classified as `unknown`
- **AND** `unknown` is treated as `branch-caused`
  (conservative)

### FR-004: Extended Result Format

The pre-flight Phase 5 result format MUST be extended
for `soft-gate` mode to include causality classification.

The Execution Results table MUST include a `Causality`
column:

```
| Tool | Command | Exit | Status | Causality |
|------|---------|------|--------|-----------|
```

The Verdict section MUST include:
- `branch-caused failures` list
- `pre-existing failures` list
- `baseline method` used (CI API / worktree /
  unavailable)

The `Result` field in the Verdict MUST be:
- `PASS` if no branch-caused or unknown failures exist
  (even if pre-existing failures exist)
- `FAIL (branch-caused)` if any branch-caused or unknown
  failures exist

#### Scenario: Verdict with pre-existing only

- **GIVEN** all failures are classified as
  `pre-existing`
- **THEN** the verdict Result is `PASS`
- **AND** the pre-existing failures list is populated

## MODIFIED Requirements

### Phase 1a of `/review-council` (review-council.md)

Phase 1a MUST use `soft-gate` mode instead of
`hard-gate` mode.

Previously: "Load the `pre-flight` skill and run in
`hard-gate` mode"

New behavior:
1. Load the pre-flight skill and run in `soft-gate`
   mode.
2. If the verdict is FAIL (branch-caused): STOP
   immediately. Report each branch-caused failure as a
   CRITICAL finding. Do NOT proceed to Phase 1b.
3. If the verdict is PASS (no branch-caused failures):
   proceed to Phase 1b. If pre-existing failures exist,
   record them for inclusion in the final report
   (Step 6).

#### Scenario: Council proceeds past pre-existing failure

- **GIVEN** a branch where `go test` fails but the
  failure is pre-existing on `main`
- **WHEN** `/review-council` runs Phase 1a in
  `soft-gate` mode
- **THEN** the council proceeds to Phase 1b and
  Divisor agent delegation
- **AND** the final report (Step 6) includes a
  "Pre-existing CI Failures" section listing the
  failure as informational

### Step 6 of `/review-council` (Final Report)

The final report MUST include a "Pre-existing CI
Failures" section when pre-existing failures were
detected in Phase 1a.

Previously: No pre-existing failure section existed.

New behavior: Between the discovery summary and the
iteration findings, include:

```
### Pre-existing CI Failures (informational)

The following failures exist on `main` and are
unrelated to the current branch:

| Tool | Exit code | Baseline method |
|------|-----------|-----------------|
| ... | ... | ... |

These do not block the review verdict.
```

This section MUST be omitted when no pre-existing
failures were detected.

## REMOVED Requirements

None.
