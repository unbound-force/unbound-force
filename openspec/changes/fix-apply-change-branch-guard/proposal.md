## Why

The `openspec-apply-change/SKILL.md` skill file has a CRITICAL/MANDATORY
rule ("NEVER switch branches or suggest archiving with uncommitted
changes") placed at line 212 in the Guardrails section -- after the
full implementation workflow (lines 16-183). An agent executing the
workflow sequentially can reach and act on branch-switching suggestions
before encountering this constraint.

This is a T2 weakness (critical rule placed after the workflow it
governs) identified in the issue #346 root cause analysis. The same
pattern exists in sibling issues #355 (cobalt-crush.md) and #361
(speckit-workflow/SKILL.md).

Fixes #359.

## What Changes

Elevate the branch-switching guardrail from the Guardrails section
at the end of the file to a Pre-condition block immediately after
the "Steps" heading and before Step 1, so agents encounter it before
executing any workflow step.

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `openspec-apply-change skill`: Branch safety rule is promoted to
  a pre-condition, ensuring agents encounter it before any workflow
  step rather than after the workflow completes

### Removed Capabilities
- None

## Impact

- **File affected**: `.opencode/skills/openspec-apply-change/SKILL.md`
- **Agent behavior**: Agents will encounter the branch safety
  constraint before executing any workflow steps, preventing
  accidental branch switches with uncommitted changes
- **No functional change**: The rule content is identical; only its
  position in the file changes
- **No code changes**: This is a skill file (Markdown) edit only

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change affects a local skill file's internal structure. It does
not alter inter-hero artifact formats, communication protocols, or
self-describing outputs.

### II. Composability First

**Assessment**: N/A

This change modifies an OpenCode skill file within this repository.
It does not introduce dependencies or affect standalone hero
functionality.

### III. Observable Quality

**Assessment**: N/A

This change does not affect machine-parseable output, provenance
metadata, or quality metrics.

### IV. Testability

**Assessment**: N/A

This change is a positional edit to a Markdown skill file. No
testable code is added or modified.
