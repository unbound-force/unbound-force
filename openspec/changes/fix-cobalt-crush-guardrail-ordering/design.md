## Context

The `/cobalt-crush` command file
(`internal/scaffold/assets/opencode/commands/cobalt-crush.md`)
currently has its "Branch Safety Guardrails" section (lines
97-111) placed after the "Instructions" section (lines 40-96).
The Instructions section contains branch operations at lines
60-95 (checking for Speckit feature branches, OpenSpec changes,
and validating branch names). An agent processing this file
sequentially encounters branch operations before learning
the CRITICAL safety pre-conditions that govern them.

This is classified as a T2 weakness in the parent audit
(issue #346): CRITICAL/MANDATORY rules placed after the
actions they govern.

The proposal confirms constitution alignment: this is a
content reordering change that improves agent reliability
(Observable Quality) without affecting inter-hero
communication or composability.

## Goals / Non-Goals

### Goals
- Move the branch safety guardrails content to appear
  BEFORE the Instructions section
- Structure as a "Pre-conditions" subsection at the top
  of the Instructions section, following the pattern
  suggested in issue #355
- Preserve the exact guardrail content -- no behavioral
  changes, only positional

### Non-Goals
- Rewriting or expanding the guardrail rules themselves
- Addressing the same T2 pattern in other files (issue
  #359 for openspec-apply-change, #361 for
  speckit-workflow -- tracked separately)
- Modifying the cobalt-crush-dev agent file (only the
  command file is affected)
- Updating already-scaffolded projects (they receive
  changes on next `uf init`)

## Decisions

**D1: Insert as Pre-conditions subsection within Instructions**

The guardrails will be added as a `### Pre-conditions`
subsection immediately after the `## Instructions` heading
and before `### When arguments are provided`. This keeps
the guardrails logically grouped with the instructions
they govern while ensuring they appear first.

Alternatives considered:
- Separate top-level section before Instructions: rejected
  because the guardrails are specific to the instruction
  logic, not general metadata.
- Inline within each branch operation: rejected because
  it duplicates content and increases maintenance burden.

**D2: Remove the original section entirely**

The `## Branch Safety Guardrails` section at the end of the
file will be removed entirely (not left as a cross-reference)
to avoid duplication and potential drift between two copies.

**D3: Preserve exact wording**

The guardrail text will be moved verbatim. No rewording,
additions, or removals. This ensures the fix is purely
structural with zero behavioral risk.

## Risks / Trade-offs

**Low risk**: This is a pure content reordering. The same
guardrail text exists in the file, just at a different
position. No logic changes, no new content.

**Scaffold drift**: Projects that have already been
scaffolded will retain the old ordering. This is expected
behavior -- scaffold updates are applied on `uf init`
re-runs. The `isToolOwned` marker in scaffold files
governs whether updates are applied.
