## Context

Maintainers manually triage GitHub issues today. The
Divisor review panel (adversary, architect, guard, sre,
testing) already evaluates PRs through specialized lenses.
This design extends that pattern to issue triage via a new
`/triage-issue` slash command.

The design follows three established patterns:
- `/review-council` for multi-agent parallel fan-out and
  verdict consolidation
- `/address-feedback` for interactive user confirmation
  before GitHub actions
- `divisor-curator` for `gh` CLI issue operations and
  duplicate checking

## Goals / Non-Goals

### Goals
- Structured multi-agent issue classification with
  per-agent provenance
- Duplicate detection before any action
- Compound issue splitting with cross-references
- Tone-appropriate commenting (warm for rejections,
  factual for valid issues)
- Machine-parseable triage artifact for downstream
  consumption
- Graceful degradation when fewer than five agents
  are available

### Non-Goals
- Batch triage (multiple issues per invocation)
- Cross-repo triage (only current repo)
- Auto-closing issues (never, by design)
- Modifying existing agent files
- Creating new Divisor agents

## Decisions

### D1: Four-Phase Pipeline

**Decision**: Follow the address-feedback four-phase
pattern: Ingest, Assess, Classify, Act.

**Rationale**: Proven pattern in the codebase. Separates
data gathering (Ingest) from analysis (Assess/Classify)
from side effects (Act). Each phase has clear inputs and
outputs. The Act phase gates all GitHub mutations behind
user confirmation.

### D2: Majority Verdict (3/5) Not Unanimous

**Decision**: Use majority consensus (3 of 5 agents)
for validity determination, not the unanimous gate used
by review-council.

**Rationale**: Issues are more nuanced than code review.
A single dissenting agent should not block the entire
panel. The review-council uses unanimous APPROVE because
merging bad code is costly and irreversible. Issue
classification is lower-stakes and reversible (labels can
be changed, comments edited). Dissenting opinions are
recorded in the artifact for transparency.

### D3: Category Taxonomy

**Decision**: Seven categories with deterministic label
mapping.

| Category             | GitHub Label         |
|----------------------|----------------------|
| `bug`                | `bug`                |
| `feature`            | `enhancement`        |
| `enhancement`        | `enhancement`        |
| `question`           | `question`           |
| `opinion`            | `design-discussion`  |
| `duplicate`          | `duplicate`          |
| `needs-clarification`| `needs-info`         |

**Rationale**: Categories map to standard GitHub labels
where possible. `feature` and `enhancement` both map to
`enhancement` (GitHub convention). `opinion` gets a new
`design-discussion` label to distinguish philosophical
disagreements from invalid issues. `needs-clarification`
uses `needs-info` to signal the reporter should provide
more context.

### D4: Labels Applied Without Confirmation (with Exception)

**Decision**: Labels are applied automatically without
user confirmation. Exception: the `duplicate` label
requires user confirmation because it carries implicit
"close" semantics. Comments and child issues also
require user confirmation.

**Rationale**: Most labels are low-risk and reversible
(one `gh issue edit --remove-label` away). The
`duplicate` label is different -- it signals to the
community that the issue should be closed, which is
a higher-stakes action. Comments are public-facing and
irreversible in practice. Child issues create new work
items. The risk/reversibility gradient drives the
confirmation policy.

### D5: Objectivity Classification

**Decision**: Binary classification (objective vs.
subjective) based on agent consensus.

- **Objective**: At least one agent provides verifiable
  evidence (reproducible bug, measurable performance
  issue, documented behavior contradiction)
- **Subjective**: All agents agree the issue is
  preference-based (naming, style, philosophy)

**Rationale**: This mirrors the address-feedback
`data-driven` vs. `subjective` classification. The
threshold is deliberately low for objective (any single
agent can elevate) because false negatives (dismissing a
real issue as subjective) are more costly than false
positives.

### D6: Comment Tone Tiers

**Decision**: Three comment tone tiers composed by the
command itself, not by individual agents.

| Validity        | Tone                                                                        |
|-----------------|-----------------------------------------------------------------------------|
| Valid            | Factual analysis, recommendations                                           |
| Invalid/Opinion  | Warm acknowledgment; design rationale; alternatives; invite engagement       |
| Needs-Clarify    | Specific questions; what info would help                                    |

**Rationale**: Individual Divisor agents are blunt by
design (adversarial reviewers). Public-facing issue
comments require a different register. The command
synthesizes agent findings into a unified comment with
appropriate tone. This separation of concerns keeps
agents honest and comments kind.

### D7: Artifact Schema with Envelope Wrapping

**Decision**: Produce a JSON artifact at
`.uf/artifacts/issue-triage/issue-<N>.json` wrapped in
the standard envelope format.

