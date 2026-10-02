## Context

The review pipeline already extracts acceptance criteria from linked GitHub issues (review-context skill Protocol 2, review-pr Step F.1). However, the extraction is ad-hoc: it relies on parsing PR bodies at review time, has no structured way to express intentional deviations from criteria, and does not leverage the `originating_issue` field being introduced in `.openspec.yaml` by #554/PR #630.

The review-council command (Phase 1c, Protocol 2) and review-pr command (Step C, Protocol 2) both call into the review-context skill for issue linking. The Guard persona in review-council Step 2 already receives acceptance criteria for drift detection. The review-pr Step F.1 already reports per-criterion status (COVERED / NOT COVERED / PARTIAL).

This change formalizes and strengthens these existing capabilities into a coherent acceptance-criteria validation pipeline with a governance escape hatch.

## Goals / Non-Goals

### Goals
- Thread originating-issue identity through `.openspec.yaml` so review-time commands can resolve the issue without re-parsing PR bodies.
- Evaluate PR changes against each acceptance criterion individually, using Given/When/Then scenarios when available.
- Provide a structured `IMPLEMENTATION_DEVIATION` finding type that routes intentional criterion non-satisfaction to governance (PRD/spec amendment) rather than silently passing.
- Maintain backward compatibility: when `originating_issue` is absent, fall back to existing PR body parsing.

### Non-Goals
- Modifying the `originating_issue` field schema (that is #554's scope).
- Enforcing acceptance-criteria satisfaction as a hard merge gate (this change surfaces deviations; governance decides).
- Changing the review-council or review-pr verdict mapping (APPROVE / REQUEST CHANGES / COMMENT semantics are unchanged).
- Addressing #539 (separate concern).

## Decisions

### D1: Originating-issue resolution order

The pipeline resolves the originating issue using this priority:
1. `originating_issue` field in `.openspec.yaml` (machine-readable, authoritative).
2. PR body parsing via Protocol 2 (existing fallback).
3. No originating issue (skip criteria evaluation, note in output).

**Rationale**: Artifact-based resolution (D1.1) aligns with constitution principle I (Autonomous Collaboration) — the issue identity is stored in a file artifact, not derived from runtime state. The fallback chain (D1.2) aligns with principle II (Composability First) — the pipeline works without the field.

### D2: Evaluation modes

Two evaluation modes based on criteria structure:
- **Rigorous mode**: When acceptance criteria contain Given/When/Then scenarios (structured or extracted from intake-kit), each scenario is individually assessed as SATISFIED / NOT SATISFIED / PARTIAL with diff evidence.
- **Best-effort mode**: When criteria are freeform (no Given/When/Then), criteria are assessed as COVERED / NOT COVERED / PARTIAL (existing behavior, now formalized).

The mode is auto-detected from the criteria structure, not user-configured.

**Rationale**: Rigorous mode provides stronger guarantees when the data is available. Best-effort mode preserves existing behavior for unstructured criteria. Auto-detection avoids configuration burden.

### D3: IMPLEMENTATION_DEVIATION finding type

A new finding type `IMPLEMENTATION_DEVIATION` is introduced:
- Severity: HIGH (governance bypass risk).
- Structure: criterion reference, reason for deviation, required governance action (amend PRD/spec or file follow-up issue).
- Auto-fixable: NO (requires human decision).
- Effect on verdict: Counts as a REQUEST CHANGES finding unless the governance action is documented in the PR description.

**Rationale**: Without a structured deviation path, reviewers either silently pass unmet criteria (governance bypass) or block the PR without a clear resolution path. The deviation finding makes the gap explicit and actionable.

### D4: Scope of changes

Changes are limited to agent commands, skills, and pipeline documentation. No Go code changes.

**Rationale**: The review pipeline is implemented as agent instructions (markdown commands and skills), not compiled code. The council-review-action scripts may need minor adjustments to thread originating-issue context into prompts, but the core logic is prompt-based.

## Risks / Trade-offs

### R1: False positives in criteria evaluation
AI-based criteria evaluation may incorrectly assess a criterion as NOT SATISFIED when the implementation addresses it indirectly. Mitigation: the evaluation includes diff evidence, and the developer can respond with an IMPLEMENTATION_DEVIATION or clarify the mapping.

### R2: Originating-issue field not yet merged
If #630 is not merged at implementation time, this change must either wait or include the field definition. Mitigation: the fallback chain (D1) means the change works without the field; the `originating_issue` threading is an enhancement, not a prerequisite.

### R3: Deviation escape hatch could be abused
Developers might use IMPLEMENTATION_DEVIATION to bypass criteria without proper governance. Mitigation: the finding is HIGH severity, requires a documented governance action, and is visible in the review output. The review-council Guard persona checks for deviation governance completeness.
