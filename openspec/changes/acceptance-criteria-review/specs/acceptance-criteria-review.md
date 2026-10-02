## ADDED Requirements

### Requirement: Originating-Issue Threading

The `.openspec.yaml` change metadata SHALL carry an `originating_issue` field (positive integer, GitHub issue number) that identifies the GitHub issue from which this change originated. When present, review-time commands SHALL use this field as the primary source for resolving the originating issue's acceptance criteria.

#### Scenario: Originating issue field present in .openspec.yaml

- **GIVEN** a change directory at `openspec/changes/<name>/` with `.openspec.yaml` containing `originating_issue: 563`
- **WHEN** the review-context skill Protocol 2 resolves the originating issue for a review
- **THEN** the pipeline SHALL use issue #563 as the originating issue without parsing the PR body for issue references

#### Scenario: Originating issue field absent

- **GIVEN** a change directory at `openspec/changes/<name>/` with `.openspec.yaml` that does NOT contain an `originating_issue` field
- **WHEN** the review-context skill Protocol 2 resolves the originating issue for a review
- **THEN** the pipeline SHALL fall back to parsing the PR body for `Fixes #N` / `Closes #N` / `Resolves #N` references (existing behavior)

#### Scenario: Originating issue field present but PR body contains different issue references

- **GIVEN** `.openspec.yaml` containing `originating_issue: 563` AND a PR body containing `Fixes #999`
- **WHEN** the review-context skill Protocol 2 resolves the originating issue
- **THEN** the pipeline SHALL use issue #563 (from `.openspec.yaml`) as the authoritative originating issue and SHALL NOT use #999

### Requirement: Acceptance-Criteria Evaluation

When the originating issue is resolved, the review pipeline SHALL fetch its acceptance criteria and evaluate the PR against them. Each criterion SHALL be individually assessed with a per-criterion status and evidence from the diff.

#### Scenario: Rigorous evaluation with Given/When/Then scenarios

- **GIVEN** an originating issue with acceptance criteria containing Given/When/Then scenarios
- **WHEN** the review pipeline evaluates the PR against those criteria
- **THEN** each scenario SHALL be assessed as SATISFIED, NOT SATISFIED, or PARTIAL
- **AND** each assessment SHALL include evidence from the diff (file paths, line references, or "no evidence found")

#### Scenario: Best-effort evaluation with freeform criteria

- **GIVEN** an originating issue with acceptance criteria that do NOT contain Given/When/Then scenarios
- **WHEN** the review pipeline evaluates the PR against those criteria
- **THEN** each criterion SHALL be assessed as COVERED, NOT COVERED, or PARTIAL
- **AND** each assessment SHALL include a brief justification

#### Scenario: No originating issue resolved

- **GIVEN** no originating issue could be resolved (neither from `.openspec.yaml` nor PR body)
- **WHEN** the review pipeline attempts acceptance-criteria evaluation
- **THEN** the pipeline SHALL skip criteria evaluation
- **AND** SHALL note "no originating issue resolved — criteria evaluation skipped" in the output

#### Scenario: Originating issue fetch fails

- **GIVEN** an originating issue number is resolved but `gh issue view` returns 404, 403, or times out
- **WHEN** the review pipeline attempts to fetch the issue
- **THEN** the pipeline SHALL skip criteria evaluation for that issue
- **AND** SHALL note "originating issue fetch failed" in the output
- **AND** SHALL NOT block the review

### Requirement: Implementation Deviation Escape Hatch

When a PR intentionally does NOT satisfy one or more acceptance criteria, the reviewer SHALL surface an `IMPLEMENTATION_DEVIATION` finding that routes the deviation to governance rather than silently passing or blocking without resolution.

#### Scenario: Intentional deviation from acceptance criterion

- **GIVEN** an acceptance criterion assessed as NOT SATISFIED or NOT COVERED
- **AND** the PR description or commit messages document an intentional deviation with a governance action (e.g., "amend PRD section X", "file follow-up issue #Y")
- **WHEN** the reviewer evaluates the deviation
- **THEN** the reviewer SHALL emit an `IMPLEMENTATION_DEVIATION` finding with severity HIGH
- **AND** the finding SHALL include: the criterion reference, the reason for deviation, and the required governance action
- **AND** the finding SHALL NOT be auto-fixable