**Rationale**: Aligns with Principle III (Observable
Quality). The envelope provides provenance metadata
(hero, version, timestamp). The payload schema
(`schemas/issue-triage/v1.0.0.schema.json`) enables
schema validation and cross-hero consumption. Mx F can
track triage patterns, Muti-Mind can detect recurring
gaps.

### D8: Dynamic Agent Discovery

**Decision**: Discover available triage agents at runtime
by reading `.opencode/agents/` and filtering for the five
target agent files.

**Rationale**: Follows the review-council pattern.
Avoids hardcoding agent availability. Supports graceful
degradation (Principle II: Composability First) -- the
command works with any subset of the five agents.
Note: `divisor-curator` is excluded from the triage
panel because its domain (documentation gap detection
and content pipeline triage) is not relevant to issue
classification.

### D9: Duplicate Check Strategy

**Decision**: Two-layer duplicate detection.

1. **Phase 1 (Ingest)**: Extract keywords from issue
   title and body, search via
   `gh issue list --search "<keywords>" --state open`
2. **Phase 2 (Assess)**: Agents independently evaluate
   whether the issue is a duplicate based on the
   candidate list from Phase 1

If Phase 1 finds candidates AND 2+ agents classify as
duplicate, the issue is marked `duplicate`.

**Rationale**: Keyword search alone produces false
positives. Agent assessment alone lacks search reach.
The two-layer approach combines broad search with
expert judgment.

### D10: Split Issue Cross-Referencing

**Decision**: When splitting, each child issue body
includes a reference to the parent issue
(`Split from #N`). The parent issue receives a comment
listing all created child issues with their numbers
and titles.

**Rationale**: Maintains traceability. The parent issue
is never auto-closed -- the reporter confirms the split
addresses their concerns.

### D11: Idempotent Re-Run

**Decision**: The command is safely re-runnable. On
re-invocation, it detects previously applied labels and
posted comments to avoid duplication. Artifacts use
round numbers to preserve history.

**Rationale**: Follows the address-feedback pattern
(round-numbered artifacts). The Act phase performs
multiple sequential GitHub mutations (label, comment,
child issues). If any step fails, the user should be
able to re-run without duplicating completed actions.
Label application is inherently idempotent
(`--add-label` on an already-labeled issue is a no-op).
Comment duplication is detected by checking for the
Divisor panel footer.

### D12: Keyword Sanitization for Duplicate Search

**Decision**: Keywords extracted from issue content are
sanitized before use in `gh issue list --search`.
Shell metacharacters and CLI flag prefixes are stripped.

**Rationale**: Issue content is attacker-controlled
(any GitHub user can create an issue). Keywords flow
into a shell command. Without sanitization, a malicious
issue title could inject shell commands or `gh` CLI
flags. This follows the same principle as FR-018
(temp file + `--input`) but applied to search queries.

## Risks / Trade-offs

### R1: Agent Disagreement on Category

**Risk**: Five agents may return five different
categories for the same issue.

**Mitigation**: Category resolution uses a specificity
hierarchy (bug > feature > enhancement >
needs-clarification > opinion > question). `duplicate`
is resolved independently by FR-013. Disagreements
are recorded in the artifact and surfaced to the user
during the Act phase. The user can override via MODIFY.

### R2: False Duplicate Detection

**Risk**: Keyword search may surface unrelated issues
as duplicates.

**Mitigation**: Duplicate classification requires both
keyword match AND 2+ agent agreement. Per D4, the
`duplicate` label is the one label that requires user
confirmation before application (it carries implicit
"close" semantics). Keywords extracted from issue
content are sanitized before use in search queries
(FR-012).

### R3: Tone Mismatch for Edge Cases

**Risk**: The three-tier tone model may not cover all
cases (e.g., a valid bug report with an aggressive tone
from the reporter).

**Mitigation**: The user reviews and can MODIFY the
comment before posting. The command never posts
autonomously. Edge cases are handled by human judgment
in the confirmation step.

### R4: Rate Limiting

**Risk**: Five parallel agent invocations plus multiple
`gh` API calls could hit rate limits on large repos.

**Mitigation**: Agent calls use the Task tool (managed
concurrency). GitHub API calls are sequential within
each phase. The command does not retry on rate limit
errors -- it reports the error and stops.

### R5: Label Creation

**Risk**: The `design-discussion` label may not exist
in the target repository.

**Mitigation**: The command checks for label existence
before applying. If the label does not exist, it creates
it via `gh label create "design-discussion"
--description "Subjective design philosophy discussion"
--color "d4c5f9"`. Same pattern for any missing labels.
