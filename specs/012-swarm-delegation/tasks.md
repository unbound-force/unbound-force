# Tasks: Swarm Delegation Workflow

**Input**: Design documents from `/specs/012-swarm-delegation/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Tests are included -- the spec requires maintaining ~87% coverage for the orchestration package and the constitution mandates testability (Principle IV).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Stage rename from "measure" to "reflect" -- the foundational constant and mapping changes that all user stories depend on.

- [x] T001 Rename `StageMeasure` to `StageReflect` and update string value from `"measure"` to `"reflect"` in `internal/orchestration/models.go`. Also update the package doc comment (line 4) to reference `reflect` instead of `measure`
- [x] T002 Update `StageOrder()` to return `StageReflect` instead of `StageMeasure` in `internal/orchestration/models.go`
- [x] T003 Add `StatusAwaitingHuman = "awaiting_human"` constant in `internal/orchestration/models.go`
- [x] T004 Add execution mode constants `ModeHuman = "human"` and `ModeSwarm = "swarm"` in `internal/orchestration/models.go`
- [x] T005 Add `ExecutionMode string` field with `json:"execution_mode,omitempty"` tag to `WorkflowStage` struct in `internal/orchestration/models.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Hero mapping and execution mode infrastructure that MUST be complete before checkpoint and reflect logic can be implemented.

**CRITICAL**: No user story work can begin until this phase is complete.

- [x] T006 Update `heroSpecs` variable to use `StageReflect` instead of `StageMeasure` for the mx-f entry in `internal/orchestration/heroes.go`
- [x] T007 Update `StageHeroMap()` to return `StageReflect: "mx-f"` instead of `StageMeasure: "mx-f"` in `internal/orchestration/heroes.go`
- [x] T008 [P] Add `StageExecutionModeMap()` function returning `map[string]string` with default mode assignments (define=human, implement=swarm, validate=swarm, review=swarm, accept=human, reflect=swarm) in `internal/orchestration/heroes.go`
- [x] T009 Update `newTestOrchestrator` helper and `allAgentFiles`/`allBinaries` vars if needed in `internal/orchestration/engine_test.go` to ensure existing test infrastructure works with the rename
- [x] T010 Update `TestStageHeroMap` assertion from `StageMeasure` to `StageReflect` in `internal/orchestration/heroes_test.go`
- [x] T010b [P] Write `TestStageExecutionModeMap` in `internal/orchestration/heroes_test.go` — verify the map returns exactly 6 entries with correct mode assignments per FR-002 (define=human, implement=swarm, validate=swarm, review=swarm, accept=human, reflect=swarm)
- [x] T011 [P] Update the three `StageMeasure` references in test assertions (`TestOrchestrator_Start_MissingHeroes`, `TestGenerateWorkflowRecord_CompletedWorkflow`, `TestGenerateWorkflowRecord_RejectedWorkflow`) in `internal/orchestration/engine_test.go`
- [x] T012 Run `go test -race -count=1 ./internal/orchestration/...` and `go build ./...` to verify rename compiles and all existing tests pass

**Checkpoint**: All existing tests pass with the measure-to-reflect rename and new constants. Execution mode map is available.

---

## Phase 3: User Story 1 - Swarm Delegation After Clarify (Priority: P1)

**Goal**: The orchestration engine pauses the workflow at swarm-to-human boundaries and resumes when the human advances.

**Independent Test**: Start a workflow, advance through define (human), verify the swarm stages advance without pause, verify the workflow pauses at `awaiting_human` before accept, verify advancing from `awaiting_human` resumes the accept stage.

### Tests for User Story 1

