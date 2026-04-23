## ADDED Requirements

### Requirement: Parent Directory Mount (FR-040)

The sandbox MUST mount the project's parent directory
at `/workspace` by default, and set the container's
working directory to `/workspace/<project-basename>`
using the `--workdir` flag.

#### Scenario: Default parent mount

- **GIVEN** `ProjectDir` is
  `/Users/j/Projects/org/myproject`
- **WHEN** `uf sandbox start` is run without
  `--no-parent`
- **THEN** the container volume mount is
  `-v /Users/j/Projects/org:/workspace`
- **AND** the container working directory is
  `--workdir /workspace/myproject`

#### Scenario: Sibling directory accessible

- **GIVEN** the parent directory contains sibling repos
  `myproject/`, `dewey/`, `gaze/`
- **WHEN** a tool inside the container reads
  `../dewey/README.md`
- **THEN** the file is accessible at
  `/workspace/dewey/README.md`

### Requirement: No-Parent Flag (FR-041)

The `uf sandbox start` command MUST accept a
`--no-parent` flag that disables parent directory
mounting. When `--no-parent` is set, the sandbox
MUST use the current behavior: mount `ProjectDir`
directly at `/workspace` without `--workdir`.

#### Scenario: Project-only mount with flag

- **GIVEN** `ProjectDir` is
  `/Users/j/Projects/org/myproject`
- **WHEN** `uf sandbox start --no-parent` is run
- **THEN** the container volume mount is
  `-v /Users/j/Projects/org/myproject:/workspace`
- **AND** no `--workdir` flag is set

### Requirement: Root Directory Fallback (FR-042)

When `filepath.Dir(ProjectDir)` returns `/` (project
is at the filesystem root), the sandbox MUST fall back
to project-only mounting (same as `--no-parent`) and
SHOULD log a debug message explaining the fallback.

#### Scenario: Project at filesystem root

- **GIVEN** `ProjectDir` is `/myproject`
- **WHEN** `uf sandbox start` is run
- **THEN** the container volume mount is
  `-v /myproject:/workspace`
- **AND** no `--workdir` flag is set
- **AND** the behavior matches `--no-parent`

### Requirement: Mode Preservation (FR-043)

The parent directory mount MUST respect the existing
mount mode: read-only (`:ro`) in isolated mode,
read-write in direct mode. The SELinux `:Z` flag
MUST be applied when `platform.SELinux` is true.

#### Scenario: Isolated mode with parent mount

- **GIVEN** `Mode` is `isolated`
- **WHEN** the sandbox starts with parent mount
- **THEN** the volume mount includes `:ro`
  (e.g., `-v /parent:/workspace:ro`)

#### Scenario: Direct mode with parent mount

- **GIVEN** `Mode` is `direct`
- **WHEN** the sandbox starts with parent mount
- **THEN** the volume mount does NOT include `:ro`

#### Scenario: SELinux with parent mount

- **GIVEN** `platform.SELinux` is true and `Mode`
  is `direct`
- **WHEN** the sandbox starts with parent mount
- **THEN** the volume mount includes `,Z`
  (e.g., `-v /parent:/workspace:Z`)

## MODIFIED Requirements

### Requirement: Volume Mount Target (FR-028 from Spec 028)

The sandbox volume mount target MUST be the project's
parent directory (not the project directory itself)
by default. The `--workdir` flag MUST be set to
`/workspace/<project-basename>` to maintain the
correct working directory for OpenCode.

Previously: The sandbox mounted `ProjectDir` directly
at `/workspace` with no `--workdir` flag.

## REMOVED Requirements

None.
