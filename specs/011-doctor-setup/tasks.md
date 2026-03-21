---
description: "Task list for Doctor and Setup Commands"
---

# Tasks: Doctor and Setup Commands

**Input**: Design documents from `/specs/011-doctor-setup/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Included per constitution Principle IV (Testability) and plan.md coverage strategy (90%+ target).

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to
- Include exact file paths in descriptions

## Path Conventions

- CLI entry point: `cmd/unbound/main.go`
- Doctor package: `internal/doctor/`
- Setup package: `internal/setup/`
- Existing reuse: `internal/orchestration/heroes.go`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create package structure and shared types

- [x] T001 Create `internal/doctor/` package directory
- [x] T002 [P] Create `internal/setup/` package directory
- [x] T003 Promote `charmbracelet/lipgloss` from indirect to direct dependency in `go.mod` by running `go get github.com/charmbracelet/lipgloss`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared data types, environment detection, and platform guard used by ALL user stories

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 Define Severity, ManagerKind, ManagerInfo, DetectedEnvironment, ToolProvenance, CheckResult, CheckGroup, Summary, Report types in `internal/doctor/models.go` per data-model.md. Include JSON tags with snake_case field names. Severity implements `MarshalJSON`/`UnmarshalJSON` for lowercase string serialization. ToolProvenance is internal-only (no JSON tags). Optional-absent items use Pass severity with populated InstallHint as the convention for informational rendering.

- [x] T005 Define Options struct in `internal/doctor/doctor.go` with fields: TargetDir (string), Format (string), Stdout (io.Writer), LookPath (func), ExecCmd (func), EvalSymlinks (func), Getenv (func), ReadFile (func). Add defaults method that fills zero-value fields with production implementations (`exec.LookPath`, `filepath.EvalSymlinks`, `os.Getenv`, `os.ReadFile`).

- [x] T006 Implement platform guard in `internal/doctor/doctor.go`: check `runtime.GOOS` at the start of `Run()`. If `"windows"`, return an error with message "Platform not supported: doctor and setup require macOS or Linux" per FR-037.

- [x] T007 Implement `DetectEnvironment()` in `internal/doctor/environ.go` per research.md R1. Detect goenv, pyenv, nvm, fnm, mise, bun, Homebrew using priority-ordered path pattern matching and env var checks. Use injected LookPath, EvalSymlinks, Getenv from Options. Return DetectedEnvironment with all detected managers and `runtime.GOOS + "/" + runtime.GOARCH` as Platform.

- [x] T008 Implement `DetectProvenance()` in `internal/doctor/environ.go` per research.md R1. Given a binary path, determine which ManagerKind installed it using the 10-step priority chain (goenv > pyenv > nvm > fnm > mise > bun > Homebrew > direct > system > unknown). Use injected EvalSymlinks for Homebrew symlink resolution.

- [x] T009 Implement `installHint()` helper in `internal/doctor/environ.go` that returns a manager-appropriate install command for a given tool name and detected environment. Maps tool names to install commands per detected manager (e.g., `go` + goenv -> `goenv install 1.24.3`, `go` + Homebrew -> `brew install go`).

- [x] T010 Write tests for DetectEnvironment and DetectProvenance in `internal/doctor/doctor_test.go`. Test with injected LookPath/EvalSymlinks/Getenv returning goenv paths, nvm paths, Homebrew symlinks, system paths, and unknown paths. Cover all 10 manager kinds. Test installHint with various manager/tool combinations. Test platform guard returns error on `"windows"`.

**Checkpoint**: Foundation ready -- environment detection working, all types defined, platform guard tested

---

## Phase 3: User Story 1 - Diagnose Environment Health (Priority: P1)

**Goal**: `unbound doctor` checks all 7 groups and reports pass/warn/fail with environment-aware install hints

**Independent Test**: Run `unbound doctor` in a temp dir with injected dependencies returning various tool states; verify correct severity and install hints for each

### Tests for User Story 1

- [x] T011 [P] [US1] Write test `TestCheckCoreTools` in `internal/doctor/doctor_test.go`. Inject LookPath that returns go (found via goenv), opencode (not found), gaze (found via Homebrew), mxf (not found), graphthulhu (not found), node (found via nvm), gh (not found), swarm (found). Verify: go=Pass with "via goenv", opencode=Fail with install hint, gaze=Pass, mxf=Warn, graphthulhu=Pass(info with hint), node=Pass with "via nvm", gh=Pass(info with hint), swarm=Pass.

- [x] T012 [P] [US1] Write test `TestCheckCoreTools_UnparseableGoVersion` in `internal/doctor/doctor_test.go`. Inject ExecCmd returning `go version devel go1.25-abcdef` for `go version`. Verify check passes with a warning that version could not be verified.

- [x] T013 [P] [US1] Write test `TestCheckScaffoldedFiles` in `internal/doctor/doctor_test.go`. Use t.TempDir() with and without `.opencode/agents/`, `.opencode/command/`, `.specify/`, `AGENTS.md`. Verify correct pass/fail per directory.

- [x] T014 [P] [US1] Write test `TestCheckHeroAvailability` in `internal/doctor/doctor_test.go`. Create agent files in t.TempDir(), inject LookPath for gaze/mxf. Verify all 5 heroes detected correctly using orchestration.DetectHeroes.

- [x] T015 [P] [US1] Write test `TestCheckMCPConfig` in `internal/doctor/doctor_test.go`. Create opencode.json with MCP server entries in t.TempDir(). Inject LookPath for server binaries. Verify pass for found binaries, warn for missing.

- [x] T016 [P] [US1] Write test `TestCheckMCPConfig_MalformedJSON` in `internal/doctor/doctor_test.go`. Create `opencode.json` with invalid JSON (`{invalid`) in t.TempDir(). Verify warn severity with "could not be parsed" message.

- [x] T017 [P] [US1] Write test `TestCheckAgentIntegrity` in `internal/doctor/doctor_test.go`. Create agent .md files with valid and invalid YAML frontmatter in t.TempDir(). Verify pass for valid, warn for missing description.

- [x] T018 [P] [US1] Write test `TestCheckSkillIntegrity` in `internal/doctor/doctor_test.go`. Create SKILL.md with valid/invalid name/description in t.TempDir(). Verify name regex validation and directory name matching.

- [x] T019 [US1] Write test `TestDoctorRun` in `internal/doctor/doctor_test.go`. Full pipeline test: inject all dependencies, call Run(), verify Report has 7 groups in correct order, Summary counts are accurate.

- [x] T020 [P] [US1] Write test `TestExitCode` (table-driven) in `internal/doctor/doctor_test.go`. Three cases: (a) all-pass Report -> no error, (b) warnings-only Report -> no error, (c) any-fail Report -> error returned.

- [x] T021 [P] [US1] Write test `TestDoctorRun_NonGitDir` in `internal/doctor/doctor_test.go`. Run doctor in t.TempDir() (not a git repo). Verify all checks execute, `.hive/` check produces appropriate message.

- [x] T022 [P] [US1] Write test `TestDoctorRun_DirFlag` in `internal/doctor/doctor_test.go`. Set TargetDir to a specific t.TempDir() path, verify checks run against that directory not cwd.

### Implementation for User Story 1

- [x] T023 [US1] Implement `checkDetectedEnvironment()` in `internal/doctor/checks.go`. Build CheckGroup "Detected Environment" listing all detected managers from DetectedEnvironment with pass severity, showing kind, path, and managed categories per FR-000a.

- [x] T024 [US1] Implement `checkCoreTools()` in `internal/doctor/checks.go`. Check 8 binaries (go, opencode, gaze, mxf, graphthulhu, node, gh, swarm) per FR-001/002/003. Use DetectProvenance for version+manager display. Parse `go version` output for >= 1.24 (FR-004), gracefully handle unparseable version. Parse `node --version` for >= 18 (FR-005). Classify as required/recommended/optional per spec. Include install hints from installHint(). Include install_url for non-trivial installs per FR-016.

- [x] T025 [US1] Implement `checkScaffoldedFiles()` in `internal/doctor/checks.go` per FR-006. Check `.opencode/agents/` (count .md files), `.opencode/command/` (count .md files), `.specify/` existence, `AGENTS.md` existence. Convention packs check `.opencode/unbound/packs/`.

- [x] T026 [US1] Implement `checkHeroAvailability()` in `internal/doctor/checks.go` per FR-007. Call `orchestration.DetectHeroes()` with agent dir and LookPath. Map HeroStatus to CheckResults with human-readable names (Muti-Mind, Cobalt-Crush, Gaze, The Divisor, Mx F). Show detection method (agent file vs binary) and Divisor persona count.

- [x] T027 [US1] Implement `checkMCPConfig()` in `internal/doctor/checks.go` per FR-011. Parse `opencode.json` if present. Extract MCP server entries, check command[0] binary exists via LookPath. Handle malformed JSON gracefully (warn, not fail).

- [x] T028 [US1] Implement `checkAgentIntegrity()` in `internal/doctor/checks.go` per FR-013. Walk `.opencode/agents/*.md`, parse YAML frontmatter per R6, validate `description` is present and non-empty. Report count of validated agents.

- [x] T029 [US1] Implement `checkSkillIntegrity()` in `internal/doctor/checks.go` per FR-014. Walk `.opencode/skill/*/SKILL.md` and `.opencode/skills/*/SKILL.md`, parse YAML frontmatter, validate `name` and `description` exist, name matches directory name, name matches `^[a-z0-9]+(-[a-z0-9]+)*$`.

- [x] T030 [US1] Implement `Run()` in `internal/doctor/doctor.go`. Call defaults() on Options, check platform guard, call DetectEnvironment(), run all 7 check group functions, compute Summary (total/passed/warned/failed), return Report. Determine exit error based on any Fail severity.

- [x] T031 [US1] Register `newDoctorCmd()` in `cmd/unbound/main.go`. Add Cobra command with `--format` (text/json, default text) and `--dir` (default cwd) flags. Validate `--format` accepts only "text" or "json" (return error for invalid values). RunE calls runDoctor(doctorParams) which calls doctor.Run() and formats output. Return error for exit code 1 when failures exist.

**Checkpoint**: `unbound doctor` shows all 7 check groups with correct pass/warn/fail. Tests pass for all scenarios including edge cases.

---

## Phase 4: User Story 2 - Validate Swarm Plugin (Priority: P1)

**Goal**: Swarm Plugin check group includes `swarm doctor` embedded output, .hive/ check, and plugin config check

**Independent Test**: Run doctor with injected ExecCmd returning canned `swarm doctor` output; verify embedded output and additional Swarm checks

### Tests for User Story 2

- [x] T032 [P] [US2] Write test `TestCheckSwarmPlugin_Installed` in `internal/doctor/doctor_test.go`. Inject LookPath finding swarm, ExecCmd returning success with 4-line swarm doctor output. Create .hive/ dir and opencode.json with plugin entry. Verify: swarm=Pass, embed contains swarm doctor output, .hive/=Pass, plugin config=Pass.

- [x] T033 [P] [US2] Write test `TestCheckSwarmPlugin_NotInstalled` in `internal/doctor/doctor_test.go`. Inject LookPath not finding swarm. Verify: swarm=Fail with install hint `npm install -g opencode-swarm-plugin@latest`.

- [x] T034 [P] [US2] Write test `TestCheckSwarmPlugin_DoctorFails` in `internal/doctor/doctor_test.go`. Inject ExecCmd returning non-zero exit with stderr output. Verify: Warn severity, stderr embedded, suggest `unbound setup`.

- [x] T035 [P] [US2] Write test `TestCheckSwarmPlugin_Timeout` in `internal/doctor/doctor_test.go`. Inject ExecCmd that blocks longer than timeout. Verify: Warn with "swarm doctor timed out".

- [x] T036 [P] [US2] Write test `TestCheckSwarmPlugin_MissingPluginConfig` in `internal/doctor/doctor_test.go`. Swarm binary found but `opencode-swarm-plugin` not in opencode.json plugin array. Verify: Warn with `Fix: unbound setup`.

### Implementation for User Story 2

- [x] T037 [US2] Implement `checkSwarmPlugin()` in `internal/doctor/checks.go` per FR-008/009/010/012. Check swarm binary via LookPath. If found: execute `swarm doctor` with 10s context timeout (FR-009), capture output, embed in CheckGroup.Embed. Check `.hive/` existence (FR-010). Parse opencode.json for `opencode-swarm-plugin` in plugin array (FR-012). Handle timeout (context.DeadlineExceeded), non-zero exit, and missing binary cases.

- [x] T038 [US2] Wire `checkSwarmPlugin()` into `Run()` in `internal/doctor/doctor.go` as the 3rd check group (after Detected Environment, Core Tools).

**Checkpoint**: Swarm Plugin section shows embedded `swarm doctor` output, .hive/ status, and plugin config status. Tests pass.

---

## Phase 5: User Story 3 - Automated Environment Setup (Priority: P2)

**Goal**: `unbound setup` installs all missing tools through detected managers, configures Swarm, runs `unbound init`. Supports `--dry-run` and curl safety confirmation.

**Independent Test**: Run setup with injected ExecCmd/LookPath simulating missing tools; verify correct install commands issued in correct order

### Tests for User Story 3

- [x] T039 [P] [US3] Write test `TestSetupRun_AllMissing` in `internal/setup/setup_test.go`. Inject LookPath finding brew and node but not opencode/gaze/swarm. Inject ExecCmd recording all commands. Verify install order: opencode (brew), gaze (brew), swarm (npm), swarm setup, opencode.json config, swarm init, unbound init.

- [x] T040 [P] [US3] Write test `TestSetupRun_AllPresent` in `internal/setup/setup_test.go`. Inject LookPath finding everything. Create opencode.json with plugin entry, .hive/, .opencode/. Verify: all steps skipped, no ExecCmd calls, "already configured" summary.

- [x] T041 [P] [US3] Write test `TestSetupRun_NoNodeJS` in `internal/setup/setup_test.go`. Inject LookPath finding brew but not node or swarm. Verify: OpenCode and Gaze installed, Swarm steps skipped with warning, unbound init still runs.

- [x] T042 [P] [US3] Write test `TestSetupRun_NpmFails` in `internal/setup/setup_test.go`. Inject ExecCmd returning error for npm install. Verify: npm error reported, swarm setup/init skipped, unbound init still runs.

- [x] T043 [P] [US3] Write test `TestSetupRun_NvmDetected` in `internal/setup/setup_test.go`. Inject LookPath finding nvm-managed node (`/.nvm/versions/...`). Inject Getenv returning NVM_DIR. Verify: Swarm installed via npm from nvm-managed node.

- [x] T044 [P] [US3] Write test `TestSetupRun_NvmInstallNode` in `internal/setup/setup_test.go`. Inject LookPath NOT finding `node` but Getenv returning `NVM_DIR=/home/user/.nvm`. Verify ExecCmd called with bash shell invocation to source nvm and install Node.js (`bash -c "source $NVM_DIR/nvm.sh && nvm install 22"`).

- [x] T045 [P] [US3] Write test `TestSetupRun_BunDetected` in `internal/setup/setup_test.go`. Inject LookPath finding `bun` but not `npm`. Verify ExecCmd called with `bun add -g opencode-swarm-plugin@latest`.

- [x] T046 [P] [US3] Write test `TestSetupRun_OpencodeJsonManipulation` in `internal/setup/setup_test.go`. Create opencode.json with existing MCP servers and no plugin key in t.TempDir(). Run setup. Verify: opencode.json now has plugin array with `opencode-swarm-plugin`, MCP servers preserved, valid JSON.

- [x] T047 [P] [US3] Write test `TestSetupRun_NoOpencodeJson` in `internal/setup/setup_test.go`. Run setup in t.TempDir() with no opencode.json. Verify a minimal opencode.json is created with `$schema` and plugin array.

- [x] T048 [P] [US3] Write test `TestSetupRun_MalformedOpencodeJson` in `internal/setup/setup_test.go`. Create opencode.json with invalid JSON in t.TempDir(). Run setup. Verify: setup refuses to modify, warns about malformed JSON, continues to other steps.

- [x] T049 [P] [US3] Write test `TestSetupRun_Idempotent` in `internal/setup/setup_test.go`. Run setup twice. Verify: second run makes no ExecCmd calls, opencode.json unchanged (no duplicate plugin entries).

- [x] T050 [P] [US3] Write test `TestSetupRun_SwarmInitFails` in `internal/setup/setup_test.go`. Inject ExecCmd returning error for `swarm init`. Verify error reported, setup continues to `unbound init`.

- [x] T051 [P] [US3] Write test `TestSetupRun_DryRun` in `internal/setup/setup_test.go`. Set DryRun=true in Options. Inject LookPath not finding opencode/gaze/swarm. Verify: no ExecCmd calls made, output shows "Would install:" messages for each missing tool.

- [x] T052 [P] [US3] Write test `TestSetupRun_CurlSafety` in `internal/setup/setup_test.go`. Inject LookPath not finding opencode or brew (curl fallback path). Set YesFlag=false, IsTTY=false. Verify: curl install is skipped with warning about requiring `--yes` flag.

### Implementation for User Story 3

- [x] T053 [US3] Define setup Options struct in `internal/setup/setup.go` with fields: TargetDir, DryRun (bool), YesFlag (bool), IsTTY (func() bool), Stdout, Stderr, LookPath, ExecCmd, EvalSymlinks, Getenv, ReadFile, WriteFile. Add defaults method. Add platform guard (same as doctor).

- [x] T054 [US3] Implement `installOpenCode()` in `internal/setup/setup.go` per FR-022/FR-036. Check `opencode` in PATH via LookPath. If missing: detect Homebrew, use `brew install anomalyco/tap/opencode`. If no Homebrew: check YesFlag or IsTTY for curl confirmation before executing `curl -fsSL https://opencode.ai/install | bash`. If neither confirmed, skip with warning and manual instructions. In dry-run mode, print "Would install" instead.

- [x] T055 [US3] Implement `installGaze()` in `internal/setup/setup.go` per FR-023. Check `gaze` in PATH. If missing and Homebrew available: `brew install unbound-force/tap/gaze`. If no Homebrew: warn with GitHub releases URL.

- [x] T056 [US3] Implement `ensureNodeJS()` in `internal/setup/setup.go` per FR-024. Check `node` in PATH and version >= 18. If missing: detect nvm (NVM_DIR) -> invoke via `bash -c "source $NVM_DIR/nvm.sh && nvm install 22"`, detect fnm -> `fnm install 22`, detect brew -> `brew install node`. If nvm invocation fails, print manual command and skip. Return whether Node.js is available for Swarm steps.

- [x] T057 [US3] Implement `installSwarmPlugin()` in `internal/setup/setup.go` per FR-025. Check `swarm` in PATH. If missing: detect bun -> `bun add -g opencode-swarm-plugin@latest`, else `npm install -g opencode-swarm-plugin@latest`. Run `swarm setup` after install (FR-026).

- [x] T058 [US3] Implement `configureOpencodeJSON()` in `internal/setup/setup.go` per FR-027/027a/028. Read opencode.json (or create minimal with $schema per FR-028). Unmarshal to map[string]json.RawMessage per R3. Add `opencode-swarm-plugin` to plugin array if missing. Write atomically (temp file + rename per FR-027a). Handle malformed JSON (refuse to modify, warn).

- [x] T059 [US3] Implement `initializeHive()` in `internal/setup/setup.go` per FR-029. Check `.hive/` existence. If missing: exec `swarm init`.

- [x] T060 [US3] Implement `runUnboundInit()` in `internal/setup/setup.go` per FR-033. Check `.opencode/` existence. If missing: call scaffold.Run() directly (same binary, no subprocess needed). Document the import dependency on `internal/scaffold/`.

- [x] T061 [US3] Implement `Run()` in `internal/setup/setup.go` per FR-021/030/032/034/035. Call defaults(), check platform guard, DetectEnvironment (reuse from doctor package), check DryRun flag. Execute install steps in order per R7 dependency chain. If DryRun: print what would be done without executing. Continue on independent failures (FR-032). Print completion summary (FR-034).

- [x] T062 [US3] Register `newSetupCmd()` in `cmd/unbound/main.go`. Add Cobra command with `--dir` (default cwd), `--dry-run`, and `--yes` flags. RunE calls runSetup(setupParams) which calls setup.Run().

**Checkpoint**: `unbound setup` installs all tools through detected managers, supports --dry-run and curl safety. Idempotent on re-run. Tests pass.

---

## Phase 6: User Story 4 - Colored Terminal Output (Priority: P2)

**Goal**: Doctor text output uses colored indicators (green/yellow/red/gray) with NO_COLOR fallback

**Independent Test**: Capture output with and without color support; verify correct symbols and fallback indicators

### Tests for User Story 4

- [x] T063 [P] [US4] Write test `TestFormatText_WithColors` in `internal/doctor/doctor_test.go`. Create Report with pass/warn/fail/info results. Call FormatText with a Writer that has color support. Verify output contains checkmark, exclamation, cross, circle symbols and install hints on the line below failures.

- [x] T064 [P] [US4] Write test `TestFormatText_NoColors` in `internal/doctor/doctor_test.go`. Call FormatText with a Writer to a buffer (pipe, no TTY). Verify output uses `[PASS]`, `[WARN]`, `[FAIL]`, `[INFO]` plain text indicators.

- [x] T065 [P] [US4] Write test `TestFormatText_SwarmDoctorEmbed` in `internal/doctor/doctor_test.go`. Create CheckGroup with Embed field. Verify embedded output appears between separator lines.

- [x] T066 [P] [US4] Write test `TestFormatText_InstallHints` in `internal/doctor/doctor_test.go`. Create CheckResult with InstallHint and InstallURL. Verify both appear in output on indented lines below the check result.

- [x] T067 [P] [US4] Write test `TestFormatSetupText` in `internal/setup/setup_test.go`. Verify: "already installed" messages for skipped steps, error messages for failed steps, completion summary format, detected environment line.

### Implementation for User Story 4

- [x] T068 [US4] Implement `FormatText()` in `internal/doctor/format.go` per FR-019 and contracts/cli-schema.md. Create lipgloss renderer from Stdout writer per R2. Define styles: pass (green), warn (yellow), fail (red), info (gray). Detect color profile: if no color support, use `[PASS]`/`[WARN]`/`[FAIL]`/`[INFO]` fallback. Render header "Unbound Force Doctor" with separator. For each group: render group name, each result with aligned columns (name, message, detail), install hints on indented line below. For Swarm group: render Embed between separator lines. Render Summary line at bottom.

- [x] T069 [US4] Implement `FormatSetupText()` in `internal/setup/setup.go`. Render "Unbound Force Setup" header, detected environment summary, per-step results (installed/skipped/failed with symbols), dry-run "Would install" messages, completion message with next steps.

- [x] T070 [US4] Wire FormatText into `cmd/unbound/main.go` doctorCmd RunE. After doctor.Run(), check format flag: "text" -> FormatText(report, stdout), "json" -> FormatJSON. Write to stdout.

**Checkpoint**: `unbound doctor` shows beautifully formatted colored output. Plain text fallback works in pipes. Install hints appear below failures. Tests pass.

---

## Phase 7: User Story 5 - Machine-Readable JSON Output (Priority: P3)

**Goal**: `--format=json` produces valid, parseable JSON matching data-model.md schema

**Independent Test**: Run `unbound doctor --format=json`, pipe through `jq .`, verify structure matches data-model.md

### Tests for User Story 5

- [x] T071 [P] [US5] Write test `TestFormatJSON` in `internal/doctor/doctor_test.go`. Create Report with all group types, pass/warn/fail results, and Swarm embed. Call FormatJSON. Verify: valid JSON, snake_case field names, severity as lowercase strings, all fields present per data-model.md, install_hint and install_url present on failed checks.

- [x] T072 [P] [US5] Write test `TestFormatJSON_EmptyReport` in `internal/doctor/doctor_test.go`. Empty Report with zero groups. Verify: valid JSON, groups is empty array (not null), summary all zeros.

### Implementation for User Story 5

- [x] T073 [US5] Implement `FormatJSON()` in `internal/doctor/format.go` per FR-017 and data-model.md. Use `json.MarshalIndent(report, "", "  ")`. Ensure all JSON tags on model types produce snake_case. Write trailing newline. Verify Severity marshal produces `"pass"`, `"warn"`, `"fail"` strings.

**Checkpoint**: `unbound doctor --format=json` produces valid JSON. Can pipe to `jq` and parse with standard tools. Tests pass.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, coverage enforcement, validation, and cleanup

- [x] T074 [P] Update `AGENTS.md` Recent Changes section with 011-doctor-setup summary
- [x] T075 [P] Add coverage enforcement: add a CI step or test helper that runs `go test -coverprofile` on `internal/doctor/` and `internal/setup/` and fails if coverage drops below 90% per Constitution Principle IV coverage ratchet requirement
- [x] T076 Run `make check` (build, test, vet, lint) and fix any issues
- [x] T077 Run `unbound doctor` manually against the project itself (dogfooding) and verify output matches contracts/cli-schema.md expected format
- [x] T078 Run `unbound setup --dry-run` manually against the project and verify it shows correct "Would install" messages without executing anything
- [x] T079 Run `unbound setup` manually against the project (idempotent verify) and confirm no-op on already-configured environment

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 -- BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Phase 2 -- core doctor checks
- **US2 (Phase 4)**: Depends on Phase 2 -- can run parallel with US1 (different check group)
- **US3 (Phase 5)**: Depends on Phase 2 -- can run parallel with US1/US2 (different package)
- **US4 (Phase 6)**: Depends on US1 + US2 completion (needs check results to format)
- **US5 (Phase 7)**: Depends on US1 + US2 completion (needs Report struct populated)
- **Polish (Phase 8)**: Depends on all user stories

### User Story Dependencies

- **US1 (P1)**: Depends on Foundation only -- core checks
- **US2 (P1)**: Depends on Foundation only -- Swarm check group (can parallel with US1)
- **US3 (P2)**: Depends on Foundation only -- setup package (can parallel with US1/US2)
- **US4 (P2)**: Depends on US1 + US2 -- needs report data to format
- **US5 (P3)**: Depends on US1 + US2 -- needs report data to serialize

### Within Each User Story

- Tests written first (TDD) -- verify they FAIL before implementation
- Models/types before logic
- Check functions before Run() orchestration
- CLI registration last

### Parallel Opportunities

- T001/T002: package dirs created in parallel
- T011-T022: all US1 tests in parallel (different test functions)
- T032-T036: all US2 tests in parallel
- T039-T052: all US3 tests in parallel
- T063-T067: all US4 tests in parallel
- T071-T072: all US5 tests in parallel
- US1/US2/US3: can run in parallel after Foundation (different files/packages)
- T074/T075: polish tasks in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all US1 tests together:
Task: "TestCheckCoreTools in internal/doctor/doctor_test.go"
Task: "TestCheckCoreTools_UnparseableGoVersion in internal/doctor/doctor_test.go"
Task: "TestCheckScaffoldedFiles in internal/doctor/doctor_test.go"
Task: "TestCheckHeroAvailability in internal/doctor/doctor_test.go"
Task: "TestCheckMCPConfig in internal/doctor/doctor_test.go"
Task: "TestCheckMCPConfig_MalformedJSON in internal/doctor/doctor_test.go"
Task: "TestCheckAgentIntegrity in internal/doctor/doctor_test.go"
Task: "TestCheckSkillIntegrity in internal/doctor/doctor_test.go"
Task: "TestExitCode in internal/doctor/doctor_test.go"
Task: "TestDoctorRun_NonGitDir in internal/doctor/doctor_test.go"
Task: "TestDoctorRun_DirFlag in internal/doctor/doctor_test.go"

# Then implement check functions in parallel:
Task: "checkCoreTools() in internal/doctor/checks.go"
Task: "checkScaffoldedFiles() in internal/doctor/checks.go"
Task: "checkHeroAvailability() in internal/doctor/checks.go"
Task: "checkMCPConfig() in internal/doctor/checks.go"
Task: "checkAgentIntegrity() in internal/doctor/checks.go"
Task: "checkSkillIntegrity() in internal/doctor/checks.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundation (T004-T010)
3. Complete Phase 3: US1 - Doctor Core Checks (T011-T031)
4. **STOP and VALIDATE**: `unbound doctor` runs all 7 groups
5. Can ship with basic output before formatting

### Incremental Delivery

1. Setup + Foundation -> types and detection working
2. Add US1 -> Doctor checks all groups (MVP)
3. Add US2 -> Swarm doctor embedded
4. Add US4 -> Beautiful colored output
5. Add US5 -> JSON output for CI
6. Add US3 -> Full automated setup with --dry-run
7. Polish -> Dogfooding, coverage enforcement, docs

### Parallel Team Strategy

With multiple developers:
1. Team completes Setup + Foundation together
2. Once Foundation is done:
   - Developer A: US1 (doctor checks)
   - Developer B: US2 (Swarm integration)
   - Developer C: US3 (setup command)
3. After US1+US2: Developer A takes US4 (formatting)
4. After US1+US2: Developer B takes US5 (JSON)

---

## Notes

- [P] tasks = different files or test functions, no dependencies
- [US?] label maps task to specific user story for traceability
- Each user story independently completable and testable
- Tests written first per TDD -- verify they FAIL before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- `internal/doctor/` is the larger package (~6 files)
- `internal/setup/` reuses `doctor.DetectEnvironment()`
- All subprocess/filesystem calls injected for testability
- Windows explicitly unsupported (FR-037) -- platform guard exits early
- `--dry-run` flag on setup prints actions without executing (FR-035)
- `curl | bash` installs require `--yes` or interactive confirmation (FR-036)
