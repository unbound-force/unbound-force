## Context

The release process currently requires 10 manual steps before
automation begins: CHANGELOG update, PR creation, review,
merge, local tag creation, and tag push. The workflow triggers
on tag push (`on: push: tags: ['v*']`), meaning all pre-tag
work is unautomated. The project lacks SBOM generation,
Sigstore signing, and Fedora packaging integration.

Reference: ComplyTime/complyctl uses workflow_dispatch +
GoReleaser + Cosign + Syft + Packit for one-click releases
with full Fedora integration.

## Goals / Non-Goals

### Goals
- Reduce human release action to one step: click "Run
  workflow" and type a version tag
- Add pre-flight validation to prevent malformed tags,
  duplicate releases, and releases on broken builds
- Add SBOM generation (Syft) and checksum signing (Cosign)
  for supply chain security
- Add Fedora packaging via Packit (rawhide, f44, f43)
- Document the release process in docs/RELEASE_PROCESS.md
- Clean up orphaned draft releases
- Decouple CHANGELOG.md from the release process

### Non-Goals
- Pre-release/alpha tag support
- CentOS Stream targets for Packit
- Homebrew tap PR-based flow (direct push is sufficient)
- Malformed tag cleanup (v.0.4.3, vanilla — manual)
- Release cadence formalization (stays opportunistic)
- Changes to the macOS signing flow (works as-is)

## Decisions

### D1: workflow_dispatch replaces tag push trigger

The release workflow trigger changes from `on: push: tags`
to `on: workflow_dispatch` with a required `tag` string
input. The workflow creates and pushes the tag itself after
pre-flight validation passes.

**Rationale**: Eliminates local git operations. Pre-flight
validation prevents errors before the tag exists. One human
action (click + type) replaces five manual steps.

**Constitution**: Supports Autonomous Collaboration — the
release pipeline operates as a self-contained workflow
producing self-describing artifacts.

### D2: Three-job pipeline (preflight → release → sign-macos)

The workflow uses three sequential jobs:

1. **preflight** (ubuntu-latest): validates tag format,
   uniqueness, semver ordering, CI status on HEAD, and
   unreleased commit count. Creates and pushes the tag.
2. **release** (ubuntu-latest): runs GoReleaser with Cosign
   and Syft. Uploads Homebrew cask/formula artifacts.
3. **sign-macos** (macos-latest, conditional): existing macOS
   signing flow, unchanged except dependency wiring.

**Rationale**: Isolating validation from build allows early
failure with clear messages. The existing sign-macos job is
proven and needs no modification — only its `needs` clause
changes.

### D3: Strict CI verification in pre-flight

Pre-flight verifies ALL required status checks passed on
the HEAD commit of main before allowing the release:

- "Build and Test" (ci_local.yml)
- "Standardized CI / Run linters" (ci_checks.yml)
- "OSV-Scanner / Trivy Source Scan" (ci_security.yml)
- "OSV-Scanner / OSV-Scanner / osv-scan" (ci_security.yml)

If any required check is not in `success` state, the
workflow aborts with a specific failure message.

**Rationale**: Prevents releasing broken builds. The current
process relies on human memory to verify CI before tagging.

### D4: Default GITHUB_TOKEN for tag creation

The workflow uses the default `GITHUB_TOKEN` (with
`contents: write` permission) to create and push tags.
Tags are attributed to `github-actions[bot]`.

**Rationale**: Simplest approach. No additional secrets or
GitHub App setup required. Tag attribution to the bot is
acceptable since the workflow_dispatch event itself records
which human triggered it.

### D5: Cosign keyless signing of checksums

Cosign signs the `checksums.txt` file using Sigstore keyless
signing with `id-token: write` permission. Produces a
`.sigstore.json` bundle alongside the checksum file.

**Rationale**: Matches complyctl's approach. Keyless signing
requires no key management — uses GitHub OIDC identity.
Signing the checksum file transitively covers all release
artifacts.

**Constitution**: Supports Observable Quality — adds
machine-parseable cryptographic provenance.

### D6: Syft SBOM for archives and source

GoReleaser generates SBOMs via Syft for both `archive`
artifacts and `source`. Produces `.sbom.json` files attached
to the GitHub release.

**Rationale**: Standard supply chain security practice.
Matches complyctl's approach.

### D7: nfpm RPM coexists with Fedora spec RPM

The existing nfpm RPM (attached to GitHub releases) is kept.
A new Fedora-compliant `unbound-force.spec` is added for
Packit. They serve different distribution channels:

- nfpm RPM: Quick `rpm -i` from GitHub releases
- Fedora spec: `dnf install` from Fedora repos

**Rationale**: Different users, different needs. The nfpm
RPM is simpler and available immediately. The Fedora spec
follows Packaging Guidelines and requires a review process.

**Constitution**: Supports Composability First — each
distribution channel operates independently.

### D8: Packit targets rawhide, f44, f43

Fedora 42 reaches EOL on 2026-05-13 and is excluded. The
Packit configuration targets:

- rawhide (always active, becomes f45)
- f44 (EOL: 2027-06-02)
- f43 (EOL: 2026-12-02)

Targets MUST be updated with each Fedora release cycle
(~6 months). The release documentation notes this.

### D9: CHANGELOG.md decoupled from releases

CHANGELOG.md remains as an AI agent context artifact,
maintained by agents during PR merges. It is NOT a release
gate, NOT updated as part of the release process, and NOT
consumed by GoReleaser.

GoReleaser auto-generates release changelogs from
conventional commits.

**Rationale**: Eliminates the highest-friction manual step
(paragraph-length changelog entries) from the release path
while preserving the value for AI agents.

### D10: Fedora package review prerequisite

Packit `propose_downstream` requires an approved Fedora
package. This is a one-time process:

1. File Bugzilla review request
2. Fedora reviewer reviews the spec
3. Package approved and created in dist-git
4. Packit automation activates

The `copr_build` job works immediately on PRs without
this prerequisite. The release documentation explains
this dependency.

## Risks / Trade-offs

### R1: Fedora package review timeline

The Fedora package review process takes 2-4 weeks. During
this period, releases will work fully except for the Fedora
dist-git proposal step. This is acceptable — the core
release pipeline (binaries, Homebrew, macOS signing, SBOM,
Cosign) is independent.

### R2: Tag creation in workflow

If the workflow fails after creating the tag but before
completing the release, an orphaned tag exists. Mitigation:
the preflight job creates the tag as its last step, after
all validation passes. GoReleaser handles tag-exists
scenarios gracefully.

### R3: CI check name stability

Pre-flight validation hardcodes CI check names. If workflow
names change, pre-flight will fail. Mitigation: check names
are stable (defined in branch protection) and the failure
message will indicate which check is missing.

### R4: Packit target maintenance

Fedora release targets must be updated ~every 6 months.
Mitigation: documented in RELEASE_PROCESS.md. The release
will not fail if a target is EOL — Packit handles this
gracefully.

### R5: No vendor directory

Unlike complyctl, UF uses Go modules without vendoring.
The Fedora spec builds with `go build` directly. If the
Fedora build environment has network restrictions, this
may require adding vendoring to the spec. Mitigation: most
Go packages in Fedora build without vendoring; test via
copr_build on PRs before the first release.
