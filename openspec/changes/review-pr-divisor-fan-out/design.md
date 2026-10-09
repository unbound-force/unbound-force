## Context

`/uf.review-pr` is the post-PR GitHub review command. It currently
delegates all analysis to a single `general` Task subagent in one
monolithic prompt. `/uf.review-council` uses a multi-agent Divisor
fan-out pattern: discover agents → load dispatch-advisor →
`plan_review_dispatch` → `invoke_agent` per agent → consolidate →
`finalize_review_dispatch`.

This design restructures `/uf.review-pr` to use the same pattern while
preserving its PR-specific features: CI check analysis, review state
awareness, fix-branch generation, and verdict-aligned PR posting.

## Goals / Non-Goals

### Goals
- Replace the single `general` subagent with multi-agent Divisor fan-out
- Move pre-flight, review-context, and convention-pack loading from
  subagent to parent, executed before agent dispatch
- Move diff fetch earlier so review-context Protocol 4 (issue linking)
  can run on the PR diff
- Preserve all existing PR-specific features: CI checks, review state,
  fix-branch, verdict-aligned PR posting
- Produce a structured council verdict via `finalize_review_dispatch`

### Non-Goals
- Changing `/uf.review-council` behavior
- Adding new Divisor agents or modifying existing agent manifests
- Modifying `plan_review_dispatch`, `invoke_agent`, or
  `finalize_review_dispatch` tool signatures
- Changing the PR posting mechanism or GitHub API integration

## Decisions

### 1. Pre-flight Mode: ci-aware

`review-council` uses `soft-gate` mode because it runs locally without
CI results. `review-pr` always has CI results from Step 3 (fetched from
GitHub Checks API), so `ci-aware` mode is appropriate — it validates
that CI results are available and surfaces failures without hard
blocking.

### 2. Diff Fetch Moved to Parent (Step 3.6)

In the current command, diff fetch happens inside the subagent (Step B).
It moves to parent Step 3.6 (before agent discovery) because
`review-context` Protocol 4 needs the diff to link issues referenced in
commit messages. This mirrors review-council's diff handling.

### 3. finalize_review_dispatch Command Enum: "review-council"

`finalize_review_dispatch` does not have a `review-pr` command variant.
Using `"review-council"` is the closest match — the tool's payload
schema, workflow result shape, and consolidation logic are identical
for both use cases. The input context distinguishes the two (PR vs
local).

### 4. Convention Pack Loading in Parent

Convention packs (default, go, typescript, severity, content) are
loaded in the parent and included in the prompt sent to each
`invoke_agent` run. This ensures every Divisor agent has access to the
same coding and severity standards without re-loading them.

### 5. Fix-Branch Preservation

The fix-branch capability (Step 8) is preserved at the end of the
command, after the council verdict is output. It remains an optional
step offered to the user.

### 6. Agent Discovery from File System

Like review-council, the command discovers available Divisor agents by
listing `.opencode/agents/divisor-*.md` files rather than hardcoding
an agent list. This keeps the command decoupled from agent inventory
changes.

## Pipeline Comparison

| Step | Current review-pr | Proposed review-pr |
|------|------------------|-------------------|
| 0 | Prerequisites | Same |
| 1 | Resolve PR Number | Same |
| 2 | Fetch PR Metadata | Same |
| 3 | CI Checks + Causality | Same |
| 3.5 | Diff Size Check | Same (advisory only) |
| 3.6 | — | **NEW**: Fetch diff (moved from subagent) |
| 3.7 | — | **NEW**: Pre-flight (ci-aware, moved from subagent) |
| 3.8 | — | **NEW**: Review-context skill (moved from subagent) |
| 3.9 | — | **NEW**: Convention packs (moved from subagent) |
| 3.10 | — | **NEW**: Fetch review state (moved from subagent) |
| 4 | (in subagent) | **NEW**: Discover divisor-* agents |
| 4a | — | **NEW**: Load dispatch-advisor skill |
| 4b | — | **NEW**: plan_review_dispatch |
| 4c | — | **NEW**: acquire_sibling_evidence |
| 5 | Task subagent (general) | invoke_agent per planned run |
| 6 | — | **NEW**: Consolidate + prepare lessons + finalize |
| 7 | Output (compact) | Output (council format) |
| 8 | Fix-branch | Fix-branch (preserved) |
| 9 | PR posting | PR posting (preserved) |

## Risks / Trade-offs

- **Token consumption**: Multi-agent fan-out will consume more tokens
  than a single subagent. This is an accepted trade-off for the
  increased review coverage and persona-specific scrutiny.
- **Latency**: Sequential `invoke_agent` calls add wall-clock time.
  Future work could parallelize independent agent runs.
- **No new tool signatures**: `finalize_review_dispatch` uses the
  existing `"review-council"` command enum. If a `"review-pr"` variant
  is added later, this command should be updated.