## Why

`/review-council` Phase 1a runs the pre-flight skill in
`hard-gate` mode, which stops the entire review council on
any CI check failure -- including failures that already
exist on `main` and are unrelated to the current branch.
This blocks the council from providing value when the
codebase has pre-existing issues.

`/review-pr` solves this with causality analysis (Step 3a):
it classifies failures as PR-caused vs pre-existing and
only blocks on PR-caused failures. This change brings the
same causality classification to `/review-council` via a
new `soft-gate` execution policy in the pre-flight skill.

Fixes: #313 (split from #177, Phase 2)

## What Changes

1. Add a third execution policy (`soft-gate`) to the
   pre-flight skill that runs all tools locally but
   classifies failures as branch-caused vs pre-existing
   before deciding whether to stop.

2. Update `/review-council` Phase 1a to use `soft-gate`
   instead of `hard-gate`.

3. Branch-caused failures remain hard gates (CRITICAL).
   Pre-existing failures are reported as informational
   findings in the council report.

## Capabilities

### New Capabilities

- `soft-gate execution policy`: A new pre-flight mode
  that runs all tools (like `hard-gate`) but classifies
  failures by causality before gating. Branch-caused
  failures stop execution. Pre-existing failures are
  reported as informational findings.

- `baseline establishment`: The `soft-gate` mode
  establishes a baseline for causality classification
  using a two-tier strategy: (1) query GitHub CI API
  via `gh api` for `main` branch check results when
  available, (2) fall back to running checks in a
  temporary git worktree of `main` when `gh` is
  unavailable or no CI data exists.

### Modified Capabilities

- `/review-council` Phase 1a: Changes from `hard-gate`
  to `soft-gate` mode. Pre-existing failures no longer
  block the council. The final report includes a new
  section for pre-existing failures classified as
  informational.

### Removed Capabilities

- None.

## Impact

### Files Modified

- `.opencode/skills/pre-flight/SKILL.md` -- add
  `soft-gate` execution policy with baseline
  establishment, causality classification, and
  decision rules
- `internal/scaffold/assets/opencode/skills/pre-flight/SKILL.md`
  -- scaffold copy (kept in sync by
  `TestEmbeddedAssets_MatchSource`)
- `.opencode/commands/review-council.md` -- Phase 1a
  switches to `soft-gate` mode, final report adds
  pre-existing failure section
- `internal/scaffold/assets/opencode/commands/review-council.md`
  -- scaffold copy (kept in sync by
  `TestEmbeddedAssets_MatchSource`)

### Consumers Not Modified

- `/unleash`: Uses `hard-gate` for phase checkpoints
  (mid-implementation, where any failure should stop)
  and delegates to `/review-council` for code review
  (inherits `soft-gate` automatically). No direct
  changes needed.
- `/review-pr`: Uses `ci-aware` mode with its own
  causality analysis in Step 3a. Unaffected.

### Behavioral Expansion

This is an intentional behavioral expansion of the
pre-flight skill (adding a third mode) and of
`/review-council` (changing gate behavior for
pre-existing failures). This is NOT strict behavioral
parity -- the council will now proceed past pre-existing
failures where it previously stopped. Per the learning
from the pre-flight skill extraction, this expansion is
explicitly declared rather than obscured.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The pre-flight skill remains a self-contained artifact
consumed asynchronously by commands. The new `soft-gate`
mode produces the same standardized result format
(CI Coverage Matrix, Execution Results, Verdict) with
an additional causality classification field. No
runtime coupling is introduced.

### II. Composability First

**Assessment**: PASS

The new mode is additive -- `hard-gate` and `ci-aware`
continue to work unchanged for their existing consumers.
No mandatory dependencies are introduced. The `gh` CLI
is used opportunistically with a local worktree fallback,
so the feature works with or without GitHub CI.

### III. Observable Quality

**Assessment**: PASS

The `soft-gate` verdict includes machine-parseable
causality classification for each failure (branch-caused
vs pre-existing vs unknown). The result format extends
the existing standardized format with an additional
column. Provenance is maintained.

### IV. Testability

**Assessment**: PASS

The causality classification logic is deterministic: it
compares exit codes from the branch run against the
baseline. Verification is manual (run `/review-council`
with a known pre-existing failure and confirm it is
reported as informational rather than blocking). This
matches the regression verification approach used for the
pre-flight skill extraction and review-context skill
migration.

### V. Security by Default

**Assessment**: PASS

The `gh api` command uses `--arg` for check name
injection safety, consistent with `/review-pr` Step 3a.
The git worktree fallback creates a temporary directory
and cleans it up after use. No new dependencies are
introduced. No secrets or elevated permissions are
required.
