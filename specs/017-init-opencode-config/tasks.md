# Tasks: Init OpenCode Config

**Input**: Design documents from `specs/017-init-opencode-config/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add injectable file I/O to scaffold Options
and import `encoding/json` for JSON manipulation.

- [x] T001 Add `ReadFile`, `WriteFile`, and `DryRun` fields to the `Options` struct in `internal/scaffold/scaffold.go`: `ReadFile func(string) ([]byte, error)`, `WriteFile func(string, []byte, os.FileMode) error`, `DryRun bool`
- [x] T002 Default `ReadFile` to `os.ReadFile` and `WriteFile` to `os.WriteFile` in the `Run()` function alongside existing defaults for `LookPath`, `ExecCmd`, and `Stdout` in `internal/scaffold/scaffold.go`
- [x] T003 Default `ReadFile` to `os.ReadFile` and `WriteFile` to `os.WriteFile` in the `initSubTools()` nil-guard block (alongside existing `Stdout` default) in `internal/scaffold/scaffold.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Implement `configureOpencodeJSON()` in scaffold -- the core function that all user stories depend on.

**CRITICAL**: No user story work can begin until this function exists.

- [x] T004 Implement `configureOpencodeJSON(opts *Options) []subToolResult` in `internal/scaffold/scaffold.go` with the following behavior: if `opts.DryRun` return "dry-run" result immediately; read `opencode.json` from `opts.TargetDir` via `opts.ReadFile`; if read error is not "not found" return "error" result; parse as `map[string]json.RawMessage`; if malformed JSON return "error" result; add `mcp.dewey` entry when `opts.LookPath("dewey")` succeeds; check for legacy `mcpServers.dewey` and treat as already configured; add `opencode-swarm-plugin` to `plugin` array when `.hive/` directory exists; write back via `opts.WriteFile` with `json.MarshalIndent` (2-space indent + trailing newline); return action values per data-model.md (created, configured, already configured, overwritten, skipped, error, failed). Update the `subToolResult` struct comment to include all action values.
- [x] T005 Add idempotent guards to `configureOpencodeJSON()`: skip `mcp.dewey` if already present in `mcp` OR `mcpServers` (unless `opts.Force` is true), skip plugin if array already contains `"opencode-swarm-plugin"`, return "already configured" when nothing to add, return "error" when file is malformed JSON (with detail "malformed JSON"), return "error" when ReadFile fails with non-"not found" error (with detail describing the read failure), return "skipped" (with detail "nothing to configure") when neither dewey nor `.hive/` is available in `internal/scaffold/scaffold.go`
- [x] T006 Add force-overwrite logic to `configureOpencodeJSON()`: when `opts.Force` is true, overwrite `mcp.dewey` entry even if it exists, return "overwritten" as action in `internal/scaffold/scaffold.go`
- [x] T007 Call `configureOpencodeJSON()` from `initSubTools()` after the Dewey init/index block (or in place of it when `.dewey/` already exists), append results to the sub-tool results slice in `internal/scaffold/scaffold.go`

- [x] T007a Update `printSummary()` in `internal/scaffold/scaffold.go` to render new action values: `"created"`, `"configured"`, `"already configured"`, `"overwritten"` → `✓`; `"skipped"`, `"dry-run"` → `—`; `"error"`, `"failed"` → `✗`
- [x] T007b Update `runUnboundInit()` in `internal/setup/setup.go` to forward `opts.ReadFile`, `opts.WriteFile`, and `opts.DryRun` to `scaffold.Options` alongside the existing `LookPath` and `ExecCmd` forwarding

**Checkpoint**: `configureOpencodeJSON()` exists, is called from `initSubTools()`, renders correctly in summary, and injection chain is complete. Ready for user story testing.

---

## Phase 3: User Story 1 -- Fresh Repo Init (Priority: P1)

**Goal**: `uf init` creates `opencode.json` with Dewey MCP and Swarm plugin entries when tools are available.

