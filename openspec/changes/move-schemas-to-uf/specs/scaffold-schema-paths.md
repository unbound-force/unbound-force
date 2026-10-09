## MODIFIED Requirements

### Requirement: Scaffold schema assets MUST deploy under `.uf/schemas/`

The `uf init` command SHALL deploy embedded JSON schema assets to
`.uf/schemas/` in the target project directory. Previously, schema
assets deployed to `schemas/`.

#### Scenario: uf init deploys schemas to .uf/schemas/

- **GIVEN** a consumer project initialized with `uf init`
- **WHEN** the scaffold processes embedded schema assets from
  `internal/scaffold/assets/uf/schemas/`
- **THEN** schema files SHALL be written to `.uf/schemas/` in the
  target project, under their original relative path (e.g.,
  `.uf/schemas/review-verdict/v1.0.0.schema.json`)
- **AND** the target project's top-level `schemas/` directory (if
  any) SHALL NOT be modified by the scaffold

#### Scenario: Stealth mode hides .uf/schemas/ via existing exclusion

- **GIVEN** a project using `uf init --stealth`
- **WHEN** the scaffold deploys schema assets to
  `.uf/schemas/review-verdict/v1.0.0.schema.json`
- **THEN** `.git/info/exclude` SHALL contain `.uf/` as an excluded
  path
- **AND** the deployed schema files SHALL be hidden from git
- **AND** the exclusion block SHALL NOT contain a separate
  `schemas/` entry

#### Scenario: Tool-owned schema files are overwritten on diff

- **GIVEN** a previously scaffolded `.uf/schemas/review-matrix/v3.schema.json`
- **WHEN** the schema content has changed in the embedded assets
- **THEN** the file SHALL be overwritten (tool-owned behavior)
- **AND** user-created `.uf/schemas/` files not matching an
  embedded asset SHALL NOT be affected

#### Scenario: Divisor-only mode still deploys schemas

- **GIVEN** `uf init --divisor` is run
- **WHEN** the scaffold filters assets to Divisor-related content
- **THEN** schema assets SHALL still be deployed to
  `.uf/schemas/` (schemas are not filtered by Divisor-only mode)

### Requirement: Scaffold path mapping SHALL NOT include a top-level `schemas/` prefix

The `mapAssetPath` function SHALL NOT contain a case for the
`schemas/` prefix. The `knownAssetPrefixes` variable SHALL NOT
include `"schemas/"`. Schema assets SHALL be deployed under the
existing `uf/` → `.uf/` mapping.

#### Scenario: Unknown asset prefix test coverage

- **GIVEN** an embedded asset under `uf/schemas/review-verdict/v1.0.0.schema.json`
- **WHEN** `mapAssetPath` processes the path
- **THEN** the `uf/` prefix case SHALL match
- **AND** the output SHALL be `.uf/schemas/review-verdict/v1.0.0.schema.json`

## REMOVED Requirements

### Requirement: Scaffold deploys schemas to top-level `schemas/`

The previous behavior deployed embedded schema assets to a
top-level `schemas/` directory in the target project. This
requirement is removed because it caused namespace collisions
with projects that maintain their own `schemas/` directory.

<!-- scaffolded by uf vdev -->