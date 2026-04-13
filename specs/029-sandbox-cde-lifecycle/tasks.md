# Tasks: Sandbox CDE Lifecycle

**Input**: Design documents from `/specs/029-sandbox-cde-lifecycle/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Verify Baseline)

**Purpose**: Confirm Spec 028 baseline is green before making changes

- [x] T001 Run `go test -race -count=1 ./internal/sandbox/...` and verify all existing Spec 028 tests pass
- [x] T002 Run `go test -race -count=1 ./cmd/unbound-force/...` and verify all existing CLI tests pass
- [x] T003 Run `go vet ./...` and `golangci-lint run` — confirm zero findings

**Checkpoint**: Baseline is green — all existing tests pass, no lint findings

---

## Phase 2: Backend Interface + Podman Persistent (Foundational)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**CRITICAL**: No user story work can begin until this phase is complete

### Backend interface definition

- [x] T004 [P] Create `internal/sandbox/backend.go` — define `Backend` interface with 7 methods (`Create`, `Start`, `Stop`, `Destroy`, `Status`, `Attach`, `Name`) per `contracts/backend-interface.md`
- [x] T005 [P] Add backend constants to `internal/sandbox/backend.go` — `BackendAuto`, `BackendPodman`, `BackendChe`, `ModePersistent`, `DefaultConfigPath` per `data-model.md`
- [x] T006 [P] Create `internal/sandbox/workspace.go` — define `WorkspaceStatus` struct with all fields (`Exists`, `Running`, `Backend`, `Name`, `ID`, `Image`, `Mode`, `ProjectDir`, `ServerURL`, `DemoEndpoints`, `StartedAt`, `ExitCode`, `Persistent`) per `data-model.md`
- [x] T007 [P] Add `DemoEndpoint` struct to `internal/sandbox/workspace.go` — fields: `Name`, `Port`, `URL`, `Protocol` per `data-model.md`
- [x] T008 [P] Add `SandboxConfig`, `CheConfig`, `OllamaConfig` structs to `internal/sandbox/workspace.go` with YAML struct tags per `data-model.md`

### Workspace helpers

- [x] T009 Add `projectName(dir string) string` function to `internal/sandbox/workspace.go` — sanitize directory basename to lowercase alphanumeric + hyphens, fallback to "default" per research R4
- [x] T010 Add `LoadConfig(opts Options) (SandboxConfig, error)` function to `internal/sandbox/workspace.go` — parse `.uf/sandbox.yaml` with env var override precedence (flag > env > config > default) per research R7
- [x] T011 Add `FormatWorkspaceStatus(ws WorkspaceStatus, w io.Writer) error` function to `internal/sandbox/workspace.go` — format status output for both Podman and CDE backends per `contracts/cobra-commands.md` status output format

### Options extension

- [x] T012 Extend `Options` struct in `internal/sandbox/sandbox.go` — add `BackendName`, `WorkspaceName`, `DemoPorts`, `ConfigPath`, `CheURL` fields per `data-model.md` (all zero-value defaults preserve Spec 028 behavior)

### Backend resolver

- [x] T013 Implement `ResolveBackend(opts Options) (Backend, error)` in `internal/sandbox/backend.go` — resolution order: `--backend` flag > `UF_SANDBOX_BACKEND` env > `.uf/sandbox.yaml` > auto-detect (CDE if chectl/UF_CHE_URL, Podman otherwise) per `contracts/backend-interface.md` resolution matrix

### PodmanBackend implementation

- [x] T014 Create `internal/sandbox/podman.go` — define `PodmanBackend` struct implementing `Backend` interface with `Name() string` returning `"podman"`
- [x] T015 Implement `PodmanBackend.Create(opts Options) error` in `internal/sandbox/podman.go` — named volume create (`uf-sandbox-<project>`), container run with volume mount, `podman cp` to seed workspace, health check wait per `contracts/backend-interface.md` Create contract
- [x] T016 Implement `PodmanBackend.Start(opts Options) error` in `internal/sandbox/podman.go` — detect persistent workspace (named volume exists) and resume via `podman start`, or fall back to ephemeral mode (Spec 028 `Start()` behavior) per research R9
- [x] T017 Implement `PodmanBackend.Stop(opts Options) error` in `internal/sandbox/podman.go` — persistent mode: `podman stop` (preserve volume); ephemeral mode: `podman stop` + `podman rm` (Spec 028 behavior) per research R9
- [x] T018 Implement `PodmanBackend.Destroy(opts Options) error` in `internal/sandbox/podman.go` — `podman rm` + `podman volume rm`, idempotent per `contracts/backend-interface.md` Destroy contract
- [x] T019 Implement `PodmanBackend.Status(opts Options) (WorkspaceStatus, error)` in `internal/sandbox/podman.go` — `podman inspect` + `podman volume inspect`, populate `WorkspaceStatus` with demo endpoints per `contracts/backend-interface.md`
- [x] T020 Implement `PodmanBackend.Attach(opts Options) error` in `internal/sandbox/podman.go` — delegate to `opencode attach http://localhost:4096` (same as Spec 028)