**Independent Test**: Run `uf init` in temp dir with dewey + hive available → verify `opencode.json` created correctly.

### Tests for User Story 1

- [x] T008 [P] [US1] Write `TestConfigureOpencodeJSON_Create` in `internal/scaffold/scaffold_test.go`: no `opencode.json`, dewey in LookPath, `.hive/` exists → verify result action is `"created"`, file contains `$schema`, `mcp.dewey` entry (type local, command array, enabled true), and `plugin` array with `opencode-swarm-plugin`. Parse JSON and assert individual fields, not raw string comparison.
- [x] T009 [P] [US1] Write `TestConfigureOpencodeJSON_DeweyOnly` in `internal/scaffold/scaffold_test.go`: no `opencode.json`, dewey in LookPath, no `.hive/` → verify result action is `"created"`, file has `mcp.dewey` only, no `plugin` key
- [x] T010 [P] [US1] Write `TestConfigureOpencodeJSON_HiveOnly` in `internal/scaffold/scaffold_test.go`: no `opencode.json`, no dewey, `.hive/` exists → verify result action is `"created"`, file has `plugin` array only, no `mcp` key
- [x] T011 [P] [US1] Write `TestConfigureOpencodeJSON_Neither` in `internal/scaffold/scaffold_test.go`: no `opencode.json`, no dewey, no `.hive/` → verify no file created, result action is `"skipped"`, detail is `"nothing to configure"`

**Checkpoint**: US1 tests pass. Fresh repo init creates correct `opencode.json`.

---

## Phase 4: User Story 2 -- Idempotent Re-run (Priority: P1)

**Goal**: Running `uf init` again adds missing entries without disturbing existing config.

**Independent Test**: Create `opencode.json` with custom MCP servers, run init, verify custom entries preserved.

### Tests for User Story 2

- [x] T012 [P] [US2] Write `TestConfigureOpencodeJSON_Idempotent` in `internal/scaffold/scaffold_test.go`: `opencode.json` already has both `mcp.dewey` and `plugin` → verify file byte-identical to input, result action is `"already configured"`
- [x] T012a [P] [US2] Write `TestConfigureOpencodeJSON_IdempotentWithOtherPlugins` in `internal/scaffold/scaffold_test.go`: `opencode.json` has `plugin: ["other-plugin", "opencode-swarm-plugin"]` and `mcp.dewey` → verify file unchanged, both plugins preserved, action is `"already configured"`
- [x] T013 [P] [US2] Write `TestConfigureOpencodeJSON_AddMissing` in `internal/scaffold/scaffold_test.go`: `opencode.json` has `plugin` but no `mcp.dewey`, dewey available → verify result action is `"configured"`, `mcp.dewey` added, existing `plugin` array entries preserved exactly
- [x] T014 [P] [US2] Write `TestConfigureOpencodeJSON_PreserveCustom` in `internal/scaffold/scaffold_test.go`: `opencode.json` has `mcp.my-custom-server` → verify custom server preserved, `mcp.dewey` added alongside
- [x] T014a [P] [US2] Write `TestConfigureOpencodeJSON_LegacyMcpServers` in `internal/scaffold/scaffold_test.go`: `opencode.json` has `mcpServers.dewey` (legacy key), dewey available → verify result action is `"already configured"`, no duplicate `mcp.dewey` added
- [x] T015 [P] [US2] Write `TestConfigureOpencodeJSON_Malformed` in `internal/scaffold/scaffold_test.go`: `opencode.json` contains `{invalid json` → verify result action is `"error"`, detail is `"malformed JSON"`, file not modified
- [x] T015a [P] [US2] Write `TestConfigureOpencodeJSON_ReadPermissionDenied` in `internal/scaffold/scaffold_test.go`: inject `ReadFile` that returns a permission error → verify result action is `"error"`, detail describes the read failure
- [x] T016 [P] [US2] Write `TestConfigureOpencodeJSON_WriteFail` in `internal/scaffold/scaffold_test.go`: inject a `WriteFile` that returns an error → verify action is `"failed"`, non-fatal
- [x] T016a [P] [US2] Write `TestConfigureOpencodeJSON_ByteIdentical` in `internal/scaffold/scaffold_test.go`: run `configureOpencodeJSON()` twice with same inputs → verify output bytes are identical (FR-016)

