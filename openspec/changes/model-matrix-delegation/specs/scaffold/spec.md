<!-- capability: scaffold
     change: model-matrix-delegation
     originating_issue: 887 -->

Requirement numbering: SC-FR-001 through SC-FR-006 are allocated
by prior changes (review-council-multi-model-fanout and
successors). SC-FR-007 is allocated here; SC-FR-008 through
SC-FR-018 are reserved for future scaffold changes. Hero-repo
deployment of the model matrix via `uf init` (pack templates,
`alwaysDeployedPacks`, per-repo matrix provisioning) is slice 2
(issue #889) and is NOT required by this requirement — this
requirement covers the asset surface registration that slice 1
needs so the mirrored plugin loads.

## ADDED Requirements

### Requirement: Model-Matrix Scaffold Asset Surface

**ID: SC-FR-007** This requirement MUST be satisfied.

The scaffold MUST register the model-matrix asset surface so the
mirrored uf-workflow plugin loads in scaffolded hero repos and
drift detection stays authoritative:

- `.opencode/lib/model-matrix.ts` MUST be a tool-owned scaffold
  asset — embedded copy under
  `internal/scaffold/assets/opencode/lib/model-matrix.ts`,
  enumerated in the lib asset lists, and covered by the
  canonical-assets drift table. This mirror is REQUIRED, not
  optional: the mirrored uf-workflow plugin imports the mirrored
  lib module, so a missing lib mirror breaks plugin load in
  scaffolded repos;
- `.opencode/plugins/uf-workflow/index.ts` MUST remain
  byte-identical to its embedded mirror (existing drift tests
  enforce this; the plugin's model-selection modifications must
  be synced);
- `.uf/model-matrix.yaml` MUST gain an embedded scaffold template
  under `internal/scaffold/assets/uf/model-matrix.yaml` mapping
  embedded `uf/` paths to runtime `.uf/` paths (the established
  `mapAssetPath` contract);
- the model-matrix JSON Schema registry MUST be embedded under
  `internal/scaffold/assets/uf/schemas/model-matrix/`,
  byte-identical to the root `schemas/model-matrix/` registry
  (MM-FR-008);
- the managed gitignore block MUST include
  `.uf/model-matrix.override.yaml`, and the repository root
  `.gitignore` MUST carry the same entry.

Drift tests MUST fail and name the mismatch when any canonical
asset differs from its mirror. The classifier subagent file
(`.opencode/agents/model-matrix-classifier.md`) and the
meta-prompt pack are canonical repo content in this slice; their
scaffold deployment is deferred to slice 2 (#889) and MUST NOT
be blocked by this requirement.

#### Scenario: Lib mirror is registered and drift-enforced

- **GIVEN** the canonical `.opencode/lib/model-matrix.ts` and
  its embedded copy differ by one byte
- **WHEN** the canonical-assets drift tests run
- **THEN** the tests fail and name the mismatched path

#### Scenario: Mirrored plugin imports the mirrored lib

- **GIVEN** a scaffolded hero repo deployed with the uf-workflow
  plugin and the model-matrix lib asset
- **WHEN** the plugin loads
- **THEN** the module import resolves without a missing-module
  failure

#### Scenario: Matrix template maps to the runtime path

- **GIVEN** the embedded asset `uf/model-matrix.yaml`
- **WHEN** `mapAssetPath` maps it
- **THEN** the result is `.uf/model-matrix.yaml`

#### Scenario: Override file is gitignored

- **GIVEN** a repository scaffolded with the managed gitignore
  block
- **WHEN** the block is inspected
- **THEN** it contains `.uf/model-matrix.override.yaml`

## REMOVED Requirements

None.
