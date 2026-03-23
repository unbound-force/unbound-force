## Context

The `/cobalt-crush` command file exists at
`.opencode/command/cobalt-crush.md` but is not included
in the scaffold asset directory at
`internal/scaffold/assets/opencode/command/`. The
scaffold engine deploys all files under the assets
directory via `embed.FS` -- adding a file there is
sufficient to include it in `uf init` output.

The file is tool-owned (has a version marker comment),
so the scaffold engine will overwrite it on re-scaffold.

## Goals / Non-Goals

### Goals
- Add `cobalt-crush.md` to the scaffold assets so
  `uf init` deploys it
- Update the scaffold file count test assertion
- Keep the live and scaffold copies in sync

### Non-Goals
- Changing the command file's content or behavior
- Adding other missing command files (muti-mind,
  workflow, gaze, opsx commands are out of scope)
- Modifying the scaffold engine logic

## Decisions

**Copy the live file as-is**: The scaffold asset is an
exact copy of `.opencode/command/cobalt-crush.md`. No
modifications needed -- the file already follows the
scaffold conventions (version marker in frontmatter,
`$ARGUMENTS` placeholder for user input).

**Tool-owned classification**: The file starts with a
YAML frontmatter block, which the scaffold engine's
`isToolOwned()` function uses to classify files. Since
it has frontmatter, it will be treated as tool-owned
and overwritten on re-scaffold. This is correct -- the
command file should be updated when the CLI binary is
updated.

**File count increase**: The scaffold produces N files
currently. Adding one file changes the count to N+1.
The test in `cmd/unbound-force/main_test.go` asserts
the exact count and must be updated.

## Risks / Trade-offs

**Low risk**: This is a copy operation with a test
update. No logic changes, no new dependencies, no
behavioral changes. The scaffold engine already handles
the deployment; we're just adding one more file to its
inventory.

**Trade-off**: The live copy and scaffold copy must
stay in sync. This is the same trade-off that exists
for all other scaffolded command files (speckit.*,
review-council, constitution-check). Future changes
to the command file must update both copies.
