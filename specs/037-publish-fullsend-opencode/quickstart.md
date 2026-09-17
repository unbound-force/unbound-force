# Quickstart: Validate the FullSend OpenCode Image

## Local Build

Prerequisites: Docker with Buildx; QEMU registration is required for the
arm64 build when running on an amd64 host.

Build and load each supported platform image:

```bash
docker buildx build --platform linux/amd64 --load \
  --tag fullsend-opencode:local-amd64 \
  --file images/fullsend-opencode/Containerfile \
  images/fullsend-opencode

docker buildx build --platform linux/arm64 --load \
  --tag fullsend-opencode:local-arm64 \
  --file images/fullsend-opencode/Containerfile \
  images/fullsend-opencode
```

For each image, verify UID 998, exact declared versions, and excluded commands:

```bash
EXPECTED_OPENCODE=$(sed -n 's/^ARG OPENCODE_VERSION=//p' images/fullsend-opencode/Containerfile)
EXPECTED_UF=$(sed -n 's/^ARG UF_VERSION=//p' images/fullsend-opencode/Containerfile)

for arch in amd64 arm64; do
  docker run --rm --platform "linux/${arch}" --entrypoint '' \
    -e EXPECTED_OPENCODE -e EXPECTED_UF \
    "fullsend-opencode:local-${arch}" sh -ceu '
      test "$(id -u)" = "998"
      test "$(opencode --version)" = "$EXPECTED_OPENCODE"
      uf --version | grep -F "unbound-force version $EXPECTED_UF "
      ! command -v dewey
      ! command -v ollama
    '
done
```

## Static Validation

```bash
actionlint .github/workflows/fullsend-opencode-image.yml
npx --yes --package renovate renovate-config-validator renovate.json
git diff --check
```

## Published Digest Validation

After a non-PR publication, set `DIGEST` to the workflow output, including its
`sha256:` prefix:

```bash
IMAGE=ghcr.io/unbound-force/fullsend-opencode
DIGEST=sha256:<64-hex-digest>
IMAGE_REF="${IMAGE}@${DIGEST}"

docker buildx imagetools inspect \
  "$IMAGE_REF"

docker logout ghcr.io
docker pull --platform linux/amd64 \
  "$IMAGE_REF"
docker pull --platform linux/arm64 \
  "$IMAGE_REF"
```

If a publication-stage job fails after the candidate image is pushed, rerun
the workflow from the same trusted ref after confirming release assets are
available. The rerun creates a new run-scoped candidate and repeats validation
before promotion; stale candidate tags can be removed by a registry maintainer
after the failure is investigated. Version-tag runs wait up to 30 minutes for
GoReleaser assets before failing.

Verify the signature and attestations with the repository's documented
certificate identity and the workflow's OIDC issuer:

```bash
IMAGE=ghcr.io/unbound-force/fullsend-opencode
DIGEST=sha256:<64-hex-digest>
IMAGE_REF="${IMAGE}@${DIGEST}"
IDENTITY='https://github.com/unbound-force/unbound-force/.github/workflows/fullsend-opencode-image.yml@refs/(heads/main|tags/v.*)'
ISSUER=https://token.actions.githubusercontent.com

cosign verify --certificate-identity-regexp="$IDENTITY" \
  --certificate-oidc-issuer="$ISSUER" "$IMAGE_REF"
cosign verify-attestation --type https://slsa.dev/provenance/v1 \
  --certificate-identity-regexp="$IDENTITY" \
  --certificate-oidc-issuer="$ISSUER" "$IMAGE_REF"
cosign verify-attestation --type https://spdx.dev/Document \
  --certificate-identity-regexp="$IDENTITY" \
  --certificate-oidc-issuer="$ISSUER" "$IMAGE_REF"
```

## FullSend Harness Validation

Configure a minimal harness with:

```yaml
image: ghcr.io/unbound-force/fullsend-opencode@sha256:<64-hex-digest>
```

Run the harness once for each supported platform and record the digest, image
platform, workflow run, and successful startup evidence.
