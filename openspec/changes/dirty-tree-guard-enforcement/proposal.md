## Why

The `openspec-propose` skill (SKILL.md) and companion command
(`opsx-propose.md`) both describe a dirty-tree guard before
`git checkout -b` entirely in prose. The guard says "STOP and
ask the user for confirmation" but never specifies using the
AskUserQuestion tool. Under context compression, the guard
reasoning can be omitted and the branch created despite
uncommitted changes -- silently applying work to the wrong
branch.

This is a T1 + T3 weakness (gate exists in prose but is not
enforced by a tool call; no session-resume guard). Fixing
this hardens the skill and command against context compression
failures.

Fixes #350.

## What Changes

Replace the prose-only dirty-tree guard in both files with an
explicit AskUserQuestion tool call. When `git status --short`
detects uncommitted changes, the agent MUST use the
AskUserQuestion tool with specific options before proceeding.

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `dirty-tree-guard`: Enforced via explicit AskUserQuestion
  tool call instead of prose-only instruction. The guard
  now specifies exact tool usage, option labels, and abort
  behavior that survives context compression.

### Removed Capabilities
- None

## Impact

- `.opencode/skills/openspec-propose/SKILL.md` lines 48-64
- `.opencode/commands/opsx-propose.md` lines 41-57
- No Go source code changes
- No CI impact
- No schema changes

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change modifies agent skill/command instructions, not
inter-hero artifact communication. No artifact formats or
hero interfaces are affected.

### II. Composability First

**Assessment**: N/A

This change is internal to the meta-repository's agent
instructions. No hero dependencies are introduced or
modified.

### III. Observable Quality

**Assessment**: PASS

The fix makes the guard's enforcement mechanism explicit
and machine-verifiable. An AskUserQuestion tool call is
a concrete, observable action -- unlike prose that can
be compressed away. This improves the observability of
the safety gate.

### IV. Testability

**Assessment**: N/A

This change modifies Markdown instruction files, not
testable source code. The fix is verifiable by inspection
of the instruction text.