### Config file support

- [x] T021 Extend `internal/sandbox/config.go` — add named volume support to `buildRunArgs()`: when `WorkspaceName` is set, use `-v <volume-name>:/workspace` instead of bind mount, add demo port `-p` mappings from `DemoPorts` per research R3

### Backward compatibility wiring

- [x] T022 Update `Start()` in `internal/sandbox/sandbox.go` — add persistent workspace detection (check for named volume via `podman volume inspect`); if persistent, dispatch to `PodmanBackend.Start()`; if not, preserve existing ephemeral behavior per research R9
- [x] T023 Update `Stop()` in `internal/sandbox/sandbox.go` — add persistent workspace detection; if persistent, dispatch to `PodmanBackend.Stop()` (preserve volume); if ephemeral, preserve existing remove behavior per research R9
- [x] T024 Update `Status()` in `internal/sandbox/sandbox.go` — check for persistent workspace first, return `WorkspaceStatus` via `FormatWorkspaceStatus()`; fall back to existing `ContainerStatus` format for ephemeral per `contracts/cobra-commands.md`

**Checkpoint**: Backend interface defined, PodmanBackend fully implemented, backward compatibility preserved — `go test -race -count=1 ./internal/sandbox/...` passes

---

## Phase 3: US1 — Persistent Workspace Creation (Priority: P1)

**Goal**: Engineer can `uf sandbox create` and `uf sandbox destroy` to manage persistent workspace lifecycle

**Independent Test**: Run `uf sandbox create`, stop, start, verify state preserved, then `uf sandbox destroy`

### Implementation

- [x] T025 [US1] Wire `Create()` dispatch in `internal/sandbox/sandbox.go` — public `Create(opts Options) error` function that calls `ResolveBackend()` then `backend.Create()`, with pre-flight checks (workspace already exists → error with hint) per spec US1 acceptance scenario 1
- [x] T026 [US1] Wire `Destroy()` dispatch in `internal/sandbox/sandbox.go` — public `Destroy(opts Options) error` function that calls `ResolveBackend()` then `backend.Destroy()`, with confirmation prompt support via `opts.Stdin` per spec US1 acceptance scenario 4
- [x] T027 [US1] Add auto-attach behavior to `Create()` in `internal/sandbox/sandbox.go` — after successful create, auto-attach TUI unless `--detach` flag is set per `contracts/cobra-commands.md` create flow
- [x] T028 [US1] Add edge case handling to `Create()` — "sandbox already exists" error with suggestion to use `start` or `destroy` per spec Edge Cases

**Checkpoint**: `Create()` and `Destroy()` work end-to-end with PodmanBackend — persistent lifecycle functional

---

## Phase 4: US4 — CDE Backend (Priority: P1)

**Goal**: Engineer can provision workspaces via Eclipse Che / Dev Spaces

**Independent Test**: Configure Che URL, run `uf sandbox create --backend che`, verify workspace created from devfile

