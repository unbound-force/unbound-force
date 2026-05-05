## ADDED Requirements

### Requirement: Command Directory Migration

The scaffold engine MUST migrate existing
`.opencode/command/` directories to `.opencode/commands/`
during `uf init`. The migration function MUST run after
`initSubTools()` completes. The migration MUST be
idempotent.

#### Scenario: Fresh repo with no command directory
- **GIVEN** a project directory with no `.opencode/command/`
  or `.opencode/commands/` directory
- **WHEN** `uf init` runs
- **THEN** embedded command assets are deployed to
  `.opencode/commands/` and no migration is reported

#### Scenario: Existing repo with only old directory
- **GIVEN** a project directory with `.opencode/command/`
  containing N command files and no `.opencode/commands/`
- **WHEN** `uf init` runs
- **THEN** `.opencode/command/` is renamed to
  `.opencode/commands/` atomically, the migration summary
  reports "migrated" with file count, and
  `.opencode/command/` no longer exists

#### Scenario: Both directories exist with unique files
- **GIVEN** `.opencode/command/` contains files A and B,
  and `.opencode/commands/` contains files C and D
- **WHEN** `uf init` runs
- **THEN** files A and B are moved to `.opencode/commands/`,
  `.opencode/command/` is removed, and
  `.opencode/commands/` contains A, B, C, and D

#### Scenario: Both directories with identical duplicate
- **GIVEN** `.opencode/command/x.md` and
  `.opencode/commands/x.md` exist with identical content
- **WHEN** `uf init` runs
- **THEN** the copy in `.opencode/command/` is removed,
  the copy in `.opencode/commands/` is preserved unchanged

#### Scenario: Both directories with conflicting duplicate
- **GIVEN** `.opencode/command/x.md` and
  `.opencode/commands/x.md` exist with different content
- **WHEN** `uf init` runs
- **THEN** the `.opencode/commands/` version is kept,
  the `.opencode/command/` version is removed, and a
  warning is printed mentioning `/uf-init` for
  AI-assisted resolution

#### Scenario: Old directory is a symlink
- **GIVEN** `.opencode/command` is a symbolic link
- **WHEN** `uf init` runs
- **THEN** migration is skipped with a warning indicating
  manual migration is required

#### Scenario: DivisorOnly mode
- **GIVEN** `uf init` is invoked with `--divisor` flag
- **WHEN** `uf init` runs
- **THEN** no migration is attempted and no migration
  result is reported

#### Scenario: Partial failure during merge
- **GIVEN** `.opencode/command/` contains files A, B, C
  and file B has restrictive permissions preventing move
- **WHEN** `uf init` runs
- **THEN** files A and C are moved to `.opencode/commands/`,
  a warning is printed for file B, and `.opencode/command/`
  is NOT removed (still contains B)

#### Scenario: Re-run after successful migration
- **GIVEN** a previous `uf init` run successfully migrated
  all files to `.opencode/commands/`
- **WHEN** `uf init` runs again
- **THEN** no migration result is reported (silent no-op)

### Requirement: File Move with Fallback

Individual file moves MUST use `os.Rename()` first. If
`os.Rename()` fails, the function MUST fall back to
read -> write -> remove.

#### Scenario: Same-filesystem move
- **GIVEN** source and destination are on the same
  filesystem
- **WHEN** a file is moved during migration
- **THEN** `os.Rename()` succeeds and the file appears
  at the destination

#### Scenario: Cross-filesystem move
- **GIVEN** source and destination are on different
  filesystems (e.g., bind mounts)
- **WHEN** `os.Rename()` fails with a cross-device error
- **THEN** the fallback reads the file content, writes
  it to the destination, and removes the source

## MODIFIED Requirements

### Requirement: Embedded Asset Directory

The embedded command assets MUST be stored at
`internal/scaffold/assets/opencode/commands/` (plural).

Previously: stored at
`internal/scaffold/assets/opencode/command/` (singular).

#### Scenario: Asset deployment path
- **GIVEN** the embedded assets include
  `opencode/commands/review-council.md`
- **WHEN** `uf init` runs
- **THEN** the file is deployed to
  `.opencode/commands/review-council.md`

### Requirement: Tool Ownership Classification

`isToolOwned()` MUST classify files under
`opencode/commands/` as tool-owned.

Previously: classified files under `opencode/command/`.

### Requirement: Divisor Asset Classification

`isDivisorAsset()` MUST recognize
`opencode/commands/review-council.md` as a Divisor asset.

Previously: recognized `opencode/command/review-council.md`.

### Requirement: Doctor Command Directory Check

`uf doctor` MUST check `.opencode/commands/` as the
primary command directory. If `.opencode/command/` (legacy)
exists, `uf doctor` SHOULD report a warning recommending
`uf init` to migrate.

Previously: checked only `.opencode/command/`.

### Requirement: Hero Contract Validation

`validate-hero-contract.sh` MUST accept
`.opencode/commands/` as the canonical command directory.
The script MUST also accept `.opencode/command/` (legacy)
with a deprecation warning. The script MUST fail only
when neither directory exists.

Previously: required `.opencode/command/`.

### Requirement: /uf-init Command Paths

The `/uf-init` slash command MUST reference
`.opencode/commands/` for all command file paths. A new
Step 0 MUST perform command directory migration before
other steps execute.

Previously: referenced `.opencode/command/` throughout.

## REMOVED Requirements

None.
