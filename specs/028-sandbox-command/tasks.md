# Tasks: Sandbox Command

**Input**: Design documents from `/specs/028-sandbox-command/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Verify Baseline)

**Purpose**: Confirm the existing codebase is green before adding new code.

- [x] T001 Run `go test -race -count=1 ./...` and verify all existing tests pass (FR-020)
- [x] T002 Run `go build ./...` and verify the binary builds cleanly
- [x] T003 Verify branch is `028-sandbox-command` and spec artifacts are committed

**Checkpoint**: Baseline is green — new code can be added safely.

---

## Phase 2: Foundation — Types and Configuration (No External Dependencies)

**Purpose**: Create the pure-logic files (`config.go`, `detect.go`) that have no external dependencies and are consumed by all user stories.

- [x] T004 [P] Create `internal/sandbox/config.go` — define package declaration, imports, and constants: `ContainerName`, `DefaultImage`, `DefaultMemory`, `DefaultCPUs`, `DefaultServerPort`, `HealthTimeout`, `ModeIsolated`, `ModeDirect` per data-model.md
- [x] T005 [P] Create `internal/sandbox/detect.go` — define `PlatformConfig` struct with `OS`, `Arch`, `SELinux` fields per data-model.md
- [x] T006 [P] Create `internal/sandbox/sandbox.go` — define `Options` struct with all injectable fields (`ProjectDir`, `Mode`, `Detach`, `Image`, `Memory`, `CPUs`, `Stdout`, `Stderr`, `LookPath`, `ExecCmd`, `ExecInteractive`, `Getenv`, `ReadFile`, `HTTPGet`) and `ContainerStatus` struct per data-model.md
- [x] T007 Create `internal/sandbox/sandbox.go` — define `PatchSummary` struct with `CommitCount`, `FilesChanged`, `Insertions`, `Deletions`, `Patch`, `StatOutput` fields per data-model.md
- [x] T008 In `internal/sandbox/sandbox.go` — implement `defaults()` method on `Options` that fills zero-value fields with production implementations (`exec.LookPath`, `exec.Command(...).CombinedOutput`, `os.Getenv`, `os.ReadFile`, `http.Get` wrapper, `os.Stdout`, `os.Stderr`)
- [x] T009 In `internal/sandbox/config.go` — implement `DefaultConfig(opts Options) Options` that resolves image from flag → `UF_SANDBOX_IMAGE` env var → `DefaultImage` constant, and memory/cpus from flag → constant defaults (FR-017, FR-018)
- [x] T010 In `internal/sandbox/config.go` — implement `forwardedEnvVars(opts Options) []string` that returns `-e` flag pairs for `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `GEMINI_API_KEY`, `OPENROUTER_API_KEY`, and `OLLAMA_HOST=host.containers.internal:11434` per research.md R7
- [x] T011 In `internal/sandbox/config.go` — implement `buildVolumeMounts(opts Options, platform PlatformConfig) []string` that constructs `-v` flags with `:ro` for isolated mode, read-write for direct mode, and `:Z` suffix when `platform.SELinux` is true (FR-008, FR-009)
- [x] T012 In `internal/sandbox/config.go` — implement `buildRunArgs(opts Options, platform PlatformConfig) []string` that assembles the complete `podman run` argument list: `-d`, `--name`, `--hostname`, `-p 4096:4096`, volume mounts, env vars, `--memory`, `--cpus`, and image name per research.md R4
- [x] T013 In `internal/sandbox/detect.go` — implement `DetectPlatform(opts Options) PlatformConfig` using `runtime.GOOS`, `runtime.GOARCH`, and injectable `ExecCmd` for `getenforce` on Linux (skip SELinux on macOS) per research.md R3

**Checkpoint**: Foundation types and pure configuration logic are complete. All functions are pure or use only injected dependencies. No container operations yet.

---

## Phase 3: US1 — Start an Isolated Agent Session (Priority: P1)

**Goal**: `uf sandbox start` launches a container, waits for health, and attaches the TUI.

**Independent Test**: Run `uf sandbox start` with Podman and Ollama available. Verify the container starts, health check passes, and terminal attaches.

### Implementation

