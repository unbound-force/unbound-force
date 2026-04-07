## ADDED Requirements

### Requirement: ensureGitignore function

The scaffold engine MUST include an `ensureGitignore()`
function that appends a standard Unbound Force ignore
block to `.gitignore` in the target directory.

The ignore block MUST cover:
- `.uf/` runtime data (databases, caches, locks, logs,
  orchestration state, metrics data)
- Legacy tool directories (`.dewey/`, `.hive/`,
  `.unbound-force/`, `.muti-mind/`, `.mx-f/`)

The function MUST be idempotent — if the marker comment
`# Unbound Force — managed by uf init` is already
present, the function MUST skip the append.

The function MUST create `.gitignore` if it does not
exist.

#### Scenario: Fresh directory with no .gitignore

- **GIVEN** a directory with no `.gitignore` file
- **WHEN** `uf init` runs
- **THEN** `.gitignore` is created containing the UF
  ignore block with the marker comment

#### Scenario: Existing .gitignore without UF block

- **GIVEN** a directory with an existing `.gitignore`
  containing project-specific patterns
- **WHEN** `uf init` runs
- **THEN** the UF ignore block is appended to the end
  of `.gitignore` with a blank line separator
- **AND** existing content is preserved unchanged

#### Scenario: Existing .gitignore with UF block

- **GIVEN** a directory with a `.gitignore` that already
  contains the UF marker comment
- **WHEN** `uf init` runs
- **THEN** `.gitignore` is not modified (idempotent)

#### Scenario: Ignore block content

- **GIVEN** the UF ignore block is appended
- **THEN** it contains patterns for:
  - `.uf/workflows/` and `.uf/artifacts/`
  - `.uf/dewey/` runtime (graph.db, cache/, locks, log)
  - `.uf/replicator/` runtime (databases, locks)
  - `.uf/muti-mind/artifacts/`
  - `.uf/mx-f/data/`
  - Legacy directories (`.dewey/`, `.hive/`,
    `.unbound-force/`, `.muti-mind/`, `.mx-f/`)

## MODIFIED Requirements

### Requirement: Run() function

The `Run()` function in the scaffold engine MUST call
`ensureGitignore()` after file scaffolding and config
generation, but before sub-tool delegation.

The result MUST be included in the scaffold summary
(created/skipped action).

## REMOVED Requirements

None.
