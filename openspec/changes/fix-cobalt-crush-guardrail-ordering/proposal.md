## Why

The "Branch Safety Guardrails" section in
`internal/scaffold/assets/opencode/commands/cobalt-crush.md`
(lines 97-111) is placed at the END of the file, after all
workflow instructions (lines 40-96). An agent following
instructions sequentially will reach and execute branch
operations (lines 60-95) without ever seeing the guardrails
that govern them.

This is a T2 weakness: CRITICAL/MANDATORY rules placed after
the actions they govern. The fix is straightforward -- move
the guardrails before the instructions section so agents
encounter them first.

Fixes #355. Related: #359, #361 (same T2 pattern in other
files, tracked separately).

## What Changes

Restructure `cobalt-crush.md` so the branch safety
pre-conditions appear before the "Instructions" section,
ensuring agents read guardrails before encountering branch
operations.

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `cobalt-crush command`: Branch safety guardrails moved to
  a "Pre-conditions" block at the top of the Instructions
  section, ensuring agents encounter them before any branch
  operations.

### Removed Capabilities
- None

## Impact

- **File**: `internal/scaffold/assets/opencode/commands/cobalt-crush.md`
- **Scaffold consumers**: Any project scaffolded by `uf init`
  will receive the updated command file with guardrails in
  the correct position.
- **Existing deployments**: Projects already scaffolded will
  retain the old ordering until they re-run `uf init` or
  manually update.
- **Behavioral change**: None -- same guardrail content, just
  reordered for correct sequential processing by agents.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change reorders content within a single scaffold asset.
It does not affect inter-hero artifact communication or
exchange formats.

### II. Composability First

**Assessment**: N/A

This change does not introduce or remove dependencies between
heroes. The cobalt-crush command remains independently usable.

### III. Observable Quality

**Assessment**: PASS

Moving guardrails before the instructions they govern improves
the reliability of agent behavior, which directly supports
observable quality. Agents will consistently apply branch
safety checks rather than potentially skipping them due to
context window truncation or sequential processing.

### IV. Testability

**Assessment**: N/A

This is a content reordering change in a Markdown template.
No testable code is affected.
