# Tasks: Replicator Migration

**Input**: Design documents from `/specs/024-replicator-migration/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Verify Baseline)

**Purpose**: Confirm the codebase compiles and all existing tests pass before any migration work begins.

- [x] T001 Run `make check` to verify all existing tests pass at baseline
- [x] T002 Run `go build ./...` to confirm clean compilation

**Checkpoint**: Baseline verified — all tests pass, code compiles. Migration work can begin.

---

## Phase 2: US1 — Replicator Installation via Setup (Priority: P1)

**Goal**: `uf setup` installs Replicator via Homebrew instead of the Swarm plugin via npm/bun. Step count changes from 15 to 12.

**Independent Test**: Run `uf setup` on a machine without Replicator installed. Verify Replicator is installed via Homebrew, `replicator setup` runs, and bun is not installed or required.

### Implementation

- [x] T003 [US1] Delete `const swarmForkSource` from `internal/setup/setup.go`
- [x] T004 [US1] Delete `ensureBun()` function from `internal/setup/setup.go` (FR-005)
- [x] T005 [US1] Delete `installSwarmPlugin()` function from `internal/setup/setup.go` (FR-005)
- [x] T006 [US1] Delete `runSwarmSetup()` function from `internal/setup/setup.go` (FR-005)
- [x] T007 [US1] Delete `initializeHive()` function from `internal/setup/setup.go` (FR-005)
- [x] T008 [US1] Add `installReplicator()` function to `internal/setup/setup.go` following `installGaze()` pattern with `brew install unbound-force/tap/replicator` (FR-001, per contract setup-replicator.md)
- [x] T009 [US1] Add `runReplicatorSetup()` function to `internal/setup/setup.go` following `runSwarmSetup()` pattern with `replicator setup` command (FR-002, per contract setup-replicator.md)
- [x] T010 [US1] Modify `installOpenSpec()` in `internal/setup/setup.go` to remove bun preference block and use npm-only: `npm install -g @fission-ai/openspec@latest` (FR-004)
- [x] T011 [US1] Modify `Run()` in `internal/setup/setup.go` to restructure step flow from 15 to 12 steps: remove bun/swarm/hive/init steps, add replicator + replicator setup steps, reorder per contract step flow (FR-006)
- [x] T012 [US1] Update embedding model completion message in `internal/setup/setup.go` to remove "Swarm" reference (per contract setup-replicator.md)
- [x] T013 [US1] Update package doc comment in `internal/setup/setup.go` to reference Replicator instead of Swarm plugin

### Tests

- [x] T014 [US1] Remove tests for `ensureBun`, `installSwarmPlugin`, `runSwarmSetup`, `initializeHive` from `internal/setup/setup_test.go`
- [x] T015 [US1] Add test for `installReplicator` — replicator already installed returns "already installed" in `internal/setup/setup_test.go`
- [x] T016 [US1] Add test for `installReplicator` — dry-run with Homebrew available in `internal/setup/setup_test.go`
- [x] T017 [US1] Add test for `installReplicator` — Homebrew not available returns skip with GitHub releases link in `internal/setup/setup_test.go`
- [x] T018 [US1] Add test for `installReplicator` — successful Homebrew install in `internal/setup/setup_test.go`
- [x] T018b [US1] Add test for `installReplicator` — Homebrew install fails returns failed result in `internal/setup/setup_test.go`
- [x] T019 [US1] Add test for `runReplicatorSetup` — dry-run returns dry-run result in `internal/setup/setup_test.go`
- [x] T020 [US1] Add test for `runReplicatorSetup` — successful execution in `internal/setup/setup_test.go`
- [x] T021 [US1] Add test for `runReplicatorSetup` — failure returns failed result in `internal/setup/setup_test.go`
- [x] T022 [US1] Update `installOpenSpec` tests to assert npm-only (no bun preference) in `internal/setup/setup_test.go`
- [x] T023 [US1] Update integration tests for new step count `[N/12]` format in `internal/setup/setup_test.go`

**Checkpoint**: `uf setup` installs Replicator via Homebrew. All setup tests pass. `go test -race -count=1 ./internal/setup/...` green.

---

## Phase 3: US2 — Project Initialization with Replicator (Priority: P1)

**Goal**: `uf init` configures `opencode.json` with an `mcp.replicator` entry instead of a `plugin` array entry. Legacy plugin entries are migrated. `replicator init` is delegated for per-repo setup.

**Independent Test**: Run `uf init` in a fresh directory with Replicator installed. Verify `opencode.json` contains `mcp.replicator` (not plugin array) and `.hive/` is created via `replicator init`.

### Implementation

- [x] T024 [US2] Replace plugin array logic (Swarm plugin entry block) in `configureOpencodeJSON()` in `internal/scaffold/scaffold.go` with Replicator MCP entry logic following Dewey MCP pattern (FR-007, FR-010, per contract init-replicator.md)
- [x] T025 [US2] Add legacy plugin migration to `configureOpencodeJSON()` in `internal/scaffold/scaffold.go` — remove `opencode-swarm-plugin` from `plugin` array, remove empty `plugin` key (FR-008, per contract init-replicator.md)
- [x] T026 [US2] Change detection condition in `configureOpencodeJSON()` in `internal/scaffold/scaffold.go` from `.hive/` directory check to `LookPath("replicator")` binary check (per contract init-replicator.md)
- [x] T027 [US2] Add Replicator init delegation to `initSubTools()` in `internal/scaffold/scaffold.go` — check `LookPath("replicator")`, check `.hive/` absence, run `replicator init` (FR-009, per contract init-replicator.md). Note: depends on `unbound-force/replicator#5`; if not available, implement with graceful skip per Constitution Principle II
- [x] T028 [US2] Update `configureOpencodeJSON()` doc comment in `internal/scaffold/scaffold.go` to reference Replicator MCP entry instead of Swarm plugin

