## Why

The release pipeline duplicates ~190 lines of inline
validation and GoReleaser setup logic that org-infra
already maintains as reusable workflows. This creates
a maintenance burden: when org-infra improves preflight
validation (e.g., smart re-run detection, spec-compliant
semver comparison), each consuming repo must independently
reimplement the improvement or fall behind.

The inline preflight has known limitations:
- Uses `sort -V` for semver ordering (breaks on
  pre-releases)
- Hardcoded CI check names (no configurable input)
- No re-run resilience (tag uniqueness check blocks
  re-runs after partial failures)
- Security scan gate uses OR logic (at least one scan)
  rather than explicit named checks

complyctl and complypack already consume org-infra
reusable workflows for release. Adopting the same pattern
across all unbound-force repositories unifies the release
pipeline and reduces per-repo maintenance.

## What Changes

Replace the inline release preflight and GoReleaser jobs
in `.github/workflows/release.yml` with calls to org-infra
reusable workflows (`reusable_release_preflight` +
`reusable_release_goreleaser`). The `sign-macos` job stays
inline because it is repo-specific (binary name, Homebrew
tap structure, signing identity).

## Capabilities

### New Capabilities
- `reusable-preflight`: Delegate preflight validation to
  org-infra's `reusable_release_preflight` workflow with
  smart re-run detection and spec-compliant semver
  comparison
- `reusable-goreleaser`: Delegate GoReleaser execution
  to org-infra's `reusable_release_goreleaser` workflow
  with supply chain artifacts (Cosign + Syft)
- `skip-inputs`: Three boolean inputs for debugging
  (`skip_semver_check`, `skip_ci_checks`,
  `skip_unreleased_check`)
- `extra-files-upload`: GoReleaser `release.extra_files`
  for atomic Homebrew artifact upload (replaces manual
  `gh release upload` post-step)
- `full-secret-validation`: Signing secrets check
  expanded to verify all 6 required secrets with
  actionable `::notice::` annotations for missing ones

### Removed Capabilities
- `inline-preflight`: ~130 lines of inline validation
  logic replaced by reusable workflow caller
- `inline-goreleaser`: ~60 lines of inline GoReleaser
  setup replaced by reusable workflow caller
- `manual-artifact-upload`: `gh release upload` step
  replaced by GoReleaser `release.extra_files`

### Unchanged Capabilities
- `sign-macos`: macOS signing, notarization, checksum
  patching, and Homebrew tap push remain inline
- `fedora-packaging`: Packit integration unchanged

## Impact

- **Risk**: LOW — mechanical adoption of existing
  patterns already validated in complyctl/complypack
- **Scope**: `.github/workflows/release.yml`,
  `.goreleaser.yaml`, `CHANGELOG.md`,
  `docs/RELEASE_PROCESS.md`
- **Dependencies**: `complytime/org-infra` @ v0.7.1
  (SHA-pinned)
- **Cross-repo**: Issue #428 covers 4 repositories;
  this change covers `unbound-force/unbound-force` only

## Acceptance Criteria

- [x] Inline preflight replaced with
  `reusable_release_preflight` caller (SHA-pinned)
- [x] Inline GoReleaser replaced with
  `reusable_release_goreleaser` caller (SHA-pinned)
- [x] `sign-macos` receives `tag` from preflight
  outputs and `has_signing_secrets` from secrets check
- [x] GoReleaser-generated cask and formula uploaded as
  release assets via `release.extra_files`
- [x] `skip_semver_check`, `skip_ci_checks`,
  `skip_unreleased_check` inputs wired through
- [x] Security scan check (`OSV-Scanner / Trivy Source
  Scan`) included in `ci_checks` input
- [x] Concurrency group prevents parallel releases
- [x] All 6 signing secrets verified before `sign-macos`
- [x] `docs/RELEASE_PROCESS.md` updated with new
  pipeline, skip controls, re-run docs, troubleshooting,
  and org-infra upgrade path
- [x] `CHANGELOG.md` entry under `## Unreleased`
