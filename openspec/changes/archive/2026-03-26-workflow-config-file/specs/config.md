## ADDED Requirements

### Requirement: workflow-config-file

A project-level config file at
`.unbound-force/config.yaml` MUST be supported for
storing default workflow execution mode overrides.

#### Scenario: config file with define=swarm

- **GIVEN** `.unbound-force/config.yaml` contains
  `workflow.execution_modes.define: swarm`
- **WHEN** `Start()` is called with no CLI overrides
- **THEN** the workflow's define stage has
  `execution_mode=swarm`

#### Scenario: CLI overrides config

- **GIVEN** config has `define: swarm`
- **WHEN** `Start()` is called with
  `overrides={"define": "human"}`
- **THEN** the workflow's define stage has
  `execution_mode=human` (CLI wins)

#### Scenario: config file missing

- **GIVEN** `.unbound-force/config.yaml` does not exist
- **WHEN** `Start()` is called
- **THEN** all execution modes use
  `StageExecutionModeMap()` defaults (define=human)

#### Scenario: config file malformed

- **GIVEN** `.unbound-force/config.yaml` contains
  invalid YAML
- **WHEN** `Start()` is called
- **THEN** a warning is logged and all defaults are
  used (no error returned)

#### Scenario: config with spec_review

- **GIVEN** config has `workflow.spec_review: true`
- **WHEN** `Start()` is called with `specReview=false`
- **THEN** spec review is enabled (config OR CLI)

### Requirement: scaffold-config-file

`uf init` MUST create `.unbound-force/config.yaml`
with commented-out defaults when the file does not
exist.

#### Scenario: fresh scaffold

- **GIVEN** `.unbound-force/config.yaml` does not exist
- **WHEN** `uf init` runs
- **THEN** the file is created with all values
  commented out

#### Scenario: re-scaffold preserves config

- **GIVEN** `.unbound-force/config.yaml` exists with
  uncommented values
- **WHEN** `uf init` runs again
- **THEN** the file is NOT overwritten (user-owned)

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
