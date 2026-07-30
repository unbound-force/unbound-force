## Why

The `speckit-workflow/SKILL.md` file contains CRITICAL branch safety
rules (lines 114-129) that are placed **after** the full workflow
instructions (lines 18-113). An agent executing the workflow
sequentially can reach branch-related actions (Phase Checkpoints,
Worker Instructions) before encountering the branch safety
constraints. This is a T2 weakness: a CRITICAL/MANDATORY rule placed
after the workflow it governs.

This is part of a broader pattern identified in issue #346's root
cause analysis, where the same T2 weakness was found in
`cobalt-crush.md` (#355) and `openspec-apply-change/SKILL.md` (#359).

Fixes #361.

## What Changes

Move the "Branch Safety" section (currently at the end of the file,
lines 114-129) to a new "Pre-conditions" section positioned before
"Reading tasks.md" (line 30). This ensures agents encounter branch
safety constraints before executing any workflow steps.

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `speckit-workflow skill`: Branch safety rules appear as
  pre-conditions before the workflow, ensuring agents process them
  before reaching any branch-affecting actions.

### Removed Capabilities
- None

## Impact

- **File**: `.opencode/skills/speckit-workflow/SKILL.md`
- **Behavioral**: Agents loading this skill will encounter branch
  safety rules earlier in the document, reducing risk of branch
  switches with dirty working trees.
- **No breaking changes**: The content is relocated, not modified.
  All existing constraints remain in effect.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change restructures a skill file's section ordering.
It does not affect artifact-based communication or
inter-hero collaboration patterns.

### II. Composability First

**Assessment**: N/A

This change modifies an internal skill file. It does not
affect hero installation, standalone functionality, or
dependency chains.

### III. Observable Quality

**Assessment**: N/A

This change does not affect machine-parseable output or
provenance metadata. It is a documentation/instruction
restructuring.

### IV. Testability

**Assessment**: N/A

This change modifies a skill instruction file, not
executable code. No testable components are affected.

### V. Security by Default

**Assessment**: PASS

Moving branch safety rules to a pre-conditions section
improves the security posture of the workflow by ensuring
agents process safety constraints before reaching any
branch-switching actions. This reduces the risk of data
loss from uncommitted work being carried to wrong branches.
