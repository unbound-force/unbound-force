## Why

The `speckit.specify` command (`.opencode/commands/speckit.specify.md`,
lines 42-50) describes a check for uncommitted changes before
`git checkout -b` entirely in prose ("STOP and ask the user for
confirmation"). No `AskUserQuestion` tool call is specified. Under
context compression, an agent can skip the check and create the
branch despite uncommitted changes, risking applying in-progress
work to the wrong branch.

This is a T1+T3 weakness pattern (gate exists in prose but is not
enforced by a tool call). The same pattern was identified in sibling
commands `opsx-propose.md` (issue #353) and
`openspec-propose/SKILL.md` (issue #350), which have corresponding
open issues but are not yet merged. This change establishes the fix
pattern for `speckit.specify.md` (#358); the sibling fixes will
adopt the same pattern.

Fixes: https://github.com/unbound-force/unbound-force/issues/358

## What Changes

Replace the prose-only dirty-tree guard in `speckit.specify.md` with
an explicit `AskUserQuestion` tool call that enforces user
confirmation before proceeding with branch creation when uncommitted
changes are detected.

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `speckit.specify dirty-tree guard`: Replaces prose-only "STOP and
  ask" instruction with an explicit `AskUserQuestion` tool call
  providing concrete options ("Stash changes and continue" /
  "Abort -- keep changes as-is"), making the gate enforceable
  under context compression.

### Removed Capabilities
- None

## Impact

- **File**: `.opencode/commands/speckit.specify.md` (lines 42-50)
- **Behavior**: Agents executing `/speckit.specify` will now see an
  explicit tool-call instruction for the dirty-tree guard, making
  it resilient to context compression and ensuring uncommitted work
  is never silently carried to a new branch.
- **Risk**: Minimal. This is a documentation/instruction fix to an
  agent command file, not a code change. The guard logic is
  unchanged; only the enforcement mechanism is strengthened.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change modifies an agent command instruction file, not
inter-hero artifact communication or output formats.

### II. Composability First

**Assessment**: N/A

This change is internal to the meta-repository's agent commands
and does not affect hero installability or dependencies.

### III. Observable Quality

**Assessment**: N/A

No machine-parseable output or provenance metadata is affected.

### IV. Testability

**Assessment**: N/A

This change modifies agent prose instructions, not testable code.
The speckit.specify command itself has no automated test harness;
the fix is verified by manual review of the instruction text.

### V. Security by Default

**Assessment**: PASS

This change strengthens a security-relevant gate. The dirty-tree
check prevents uncommitted work from being silently applied to the
wrong branch -- a data integrity concern. Replacing prose with an
explicit tool-call enforcement aligns with the principle of
enforcing security by design rather than relying on review-time
discovery.
