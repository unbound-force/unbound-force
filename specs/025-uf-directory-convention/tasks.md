# Tasks: Unified .uf/ Directory Convention

**Input**: Design documents from `/specs/025-uf-directory-convention/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Verify Baseline)

**Purpose**: Confirm all existing tests pass before any changes

- [x] T001 Run `make check` to verify baseline — all tests, vet, and lint must pass before any modifications

**Checkpoint**: Baseline green — proceed with changes

---

## Phase 2: Scaffold Engine + Asset Directory Rename (US1 + US3)

**Purpose**: The scaffold engine determines where files are deployed. It must be updated first because all other changes depend on the asset directory structure being correct.

**Goal**: `uf init` creates `.uf/` with subdirectories and deploys convention packs to `.opencode/uf/packs/`

### Asset Directory Rename

- [x] T002 [US3] Run `git mv internal/scaffold/assets/opencode/unbound internal/scaffold/assets/opencode/uf` to rename the scaffold asset directory (FR-012)

### Scaffold Engine Source Changes

- [x] T003 [US3] Update `isConventionPack()` in `internal/scaffold/scaffold.go` — change prefix from `opencode/unbound/packs/` to `opencode/uf/packs/` (FR-005, FR-012)
- [x] T004 [P] [US3] Update any comment in `isDivisorAsset()` in `internal/scaffold/scaffold.go` referencing `opencode/unbound/packs/` to `opencode/uf/packs/`
- [x] T005 [US1] Update `workflowConfigContent` constant in `internal/scaffold/scaffold.go` — change header comment from `# .unbound-force/config.yaml` to `# .uf/config.yaml` (FR-001)
- [x] T006 [US1] Update `initSubTools()` in `internal/scaffold/scaffold.go` — change `.unbound-force` to `.uf` for workflow config directory path (FR-001)
- [x] T007 [US1] Update `initSubTools()` in `internal/scaffold/scaffold.go` — change `.dewey` to `.uf/dewey` for Dewey workspace check path (FR-002)
- [x] T008 [US1] Update `initSubTools()` in `internal/scaffold/scaffold.go` — change `.hive` to `.uf/replicator` for Replicator workspace check path, rename variable from `hiveDir` to `replicatorDir` (FR-003)
- [x] T009 [US1] Update `initSubTools()` display names in `internal/scaffold/scaffold.go` — change `.unbound-force/config.yaml` to `.uf/config.yaml`, `.dewey/` to `.uf/dewey/`, `.hive/` to `.uf/replicator/` in subToolResult strings
- [x] T010 [US1] Update `generateDeweySources()` in `internal/scaffold/scaffold.go` — change `.dewey` to `.uf/dewey` in the sources.yaml path (FR-002)
- [x] T011 [P] [US1] Update `configureOpencodeJSON()` in `internal/scaffold/scaffold.go` — verify no path changes needed (research R10 confirms `--vault .` is correct), update any comments referencing old paths

### Scaffold Test Updates

- [x] T012 [US1] [US3] Update `expectedAssetPaths` in `internal/scaffold/scaffold_test.go` — replace all `opencode/unbound/packs/*` entries with `opencode/uf/packs/*`
- [x] T013 [US1] Update all `initSubTools` test assertions in `internal/scaffold/scaffold_test.go` — replace `.unbound-force`, `.dewey`, `.hive` path references with `.uf`, `.uf/dewey`, `.uf/replicator`
- [x] T014 [US1] Update drift detection test baselines in `internal/scaffold/scaffold_test.go` — ensure asset paths reference `opencode/uf/packs/`
- [x] T015 [US1] [US3] Run `go test -race -count=1 ./internal/scaffold/...` to verify scaffold tests pass

**Checkpoint**: Scaffold engine uses new paths. `uf init` would deploy to `.uf/` and `.opencode/uf/packs/`. Scaffold tests pass.

---

## Phase 3: Doctor Checks (US2)

**Purpose**: Doctor validates the environment. Must use new paths to produce correct diagnostics.

**Goal**: `uf doctor` checks `.uf/dewey/`, `.uf/replicator/`, and `.opencode/uf/packs/` instead of old paths

### Doctor Source Changes

- [x] T016 [US2] Update `checkDewey()` in `internal/doctor/checks.go` — change workspace directory from `.dewey` to `.uf/dewey` (FR-007)
- [x] T017 [US2] Update `checkReplicator()` in `internal/doctor/checks.go` — change `.hive` to `.uf/replicator`, update display name from `.hive/` to `.uf/replicator/` (FR-008)
- [x] T018 [US2] Update `checkScaffoldedFiles()` in `internal/doctor/checks.go` — change packs directory from `.opencode/unbound/packs` to `.opencode/uf/packs`, update display name (FR-009)
- [x] T019 [P] [US2] Update any install hint strings in `internal/doctor/checks.go` and `internal/doctor/environ.go` that reference old paths (FR-010)

