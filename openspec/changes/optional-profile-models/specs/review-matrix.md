## ADDED Requirements

> **Numbering**: RD-FR-019 through RD-FR-024 are allocated for review-matrix
> delta requirements. RD-FR-009 through RD-FR-018 are reserved for future
> review-matrix changes in other OpenSpec deltas. This change builds on parent
> change `review-council-multi-model-fanout` (RD-FR-001 through RD-FR-008).

### Requirement: RD-FR-019 — Optional Profile Model

The `model` field within each profile entry in `.uf/review-
matrix.yaml` is now optional. When `model` is omitted from a
profile, all review runs dispatched via that profile MUST use
the parent agent's model (host resolution) rather than an
explicitly specified model.

#### Scenario: Profile with no model falls back to host
- **GIVEN** a review matrix with `profiles.standard: {}` (no
  model field) and `defaults.code: standard`
- **WHEN** a review council is dispatched for a code change
- **THEN** each advisor-generated run for code-review agents
  has `model: null` in its plan entry
- **AND** the run uses the parent session's model via host
  resolution in `invoke_agent`

#### Scenario: Profile with explicit model uses that model
- **GIVEN** a review matrix with `profiles.standard.model:
  opencode-go/deepseek-v4-pro` and `defaults.code: standard`
- **WHEN** a review council is dispatched for a code change
- **THEN** each advisor-generated run for code-review agents
  has `model: "opencode-go/deepseek-v4-pro"` in its plan entry
- **AND** the run uses that model explicitly via `invoke_agent`

### Requirement: RD-FR-020 — Review Matrix Override File

The review dispatch plugin MUST support a `.uf/review-matrix.
override.yaml` file. When this file exists, the override is
constrained to `profiles.*.model` and `profiles.*.variant`
fields only (whitelist). Any other key or nested path in the
override MUST cause validation failure, preventing silent
weakening of team governance (e.g., replacing `defaults`,
`always`, or tier structure).

Whitelisted override fields perform a nested-path merge with the
base `.uf/review-matrix.yaml` before schema validation. Fields
present in the override replace their counterparts in the
base; fields absent in the override fall through to the base.

The override file MUST be gitignored and MUST NOT be tracked
in version control.

#### Scenario: Override supplies model for a profile
- **GIVEN** a tracked review matrix with `profiles.standard:
  {}` (no model)
- **AND** a gitignored override file with `profiles.standard.
  model: opencode-go/deepseek-v4-pro`
- **WHEN** `loadPolicies()` reads the matrix
- **THEN** the merged result has `profiles.standard.model:
  "opencode-go/deepseek-v4-pro"`
- **AND** the merged result passes schema validation

#### Scenario: Override not present, base used as-is
- **GIVEN** a tracked review matrix with `profiles.standard.
  model: opencode-go/qwen3.8-flash`
- **AND** no override file exists
- **WHEN** `loadPolicies()` reads the matrix
- **THEN** the base matrix is used without modification

#### Scenario: Invalid override fails validation
- **GIVEN** a valid base review matrix
- **AND** an override file with invalid syntax or schema
- **WHEN** `loadPolicies()` reads and merges the matrix
- **THEN** the merged result fails schema validation
- **AND** the error message identifies the validation failure

#### Scenario: Override absent, no profile models set, soft gate
- **GIVEN** a review matrix where no profile has a `model`
  field
- **AND** no override file is present
- **WHEN** `loadPolicies()` reads the matrix
- **THEN** the matrix loads successfully
- **AND** all profile-driven runs use host model resolution
- **AND** a soft-gate warning SHOULD be logged (not a hard
  failure)

#### Scenario: Override contains non-whitelisted key
- **GIVEN** a valid base review matrix
- **AND** an override file with `defaults.code:
  nonstandard` (non-whitelisted key `defaults`)
- **WHEN** `loadPolicies()` reads and validates the
  override
- **THEN** the override is rejected with a validation
  error at the `defaults` key path
- **AND** the error message states `.uf/review-matrix.
  override.yaml` only permits `profiles.*.model` and
  `profiles.*.variant`

#### Scenario: Empty override file is no-op
- **GIVEN** a tracked review matrix with explicit models
- **AND** an override file that exists but is empty
  (zero bytes)
- **WHEN** `loadPolicies()` reads and merges
- **THEN** the base matrix is used without modification

#### Scenario: Comments-only override file is no-op
- **GIVEN** a tracked review matrix with explicit models
- **AND** an override file containing only YAML comments
  (no key-value pairs)
