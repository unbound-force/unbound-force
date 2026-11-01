<!-- capability: scaffold
     change: model-matrix-scaffold-deployment
     originating_issue: 889 -->

Requirement numbering: SC-FR-001 through SC-FR-006 are allocated
by prior changes (review-council-multi-model-fanout and
successors); SC-FR-007 is allocated by the
model-matrix-delegation change (slice 1, issue #887), whose
delta explicitly leaves hero-repo deployment of the model matrix
to slice 2. SC-FR-008 through SC-FR-010 are allocated here;
SC-FR-011 through SC-FR-018 remain reserved for future scaffold
changes. Sibling capabilities: matrix loading and selection
semantics live in specs/model-matrix/spec.md (MM-FR-001 through
MM-FR-008, unchanged by this change); the delegation-tool
integration lives in specs/unleash/spec.md (UF-FR-001,
UF-FR-002, unchanged).

## ADDED Requirements

### Requirement: Model-Matrix Pack and Classifier Deployment

**ID: SC-FR-008** This requirement MUST be satisfied.

The scaffold MUST deploy the model-matrix pack surface and
classifier subagent that the slice-1 asset registration
references but does not ship:

- `.opencode/uf/packs/meta-prompt.md` MUST have a
  byte-identical embedded mirror under
  `internal/scaffold/assets/opencode/uf/packs/meta-prompt.md`,
  covered by the canonical-assets drift table; the embedded
  mirror and canonical file MUST be registered in the embedded
  asset list. This is load-bearing: the tracked starter matrix
  maps the `meta-prompt` pack, the matrix loader fails closed
  when a `packs:` key has no pack file (MM-FR-001), and the
  full scaffold deploys every embedded pack — a hero repo
  whose matrix names `meta-prompt` without the pack file has a
  structurally invalid matrix;
- a `meta-prompt-custom.md` initial-deployment template MUST
  exist canonically at `.opencode/uf/packs/` and in the
  embedded assets, following the existing `*-custom.md`
  pattern (frontmatter, custom-rule prefix guidance, sentinel
  Custom Rules stub, no version marker in the canonical
  file). It is user-owned and exempt from drift checks under
  the existing custom-pack skip rule;
- `alwaysDeployedPacks` MUST list `meta-prompt` and
  `meta-prompt-custom`, and `collectDeployedPacks` MUST list
  `meta-prompt.md` and `meta-prompt-custom.md`, so the pack
  deploys regardless of detected language and the scaffolded
  AGENTS.md Convention Packs section directs agents to it;
- `.opencode/agents/model-matrix-classifier.md` MUST have a
  byte-identical embedded mirror under
  `internal/scaffold/assets/opencode/agents/`, registered in
  the embedded asset list and the canonical-assets drift
  table, and deployed by full `uf init` runs. Agent files
  remain user-owned on re-init per the existing agent
  ownership precedent;
- the `knownNonEmbeddedFiles` entries slice 1 added for
  `.opencode/uf/packs/meta-prompt.md` and
  `.opencode/agents/model-matrix-classifier.md` (marked
  deferred to this slice) MUST be removed.

#### Scenario: Fresh full scaffold deploys the pack surface

- **GIVEN** an empty target directory
- **WHEN** `uf init` runs in full (non-DivisorOnly) mode
- **THEN** the run's created list includes
  `.uf/model-matrix.yaml`,
  `.uf/schemas/model-matrix/v1.schema.json`,
  `.opencode/uf/packs/meta-prompt.md`,
  `.opencode/uf/packs/meta-prompt-custom.md`, and
  `.opencode/agents/model-matrix-classifier.md`

#### Scenario: Canonical pack drift fails the suite

- **GIVEN** the canonical `.opencode/uf/packs/meta-prompt.md`
  differs by one byte from its embedded mirror
- **WHEN** the canonical-assets drift tests run
- **THEN** the tests fail and name the mismatched path

#### Scenario: DivisorOnly deploys the packs, not the matrix

- **GIVEN** `uf init` runs with DivisorOnly mode
- **WHEN** the asset walk filters deployments
- **THEN** `.opencode/uf/packs/meta-prompt.md` and
  `.opencode/uf/packs/meta-prompt-custom.md` deploy (DivisorOnly
  deploys all convention packs), while `.uf/model-matrix.yaml`
  and `.uf/schemas/model-matrix/v1.schema.json` do not (the
  model matrix serves content-agent delegation, not review
  dispatch)

### Requirement: Deployed Matrix Load Validity

**ID: SC-FR-009** This requirement MUST be satisfied.

The scaffold MUST keep the deployed model-matrix surface
self-consistent: every `packs:` key of the embedded
`uf/model-matrix.yaml` template MUST have a corresponding
embedded pack file under
`internal/scaffold/assets/opencode/uf/packs/`. Because the
full scaffold deploys every embedded pack, embedded presence
implies deployed presence, and a template satisfying this
invariant yields hero-repo matrices that pass the loader's
pack-existence rule (MM-FR-001). A scaffold test MUST enforce
the invariant by parsing the embedded template and failing with
an error that names the pack file path of every `packs:` key
without an embedded pack file.

#### Scenario: Guard fails naming the missing pack

- **GIVEN** the embedded `uf/model-matrix.yaml` template names
  a `packs:` key whose pack file is not embedded
- **WHEN** the deployment completeness guard test runs
- **THEN** the test fails and names the missing
  `.opencode/uf/packs/<key>.md` path

#### Scenario: Starter set satisfies the guard

- **GIVEN** the starter template's `packs:` keys
  (`meta-prompt`, `go`, `typescript`, `python`, `content`) and
  the embedded pack set including `meta-prompt.md`
- **WHEN** the deployment completeness guard test runs
- **THEN** the test passes with no missing pack files

### Requirement: Deployed Matrix Re-Init Ownership

**ID: SC-FR-010** This requirement MUST be satisfied.

The deployed `.uf/model-matrix.yaml` MUST be a project-owned
extension point, mirroring the `uf/review-matrix.yaml` and
`uf/reviewer-capabilities.yaml` init-survival precedent
(issues #884/#886):

- a fresh `uf init` run MUST seed `.uf/model-matrix.yaml` from
  the embedded template (the new-file creation path);
- a plain `uf init` re-run MUST NOT overwrite a deployed
  `.uf/model-matrix.yaml` whose content differs from the
  template — `isToolOwned("uf/model-matrix.yaml")` MUST return
  false, so team-edited tracked matrices survive re-init, and
  `--force` remains the explicit path that replaces them;
- `.uf/schemas/model-matrix/` files and canonical convention
  packs (including `meta-prompt.md`) MUST remain tool-owned
  and refresh on re-run when content differs;
- `*-custom.md` packs and the gitignored
  `.uf/model-matrix.override.yaml` MUST remain preserved and
  never scaffolded beyond the gitignore entry.

#### Scenario: Team-edited matrix survives re-init

- **GIVEN** a repo scaffolded by `uf init` whose tracked
  `.uf/model-matrix.yaml` was edited by the team after
  deployment
- **WHEN** `uf init` re-runs without `--force`
- **THEN** the tracked matrix appears in the run's skipped
  list with its edited content unchanged

#### Scenario: Tool-owned matrix files refresh on diff

- **GIVEN** a scaffolded repo whose deployed
  `.uf/schemas/model-matrix/v1.schema.json` (or canonical
  `meta-prompt.md`) differs from the embedded copy
- **WHEN** `uf init` re-runs without `--force`
- **THEN** the differing file appears in the run's updated
  list and matches the embedded content afterward

#### Scenario: Identical content produces no spurious updates

- **GIVEN** a scaffolded repo whose tool-owned matrix files
  are byte-identical to the embedded copies
- **WHEN** `uf init` re-runs without `--force`
- **THEN** those files appear in the skipped list, not the
  updated list

#### Scenario: Force replaces the team-edited matrix

- **GIVEN** a scaffolded repo with a team-edited tracked
  `.uf/model-matrix.yaml`
- **WHEN** `uf init` re-runs with `--force`
- **THEN** the tracked matrix is overwritten with the embedded
  template content

## MODIFIED Requirements

None. SC-FR-007 (slice 1) registered the asset surface and its
delta text explicitly leaves hero-repo deployment semantics to
this slice, so no prior requirement's contract changes; the
deployed-copy ownership classification this change defines was
outside SC-FR-007's stated scope.

## REMOVED Requirements

None.
