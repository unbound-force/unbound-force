# Feature Specification: Publish FullSend OpenCode Sandbox Image

**Feature Branch**: `speckit/037-publish-fullsend-opencode`
**Created**: 2026-09-10
**Status**: In Progress
**Input**: User description: "https://github.com/unbound-force/unbound-force/issues/511"

## Clarifications

### Session 2026-09-10

- Q: How should published-image vulnerability scans affect signing? → A:
  Scanner execution errors block signing and attestation; vulnerability
  findings are reported but do not block signing.
- Q: Which publication policy should the specification require? → A: Use
  FullSend's tag convention: main pushes publish latest/dev/SHA, version tags
  publish semver/minor/dev/SHA, manual dispatch publishes dev/SHA, and pull
  requests never publish.
- Q: When should the external FullSend documentation be completed relative to
  this repository's image PR? → A: Track it after the image merge; keep issue
  #511 open until the coordinated FullSend documentation PR lands.
- Q: What evidence should satisfy the FullSend harness usability requirement?
  → A: Verify anonymous digest pulls on both architectures and run a minimal
  live FullSend harness using the digest.
- Q: Which refs may manual dispatch publish from? → A: Only trusted `main`
  and version-tag refs publish; dispatches from other branches do not publish.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Run OpenCode in a Known Sandbox (Priority: P1)

As a FullSend operator, I want a published sandbox image containing
OpenCode and the uf CLI so that a harness can start with a known runtime
without installing tools during agent execution.

**Why this priority**: A deterministic runtime is the primary value of the
feature and is required before the image can be used by a harness.

**Independent Test**: Build the image for each supported Linux architecture,
start it as the inherited non-root sandbox user, and verify the required
commands and versions are available while the excluded services are absent.

**Acceptance Scenarios**:

1. **Given** a supported architecture, **when** the image is built, **then**
   the image contains the pinned OpenCode version and the matching uf release.
2. **Given** a started image, **when** the runtime identity is inspected,
   **then** it runs as the base image's `sandbox` user at UID 998 and not as
   root.
3. **Given** a started image, **when** installed commands are inspected,
   **then** Dewey and Ollama are absent.

---

### User Story 2 - Consume an Auditable Image (Priority: P1)

As a FullSend operator, I want the image published for both supported
architectures with immutable references and supply-chain evidence so that a
harness can consume and verify the exact runtime it executes.

**Why this priority**: An unverified or architecture-specific artifact does
not provide the deterministic runtime promised by the feature.

**Independent Test**: Publish the image, inspect its manifest, pull it by
digest for both architectures, and verify its signature and attestations.

**Acceptance Scenarios**:

1. **Given** a published image, **when** its immutable reference is used,
   **then** both amd64 and arm64 variants are available.
2. **Given** an immutable image reference, **when** its provenance, SBOM, and
   signature are verified, **then** each verification succeeds.
3. **Given** a FullSend harness configured with the immutable image reference,
   **when** the harness starts, **then** it can pull and use the image.

---

### User Story 3 - Maintain Pinned Runtime Inputs (Priority: P2)

As a maintainer, I want pinned OpenCode and uf versions to be discoverable by
dependency-update automation so that upgrades are reviewable instead of being
silently introduced by rebuilds.

**Why this priority**: Reviewable updates reduce version drift while keeping
the runtime reproducible.

**Independent Test**: Validate the dependency-update configuration and confirm
that both version declarations are discoverable with their intended release
sources.

**Acceptance Scenarios**:

1. **Given** a new OpenCode or uf release, **when** dependency automation
   evaluates the image definition, **then** it identifies the corresponding
   pinned version for review.
2. **Given** an uf version update, **when** the update is prepared, **then**
   the required amd64 and arm64 checksums are reviewed before the image can
   build successfully.

### Edge Cases

- A build for an unsupported architecture MUST fail rather than install an
  unverified artifact.
- A release tag that does not match the pinned uf version MUST fail before
  publication.
- A stale architecture checksum MUST fail image construction.
- A pull request MUST validate the image without publishing or granting
  registry write access.
- A failed published-image validation or scanner execution MUST prevent
  signing and attestation; vulnerability findings remain report-only.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The published image MUST be named
  `ghcr.io/unbound-force/fullsend-opencode`.
- **FR-002**: The image MUST extend the FullSend sandbox image through an
  immutable parent-image digest.
- **FR-003**: The image MUST contain a pinned OpenCode release.
- **FR-004**: The image MUST contain a pinned uf release for amd64 and arm64,
  with an independently recorded SHA256 checksum for each architecture.
- **FR-005**: The image MUST run as the inherited FullSend `sandbox` user at
  UID 998 and MUST NOT run as root at runtime.
- **FR-006**: The image MUST omit Dewey and Ollama.
- **FR-007**: The publication process MUST produce amd64 and arm64 image
  variants and an immutable manifest digest.
- **FR-008**: The publication process MUST provide keyless signing, machine-
  readable provenance, and an SPDX SBOM for the immutable digest.
- **FR-009**: The publication process MUST verify the signature, provenance,
  and SBOM before considering the publication successful.
- **FR-010**: Pull requests MUST build and validate without publishing and
  without registry write permission.
- **FR-011**: The image MUST be usable as the `image` value in a FullSend
  harness when referenced by digest; acceptance MUST include anonymous pulls
  on both architectures and a minimal live harness run.
- **FR-012**: Dependency-update configuration MUST discover the OpenCode and
  uf version declarations; uf checksum updates MUST remain reviewable and
  non-automerged.
- **FR-013**: The upstream FullSend documentation MUST receive an OpenCode
  runtime entry and a security-matrix entry describing the temporary hooks
  gap. This external deliverable is tracked after the image PR merges and
  does not block merging the image implementation itself.
- **FR-014**: The publication process MUST fail when the vulnerability scanner
  cannot complete or produce a result, while recording detected vulnerability
  findings without using them as a new repository-wide merge gate.
- **FR-015**: Publication triggers and tags MUST follow the FullSend convention:
  main pushes publish `latest`, `dev`, and a commit-SHA tag; version tags
  publish semver, minor, `dev`, and a commit-SHA tags; manual dispatch from
  trusted `main` or version-tag refs publishes `dev` and a commit-SHA tag;
  manual dispatches from other branches and pull requests MUST NOT publish.

### Scope and Deferred Work

- The OpenCodeRuntime implementation is out of scope.
- Write-capable security hooks are out of scope for the initial image and are
  deferred to issue #515.
- The full uf methodology remains in the repository's `.opencode/` content.
- The FullSend documentation deliverable is an external repository change;
  this repository tracks it through issue #511.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of supported architecture builds contain the pinned
  OpenCode and uf versions and pass runtime validation.
- **SC-002**: 100% of published manifest digests expose both supported Linux
  architectures.
- **SC-003**: 100% of successful publication runs produce verifiable signature,
  provenance, and SBOM evidence for the published digest.
- **SC-004**: 100% of pull-request validation runs complete without registry
  write permission and without publishing an image.
- **SC-005**: A FullSend harness can start successfully from the published
  digest reference on both supported architectures.

## Assumptions and Dependencies

- The FullSend sandbox base remains Debian/Ubuntu-based and defines the
  `sandbox` runtime user at UID 998.
- FullSend ADR 0036 is the external authority for digest pinning, signing,
  and BYOA image usage.
- Issue #515 owns the OpenCode hooks implementation and its later integration
  into this image.
- The upstream FullSend repository owns the runtime documentation and security
  matrix referenced by this feature.
