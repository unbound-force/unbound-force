## Why

The review council and review-pr commands already extract acceptance criteria from linked issues (review-context skill Protocol 2, review-pr Step F.1), but the validation is best-effort and ad-hoc. There is no formal mechanism to:

1. Thread the originating-issue identity as machine-readable state through the OpenSpec pipeline to review time, so reviewers know *which* issue's criteria to evaluate against without re-parsing PR bodies.
2. Rigorously evaluate a PR against Given/When/Then acceptance criteria when the intake-kit is in the flow (currently per-criterion status is reported but not enforced).
3. Surface an explicit `IMPLEMENTATION_DEVIATION` path when a PR intentionally does NOT satisfy an acceptance criterion — currently, unmet criteria are flagged as findings but there is no structured route to governance (amend the PRD/spec) instead of silently passing or blocking.

This addresses GitHub issue #563: "Review council should validate that PR addresses the originating issue's acceptance criteria."

## What Changes

Three coordinated changes across the review pipeline:

1. **Originating-issue threading**: Extend `.openspec.yaml` (building on the `originating_issue` field from #554/PR #630) to carry the originating GitHub issue number as machine-readable state. The review-context skill (Protocol 2) and review-pr command consume this field to resolve the originating issue without relying solely on PR body parsing.

2. **Acceptance-criteria grounding**: When the originating issue is known, the review pipeline fetches its acceptance criteria and evaluates the PR against them. Rigorous mode (when the intake-kit / structured criteria with Given/When/Then scenarios is available): each scenario is individually assessed as SATISFIED / NOT SATISFIED / PARTIAL, with evidence from the diff. Best-effort mode (freeform criteria or no intake-kit): criteria are extracted and assessed as COVERED / NOT COVERED / PARTIAL (existing behavior, now formalized).

3. **Deviation escape hatch**: When a PR intentionally does NOT satisfy one or more acceptance criteria, the reviewer surfaces an `IMPLEMENTATION_DEVIATION` finding. This finding:
   - Identifies which criterion is intentionally unmet and why.
   - Routes to governance: requires amending the PRD/spec (or filing a follow-up issue) before merge.
   - Is classified as HIGH severity (governance bypass risk).
   - Is NOT auto-fixable — it requires human decision.

## Capabilities

### New Capabilities
- `originating-issue-threading`: Machine-readable originating-issue identity carried through `.openspec.yaml` to review time.
- `acceptance-criteria-evaluation`: Rigorous per-scenario evaluation of PR against originating issue acceptance criteria.
- `implementation-deviation`: Structured escape hatch for intentional acceptance-criteria deviations that routes to governance.

### Modified Capabilities
- `review-context-protocol-2`: Extended to consume `originating_issue` from `.openspec.yaml` when available, falling back to PR body parsing.
- `review-pr-step-f1`: Enhanced to perform rigorous Given/When/Then evaluation when structured criteria are available, and to surface `IMPLEMENTATION_DEVIATION` findings.
- `review-council-step-2`: Guard persona receives originating-issue criteria for drift detection and deviation surfacing.

### Removed Capabilities
None.

## Impact

- **Commands**: `.opencode/commands/uf.review-council.md`, `.opencode/commands/uf.review-pr.md`
- **Skills**: `.opencode/skills/review-context/SKILL.md` (Protocol 2)
- **Scaffold**: `.openspec.yaml` template (originating_issue field documentation)
- **Council review action**: `council-review-action/scripts/` (prompt preparation may need to thread originating-issue context)
- **No Go code changes**: This change is entirely to agent commands, skills, and pipeline documentation.

## Dependencies

- **GitHub issue #554 / PR #630**: Introduces the `originating_issue` field in `.openspec.yaml`. This change builds on that field — it MUST be merged (or this change must include the field definition if #630 is not yet merged at implementation time).
- **GitHub issue #539**: Separate concern (not addressed here).
- **Separation**: This change is explicitly SEPARATE from #539 and #554. It consumes the `originating_issue` field but does not modify its schema definition.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The originating-issue threading uses artifact-based communication (`.openspec.yaml` field) rather than runtime coupling. Review commands consume the field asynchronously from the file. No synchronous interaction between heroes is introduced.

### II. Composability First

**Assessment**: PASS

The acceptance-criteria evaluation degrades gracefully: when `originating_issue` is absent, the pipeline falls back to PR body parsing (existing behavior). No mandatory dependency on #554's field is introduced — it is an enhancement, not a prerequisite. Each review command remains independently usable.

### III. Observable Quality

**Assessment**: PASS

The `IMPLEMENTATION_DEVIATION` finding is machine-parseable (structured finding type with severity, criterion reference, and governance route). Per-criterion evaluation results (SATISFIED / NOT SATISFIED / PARTIAL) are machine-parseable and reproducible. Provenance is maintained via the originating-issue reference.

### IV. Testability

**Assessment**: PASS

The acceptance-criteria evaluation logic is testable in isolation: given a set of criteria and a diff, the evaluation produces deterministic results. The deviation escape hatch is a finding classification rule — testable via prompt-based test fixtures. No external services required.
