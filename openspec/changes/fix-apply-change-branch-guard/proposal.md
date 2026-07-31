## Why

The `openspec-apply-change/SKILL.md` contains a CRITICAL/MANDATORY
rule at line 212 -- "NEVER switch branches or suggest archiving with
uncommitted changes" -- buried in the Guardrails section at the end
of a 219-line file. An agent executing the workflow sequentially
reaches and acts on branch-switching suggestions (e.g., Step 8's
completion flow at lines 139-149) before encountering this
constraint.

This is a T2 weakness: a CRITICAL rule placed after the workflow
it governs. The fix is to elevate it to a pre-condition block
before Step 1, ensuring agents encounter the constraint before
any workflow step that might violate it.

Fixes: https://github.com/unbound-force/unbound-force/issues/359

Related sibling issues with the same T2 pattern:
- #355 (cobalt-crush.md)
- #361 (speckit-workflow/SKILL.md)

## What Changes

Elevate the branch-safety guardrail from the end-of-file Guardrails
section to a **Pre-condition** block placed immediately after the
"Steps" heading and before Step 1. The rule remains in Guardrails
as well (for completeness), but the pre-condition ensures agents
process it before any workflow step.

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `openspec-apply-change skill`: Branch-safety rule promoted to
  pre-condition position, ensuring agents encounter it before
  executing any workflow step

### Removed Capabilities
- None

## Impact

- **File**: `.opencode/skills/openspec-apply-change/SKILL.md`
- **Behavioral**: Agents will encounter the "NEVER switch branches
  with uncommitted changes" rule before Step 1, reducing the risk
  of branch-switching violations during implementation workflows
- **No runtime code changes**: This is a skill file (agent
  instruction) modification only

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change modifies agent instructions, not artifact formats
or inter-hero communication. No impact on autonomous collaboration.

### II. Composability First

**Assessment**: N/A

No dependencies are introduced or modified. The skill file remains
standalone and independently usable.

### III. Observable Quality

**Assessment**: N/A

No machine-parseable output or provenance metadata is affected.
This is an agent instruction ordering fix.

### IV. Testability

**Assessment**: N/A

No testable components are added or modified. The change is to
a Markdown skill file that guides agent behavior.

### V. Security by Default

**Assessment**: N/A

No security-sensitive operations, supply chain changes, input
validation, or dependency modifications are introduced. This is
an agent instruction ordering fix.