**Checkpoint**: US2 tests pass. Idempotent behavior verified.

---

## Phase 5: User Story 3 -- Force Overwrite (Priority: P2)

**Goal**: `uf init --force` overwrites stale `mcp.dewey` entries.

**Independent Test**: Create stale `mcp.dewey`, run with Force=true, verify replaced.

### Tests for User Story 3

- [x] T017 [P] [US3] Write `TestConfigureOpencodeJSON_Force` in `internal/scaffold/scaffold_test.go`: `opencode.json` has stale `mcp.dewey` with `--include-hidden` flag AND `plugin: ["opencode-swarm-plugin"]`, Force=true, dewey available, `.hive/` exists → verify `mcp.dewey` overwritten with correct command, action is `"overwritten"`, plugin array is NOT duplicated (still exactly one entry)
- [x] T018 [P] [US3] Write `TestConfigureOpencodeJSON_ForceCorrect` in `internal/scaffold/scaffold_test.go`: `opencode.json` has correct `mcp.dewey`, Force=true → verify overwritten (same content), action is `"overwritten"`
- [x] T018a [P] [US3] Write `TestConfigureOpencodeJSON_DryRun` in `internal/scaffold/scaffold_test.go`: DryRun=true → verify no file written, result action is `"dry-run"`

**Checkpoint**: US3 tests pass. Force overwrite verified.

---

## Phase 6: User Story 4 -- Setup Delegates to Init (Priority: P2)

**Goal**: Remove `configureOpencodeJSON()` from setup.go, renumber steps from 16 to 15.

**Independent Test**: Run setup, verify no direct opencode.json step, step count is 15.

### Implementation for User Story 4

- [x] T019 [US4] Remove the `configureOpencodeJSON()` function from `internal/setup/setup.go`
- [x] T020 [US4] Remove step 10 (`opencode.json` configuration) from the `Run()` function in `internal/setup/setup.go` -- remove both the normal path and the `swarmResult.action == "already installed"` path that call `configureOpencodeJSON()`
- [x] T021 [US4] Remove the skipped step result for `opencode.json` in the `!nodeAvailable` else branch in `internal/setup/setup.go`
- [x] T022 [US4] Renumber all step progress messages from `[N/16]` to `[N/15]` in `internal/setup/setup.go` (steps after the removed step shift down by 1)
- [x] T023 [US4] Update `TestSetupRun_OpencodeJsonManipulation` in `internal/setup/setup_test.go`: verify setup does NOT write `opencode.json` (remove assertions about plugin entry being added by setup)
- [x] T024 [US4] Update `TestSetupRun_NoOpencodeJson` in `internal/setup/setup_test.go`: verify setup does NOT create `opencode.json`
- [x] T025 [US4] Update `TestSetupRun_MalformedOpencodeJson` in `internal/setup/setup_test.go`: remove or update assertions about opencode.json handling since setup no longer touches it
- [x] T026 [US4] Update any step-count assertions in `internal/setup/setup_test.go` that reference 16 steps → 15 steps

**Checkpoint**: US4 complete. Setup no longer touches opencode.json.

---

## Phase 7: Doctor MCP Fix (Cross-Cutting)

**Goal**: Fix `uf doctor` to check both `"mcp"` and `"mcpServers"` keys and handle array-style command fields.

