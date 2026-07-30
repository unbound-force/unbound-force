## Context

The `speckit.specify` command in `.opencode/commands/speckit.specify.md`
contains a dirty-tree guard at lines 42-50 that instructs the agent
to "STOP and ask the user for confirmation" before switching branches
when uncommitted changes exist. This instruction is prose-only -- no
`AskUserQuestion` tool call is specified. Under context compression
(long sessions, large contexts), agents can skip prose-only gates
and proceed to `git checkout -b`, silently carrying uncommitted work
to a new branch.

The sibling commands `opsx-propose.md` and `openspec-propose/SKILL.md`
had the same T1+T3 weakness pattern and have corresponding fix issues
(#353, #350). This design covers the fix for `speckit.specify.md`
(#358).

## Goals / Non-Goals

### Goals
- Replace prose-only dirty-tree guard with an explicit
  `AskUserQuestion` tool-call instruction
- Provide concrete options that match the guard's intent:
  stash and continue, or abort
- Maintain the existing exception rule (skip check when user
  explicitly requested a new spec in the same message) while
  still enforcing confirmation even in that case
- Match the pattern used in the related fixes for
  `opsx-propose.md` and `openspec-propose/SKILL.md`

### Non-Goals
- Changing the branch-naming logic or any other part of the
  `speckit.specify` workflow
- Adding automated tests for agent command files (these are
  instruction documents, not executable code)
- Modifying the branch check logic (step 3 in the command)
- Addressing other prose-only gates in other commands beyond
  what is scoped to issue #358

## Decisions

### D1: Use AskUserQuestion with two options

Replace the prose "STOP and ask the user" block with an explicit
instruction to use the `AskUserQuestion` tool with two options:

1. "Stash changes and continue"
2. "Abort -- keep changes as-is"

**Rationale**: These options match the existing guard's intent
(proceed or abort) while adding an actionable "stash" step that
protects the user's uncommitted work. The issue text suggests
exactly these options.

### D2: Preserve the exception clause with confirmation

The current text includes an exception: "only skip this check
if the user explicitly said to create a new spec in the same
message." The fix preserves this exception but clarifies that
even in this case, the AskUserQuestion tool MUST still be used
when uncommitted changes are detected. The exception means the
agent does not need to proactively check for uncommitted changes
when the user gave a direct command, but if changes are detected,
confirmation is still required.

**Rationale**: The issue specifically calls out that "this still
requires confirmation -- never silently switch branches with
uncommitted work" even in exception cases.

### D3: Match sibling command patterns

The replacement text establishes the pattern that the pending
sibling fixes for `opsx-propose.md` (#353) and
`openspec-propose/SKILL.md` (#350) will also adopt, maintaining
consistency across all three commands that share this guard.

**Rationale**: Consistency reduces cognitive load for agents and
maintainers. The same guard pattern should use the same enforcement
mechanism across all commands. Since the sibling fixes are not yet
merged, this change defines the canonical pattern.

## Risks / Trade-offs

### Low Risk: Instruction-only change

This modifies agent instruction text, not executable code. The
risk of regression is minimal -- the worst case is the agent
misinterprets the new instruction, which is mitigated by using
explicit tool-call syntax.

### Trade-off: Two options vs. three

The issue suggests two options ("Stash changes and continue" and
"Abort"). An alternative would be three options (adding "Continue
without stashing"). We chose two options to match the issue's
recommendation and to avoid the risk of uncommitted changes being
carried to the wrong branch, which is the exact problem this fix
addresses.
