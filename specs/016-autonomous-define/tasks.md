# Tasks: Autonomous Define with Dewey

**Input**: Design documents from `/specs/016-autonomous-define/`
**Prerequisites**: plan.md, spec.md, research.md, contracts/workflow-seed.md, quickstart.md

**Tests**: Tests are included -- the spec requires zero regressions (SC-005, SC-006) and the constitution mandates testability (Principle IV).

**Organization**: US1 and US2 are both P1 but US1 is the engine layer (Go code) while US2 is the agent layer (Markdown). They can be implemented in parallel since they touch different files.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to
- Include exact file paths in descriptions

## Phase 1: Setup (Data Model Changes)

**Purpose**: Add the execution mode override mechanism and spec review flag to the workflow data model.

- [x] T001 Add `SpecReviewEnabled bool` field with `json:"spec_review_enabled,omitempty"` tag to the `WorkflowInstance` struct in `internal/orchestration/models.go`
- [x] T002 Update `NewWorkflow()` signature in `internal/orchestration/engine.go` to accept `overrides map[string]string` and `specReview bool` parameters. Return `(*WorkflowInstance, error)` instead of `*WorkflowInstance`. Apply overrides on top of `StageExecutionModeMap()` defaults. Set `SpecReviewEnabled` from `specReview`. Validate override keys against `StageOrder()` (reject unknown stage names) and values against `ModeHuman`/`ModeSwarm` (reject unknown modes).
- [x] T002b Update `Start()` signature in `internal/orchestration/engine.go` to accept `overrides map[string]string` and `specReview bool`, forwarding them to `NewWorkflow()`. Handle the new error return from `NewWorkflow()`.
- [x] T003 Update all existing `Start()` call sites (20+ in `engine_test.go`, 1 in `store_test.go`) to pass `nil, false` for the new parameters. Update the 1 direct `NewWorkflow()` call in `engine_test.go` similarly. Explicitly verify `Start()` at `engine.go:117` is updated.
- [x] T004 Run `go build ./...` and `go test -race -count=1 ./internal/orchestration/...` to verify compilation and all existing tests pass with the new signature

---

## Phase 2: Foundational (Checkpoint Logic)

**Purpose**: Add the spec review checkpoint to `Advance()`. MUST complete before user story work begins.

- [x] T005 Add spec review checkpoint logic to `Advance()` in `internal/orchestration/engine.go`: after completing the define stage, if `workflow.SpecReviewEnabled` is true AND the define stage's execution mode is `swarm`, set workflow status to `StatusAwaitingHuman` and return without activating the next stage. This fires based on define=swarm (not implement's mode), since the checkpoint is about reviewing the autonomously-drafted spec.
- [x] T006 Write `TestOrchestrator_Advance_SpecReviewCheckpoint` in `internal/orchestration/engine_test.go`: create a workflow with `define=swarm`, `SpecReviewEnabled=true`, advance through define, verify workflow pauses with `StatusAwaitingHuman` before implement
- [x] T006b Write `TestOrchestrator_Advance_SpecReviewCheckpoint_DefineHuman` in `internal/orchestration/engine_test.go`: create a workflow with `define=human`, `SpecReviewEnabled=true`, advance through define, verify the workflow proceeds directly to implement WITHOUT pausing (the checkpoint is silently skipped because the human was already involved)
- [x] T007 Write `TestOrchestrator_Advance_SpecReviewDisabled` in `internal/orchestration/engine_test.go`: create a workflow with `define=swarm`, `SpecReviewEnabled=false`, advance through define, verify workflow proceeds directly to implement without pausing
- [x] T008 Write `TestOrchestrator_NewWorkflow_WithOverrides` in `internal/orchestration/engine_test.go`: create a workflow with `overrides={"define": "swarm"}`, verify the define stage has `execution_mode=swarm` while all other stages retain their defaults
- [x] T009 Write `TestOrchestrator_NewWorkflow_DefaultOverrides` in `internal/orchestration/engine_test.go`: create a workflow with `nil` overrides, verify all stages match the default `StageExecutionModeMap()` values (backward compatibility)
- [x] T010 Run `go test -race -count=1 ./internal/orchestration/...` to verify all new and existing tests pass

**Checkpoint**: The orchestration engine supports configurable execution modes and the spec review checkpoint. All existing tests pass.

---

## Phase 3: User Story 1 - Configurable Define Stage (Priority: P1)

**Goal**: The define stage can be configured as `[swarm]` via the workflow start command and the execution mode is persisted in the workflow JSON.

**Independent Test**: Start a workflow with `define=swarm`, verify the define stage executes in swarm mode and the workflow proceeds through all swarm stages to the accept checkpoint without human intervention.

