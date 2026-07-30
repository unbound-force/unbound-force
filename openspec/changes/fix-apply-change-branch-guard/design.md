## Context

The `openspec-apply-change/SKILL.md` file has a CRITICAL guardrail
rule at line 212 (Guardrails section) that prohibits switching
branches or suggesting archiving with uncommitted changes. The full
implementation workflow spans lines 16-183, meaning an agent
executing sequentially reaches the workflow steps before encountering
this safety constraint.

This is a T2 weakness pattern (critical rule placed after the
workflow it governs) identified in the issue #346 root cause
analysis.

## Goals / Non-Goals

### Goals
- Ensure agents encounter the branch safety constraint before
  executing any workflow step
- Preserve the existing rule text exactly -- no semantic changes
- Follow the pattern established by the issue's prescribed fix

### Non-Goals
- Rewriting or expanding the guardrail rule's content
- Refactoring other sections of the skill file
- Fixing the same T2 pattern in other skill files (those are
  tracked by separate issues #355 and #361)
- Adding new guardrails or pre-conditions

## Decisions

**D1: Add pre-condition block, keep original guardrail**

Add a Pre-condition block immediately after the "Steps" heading
and before Step 1 that states the branch safety constraint. Keep
the original guardrail bullet in the Guardrails section as-is for
redundancy -- the Guardrails section serves as a reference
checklist, while the pre-condition ensures early encounter.

Rationale: The issue prescribes adding a pre-condition block. The
Guardrails section may be consulted independently as a checklist,
so removing the original would reduce its completeness. Duplication
is acceptable for safety-critical rules.

**D2: Use the exact wording from the issue's prescribed fix**

The issue provides specific pre-condition text. Use it verbatim
to maintain traceability between the fix and its root cause
analysis.

## Risks / Trade-offs

- **Minor duplication**: The rule appears in both the pre-condition
  block and the Guardrails section. This is intentional -- the
  pre-condition ensures early encounter while the Guardrails section
  maintains its role as a complete checklist.
- **Low risk overall**: This is a positional edit to a Markdown file
  with no functional code changes.
