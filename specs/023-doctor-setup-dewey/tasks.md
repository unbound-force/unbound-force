# Tasks: Doctor & Setup Dewey Alignment

**Input**: Design documents from `/specs/023-doctor-setup-dewey/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Tests are required — FR-008 mandates all existing tests pass, and new branches require unit test coverage per the coverage strategy in plan.md.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Project type**: Go CLI binary (single project)
- **Source**: `internal/doctor/`, `internal/setup/`
- **Tests**: Co-located `*_test.go` files

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Verify existing tests pass and understand current code before making changes

- [ ] T001 Run full test suite (`go test -race -count=1 ./...`) to establish green baseline per FR-008
- [ ] T002 [P] Read `internal/doctor/doctor.go` to understand `Options` struct and `defaults()` pattern for `EmbedCheck` field placement
- [ ] T003 [P] Read `internal/doctor/checks.go` `checkDewey()` function to understand current check flow and insertion point for embedding capability check
- [ ] T004 [P] Read `internal/setup/setup.go` `installSwarmPlugin()` function to understand current install flow and all `opencode-swarm-plugin@latest` references

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add the `EmbedCheck` injectable function field to `Options` — required by both US1 (embedding check) and US3 (Ollama demotion context)

**⚠️ CRITICAL**: US1 implementation depends on this field existing

- [x] T005 [US1] Add `EmbedCheck func(model string) error` field to `Options` struct in `internal/doctor/doctor.go` with GoDoc comment per data-model.md
- [x] T006 [US1] Add `defaultEmbedCheck(getenv func(string) string) func(model string) error` function in `internal/doctor/checks.go` — sends POST to Ollama `/api/embed` endpoint with 5-second timeout, reads `OLLAMA_HOST` env var (default `http://localhost:11434`), returns nil on success or descriptive error per contracts/doctor-checks.md
- [x] T007 [US1] Add `net/http` import to `internal/doctor/checks.go` (standard library, no new external dependency per plan.md constraints)
- [x] T008 [US1] Wire `EmbedCheck` default in `defaults()` method in `internal/doctor/doctor.go`: `if o.EmbedCheck == nil { o.EmbedCheck = defaultEmbedCheck(o.Getenv) }`

**Checkpoint**: `Options.EmbedCheck` field exists with production default. Existing tests still pass.

---

## Phase 3: User Story 1 — Dewey Embedding Health Check (Priority: P1) 🎯 MVP

**Goal**: `uf doctor` verifies Dewey can generate embeddings end-to-end, catching the most common failure mode (Ollama not serving or model not loaded) before engineers encounter cryptic errors during agent operations.

**Independent Test**: Run `uf doctor` with Dewey installed and embedding model pulled → embedding capability check PASS. Stop Ollama → embedding capability check WARN with actionable hint. Remove Dewey → embedding capability check skipped.

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T009 [P] [US1] Add `TestCheckDewey_EmbeddingCapability_Pass` in `internal/doctor/doctor_test.go` — inject mock `EmbedCheck` returning nil, verify CheckResult has `Name: "embedding capability"`, `Severity: Pass`, message contains model name per contracts/doctor-checks.md behavior matrix. Also verify `--format=json` output includes the check with correct `name`, `severity`, and `message` fields per FR-007 (automatic via existing `CheckResult` JSON serialization)
- [x] T010 [P] [US1] Add `TestCheckDewey_EmbeddingCapability_Fail` in `internal/doctor/doctor_test.go` — inject mock `EmbedCheck` returning error, verify CheckResult has `Severity: Warn`, `InstallHint` contains actionable fix command per contracts/doctor-checks.md
- [x] T011 [P] [US1] Add `TestCheckDewey_EmbeddingCapability_Skip` in `internal/doctor/doctor_test.go` — set `LookPath("dewey")` to return error, verify embedding capability check result has message "skipped: dewey not installed" per FR-003
- [x] T012 [P] [US1] Add `TestCheckDewey_EmbeddingCapability_ConnectionRefused` in `internal/doctor/doctor_test.go` — inject `EmbedCheck` returning error containing "connection refused", verify hint says "Start Ollama: ollama serve" per contracts/doctor-checks.md error categories