### Doctor Test Updates

- [x] T020 [US2] Update `internal/doctor/doctor_test.go` — replace all `.dewey/` directory creation with `.uf/dewey/` in test setup
- [x] T021 [US2] Update `internal/doctor/doctor_test.go` — replace all `.hive/` directory creation with `.uf/replicator/` in test setup
- [x] T022 [US2] Update `internal/doctor/doctor_test.go` — replace all `.opencode/unbound/packs/` directory creation with `.opencode/uf/packs/` in test setup
- [x] T023 [US2] Update `internal/doctor/doctor_test.go` — replace all path assertion strings (`.dewey/`, `.hive/`, `.opencode/unbound/packs/`) with new paths
- [x] T024 [US2] Run `go test -race -count=1 ./internal/doctor/...` to verify doctor tests pass

**Checkpoint**: Doctor checks use new paths. All doctor tests pass.

---

## Phase 4: Orchestration Engine (US5)

**Purpose**: Internal orchestration packages reference `.unbound-force/` for workflow state and artifact storage.

**Goal**: All orchestration paths reference `.uf/workflows/` and `.uf/artifacts/`

### Orchestration Source Changes

- [x] T025 [P] [US5] Update `LoadWorkflowConfig()` path and comments in `internal/orchestration/config.go` — replace `.unbound-force/` with `.uf/` (FR-013)
- [x] T026 [P] [US5] Update `Orchestrator` struct field comments in `internal/orchestration/engine.go` — replace `.unbound-force/workflows/` with `.uf/workflows/` and `.unbound-force/artifacts/` with `.uf/artifacts/` (FR-013)
- [x] T027 [P] [US5] Update `WorkflowInstance` comment in `internal/orchestration/models.go` — replace `.unbound-force/` with `.uf/`

### Orchestration Test Updates

- [x] T028 [US5] Update `internal/orchestration/config_test.go` — replace `.unbound-force/` path references with `.uf/`
- [x] T029 [US5] Run `go test -race -count=1 ./internal/orchestration/...` to verify orchestration tests pass

**Checkpoint**: Orchestration engine uses `.uf/` paths. All orchestration tests pass.

---

## Phase 5: Hero CLIs — Muti-Mind, Mx F (US5)

**Purpose**: Hero CLI commands reference old per-hero directories for backlog and metrics data.

**Goal**: Muti-Mind defaults to `.uf/muti-mind/`, Mx F references `.uf/mx-f/`

- [x] T030 [P] [US5] Update `cmd/mutimind/main.go` — change default `--backlog-dir` from `.muti-mind/backlog` to `.uf/muti-mind/backlog` (FR-014)
- [x] T031 [P] [US5] Update `cmd/mutimind/main.go` — change default `--artifacts-dir` from `.muti-mind/artifacts` to `.uf/muti-mind/artifacts` (FR-014)
- [x] T032 [P] [US5] Update `cmd/mutimind/main.go` — change any other `.muti-mind/` references (e.g., config path) to `.uf/muti-mind/` (FR-014)
- [x] T033 [P] [US5] Update `internal/metrics/store.go` — change `.mx-f/` comment reference to `.uf/mx-f/` (FR-015)

**Checkpoint**: Hero CLIs use `.uf/` paths.

---

## Phase 6: Agent/Command Markdown Files (US3 + US5)

**Purpose**: Agent and command Markdown files reference convention pack paths and workflow/artifact paths. Both scaffold asset copies (canonical) and live deployed copies must be updated.

**Goal**: All agent/command files reference `.opencode/uf/packs/`, `.uf/workflows/`, `.uf/artifacts/`, `.uf/muti-mind/`, `.uf/mx-f/`

### Scaffold Asset — Agent Files (canonical sources)

- [x] T034 [P] [US3] Update `internal/scaffold/assets/opencode/agents/cobalt-crush-dev.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/` and `.unbound-force/` with `.uf/`
- [x] T035 [P] [US3] Update `internal/scaffold/assets/opencode/agents/divisor-adversary.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/`
- [x] T036 [P] [US3] Update `internal/scaffold/assets/opencode/agents/divisor-architect.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/`
- [x] T037 [P] [US3] Update `internal/scaffold/assets/opencode/agents/divisor-guard.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/`
- [x] T038 [P] [US3] Update `internal/scaffold/assets/opencode/agents/divisor-sre.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/`
- [x] T039 [P] [US3] Update `internal/scaffold/assets/opencode/agents/divisor-testing.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/`
- [x] T040 [P] [US3] Update `internal/scaffold/assets/opencode/agents/divisor-envoy.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/`
- [x] T041 [P] [US3] Update `internal/scaffold/assets/opencode/agents/divisor-herald.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/`
- [x] T042 [P] [US3] Update `internal/scaffold/assets/opencode/agents/divisor-scribe.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/`
- [x] T043 [P] [US5] Update `internal/scaffold/assets/opencode/agents/mx-f-coach.md` — replace `.mx-f/` with `.uf/mx-f/`

