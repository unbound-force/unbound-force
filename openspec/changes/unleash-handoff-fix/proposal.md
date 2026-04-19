## Why

When spec-phase commands (`/opsx-propose`, `/speckit.plan`,
etc.) finish creating artifacts, they prompt the user with
a "next step" handoff message. Currently, these handoffs
either omit `/unleash` entirely or list it last after
`/opsx-apply` and `/cobalt-crush`. This causes agents to
recommend single-agent sequential implementation even when
the change has many independent task groups that would
benefit from parallel swarm execution.

The `/unleash` command is the primary autonomous pipeline
and should be the first recommendation for multi-task
changes. A user reported this exact issue: after running
`/opsx-propose` on a 22-task change, the agent suggested
only `/opsx-apply` instead of `/unleash`.

## What Changes

Update all implementation handoff prompts, STOP blocks,
and guardrails across 13 files to consistently recommend
`/unleash` as the primary implementation option.

## Capabilities

### New Capabilities

- None — no new functionality is added.

### Modified Capabilities

- `Handoff prompts`: All spec-phase command STOP blocks
  and output prompts now list `/unleash` first, followed
  by `/cobalt-crush` and `/opsx-apply`.
- `Guardrails`: All "NEVER run" lists in spec-phase
  commands now include `/unleash` alongside `/opsx-apply`
  and `/cobalt-crush`.
- `Fallback routing`: The `/cobalt-crush` command's
  no-context fallback prompt now includes `/unleash` as
  the first option.

### Removed Capabilities

- None.

## Impact

- 13 Markdown files modified (commands, skills)
- No Go source code changes
- No test changes
- No convention pack changes
- No agent persona changes (only command/skill files)

Files affected:
- `.opencode/command/opsx-propose.md`
- `.opencode/command/cobalt-crush.md`
- `.opencode/command/opsx-explore.md`
- `.opencode/command/uf-init.md`
- `.opencode/command/speckit.specify.md`
- `.opencode/command/speckit.clarify.md`
- `.opencode/command/speckit.plan.md`
- `.opencode/command/speckit.tasks.md`
- `.opencode/command/speckit.analyze.md`
- `.opencode/command/speckit.checklist.md`
- `.opencode/command/speckit.testreview.md`
- `.opencode/skills/openspec-propose/SKILL.md`
- `.opencode/skill/speckit-workflow/SKILL.md`

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change modifies prompt text only. It does not
affect artifact-based communication between heroes or
inter-hero data formats.

### II. Composability First

**Assessment**: PASS

The change adds `/unleash` as an option alongside
existing commands. No mandatory dependencies are
introduced — users can still choose `/opsx-apply` or
`/cobalt-crush` for sequential execution.

### III. Observable Quality

**Assessment**: N/A

No machine-parseable output or provenance metadata is
affected. This is a prompt text change only.

### IV. Testability

**Assessment**: N/A

No testable components are modified. The change affects
natural language prompts, not executable code. The
existing `TestScaffoldOutput_*` regression tests in
`scaffold_test.go` will catch any scaffold asset drift
if the command files are also embedded assets.