### Implementation for User Story 1

- [x] T013 [US1] Add `checkEmbeddingCapability(opts *Options) CheckResult` function in `internal/doctor/checks.go` per contracts/doctor-checks.md — calls `opts.EmbedCheck(graniteModel)`, returns Pass on nil, Warn with categorized hints on error (connection refused → "Start Ollama", model not found → "ollama pull", other → combined hint)
- [x] T014 [US1] Modify `checkDewey()` in `internal/doctor/checks.go` to call `checkEmbeddingCapability(opts)` after the existing "embedding model" check (check position 3 of 4 per contracts/doctor-checks.md updated check order)
- [x] T015 [US1] Add "embedding capability" skip result in `checkDewey()` skip block (when Dewey binary not found) — add `CheckResult{Name: "embedding capability", Severity: Pass, Message: "skipped: dewey not installed"}` alongside existing "embedding model" and "workspace" skip results per FR-003

**Checkpoint**: `uf doctor` shows embedding capability check in Dewey Knowledge Layer group. All 3 acceptance scenarios pass. JSON output includes the check.

---

## Phase 4: User Story 2 — Forked Swarm Plugin Installation (Priority: P1)

**Goal**: `uf setup` installs the Swarm plugin from the organization's fork (`unbound-force/swarm-tools`) instead of upstream (`joelhooks/swarm-tools`), ensuring engineers' environments match the team's expected configuration.

**Independent Test**: Run `uf setup` → verify install command references `github:unbound-force/swarm-tools`. Verify existing swarm binary is updated (not skipped) to ensure fork version.

### Tests for User Story 2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T017 [P] [US2] Add `TestInstallSwarmPlugin_ForkSource_Bun` in `internal/setup/setup_test.go` — verify `ExecCmd` receives `"bun", "add", "-g", "github:unbound-force/swarm-tools"` (not `opencode-swarm-plugin@latest`)
- [x] T018 [P] [US2] Add `TestInstallSwarmPlugin_ForkSource_Npm` in `internal/setup/setup_test.go` — verify `ExecCmd` receives `"npm", "install", "-g", "github:unbound-force/swarm-tools"` when bun is not available
- [x] T019 [P] [US2] Add `TestInstallSwarmPlugin_AlwaysInstalls` in `internal/setup/setup_test.go` — set `LookPath("swarm")` to return success (swarm already installed), verify install command still runs (no early return) per contracts/setup-swarm.md idempotent update behavior
- [x] T020 [P] [US2] Update existing `TestSetupRun_SwarmPluginNpmFails` in `internal/setup/setup_test.go` — change mock command key from `"npm install -g opencode-swarm-plugin@latest"` to `"npm install -g github:unbound-force/swarm-tools"` to match new install source

### Implementation for User Story 2

