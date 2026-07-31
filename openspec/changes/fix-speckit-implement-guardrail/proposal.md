## Why

`.opencode/commands/speckit.implement.md` contains a
guardrail copied from spec-authoring commands that was
never corrected for its new context. Lines 156-158 read:

```
- **NEVER modify source code** — this command updates
  spec artifacts ONLY. Implementation changes belong in
  `/speckit.implement`, `/unleash`, or `/cobalt-crush`.
```

`/speckit.implement` is the command that contains this
guardrail — listing itself as the implementation
destination contradicts the guardrail's own prohibition.
An agent reading this guardrail receives contradictory
instructions: "never implement here" and "implement
here". This is a copy-paste error from spec-authoring
commands (e.g., `/speckit.plan`) where the guardrail
is appropriate and `/speckit.implement` is legitimately
the correct delegation target.

Fixes unbound-force/unbound-force#238.

## What Changes

Remove `/speckit.implement` from the guardrail's list
of implementation destinations in
`.opencode/commands/speckit.implement.md` line 158.

## Capabilities

### Modified Capabilities
- `/speckit.implement` guardrail: no longer lists
  itself as a valid implementation destination

### New Capabilities
- None

### Removed Capabilities
- None

## Impact

### Files Affected

| Area | Changes |
|------|---------|
| `.opencode/commands/speckit.implement.md` | Remove self-reference on line 158 |

### External Dependencies
- None

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A — this change fixes documentation
in a slash command file. It does not affect
artifact-based communication between heroes.

### II. Composability First

**Assessment**: N/A — this change does not introduce
or remove any dependencies between components.

### III. Observable Quality

**Assessment**: PASS — correcting a self-contradictory
guardrail improves the accuracy of the system's
self-description. Agents receive unambiguous
instructions, which is a form of observable quality
in the prompt layer.

### IV. Testability

**Assessment**: N/A — slash command markdown files are
not directly unit-tested. The correctness of the
change is verified by reading the file after the edit.
