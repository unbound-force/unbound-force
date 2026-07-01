## Context

The `uf` binary embeds 10 slash command files in
`internal/scaffold/assets/opencode/commands/`. When
`uf init` runs, these files are deployed to the
target repo's `.opencode/commands/` directory. The
scaffold engine (`internal/scaffold/scaffold.go`)
manages file ownership via `isToolOwned()`, which
returns true for all paths under `opencode/commands/`.

Currently, the 10 uf-owned commands use unprefixed
names (e.g., `review-council.md`, `unleash.md`).
Other tools in the ecosystem already use clear
namespacing: `speckit.*`, `muti-mind.*`, `opsx-*`.
This change renames uf commands to use `uf.`
dot-notation for consistency.

## Goals / Non-Goals

### Goals
- Rename all 10 uf-embedded command files from
  unprefixed to `uf.` dot-notation
- Add a migration map to the scaffold engine to
  automatically remove orphaned old-name files
- Update all hard references in Go source, tests,
  and embedded assets
- Update all documentation and cross-command
  references
- Provide a CHANGELOG entry and migration
  documentation
- Maintain backward compatibility during transition
  (orphan cleanup, not hard breakage)

### Non-Goals
- Rename commands owned by other tools (Replicator,
  Gaze, OpenSpec, Speckit, Muti-Mind)
- Rename the hero identity `cobalt-crush` in
  orchestration Go code
- Modify historical CHANGELOG entries
- Add OpenCode-level grouping behavior (this is an
  OpenCode upstream concern, not uf's)

## Decisions

### D1: Dot-notation over dash-prefix

**Decision**: Use `uf.command-name` (dot-notation).

**Rationale**: Dot-notation is the dominant convention
in this project (22 of 48 commands). It provides
alphabetical clustering when commands are listed.
While OpenCode treats dots as regular characters
with no special grouping behavior, the visual
clustering from alphabetical sorting is sufficient.
Consistent with `speckit.*` and `muti-mind.*`.

**Alternative considered**: `uf-command-name`
(dash-prefix). Only `uf-init` uses this today.
Dash-prefix does not cluster as well and is the
minority convention (19 of 48).

### D2: Migration map in scaffold engine

**Decision**: Add a `renamedCommands` map
(`map[string]string`) to `scaffold.go` that maps old
embedded asset relative paths to new ones. During
the scaffold walk, after deploying new files, iterate
the migration map and remove any old-path files that
exist in the target directory.

**Rationale**: This approach is:
- Automatic: users just re-run `uf init`
- Safe: only removes files at known old paths that
  were previously tool-owned
- Self-documenting: the map is a clear record of
  what was renamed
- Idempotent: safe to run multiple times

**Location**: Place the map as a package-level
variable near `isToolOwned()` (around line 302).
Execute cleanup in `Run()` after the main scaffold
walk completes.

### D3: isDivisorAsset path update

**Decision**: Update the hardcoded path check in
`isDivisorAsset()` (line 338) from
`"opencode/commands/review-council.md"` to
`"opencode/commands/uf.review-council.md"`.

**Rationale**: This function controls which assets
are deployed in `--divisor` mode. The path must
match the new embedded asset filename.

### D4: Hero identity unchanged

**Decision**: The string `"cobalt-crush"` in
`internal/orchestration/heroes.go`,
`internal/orchestration/learning.go`, and related
test files is the hero identity, not the command
name. It stays unchanged.

**Rationale**: The command `/uf.cobalt-crush` invokes
the agent `cobalt-crush-dev`. The hero identity
(`Name: "cobalt-crush"`) is an internal constant
used for orchestration routing, learning feedback,
and schema samples. Renaming the hero identity would
be a separate, larger change with different blast
radius.

### D5: Cross-command reference updates

**Decision**: Update all command files (including
those owned by other tools like `/opsx-propose`) that
reference uf commands by old names. This ensures
functionality after the rename regardless of which
tool originally created the file.

**Rationale**: A user running `/opsx-propose` that
hands off to `/unleash` (old name) will fail if
`/unleash` no longer exists. Functional correctness
takes priority over strict ownership boundaries for
cross-references.

### D6: CHANGELOG handling

**Decision**: Add a new entry in CHANGELOG.md under
the Unreleased section documenting the rename with a
migration reference. Do not modify historical entries.

**Rationale**: Historical entries reflect the state
at time of release. A new entry is informative and
points users to the migration guide.

### D7: Migration documentation

**Decision**: Include a migration guide in the
CHANGELOG entry and the GitHub issue body. The guide
covers: update binary, re-run `uf init`, update
custom references, commit.

**Rationale**: Users of other repos that have run
`uf init` need clear, actionable steps. The issue
(#302) already contains this guide.

## Risks / Trade-offs

### R1: Broad blast radius

~500+ soft references across documentation, specs,
and OpenSpec change artifacts. Risk of missing
references is non-trivial.

**Mitigation**: Use systematic grep after
implementation to verify no old names remain in
active files. Historical/archived OpenSpec changes
are exempt.

### R2: Other tools' command files become stale

Commands owned by Replicator, OpenSpec, Gaze, etc.
may contain references to old uf command names. We
update these in this change, but future re-runs of
those tools' `init` commands may overwrite our fixes
with old names.

**Mitigation**: File issues against those tools to
update their embedded templates. This is a
coordination concern, not a blocking one.

### R3: User confusion during transition

Users who see both `/review-council` (orphan) and
`/uf.review-council` (new) may be confused if they
don't run `uf init` cleanly.

**Mitigation**: The migration map in `uf init`
handles orphan cleanup automatically. The CHANGELOG
entry and issue #302 document the transition.