### CheBackend implementation

- [x] T029 [US4] Create `internal/sandbox/che.go` — define `CheBackend` struct with `cheURL string` and `useChectl bool` fields, implement `Name() string` returning `"che"` per `data-model.md`
- [x] T030 [US4] Implement `CheBackend.Create(opts Options) error` in `internal/sandbox/che.go` — chectl primary path: `chectl workspace:create --devfile=devfile.yaml --name=uf-<project>`, REST API fallback: `POST /api/workspace/devfile`, devfile existence check, wait for RUNNING state per `contracts/backend-interface.md` CheBackend Create contract
- [x] T031 [US4] Implement `CheBackend.Start(opts Options) error` in `internal/sandbox/che.go` — chectl: `chectl workspace:start --name=uf-<project>`, REST API: `PATCH /api/workspace/{id}/runtime` per research R2
- [x] T032 [US4] Implement `CheBackend.Stop(opts Options) error` in `internal/sandbox/che.go` — chectl: `chectl workspace:stop --name=uf-<project>`, REST API: `PATCH /api/workspace/{id}/runtime` per research R2
- [x] T033 [US4] Implement `CheBackend.Destroy(opts Options) error` in `internal/sandbox/che.go` — chectl: `chectl workspace:delete --name=uf-<project> --yes`, REST API: `DELETE /api/workspace/{id}` per research R2
- [x] T034 [US4] Implement `CheBackend.Status(opts Options) (WorkspaceStatus, error)` in `internal/sandbox/che.go` — chectl: `chectl workspace:list --output=json`, REST API: `GET /api/workspace/{id}`, parse endpoint URLs into `DemoEndpoints` per `contracts/backend-interface.md`
- [x] T035 [US4] Implement `CheBackend.Attach(opts Options) error` in `internal/sandbox/che.go` — construct Che endpoint URL for OpenCode server, delegate to `opencode attach <che-endpoint-url>` per `contracts/backend-interface.md` Attach contract
- [x] T036 [US4] Add CDE error handling in `internal/sandbox/che.go` — actionable error messages: "devfile.yaml not found" with `--backend podman` suggestion, "cannot reach Che" with URL/auth hint, "workspace already exists" with start/destroy suggestion per spec Edge Cases
- [x] T037 [US4] Update `ResolveBackend()` in `internal/sandbox/backend.go` — wire `CheBackend` construction: set `cheURL` from flag/env/config, set `useChectl` from `LookPath("chectl")` per `contracts/backend-interface.md` resolution matrix
- [x] T038 [US4] Add `HTTPClient` injection point to `Options` struct in `internal/sandbox/sandbox.go` — injectable `func(req *http.Request) (*http.Response, error)` for CDE REST API calls, defaulting to `http.DefaultClient.Do` per FR-017

**Checkpoint**: CheBackend fully implemented — both chectl and REST API paths work with injected dependencies

---

## Phase 5: US2 — Demo Loop Integration (Priority: P1)

**Goal**: Demo review happens inside the workspace without extraction — iterative `/unleash` → demo → clarify → `/unleash` loop

**Independent Test**: Create sandbox, run `/unleash` to demo step, review in IDE, provide feedback, re-run `/unleash`, verify context preserved

### Implementation

- [x] T039 [US2] Add demo endpoint discovery to `PodmanBackend.Create()` in `internal/sandbox/podman.go` — read `DemoPorts` from config file and CLI flags, add `-p` mappings to container run args per research R6
- [x] T040 [US2] Add demo endpoint discovery to `CheBackend.Create()` in `internal/sandbox/che.go` — parse devfile endpoints section to populate `DemoEndpoints` in status output per research R6
- [x] T041 [US2] Add demo endpoint display to `FormatWorkspaceStatus()` in `internal/sandbox/workspace.go` — list each `DemoEndpoint` with name and URL per `contracts/cobra-commands.md` status output format
- [x] T042 [US2] Display demo endpoint URLs after `Create()` completes in `internal/sandbox/sandbox.go` — print server URL and demo endpoint URLs to stdout for engineer review per quickstart.md output examples

