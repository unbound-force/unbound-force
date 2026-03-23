## Why

The `/cobalt-crush` command exists as a live file in
`.opencode/command/cobalt-crush.md` but is not included
in the scaffold engine. This means `uf init` does not
deploy it to new projects. Developers scaffolding a new
project get all speckit commands, the review council, and
the constitution check -- but not the Cobalt-Crush
developer persona command. They have to discover and
create it manually.

## What Changes

Add `cobalt-crush.md` to the scaffold assets at
`internal/scaffold/assets/opencode/command/` so that
`uf init` deploys it alongside the other command files.

## Capabilities

### New Capabilities
- `scaffold cobalt-crush command`: `uf init` deploys
  the `/cobalt-crush` command file to new projects,
  giving developers immediate access to the Cobalt-Crush
  developer persona.

### Modified Capabilities
- `scaffold file count`: The total scaffold file count
  increases by 1. The test assertion in
  `cmd/unbound-force/main_test.go` must be updated.

### Removed Capabilities
- None

## Impact

- `internal/scaffold/assets/opencode/command/cobalt-crush.md`
  -- new scaffold asset (copy of the live command file)
- `cmd/unbound-force/main_test.go` -- file count
  assertion update
- No Go logic changes -- the scaffold engine already
  deploys all files under `internal/scaffold/assets/`

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

No artifact communication changes. This adds a command
file to the scaffold -- it's a deployment concern, not a
collaboration concern.

### II. Composability First

**Assessment**: PASS

The command file works independently. It detects the
active workflow (speckit or openspec) and delegates
accordingly. No new dependencies introduced.

### III. Observable Quality

**Assessment**: N/A

No output format changes. The command file's behavior
is unchanged -- only its deployment method changes.

### IV. Testability

**Assessment**: PASS

The scaffold file count test will be updated to account
for the new file. The existing
`TestScaffoldOutput_NoBareUnboundReferences` and
`TestScaffoldOutput_NoGraphthulhuReferences` regression
tests will automatically include the new file in their
sweeps.
