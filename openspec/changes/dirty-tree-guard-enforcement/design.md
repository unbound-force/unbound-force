## Context

The `openspec-propose` skill and its companion command both contain a
dirty-tree guard that instructs the agent to check `git status --short`
before creating a branch. When uncommitted changes are detected, the
guard says "STOP and ask the user for confirmation" but does not
specify the AskUserQuestion tool call or its options. Under context
compression (T1 weakness), the agent can skip the guard entirely and
proceed to `git checkout -b` with a dirty working tree.

The proposal (constitution alignment: PASS on Principle V, N/A on
I-IV) establishes that this is a Security by Default hardening --
turning a prose-only gate into a tool-enforced gate.

## Goals / Non-Goals

### Goals
- Replace the prose-only dirty-tree confirmation with an explicit
  AskUserQuestion tool call specifying concrete options
- Ensure the guard survives context compression by making it a
  tool invocation instruction rather than reasoning prose
- Apply the fix to both affected files: `SKILL.md` and
  `opsx-propose.md`
- Add stash behavior when the user confirms (so changes are not
  lost)

### Non-Goals
- Changing the branch check logic (part b of the guard) -- that
  is already adequately specified
- Adding session-resume guards (T3 weakness) -- that is a broader
  concern tracked separately
- Modifying any other skills or commands -- only the two files
  identified in the issue are in scope
- Adding automated tests for skill file behavior -- skill files
  are agent instructions, not executable code

## Decisions

### D1: AskUserQuestion with two explicit options

The dirty-tree guard will specify an AskUserQuestion call with
exactly two options:

1. **"Stash changes and continue"** -- the agent runs
   `git stash push -m "openspec-propose: auto-stash before
   branch switch"` and proceeds to create the branch
2. **"Abort -- keep changes as-is"** -- the agent stops
   immediately and does not create the branch

**Rationale**: Two options are sufficient. "Stash and continue"
preserves uncommitted work safely. "Abort" gives the user full
control. A third option like "Continue without stashing" would
defeat the purpose of the guard by allowing changes to leak
across branches.

### D2: Show uncommitted changes before prompting

Before presenting the AskUserQuestion, the agent MUST display
the output of `git status --short` so the user can make an
informed decision. This is already implied by the existing prose
but will be made explicit.

### D3: Identical text in both files

Both `SKILL.md` and `opsx-propose.md` will receive the same
replacement text. The two files serve different contexts (skill
vs. command) but the guard logic is identical. Keeping them
in sync reduces drift risk.

### D4: Preserve existing exception clause

The existing exception clause ("if the user explicitly requested
a new change in the same message, this still requires
confirmation") is retained. The AskUserQuestion enforcement
applies regardless of how the change was initiated.

## Risks / Trade-offs

### R1: Stash may surprise users unfamiliar with git stash

Users who choose "Stash changes and continue" may not know how
to retrieve stashed changes later. **Mitigation**: The agent
will print a message after stashing: "Changes stashed. Run
`git stash pop` to restore them when ready."

### R2: Two files with identical guard text may drift

If one file is updated and the other is not, the guard behavior
diverges. **Mitigation**: Issue #353 tracks the companion command
separately, but this change addresses both files together. Future
scaffolding sync (via `uf init`) should detect drift.

### R3: AskUserQuestion tool may not be available in all contexts

The AskUserQuestion tool is a standard part of the agent
framework. If it were unavailable, the agent would fall back to
its default behavior (which is the current prose-only state).
**Acceptance**: This is acceptable -- the fix improves the common
case without breaking edge cases.

### R4: Other files with prose-only dirty-tree guards remain unfixed

At least four other files contain the same prose-only dirty-tree
guard pattern vulnerable to T1 context-compression bypass:

- `.opencode/skills/openspec-archive-change/SKILL.md`
- `.opencode/skills/speckit-workflow/SKILL.md`
- `.opencode/commands/cobalt-crush.md`
- `.opencode/commands/speckit.specify.md`

These are explicitly out of scope for this change (see
Non-Goals). The parent audit issue (#346) tracks the broader
hardening effort. After this change lands, the project will
have two files with tool-enforced guards and four with
prose-only guards. Future work should apply the same pattern
to the remaining files.

## Coverage Strategy

This change modifies agent instruction markdown files, not
executable Go code. No automated test coverage applies.
Verification is manual per task 2.1 -- the implementer
compares the replaced sections in both files to confirm
identical guard text, correct AskUserQuestion options, stash
command, recovery message, and abort behavior.
