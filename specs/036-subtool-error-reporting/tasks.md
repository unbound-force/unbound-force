# Tasks: Sub-tool Error Reporting

**Input**: Design documents from `specs/036-subtool-error-reporting/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: Tests are included -- the spec requires
verifiable error output behavior and the constitution
mandates a coverage strategy.

**Organization**: Tasks are grouped by user story to
enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Foundational (Blocking Prerequisites)

**Purpose**: Extend result structs and test infrastructure.
MUST complete before user story implementation.

**CRITICAL**: No user story work can begin until this
phase is complete.

- [x] T001 [P] Add `err error` and `output []byte` fields to `subToolResult` struct in `internal/scaffold/scaffold.go` (currently at line 476). Update the GoDoc comment on lines 473-475 to document the new fields.
- [x] T002 [P] Add `output []byte` field to `stepResult` struct in `internal/setup/setup.go` (currently at line 133). Update the comment on line 134 to document the new field.
- [x] T003 [P] Add `truncateOutput` helper function in `internal/scaffold/scaffold.go` that takes `[]byte` and `maxLines int`, returns a `string` with the last `maxLines/2` lines (integer division) when output exceeds `maxLines` lines, prefixed with `... (N lines omitted)\n`. Empty input returns empty string. Trim trailing whitespace. Default `maxLines` is 20, showing the last 10 lines.
- [x] T004 [P] Extend `scaffoldCmdRecorder` in `internal/scaffold/scaffold_test.go` with an `outputs map[string]string` field. Update its `execCmd` method to return `[]byte(outputs[key])` alongside the error (mirroring the `cmdRecorder` pattern in `internal/setup/setup_test.go` which already has an `outputs` map).

**Checkpoint**: Result structs extended, test infrastructure ready -- user story implementation can now begin.

---

## Phase 2: User Story 1 - See Why a Sub-tool Failed During Init (Priority: P1) MVP

**Goal**: When a sub-tool invocation fails during
`uf init`, the failure message includes the sub-tool's
actual error output instead of a hardcoded summary.

**Independent Test**: Trigger a sub-tool failure during
`uf init` and verify the error output includes the
actual error message from the sub-tool.

### Tests for User Story 1

- [x] T005 [P] [US1] Add test `TestInitSubTools_DeweyInitFails_ShowsError` in `internal/scaffold/scaffold_test.go` that configures `scaffoldCmdRecorder` to return both output bytes (e.g., `[]byte("Error: vault not found\n")`) and an error for `dewey init`, runs `initDewey`, and asserts the `subToolResult.output` contains the error text and the rendered summary (via `printSummary` to a `bytes.Buffer`) includes "vault not found".
- [x] T006 [P] [US1] Add test `TestInitSubTools_SimpleToolFails_ShowsError` in `internal/scaffold/scaffold_test.go` that configures `scaffoldCmdRecorder` to return output bytes and an error for `specify init`, runs `initSimpleTool`, and asserts the result contains the actual error text.
- [x] T007 [P] [US1] Add test `TestTruncateOutput` in `internal/scaffold/scaffold_test.go` as a table-driven test covering: empty input, short output (< 20 lines), exactly 20 lines, 21 lines (boundary -- first case that triggers truncation), long output (50 lines, verify last 10 shown with omission prefix), output with no trailing newline.
- [x] T008 [P] [US1] Add test `TestInitSubTools_ExitCodeOnly` in `internal/scaffold/scaffold_test.go` that configures `scaffoldCmdRecorder` to return empty output bytes with a non-nil error, and asserts the rendered summary includes the error message (exit code) as fallback diagnostic (FR-004).
- [x] T009 [P] [US1] Add test `TestPrintSummary_SuccessUnchanged` in `internal/scaffold/scaffold_test.go` that creates a `subToolResult` with `action: "initialized"`, `err: nil`, `output: nil`, and asserts `printSummary` produces the same single-line output as before (no `Output:` or `Error:` lines), verifying FR-007 regression protection.

### Implementation for User Story 1

- [x] T010 [US1] Update `initDewey` function in `internal/scaffold/scaffold.go` to capture ExecCmd output: change all `_, initErr :=` and `_, idxErr :=` patterns to `out, initErr :=` and `out, idxErr :=`, and populate `output` and `err` fields on `subToolResult` for failure cases. Update `detail` to use `fmt.Sprintf("%s: %s", toolName, initErr)` instead of hardcoded strings. There are 3 ExecCmd calls in this function (dewey init, dewey index first-run, dewey index re-index).
- [x] T011 [US1] Update `initSimpleTool` function in `internal/scaffold/scaffold.go` to capture ExecCmd output: change `_, initErr :=` to `out, initErr :=`, populate `output` and `err` fields on the returned `*subToolResult`, and update `detail` to include the actual error. This function handles specify, openspec, replicator, and gaze init.
- [x] T012 [US1] Update `printSummary` function in `internal/scaffold/scaffold.go` to display `output` field for failed sub-tool results. After the main result line, if `sr.action` is `"failed"` or `"error"` and `sr.output` is non-empty, print the output using `truncateOutput(sr.output, 20)` with indentation matching the existing format (prefix each line with enough spaces to align under the detail text). Also print `sr.err` on a separate `Error:` line if non-nil (this is a new pattern for scaffold, modeled after setup.go's `printStepResult`).

**Checkpoint**: `uf init` sub-tool failures now show actual error messages. Run `go test -race -count=1 ./internal/scaffold/...` to verify.

---

## Phase 3: User Story 2 - See Why a Tool Installation Failed During Setup (Priority: P2)

**Goal**: When a tool installation fails during
`uf setup`, the failure message includes the package
manager's actual error output.

**Independent Test**: Trigger a setup installation
failure and verify the output includes the package
manager's error message.

### Tests for User Story 2

- [x] T013 [P] [US2] Add test `TestInstallViaBrew_ShowsOutput` in `internal/setup/setup_test.go` that configures `cmdRecorder` to return output bytes (e.g., `[]byte("Error: No available formula\n")`) and an error for `brew install`, calls `installViaBrew`, and asserts `stepResult.output` contains the error text and `printStepResult` renders the `Output:` line.
- [x] T014 [P] [US2] Add test `TestInstallViaGo_ShowsOutput` in `internal/setup/setup_test.go` that configures `cmdRecorder` to return output bytes with Go compiler errors and an error for `go install`, and asserts `stepResult.output` and rendered output include the compiler error text.
- [x] T015 [P] [US2] Add test `TestPrintStepResult_WithOutput` in `internal/setup/setup_test.go` that directly tests `printStepResult` with a `stepResult` that has `output` populated, verifying the `Output:` line appears in the rendered output for failures and does NOT appear for successful results (FR-007 regression protection for setup).

### Implementation for User Story 2

- [x] T016 [US2] Add `truncateOutput` helper function in `internal/setup/setup.go` (same logic as scaffold's version -- small function, not worth a shared package per research.md RQ-2 decision). If either copy is modified in the future, both MUST be updated together.
- [x] T017 [US2] Update all `_, err := opts.ExecCmd(...)` call sites in `internal/setup/setup.go` to `out, err := opts.ExecCmd(...)` and populate the `output` field on `stepResult` for failure cases. There are approximately 25 call sites across functions: `installViaBrew`, `installViaGo`, `installViaDnf`, `installViaRpm`, and per-tool install functions. Only populate `output` when `err != nil` -- successful calls can continue to discard output (except where output is already used, like `node --version` and `devpod provider list`).
- [x] T018 [US2] Update `printStepResult` function in `internal/setup/setup.go` to display `output` field. After the existing `Error:` line, if `r.output` is non-empty and `r.action` is `"failed"`, print an `Output:` line with the truncated output using `truncateOutput(r.output, 20)`, indented to align with the `Error:` prefix (21 spaces).

**Checkpoint**: `uf setup` installation failures now show actual error output. Run `go test -race -count=1 ./internal/setup/...` to verify.

---

## Phase 4: User Story 3 - Detailed Output on Demand (Priority: P3)

**Goal**: Provide a mechanism for users to access full
command output from failed sub-tool invocations.

**Independent Test**: Trigger a failure with multi-line
output and verify the full output is accessible.

### Implementation for User Story 3

- [x] T019 [US3] Add test `TestTruncateOutput_Setup` in `internal/setup/setup_test.go` as a table-driven test covering: empty input, short output, exactly 20 lines (no truncation), 21 lines (boundary -- first truncation trigger), long output (50 lines showing truncation with omission count), verifying the truncation helper works correctly in the setup package context.
- [x] T020 [US3] Verify truncation integration by adding a test case to `TestInitSubTools_DeweyInitFails_ShowsError` (scaffold) and `TestInstallViaBrew_ShowsOutput` (setup) that uses output exceeding 20 lines and asserts the rendered summary shows `... (N lines omitted)` prefix followed by the last 10 lines.

**Checkpoint**: Long error output is truncated with actionable omission message. Run `go test -race -count=1 ./internal/scaffold/... ./internal/setup/...` to verify.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and documentation.

- [x] T021 Run `go vet ./...` and `golangci-lint run` to verify no lint issues introduced.
- [x] T022 Run full test suite `go test -race -count=1 ./...` to verify no regressions.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 1)**: No dependencies -- can start immediately
- **US1 (Phase 2)**: Depends on T001 (struct fields), T003 (truncateOutput), T004 (test recorder)
- **US2 (Phase 3)**: Depends on T002 (struct field). Can run in PARALLEL with US1
- **US3 (Phase 4)**: Depends on US1 and US2 (truncation tested end-to-end)
- **Polish (Phase 5)**: Depends on all user stories

### User Story Dependencies

- **US1 (P1)**: Depends on T001, T003, T004 (Foundational phase)
- **US2 (P2)**: Depends on T002 (Foundational phase). Can run in PARALLEL with US1
- **US3 (P3)**: Depends on US1 and US2 completion (integration tests)

### Parallel Opportunities

- T001-T004 can run in parallel (different files or independent changes)
- T005-T009 can run in parallel (different test functions, same file)
- T013-T015 can run in parallel (different test functions, same file)
- US1 and US2 can run in parallel after Foundational phase (different packages)

---

## Parallel Example: Foundational Phase

```bash
# Launch all foundational tasks in parallel:
Task T001: "Add err/output fields to subToolResult in internal/scaffold/scaffold.go"
Task T002: "Add output field to stepResult in internal/setup/setup.go"
Task T003: "Add truncateOutput helper in internal/scaffold/scaffold.go"
Task T004: "Extend scaffoldCmdRecorder in internal/scaffold/scaffold_test.go"
```

## Parallel Example: User Story 1 Tests

```bash
# Launch all US1 tests in parallel (same file, different functions):
Task T005: "TestInitSubTools_DeweyInitFails_ShowsError"
Task T006: "TestInitSubTools_SimpleToolFails_ShowsError"
Task T007: "TestTruncateOutput"
Task T008: "TestInitSubTools_ExitCodeOnly"
Task T009: "TestPrintSummary_SuccessUnchanged"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Foundational (T001-T004)
2. Complete Phase 2: User Story 1 (T005-T012)
3. **STOP and VALIDATE**: Run `go test -race -count=1 ./internal/scaffold/...`
4. `uf init` failures now show actual errors -- core problem solved

### Incremental Delivery

1. Foundational -> struct extensions, helpers, and test recorder ready
2. US1 -> scaffold error reporting fixed -> validate independently
3. US2 -> setup error reporting fixed -> validate independently
4. US3 -> truncation integration verified -> validate end-to-end
5. Polish -> lint, full test suite verification

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story
- T004 extends `scaffoldCmdRecorder` with an `outputs` map -- this is a blocking prerequisite for US1 test tasks (T005-T009)
- Existing failure tests (e.g., `TestInitSubTools_DeweyInitFails`) can be extended rather than duplicated -- add assertions for the new output field
- The `cmdRecorder` in setup tests already has an `outputs` map -- no test infrastructure changes needed there
- If `truncateOutput` behavior is modified in the future, both copies (scaffold and setup) MUST be updated together

<!-- spec-review: passed -->
<!-- code-review: passed -->