- **WHEN** `loadPolicies()` reads and merges
- **THEN** the base matrix is used without modification

#### Scenario: Override sets model to null explicitly
- **GIVEN** a tracked review matrix with
  `profiles.standard.model: opencode-go/grok-4.7`
- **AND** an override file with `profiles.standard.
  model: null` (null YAML value)
- **WHEN** `loadPolicies()` reads and merges
- **THEN** the merged result has `profiles.standard.
  model: null`
- **AND** the null value propagates to
  `profilePair()`, triggering host fallback

#### Scenario: Override file unreadable or unparseable
- **GIVEN** a valid base review matrix
- **AND** an override file exists but contains malformed
  YAML or is unreadable
- **WHEN** `loadPolicies()` attempts to read the override
- **THEN** the dispatch plan MUST return `INCONCLUSIVE`
- **AND** the error message identifies the override file
  as the failure source and includes the read error
  detail

### Requirement: RD-FR-021 — All-or-Nothing Merge Result

After nested-path merge of the override file, the combined result
MUST be validated against the identical `ReviewMatrixSchema`
used for the base file. If the merged result fails validation,
the dispatch plan MUST fall back to the `INCONCLUSIVE` workflow
result with an error message indicating the override merge
failure.

This ensures any invalid override (e.g., referencing an unknown
profile in `defaults`, violating limits, missing required fields)
is caught before dispatch begins.

#### Scenario: Override adds invalid profile reference
- **GIVEN** a valid base matrix
- **AND** an override that sets `profiles.nonexistent.model:
  foo` where `nonexistent` is not a defined tier key in the
  base matrix's `profiles` section
- **WHEN** `loadPolicies()` reads and merges
- **THEN** schema validation fails with an error at path
  `profiles.nonexistent`

### Requirement: RD-FR-022 — Profile Model Resolution

The `profilePair()` function SHALL return `model: null` when
the resolved profile does not contain a `model` field. Callers
SHALL treat `model: null` identically to host-fallback runs
— the parent session's model is used at invocation time.

#### Scenario: profilePair returns null model
- **GIVEN** a profile entry `standard: {}` (no model)
- **WHEN** `profilePair(matrix, "standard")` is called
- **THEN** the result is `{ model: null, variant: null,
  tier: "standard" }`

#### Scenario: Null model resolves to host at dispatch
- **GIVEN** a dispatch plan entry with `model: null`
  (profile had no model)
- **WHEN** `invoke_agent` is called with `model: null`
- **THEN** the child session resolves the parent session's
  model at runtime
- **AND** the HIC envelope records the resolved model in
  `reported_model` and `resolved_parent_model`

### Requirement: RD-FR-023 — Review Matrix Schema Version

The review matrix schema version SHALL increment from `2` to
`3`. Version `2` matrices use `schemas/review-matrix/v2.
schema.json` where `model` remains required. Version `3`
matrices use `schemas/review-matrix/v3.schema.json` where
`model` is optional. Both schema versions coexist — this
change does not retire v2.

#### Scenario: Version 2 matrix loads under v2 schema
- **GIVEN** a version-2 review matrix with all model fields
  set
- **WHEN** `parseReviewMatrix()` validates it against
  `v2.schema.json`
- **THEN** the matrix parses successfully

#### Scenario: Version 2 matrix missing model field fails
- **GIVEN** a version-2 review matrix where a profile omits
  the `model` field
- **WHEN** `parseReviewMatrix()` validates it against
  `v2.schema.json`
- **THEN** validation fails with an error at the missing
  field path

#### Scenario: Version 3 matrix omits model field succeeds
- **GIVEN** a version-3 review matrix where a profile omits
  the `model` field
- **WHEN** `parseReviewMatrix()` validates it against
  `v3.schema.json`
- **THEN** the matrix parses successfully

## REMOVED Requirements

None.

## ADDED Requirements (continued)

### Requirement: RD-FR-024 — Review Matrix Version History

The review matrix was previously defined at version `2` with
`model` required in every profile entry. The `ProfileSchema`
in `schemas/review-matrix/v2.schema.json` enforces this with
`.min(3)` on the `model` field. This change makes the `model`
field optional (per RD-FR-019) and increments the schema
version to `3` (per RD-FR-023). The two schema versions
coexist — v2 supports legacy validation where `model` is
required, and v3 enables optional models.

Previously: `ProfileSchema.model` was required; every profile
had to specify a concrete model string. The schema enforced
this with `.min(3)` on the model field. Version was `2`. No
parent Speckit spec exists for review-dispatch; this
requirement serves as the self-documenting version history.