## Context

The `openspec-propose` skill (SKILL.md lines 48-64) and the
companion `opsx-propose.md` command (lines 41-57) both contain
a dirty-tree guard that instructs the agent to "STOP and ask
the user for confirmation" before `git checkout -b` when
uncommitted changes are detected. However, neither file
specifies using the AskUserQuestion tool -- the guard is
prose-only.

Under context compression (long sessions, large codebases),
prose-only guards can be dropped from the agent's working
context. When this happens, the agent proceeds directly to
`git checkout -b`, potentially applying uncommitted work to
the wrong branch.

The proposal's constitution alignment confirmed this change
is N/A for Autonomous Collaboration, Composability First, and
Testability. It passes Observable Quality because the fix
replaces implicit reasoning with an explicit, observable tool
call.

## Goals / Non-Goals

### Goals
- Replace prose-only dirty-tree guard with explicit
  AskUserQuestion tool call in both files
- Specify exact option labels and abort behavior
- Ensure the guard survives context compression
- Keep both files (SKILL.md and opsx-propose.md) in sync

### Non-Goals
- Adding dirty-tree guards to other commands (e.g.,
  review-pr.md, speckit.specify.md) -- those are separate
  issues
- Adding session-resume guards (T3 weakness) -- that is a
  broader pattern requiring a separate design
- Changing the branch-check logic (step 3b) -- that logic
  is already explicit enough

## Decisions

### D1: Use AskUserQuestion with two fixed options

The guard will use the AskUserQuestion tool with exactly
two options:

1. "Stash changes and continue" -- the agent runs
   `git stash` then proceeds to `git checkout -b`
2. "Abort -- keep changes as-is" -- the agent stops
   and reports that the user should handle their
   uncommitted work before retrying

**Rationale**: These two options cover the practical
choices. A third option like "Continue without stashing"
would silently carry uncommitted changes to the new branch
which is the exact problem we are fixing.

### D2: Identical guard text in both files

The SKILL.md and opsx-propose.md files will receive
identical replacement text for the dirty-tree guard
section. This avoids drift between the two files.

**Rationale**: The issue notes that #353 was filed
separately for the command file, but the fix is identical.
Applying both in one change avoids redundant work and
ensures consistency.

### D3: Show uncommitted changes in the question context

Before presenting the AskUserQuestion, the agent MUST
display the output of `git status --short` so the user
can see exactly what uncommitted changes exist before
making a decision.

**Rationale**: Users cannot make an informed decision
without seeing what is at risk.

## Risks / Trade-offs

### Risk: AskUserQuestion tool availability

If the AskUserQuestion tool is unavailable or renamed in
a future OpenCode version, the guard will fail. This is
acceptable because a failing guard is safer than a
missing guard -- the agent will error rather than silently
skipping the check.

### Trade-off: Stash vs. no-stash option

We chose not to offer a "Continue without stashing"
option. This means users who intentionally want to carry
changes to the new branch must abort, manually checkout,
and re-run the command. This trades convenience for
safety -- the primary goal of this fix.