- [x] T021 [US2] Modify `installSwarmPlugin()` in `internal/setup/setup.go` — remove the early-return `LookPath("swarm")` check (lines 539-541) so the install always runs per contracts/setup-swarm.md idempotent update behavior
- [x] T022 [US2] Change install source in `installSwarmPlugin()` bun path from `"opencode-swarm-plugin@latest"` to `"github:unbound-force/swarm-tools"` in `internal/setup/setup.go`
- [x] T023 [US2] Change install source in `installSwarmPlugin()` npm path from `"opencode-swarm-plugin@latest"` to `"github:unbound-force/swarm-tools"` in `internal/setup/setup.go`
- [x] T024 [US2] Update dry-run messages in `installSwarmPlugin()` to reference `github:unbound-force/swarm-tools` instead of `opencode-swarm-plugin@latest` in `internal/setup/setup.go`
- [x] T025 [P] [US2] Update `managerInstallCmd("swarm", ManagerBun)` in `internal/doctor/environ.go` from `"bun add -g opencode-swarm-plugin@latest"` to `"bun add -g github:unbound-force/swarm-tools"` per contracts/setup-swarm.md install hint updates
- [x] T026 [P] [US2] Update `homebrewInstallCmd("swarm")` in `internal/doctor/environ.go` from `"npm install -g opencode-swarm-plugin@latest"` to `"npm install -g github:unbound-force/swarm-tools"`
- [x] T027 [P] [US2] Update `genericInstallCmd("swarm")` in `internal/doctor/environ.go` from `"npm install -g opencode-swarm-plugin@latest"` to `"npm install -g github:unbound-force/swarm-tools"`
- [x] T028 [US2] Update `checkSwarmPlugin()` install hint in `internal/doctor/checks.go` (line 337) from `"npm install -g opencode-swarm-plugin@latest"` to `"npm install -g github:unbound-force/swarm-tools"`
- [x] T029 [US2] Update all existing test assertions in `internal/doctor/doctor_test.go` that reference `"opencode-swarm-plugin@latest"` to use `"github:unbound-force/swarm-tools"` — affects `TestCheckSwarmPlugin_NotInstalled`, `TestInstallHint_*`, `TestManagerInstallCmd_*`
- [x] T030 [US2] Update all existing test assertions in `internal/setup/setup_test.go` that reference `"opencode-swarm-plugin@latest"` to use `"github:unbound-force/swarm-tools"` — affects mock command maps and assertion strings

**Checkpoint**: `uf setup` installs from fork. All 3 acceptance scenarios pass. All existing swarm-related tests updated and passing.

---

## Phase 5: User Story 3 — Ollama Check Demotion (Priority: P3)

**Goal**: Demote the direct Ollama serving check to informational status, reflecting that Dewey manages the Ollama lifecycle. Engineers no longer see misleading "Ollama: FAIL" when Dewey handles Ollama automatically.

**Independent Test**: Run `uf doctor` with Ollama not running → Ollama check shows informational message (not FAIL). Run with Ollama running → still shows PASS.

### Tests for User Story 3

- [x] T031 [P] [US3] Add `TestCheckDewey_OllamaDemotion` in `internal/doctor/doctor_test.go` — when Dewey is installed and embedding model check runs, verify the "embedding model" check message includes "(Dewey manages Ollama lifecycle)" annotation per contracts/doctor-checks.md

### Implementation for User Story 3

- [x] T032 [US3] Modify `checkDewey()` in `internal/doctor/checks.go` — when Dewey binary is found and the embedding model check passes, annotate the "embedding model" result message with "(Dewey manages Ollama lifecycle)" per contracts/doctor-checks.md and research.md R2
- [x] T033 [US3] Verify the Ollama demotion annotation appears in `--format=json` output per FR-007 (automatic via `CheckResult.Message` field — verify in test T031)

**Checkpoint**: Ollama check shows informational annotation. Both acceptance scenarios pass. JSON output reflects the annotation.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation updates, full test suite validation, and cross-cutting verification

- [x] T034 Run full test suite (`go test -race -count=1 ./...`) to verify all existing tests pass after all changes per FR-008
- [x] T035 Run `golangci-lint run` to verify no lint regressions
- [x] T036 [P] Update `AGENTS.md` "Active Technologies" section — add `net/http` (standard library — new import in `checks.go`) to the 023-doctor-setup-dewey entry
- [x] T037 [P] Update `AGENTS.md` "Recent Changes" section — add 023-doctor-setup-dewey entry summarizing: embedding capability check added to `uf doctor`, Swarm plugin source changed to fork, Ollama check demoted, files modified, user stories and task count
- [x] T038 Run quickstart.md verification — execute the manual verification steps from `specs/023-doctor-setup-dewey/quickstart.md` to confirm end-to-end behavior

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS US1 implementation (US2 and US3 do not depend on Phase 2)
- **US1 (Phase 3)**: Depends on Phase 2 (`EmbedCheck` field must exist)
- **US2 (Phase 4)**: Depends on Phase 1 only — can run in parallel with Phase 2 and US1
- **US3 (Phase 5)**: Depends on Phase 1 only — can run in parallel with Phase 2, US1, and US2
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Depends on Phase 2 (`EmbedCheck` field on `Options`). Touches `internal/doctor/doctor.go` and `internal/doctor/checks.go`.
- **User Story 2 (P1)**: Independent of other stories. Touches `internal/setup/setup.go`, `internal/doctor/environ.go`, `internal/doctor/checks.go` (line 337 only), and test files.
- **User Story 3 (P3)**: Independent of other stories. Touches `internal/doctor/checks.go` (`checkDewey()` function only).

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Implementation follows contract specifications
- Story complete before moving to next priority

