## Why

The `/opsx-propose` command creates OpenSpec artifacts
(proposal, design, specs, tasks) but has no guardrail
preventing the agent from immediately jumping into
implementation and `/finale` afterward. This happened
in practice during the `sandbox-vertex-envvars` change
— the agent created the artifacts, then implemented
the code change and merged the PR, all in response to
a single `/opsx-propose` invocation.

The Speckit pipeline commands received guardrails in
the `opsx/workflow-phase-boundaries` change, but the
OpenSpec commands (`/opsx-propose`, `/opsx-explore`,
`/opsx-archive`) were not covered.

## What Changes

### Modified Capabilities

- `/opsx-propose` (`.opencode/command/opsx-propose.md`
  or equivalent OpenSpec-managed file): Add a Guardrails
  section stating the command creates artifacts only —
  never implements, commits, pushes, or creates PRs.

### New Capabilities

None.

### Removed Capabilities

None.

## Impact

- 1 command file modified (via `/uf-init` customization,
  since OpenSpec commands are created by `openspec init`)
- No Go code changes
- No test changes

## Constitution Alignment

All N/A — this is a behavioral instruction addition to
a Markdown command file.
