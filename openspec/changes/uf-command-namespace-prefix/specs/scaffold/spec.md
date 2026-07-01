## ADDED Requirements

### Requirement: Migration Map for Renamed Commands

The scaffold engine MUST maintain a migration map
(`renamedCommands`) that maps old embedded command
file paths to their new `uf.`-prefixed paths. The
map MUST cover all 10 renamed commands.

#### Scenario: Clean orphan removal on re-run

- **GIVEN** a repository that previously ran
  `uf init` and has old-name command files (e.g.,
  `.opencode/commands/review-council.md`)
- **WHEN** the user runs `uf init` with the updated
  binary
- **THEN** the scaffold engine MUST deploy new
  `uf.*` command files AND remove any old-name files
  listed in the migration map

#### Scenario: Idempotent migration

- **GIVEN** a repository that has already been
  migrated (only `uf.*` files exist)
- **WHEN** the user runs `uf init` again
- **THEN** no errors MUST occur and no old-name
  files MUST be created or referenced

#### Scenario: Fresh repository

- **GIVEN** a repository that has never run
  `uf init`
- **WHEN** the user runs `uf init` for the first
  time
- **THEN** only `uf.*` command files MUST be created;
  no old-name files MUST be created

### Requirement: Orphan Cleanup Reporting

The scaffold engine SHOULD report removed orphan
files in the scaffold summary output, using a
distinct category (e.g., "migrated" or "removed").

#### Scenario: Summary includes migrated files

- **GIVEN** a repository with 3 old-name command
  files present
- **WHEN** `uf init` runs and removes them
- **THEN** the summary output SHOULD list the
  removed files with a migration indicator

## MODIFIED Requirements

### Requirement: isDivisorAsset Path

The `isDivisorAsset()` function MUST check for
`"opencode/commands/uf.review-council.md"` instead
of `"opencode/commands/review-council.md"`.

Previously: `relPath == "opencode/commands/review-council.md"`

#### Scenario: Divisor-only mode deploys renamed
command

- **GIVEN** `uf init --divisor` is invoked
- **WHEN** the scaffold engine filters assets via
  `isDivisorAsset()`
- **THEN** the `uf.review-council.md` command file
  MUST be included in the deployed assets

### Requirement: Doctor InstallHint References

All `InstallHint` strings in
`internal/doctor/checks.go` that reference
`/agent-brief` MUST be updated to reference
`/uf.agent-brief`.

Previously: `"Run: /agent-brief in OpenCode"`

#### Scenario: Doctor check suggests correct command

- **GIVEN** a repository missing a bridge file
- **WHEN** the doctor check reports a warning
- **THEN** the `InstallHint` MUST read
  `"Run: /uf.agent-brief in OpenCode"`

### Requirement: Scaffold Hint String

The `hintDivisor` constant MUST reference
`/uf.review-council` instead of `/review-council`.

Previously: `"Run /review-council to start a code review."`

#### Scenario: Post-scaffold hint shows new name

- **GIVEN** `uf init --divisor` completes
- **WHEN** the scaffold summary is printed
- **THEN** the hint MUST read
  `"Run /uf.review-council to start a code review."`

### Requirement: Warning Message Path

The warning message in the `command/` to `commands/`
migration logic MUST reference `/uf.init` instead of
`/uf-init`.

Previously: `"run /uf-init for AI-assisted resolution"`

#### Scenario: Conflict warning shows new command

- **GIVEN** conflicting command files exist in both
  `command/` and `commands/` directories
- **WHEN** the scaffold engine detects the conflict
- **THEN** the warning MUST suggest
  `"run /uf.init for AI-assisted resolution"`

### Requirement: Embedded Asset Paths

The `expectedAssetPaths` list in
`scaffold_test.go` MUST reflect the new `uf.*`
filenames for all 10 renamed commands.

#### Scenario: Drift detection passes

- **GIVEN** the embedded assets use new `uf.*`
  filenames
- **WHEN** `TestAssetPaths_MatchExpected` runs
- **THEN** all 10 renamed paths MUST appear in
  `expectedAssetPaths` and the test MUST pass

## REMOVED Requirements

None.
