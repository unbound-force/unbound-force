## ADDED Requirements

### Requirement: CI Convention Pack Deployment

FR-001: The scaffold engine MUST deploy `ci.md` and
`ci-custom.md` to `.opencode/uf/packs/` for all projects,
regardless of detected language.

FR-002: `ci.md` MUST be tool-owned (auto-updated on
`uf init` re-runs). `ci-custom.md` MUST be user-owned
(never overwritten unless `--force` is used).

FR-003: `ci.md` and `ci-custom.md` MUST be included in
the `collectDeployedPacks()` candidates list alongside
default, severity, and content packs.

FR-004: `shouldDeployPack()` MUST return true for `ci`
and `ci-custom` pack names regardless of the resolved
language.

#### Scenario: Pack deployed for Go project

- **GIVEN** a project with `go.mod` present
- **WHEN** `uf init` is run
- **THEN** `.opencode/uf/packs/ci.md` and
  `.opencode/uf/packs/ci-custom.md` are created alongside
  `default.md`, `severity.md`, `content.md`, and `go.md`

#### Scenario: Pack deployed for project with no detected language

- **GIVEN** a project with no language marker files
- **WHEN** `uf init` is run
- **THEN** `.opencode/uf/packs/ci.md` and
  `.opencode/uf/packs/ci-custom.md` are created alongside
  default, severity, and content packs

#### Scenario: Custom pack preserved on re-run

- **GIVEN** a project where `ci-custom.md` has been
  populated with project-specific rules
- **WHEN** `uf init` is run again
- **THEN** `ci-custom.md` is NOT overwritten
- **AND** `ci.md` IS updated to the latest canonical version

### Requirement: CI Convention Pack Structure

FR-010: The `ci.md` pack MUST have YAML frontmatter with
`pack_id: ci`, `language: Any`, and `version: 1.0.0`.

FR-011: The `ci.md` pack MUST contain the following H2
sections: `Action Pinning & Supply Chain`,
`Workflow Structure`, `Permissions & Secrets`,
`Reusable Workflow Design`, `Custom Rules`.

FR-012: The `ci.md` pack MUST be exempt from the pack
validator's standard 6-section requirement (Coding Style,
Architectural Patterns, etc.), following the established
exemption pattern for `severity.md` and `content.md`.

#### Scenario: Pack validator skips ci.md

- **GIVEN** the `ci.md` pack exists in `.opencode/uf/packs/`
- **WHEN** `TestValidateConventionPack_AllPacksValid` runs
- **THEN** `ci.md` is skipped (not validated against the
  6-section coding pack requirement)
- **AND** no validation error is reported

### Requirement: Drift Detection

FR-020: The embedded copy at
`internal/scaffold/assets/opencode/uf/packs/ci.md` MUST be
byte-identical to `.opencode/uf/packs/ci.md`.

FR-021: The embedded copy at
`internal/scaffold/assets/opencode/uf/packs/ci-custom.md`
MUST be byte-identical to `.opencode/uf/packs/ci-custom.md`.

FR-022: Both files MUST be listed in the `expectedAssetPaths`
manifest in `scaffold_test.go`.

#### Scenario: Drift detected when embedded diverges

- **GIVEN** `ci.md` is modified in `.opencode/uf/packs/`
  but not in `internal/scaffold/assets/opencode/uf/packs/`
- **WHEN** `TestEmbeddedAssets_MatchSource` runs
- **THEN** the test fails with a drift error

### Requirement: Action Pinning Rules

FR-030: CI-001 MUST require that every `uses:` reference
for actions and reusable workflows in GitHub Actions
workflow files is pinned to a full 40-character commit SHA.
Docker container references (`docker://image:tag`) are
excluded from this requirement; they SHOULD be pinned to
image digests (`@sha256:...`) when available but are not
subject to CI-001.

FR-031: CI-002 MUST require that SHAs are verified
against the source repository at authoring time. The tag
MUST be resolved first, and the SHA derived from it. SHAs
MUST NOT be sourced from memory or training data. When the
API response object type is `tag` (annotated tag), the
agent MUST dereference to get the commit SHA by following
the tag object's target.

FR-032: CI-003 SHOULD recommend including a version
comment (e.g., `# v6.0.2`) after the SHA for human
readability. Missing comments are a warning, not a
failure. When a version comment IS included, it MUST
accurately reflect the tag used to derive the SHA.

FR-033: CI-002 verification MUST fail safe. When SHA
verification cannot be completed (network unavailable,
API rate limit, missing authentication, tag not found),
the agent MUST NOT write the unverified `uses:` reference
and MUST report the verification failure.