**Checkpoint**: Demo endpoints are discoverable via `status` and displayed after `create` — iterative loop is supported

---

## Phase 6: US3 — Bidirectional Git Sync (Priority: P2)

**Goal**: Changes flow bidirectionally between host and workspace via git push/pull

**Independent Test**: Agent commits inside workspace and pushes; engineer pulls on host; engineer pushes spec edit; workspace sees it

### Implementation

- [x] T043 [US3] Add `setupGitSync(opts Options) error` function to `internal/sandbox/workspace.go` — configure git remote in workspace, detect current branch, set up credential forwarding per `contracts/git-sync.md` Workspace Git Setup
- [x] T044 [US3] Call `setupGitSync()` from `PodmanBackend.Create()` in `internal/sandbox/podman.go` — after `podman cp` seeds the workspace, configure git inside the container per `contracts/git-sync.md` Podman Backend section
- [x] T045 [US3] Call `setupGitSync()` from `CheBackend.Create()` in `internal/sandbox/che.go` — Che handles git clone from devfile `projects` section, verify remote is configured per `contracts/git-sync.md` CDE Backend section
- [x] T046 [US3] Add `checkGitSync(opts Options) error` function to `internal/sandbox/workspace.go` — verify workspace git state is clean and up-to-date, auto-pull if fast-forward possible, error on divergence per `contracts/git-sync.md` Conflict Detection
- [x] T047 [US3] Update `Extract()` in `internal/sandbox/sandbox.go` — for persistent workspaces with push access, suggest `git pull` instead of extract; for workspaces without push access, fall back to Spec 028 format-patch/am workflow per `contracts/git-sync.md` Extract Compatibility

**Checkpoint**: Bidirectional git sync works for both backends — `go test -race -count=1 ./internal/sandbox/...` passes

---

## Phase 7: Cobra Commands

**Purpose**: Wire new and updated subcommands in the CLI layer

### New commands

- [x] T048 [US1] Add `newSandboxCreateCmd()` to `cmd/unbound-force/sandbox.go` — Cobra command with flags: `--backend`, `--image`, `--memory`, `--cpus`, `--name`, `--detach`, `--demo-ports`; delegates to `runSandboxCreate(params)` per `contracts/cobra-commands.md` create spec
- [x] T049 [US1] Implement `runSandboxCreate(p sandboxCreateParams) error` in `cmd/unbound-force/sandbox.go` — construct `sandbox.Options` from params, call `sandbox.Create()`, handle errors with user-facing messages per `contracts/cobra-commands.md` delegation pattern
- [x] T050 [US1] Add `newSandboxDestroyCmd()` to `cmd/unbound-force/sandbox.go` — Cobra command with flags: `--yes`, `--force`; delegates to `runSandboxDestroy(params)` per `contracts/cobra-commands.md` destroy spec
- [x] T051 [US1] Implement `runSandboxDestroy(p sandboxDestroyParams) error` in `cmd/unbound-force/sandbox.go` — construct `sandbox.Options`, show confirmation prompt (unless `--yes`), call `sandbox.Destroy()` per `contracts/cobra-commands.md` confirmation prompt

### Updated commands

- [x] T052 [US1] Update `newSandboxStartCmd()` in `cmd/unbound-force/sandbox.go` — add `--backend` flag, update `runSandboxStart()` to pass `BackendName` to `sandbox.Options` for persistent workspace detection per `contracts/cobra-commands.md` start update
- [x] T053 [US1] Update `newSandboxStopCmd()` in `cmd/unbound-force/sandbox.go` — update `runSandboxStop()` to pass persistent mode context to `sandbox.Stop()` per `contracts/cobra-commands.md` stop update
- [x] T054 [US1] Update `newSandboxStatusCmd()` in `cmd/unbound-force/sandbox.go` — update `runSandboxStatus()` to use `FormatWorkspaceStatus()` for persistent workspaces, fall back to existing format for ephemeral per `contracts/cobra-commands.md` status update

