---

description: "Task list for externalizing Speckit, OpenSpec, and Gaze initialization"
---
<!-- scaffolded by uf vdev -->

# Tasks: Externalize Tool Initialization

**Input**: Design documents from `/specs/027-externalize-tool-init/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 1: Setup Prerequisites (US4 — uf setup)

**Purpose**: Add `uv` and `specify` CLI installation to `uf setup` so the tools are available for `uf init` delegation.

**Goal**: `uf setup` installs `uv` (Python package manager) and `specify` CLI as new steps, increasing the step count from 12 to 14.

**Independent Test**: Run `uf setup --dry-run` and verify 14 steps are listed with `uv` at step 7 and `Specify CLI` at step 8.

### Implementation for US4

- [x] T001 [US4] Add `installUV()` function to `internal/setup/setup.go` — Homebrew-first pattern with curl fallback (`curl -LsSf https://astral.sh/uv/install.sh | sh`), interactive guard for curl (requires `--yes` or TTY), follows `installOpenCode()` pattern
- [x] T002 [US4] Add `installSpecify()` function to `internal/setup/setup.go` — checks `LookPath("specify")`, dry-run support, gates on `LookPath("uv")` availability, installs via `uv tool install specify-cli`, follows `installOpenSpec()` pattern
- [x] T003 [US4] Update `Run()` step flow in `internal/setup/setup.go` — change all `[N/12]` format strings to `[N/14]`, insert `installUV()` at step 7 and `installSpecify()` at step 8 (gated by `uvResult` success matching Node.js→OpenSpec pattern), renumber steps 7–12 to 9–14

### Tests for US4

- [x] T004 [P] [US4] Add `TestInstallUV_AlreadyInstalled` to `internal/setup/setup_test.go` — verify skip when `uv` is in PATH
- [x] T005 [P] [US4] Add `TestInstallUV_DryRun_Homebrew` to `internal/setup/setup_test.go` — verify dry-run output with Homebrew available
- [x] T006 [P] [US4] Add `TestInstallUV_DryRun_Curl` to `internal/setup/setup_test.go` — verify dry-run output without Homebrew
- [x] T007 [P] [US4] Add `TestInstallUV_Homebrew` to `internal/setup/setup_test.go` — verify `brew install uv` is called
- [x] T008 [P] [US4] Add `TestInstallUV_Curl` to `internal/setup/setup_test.go` — verify curl fallback is called when no Homebrew
- [x] T009 [P] [US4] Add `TestInstallUV_CurlSkipped` to `internal/setup/setup_test.go` — verify skip when no Homebrew, no TTY, no `--yes`
- [x] T010 [P] [US4] Add `TestInstallSpecify_AlreadyInstalled` to `internal/setup/setup_test.go` — verify skip when `specify` is in PATH
- [x] T011 [P] [US4] Add `TestInstallSpecify_DryRun` to `internal/setup/setup_test.go` — verify dry-run output
- [x] T012 [P] [US4] Add `TestInstallSpecify_NoUV` to `internal/setup/setup_test.go` — verify skip when `uv` not in PATH
- [x] T013 [P] [US4] Add `TestInstallSpecify_Success` to `internal/setup/setup_test.go` — verify `uv tool install specify-cli` is called
- [x] T014 [P] [US4] Add `TestInstallSpecify_Failed` to `internal/setup/setup_test.go` — verify failure result when `uv tool install` fails
- [x] T015 [US4] Update existing setup test step count assertions in `internal/setup/setup_test.go` — change all `"/12]"` references to `"/14]"` and verify step numbering

**Checkpoint**: `go test -race -count=1 ./internal/setup/...` passes. `uf setup --dry-run` shows 14 steps.

---

## Phase 2: Embedded Asset Removal (US1, US2)

**Purpose**: Remove Speckit and OpenSpec base files from embedded scaffold assets. This is a prerequisite for the delegation work — the files must be removed before delegations are added, otherwise `uf init` would deploy both embedded files AND call the external CLI.

**Goal**: Reduce embedded asset count from 55 to 42 (remove 12 Speckit files + 1 OpenSpec config).

**Independent Test**: `go test -race -count=1 ./internal/scaffold/...` passes with updated asset count.

### Implementation for Asset Removal

- [x] T016 [US1] Delete the entire `internal/scaffold/assets/specify/` directory (12 files: 6 templates, 5 scripts, 1 config)
- [x] T017 [US2] Delete `internal/scaffold/assets/openspec/config.yaml` (1 file) — keep `openspec/schemas/unbound-force/` (5 files)
- [x] T018 [US1] Remove `"specify/"` from `knownAssetPrefixes` in `internal/scaffold/scaffold.go` (line 251)
- [x] T019 [US1] Remove the `specify/` case from `mapAssetPath()` in `internal/scaffold/scaffold.go` (lines 262–263) — dead code after asset removal

