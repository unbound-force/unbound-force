# Data Model: FullSend OpenCode Image

This feature publishes and verifies artifacts; it does not introduce
application persistence or a database schema.

## Image Artifact

- **Name**: `ghcr.io/unbound-force/fullsend-opencode`.
- **Manifest digest**: immutable `sha256:<64-hex>` identity for the
  multi-architecture manifest.
- **Platforms**: `linux/amd64` and `linux/arm64`.
- **Tags**: event-derived mutable references used for discovery and a
  commit-SHA reference used for traceability.
- **Runtime identity**: FullSend `sandbox`, UID 998.

## Version Pin Set

- **Parent digest**: immutable FullSend sandbox manifest digest.
- **OpenCode version**: exact npm release version.
- **uf version**: exact release version.
- **uf checksums**: one SHA256 value for each supported architecture.

Validation rules:

- Every supported architecture selects its own checksum.
- Unsupported architectures fail before installation.
- The installed OpenCode and uf versions match their declarations.
- A stale checksum prevents image construction.

## Publication Event

- **Pull request**: build and validate only; no registry publication.
- **Main push**: publish `latest`, `dev`, and commit-SHA tags.
- **Version tag**: publish semver, minor, `dev`, and commit-SHA tags.
- **Manual dispatch from trusted `main` or version-tag refs**: publish `dev`
  and commit-SHA tags; other branch refs do not publish.

## Evidence Bundle

All evidence is bound to the manifest digest:

- Keyless cosign signature.
- SLSA provenance attestation.
- SPDX SBOM attestation.
- Vulnerability scan SARIF and scanner-completion result.

## Harness Reference

- **Image**: the published manifest digest, never a mutable tag.
- **Required proof**: anonymous pulls for both platforms and a minimal live
  FullSend harness run.
