## Context

The pre-flight skill currently supports two execution
policies: `hard-gate` (run all, stop on first failure)
and `ci-aware` (skip tools CI already covers, don't
stop). `/review-council` uses `hard-gate`, which blocks
the entire review council on any failure -- including
failures pre-existing on `main`.

`/review-pr` has its own causality analysis (Step 3a)
that classifies failures as PR-caused vs pre-existing
by querying the GitHub CI API for base branch check
results. This logic lives in the review-pr command, not
in the pre-flight skill.

This design adds a third execution policy (`soft-gate`)
to the pre-flight skill that combines `hard-gate`'s
"run everything" behavior with causality classification
to determine which failures should block.

## Goals / Non-Goals

### Goals

- Add `soft-gate` execution policy to the pre-flight
  skill with baseline establishment and causality
  classification
- Update `/review-council` Phase 1a to use `soft-gate`
- Maintain backward compatibility for `hard-gate` and
  `ci-aware` consumers
- Classify failures using the same model as `/review-pr`
  Step 3a (branch-caused / pre-existing / unknown)

### Non-Goals

- Modifying `/unleash` -- it inherits the improvement
  via `/review-council` delegation; phase checkpoints
  remain `hard-gate`
- Modifying `/review-pr` -- it has its own causality
  analysis in Step 3a that works against PR check data
- Baseline result caching -- deferred to a follow-up
  issue if performance becomes a concern
- Error output comparison for causality classification
  -- exit-code-only classification matches `/review-pr`
  and is sufficient for v1
- Fix-branch creation for pre-existing failures --
  deferred; this may become a standalone command

## Decisions

### D1: New mode vs extending `hard-gate`

**Decision**: Add a new `soft-gate` mode rather than
extending `hard-gate` with optional parameters.

**Rationale**: The existing two modes have clean,
distinct behaviors. Adding optional baseline parameters
to `hard-gate` would make it behave differently depending
on whether baseline data is provided, breaking the
current clean contract. A third mode preserves the
principle that each mode has one unambiguous behavior.
This aligns with Composability First -- existing consumers
are unaffected.

### D2: Baseline establishment strategy

**Decision**: Two-tier baseline: (1) GitHub CI API first,
(2) local git worktree fallback.

**Rationale**: Most repositories using `/review-council`
have CI configured. Querying `gh api` for `main` branch
check results is fast and avoids running the full tool
suite twice. When `gh` is unavailable or returns no data
(no CI configured, private repo without token), the skill
falls back to creating a temporary git worktree of `main`,
running the same tools there, and comparing exit codes.
This ensures the feature works in all environments.

**Tier 1 -- CI API baseline**:
```bash
gh api repos/{owner}/{repo}/commits/main/check-runs \
  --jq '.check_runs[] | {name, conclusion}'
```
Maps CI check names to tool names using the same
coverage matrix logic from Phase 3.

**Tier 2 -- Local worktree baseline**:
```bash
git worktree add /tmp/preflight-baseline-<SHORT_SHA> \
  main --detach
# Run each failing tool in the worktree
# Compare exit codes
git worktree remove /tmp/preflight-baseline-<SHORT_SHA>
```
Only tools that failed on the branch need to be run
against the baseline -- passing tools are not
branch-caused by definition.

**Fallback to conservative**: If both tiers fail (e.g.,
`gh` unavailable AND worktree creation fails), classify
all failures as `unknown` and treat them as branch-caused
(conservative, matching `/review-pr` behavior).

### D3: Causality classification model

**Decision**: Reuse `/review-pr` Step 3a's classification
model exactly.

| Baseline status | Branch status | Classification |
|-----------------|---------------|----------------|
| Pass | Fail | **branch-caused** |
| Fail | Fail | **pre-existing** |
| No data | Fail | **unknown** (treat as branch-caused) |

**Rationale**: Consistency with `/review-pr`. Users see
the same classification vocabulary in both commands. The
`unknown` → branch-caused conservative default prevents
false negatives (missing a real branch-caused failure).

### D4: Gate behavior in `soft-gate` mode

**Decision**: `soft-gate` runs ALL tools (same as
`hard-gate`). After execution:

- **branch-caused failures**: STOP (hard gate). Report
  as CRITICAL findings. Do not proceed to council.
- **pre-existing failures**: CONTINUE. Report as
  informational findings. Proceed to council with the
  findings included in context.
- **unknown failures**: Treat as branch-caused (STOP).

This differs from `hard-gate` (stops on ANY failure)
and `ci-aware` (never stops, leaves gating to the
consuming command).

### D5: Execution order within `soft-gate`

**Decision**: Run all tools first, then classify. Do not
stop on first failure during execution.

**Rationale**: Unlike `hard-gate` which stops on first
failure, `soft-gate` needs to run all tools to build a
complete picture before establishing the baseline. A tool
that fails might be pre-existing, and stopping early
would prevent discovering branch-caused failures in later
tools. The baseline is only queried for tools that
actually failed, minimizing baseline overhead.

### D6: Result format extension

**Decision**: Extend Phase 5's result format with a
causality column in the Execution Results table and a
causality breakdown in the Verdict.

```
### Execution Results
| Tool | Command | Exit | Status | Causality |
|------|---------|------|--------|-----------|
| go test | go test ... | 1 | FAIL | pre-existing |
| golangci-lint | golangci-lint ... | 1 | FAIL | branch-caused |

### Verdict
- **Mode**: soft-gate
- **Result**: FAIL (branch-caused)
- **Branch-caused failures**: [golangci-lint]
- **Pre-existing failures**: [go test]
- **Baseline method**: CI API | worktree | unavailable
```

The consuming command uses the verdict to decide:
- If branch-caused failures exist → STOP
- If only pre-existing failures → CONTINUE with
  informational findings

### D7: Review-council report changes

**Decision**: Add a "Pre-existing CI Failures" section
to the final report (Step 6), between the discovery
summary and the iteration findings.

```
### Pre-existing CI Failures (informational)

The following failures exist on `main` and are
unrelated to the current branch:

| Tool | Exit code | Baseline method |
|------|-----------|-----------------|
| go test | 1 | CI API |

These do not block the review verdict.
```

## Risks / Trade-offs

### R1: Baseline execution adds latency

The CI API tier is fast (single HTTP request). The
worktree tier is slow (creates a worktree, runs tools,
cleans up). Mitigation: CI API is tried first, and the
worktree fallback only runs failing tools against the
baseline (not all tools). Caching is explicitly deferred
to a follow-up issue.

### R2: False pre-existing classification

If a tool fails on `main` for a different reason than it
fails on the branch, the failure is classified as
pre-existing when it may be branch-caused. Mitigation:
exit-code-only classification is the same approach
`/review-pr` uses successfully. Error output comparison
is deferred to a follow-up if this becomes a practical
problem.

### R3: Worktree creation may fail

Dirty working tree, insufficient disk space, or git
configuration issues could prevent worktree creation.
Mitigation: fallback to conservative classification
(all failures treated as branch-caused / unknown). The
user sees the same behavior as the current `hard-gate`
mode in this case.

### R4: `gh` CLI authentication

The CI API tier requires `gh` to be installed and
authenticated. Mitigation: binary availability is
checked via `which gh` before attempting the API call.
If unavailable, the skill falls through to the worktree
tier silently.

### R5: Flaky tests

A test that passes on `main` intermittently but fails on
the branch could be misclassified as branch-caused.
Mitigation: this is inherent to any causality analysis
based on point-in-time comparison. Accepted as a known
limitation, consistent with `/review-pr` behavior.