### Test Updates for Asset Removal

- [x] T020 [US1] Remove 12 Speckit entries from `expectedAssetPaths` in `internal/scaffold/scaffold_test.go` (lines 88–102: 6 templates, 5 scripts, 1 config)
- [x] T021 [US2] Remove 1 OpenSpec config entry from `expectedAssetPaths` in `internal/scaffold/scaffold_test.go` (line 151: `"openspec/config.yaml"`)
- [x] T022 [US1] Add 12 Speckit file entries to `knownNonEmbeddedFiles` in `internal/scaffold/scaffold_test.go` — files now created by `specify init` (`.specify/config.yaml`, 6 templates, 5 scripts)
- [x] T023 [US2] Add 1 OpenSpec config entry to `knownNonEmbeddedFiles` in `internal/scaffold/scaffold_test.go` — `"openspec/config.yaml"` now created by `openspec init`
- [x] T024 [US1] Update `TestCanonicalSources_AreEmbedded` standalone config check in `internal/scaffold/scaffold_test.go` (line 984) — remove `.specify/config.yaml` from the standalone check list (it is now in `knownNonEmbeddedFiles`), keep `openspec/config.yaml` check but it will pass via `knownNonEmbeddedFiles`
- [x] T025 [US1] Remove `mapAssetPath` test cases for `specify/` prefix in `internal/scaffold/scaffold_test.go` — update `TestMapAssetPath` and `TestAllAssets_HaveKnownPrefix` to reflect removal

**Checkpoint**: `go test -race -count=1 ./internal/scaffold/...` passes with 42 embedded assets. `internal/scaffold/assets/specify/` directory does not exist. `internal/scaffold/assets/openspec/config.yaml` does not exist.

---

## Phase 3: Tool Delegations (US1, US2, US3)

**Purpose**: Add `specify init`, `openspec init --tools opencode`, and `gaze init` delegations to `initSubTools()` in the scaffold engine.

**Goal**: `uf init` delegates initialization to external CLIs when available, following the established Dewey/Replicator pattern.

**Independent Test**: Run `uf init` in a fresh directory with all tools installed. Verify `.specify/`, `openspec/config.yaml`, and Gaze agent files are created by CLI delegation.

### Implementation for US1 — Speckit Delegation

- [x] T026 [US1] Add Specify CLI delegation block to `initSubTools()` in `internal/scaffold/scaffold.go` — gate on `LookPath("specify")` + `os.Stat(".specify/")`, call `specify init`, report `subToolResult{name: ".specify/"}`, place after Replicator delegation

### Implementation for US2 — OpenSpec Delegation

- [x] T027 [US2] Add OpenSpec CLI delegation block to `initSubTools()` in `internal/scaffold/scaffold.go` — gate on `LookPath("openspec")` + `os.Stat("openspec/config.yaml")` (NOT directory — see R4), call `openspec init --tools opencode`, report `subToolResult{name: "openspec/"}`, place after Specify delegation

### Implementation for US3 — Gaze Delegation

- [x] T028 [US3] Add Gaze CLI delegation block to `initSubTools()` in `internal/scaffold/scaffold.go` — gate on `LookPath("gaze")` + `os.Stat(".opencode/agents/gaze-reporter.md")` (file, not directory — see R6), call `gaze init`, report `subToolResult{name: "gaze"}`, place after OpenSpec delegation

### Tests for Delegations

- [x] T029 [P] [US1] Add `TestInitSubTools_SpecifyInit` to `internal/scaffold/scaffold_test.go` — verify `specify init` is called when binary available and `.specify/` absent, verify `subToolResult` with action `"initialized"`
- [x] T030 [P] [US1] Add `TestInitSubTools_SpecifySkipped` to `internal/scaffold/scaffold_test.go` — verify skip when `.specify/` already exists
- [x] T031 [P] [US1] Add `TestInitSubTools_SpecifyNotInstalled` to `internal/scaffold/scaffold_test.go` — verify skip when `specify` binary not in PATH
- [x] T032 [P] [US1] Add `TestInitSubTools_SpecifyFailed` to `internal/scaffold/scaffold_test.go` — verify `subToolResult` with action `"failed"` when `specify init` returns error
- [x] T033 [P] [US2] Add `TestInitSubTools_OpenSpecInit` to `internal/scaffold/scaffold_test.go` — verify `openspec init --tools opencode` is called when binary available and `openspec/config.yaml` absent
- [x] T034 [P] [US2] Add `TestInitSubTools_OpenSpecSkipped` to `internal/scaffold/scaffold_test.go` — verify skip when `openspec/config.yaml` already exists
- [x] T035 [P] [US2] Add `TestInitSubTools_OpenSpecFailed` to `internal/scaffold/scaffold_test.go` — verify `subToolResult` with action `"failed"` when `openspec init` returns error
- [x] T036 [P] [US3] Add `TestInitSubTools_GazeInit` to `internal/scaffold/scaffold_test.go` — verify `gaze init` is called when binary available and `gaze-reporter.md` absent
- [x] T037 [P] [US3] Add `TestInitSubTools_GazeSkipped` to `internal/scaffold/scaffold_test.go` — verify skip when `gaze-reporter.md` already exists
- [x] T038 [P] [US3] Add `TestInitSubTools_GazeFailed` to `internal/scaffold/scaffold_test.go` — verify `subToolResult` with action `"failed"` when `gaze init` returns error

