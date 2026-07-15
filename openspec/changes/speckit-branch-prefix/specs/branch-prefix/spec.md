## ADDED Requirements

### FR-001: Speckit branch folder prefix

New Speckit feature branches MUST be created with the
`speckit/` folder prefix. The branch name format SHALL be
`speckit/NNN-<name>` where `NNN` is a zero-padded 3-digit
sequential number and `<name>` is the kebab-case feature
name.

#### Scenario: New feature branch creation

- **GIVEN** a developer runs the Speckit `create-new-feature`
  script with description "add user auth"
- **WHEN** the script generates the branch name
- **THEN** the branch name MUST be `speckit/NNN-user-auth`
  (e.g., `speckit/035-user-auth`)
- **AND** the spec directory on disk MUST be
  `specs/NNN-user-auth/` (without the `speckit/` prefix)

### FR-002: Backward-compatible branch detection

All branch-detection logic MUST accept both the legacy
`NNN-<name>` pattern and the new `speckit/NNN-<name>`
pattern. Legacy branches MUST NOT be rejected or require
migration.

#### Scenario: Legacy branch detected by unleash command

- **GIVEN** a developer is on branch `031-unleash-openspec`
  (legacy format)
- **WHEN** the `/unleash` command detects the branch type
- **THEN** the command MUST identify it as a Speckit branch
- **AND** the command MUST locate spec artifacts at
  `specs/031-unleash-openspec/`

#### Scenario: New-format branch detected by unleash command

- **GIVEN** a developer is on branch
  `speckit/035-user-auth` (new format)
- **WHEN** the `/unleash` command detects the branch type
- **THEN** the command MUST identify it as a Speckit branch
- **AND** the command MUST locate spec artifacts at
  `specs/035-user-auth/` (prefix stripped)

### FR-003: Prefix stripping for path construction

When mapping a branch name to a filesystem path, the
`speckit/` prefix MUST be stripped before constructing the
spec directory path. The stripping MUST occur in
`common.sh` functions (`find_feature_dir_by_prefix`,
`get_current_branch`) so that downstream consumers
receive a clean `NNN-<name>` identifier.

#### Scenario: Prefix stripped for spec directory lookup

- **GIVEN** the current branch is `speckit/035-user-auth`
- **WHEN** `find_feature_dir_by_prefix()` resolves the
  feature directory
- **THEN** it MUST return `$REPO_ROOT/specs/035-user-auth`
- **AND** it MUST NOT return
  `$REPO_ROOT/specs/speckit/035-user-auth`

#### Scenario: Legacy branch needs no stripping

- **GIVEN** the current branch is `031-unleash-openspec`
- **WHEN** `find_feature_dir_by_prefix()` resolves the
  feature directory
- **THEN** it MUST return
  `$REPO_ROOT/specs/031-unleash-openspec`

### FR-004: Spec directory naming unchanged

Spec directories on disk MUST continue to use the
`NNN-<name>` naming convention under `specs/`. The
`speckit/` prefix applies only to git branch names, not
to filesystem directories.

#### Scenario: Directory created without prefix

- **GIVEN** the `create-new-feature` script runs for a new
  feature
- **WHEN** the spec directory is created
- **THEN** the directory MUST be at
  `specs/NNN-<name>/` (e.g., `specs/035-user-auth/`)
- **AND** the directory MUST NOT be at
  `specs/speckit/NNN-<name>/`

## MODIFIED Requirements

### Requirement: Branch naming convention documentation

All documentation, commands, skills, and agent files that
reference the Speckit branch convention MUST be updated to
show `speckit/NNN-<name>` as the primary format. Legacy
`NNN-<name>` SHOULD be noted as accepted for backward
compatibility.

Previously: Documentation referenced `NNN-<name>` as the
sole branch naming convention for Speckit.

### Requirement: Scaffold asset synchronization

All scaffold assets under `internal/scaffold/assets/` that
mirror live command, skill, or agent files MUST be updated
in sync with their live counterparts. Scaffold drift
detection tests MUST pass after the update.

Previously: Scaffold assets referenced `NNN-<name>` patterns.

## REMOVED Requirements

None.
