## ADDED Requirements

### Requirement: Persistent workspace working directory

`buildPersistentRunArgs()` MUST set `--workdir` to
`/workspace/<project-basename>` where `<project-basename>`
is `filepath.Base(opts.ProjectDir)`.

`buildPersistentRunArgs()` MUST set the `WORKSPACE`
environment variable to the same path via
`-e WORKSPACE=/workspace/<project-basename>`.

This ensures the container image's entrypoint
(`cd "$WORKSPACE"`) starts OpenCode in the project
directory, not the parent `/workspace/` directory.

#### Scenario: Create persistent Podman workspace

- **GIVEN** the user runs
  `uf sandbox create --backend podman` from a project
  directory named `my-project`
- **WHEN** `buildPersistentRunArgs()` constructs the
  `podman run` arguments
- **THEN** the arguments MUST include
  `--workdir /workspace/my-project`
- **AND** the arguments MUST include
  `-e WORKSPACE=/workspace/my-project`

#### Scenario: OpenCode starts in project directory

- **GIVEN** a persistent Podman workspace was created
  for a project named `my-project`
- **WHEN** the container starts and the entrypoint runs
- **THEN** the entrypoint MUST `cd` to
  `/workspace/my-project`
- **AND** `opencode serve` MUST start with
  `/workspace/my-project` as the working directory

#### Scenario: Resumed workspace retains workdir

- **GIVEN** a persistent Podman workspace was created
  with `--workdir /workspace/my-project` and
  `WORKSPACE=/workspace/my-project`
- **WHEN** the workspace is stopped and resumed via
  `uf sandbox start`
- **THEN** the container MUST retain the original
  `--workdir` and `WORKSPACE` env var
- **AND** OpenCode MUST start in
  `/workspace/my-project`

## MODIFIED Requirements

(none)

## REMOVED Requirements

(none)
