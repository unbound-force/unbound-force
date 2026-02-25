# Quickstart: Unbound Force Organization Constitution

**Spec**: [spec.md](spec.md) | **Date**: 2026-02-25

## What is the Org Constitution?

The Unbound Force organization constitution is the highest-authority governance document for all hero repositories. It defines three core principles that every hero must align with:

1. **Autonomous Collaboration** -- Heroes communicate through artifacts, not runtime coupling
2. **Composability First** -- Every hero works standalone; combining them adds value
3. **Observable Quality** -- All outputs are machine-parseable with provenance metadata

The constitution lives at `.specify/memory/constitution.md` and is versioned using semantic versioning.

## Reading the Constitution

The constitution is at:

```
.specify/memory/constitution.md
```

It contains:

- **Core Principles** -- Three numbered principles (I, II, III) with MUST/SHOULD/MAY rules
- **Hero Constitution Alignment** -- Rules for how hero constitutions relate to the org constitution
- **Development Workflow** -- Branching, review, CI, versioning, and commit conventions
- **Governance** -- Amendment process, versioning, supremacy clause, compliance review

## Checking Hero Constitution Alignment

### Prerequisites

- OpenCode installed and configured
- The alignment agent (`constitution-check.md`) installed in `.opencode/agents/`
- The `/constitution-check` command installed in `.opencode/command/`

### Running an Alignment Check

From any hero repository with the alignment agent installed:

```
/constitution-check
```

The agent reads:

1. The org constitution (from the unbound-force meta repo or a local copy)
2. The hero constitution (from `.specify/memory/constitution.md` in the current repo)

It produces a structured report showing:

- Whether each org principle is supported by at least one hero principle
- Whether any hero principle contradicts an org principle
- Whether the hero constitution includes a `parent_constitution` reference
- An overall ALIGNED or NON-ALIGNED verdict

### Example Output

```markdown
# Constitution Alignment Report

**Hero**: Gaze
**Hero Constitution Version**: 1.0.0
**Org Constitution Version**: 1.0.0
**Checked**: 2026-02-25T14:30:00Z
**Overall Status**: ALIGNED

## Findings

### ALIGNED: I. Autonomous Collaboration ↔ III. Actionable Output

**Org Principle**: Autonomous Collaboration
**Hero Principle**: Actionable Output
**Status**: ALIGNED
**Rationale**: Actionable Output requires machine-readable JSON
  alongside human-readable output, supporting the org requirement
  for self-describing artifacts with provenance metadata.

### ALIGNED: II. Composability First ↔ II. Minimal Assumptions

**Org Principle**: Composability First
**Hero Principle**: Minimal Assumptions
**Status**: ALIGNED
**Rationale**: Minimal Assumptions requires Gaze to work with
  existing code without annotation or restructuring, supporting
  the org requirement for standalone usability.

### ALIGNED: III. Observable Quality ↔ I. Accuracy

**Org Principle**: Observable Quality
**Hero Principle**: Accuracy
**Status**: ALIGNED
**Rationale**: Accuracy requires claims backed by automated
  regression tests, directly supporting the org requirement
  for quality claims backed by automated evidence.

## Summary

- Principles checked: 3
- Aligned: 3
- Gaps: 0
- Contradictions: 0
- Parent constitution reference: MISSING (pre-dates org constitution)
```

## Amending the Constitution

1. Create a feature branch
2. Edit `.specify/memory/constitution.md`
3. Update the version line per semantic versioning:
   - MAJOR: Removing or redefining a MUST rule
   - MINOR: Adding a principle or expanding guidance
   - PATCH: Clarifying wording without changing meaning
4. Update `Last Amended` date
5. Submit a PR with a migration plan (if MAJOR/MINOR)
6. After merge, run `/constitution-check` in each hero repo to verify alignment

## Constitution Check in the Speckit Pipeline

The Constitution Check is a mandatory gate during `/speckit.plan`. At the plan phase, each of the three org principles is evaluated against the planned work:

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. Autonomous Collaboration | PASS/FAIL | [how the plan satisfies or violates this principle] |
| II. Composability First | PASS/FAIL | [how the plan satisfies or violates this principle] |
| III. Observable Quality | PASS/FAIL | [how the plan satisfies or violates this principle] |

Any FAIL is CRITICAL severity and blocks proceeding to implementation.

## For New Hero Repositories

When bootstrapping a new hero repository:

1. Copy the constitution template from speckit
2. Run `/speckit.constitution` to fill in hero-specific principles
3. Include `parent_constitution` reference to the org constitution version
4. Run `/constitution-check` to verify alignment before proceeding
5. Ensure hero principles extend (never contradict) the org constitution