### Tests

- [x] T029 [US2] Update tests in `internal/scaffold/scaffold_test.go` that assert on `plugin` array to assert on `mcp.replicator` entry instead
- [x] T030 [US2] Add test for legacy plugin migration — `opencode-swarm-plugin` in plugin array is removed and `mcp.replicator` added in `internal/scaffold/scaffold_test.go`
- [x] T031 [US2] Add test for legacy plugin migration — empty plugin array after removal causes `plugin` key deletion in `internal/scaffold/scaffold_test.go`
- [x] T032 [US2] Add test for idempotent behavior — existing `mcp.replicator` entry is preserved in `internal/scaffold/scaffold_test.go`
- [x] T033 [US2] Add test for Replicator not installed — no `mcp.replicator` entry and no `plugin` array added in `internal/scaffold/scaffold_test.go`
- [x] T034 [US2] Add test for `replicator init` delegation — `.hive/` absent triggers `replicator init` in `internal/scaffold/scaffold_test.go`
- [x] T035 [US2] Add test for `replicator init` delegation — `.hive/` present skips init in `internal/scaffold/scaffold_test.go`
- [x] T036 [US2] Add test for `replicator init` failure — error captured as warning, does not block remaining init in `internal/scaffold/scaffold_test.go`

**Checkpoint**: `uf init` produces correct `opencode.json` with `mcp.replicator`. Legacy migration works. All scaffold tests pass. `go test -race -count=1 ./internal/scaffold/...` green.

---

## Phase 4: US3 — Doctor Replicator Health Checks (Priority: P2)

**Goal**: `uf doctor` verifies Replicator installation and configuration instead of checking for the Swarm plugin.

**Independent Test**: Run `uf doctor` with Replicator installed and configured. Verify "Replicator" check group with binary, doctor, `.hive/`, and MCP config checks.

### Implementation

- [x] T037 [US3] Delete `checkSwarmPlugin()` function from `internal/doctor/checks.go` (FR-011)
- [x] T038 [US3] Add `checkReplicator()` function to `internal/doctor/checks.go` with 4 sub-checks: binary (LookPath), doctor delegation (ExecCmdTimeout 10s), `.hive/` (os.Stat), MCP config (mcp.replicator in opencode.json) (FR-011, per contract doctor-replicator.md)
- [x] T039 [US3] Update `coreToolSpecs` in `internal/doctor/checks.go` to replace `"swarm"` entry with `"replicator"` entry (per research R11)
- [x] T040 [P] [US3] Replace `checkSwarmPlugin(&opts)` with `checkReplicator(&opts)` in `groups` slice in `internal/doctor/doctor.go` (per contract doctor-replicator.md)
- [x] T041 [P] [US3] Update `managerInstallCmd()` in `internal/doctor/environ.go` — remove `ManagerBun` case for `"swarm"` (FR-013, per contract doctor-replicator.md)
- [x] T042 [P] [US3] Update `homebrewInstallCmd()` in `internal/doctor/environ.go` — replace `"swarm"` case with `"replicator"` returning `"brew install unbound-force/tap/replicator"` (FR-012)
- [x] T043 [P] [US3] Update `genericInstallCmd()` in `internal/doctor/environ.go` — replace `"swarm"` case with `"replicator"` returning `"brew install unbound-force/tap/replicator"` (FR-012)
- [x] T044 [P] [US3] Add `"replicator"` entry to `installURL()` in `internal/doctor/environ.go` returning `"https://github.com/unbound-force/replicator"` (per contract doctor-replicator.md)

### Tests

- [x] T045 [US3] Remove `checkSwarmPlugin` tests from `internal/doctor/doctor_test.go`
- [x] T046 [US3] Add test for `checkReplicator` — replicator installed, all checks pass in `internal/doctor/doctor_test.go`
- [x] T047 [US3] Add test for `checkReplicator` — replicator not installed, returns Warn with Homebrew install hint in `internal/doctor/doctor_test.go`
- [x] T048 [US3] Add test for `checkReplicator` — replicator doctor timeout returns Warn in `internal/doctor/doctor_test.go`
- [x] T049 [US3] Add test for `checkReplicator` — `.hive/` missing returns Warn with `uf init` hint in `internal/doctor/doctor_test.go`
- [x] T050 [US3] Add test for `checkReplicator` — `mcp.replicator` missing from opencode.json returns Warn in `internal/doctor/doctor_test.go`
- [x] T051 [US3] Update install hint assertion strings in existing tests to reference `"replicator"` instead of `"swarm"` in `internal/doctor/doctor_test.go`

