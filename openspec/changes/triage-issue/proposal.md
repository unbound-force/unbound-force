## Why

GitHub issues arrive in many forms: genuine bugs, feature
requests mislabeled as bugs, subjective disagreements with
design philosophy, questions that belong in discussions, and
compound issues that should be split. Today maintainers
manually read, classify, and respond to each one. This is
time-consuming, inconsistent, and prone to tone missteps
when declining an issue.

The Divisor review panel already has the domain expertise
to evaluate issues across security, architecture,
governance, operations, and testing. This change extends
their reach from PR review to issue triage, bringing
structured multi-agent assessment to the issue lifecycle.

## What Changes

A new `/triage-issue` slash command that accepts a GitHub
issue number and runs a five-agent Divisor panel to:

1. Classify the issue (bug, feature, enhancement,
   question, opinion, duplicate)
2. Determine validity (valid, invalid, needs
   clarification)
3. Assess objectivity (objective defect vs. subjective
   preference)
4. Check for duplicates against open issues
5. Recommend splitting compound issues into focused
   child issues
6. Post a triage comment with the analysis and
   recommendations
7. Apply appropriate labels automatically

A new `issue-triage` JSON Schema captures the triage
artifact for consumption by Mx F, Muti-Mind, and future
analytics.

## Capabilities

### New Capabilities
- `/triage-issue <N>`: Multi-agent issue triage with
  structured classification, duplicate detection, split
  recommendations, and tone-appropriate commenting
- `issue-triage` schema (v1.0.0): Machine-parseable
  artifact for triage outcomes with full agent assessment
  provenance

### Modified Capabilities
- None

### Removed Capabilities
- None

## Impact

- **New files**:
  - `.opencode/commands/triage-issue.md` (slash command)
  - `schemas/issue-triage/v1.0.0.schema.json` (payload
    schema)
  - `schemas/issue-triage/README.md` (schema docs)
  - `schemas/issue-triage/samples/sample-issue-triage.json`
    (valid sample)
  - `schemas/issue-triage/samples/invalid-extra-property.json`
    (invalid sample for validation testing)
- **No modified files**: The command is self-contained.
  Agents discover commands by filename. The schema
  directory is a flat registry.
- **GitHub interactions**: Reads issues via `gh issue view`,
  searches for duplicates via `gh issue list --search`,
  applies labels via `gh issue edit --add-label`, posts
  comments via `gh api`, creates child issues via
  `gh issue create`
- **New GitHub label**: `design-discussion` for issues
  classified as subjective opinion/philosophy differences

## Constitution Alignment

Assessed against the Unbound Force org constitution v1.2.0.

### I. Autonomous Collaboration

**Assessment**: PASS

The command produces a self-describing JSON artifact
(`.uf/artifacts/issue-triage/issue-<N>.json`) wrapped in
the standard envelope format with full provenance metadata.
Each agent assessment is recorded independently. Mx F can
consume triage artifacts for trend analysis without
consulting the producing agents. The five Divisor agents
operate asynchronously via parallel Task tool fan-out,
with no synchronous inter-agent coupling.

### II. Composability First

**Assessment**: PASS

The command gracefully degrades when fewer than five
Divisor agents are available. If only three agents are
deployed, the command proceeds with those three and notes
which are missing. The command requires no other heroes
(Muti-Mind, Mx F, Gaze) to function. Those heroes
benefit from consuming the artifact but are not
prerequisites.

### III. Observable Quality

**Assessment**: PASS

The triage artifact is JSON with a formal schema
(`schemas/issue-triage/v1.0.0.schema.json`), includes
provenance metadata (producing hero, timestamp, repo
context), and is validated against the schema. Valid
and invalid samples enable automated regression testing.
The artifact includes per-agent assessments so every
classification decision is traceable to its source.

### IV. Testability

**Assessment**: PASS

The schema includes valid and invalid samples for
validation testing. The command uses `gh` CLI exclusively
(testable via mock runners, consistent with
`internal/sync/sync.go` patterns). Agent assessments
are independently verifiable. The slash command is a
declarative instruction file with no compiled code,
so testability concerns are limited to schema validation
and integration testing of the `gh` CLI interactions.

### V. Security by Default

**Assessment**: PASS

All GitHub comment text is written to temporary files
and posted via `--input`, preventing shell injection
(consistent with `address-feedback` and `review-pr`
guardrails). The command uses dynamic repo detection
(`gh repo view --json nameWithOwner`) rather than
hardcoded values. No new dependencies are introduced.
Labels are applied via `gh issue edit`, which respects
the authenticated user's permissions.
