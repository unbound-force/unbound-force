## ADDED Requirements

### Requirement: force-regenerate-sources

`uf init --force` MUST regenerate `.dewey/sources.yaml`
before re-indexing, even if the file has been previously
customized.

#### Scenario: force regenerates sources
- **GIVEN** `.dewey/sources.yaml` exists with custom
  entries
- **WHEN** the user runs `uf init --force` with dewey
  available
- **THEN** `sources.yaml` is overwritten with the
  auto-detected config (including `recursive: false`
  on `disk-org`) and `dewey index` runs with the new
  config

#### Scenario: non-force preserves custom sources
- **GIVEN** `.dewey/sources.yaml` exists with custom
  entries
- **WHEN** the user runs `uf init` (no `--force`)
- **THEN** `sources.yaml` is not modified (existing
  behavior preserved)

## MODIFIED Requirements

### Requirement: disk-org-recursive

The `disk-org` entry in `.dewey/sources.yaml` MUST
include `recursive: false` to prevent recursive
indexing of the parent directory.

Previously: The `disk-org` entry had no `recursive`
setting, causing Dewey to recursively index all
contents of the parent directory including sibling
repos that are already indexed individually.

#### Scenario: disk-org is non-recursive
- **GIVEN** `generateDeweySources()` creates a
  `sources.yaml` with a `disk-org` entry
- **WHEN** the generated YAML is examined
- **THEN** the `disk-org` entry contains
  `recursive: false` under its `config` section

#### Scenario: per-repo entries remain recursive
- **GIVEN** `generateDeweySources()` detects sibling
  repos
- **WHEN** the generated YAML is examined
- **THEN** the per-repo `disk-<name>` entries do NOT
  contain `recursive: false` (they remain recursive
  by default)

## REMOVED Requirements

None.