### Command registration

- [x] T055 [US1] Register `newSandboxCreateCmd()` and `newSandboxDestroyCmd()` in `newSandboxCmd()` in `cmd/unbound-force/sandbox.go` — add to `cmd.AddCommand()` call per `contracts/cobra-commands.md` Command Registration

**Checkpoint**: All CLI commands wired — `go build ./cmd/unbound-force/...` succeeds, `uf sandbox --help` shows all 7 subcommands

---

## Phase 8: Tests

**Purpose**: Comprehensive test coverage for all new code per `contracts/testing-strategy.md`

### Backend resolver tests

- [x] T056 [P] Add `TestResolveBackend_AutoPodman` to `internal/sandbox/sandbox_test.go` — no CDE configured, auto-detect selects Podman
- [x] T057 [P] Add `TestResolveBackend_AutoChe` to `internal/sandbox/sandbox_test.go` — `UF_CHE_URL` set, auto-detect selects Che
- [x] T058 [P] Add `TestResolveBackend_ExplicitPodman` to `internal/sandbox/sandbox_test.go` — `--backend podman` selects Podman regardless of CDE config
- [x] T059 [P] Add `TestResolveBackend_ExplicitChe` to `internal/sandbox/sandbox_test.go` — `--backend che` with CDE configured selects Che
- [x] T060 [P] Add `TestResolveBackend_CheNotConfigured` to `internal/sandbox/sandbox_test.go` — `--backend che` without CDE returns actionable error
- [x] T061 [P] Add `TestResolveBackend_UnknownBackend` to `internal/sandbox/sandbox_test.go` — unknown backend name returns error

### PodmanBackend tests

- [x] T062 [P] Add `TestPodmanCreate_HappyPath` to `internal/sandbox/sandbox_test.go` — verify volume create, container run, source copy, health check sequence
- [x] T063 [P] Add `TestPodmanCreate_AlreadyExists` to `internal/sandbox/sandbox_test.go` — verify error when named volume already exists
- [x] T064 [P] Add `TestPodmanCreate_VolumeCreateFails` to `internal/sandbox/sandbox_test.go` — verify error propagation with context
- [x] T065 [P] Add `TestPodmanCreate_WithDemoPorts` to `internal/sandbox/sandbox_test.go` — verify demo port `-p` mappings in container run args
- [x] T066 [P] Add `TestPodmanStart_PersistentResume` to `internal/sandbox/sandbox_test.go` — verify `podman start` for existing named volume
- [x] T067 [P] Add `TestPodmanStart_EphemeralFallback` to `internal/sandbox/sandbox_test.go` — verify Spec 028 behavior when no named volume
- [x] T068 [P] Add `TestPodmanStop_PersistentPreservesVolume` to `internal/sandbox/sandbox_test.go` — verify `podman stop` without `podman rm`
- [x] T069 [P] Add `TestPodmanStop_EphemeralRemoves` to `internal/sandbox/sandbox_test.go` — verify `podman stop` + `podman rm` (Spec 028)
- [x] T070 [P] Add `TestPodmanDestroy_HappyPath` to `internal/sandbox/sandbox_test.go` — verify `podman rm` + `podman volume rm`
- [x] T071 [P] Add `TestPodmanDestroy_NoWorkspace` to `internal/sandbox/sandbox_test.go` — verify idempotent (no error when nothing exists)
- [x] T072 [P] Add `TestPodmanDestroy_RunningWorkspace` to `internal/sandbox/sandbox_test.go` — verify stop then destroy sequence
- [x] T073 [P] Add `TestPodmanStatus_PersistentRunning` to `internal/sandbox/sandbox_test.go` — verify `WorkspaceStatus` fields for running persistent workspace
- [x] T074 [P] Add `TestPodmanStatus_PersistentStopped` to `internal/sandbox/sandbox_test.go` — verify `WorkspaceStatus` fields for stopped persistent workspace

