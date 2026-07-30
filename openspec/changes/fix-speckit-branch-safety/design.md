## Context

The `speckit-workflow/SKILL.md` file contains a "Branch Safety"
section at lines 114-129 that includes CRITICAL rules about
committing work before branch switches and never silently switching
branches with a dirty working tree. These rules are placed after
the full workflow (lines 18-113), meaning an agent executing the
workflow can reach branch-affecting actions before encountering
the safety constraints.

This is a T2 weakness pattern (CRITICAL rule placed after the
workflow it governs), previously identified in `cobalt-crush.md`
(#355) and `openspec-apply-change/SKILL.md` (#359).

## Goals / Non-Goals

### Goals
- Move branch safety rules to appear before any workflow steps
- Preserve the exact content of the branch safety rules
- Follow the same fix pattern used for sibling T2 issues (#355,
  #359)

### Non-Goals
- Rewriting or expanding the branch safety rules
- Modifying other sections of the skill file
- Addressing branch safety in other skill files (separate issues)

## Decisions

**D1: Create a "Pre-conditions" section before "Reading tasks.md"**

The branch safety content will be extracted from its current
location (lines 114-129) and placed into a new "Pre-conditions"
section immediately before "Reading tasks.md" (currently line 30).
This mirrors the fix pattern described in issue #361 and matches
the approach taken for similar T2 fixes.

Rationale: Placing constraints as pre-conditions ensures agents
process them before any workflow step that could trigger a branch
switch. The "Pre-conditions" heading clearly signals that these
rules apply to the entire workflow.

**D2: Remove the original "Branch Safety" section**

After relocating the content, the original section at lines
114-129 is deleted entirely to avoid duplication and potential
drift between two copies of the same rules.

**D3: Preserve exact rule text**

The branch safety rules will be moved verbatim. No rewording
or expansion. This keeps the change minimal and focused on
structural placement only.

## Risks / Trade-offs

**Low risk**: This is a content relocation within a single file.
No behavioral logic changes, no code changes, no dependency
changes.

**Trade-off**: The "Pre-conditions" section adds a heading before
the workflow begins, slightly increasing the file's structure.
This is acceptable because it makes the constraint ordering
explicit.
