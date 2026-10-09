## MODIFIED Requirements

### Requirement: Scaffold Asset Drift Detection

Scaffold embedded assets MUST be byte-identical to their
canonical source files. Any modification to a canonical
file under `.opencode/commands/`, `.opencode/agents/`,
`.opencode/skills/`, or `.uf/` that has a scaffold copy
under `internal/scaffold/assets/` MUST include a
corresponding update to the scaffold copy.

Previously: Drift detection was implicitly enforced by
`TestEmbeddedAssets_MatchSource` and
`TestSCFR001_CanonicalAssetSurfaceMatchesMirrors`, but the
canonical sources and scaffold copies diverged.

#### Scenario: Canonical command file update syncs scaffold copy

- **GIVEN** the file `.opencode/commands/uf.review-pr.md`
  has been updated with new content
- **WHEN** the scaffold drift test runs
- **THEN** `internal/scaffold/assets/opencode/commands/uf.review-pr.md`
  MUST be byte-identical to `.opencode/commands/uf.review-pr.md`

#### Scenario: Canonical review matrix update syncs scaffold copy

- **GIVEN** the file `.uf/review-matrix.yaml` has been
  updated with v3 schema changes
- **WHEN** the scaffold drift test runs
- **THEN** `internal/scaffold/assets/uf/review-matrix.yaml`
  MUST be byte-identical to `.uf/review-matrix.yaml`

### Requirement: Review Matrix Schema Test Accepts Optional Model

The CI test `TestReviewMatrixSchema_CanonicalPolicyValidates`
MUST validate that the canonical review matrix is
schema-valid, but MUST NOT assert specific model values on
profiles that intentionally use empty model maps for
host-resolved dispatch.

Previously: The test asserted specific model strings
(`opencode-go/qwen3.8-flash`, ...) for each profile, which
conflicted with the v3 schema change that made `model`
optional.

#### Scenario: Empty model on profile passes validation

- **GIVEN** the canonical review matrix defines a profile
  with `model: {}` (empty, host-resolved)
- **WHEN** `TestReviewMatrixSchema_CanonicalPolicyValidates`
  runs
- **THEN** the profile MUST validate against the v3 schema
- **AND** the test MUST NOT assert a specific non-empty
  model string for that profile

#### Scenario: Profile with explicit model still validates

- **GIVEN** the canonical review matrix defines a profile
  with a specific model name
- **WHEN** `TestReviewMatrixSchema_CanonicalPolicyValidates`
  runs
- **THEN** the profile MUST validate against the v3 schema
- **AND** the test MUST assert the model value matches the
  configured model name