- [x] T013 [P] [US1] Write `TestOrchestrator_NewWorkflow_SetsExecutionModes` -- verify each stage gets correct execution_mode from `StageExecutionModeMap()` in `internal/orchestration/engine_test.go`
- [x] T014 [P] [US1] Write `TestOrchestrator_Advance_PausesAtHumanCheckpoint` -- advance through define (human) and swarm stages, verify workflow transitions to `awaiting_human` when next stage is accept (human) in `internal/orchestration/engine_test.go`
- [x] T015 [P] [US1] Write `TestOrchestrator_Advance_ResumesFromCheckpoint` -- set workflow to `awaiting_human`, call Advance, verify accept stage activates and status returns to `active` in `internal/orchestration/engine_test.go`
- [x] T016 [P] [US1] Write `TestOrchestrator_Advance_SwarmToSwarmNoPause` -- verify swarm-to-swarm transitions (implement→validate, validate→review) do NOT trigger `awaiting_human` in `internal/orchestration/engine_test.go`

### Implementation for User Story 1

- [x] T017 [US1] Update `NewWorkflow()` in `internal/orchestration/engine.go` to populate `ExecutionMode` on each stage from `StageExecutionModeMap()`
- [x] T018 [US1] Add checkpoint detection logic to `Advance()` in `internal/orchestration/engine.go`: when completing a swarm-mode stage and the next non-skipped stage is human-mode, set workflow status to `StatusAwaitingHuman` and return without activating the next stage
- [x] T019 [US1] Add resume logic to `Advance()` in `internal/orchestration/engine.go`: first, modify the status guard at line 162 to accept both `StatusActive` and `StatusAwaitingHuman` (reject completed, failed, escalated). When status is `StatusAwaitingHuman`, skip the current-stage-completion logic and proceed directly to finding the next pending non-skipped stage, activate it, and set status back to `StatusActive`
- [x] T020 [US1] Run `go test -race -count=1 ./internal/orchestration/...` to verify all US1 tests pass

**Checkpoint**: The core delegation mechanism works -- workflows pause at swarm→human boundaries and resume on advance.

---

## Phase 4: User Story 2 - Execution Mode Per Stage (Priority: P1)

**Goal**: Execution modes are persisted, backward compatible, and discoverable via `Latest()`.

**Independent Test**: Create a workflow, inspect persisted JSON for `execution_mode` fields. Load a legacy JSON file without `execution_mode`, verify it works as all-human. Call `Latest()` on an `awaiting_human` workflow and verify it's found.

### Tests for User Story 2

- [x] T021 [P] [US2] Write `TestOrchestrator_Advance_LegacyWorkflowNoCheckpoints` -- create a workflow, manually clear all `execution_mode` fields, advance through all stages, verify no `awaiting_human` pause occurs in `internal/orchestration/engine_test.go`
- [x] T022 [P] [US2] Write `TestWorkflowStore_Latest_FindsAwaitingHuman` -- save a workflow with `StatusAwaitingHuman`, verify `Latest()` returns it in `internal/orchestration/store_test.go`

### Implementation for User Story 2

- [x] T023 [US2] Update `WorkflowStore.Latest()` in `internal/orchestration/store.go` to query both `StatusActive` and `StatusAwaitingHuman` workflows when finding the latest for a branch
- [x] T024 [P] [US2] Write `TestOrchestrator_Advance_AllSwarmSkipped_NoCheckpoint` -- skip all swarm stages (implement, validate, review), advance from define, verify workflow transitions directly to accept without entering `awaiting_human` in `internal/orchestration/engine_test.go`
- [x] T025 [P] [US2] Write `TestOrchestrator_Advance_EscalationWithExecutionModes` -- set stages to swarm mode with iteration count at max, advance to review, verify escalation still fires correctly regardless of execution mode in `internal/orchestration/engine_test.go`
- [x] T026 [US2] Run `go test -race -count=1 ./internal/orchestration/...` to verify all US2 tests pass

**Checkpoint**: Execution modes persist correctly, legacy workflows are backward compatible, and `Latest()` discovers paused workflows.

---

## Phase 5: User Story 3 - Reflect Stage (Priority: P2)

**Goal**: The final stage is named "reflect," assigned to Mx F, and the SKILL.md documents its enriched behavior (metrics + learning + retrospective).

