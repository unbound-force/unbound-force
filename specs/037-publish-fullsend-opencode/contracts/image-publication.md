# Image Publication Contract

## Image Reference

Consumers MUST use:

```text
ghcr.io/unbound-force/fullsend-opencode@sha256:<manifest-digest>
```

Tags MAY be used for discovery and development, but MUST NOT be the acceptance
reference for a production harness.

## Event-to-Tag Contract

| Event | Published tags | PR publication |
|---|---|---|
| Pull request | None | MUST NOT publish |
| Push to `main` | `latest`, `dev`, commit SHA | N/A |
| Version tag | semver, major.minor, `dev`, commit SHA | N/A |
| Manual dispatch from trusted `main` or version-tag ref | `dev`, commit SHA | N/A |
| Manual dispatch from other branch | None | MUST NOT publish |

## Required Manifest

The digest MUST resolve to a manifest containing:

- `linux/amd64`
- `linux/arm64`

## Required Evidence

The same manifest digest MUST have verifiable:

- Keyless signature from the image workflow.
- SLSA provenance.
- SPDX SBOM.

The vulnerability scanner MUST complete and produce a result before signing
and attestation. Vulnerability findings are reported through SARIF and do not
create a new repository-wide severity gate in this feature.

## Consumer Acceptance

The acceptance procedure MUST:

1. Pull the digest anonymously for both supported platforms.
2. Run a minimal FullSend harness using that digest.
3. Record the digest, workflow run, platform, and verification results.