- [x] T011 [US1] Write `TestOrchestrator_Advance_AutonomousDefine_ThroughAllStages` in `internal/orchestration/engine_test.go`: create a workflow with `define=swarm` and `SpecReviewEnabled=false`, advance through all stages, verify the workflow passes through `awaiting_human` exactly once (before accept), and all 6 stages complete
- [x] T012 [US1] Add execution mode validation to `NewWorkflow()` in `internal/orchestration/engine.go`: if an override value is not `"human"` or `"swarm"`, return an error with a clear message listing valid values (FR-004)
- [x] T013 [US1] Write `TestOrchestrator_NewWorkflow_InvalidOverrideValue` in `internal/orchestration/engine_test.go`: pass `overrides={"define": "auto"}`, verify an error is returned with a message containing "human" and "swarm"
- [x] T013b [US1] Write `TestOrchestrator_NewWorkflow_InvalidOverrideKey` in `internal/orchestration/engine_test.go`: pass `overrides={"nonexistent": "swarm"}`, verify an error is returned with a message indicating the stage name is invalid
- [x] T013c [US1] Write `TestOrchestrator_Start_InvalidOverride_ReturnsError` in `internal/orchestration/engine_test.go`: call `Start()` with an invalid override, verify the error propagates correctly through `Start()` (not a nil-pointer panic or silent success)
- [x] T014 [US1] Run `go test -race -count=1 ./internal/orchestration/...` to verify all US1 tests pass

**Checkpoint**: The define stage can be configured as swarm, the config is validated, and the workflow advances correctly.

---

## Phase 4: User Story 2 - Muti-Mind Autonomous Specification (Priority: P1)

**Goal**: Muti-Mind's agent file includes instructions for autonomous specification drafting using Dewey context.

**Independent Test**: Read the Muti-Mind agent file. Verify it includes an "Autonomous Specification" section with steps for seed intake, Dewey context retrieval, self-clarification, and spec output.

- [x] T015 [US2] Add an "Autonomous Specification Workflow" section to `.opencode/agents/muti-mind-po.md`: describe the step-by-step workflow for autonomous spec drafting: (1) accept seed description, (2) query Dewey for related specs/issues/docs, (3) draft spec using speckit template, (4) self-clarify by querying Dewey for ambiguities instead of asking the human, (5) reference learning feedback from past workflows if 3+ records exist, (6) produce the spec as a file artifact
- [x] T016 [US2] Add Dewey context retrieval examples to the Autonomous Specification section in `.opencode/agents/muti-mind-po.md`: specific query examples for each context type (e.g., `dewey_semantic_search "authentication patterns"` for related specs, `dewey_semantic_search_filtered source_type=github` for related issues)
- [x] T017 [US2] Add a Tier 1 fallback subsection to the Autonomous Specification section in `.opencode/agents/muti-mind-po.md`: when Dewey is unavailable, use Read tool on local backlog items and convention packs to produce a less contextual but valid spec

**Checkpoint**: Muti-Mind has clear instructions for autonomous spec drafting with Dewey context and Tier 1 fallback.

---

## Phase 5: User Story 3 - Optional Spec Review Checkpoint (Priority: P2)

**Goal**: The spec review checkpoint is configurable and documented in the workflow commands.

**Independent Test**: Start a workflow with spec review enabled. Verify the workflow pauses after define with a clear message.

- [x] T018 [US3] Update `.opencode/command/workflow-start.md`: document the `--define-mode` flag (accepts `human` or `swarm`, default `human`) and the `--spec-review` flag (enables the spec review checkpoint between define and implement)
- [x] T019 [US3] Update `.opencode/command/workflow-advance.md`: document the "Spec review" output format when the workflow pauses after autonomous define with spec review enabled (similar to the "Checkpoint Reached" format from Spec 012)
- [x] T020 [US3] Update `.opencode/command/workflow-status.md`: document the `⏸ define (muti-mind) completed [swarm] ← spec ready for review` indicator when the workflow is paused at the spec review checkpoint

**Checkpoint**: Workflow commands are documented for the spec review checkpoint.

---

## Phase 6: User Story 4 - Seed Workflow Command (Priority: P2)

**Goal**: A `/workflow seed` command exists that starts a workflow with one sentence.

**Independent Test**: Run the seed command with a description. Verify a workflow starts with define in swarm mode.

- [x] T021 [US4] Create `.opencode/command/workflow-seed.md` with the seed command definition: accepts a feature description, creates a backlog item, starts a workflow with `define=swarm`, returns the workflow ID. Include the output format from the contract.
- [x] T022 [US4] Update `.opencode/skill/unbound-force-heroes/SKILL.md`: add a "Seed Workflow" section describing the seed-to-accept workflow, the configurable define stage, and the optional spec review checkpoint. Include the comparison table from the quickstart (default vs autonomous vs autonomous+review).

