# Tasks: Publish FullSend OpenCode Sandbox Image

**Input**: Design documents from
`/specs/037-publish-fullsend-opencode/`
**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`,
`contracts/`, `quickstart.md`

**Organization**: Tasks are grouped by user story. Tests are included because
the feature specification requires build, integration, and harness validation.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish the image, workflow, and maintenance file locations.

- [x] T001 Create the image directory and Containerfile placeholder at `images/fullsend-opencode/Containerfile`.
- [x] T002 Create the image workflow at `.github/workflows/fullsend-opencode-image.yml` with the repository SPDX header, path filters, concurrency, and workflow-level least-privilege permissions.
- [x] T003 Create the Renovate configuration at `renovate.json` with custom-manager scope limited to the OpenCode image Containerfile.
- [x] T004 [P] Update the project structure and workflow inventory in `AGENTS.md`.
- [x] T005 [P] Add the user-facing image change entry to `CHANGELOG.md` under `Unreleased/Added`, referencing issue #511.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Define immutable inputs, artifact identity, publication policy, and
job dependency boundaries before story-specific validation is added.

- [x] T006 Pin the FullSend parent manifest digest and declare OpenCode, uf, target-architecture, and per-architecture checksum arguments in `images/fullsend-opencode/Containerfile`.
- [x] T007 Define the image name, event-to-tag matrix, digest output, and release-tag-to-uf-version validation in `.github/workflows/fullsend-opencode-image.yml`.
- [x] T008 Define separate pull-request, publication, published-validation, scanner, and supply-chain jobs with explicit `needs` relationships and least-privilege permissions in `.github/workflows/fullsend-opencode-image.yml`.
- [x] T009 Add the Renovate OpenCode and uf custom managers, non-automerge rule, and manual checksum-update note in `renovate.json`.
- [x] T010 [P] Write the final image publication contract and digest-consumer reference in `specs/037-publish-fullsend-opencode/contracts/image-publication.md`.

**Checkpoint**: Immutable inputs, publication policy, and job boundaries are
defined; user-story implementation can begin.

---

## Phase 3: User Story 1 - Run OpenCode in a Known Sandbox (Priority: P1) 🎯 MVP

**Goal**: Build a reproducible non-root image containing the pinned OpenCode
and uf runtimes while excluding Dewey and Ollama.

**Independent Test**: Build and load amd64 and arm64 images, then assert UID
998, exact command versions, and absence of excluded commands.

### Implementation

- [x] T011 [US1] Install the exact `opencode-ai` version as root during image construction and assert its version in `images/fullsend-opencode/Containerfile`.
- [x] T012 [US1] Download the architecture-specific uf release, select the matching checksum, reject unsupported architectures, verify with `sha256sum -c`, install `uf`, and remove temporary artifacts in `images/fullsend-opencode/Containerfile`.
- [x] T013 [US1] Restore the inherited `sandbox` runtime user at UID 998 in `images/fullsend-opencode/Containerfile`.
- [x] T014 [US1] Add fail-fast amd64 and arm64 pull-request validation with exact OpenCode/uf version assertions, UID 998 assertion, and Dewey/Ollama absence checks in `.github/workflows/fullsend-opencode-image.yml`.
- [x] T015 [US1] Execute the local amd64 and emulated arm64 build-and-run checks from `specs/037-publish-fullsend-opencode/quickstart.md` and record any required fixes in the implementation branch.

**Checkpoint**: Both supported architecture images build and independently
pass runtime validation without publication.

---

## Phase 4: User Story 2 - Consume an Auditable Image (Priority: P1)

**Goal**: Publish and validate an immutable multi-architecture image with
machine-readable supply-chain evidence and a digest-based consumer contract.

**Independent Test**: Publish a non-PR image, validate its digest on both
architectures, scan it, verify its signature and attestations, and run a
minimal digest-pinned harness.

### Implementation

- [x] T016 [US2] Implement the non-PR multi-architecture Buildx publication and manifest-digest output in `.github/workflows/fullsend-opencode-image.yml`.
- [x] T017 [US2] Pull the published digest with explicit amd64 and arm64 platforms and run the same runtime assertions in `.github/workflows/fullsend-opencode-image.yml`.
- [x] T018 [US2] Add published-digest Trivy scanning with SARIF upload, scanner-execution failure handling, and report-only vulnerability findings in `.github/workflows/fullsend-opencode-image.yml`.
- [x] T019 [US2] Generate SLSA provenance and SPDX SBOM attestations bound to the manifest digest in `.github/workflows/fullsend-opencode-image.yml`.
- [x] T020 [US2] Sign the immutable digest with keyless cosign and verify the signature, provenance, and SBOM using the repository/workflow certificate identity in `.github/workflows/fullsend-opencode-image.yml`.
- [ ] T021 [US2] Run manifest inspection and anonymous digest pulls for both architectures after publication, recording the digest and platform evidence in `specs/037-publish-fullsend-opencode/quickstart.md` and issue #511.
- [ ] T022 [US2] Configure a minimal FullSend harness with the published digest and execute a live harness run for each supported platform, recording startup and selected-digest evidence in `specs/037-publish-fullsend-opencode/quickstart.md` and issue #511.

**Checkpoint**: The published digest is multi-architecture, publicly pullable,
verified, and usable by a digest-pinned FullSend harness.

---

## Phase 5: User Story 3 - Maintain Pinned Runtime Inputs (Priority: P2)

**Goal**: Keep runtime version upgrades reviewable and prevent incomplete uf
checksum updates from being silently accepted.

**Independent Test**: Validate Renovate configuration, inspect both discovered
version managers, and confirm uf updates are explicitly non-automerged with
manual checksum instructions.

### Implementation

- [x] T023 [US3] Validate that the Renovate custom manager discovers `OPENCODE_VERSION` from npm and `UF_VERSION` from GitHub releases in `renovate.json`.
- [x] T024 [US3] Document and validate the reviewed update unit consisting of `UF_VERSION`, `UF_SHA256_AMD64`, and `UF_SHA256_ARM64` in `renovate.json` and `images/fullsend-opencode/Containerfile`.
- [x] T025 [US3] Confirm the image build fails closed for stale or swapped uf checksums and unsupported `TARGETARCH` values using the image validation workflow.

**Checkpoint**: Runtime version updates are discoverable, reviewable, and
cannot silently bypass architecture-specific integrity checks.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Validate the complete change, synchronize documentation, and
close the implementation loop without implementing deferred hooks.

- [x] T026 [P] Run `actionlint` and Renovate configuration validation against `.github/workflows/fullsend-opencode-image.yml` and `renovate.json`.
- [x] T027 [P] Run repository CI-parity checks from `.github/workflows/`, including `go test -race -count=1 ./...`, and record unrelated baseline failures without weakening gates.
- [x] T028 [P] Run the final local image builds and quickstart assertions from `specs/037-publish-fullsend-opencode/quickstart.md`.
- [ ] T029 [P] Prepare the coordinated external FullSend documentation change for `fullsend/docs/runtimes.md`, including the OpenCode runtime row, security matrix, digest usage, and deferred hooks gap; keep it tracked after the image PR merges.
- [x] T030 Run the code review council and record the implementation-review result in `specs/037-publish-fullsend-opencode/tasks.md` before replacing PR #586 with the Speckit PR.
- [x] T031 Record the deferred OpenCode hooks integration and issue #515 relationship in `specs/037-publish-fullsend-opencode/research.md` without adding the plugin to the initial image.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No implementation dependency; establishes target files.
- **Foundational (Phase 2)**: Depends on Setup and blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational; is the MVP image slice.
- **User Story 2 (Phase 4)**: Depends on User Story 1 runtime contents.
- **User Story 3 (Phase 5)**: Depends on the version declarations from Foundational and User Story 1.
- **Polish (Phase 6)**: Depends on all desired user-story checkpoints.

### User Story Dependencies

- **US1**: Independent after Foundational.
- **US2**: Depends on US1 because publication validation consumes the built image contents.
- **US3**: Depends on US1 version declarations but can validate in parallel with US2 after those declarations exist.

### Parallel Opportunities

- T004 and T005 can run in parallel after the initial file layout is agreed.
- T010 can run in parallel with T006-T009 because it only changes a spec contract.
- T026, T027, T028, and T029 can run in parallel after implementation checkpoints.
- US2 and US3 can be worked in parallel after US1 is complete, subject to workflow-file reservation.

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Setup and Foundational phases.
2. Implement the Containerfile and pull-request architecture validation.
3. Stop and validate both local platform images independently.

### Incremental Delivery

1. Add US1 for a reproducible local sandbox.
2. Add US2 for publication, evidence, digest validation, and harness proof.
3. Add US3 for long-term version maintenance.
4. Complete Polish tasks and replace PR #586 with the reviewed Speckit PR.

### Deferred Scope

The OpenCode hooks plugin remains owned by issue #515 and is not part of the
initial implementation task set beyond documenting the integration follow-up.

## Execution Notes

- Local amd64 and emulated arm64 image builds and runtime assertions pass.
- Local amd64 image manifest: `sha256:7e59c4ee68b684b70ecbe97bb5eafcba306ec362ab90f7fdf28640f2ad3c7113`.
- Local arm64 image manifest: `sha256:86b0793edf48ad642f863ab6488717ae2013b7b1de4305272f50a4a2c96b6055`.
- Both local runs reported OpenCode `1.18.29`, uf `0.17.0`, UID 998, and no
  Dewey or Ollama commands.
- Stale checksum and unsupported architecture negative tests fail as expected.
- The publication workflow now builds a run-specific candidate tag, waits for
  release assets on version-tag runs, and promotes discovery tags only after
  validation, scanning, signing, and attestation succeed.
- Pull-request CI now runs automated stale-checksum and unsupported-architecture
  negative builds in the `negative-validation` job.
- Untrusted manual dispatches run the read-only validation path without
  publishing; anonymous validation uses an isolated Docker configuration and
  clears registry credentials before pulling.
- Final local council review ran against the implementation. It confirmed the
  post-publication and external documentation tasks remain open; the npm
  registry-integrity tradeoff remains the documented issue #511 decision.
- T030 council result: APPROVE for the pre-push implementation gate on
  2026-09-10 at commit `9979eee`. Findings were resolved without weakening
  quality or governance gates. T021, T022, and T029 remain intentionally
  pending post-publication or external follow-up work.
- `go test -race -count=1 ./...` passes.
- `make check` reaches the existing lint gate but reports six unrelated
  SA5011 staticcheck findings in `internal/doctor/doctor_test.go`; no gate was
  weakened.
- T021 and T022 remain post-publication acceptance tasks.
- T029 remains the post-merge external FullSend documentation follow-up;
  issue #511 intentionally remains open until it lands.
- Implementation commits are local-only by user instruction; no remote push
  or replacement PR has been created.