#### Scenario: Agent writes a new action reference

- **GIVEN** an agent is adding `actions/checkout` to a
  workflow file
- **WHEN** the agent follows CI-002
- **THEN** the agent resolves the desired tag via
  `gh api repos/actions/checkout/git/ref/tags/v4`
- **AND** if the response object type is `tag` (annotated
  tag), the agent dereferences to get the commit SHA
- **AND** derives the commit SHA from the API response
- **AND** writes `uses: actions/checkout@<sha>  # v4.x.x`

#### Scenario: Agent writes SHA from memory

- **GIVEN** an agent writes a `uses:` reference with a
  SHA sourced from training data
- **WHEN** a Divisor agent reviews the change or the CI
  guardrail runs
- **THEN** the reviewer flags a CI-002 violation if the
  SHA does not resolve or does not match the claimed tag

#### Scenario: SHA verification fails

- **GIVEN** an agent is adding an action reference
- **WHEN** the `gh api` call fails (network, auth, rate
  limit, missing tag)
- **THEN** the agent MUST NOT write the `uses:` reference
- **AND** the agent MUST report the verification failure

### Requirement: Non-Pinning CI Rules

FR-034: The Permissions & Secrets section (CI-020, CI-021,
CI-022) MUST codify the CI workflow permission and secret
management rules. The normative source for these rules is
the "Security" subsection of the "YAML / GitHub Actions
Workflows" section in the project's coding standards.
CI-020 MUST require least-privilege permissions. CI-021
SHOULD prefer job-level over workflow-level permissions.
CI-022 MUST prohibit hardcoded secrets.

FR-035: The Workflow Structure section (CI-010, CI-011,
CI-012) MUST codify CI workflow naming and formatting
rules. The normative source is the "Naming Conventions"
and "Formatting" subsections of the coding standards.
CI-010 MUST require `reusable_` prefix for reusable
workflows and `ci_` prefix for consumers. CI-011 MUST
require header comment blocks. CI-012 SHOULD recommend
concurrency groups.

FR-036: The Reusable Workflow Design section (CI-030,
CI-031, CI-032) MUST codify reusable workflow interface
rules. CI-030 MUST require descriptive input descriptions
and `required: true` for required inputs. CI-031 SHOULD
recommend sensible defaults. CI-032 MUST require reusable
workflows to be generic enough for any consumer.

### Requirement: Coding Standards Centralization

FR-040: The `ci.md` pack MUST be the canonical source of
truth for CI workflow conventions in repos that use the
Unbound Force scaffold.

FR-041: The "YAML / GitHub Actions Workflows" section in
`~/.config/opencode/coding-standards.md` MUST be replaced
with a condensed inline summary of essential CI rules
(pin to SHA, least-privilege, naming) plus a reference to
the `ci.md` pack for expanded enforceable rules. This
preserves baseline guidance for non-UF repos.

FR-042: The "Containers" section in `coding-standards.md`
MUST remain unchanged (different concern).

#### Scenario: Non-UF repo retains baseline CI guidance

- **GIVEN** a developer opens a non-UF repo (no
  `.opencode/uf/packs/ci.md` present)
- **WHEN** the agent reads `coding-standards.md`
- **THEN** essential CI baseline rules (pin to SHA,
  least-privilege, naming) are available inline
- **AND** a note indicates expanded rules are in the
  `ci.md` pack for UF repos

#### Scenario: UF repo uses pack as canonical source

- **GIVEN** the `ci.md` pack is deployed in a UF repo
- **WHEN** an agent reads both `coding-standards.md` and
  the `ci.md` pack
- **THEN** the pack contains the full CI rule set
- **AND** `coding-standards.md` defers to the pack for
  expanded rules

## MODIFIED Requirements

### Requirement: collectDeployedPacks candidates list

Previously: candidates list contained 5 always-deployed
entries (default, default-custom, severity, content,
content-custom).

Now: candidates list MUST contain 7 always-deployed entries
(default, default-custom, severity, content, content-custom,
ci, ci-custom).

### Requirement: shouldDeployPack always-deploy set

Previously: always-deploy set contained default,
default-custom, severity, content, content-custom.

Now: always-deploy set MUST also contain ci and ci-custom.

### Requirement: expectedAssetPaths manifest

Previously: convention packs section listed 11 entries.

Now: convention packs section MUST list 13 entries,
including `opencode/uf/packs/ci.md` and
`opencode/uf/packs/ci-custom.md`.

## REMOVED Requirements

None.