### CheBackend tests

- [x] T075 [P] Add `TestCheCreate_WithChectl` to `internal/sandbox/sandbox_test.go` — verify `chectl workspace:create` command construction
- [x] T076 [P] Add `TestCheCreate_WithRestAPI` to `internal/sandbox/sandbox_test.go` — verify REST API `POST /api/workspace/devfile` with injected HTTPClient
- [x] T077 [P] Add `TestCheCreate_NoDevfile` to `internal/sandbox/sandbox_test.go` — verify error with `--backend podman` suggestion
- [x] T078 [P] Add `TestCheCreate_Unreachable` to `internal/sandbox/sandbox_test.go` — verify error with URL/auth hint
- [x] T079 [P] Add `TestCheStart_WithChectl` to `internal/sandbox/sandbox_test.go` — verify `chectl workspace:start` command
- [x] T080 [P] Add `TestCheStop_WithChectl` to `internal/sandbox/sandbox_test.go` — verify `chectl workspace:stop` command
- [x] T081 [P] Add `TestCheDestroy_WithChectl` to `internal/sandbox/sandbox_test.go` — verify `chectl workspace:delete --yes` command
- [x] T082 [P] Add `TestCheStatus_Running` to `internal/sandbox/sandbox_test.go` — verify `WorkspaceStatus` from chectl JSON output
- [x] T083 [P] Add `TestCheAttach_EndpointURL` to `internal/sandbox/sandbox_test.go` — verify Che endpoint URL construction for `opencode attach`

### Workspace helper tests

- [x] T084 [P] Add `TestProjectName_Simple` to `internal/sandbox/sandbox_test.go` — verify basename extraction and lowercase
- [x] T085 [P] Add `TestProjectName_SpecialChars` to `internal/sandbox/sandbox_test.go` — verify sanitization of special characters to hyphens
- [x] T086 [P] Add `TestProjectName_Empty` to `internal/sandbox/sandbox_test.go` — verify fallback to "default"
- [x] T087 [P] Add `TestLoadConfig_HappyPath` to `internal/sandbox/sandbox_test.go` — verify YAML parsing with all fields
- [x] T088 [P] Add `TestLoadConfig_Missing` to `internal/sandbox/sandbox_test.go` — verify default config returned when file missing
- [x] T089 [P] Add `TestLoadConfig_EnvOverride` to `internal/sandbox/sandbox_test.go` — verify env var overrides config file values
- [x] T090 [P] Add `TestFormatWorkspaceStatus_Podman` to `internal/sandbox/sandbox_test.go` — verify output format for Podman persistent workspace
- [x] T091 [P] Add `TestFormatWorkspaceStatus_Che` to `internal/sandbox/sandbox_test.go` — verify output format for CDE workspace
- [x] T092 [P] Add `TestFormatWorkspaceStatus_WithDemoEndpoints` to `internal/sandbox/sandbox_test.go` — verify demo endpoint lines in output

### Backward compatibility tests

- [x] T093 [P] Add `TestStart_EphemeralMode` to `internal/sandbox/sandbox_test.go` — verify `uf sandbox start` without prior `create` uses ephemeral mode (Spec 028 behavior)
- [x] T094 [P] Add `TestStop_EphemeralMode` to `internal/sandbox/sandbox_test.go` — verify `uf sandbox stop` in ephemeral mode removes container (Spec 028 behavior)
- [x] T095 [P] Add `TestAttach_Unchanged` to `internal/sandbox/sandbox_test.go` — verify `uf sandbox attach` works with both persistent and ephemeral
- [x] T096 [P] Add `TestExtract_Unchanged` to `internal/sandbox/sandbox_test.go` — verify `uf sandbox extract` works with both persistent and ephemeral
- [x] T097 [P] Add `TestStatus_EphemeralFallback` to `internal/sandbox/sandbox_test.go` — verify `uf sandbox status` shows Spec 028 format for ephemeral

### Git sync tests