**Independent Test**: Complete a full workflow and verify the final stage is named "reflect" in the persisted JSON and workflow record. Verify the SKILL.md describes artifact consumption from validate and review stages.

### Implementation for User Story 3

- [x] T027 [US3] Update the Workflow Stages section in `.opencode/skill/unbound-force-heroes/SKILL.md`: rename stage 6 from `measure` to `reflect`, update artifact output description to include metrics snapshot, learning feedback, and retrospective summary
- [x] T028 [US3] Add a "Reflect Stage" subsection in `.opencode/skill/unbound-force-heroes/SKILL.md` documenting that the stage consumes quality-report from validate and review-verdict from review, runs `AnalyzeWorkflows` for learning patterns, and produces a retrospective summary with empirical data
- [x] T029 [US3] Rewrite `TestOrchestrator_Advance_ThroughAllStages` in `internal/orchestration/engine_test.go`: the advance loop must handle the `awaiting_human` intermediate state (the 4th advance triggers `awaiting_human` at the review→accept boundary, the 5th advance resumes, the 6th and 7th complete accept and reflect). Verify the workflow passes through `awaiting_human` exactly once, the final stage is named `"reflect"`, and all stages are completed
- [x] T030 [US3] Run `go test -race -count=1 ./internal/orchestration/...` to verify reflect stage tests pass

**Checkpoint**: The reflect stage is fully named, mapped, and documented with enriched behavior in the SKILL.md.

---

## Phase 6: User Story 4 - Workflow Status Shows Execution Mode (Priority: P3)

**Goal**: The `/workflow` command documentation shows execution mode indicators per stage and a distinct `awaiting_human` display.

**Independent Test**: Review the workflow command Markdown files and verify they include `[human]`/`[swarm]` mode indicators and the `⏸` indicator for the awaiting_human state.

### Implementation for User Story 4

- [x] T031 [P] [US4] Update `.opencode/command/workflow-start.md`: rename `measure` to `reflect` in stage list and sample output, add `[human]`/`[swarm]` mode indicators to each stage line in the hero availability output
- [x] T032 [P] [US4] Update `.opencode/command/workflow-status.md`: rename `measure` to `reflect`, add `[human]`/`[swarm]` mode indicators to each stage row, add `⏸` indicator and `← your turn` annotation for awaiting_human display, add resume instruction
- [x] T033 [P] [US4] Update `.opencode/command/workflow-advance.md`: rename `measure` to `reflect`, add "Checkpoint Reached" output format for swarm→human boundary, add "Resumed from Checkpoint" output format, add `[human]`/`[swarm]` mode indicators to all stage rows
- [x] T034 [US4] Add "Execution Modes" section to `.opencode/skill/unbound-force-heroes/SKILL.md` documenting the human/swarm assignment per stage and the "Swarm Delegation" handoff pattern (depends on T027/T028 -- same file)

**Checkpoint**: All workflow command documentation reflects execution modes, the reflect rename, and the awaiting_human state.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Documentation updates, final validation, and cross-cutting cleanup.

