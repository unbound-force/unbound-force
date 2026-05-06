## ADDED Requirements

### FR-001: Workflow Dispatch Release Trigger

The release workflow MUST be triggered via workflow_dispatch
with a required `tag` input of type string. The workflow
MUST NOT trigger on tag push events.

#### Scenario: Maintainer triggers release
- **GIVEN** the release workflow exists in GitHub Actions
- **WHEN** a maintainer navigates to Actions, selects
  "Release", clicks "Run workflow", and enters "v0.15.0"
- **THEN** the workflow starts with the tag input set to
  "v0.15.0"

#### Scenario: Tag push does not trigger release
- **GIVEN** a tag `v0.15.0` is pushed to the repository
- **WHEN** GitHub processes the push event
- **THEN** the release workflow MUST NOT be triggered

### FR-002: Pre-flight Tag Format Validation

The preflight job MUST validate that the tag input matches
the pattern `^v[0-9]+\.[0-9]+\.[0-9]+$` (strict semver
with `v` prefix). The workflow MUST abort with a clear
error message if the format is invalid.

#### Scenario: Valid tag format
- **GIVEN** the workflow is triggered with tag "v0.15.0"
- **WHEN** the preflight job validates the tag format
- **THEN** validation passes and the job continues

#### Scenario: Malformed tag rejected
- **GIVEN** the workflow is triggered with tag "v.0.4.3"
- **WHEN** the preflight job validates the tag format
- **THEN** the job fails with message indicating the tag
  does not match the required semver format

#### Scenario: Missing v prefix rejected
- **GIVEN** the workflow is triggered with tag "0.15.0"
- **WHEN** the preflight job validates the tag format
- **THEN** the job fails with message indicating the `v`
  prefix is required

### FR-003: Pre-flight Tag Uniqueness

The preflight job MUST verify that the specified tag does
not already exist in the repository. The workflow MUST
abort if the tag exists.

#### Scenario: New tag accepted
- **GIVEN** tag "v0.15.0" does not exist in the repository
- **WHEN** the preflight job checks tag uniqueness
- **THEN** validation passes

#### Scenario: Duplicate tag rejected
- **GIVEN** tag "v0.14.0" already exists in the repository
- **WHEN** the preflight job checks tag uniqueness
- **THEN** the job fails with message indicating the tag
  already exists

### FR-004: Pre-flight Semver Ordering

The preflight job MUST verify that the new tag version is
greater than the latest existing tag version. The workflow
MUST abort if the new version is not greater.

#### Scenario: Higher version accepted
- **GIVEN** the latest existing tag is "v0.14.0"
- **WHEN** the workflow is triggered with tag "v0.15.0"
- **THEN** validation passes

#### Scenario: Lower version rejected
- **GIVEN** the latest existing tag is "v0.14.0"
- **WHEN** the workflow is triggered with tag "v0.13.0"
- **THEN** the job fails with message indicating the new
  version must be greater than the latest release

### FR-005: Pre-flight CI Status Verification

The preflight job MUST verify that all required status
checks passed on the HEAD commit of the default branch.
Required checks:

- "Build and Test"
- "Standardized CI / Run linters"
- "OSV-Scanner / Trivy Source Scan"
- "OSV-Scanner / OSV-Scanner / osv-scan"

The workflow MUST abort if any required check is not in
`success` conclusion.

#### Scenario: All checks passed
- **GIVEN** HEAD of main has all four required checks in
  `success` state
- **WHEN** the preflight job verifies CI status
- **THEN** validation passes

#### Scenario: Failing check blocks release
- **GIVEN** "Build and Test" check on HEAD is in `failure`
  state
- **WHEN** the preflight job verifies CI status
- **THEN** the job fails with message identifying the
  specific failed check

#### Scenario: Missing check blocks release
- **GIVEN** "Standardized CI / Run linters" check has not
  run on HEAD
- **WHEN** the preflight job verifies CI status
- **THEN** the job fails with message indicating the check
  has not completed

### FR-006: Pre-flight Unreleased Commits

The preflight job MUST verify that there are commits
between the latest existing tag and HEAD. The workflow
MUST abort if there are no unreleased commits.

#### Scenario: Unreleased commits exist
- **GIVEN** 5 commits exist between v0.14.0 and HEAD
- **WHEN** the preflight job checks for unreleased commits
- **THEN** validation passes

#### Scenario: No unreleased commits
- **GIVEN** HEAD is the same commit as v0.14.0
- **WHEN** the preflight job checks for unreleased commits
- **THEN** the job fails with message indicating there are
  no changes to release

### FR-007: Automated Tag Creation

After all pre-flight validations pass, the preflight job
MUST create an annotated tag at HEAD and push it to the
repository using the default GITHUB_TOKEN.