- [x] T014 In `internal/sandbox/sandbox.go` — implement `isContainerRunning(opts Options) (bool, error)` using `podman inspect --format '{{.State.Running}}' uf-sandbox` via injected `ExecCmd` per contracts/sandbox-api.md
- [x] T015 In `internal/sandbox/sandbox.go` — implement `waitForHealth(opts *Options, timeout time.Duration) error` with exponential backoff polling (500ms initial, doubling to 5s max, 60s total timeout) using injected `HTTPGet` to `http://localhost:4096` per research.md R5 and FR-005
- [x] T016 [US1] In `internal/sandbox/sandbox.go` — implement `Start(opts Options) error` prerequisite checks: verify `podman` in PATH via `LookPath` (FR-001), verify `opencode` in PATH when `Detach` is false, check for already-running container via `isContainerRunning` (FR-016)
- [x] T017 [US1] In `internal/sandbox/sandbox.go` — implement `Start(opts Options) error` container launch: call `DefaultConfig` to resolve image/memory/cpus, call `DetectPlatform`, call `buildRunArgs`, execute `podman run` via `ExecCmd` (FR-003, FR-004)
- [x] T018 [US1] In `internal/sandbox/sandbox.go` — implement `Start(opts Options) error` post-launch: call `waitForHealth`, then either print server URL and return (when `Detach` is true, FR-007) or call `ExecInteractive("opencode", "attach", "http://localhost:4096")` (FR-006)
- [x] T019 [US1] In `internal/sandbox/sandbox.go` — implement `Start()` Ollama warning: when `Getenv("OLLAMA_HOST")` is empty and Ollama is not detected, print a warning to `Stderr` but continue (per edge cases)
- [x] T020 [US1] In `internal/sandbox/sandbox.go` — implement `Start()` dead container cleanup: when `podman inspect` finds a stopped container, run `podman rm uf-sandbox` before starting a new one (per edge cases)
- [x] T021 [US1] In `internal/sandbox/sandbox.go` — implement `Start()` image pull: before `podman run`, execute `podman image exists <image>` and if it fails, run `podman pull <image>` with progress output to `Stderr` (FR-003)

**Checkpoint**: `uf sandbox start` works end-to-end with injected dependencies. All US1 acceptance scenarios are covered by the implementation logic.

---

## Phase 4: US2 — Extract Changes from Container (Priority: P1)

**Goal**: `uf sandbox extract` generates a patch, shows a review, and applies on confirmation.

**Independent Test**: Start a sandbox, make a change inside, run `uf sandbox extract`. Verify patch is presented and applied on confirmation.

### Implementation

- [x] T022 [US2] In `internal/sandbox/sandbox.go` — implement `Extract(opts Options) error` precondition checks: verify container is running via `isContainerRunning`, detect direct mode and return early with "changes are already on the host filesystem" message per contracts/sandbox-api.md
- [x] T023 [US2] In `internal/sandbox/sandbox.go` — implement `Extract()` patch generation: execute `podman exec uf-sandbox git -C /workspace log --oneline origin/HEAD..HEAD` to count commits, then `podman exec uf-sandbox git -C /workspace format-patch origin/HEAD..HEAD --stdout` to generate patch (FR-010)
- [x] T024 [US2] In `internal/sandbox/sandbox.go` — implement `Extract()` patch summary: parse commit count, run `git apply --stat` on the patch content to get `FilesChanged`, `Insertions`, `Deletions`, display `PatchSummary` to `Stdout` (FR-011)
- [x] T025 [US2] In `internal/sandbox/sandbox.go` — implement `Extract()` confirmation and apply: prompt user via `Stdin` (or skip if `--yes`), on confirmation run `git am` on host with patch content, on decline exit cleanly (FR-012)
- [x] T026 [US2] In `internal/sandbox/sandbox.go` — implement `Extract()` error handling: handle "no changes to extract" (empty commit list), handle `git am` merge conflict with actionable message suggesting `git am --abort` per contracts/sandbox-api.md

**Checkpoint**: `uf sandbox extract` works end-to-end. Round-trip workflow (start → work → extract) is complete.

---

## Phase 5: US3 — Manage Sandbox Lifecycle (Priority: P2)

**Goal**: `uf sandbox attach`, `uf sandbox stop`, and `uf sandbox status` manage the running sandbox.