**Checkpoint**: `uf doctor` displays "Replicator" check group. All doctor tests pass. `go test -race -count=1 ./internal/doctor/...` green.

---

## Phase 5: US4 — Bun Removal (Priority: P3)

**Goal**: Setup no longer installs bun as a prerequisite. OpenSpec CLI installs via npm-only.

**Independent Test**: Run `uf setup` and verify bun is not installed or referenced.

> Note: Most bun removal tasks are already covered by Phase 2 (T004 deletes `ensureBun`, T010 makes OpenSpec npm-only, T011 removes bun step from flow). This phase covers verification and any remaining bun references.

- [x] T052 [US4] Verify no bun install hints remain in `internal/doctor/environ.go` for any Unbound Force tool (FR-013) — the `ManagerBun` detection remains for user projects but no UF tool uses it
- [x] T053 [US4] Verify no bun references remain in `internal/setup/setup.go` (FR-003)
- [x] T054 [US4] Verify no `bun add -g` references remain in any production source file (SC-004)

**Checkpoint**: Zero bun references in production code for Unbound Force tool installation.

---

## Phase 6: Polish (Docs, Config, Final Validation)

**Purpose**: Update configuration files, documentation, agent/command files, and run final validation.

### Configuration

- [x] T055 [US2] Update live `opencode.json` at repo root — remove `plugin` array, add `mcp.replicator` entry per data-model.md target schema (FR-014)

### Agent/Command Files

- [x] T056 [P] Update `.opencode/command/unleash.md` — change "Swarm plugin" install hint text to "Replicator" (FR-016)
- [x] T057 [P] Update `internal/scaffold/assets/opencode/command/unleash.md` — same change as T056 to keep scaffold asset in sync (FR-016)

### Regression Tests

- [x] T058 Add regression test `TestScaffoldOutput_NoSwarmPluginReferences` to `internal/scaffold/scaffold_test.go` verifying zero occurrences of `opencode-swarm-plugin`, `installSwarmPlugin`, `ensureBun`, `swarmForkSource` in scaffold output (SC-004)

### Documentation

- [x] T059 [P] Add Recent Changes entry to `AGENTS.md` summarizing the Replicator migration (user stories, task count, key changes)

### Final Validation

- [x] T060 Run `make check` to verify all tests pass after migration (FR-015, SC-005)
- [x] T061 Run `go test -race -count=1 ./...` to verify full test suite (SC-005)
- [x] T062 Grep production source for stale references: `opencode-swarm-plugin`, `ensureBun`, `installSwarmPlugin`, `bun add -g` — verify zero matches (SC-004)

**Checkpoint**: All tests pass. Zero stale references. Documentation updated. Migration complete.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — verify baseline first
- **Phase 2 (US1)**: Depends on Phase 1 — modifies `setup.go` and `setup_test.go`
- **Phase 3 (US2)**: Depends on Phase 1 — modifies `scaffold.go` and `scaffold_test.go` (independent of Phase 2)
- **Phase 4 (US3)**: Depends on Phase 1 — modifies `checks.go`, `doctor.go`, `environ.go`, `doctor_test.go` (independent of Phases 2-3)
- **Phase 5 (US4)**: Depends on Phases 2-4 — verification of bun removal across all modified files
- **Phase 6 (Polish)**: Depends on Phases 2-4 — documentation and config updates after code changes

### Parallel Opportunities

- **Phases 2, 3, 4** can proceed in parallel after Phase 1 (different source files, no cross-dependencies):
  - Phase 2 touches `internal/setup/`
  - Phase 3 touches `internal/scaffold/`
  - Phase 4 touches `internal/doctor/`
- Within Phase 4: T040-T044 are marked `[P]` (different files: `doctor.go`, `environ.go`)
- Within Phase 6: T056-T057 are marked `[P]` (different files), T059 is independent

### Within Each Phase

- Delete functions before adding replacements (clean compile state)
- Implementation tasks before test tasks (tests need functions to exist)
- Core logic before doc comments (substance before polish)

---

## Summary

| Phase | Story | Tasks | Files |
|-------|-------|-------|-------|
| 1 | — | 2 | — |
| 2 | US1 (P1) | 22 | `setup.go`, `setup_test.go` |
| 3 | US2 (P1) | 13 | `scaffold.go`, `scaffold_test.go` |
| 4 | US3 (P2) | 15 | `checks.go`, `doctor.go`, `environ.go`, `doctor_test.go` |
| 5 | US4 (P3) | 3 | (verification only) |
| 6 | Polish | 8 | `opencode.json`, `unleash.md` (x2), `scaffold_test.go`, `AGENTS.md` |
| **Total** | **4 stories** | **63** | **11 files** |

<!-- spec-review: passed -->