#### Scenario: Tag created and pushed
- **GIVEN** all pre-flight validations pass for "v0.15.0"
- **WHEN** the preflight job creates the tag
- **THEN** an annotated tag "v0.15.0" exists at HEAD and
  is visible in the repository's tag list

### FR-008: SBOM Generation

GoReleaser MUST generate SBOMs via Syft for both archive
artifacts and source. SBOM files MUST be attached to the
GitHub release.

#### Scenario: SBOMs attached to release
- **GIVEN** a release is triggered for v0.15.0
- **WHEN** GoReleaser completes
- **THEN** `.sbom.json` files exist for each archive
  artifact and for the source archive

### FR-009: Cosign Checksum Signing

GoReleaser MUST sign the checksums file using Cosign
keyless signing. The workflow MUST have `id-token: write`
permission for Sigstore OIDC. A `.sigstore.json` bundle
MUST be attached to the release alongside the checksums.

#### Scenario: Checksums signed
- **GIVEN** a release is triggered for v0.15.0
- **WHEN** GoReleaser completes signing
- **THEN** `checksums.txt.sigstore.json` exists as a
  release artifact

### FR-010: Packit Fedora Integration

The repository MUST include a `.packit.yaml` that
configures:

- `copr_build` on pull requests for fedora-rawhide,
  fedora-44, and fedora-43 (x86_64)
- `propose_downstream` on release for rawhide, f44, f43
- `koji_build` on dist-git commit for rawhide, f44, f43
- `bodhi_update` on dist-git commit for f44, f43

#### Scenario: PR triggers COPR build
- **GIVEN** a pull request is opened against main
- **WHEN** Packit detects the PR event
- **THEN** COPR builds are triggered for rawhide, f44,
  and f43

#### Scenario: Release triggers Fedora proposals
- **GIVEN** a GitHub release is published for v0.15.0
- **WHEN** Packit detects the release event
- **THEN** PRs are created in Fedora dist-git for
  rawhide, f44, and f43 branches

### FR-011: Fedora RPM Spec

The repository MUST include an `unbound-force.spec` file
that follows Fedora Packaging Guidelines. The spec MUST:

- Build the `unbound-force` binary from `cmd/unbound-force`
- Install `/usr/bin/unbound-force` and a `/usr/bin/uf`
  symlink
- Use `go build -buildmode=pie` with ldflags for version
  injection. Ldflags variable paths MUST match the
  existing `.goreleaser.yaml` configuration.
- Run `go test -race -count=1 ./...` in the `%check` phase
- Include LICENSE and README.md

#### Scenario: RPM builds successfully
- **GIVEN** the unbound-force.spec and Go source
- **WHEN** `rpmbuild` or COPR processes the spec
- **THEN** an RPM is produced containing
  `/usr/bin/unbound-force` and `/usr/bin/uf` symlink

### FR-012: Release Documentation

A `docs/RELEASE_PROCESS.md` file MUST document:

- How to trigger a release (workflow_dispatch steps)
- What happens automatically (full pipeline description)
- Release cadence expectations (opportunistic)
- Version scheme (semantic versioning)
- Fedora packaging overview and prerequisites
- Packit target maintenance schedule
- Rollback procedures (how to delete a failed release,
  remove the tag, handle Homebrew tap and Fedora update
  state)

#### Scenario: Maintainer follows release process
- **GIVEN** a maintainer reads docs/RELEASE_PROCESS.md
- **WHEN** they follow the documented steps
- **THEN** they can successfully trigger a release with
  one action

### FR-013: Testing Farm Metadata

The repository MUST include `.fmf/version` containing `1`
for Testing Farm metadata compatibility with Packit.

### FR-014: Draft Release Cleanup

Orphaned draft releases MUST be deleted:
- v0.13.0 draft release
- v0.4.1 draft release

#### Scenario: Draft releases removed
- **GIVEN** draft releases v0.13.0 and v0.4.1 exist
- **WHEN** cleanup is performed
- **THEN** only published releases remain

## MODIFIED Requirements

### FR-M01: Release Workflow Trigger

The release workflow trigger changes from
`on: push: tags: ['v*']` to
`on: workflow_dispatch: inputs: tag: required: true`.

Previously: Release was triggered by pushing a `v*` tag.

### FR-M02: GoReleaser Configuration

The `.goreleaser.yaml` adds `sboms` and `signs` top-level
sections. All existing configuration (builds, archives,
changelog, nfpms, brews, homebrew_casks) is unchanged.

Previously: No SBOM or signing configuration existed.

### FR-M03: Release Workflow Permissions

The release job adds `id-token: write` permission for
Cosign Sigstore keyless signing.

Previously: Only `contents: write` was required.

## REMOVED Requirements

### FR-R01: Tag Push Release Trigger

The `on: push: tags: ['v*']` trigger is removed from the
release workflow.

Reason: Replaced by workflow_dispatch trigger with
pre-flight validation. Tags are now created by the
workflow itself, not by maintainers.
