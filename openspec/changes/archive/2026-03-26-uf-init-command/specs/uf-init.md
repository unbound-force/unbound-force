## ADDED Requirements

### Requirement: uf-init-command

A `/uf-init` slash command MUST exist that applies
project-specific customizations to third-party tool
files using LLM reasoning.

#### Scenario: fresh run (no customizations present)

- **GIVEN** OpenSpec skill and command files exist
  without any Unbound Force customizations
- **WHEN** the user runs `/uf-init` in an OpenCode
  session
- **THEN** all 14 customizations are inserted at
  correct locations and a summary reports 14 applied

#### Scenario: re-run (all customizations present)

- **GIVEN** all customizations were previously applied
- **WHEN** the user runs `/uf-init` again
- **THEN** all 14 customizations are detected as
  already present and the summary reports 0 applied,
  14 skipped

#### Scenario: after OpenSpec update

- **GIVEN** the OpenSpec CLI was updated, overwriting
  skill files (removing customizations)
- **WHEN** the user runs `/uf-init`
- **THEN** the customizations in skill files are
  re-applied while command files (not overwritten by
  OpenSpec) report as already present

#### Scenario: missing OpenSpec files

- **GIVEN** OpenSpec is not installed (skill files
  don't exist)
- **WHEN** the user runs `/uf-init`
- **THEN** errors are reported for each missing file
  with a fix suggestion ("Run `uf setup` to install
  OpenSpec, then `uf init`"), and the command continues
  checking remaining files

#### Scenario: uf init not run

- **GIVEN** `.opencode/` directory does not exist
- **WHEN** the user runs `/uf-init`
- **THEN** the command errors: "`.opencode/` not found.
  Run `uf init` from your terminal first."

### Requirement: scaffold-uf-init-command

`uf init` MUST scaffold `.opencode/command/uf-init.md`
as a tool-owned asset.

#### Scenario: scaffold includes command

- **GIVEN** a fresh project directory
- **WHEN** `uf init` runs
- **THEN** `.opencode/command/uf-init.md` is created

#### Scenario: re-scaffold updates command

- **GIVEN** an older version of `uf-init.md` exists
- **WHEN** `uf init` runs again
- **THEN** the file is overwritten with the latest
  version (tool-owned)

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