- [x] T035 [P] Update the Recent Changes entry for `008-swarm-orchestration` in `AGENTS.md` to mention the measure→reflect rename and execution mode additions
- [x] T036 [P] Update the inter-hero artifact types table in `AGENTS.md` to change `metrics-snapshot` producer stage reference from `measure` to `reflect` (if referenced)
- [x] T037 Create `schemas/workflow-record/v1.1.0.schema.json` by copying v1.0.0 and adding `execution_mode` as an optional string property on the stage items object (the existing schema has `additionalProperties: false` which rejects unknown fields)
- [x] T038 [P] Update `schemas/workflow-record/samples/` with a sample that includes `execution_mode` fields and the `"reflect"` stage name
- [x] T039 [P] Update `schemas/workflow-record/README.md` with a changelog entry for v1.1.0
- [x] T040 Run `go test -race -count=1 ./internal/schemas/...` to verify schema validation tests pass with the new schema version
- [x] T041 Run full test suite `go test -race -count=1 ./...` and `go build ./...` to verify zero regressions across all packages
- [x] T042 Run quickstart.md validation: manually trace the quickstart scenario against the updated command docs and engine behavior to verify consistency

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion -- BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - US1 and US2 are both P1 but US2's `Latest()` change depends on US1's `StatusAwaitingHuman` usage being in place
  - US3 and US4 can run in parallel with each other
  - US4 can run in parallel with US1/US2 (Markdown-only, no Go code overlap)
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) -- No dependencies on other stories
- **User Story 2 (P1)**: Depends on US1 (needs `StatusAwaitingHuman` and checkpoint logic to test against)
- **User Story 3 (P2)**: Can start after Foundational (Phase 2) -- Reflect rename is in Setup/Foundational, SKILL.md is independent
- **User Story 4 (P3)**: T031-T033 can start after Foundational (Phase 2) -- different Markdown files, no Go code dependencies. T034 depends on US3's T027/T028 (same SKILL.md file)

### Within Each User Story

- Tests written first (where included), then implementation
- Engine changes before store changes
- Go code before Markdown documentation
- Run test checkpoint at end of each story

### Parallel Opportunities

- T004, T005 are sequential (same file `models.go`)
- T006, T007, T008 can be parallelized (T008 is independent, T006/T007 are in same function but different lines)
- T013, T014, T015, T016 can all run in parallel (independent test functions)
- T021, T022, T024, T025 can run in parallel (independent test functions)
- T031, T032, T033 can all run in parallel (different Markdown files)
- T035, T036 can run in parallel (different sections of AGENTS.md)
- US3 and US4 can run in parallel with each other (T034 depends on T027/T028 but other US4 tasks are independent)

---

## Parallel Example: User Story 1

```bash
# Launch all US1 tests in parallel:
Task: "T013 Write TestOrchestrator_NewWorkflow_SetsExecutionModes in engine_test.go"
Task: "T014 Write TestOrchestrator_Advance_PausesAtHumanCheckpoint in engine_test.go"
Task: "T015 Write TestOrchestrator_Advance_ResumesFromCheckpoint in engine_test.go"
Task: "T016 Write TestOrchestrator_Advance_SwarmToSwarmNoPause in engine_test.go"
```

## Parallel Example: User Story 4

```bash
# Launch all US4 Markdown updates in parallel:
Task: "T031 Update workflow-start.md"
Task: "T032 Update workflow-status.md"
Task: "T033 Update workflow-advance.md"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (rename + constants)
2. Complete Phase 2: Foundational (hero maps + test fixes)
3. Complete Phase 3: User Story 1 (checkpoint logic)
4. **STOP and VALIDATE**: Run `go test -race -count=1 ./internal/orchestration/...` -- the core delegation works
5. The swarm can now pause at human checkpoints

### Incremental Delivery

1. Setup + Foundational → All existing tests pass with rename
2. Add US1 → Checkpoint pause/resume works → Core value delivered
3. Add US2 → Backward compat + Latest() discovery → Production ready
4. Add US3 → Reflect stage fully documented → Learning loop enriched
5. Add US4 → Command docs updated → Operator experience complete
6. Polish → Full regression pass → Ship it

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (engine.go changes)
   - Developer B: User Story 4 (Markdown files -- no Go overlap)
3. After US1 completes:
   - Developer A: User Story 2 (store.go + edge case tests)
   - Developer B: User Story 3 (SKILL.md + reflect docs)
4. Polish phase together

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each phase checkpoint
- The reflect stage's enriched runtime behavior (consuming artifacts, running AnalyzeWorkflows) is documented in SKILL.md for the Swarm coordinator to follow -- the Go engine only tracks metadata
- No scaffold asset count changes -- the heroes SKILL.md is a local-only file