### Scaffold Asset — Command Files (canonical sources)

- [x] T044 [P] [US3] [US5] Update `internal/scaffold/assets/opencode/command/cobalt-crush.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/` and `.unbound-force/` with `.uf/`
- [x] T045 [P] [US3] Update `internal/scaffold/assets/opencode/command/review-council.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/`
- [x] T046 [P] [US3] Update `internal/scaffold/assets/opencode/command/uf-init.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/`

### Live Agent Files (deployed copies — must match scaffold assets)

- [x] T047 [P] [US3] Sync `.opencode/agents/cobalt-crush-dev.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/` and `.unbound-force/` with `.uf/`
- [x] T048 [P] [US3] Sync `.opencode/agents/divisor-adversary.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/`
- [x] T049 [P] [US3] Sync `.opencode/agents/divisor-architect.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/`
- [x] T050 [P] [US3] Sync `.opencode/agents/divisor-guard.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/`
- [x] T051 [P] [US3] Sync `.opencode/agents/divisor-sre.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/`
- [x] T052 [P] [US3] Sync `.opencode/agents/divisor-testing.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/`
- [x] T053 [P] [US3] Sync `.opencode/agents/divisor-envoy.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/`
- [x] T054 [P] [US3] Sync `.opencode/agents/divisor-herald.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/`
- [x] T055 [P] [US3] Sync `.opencode/agents/divisor-scribe.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/`
- [x] T056 [P] [US5] Sync `.opencode/agents/muti-mind-po.md` — replace `.muti-mind/` with `.uf/muti-mind/`
- [x] T057 [P] [US5] Sync `.opencode/agents/mx-f-coach.md` — replace `.mx-f/` with `.uf/mx-f/`

### Live Command Files (deployed copies)

- [x] T058 [P] [US3] [US5] Sync `.opencode/command/cobalt-crush.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/` and `.unbound-force/` with `.uf/`
- [x] T059 [P] [US3] Sync `.opencode/command/review-council.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/`
- [x] T060 [P] [US3] Sync `.opencode/command/uf-init.md` — replace `opencode/unbound/packs/` with `opencode/uf/packs/`
- [x] T061 [P] [US5] Sync `.opencode/command/workflow-start.md` — replace `.unbound-force/` with `.uf/`
- [x] T062 [P] [US5] Sync `.opencode/command/workflow-status.md` — replace `.unbound-force/` with `.uf/`
- [x] T063 [P] [US5] Sync `.opencode/command/workflow-list.md` — replace `.unbound-force/` with `.uf/`
- [x] T064 [P] [US5] Sync `.opencode/command/workflow-advance.md` — replace `.unbound-force/` with `.uf/`
- [x] T065 [P] [US5] Sync `.opencode/command/muti-mind.init.md` — replace `.muti-mind/` with `.uf/muti-mind/`
- [x] T066 [P] [US5] Sync `.opencode/command/muti-mind.backlog-add.md` — replace `.muti-mind/` with `.uf/muti-mind/`
- [x] T067 [P] [US5] Sync `.opencode/command/muti-mind.backlog-list.md` — replace `.muti-mind/` with `.uf/muti-mind/`

### Scaffold Drift Verification

- [x] T068 [US3] Run `go test -race -count=1 -run TestScaffold ./internal/scaffold/...` to verify scaffold drift detection passes (all scaffold asset copies match canonical sources)

**Checkpoint**: All agent/command files reference new paths. Scaffold drift tests pass.

---

## Phase 7: Config, Scripts, Schemas (Cross-Cutting)

**Purpose**: Update remaining config files, scripts, and schema descriptions

### Config Files

- [x] T069 [P] [US1] Update `.gitignore` — replace `.unbound-force/` and `.dewey/` entries with single `.uf/` entry (FR-018)
- [x] T070 [P] [US1] Verify `opencode.json` Dewey serve command — confirm `--vault .` is correct and no changes needed (research R10)

### Scripts

- [x] T071 [US5] Update `scripts/validate-hero-contract.sh` — replace all `.unbound-force/hero.json` references with `.uf/hero.json` (FR-016)

### Schemas