**Checkpoint**: `go test -race -count=1 ./internal/scaffold/...` passes. All 3 delegations work with injected stubs.

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Documentation updates and final validation.

- [x] T039 [US1] Update the `/uf-init` slash command at `internal/scaffold/assets/opencode/command/uf-init.md` — document that `.specify/` is created by `specify init` (not embedded), `openspec/` base structure by `openspec init`, and Gaze files by `gaze init`; note post-init customization steps if needed
- [x] T040 Update `AGENTS.md` Recent Changes section — add entry for `027-externalize-tool-init` summarizing: removed 13 embedded assets, added 3 tool delegations to `uf init`, added 2 setup steps (12→14), updated `expectedAssetPaths` (55→42)
- [x] T041 Run full test suite verification — `go test -race -count=1 ./...` and `go build ./...` and `golangci-lint run`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (US4 — Setup)**: No dependencies — can start immediately. Modifies only `internal/setup/` files.
- **Phase 2 (Asset Removal)**: No dependencies on Phase 1 — can run in parallel. Modifies only `internal/scaffold/` files.
- **Phase 3 (Delegations)**: Depends on Phase 2 completion — the embedded assets must be removed before adding delegations to avoid deploying files twice. Modifies only `internal/scaffold/` files.
- **Phase 4 (Polish)**: Depends on Phases 1–3 completion.

### User Story Dependencies

- **US4 (Setup)**: Independent — can start immediately. Only touches `internal/setup/`.
- **US1 (Speckit)**: Phases 2 + 3 — asset removal then delegation. Only touches `internal/scaffold/`.
- **US2 (OpenSpec)**: Phases 2 + 3 — asset removal then delegation. Only touches `internal/scaffold/`.
- **US3 (Gaze)**: Phase 3 only — no asset removal needed (Gaze files were never embedded). Only touches `internal/scaffold/`.

### Parallel Opportunities

- **Phase 1 and Phase 2** can run in parallel (different files: `setup.go`/`setup_test.go` vs `scaffold.go`/`scaffold_test.go`)
- **T004–T015** (US4 tests) can all run in parallel — each is an independent test function
- **T016–T017** (asset deletion) can run in parallel — different directories
- **T020–T025** (test updates) should run after T016–T017 but can run in parallel with each other
- **T029–T038** (delegation tests) can all run in parallel — each is an independent test function
- **T026–T028** (delegation implementations) can run in parallel — each adds an independent block to `initSubTools()`

### Within Each Phase

- Implementation tasks before test tasks (tests reference the implementation)
- Asset removal (T016–T017) before test updates (T020–T025)
- Delegation implementation (T026–T028) before delegation tests (T029–T038)

---

## Implementation Strategy

### MVP First (US4 + US1)

1. Complete Phase 1: Setup (US4) — `uv` + `specify` installation
2. Complete Phase 2: Asset removal for Speckit
3. Complete Phase 3: Specify delegation
4. **STOP and VALIDATE**: `uf setup --dry-run` shows 14 steps, `uf init` delegates to `specify init`

### Incremental Delivery

1. US4 (Setup) → Test independently → `uf setup` installs specify
2. US1 (Speckit) → Test independently → `uf init` delegates to specify
3. US2 (OpenSpec) → Test independently → `uf init` delegates to openspec
4. US3 (Gaze) → Test independently → `uf init` delegates to gaze
5. Documentation → Final validation

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Gate on `openspec/config.yaml` (not `openspec/` directory) per R4 — the embedded custom schema creates `openspec/schemas/` before `initSubTools()` runs
- Gate on `.opencode/agents/gaze-reporter.md` (not a directory) per R6 — Gaze creates files inside `.opencode/`, not a workspace directory
- Step count is 14 (not 15) per R5 — the spec header "12→15" was a calculation error; 12 + 2 = 14
- All tool delegations follow the Replicator pattern: `LookPath` → `os.Stat` → `ExecCmd` → `subToolResult`
- Errors in tool delegations are warnings, not hard failures (Constitution Principle II — Composability First)

<!-- spec-review: passed -->
