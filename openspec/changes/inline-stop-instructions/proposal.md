## Why

AI agents repeatedly cross workflow phase boundaries
during spec-phase commands. The agent creates artifacts
correctly, then continues into implementation — editing
source code, syncing scaffold assets, running tests —
without the user reviewing the plan first. This has
happened 3 times in the current session (issue #109).

The root cause: guardrails are at the **end** of the
command file in a separate section. By the time the
agent reads them, it has completed artifact creation
and is in "what next?" mode. The momentum carries it
into implementation.

The fix: add **inline STOP instructions** at the
exact point where each command's work is done, plus
add "why" reasoning to existing guardrails.

## What Changes

### Modified Capabilities

- 7 speckit command files (specify, clarify, plan,
  tasks, analyze, checklist, testreview) +
  1 opsx-propose command + 1 openspec-propose skill:
  Add inline STOP after the command's completion
  point, add "why" to guardrails.

### New Capabilities

None.

### Removed Capabilities

None.

## Impact

- 9 Markdown files modified (~5 lines added to each)
- No Go code changes
- No scaffold asset syncs needed (all target files
  are externalized or OpenSpec-owned)
- No test changes

## Constitution Alignment

All N/A — behavioral instruction improvements to
Markdown command files.