**Independent Test**: Start a sandbox with `--detach`, then use `attach`, `status`, and `stop` subcommands.

### Implementation

- [x] T027 [US3] In `internal/sandbox/sandbox.go` — implement `Attach(opts Options) error`: verify `opencode` in PATH via `LookPath`, verify container is running via `isContainerRunning`, call `ExecInteractive("opencode", "attach", "http://localhost:4096")` (FR-013)
- [x] T028 [US3] In `internal/sandbox/sandbox.go` — implement `Stop(opts Options) error`: call `podman stop uf-sandbox` then `podman rm uf-sandbox` via `ExecCmd`, return nil if no container exists (idempotent) (FR-014)
- [x] T029 [US3] In `internal/sandbox/sandbox.go` — implement `Status(opts Options) (ContainerStatus, error)`: execute `podman inspect uf-sandbox --format json` via `ExecCmd`, parse JSON into `ContainerStatus` struct, determine mode from volume mount flags, format and print output to `Stdout` (FR-015)
- [x] T030 [US3] In `internal/sandbox/sandbox.go` — implement `Status()` output formatting: print container name, ID, image, mode, project dir, server URL, and uptime when running; print "No sandbox running." when no container exists per contracts/sandbox-api.md

**Checkpoint**: All lifecycle management functions work. US3 acceptance scenarios are covered.

---

## Phase 6: US4 — Platform-Aware Container Configuration (Priority: P2)

**Goal**: SELinux detection and platform-specific volume mount flags work correctly.

**Independent Test**: Run on macOS and Fedora, verify correct architecture and SELinux flags.

### Implementation

- [x] T031 [US4] In `internal/sandbox/detect.go` — implement SELinux detection: on Linux, read `/etc/selinux/config` via `ReadFile` to check for `SELINUX=enforcing`, then verify with `getenforce` command via `ExecCmd`; on macOS, always return `SELinux: false` per research.md R3
- [x] T032 [US4] In `internal/sandbox/config.go` — verify `buildVolumeMounts()` appends `:Z` to all volume mount paths when `platform.SELinux` is true, and omits it otherwise (already implemented in T011, this task verifies the integration)
- [x] T033 [US4] In `internal/sandbox/config.go` — verify `buildRunArgs()` includes `--platform linux/arm64` or `--platform linux/amd64` based on `platform.Arch` for correct image variant selection

**Checkpoint**: Platform detection is complete. Volume mounts are correct on all target platforms.

---

## Phase 7: Cobra Commands

**Purpose**: Wire the `internal/sandbox` package to the CLI via Cobra commands.

- [x] T034 Create `cmd/unbound-force/sandbox.go` — implement `newSandboxCmd() *cobra.Command` returning the parent `uf sandbox` command with `Use`, `Short`, and `Long` descriptions per contracts/cobra-commands.md
- [x] T035 In `cmd/unbound-force/sandbox.go` — define `sandboxStartParams` struct and implement `runSandboxStart(p sandboxStartParams) error` that constructs `sandbox.Options` from params and calls `sandbox.Start()` per contracts/cobra-commands.md
- [x] T036 In `cmd/unbound-force/sandbox.go` — implement `newSandboxStartCmd() *cobra.Command` with `--mode`, `--detach`, `--image`, `--memory`, `--cpus` flags, delegating to `runSandboxStart()` per contracts/cobra-commands.md
- [x] T037 In `cmd/unbound-force/sandbox.go` — define `sandboxStopParams` struct and implement `runSandboxStop(p sandboxStopParams) error` that calls `sandbox.Stop()`
- [x] T038 In `cmd/unbound-force/sandbox.go` — implement `newSandboxStopCmd() *cobra.Command` delegating to `runSandboxStop()`
- [x] T039 In `cmd/unbound-force/sandbox.go` — define `sandboxAttachParams` struct and implement `runSandboxAttach(p sandboxAttachParams) error` that calls `sandbox.Attach()`
- [x] T040 In `cmd/unbound-force/sandbox.go` — implement `newSandboxAttachCmd() *cobra.Command` delegating to `runSandboxAttach()`
- [x] T041 In `cmd/unbound-force/sandbox.go` — define `sandboxExtractParams` struct (with `yes bool`, `stdin io.Reader`) and implement `runSandboxExtract(p sandboxExtractParams) error` that calls `sandbox.Extract()`
- [x] T042 In `cmd/unbound-force/sandbox.go` — implement `newSandboxExtractCmd() *cobra.Command` with `--yes` flag, delegating to `runSandboxExtract()` per contracts/cobra-commands.md
- [x] T043 In `cmd/unbound-force/sandbox.go` — define `sandboxStatusParams` struct and implement `runSandboxStatus(p sandboxStatusParams) error` that calls `sandbox.Status()` and formats output
- [x] T044 In `cmd/unbound-force/sandbox.go` — implement `newSandboxStatusCmd() *cobra.Command` delegating to `runSandboxStatus()`
- [x] T045 In `cmd/unbound-force/sandbox.go` — register all subcommands: `newSandboxCmd().AddCommand(newSandboxStartCmd(), newSandboxStopCmd(), newSandboxAttachCmd(), newSandboxExtractCmd(), newSandboxStatusCmd())`
- [x] T046 In `cmd/unbound-force/main.go` — add `root.AddCommand(newSandboxCmd())` after the existing `newSetupCmd()` registration (line 31)

