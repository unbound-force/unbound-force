## Why

During PR #320 review, @trevor-vaughan observed that when
`/finale` detects a merge conflict (step 6b), the current
options are all manual: merge, rebase, stop, or skip. Since
the user is already running an AI agent, they will likely
want the agent to attempt conflict resolution rather than
switching to a terminal to do it themselves.

Adding a sub-agent option for automated conflict resolution
reduces context-switching and keeps the user in the AI-assisted
workflow. The agent can read conflicting files, understand both
sides of the conflict using the PR context, and attempt a
resolution -- falling back to manual if it fails.

Ref: https://github.com/unbound-force/unbound-force/issues/325

## What Changes

Add a new option to the conflict recovery flow (step 6b) in
`/finale` that spawns a `cobalt-crush-dev` sub-agent to
attempt automated merge conflict resolution.

The sub-agent:

1. Reads the conflicting files and both sides of the diff
2. Understands the intent of both the PR branch and the
   target branch changes
3. Resolves the conflict markers in each file
4. Stages the resolved files
5. Reports success or failure back to `/finale`

If the sub-agent fails to resolve any conflict, `/finale`
falls back to the existing manual resolution options.

## Capabilities

### New Capabilities
- `automated-conflict-resolution`: Spawns a sub-agent to
  attempt merge conflict resolution when `/finale` detects
  a `CONFLICTING` PR state

### Modified Capabilities
- `conflict-recovery` (step 6b): Adds a fifth option
  "Spawn sub-agent to resolve conflicts" alongside the
  existing merge, rebase, stop, and continue options

### Removed Capabilities
- None

## Impact

- **Files affected**:
  `.opencode/commands/finale.md` -- step 6b modification
- **Behavioral change**: Users see a new option (5) in
  the conflict recovery menu. Existing options 1-4 are
  unchanged.
- **Dependencies**: Requires `cobalt-crush-dev` sub-agent
  type to be available in the Task tool. This is already
  available in the current OpenCode configuration.
- **Risk**: Low. The sub-agent option is additive and
  opt-in. If the sub-agent fails, the flow falls back to
  manual resolution. No existing behavior is modified.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The sub-agent operates through well-defined inputs (list of
conflicting files, diff context) and outputs (resolved files
or failure report). It does not require synchronous coupling
with other heroes. The conflict resolution result is an
artifact (resolved source files) that `/finale` consumes
asynchronously.

### II. Composability First

**Assessment**: N/A

This change modifies the `/finale` slash command, which is
an agent instruction file, not a hero component. The
`cobalt-crush-dev` sub-agent is already independently
available. No new mandatory dependencies are introduced.

### III. Observable Quality

**Assessment**: PASS

The sub-agent reports which files it resolved and which it
could not. The user sees the resolution diff before
committing. All conflict resolution attempts are observable
in the conversation log.

### IV. Testability

**Assessment**: N/A

This change is to a Markdown command file (agent
instructions), not to Go source code. There are no
runtime components to test in isolation. The behavior is
verified through manual execution of the `/finale` flow.

### V. Security by Default

**Assessment**: N/A

This change does not introduce new external inputs,
dependencies, or privilege escalation. The sub-agent
operates on local files with conflict markers and
stages resolved files through standard git commands.
No new attack surface is created.
