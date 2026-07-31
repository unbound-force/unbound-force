## Context

`.opencode/commands/speckit.implement.md` is the slash
command that advances a Speckit spec through its
implementation phase. It manages checkbox state in
`tasks.md`, tracks progress, and delegates actual
coding to `/unleash` or `/cobalt-crush`. It never
modifies source code directly.

Lines 156-158 contain a guardrail that was copied from
a spec-authoring command without being updated. The
guardrail correctly forbids source code changes but
then lists `/speckit.implement` (this command) as a
valid destination for implementation work, creating a
self-referential contradiction.

The fix is a one-line text edit: remove the
`/speckit.implement,` token from line 158, leaving
`/unleash` and `/cobalt-crush` as the only valid
implementation destinations.

## Goals / Non-Goals

### Goals
- Remove the self-reference from the guardrail on
  line 158 of `.opencode/commands/speckit.implement.md`
- Preserve the intent of the guardrail (this command
  does not modify source code)
- Leave the corrected guardrail pointing to `/unleash`
  and `/cobalt-crush` as the proper implementation
  commands

### Non-Goals
- Do not change the guardrail's intent or any other
  part of the file
- Do not audit or fix other commands for similar issues
- Do not add tests (the file is prompt content, not
  executable code)

## Decisions

**Single-file, single-line edit**: The fix is confined
to one line in one file. No other files reference this
guardrail text. The scaffold copy of this command (if
one exists in `internal/scaffold/assets/`) must also
be checked and updated if present.

**No spec delta**: The change does not alter any
functional requirement, API, or data model. A full
delta spec is not warranted. The proposal and this
design document are sufficient specification.

## Risks / Trade-offs

**Risk**: Scaffold copy out of sync. If
`internal/scaffold/assets/opencode/commands/speckit.implement.md`
exists, it must receive the same fix or newly scaffolded
repos will receive the broken guardrail. The task list
includes an explicit check for this.

**Trade-off**: None. The fix is unambiguous and has no
behavioral side effects.