### Parallel Opportunities

```text
Phase 1: T002, T003, T004 can run in parallel (read-only)
Phase 2: T005 and T006 are sequential (T006 depends on T005 for type reference)
Phase 3 tests: T009, T010, T011, T012 can run in parallel (different test functions)
Phase 4 tests: T017, T018, T019, T020 can run in parallel (different test functions)
Phase 4 impl: T025, T026, T027 can run in parallel (different functions in environ.go)
Phase 6: T036, T037 can run in parallel (different sections of AGENTS.md)

US2 can run in parallel with US1 (different files except checks.go line 337)
US3 can run in parallel with US2 (different functions in checks.go)
```

---

## Parallel Example: User Story 1 Tests

```bash
# Launch all US1 tests together (different test functions, no dependencies):
Task: "TestCheckDewey_EmbeddingCapability_Pass in internal/doctor/doctor_test.go"
Task: "TestCheckDewey_EmbeddingCapability_Fail in internal/doctor/doctor_test.go"
Task: "TestCheckDewey_EmbeddingCapability_Skip in internal/doctor/doctor_test.go"
Task: "TestCheckDewey_EmbeddingCapability_ConnectionRefused in internal/doctor/doctor_test.go"
```

## Parallel Example: User Story 2 Install Hints

```bash
# Launch all environ.go hint updates together (different functions):
Task: "Update managerInstallCmd in internal/doctor/environ.go"
Task: "Update homebrewInstallCmd in internal/doctor/environ.go"
Task: "Update genericInstallCmd in internal/doctor/environ.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (verify green baseline)
2. Complete Phase 2: Foundational (`EmbedCheck` field)
3. Complete Phase 3: User Story 1 (embedding capability check)
4. **STOP and VALIDATE**: Run `uf doctor` with Dewey — verify embedding check appears
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → `EmbedCheck` field ready
2. Add User Story 1 → Test independently → Embedding check works (MVP!)
3. Add User Story 2 → Test independently → Fork install works
4. Add User Story 3 → Test independently → Ollama demotion works
5. Each story adds value without breaking previous stories

### File Modification Summary

| File | Stories | Changes |
|------|---------|---------|
| `internal/doctor/doctor.go` | US1 | Add `EmbedCheck` field, wire default in `defaults()` |
| `internal/doctor/checks.go` | US1, US2, US3 | Add `checkEmbeddingCapability()`, `defaultEmbedCheck()`, modify `checkDewey()`, update swarm install hint |
| `internal/doctor/environ.go` | US2 | Update 3 install hint functions for fork source |
| `internal/doctor/doctor_test.go` | US1, US2, US3 | Add 5 new tests, update existing swarm hint assertions |
| `internal/setup/setup.go` | US2 | Modify `installSwarmPlugin()` source and early-return |
| `internal/setup/setup_test.go` | US2 | Add 4 new tests, update existing mock command strings |
| `AGENTS.md` | All | Update Active Technologies and Recent Changes |

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- No new external Go dependencies — only `net/http` from standard library
- All new code follows existing `Options` injection pattern for testability
- `graniteModel` constant already defined in `internal/setup/setup.go` — reuse or define locally in `checks.go`

<!-- spec-review: passed -->