#### Scenario: Unexplained missing criterion (not an intentional deviation)

- **GIVEN** an acceptance criterion assessed as NOT SATISFIED or NOT COVERED
- **AND** the PR description does NOT document an intentional deviation or governance action
- **WHEN** the reviewer evaluates the criterion
- **THEN** the reviewer SHALL emit a standard alignment finding (not IMPLEMENTATION_DEVIATION) with severity based on the criterion's importance
- **AND** the finding SHALL indicate that the criterion is not addressed and no deviation governance is documented

#### Scenario: Deviation with documented governance action

- **GIVEN** an `IMPLEMENTATION_DEVIATION` finding where the governance action is documented in the PR description (e.g., "Follow-up issue #1234 filed to amend PRD section 3.2")
- **WHEN** the review verdict is computed
- **THEN** the IMPLEMENTATION_DEVIATION finding SHALL still be reported in the output
- **AND** the finding SHALL note that the governance action is documented
- **AND** the verdict SHALL be COMMENT (not REQUEST CHANGES) if the governance action is adequate and no other REQUEST CHANGES findings exist

## MODIFIED Requirements

### Requirement: Review-Context Protocol 2 (Issue Linking)

The review-context skill Protocol 2 SHALL resolve the originating issue using this priority order:
1. `originating_issue` field in `.openspec.yaml` (authoritative when present).
2. PR body parsing for `Fixes #N` / `Closes #N` / `Resolves #N` references (fallback).
3. No originating issue (skip criteria evaluation).

Previously: Protocol 2 resolved linked issues solely from PR body parsing.

#### Scenario: Protocol 2 with originating_issue field

- **GIVEN** `.openspec.yaml` containing `originating_issue: 563`
- **WHEN** Protocol 2 executes
- **THEN** Protocol 2 SHALL resolve issue #563 as the originating issue
- **AND** SHALL NOT parse the PR body for issue references (the field is authoritative)

#### Scenario: Protocol 2 without originating_issue field

- **GIVEN** `.openspec.yaml` without an `originating_issue` field
- **WHEN** Protocol 2 executes
- **THEN** Protocol 2 SHALL parse the PR body for issue references (existing behavior)

### Requirement: Review-PR Step F.1 (Alignment Check — Issue Criteria Coverage)

Step F.1 SHALL perform rigorous evaluation (SATISFIED / NOT SATISFIED / PARTIAL per scenario) when Given/When/Then scenarios are available, and best-effort evaluation (COVERED / NOT COVERED / PARTIAL) otherwise. Step F.1 SHALL emit `IMPLEMENTATION_DEVIATION` findings for intentional deviations with documented governance actions.

Previously: Step F.1 reported per-criterion status (COVERED / NOT COVERED / PARTIAL) without distinguishing evaluation modes or providing a deviation escape hatch.

#### Scenario: Rigorous evaluation in Step F.1

- **GIVEN** linked issue acceptance criteria with Given/When/Then scenarios
- **WHEN** Step F.1 evaluates issue criteria coverage
- **THEN** each scenario SHALL be assessed as SATISFIED, NOT SATISFIED, or PARTIAL with diff evidence

#### Scenario: IMPLEMENTATION_DEVIATION in Step F.1

- **GIVEN** a criterion assessed as NOT SATISFIED with a documented governance action in the PR description
- **WHEN** Step F.1 evaluates the criterion
- **THEN** Step F.1 SHALL emit an `IMPLEMENTATION_DEVIATION` finding with severity HIGH

### Requirement: Review-Council Step 2 (Guard Persona Delegation)

The Guard persona SHALL receive the originating issue's acceptance criteria (when resolved) for drift detection and SHALL evaluate whether unmet criteria have an `IMPLEMENTATION_DEVIATION` with adequate governance.

Previously: The Guard persona received acceptance criteria for drift detection but did not evaluate deviation governance.

#### Scenario: Guard persona receives criteria with deviation

- **GIVEN** originating issue criteria resolved and at least one criterion assessed as NOT SATISFIED with an IMPLEMENTATION_DEVIATION finding
- **WHEN** the Guard persona evaluates the review
- **THEN** the Guard persona SHALL verify that the IMPLEMENTATION_DEVIATION includes a documented governance action
- **AND** SHALL flag missing governance as a HIGH-severity finding