**Checkpoint**: All 5 subcommands are wired to the CLI. `go build ./...` succeeds. `uf sandbox --help` shows the command tree.

---

## Phase 8: Tests

**Purpose**: Comprehensive test coverage for all sandbox functions per contracts/testing-strategy.md.

### detect.go tests

- [x] T047 [P] In `internal/sandbox/sandbox_test.go` — implement `TestDetectPlatform_MacOSArm64`: inject `ExecCmd` that is never called (macOS skips getenforce), verify `PlatformConfig{OS: "darwin", Arch: "arm64", SELinux: false}`
- [x] T048 [P] In `internal/sandbox/sandbox_test.go` — implement `TestDetectPlatform_FedoraSELinux`: inject `ExecCmd` returning "Enforcing" for `getenforce` and `ReadFile` returning `SELINUX=enforcing`, verify `SELinux: true`
- [x] T049 [P] In `internal/sandbox/sandbox_test.go` — implement `TestDetectPlatform_FedoraNoSELinux`: inject `ExecCmd` returning "Disabled" for `getenforce`, verify `SELinux: false`

### config.go tests

- [x] T050 [P] In `internal/sandbox/sandbox_test.go` — implement `TestBuildRunArgs_Isolated`: verify args include `-v <project>:/workspace:ro`, `--name uf-sandbox`, `-p 4096:4096`, `--memory 8g`, `--cpus 4`
- [x] T051 [P] In `internal/sandbox/sandbox_test.go` — implement `TestBuildRunArgs_Direct`: verify args include `-v <project>:/workspace` (no `:ro`), read-write mount
- [x] T052 [P] In `internal/sandbox/sandbox_test.go` — implement `TestBuildRunArgs_SELinux`: verify args include `:Z` suffix on volume mounts when `platform.SELinux` is true
- [x] T053 [P] In `internal/sandbox/sandbox_test.go` — implement `TestBuildRunArgs_CustomImage`: verify custom image name appears in args when `opts.Image` is set
- [x] T054 [P] In `internal/sandbox/sandbox_test.go` — implement `TestDefaultConfig_ImagePrecedence`: verify flag → env var → constant precedence for image resolution (FR-017)
- [x] T055 [P] In `internal/sandbox/sandbox_test.go` — implement `TestDefaultConfig_MemoryAndCPUsPrecedence`: verify flag value overrides constant default for both memory and cpus fields
- [x] T056a [P] In `internal/sandbox/sandbox_test.go` — implement `TestForwardedEnvVars`: verify all expected env vars are included and `OLLAMA_HOST` is set to `host.containers.internal:11434`

### sandbox.go Start() tests

