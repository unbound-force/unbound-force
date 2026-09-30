# Research: FullSend OpenCode Sandbox Image

**Date**: 2026-09-10

## Decision: Use the FullSend sandbox image hierarchy

**Decision**: Publish an organization-owned image that extends
`ghcr.io/fullsend-ai/fullsend-sandbox` and use the inherited `sandbox` user
at UID 998.

**Rationale**: FullSend documents one image per directory under `images/`,
supports amd64 and arm64, and uses the sandbox image as the common runtime
base. The issue decision comment records that forcing UID 1000 would conflict
with the base image's ownership model.

**Alternatives considered**: Force UID 1000, use a separate base image, or
install tools at harness startup. These weaken ownership compatibility,
duplicate FullSend isolation behavior, or reduce reproducibility.

**Evidence**: FullSend `images/README.md:9-22`; issue #511 comments
`5601329679` and `5603452429`.

## Decision: Pin and verify runtime inputs

**Decision**: Pin the parent image by manifest digest, pin OpenCode and uf
versions, and verify architecture-specific uf release tarballs with SHA256.
Reject unsupported architectures and validate installed versions at build or
runtime validation time.

**Rationale**: FullSend's image convention uses immutable parent digests and
per-architecture checksums for downloaded tools. The Debian-based base image
means the official uf tarball is appropriate; the RPM wording in the issue is
superseded by the issue's later implementation notes.

**Alternatives considered**: Mutable parent tags, RPM installation, or
unverified downloads. These weaken reproducibility or conflict with the base
image distribution.

**Evidence**: FullSend `images/sandbox/Containerfile:33,49-78`; issue #511
comment `5603582924`.

## Decision: Separate pull-request validation from publication

**Decision**: Pull requests build and validate both architectures without
publishing or registry write permission. Main, version-tag, and manual events
publish according to FullSend's tag convention and expose an immutable digest.

**Rationale**: Separate jobs allow least-privilege PR validation. FullSend's
documented convention is latest/dev/SHA for main, semver/minor/dev/SHA for
version tags, dev/SHA for manual dispatch, and no publication for PRs.

**Alternatives considered**: One job with package-write permission for every
event, or tag-only publication. The former violates least privilege; the
latter omits the documented development workflow.

**Evidence**: FullSend `images/README.md:55-80`; FullSend
`.github/workflows/sandbox-images.yml:61-88`.

## Decision: Gate evidence on scanner execution, not vulnerability severity

**Decision**: A scanner execution or result-generation failure blocks signing
and attestation. Vulnerability findings are uploaded as SARIF and remain
report-only for this feature.

**Rationale**: The clarification preserves existing repository severity gates
and avoids changing governance thresholds. It still prevents an unavailable
or malformed scan from being treated as successful evidence.

**Alternatives considered**: Block signing on HIGH/CRITICAL findings, or make
the scan fully non-blocking. The first changes a protected gate; the second
would allow missing evidence.

## Decision: Attach evidence to the immutable manifest digest

**Decision**: Generate SLSA provenance and SPDX SBOM attestations, sign the
manifest digest with keyless cosign, and verify all evidence against the same
digest after published-image validation.

**Rationale**: Tags are mutable; the digest is the identity consumed by the
harness and the only stable subject for signature, provenance, and SBOM
verification.

**Evidence**: FullSend ADR 0036, `Implementation Details` and `Mitigations`;
FullSend `images/README.md:129-156`.

## Decision: Treat Renovate uf checksums as a reviewed update unit

**Decision**: Renovate discovers OpenCode and uf version declarations. uf
updates are non-automerged and include instructions to update both
architecture checksums from the matching signed release asset. The build
fails closed when checksums are stale.

**Rationale**: Renovate custom managers can discover the version but cannot
infer release-specific architecture checksums without trusted update code.
The neighboring FullSend configuration documents the same manual checksum
pattern for image tools.

**Alternatives considered**: Hosted Renovate post-upgrade scripts or silently
accepting version-only updates. Scripts add trusted execution complexity;
version-only updates create guaranteed build failures without a clear update
contract.

**Evidence**: FullSend `renovate.json:108-116,160-178`; issue #511
implementation notes.

## Decision: Demonstrate harness usability after publication

**Decision**: Acceptance evidence includes anonymous digest pulls for amd64
and arm64 plus a minimal live FullSend harness run using the digest. The
external FullSend runtime documentation follows image merge and keeps issue
#511 open until the coordinated documentation change lands.

**Rationale**: Local PR builds prove construction but cannot prove public
registry access or the actual published manifest. The harness test proves the
consumer contract.

**Evidence**: Issue #511 acceptance criteria and clarification session;
FullSend `docs/guides/user/bring-your-own-agent.md:69-97`.