- [x] T027 Update `checkMCPConfig()` in `internal/doctor/checks.go` to check for `"mcp"` key first, then fall back to `"mcpServers"` key. If both exist, prefer `"mcp"`
- [x] T028 Update the MCP server command binary extraction in `checkMCPConfig()` in `internal/doctor/checks.go` to handle both string-style (`"command": "dewey"`) and array-style (`"command": ["dewey", "serve", "--vault", "."]`) command fields. For array-style, extract the first element as the binary name
- [x] T029 [P] Write `TestCheckMCPConfig_McpKey` in `internal/doctor/doctor_test.go`: `opencode.json` uses `"mcp"` key with dewey server → verify dewey binary check passes
- [x] T030 [P] Write `TestCheckMCPConfig_McpServersKey` in `internal/doctor/doctor_test.go`: `opencode.json` uses legacy `"mcpServers"` key → verify fallback works
- [x] T031 [P] Write `TestCheckMCPConfig_ArrayCommand` in `internal/doctor/doctor_test.go`: command is `["dewey", "serve", "--vault", "."]` → verify extracts `dewey` as binary
- [x] T032 [P] Write `TestCheckMCPConfig_StringCommand` in `internal/doctor/doctor_test.go`: command is `"dewey"` (string) → verify backward compat

**Checkpoint**: Doctor correctly detects MCP servers regardless of key format or command style.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, README updates, full test suite verification.

- [x] T032a Write `TestInitSubTools_OpencodeJSON` integration test in `internal/scaffold/scaffold_test.go`: call `initSubTools()` with `LookPath` returning success for `"dewey"`, `.hive/` directory present → verify `opencode.json` exists with expected content (tests the T007 wiring)
- [x] T032b [US4] Verify setup step numbering is contiguous `[1/15]` through `[15/15]` with no gaps or duplicates in `internal/setup/setup_test.go`
- [x] T033 Verify `README.md` scaffold file count is still accurate (no new files added by this spec)
- [x] T034 Run `go build ./...` to verify clean build
- [x] T035 Run `go test -race -count=1 ./...` to verify all tests pass including coverage ratchets
- [x] T036 Run `go vet ./...` to verify no vet warnings
- [x] T037 Update AGENTS.md "Recent Changes" section with this spec's summary

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 -- BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Phase 2
- **US2 (Phase 4)**: Depends on Phase 2 (independent of US1)
- **US3 (Phase 5)**: Depends on Phase 2 (independent of US1, US2)
- **US4 (Phase 6)**: Depends on Phase 2 (independent of US1-3 but should run after to verify no regressions)
- **Doctor Fix (Phase 7)**: Independent of US1-4 (different package)
- **Polish (Phase 8)**: Depends on all phases complete

### User Story Dependencies

- **US1 (P1)**: Independent -- tests fresh repo init
- **US2 (P1)**: Independent -- tests idempotent re-run
- **US3 (P2)**: Independent -- tests force overwrite
- **US4 (P2)**: Independent -- tests setup removal (different package)
- **Doctor Fix**: Independent -- tests doctor MCP check (different package)

### Parallel Opportunities

- T001, T002, T003 are sequential (same struct, same function)
- T008-T011 (US1 tests) can run in parallel
- T012-T016 (US2 tests) can run in parallel
- T017-T018 (US3 tests) can run in parallel
- T029-T032 (Doctor tests) can run in parallel
- US1-US3 test phases can run in parallel after Phase 2
- Phase 7 (Doctor) can run in parallel with US1-US4

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T007)
3. Complete Phase 3: US1 tests (T008-T011)
4. **STOP and VALIDATE**: `go test ./internal/scaffold/`
5. Fresh repo init works correctly

### Incremental Delivery

1. Setup + Foundational → Core function ready
2. US1 → Fresh init tested
3. US2 → Idempotent behavior tested
4. US3 → Force overwrite tested
5. US4 → Setup refactored
6. Doctor → MCP check fixed
7. Polish → Full suite passes, docs updated