- [x] T056 [P] In `internal/sandbox/sandbox_test.go` — implement `TestStart_PodmanMissing`: inject `LookPath` returning error for "podman", verify error message includes install hint (FR-001)
- [x] T057 [P] In `internal/sandbox/sandbox_test.go` — implement `TestStart_AlreadyRunning`: inject `ExecCmd` returning "true" for `podman inspect`, verify error message suggests `attach` or `stop` (FR-016)
- [x] T058 [P] In `internal/sandbox/sandbox_test.go` — implement `TestStart_DetachMode`: inject successful `ExecCmd` for all podman calls and `HTTPGet` returning 200, verify `ExecInteractive` is NOT called and server URL is printed (FR-007)
- [x] T059 [P] In `internal/sandbox/sandbox_test.go` — implement `TestStart_IsolatedMount`: verify `buildRunArgs` produces `:ro` volume mount flag (FR-008)
- [x] T060 [P] In `internal/sandbox/sandbox_test.go` — implement `TestStart_DirectMount`: verify `buildRunArgs` produces read-write volume mount (no `:ro`) (FR-009)
- [x] T061 [P] In `internal/sandbox/sandbox_test.go` — implement `TestStart_HealthTimeout`: inject `HTTPGet` always returning error, verify error after timeout (FR-005)
- [x] T062 [P] In `internal/sandbox/sandbox_test.go` — implement `TestStart_DeadContainerCleanup`: inject `ExecCmd` returning stopped container on inspect, verify `podman rm` is called before `podman run`

### sandbox.go Stop() tests

- [x] T063 [P] In `internal/sandbox/sandbox_test.go` — implement `TestStop_RunningContainer`: inject `ExecCmd` succeeding for `podman stop` and `podman rm`, verify both are called (FR-014)
- [x] T064 [P] In `internal/sandbox/sandbox_test.go` — implement `TestStop_NoContainer`: inject `ExecCmd` returning error for `podman stop` (no such container), verify nil return and "no sandbox to stop" message

### sandbox.go Attach() tests

- [x] T065 [P] In `internal/sandbox/sandbox_test.go` — implement `TestAttach_NoContainer`: inject `isContainerRunning` returning false, verify error message suggests `uf sandbox start` (FR-013)
- [x] T066 [P] In `internal/sandbox/sandbox_test.go` — implement `TestAttach_OpenCodeMissing`: inject `LookPath` returning error for "opencode", verify error message includes install hint

### sandbox.go Extract() tests

- [x] T067 [P] In `internal/sandbox/sandbox_test.go` — implement `TestExtract_NoChanges`: inject `ExecCmd` returning empty commit list, verify "no changes to extract" message (FR-010)
- [x] T068 [P] In `internal/sandbox/sandbox_test.go` — implement `TestExtract_UserDeclines`: inject `ExecCmd` returning commits and patch, simulate "n" on stdin, verify patch is NOT applied (FR-012)
- [x] T069 [P] In `internal/sandbox/sandbox_test.go` — implement `TestExtract_DirectModeWarning`: set `opts.Mode` to "direct", verify early return with "changes are already on the host filesystem" message
- [x] T070 [P] In `internal/sandbox/sandbox_test.go` — implement `TestExtract_NoContainer`: inject `isContainerRunning` returning false, verify "no sandbox running" error

### sandbox.go Status() tests

- [x] T071 [P] In `internal/sandbox/sandbox_test.go` — implement `TestStatus_Running`: inject `ExecCmd` returning valid `podman inspect` JSON, verify `ContainerStatus` fields are correctly parsed (FR-015)
- [x] T072 [P] In `internal/sandbox/sandbox_test.go` — implement `TestStatus_Stopped`: inject `ExecCmd` returning inspect JSON with `Running: false` and exit code, verify `ExitCode` is set
- [x] T073 [P] In `internal/sandbox/sandbox_test.go` — implement `TestStatus_NoContainer`: inject `ExecCmd` returning error (no such container), verify `Running: false` and no error

### Health check tests

- [x] T074 [P] In `internal/sandbox/sandbox_test.go` — implement `TestWaitForHealth_ImmediateSuccess`: inject `HTTPGet` returning 200 on first call, verify immediate return
- [x] T075 [P] In `internal/sandbox/sandbox_test.go` — implement `TestWaitForHealth_DelayedSuccess`: inject `HTTPGet` returning error 3 times then 200, verify success after retries
- [x] T076 [P] In `internal/sandbox/sandbox_test.go` — implement `TestWaitForHealth_Timeout`: inject `HTTPGet` always returning error, verify timeout error with correct duration message

### isContainerRunning tests

