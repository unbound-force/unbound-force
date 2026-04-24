## ADDED Requirements

### Requirement: WORKSPACE Environment Variable (FR-044)

When parent mount is active, `buildRunArgs()` MUST pass
`-e WORKSPACE=/workspace/<project-basename>` to the
container so the entrypoint's `cd "$WORKSPACE"` sets
the correct working directory for OpenCode.

#### Scenario: Parent mount sets WORKSPACE

- **GIVEN** `ProjectDir` is
  `/Users/j/Projects/org/myproject`
- **AND** `NoParent` is false
- **WHEN** `buildRunArgs()` constructs the podman args
- **THEN** the args include
  `-e WORKSPACE=/workspace/myproject`

#### Scenario: No-parent mode omits WORKSPACE

- **GIVEN** `NoParent` is true
- **WHEN** `buildRunArgs()` constructs the podman args
- **THEN** the args do NOT include `-e WORKSPACE=...`
  (entrypoint uses its default `/workspace`)

#### Scenario: Root fallback omits WORKSPACE

- **GIVEN** `ProjectDir` is `/myproject`
  (parent is `/`)
- **WHEN** `buildRunArgs()` constructs the podman args
- **THEN** the args do NOT include `-e WORKSPACE=...`
  (parent mount is not active due to root fallback)

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
