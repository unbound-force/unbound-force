## Context

Issue-generating agents (triage, tasks-to-issues,
uf.init taskstoissues, divisor-curator) produce issue
content via LLM generation. Without explicit guardrails,
the LLM defaults to org-config-specific terminology
regardless of the target repository's actual type. This
is a systemic problem that affects all issue-generating
paths.

## Goals / Non-Goals

### Goals
- Prevent LLM-generated issue content from using
  org-config-specific terminology for non-org-config repos
- Add convention pack rules that all agents inherit
- Add targeted guardrails at each issue-generating
  component
- Regression test coverage for all guardrails

### Non-Goals
- Modifying existing issue content (this is forward-looking)
- Changing non-issue-generating agents
- Detecting org-config-specific terms at runtime (this
  uses instruction-level guardrails, not content filters)

## Decisions

### D1: Convention pack rules (systemic layer)

Add CP-001, CP-002, CP-003 to the default convention pack.
These rules use RFC 2119 language (MUST/SHOULD) and provide
generic equivalents for common org-config-specific terms.

**Rationale**: Convention packs are loaded by all agents
automatically, providing a systemic baseline that catches
new issue-generating paths without requiring per-component
changes.

### D2: Targeted guardrails (component layer)

Add explicit content portability sections to each of the
four issue-generating components, cross-referencing the
convention pack rules.

**Rationale**: Defense in depth. Convention pack rules may
be overlooked by agents focused on their primary task.
Placing the guardrail adjacent to the issue creation
instructions ensures it is in context when the LLM
generates issue content.

### D3: Scaffold asset sync

Both live copies (`.opencode/`) and canonical scaffold
copies (`internal/scaffold/assets/opencode/`) must be
updated in sync. Drift detection tests enforce this.

**Rationale**: Existing codebase pattern. The scaffold
engine embeds canonical copies; live copies are what
agents read at runtime. Both must match.