- [x] T077 [P] In `internal/sandbox/sandbox_test.go` — implement `TestIsContainerRunning_Running`: inject `ExecCmd` returning "true" for inspect, verify `true`
- [x] T078 [P] In `internal/sandbox/sandbox_test.go` — implement `TestIsContainerRunning_NotRunning`: inject `ExecCmd` returning "false" for inspect, verify `false`
- [x] T079 [P] In `internal/sandbox/sandbox_test.go` — implement `TestIsContainerRunning_NoContainer`: inject `ExecCmd` returning error (no such container), verify `false` and no error

**Checkpoint**: All 33 test cases pass. Run `go test -race -count=1 ./internal/sandbox/` to verify. Coverage target: ≥ 80% for `internal/sandbox/`.

---

## Phase 9: Documentation and Verification

**Purpose**: Update documentation, verify the full build, and run the complete test suite.

- [x] T080 Update `AGENTS.md` — add `internal/sandbox/` to the Project Structure tree under `internal/`
- [x] T081 Update `AGENTS.md` — add `cmd/unbound-force/sandbox.go` to the Project Structure tree under `cmd/unbound-force/`
- [x] T082 Update `AGENTS.md` — add Spec 028 entry to the Recent Changes section with summary of all user stories and task count
- [x] T083 Run `go build ./...` and verify the binary builds with the new sandbox command
- [x] T084 Run `go test -race -count=1 ./...` and verify ALL tests pass (FR-020)
- [x] T085 Run `go vet ./...` and verify no vet warnings
- [x] T086 Verify `uf sandbox --help` shows all 5 subcommands with correct descriptions
- [x] T087 Run quickstart.md validation — verify documented commands match implemented CLI flags and behavior

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundation)**: Depends on Phase 1 — creates types consumed by all later phases
- **Phase 3 (US1 Start)**: Depends on Phase 2 — uses `Options`, `buildRunArgs`, `DetectPlatform`
- **Phase 4 (US2 Extract)**: Depends on Phase 2 + T014 (`isContainerRunning`) from Phase 3
- **Phase 5 (US3 Lifecycle)**: Depends on Phase 2 + T014 (`isContainerRunning`) from Phase 3
- **Phase 6 (US4 Platform)**: Depends on Phase 2 — extends `DetectPlatform` and verifies `buildVolumeMounts`
- **Phase 7 (Cobra)**: Depends on Phases 3-6 — wires all functions to CLI
- **Phase 8 (Tests)**: Depends on Phases 2-6 — tests all functions. Can start per-function as each is completed.
- **Phase 9 (Docs)**: Depends on all previous phases

### User Story Dependencies

- **US1 (Start)**: Can start after Phase 2. No dependencies on other stories.
- **US2 (Extract)**: Can start after Phase 2 + `isContainerRunning` from US1. Independent of US3/US4.
- **US3 (Lifecycle)**: Can start after Phase 2 + `isContainerRunning` from US1. Independent of US2/US4.
- **US4 (Platform)**: Can start after Phase 2. Independent of US1/US2/US3 (extends `detect.go`).

### Parallel Opportunities

- **Phase 2**: T004, T005, T006, T007 can run in parallel (different files)
- **Phase 3-6**: US4 (Phase 6) can run in parallel with US1 (Phase 3) since they touch different files (`detect.go` vs `sandbox.go`)
- **Phase 7**: T034-T045 are sequential within `sandbox.go` but T046 (`main.go`) can run in parallel
- **Phase 8**: All test tasks marked [P] can run in parallel (single test file, but independent test functions)
- **Phase 9**: T080-T082 (docs) can run in parallel with T083-T087 (verification)

---

## Summary

| Phase | Tasks | Files |
|-------|-------|-------|
| 1. Setup | 3 | (verification only) |
| 2. Foundation | 10 | `config.go`, `detect.go`, `sandbox.go` |
| 3. US1 Start | 8 | `sandbox.go` |
| 4. US2 Extract | 5 | `sandbox.go` |
| 5. US3 Lifecycle | 4 | `sandbox.go` |
| 6. US4 Platform | 3 | `detect.go`, `config.go` |
| 7. Cobra | 13 | `cmd/unbound-force/sandbox.go`, `main.go` |
| 8. Tests | 34 | `sandbox_test.go` |
| 9. Docs | 8 | `AGENTS.md`, verification |
| **Total** | **88** | |

<!-- spec-review: passed -->
