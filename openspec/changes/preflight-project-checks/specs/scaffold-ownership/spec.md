<!-- capability: scaffold-ownership
     change: preflight-project-checks -->

Requirement numbering: this change allocates SO-FR-001
and SO-FR-002 for the two requirements below. The
scaffold-ownership capability's earlier requirements
(three, currently unnumbered) live in the unarchived
openspec/changes/reviewer-capabilities-init-survival
change and will need SO-FR ids assigned before that
change archives; SO-FR-003 onward stay reserved for
that reconciliation and future scaffold-ownership
changes.

## ADDED Requirements

### Requirement: Pre-flight check declarations are a project-owned extension point

**ID: SO-FR-001** This requirement MUST be satisfied.

The scaffold MUST classify `uf/preflight.yaml` as project-owned:
- created from the embedded template when absent;
- skipped on plain `uf init` re-runs when present;
- replaced with the template only under `--force`.

The embedded template MUST declare `version: 1` and an empty
`checks` list, with commented field documentation. The template
MUST be byte-identical to this repository's deployed
`.uf/preflight.yaml`.

#### Scenario: First deploy seeds the template

- **GIVEN** a target directory without `.uf/preflight.yaml`
- **WHEN** `uf init` runs
- **THEN** the file is created from the embedded template and
  reported in `Created`

#### Scenario: Declared checks survive re-init

- **GIVEN** a deployed `.uf/preflight.yaml` with locally
  declared checks
- **WHEN** `uf init` re-runs without `--force`
- **THEN** the file is reported in `Skipped` and its content is
  unchanged

#### Scenario: Force resets to the template

- **GIVEN** a deployed `.uf/preflight.yaml` with local checks
- **WHEN** `uf init --force` runs
- **THEN** the file is replaced with the embedded template

### Requirement: Declared-check library ships with the review tooling

**ID: SO-FR-002** This requirement MUST be satisfied.

The scaffold MUST classify `opencode/lib/preflight-declared.ts`
as tool-owned, MUST include it in the Divisor-only asset subset,
and MUST stage it with the review plugin sources. That way
`uf_preflight_declared_checks` resolves its import wherever the
uf-workflow plugin is deployed.

#### Scenario: Divisor-only deploy includes the library

- **GIVEN** a Divisor-only `uf init`
- **WHEN** assets are deployed
- **THEN** `.opencode/lib/preflight-declared.ts` is deployed
  and `.uf/preflight.yaml` is not part of the Divisor subset
