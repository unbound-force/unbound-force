<!--
  originating_issue: 113
-->

## Why

The `/uf.review-pr` command currently delegates all analysis to a single
`general` Task subagent. This monolithic delegation lacks persona-specific
scrutiny (security, architecture, testing, alignment) and produces a single
compact summary instead of a structured, consolidatable council verdict. By
contrast, `/uf.review-council` already uses the dispatch-advisor skill with
`plan_review_dispatch` + `invoke_agent` to fan out review across multiple
Divisor agents, then consolidates findings via `finalize_review_dispatch`.

Aligning `/uf.review-pr` with this pattern eliminates a divergent code path,
increases review coverage, and ensures PR reviews benefit from the same
multi-agent scrutiny as pre-PR council reviews.

## What Changes

- Replace the single `general` Task subagent (Steps A-F in current command)
  with a multi-agent Divisor fan-out pipeline matching review-council's
  pattern: discover divisor-* agents → load dispatch-advisor →
  `plan_review_dispatch` → `invoke_agent` per agent → consolidate →
  `finalize_review_dispatch`
- Move pre-flight, review-context, convention packs, and AI review steps
  from the subagent prompt into the parent command, executed before dispatch
- Move diff fetch earlier in the pipeline so review-context Protocol 4
  (issue linking from diff) can run
- Load the dispatch-advisor skill in the parent command
- Preserve all PR-specific steps: CI checks + causality, diff size check,
  review state fetching, fix-branch offer, and verdict-aligned PR posting

## Capabilities

### New Capabilities
- `multi-agent-pr-review`: Review PR diffs with multiple Divisor agents
  (adversary, architect, guard, testing, sre, scribe) instead of a single
  general-purpose agent
- `pr-council-consolidation`: Produce structured council verdicts for PR
  reviews via `finalize_review_dispatch`

### Modified Capabilities
- `pr-review`: Pipeline reordered so pre-flight (ci-aware), review-context,
  and convention-pack loading happen in the parent before agent dispatch

### Removed Capabilities
- `single-agent-pr-analysis`: The monolithic subagent prompt that performed
  all analysis in one agent context

## Impact

- File: `.opencode/commands/uf.review-pr.md` (full rewrite of Steps 3-7 to
  match review-council fan-out pattern)
- No Go code changes required (this is a command markdown file only)
- No CI or test changes required (command invocation pattern does not change)

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

This change makes `/uf.review-pr` produce the same structured
`finalize_review_dispatch` artifact as `/uf.review-council`, with provenance
metadata (agent identity, run results, verdict). The artifact is
self-describing and consumable without consulting any producing hero.

### II. Composability First

**Assessment**: PASS

Each Divisor agent (adversary, architect, guard, etc.) remains independently
invocable via `invoke_agent`. The command orchestrates them through
well-defined extension points (`plan_review_dispatch`, dispatch-advisor)
without modifying agent internals. No hero requires another as a hard
prerequisite.

### III. Observable Quality

**Assessment**: PASS

The `finalize_review_dispatch` output includes structured findings with
severity, category, file, line, and run-level provenance. The council verdict
(APPROVE/REQUEST CHANGES/APPROVE WITH ADVISORIES) is machine-parseable and
backed by per-agent evidence.

### IV. Testability

**Assessment**: PASS

Each `invoke_agent` run is isolated with its own bounded context. The
dispatch plan is deterministic (`plan_review_dispatch`). The command's
behavior can be verified by inspecting the `finalize_review_dispatch`
output artifact, which includes per-run status, findings, and verdict.

### V. Security by Default

**Assessment**: PASS

No new dependencies are introduced. The change uses existing tools
(`plan_review_dispatch`, `invoke_agent`, `finalize_review_dispatch`,
dispatch-advisor skill) already trusted in the review-council workflow.
The parent command retains pre-flight validation in ci-aware mode.