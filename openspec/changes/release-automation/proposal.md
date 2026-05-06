## Why

The current release process requires 10 manual steps spanning
two PRs, local git operations, and manual file editing before
automation kicks in. This friction is incompatible with the
project's need for frequent releases to deliver contribution
value quickly. Manual steps have produced errors: malformed
tags (`v.0.4.3`), orphaned draft releases (v0.13.0, v0.4.1),
and CHANGELOG maintenance that blocks releases.

The project also lacks supply chain security artifacts (SBOM,
Sigstore signing) and has no pathway into Fedora package
repositories despite already producing RPMs via nfpm.

Reference implementation: ComplyTime/complyctl uses
workflow_dispatch + GoReleaser + Cosign + Packit to achieve
one-click releases with full Fedora integration.

## What Changes

Replace the tag-push-triggered release workflow with a
workflow_dispatch-triggered pipeline that automates tag
creation, adds pre-flight validation, integrates supply chain
security tooling, and enables Fedora packaging via Packit.

## Capabilities

### New Capabilities
- `workflow-dispatch-release`: One-click release trigger via
  GitHub Actions UI with version tag input
- `pre-flight-validation`: Automated checks before release
  (semver format, tag uniqueness, semver ordering, CI status
  verification, unreleased commit detection)
- `automated-tag-creation`: Tag creation and push handled
  inside the workflow, eliminating local git operations
- `sbom-generation`: Software Bill of Materials via Syft for
  archive and source artifacts
- `cosign-signing`: Sigstore keyless signing of checksums
  with Rekor transparency log entries
- `fedora-packaging`: Packit integration for automated Fedora
  dist-git proposals, Koji builds, and Bodhi updates
  (targeting rawhide, f44, f43)
- `fedora-rpm-spec`: RPM spec file following Fedora Packaging
  Guidelines for `dnf install unbound-force`
- `release-documentation`: docs/RELEASE_PROCESS.md setting
  expectations for trigger, automation, cadence, and Fedora
- `draft-release-cleanup`: Remove orphaned draft releases
  (v0.13.0 draft, v0.4.1 draft)

### Modified Capabilities
- `release-workflow`: Trigger changes from `push tags v*` to
  `workflow_dispatch`. Adds preflight job, Cosign and Syft
  installation. Existing macOS signing job unchanged except
  dependency wiring.
- `goreleaser-config`: Adds `sboms` and `signs` top-level
  sections. Existing builds, archives, changelog, nfpms,
  brews, and homebrew_casks unchanged.

### Removed Capabilities
- `manual-tag-push-trigger`: Tag push no longer triggers
  releases. Tags are created by the workflow itself.

## Impact

- `.github/workflows/release.yml`: Major redesign (new
  trigger, new preflight job, new tool installations)
- `.goreleaser.yaml`: Two new sections (sboms, signs)
- `.packit.yaml`: New file (Fedora packaging automation)
- `unbound-force.spec`: New file (Fedora RPM spec)
- `docs/RELEASE_PROCESS.md`: New file (release documentation)
- `.fmf/version`: New file (Testing Farm metadata)
- CHANGELOG.md is explicitly decoupled from releases (no
  changes, no dependency)
- nfpm RPM (GitHub release artifact) coexists with Fedora
  spec RPM (dnf repos) — they serve different users
- Fedora package review (Bugzilla) is a prerequisite for
  Packit propose_downstream; copr_build works immediately
- Default GITHUB_TOKEN used for tag creation (tags attributed
  to github-actions[bot])

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

Release artifacts (binaries, SBOMs, signatures, checksums)
are self-describing with provenance metadata injected by
GoReleaser ldflags (version, commit, date). Packit consumes
the GitHub Release event asynchronously to propose Fedora
packaging without requiring synchronous interaction. The
release workflow produces artifacts that downstream consumers
(Homebrew tap, Fedora dist-git, users) consume independently.

### II. Composability First

**Assessment**: PASS

The release workflow is independently functional. Packit
integration is additive — if Packit is not installed or
the Fedora package review is incomplete, the core release
(binaries, Homebrew, macOS signing) works without it. The
nfpm RPM and Fedora spec RPM coexist without conflict.
Each distribution channel operates independently.

### III. Observable Quality

**Assessment**: PASS

Adds SBOM (Syft) and Cosign signing, providing
machine-parseable supply chain artifacts with provenance.
Pre-flight validation produces clear pass/fail signals
with specific failure messages. GoReleaser changelog groups
commits by conventional commit type. All artifacts include
version, commit, and build date metadata.

### IV. Testability

**Assessment**: PASS

Pre-flight validation logic is testable via the GitHub API
(check-runs endpoint). The release workflow can be validated
via dry-run (GoReleaser `--snapshot`). Packit copr_build on
PRs validates the RPM spec before any release. The workflow
structure (preflight → release → sign-macos) isolates
concerns for independent verification.