- [x] T072 [P] [US5] Update `schemas/hero-manifest/v1.0.0.schema.json` — replace `.unbound-force/hero.json` in description text with `.uf/hero.json` (FR-017)
- [x] T073 [P] [US5] Update `schemas/acceptance-decision/samples/sample-acceptance-decision.json` — replace `.unbound-force/artifacts/` path reference with `.uf/artifacts/`

### Schema Test Paths

- [x] T074 [P] [US3] Update `internal/schemas/packvalidator_test.go` — replace `.opencode/unbound/packs/` comment reference with `.opencode/uf/packs/`

**Checkpoint**: All config, scripts, and schemas reference new paths.

---

## Phase 8: Documentation + Regression Test (Polish)

**Purpose**: Update living documentation and add regression test to prevent old path reintroduction

### Documentation

- [x] T075 [US1] [US2] [US3] [US5] Update `AGENTS.md` Project Structure section — replace `.dewey/`, `.hive/`, `.unbound-force/`, `.muti-mind/`, `.mx-f/` with `.uf/` equivalents, replace `.opencode/unbound/packs/` with `.opencode/uf/packs/`
- [x] T076 [US1] [US5] Update `AGENTS.md` pack listings and any prose referencing old directory paths
- [x] T077 Update `AGENTS.md` Recent Changes section — add entry for `025-uf-directory-convention`

### Regression Test

- [x] T078 [US1] [US3] Add `TestScaffoldOutput_NoOldPathReferences` in `internal/scaffold/scaffold_test.go` — grep all scaffold asset content for old path patterns (`.dewey/`, `.hive/`, `.unbound-force/`, `.muti-mind/`, `.mx-f/`, `opencode/unbound/`) and fail if any found (SC-001, SC-002)

### Final Verification

- [x] T079 Run `make check` to verify all tests, vet, and lint pass (SC-005, SC-006)
- [x] T080 Run grep verification from `contracts/path-mapping.md` — confirm zero matches for old path patterns in production code (SC-001, SC-002)

**Checkpoint**: All documentation updated. Regression test prevents reintroduction. Full test suite passes.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — verify baseline first
- **Phase 2 (Scaffold)**: Depends on Phase 1 — BLOCKS all subsequent phases (scaffold determines deployment paths)
- **Phase 3 (Doctor)**: Depends on Phase 2 — doctor checks paths that scaffold creates
- **Phase 4 (Orchestration)**: Depends on Phase 2 — can run in parallel with Phase 3
- **Phase 5 (Hero CLIs)**: Depends on Phase 2 — can run in parallel with Phases 3-4
- **Phase 6 (Markdown)**: Depends on Phase 2 — can run in parallel with Phases 3-5 (different files)
- **Phase 7 (Config/Scripts)**: Depends on Phase 2 — can run in parallel with Phases 3-6
- **Phase 8 (Docs/Regression)**: Depends on ALL previous phases — final verification

### Within-Phase Dependencies

- **Phase 2**: T002 (git mv) must complete before T003-T014. T015 (test run) depends on all T002-T014.
- **Phase 3**: T016-T019 (source changes) before T020-T023 (test updates). T024 depends on all.
- **Phase 6**: Scaffold assets (T034-T046) should be updated before live copies (T047-T067) to maintain canonical-first ordering. T068 depends on all.

### Parallel Opportunities

After Phase 2 completes, Phases 3-7 can proceed in parallel since they touch different files:
- Phase 3: `internal/doctor/`
- Phase 4: `internal/orchestration/`
- Phase 5: `cmd/mutimind/`, `internal/metrics/`
- Phase 6: `.opencode/agents/`, `.opencode/command/`, `internal/scaffold/assets/opencode/agents/`, `internal/scaffold/assets/opencode/command/`
- Phase 7: `.gitignore`, `scripts/`, `schemas/`

Within Phase 6, all [P]-marked tasks can run in parallel (each touches a different file).

---

## Implementation Strategy

### Recommended Approach (Sequential)

1. Phase 1: Verify baseline (1 task)
2. Phase 2: Scaffold engine — the foundation (14 tasks)
3. Phases 3-7 in priority order: Doctor → Orchestration → Hero CLIs → Markdown → Config (55 tasks)
4. Phase 8: Documentation + regression test + final verification (6 tasks)

### Total: 80 tasks across 8 phases

---

## Notes

- All changes are mechanical path renames — no logic changes, no new features
- Historical spec documents under `specs/` and `openspec/changes/archive/` are NOT updated (archival records)
- The `opencode.json` Dewey serve command (`--vault .`) does NOT change — Dewey handles the workspace subdirectory internally
- External dependencies (Dewey #33, Replicator #9) are hard blockers for integration testing but NOT for code changes in this repo
- The regression test (T078) is the safety net — it prevents old path references from being reintroduced in future changes

<!-- spec-review: passed -->
