# Implementation Plan: Publish FullSend OpenCode Sandbox Image

**Branch**: `speckit/037-publish-fullsend-opencode` | **Date**: 2026-09-10 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/037-publish-fullsend-opencode/spec.md`

## Summary

Publish a digest-pinned, multi-architecture FullSend OpenCode sandbox image
containing pinned OpenCode and uf runtimes. The workflow will separate
read-only pull-request validation from publication, validate the published
manifest digest on both architectures, attach and verify supply-chain
evidence, and expose a digest suitable for a FullSend harness. Renovate will
discover version pins while keeping uf checksum updates human-reviewed.

## Technical Context

**Language/Version**: N/A; Containerfile, YAML, and shell workflow definitions
**Primary Dependencies**: Docker Buildx, QEMU, GHCR, cosign, GitHub attestations, SPDX SBOM, Trivy, Renovate
**Storage**: N/A; OCI registry artifacts and CI evidence
**Testing**: Buildx image builds, container assertions, actionlint, Renovate config validation, manifest inspection, digest pulls, harness run
**Target Platform**: Linux containers on `amd64` and `arm64`
**Project Type**: CI-published container artifact
**Performance Goals**: Complete per-architecture validation within the CI workflow timeout; no runtime latency target
**Constraints**: Pull requests MUST have no registry write permission; published consumers MUST use immutable digests; scanner execution failures MUST block signing
**Scale/Scope**: One public image with two Linux platform variants and one digest-bound evidence bundle

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

* **I. Autonomous Collaboration**: PASS. Build outputs, digests, attestations,
  and harness evidence are artifact-based; no synchronous hero dependency is
  introduced.
* **II. Composability First**: PASS. The image is independently pullable and
  usable by FullSend through its image contract.
* **III. Observable Quality**: PASS. Provenance, SBOM, signature, scan output,
  and recorded digest provide machine-readable evidence.
* **IV. Testability**: PASS. Both architectures have isolated build/runtime
  checks; publication and harness acceptance are explicit integration tests.
* **V. Security by Default**: PASS with documented tradeoff. Parent and
  actions are digest-pinned, uf archives use architecture-specific SHA256,
  PRs are least-privileged, and scanner execution failures block signing.
  Vulnerability findings remain report-only to avoid changing protected
  repository severity gates.
* **Development workflow**: PASS for this branch. This plan follows the
  retrospective replacement-PR workflow; implementation will not begin until
  the spec, plan, and tasks are committed.
* **Cross-repo documentation**: REQUIRED follow-up. The external FullSend
  documentation is tracked by issue #511; a website issue is not required for
  this CI-only change under the constitution exemption.

**Post-design recheck**: PASS. The research, artifact model, publication
contract, and quickstart preserve least privilege, immutable evidence,
isolated architecture validation, and the report-only vulnerability policy
without changing protected repository gates.

## Project Structure

### Documentation (this feature)

```text
specs/037-publish-fullsend-opencode/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/image-publication.md
└── tasks.md
```

### Source Code (repository root)

```text
images/
└── fullsend-opencode/
    ├── Containerfile
    └── .dockerignore
.github/workflows/
└── fullsend-opencode-image.yml
renovate.json
AGENTS.md
CHANGELOG.md
```

**Structure Decision**: Use the repository's existing image and workflow
locations. Keep all image-specific source and CI changes under
`images/fullsend-opencode/` and `.github/workflows/`; keep planning artifacts
under the feature directory. No Go package or application data model changes
are required.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|---|---|---|
| None | N/A | The feature adds one CI-published image artifact within the existing repository structure. |