- [x] T098 [P] [US3] Add `TestSetupGitSync_PodmanBackend` to `internal/sandbox/sandbox_test.go` — verify git remote configuration inside container
- [x] T099 [P] [US3] Add `TestCheckGitSync_Clean` to `internal/sandbox/sandbox_test.go` — verify nil return when workspace is clean and up-to-date
- [x] T100 [P] [US3] Add `TestCheckGitSync_Diverged` to `internal/sandbox/sandbox_test.go` — verify error with conflict details
- [x] T101 [P] [US3] Add `TestExtract_PersistentSuggestsGitPull` to `internal/sandbox/sandbox_test.go` — verify suggestion message for persistent workspaces with push access

### Full suite verification

- [x] T102 Run `go test -race -count=1 ./internal/sandbox/...` — verify all new and existing tests pass
- [x] T103 Run `go test -race -count=1 -coverprofile=coverage.out ./internal/sandbox/...` — verify coverage ≥ 80% overall; review per-file coverage against targets in `contracts/testing-strategy.md`
- [x] T104_a Run `go test -race -count=1 ./...` — verify no regressions across the entire project

**Checkpoint**: All 37+ new tests pass, all existing Spec 028 tests pass, coverage targets met (≥ 80% overall for `internal/sandbox/`)

---

## Phase 9: Documentation + Verification

**Purpose**: Update documentation and run final validation

- [x] T104 [P] Update `AGENTS.md` Recent Changes section — add `029-sandbox-cde-lifecycle` entry with summary of changes (new files, updated files, user story count, task count)
- [x] T105 [P] Update `AGENTS.md` Project Structure — add `backend.go`, `podman.go`, `che.go`, `workspace.go` entries under `internal/sandbox/`
- [x] T106 [P] Update `AGENTS.md` Active Technologies — add `gopkg.in/yaml.v3` for sandbox config parsing if not already listed for this package
- [x] T107 Run `go vet ./...` and `golangci-lint run` — verify zero findings after all changes
- [x] T108 Run `go build ./cmd/unbound-force/...` — verify binary builds successfully
- [x] T109 Run `go test -race -count=1 ./...` — final full-suite regression check
- [x] T110 Validate `quickstart.md` — verify documented commands match implemented CLI flags and output format

**Checkpoint**: All documentation updated, all tests pass, binary builds, lint clean — feature complete

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 green baseline — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Phase 2 (Backend interface + PodmanBackend)
- **US4 (Phase 4)**: Depends on Phase 2 (Backend interface + ResolveBackend)
- **US2 (Phase 5)**: Depends on Phase 3 (Create) and Phase 4 (CheBackend) for demo endpoint discovery
- **US3 (Phase 6)**: Depends on Phase 3 (Create) for git setup during workspace provisioning
- **Cobra Commands (Phase 7)**: Depends on Phase 3 (Create/Destroy) and Phase 4 (CheBackend) for full backend support
- **Tests (Phase 8)**: Depends on Phases 2-7 (all implementation complete)
- **Documentation (Phase 9)**: Depends on Phase 8 (all tests pass)

### User Story Dependencies

- **US1 (P1)**: Can start after Phase 2 — no dependencies on other stories
- **US4 (P1)**: Can start after Phase 2 — independent of US1 (different backend)
- **US2 (P1)**: Depends on US1 (Create) and US4 (CheBackend) for demo endpoint integration
- **US3 (P2)**: Depends on US1 (Create) for git setup during provisioning

### Parallel Opportunities

- **Phase 2**: T004-T008 can run in parallel (different structs/files, no dependencies)
- **Phase 3 + Phase 4**: US1 and US4 can proceed in parallel after Phase 2 (different files: `podman.go` vs. `che.go`)
- **Phase 8**: All test tasks marked [P] can run in parallel (all write to `sandbox_test.go` but are independent test functions)

### Within Each Phase

- Types/structs before functions that use them
- Interface definition before implementations
- Backend implementations before CLI wiring
- Implementation before tests
- Tests before documentation

<!-- spec-review: passed -->