**Checkpoint**: The seed command is defined and the SKILL.md documents the autonomous workflow.

---

## Phase 7: User Story 5 - Documentation (Priority: P3)

**Goal**: AGENTS.md is updated with the new spec and workflow changes.

- [x] T023 [P] [US5] Update `AGENTS.md` Spec Organization section: add Spec 016 to Phase 2 (Cross-Cutting) with description "Autonomous define with Dewey -- configurable execution modes, seed command, spec review checkpoint"
- [x] T024 [P] [US5] Update `AGENTS.md` Dependency Graph: add `016-autonomous-define` under Phase 2 with dependencies on 012, 014, 015
- [x] T025 [US5] Update `AGENTS.md` Recent Changes section: add entry for `016-autonomous-define` describing the autonomous define feature

---

## Phase 8: Polish & Cross-Cutting Concerns

- [x] T026 Run full test suite `go test -race -count=1 ./...` and `go build ./...` to verify zero regressions across all packages
- [x] T027 Verify backward compatibility: run existing `TestOrchestrator_Advance_ThroughAllStages` (updated in Spec 012) and confirm it still passes with the default mode map (define=human)
- [x] T028 [P] Verify that `spec_review_enabled` is on `WorkflowInstance` only (NOT on `WorkflowRecord`). Confirm that `GenerateWorkflowRecord()` in `internal/orchestration/record.go` does NOT copy the field to the record. The workflow-record schema at `schemas/workflow-record/v1.1.0.schema.json` does NOT need updating (it has `additionalProperties: false` and the field is instance-only).
- [x] T028b Write `TestWorkflowStore_Load_LegacyJSON_SpecReviewDefaultsFalse` in `internal/orchestration/store_test.go`: create a workflow JSON file without the `spec_review_enabled` field, load it, verify `SpecReviewEnabled == false` (backward compatibility regression guard)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- data model changes first
- **Foundational (Phase 2)**: Depends on Phase 1 -- checkpoint logic needs the new fields
- **US1 (Phase 3)**: Depends on Phase 2 -- configurable modes + validation
- **US2 (Phase 4)**: Can run in parallel with Phase 3 (Markdown agent file vs Go code)
- **US3 (Phase 5)**: Depends on Phase 2 (checkpoint logic exists before documenting it)
- **US4 (Phase 6)**: Depends on US1 (seed command uses define=swarm override)
- **US5 (Phase 7)**: Depends on all other phases
- **Polish (Phase 8)**: Depends on all phases

### User Story Dependencies

- **US1 (P1)**: Depends on Phase 2 (checkpoint + overrides)
- **US2 (P1)**: Independent of US1 (Markdown vs Go -- parallel OK)
- **US3 (P2)**: Depends on Phase 2
- **US4 (P2)**: Depends on US1 (seed uses define=swarm)
- **US5 (P3)**: Depends on all

### Parallel Opportunities

- T015, T016 can run in parallel (same file, different sections -- but sequential is safer)
- Phase 4 (US2, Markdown) can run in parallel with Phase 3 (US1, Go code)
- T023, T024 can run in parallel (different AGENTS.md sections)

---

## Implementation Strategy

### MVP First (US1 + US2)

1. Phase 1: Data model changes (T001-T004)
2. Phase 2: Checkpoint logic (T005-T010)
3. Phase 3: Configurable define + validation (T011-T014)
4. Phase 4: Muti-Mind agent instructions (T015-T017)
5. **STOP and VALIDATE**: The autonomous define works end-to-end
6. Core value delivered -- seed a feature, swarm handles everything to accept

### Incremental Delivery

1. Setup + Foundational → Engine supports overrides + checkpoint
2. US1 → Configurable define works → Core mechanism
3. US2 → Muti-Mind knows how to draft specs → Autonomous capability
4. US3 → Spec review documented → Risk mitigation for high-stakes
5. US4 → Seed command exists → UX polish
6. US5 → Docs updated → Discoverability
7. Polish → Full regression pass → Ship it

---

## Notes

- The autonomous spec drafting (US2) is agent file instructions, not Go code. Muti-Mind follows the instructions in its agent file when the Swarm coordinator invokes it during the define stage.
- The spec review checkpoint reuses `StatusAwaitingHuman` from Spec 012. No new status constants.
- The seed command is an OpenCode command (Markdown), not a Go CLI command. The Swarm coordinator interprets it and calls the orchestration engine's `NewWorkflow()` with the appropriate overrides.
- The `--define-mode` and `--spec-review` flags are documented in the workflow command Markdown files. They are interpreted by the Swarm coordinator, not by Go code.
- Completed specs (001-015) are NOT modified.
