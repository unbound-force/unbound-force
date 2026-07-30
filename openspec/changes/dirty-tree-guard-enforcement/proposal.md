## Why

The `openspec-propose` skill (`SKILL.md`) and its companion command
(`opsx-propose.md`) describe a dirty-tree guard before `git checkout -b`
entirely in prose. The guard says "STOP and ask the user for confirmation"
but never specifies an explicit `AskUserQuestion` tool call with concrete
options. Under context compression, the LLM can omit the guard reasoning
entirely and create the branch despite uncommitted changes, silently
applying work-in-progress files to the wrong branch.

This is a T1 (prose-only gate) + T3 (no session-resume guard) weakness
identified in the parent audit (issue #346). Issue #350 covers the skill
file; issue #353 covers the companion command file.

## What Changes

Replace the prose-only dirty-tree guard in both `SKILL.md` and
`opsx-propose.md` with an explicit `AskUserQuestion` tool call that
presents concrete options when uncommitted changes are detected.

The guard currently reads:
> "STOP and ask the user for confirmation before switching branches."

It will be replaced with a structured tool call specification:
> Use the **AskUserQuestion tool** with options:
> `["Stash changes and continue", "Abort -- keep changes as-is"]`
> Only proceed to `git checkout -b` if the user selects "Stash changes
> and continue".

## Capabilities

### New Capabilities
- None (this hardens an existing guard, no new features)

### Modified Capabilities
- `dirty-tree-guard`: Strengthened from prose-only instruction to
  explicit AskUserQuestion enforcement with concrete options. The guard
  now survives context compression because it is a tool call instruction,
  not reasoning prose.

### Removed Capabilities
- None

## Impact

- **Files affected**:
  - `.opencode/skills/openspec-propose/SKILL.md` (dirty-tree
    guard section under step 3, `a. **Dirty working tree check**`)
  - `.opencode/commands/opsx-propose.md` (dirty-tree guard
    section under step 3, `a. **Dirty working tree check**`)
- **Behavioral change**: When uncommitted changes are detected, the agent
  MUST call AskUserQuestion with two explicit options before proceeding.
  Previously, the agent was told to "ask" but the mechanism was
  unspecified.
- **Risk**: Low. This is a tightening of an existing guard. No functional
  behavior changes for clean working trees. Users with dirty trees will
  now reliably see a confirmation prompt.
- **Backward compatibility**: Fully compatible. The guard only activates
  when `git status --short` returns output.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change modifies agent instruction files (skill/command markdown),
not inter-hero artifact formats or communication protocols. No impact
on artifact-based collaboration.

### II. Composability First

**Assessment**: N/A

No dependencies are introduced or modified. The change is internal to
the openspec-propose workflow and does not affect hero composability.

### III. Observable Quality

**Assessment**: N/A

No output formats or provenance metadata are affected. The change
hardens an input validation gate within the agent workflow.

### IV. Testability

**Assessment**: N/A

The modified files are agent instruction markdown, not executable code.
Testability of the skill/command behavior is governed by the agent
framework, not by this change.

### V. Security by Default

**Assessment**: PASS

This change directly supports Security by Default. The dirty-tree guard
is an input validation gate that prevents the agent from silently
switching branches with uncommitted work. Replacing prose-only
instructions with explicit tool call enforcement ensures the gate
cannot be bypassed under context compression -- a form of structural
security hardening.
