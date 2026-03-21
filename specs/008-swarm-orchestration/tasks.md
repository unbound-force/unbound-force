# Tasks: Swarm Orchestration

**Input**: Design documents from `/specs/008-swarm-orchestration/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/workflow-commands.md

**Tests**: Tests are included — the plan specifies 85% coverage target and the constitution's Testability principle mandates coverage strategy.

**Organization**: Tasks are grouped by user story. A foundational phase creates the artifact context and orchestration package infrastructure.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Create the orchestration package and artifact context infrastructure.

- [ ] T001 Create `internal/orchestration/` package directory.
- [ ] T002 Add `ArtifactContext` struct to `internal/artifacts/artifacts.go` per research.md R3: `Branch`, `Commit`, `BacklogItemID`, `CorrelationID`, `WorkflowID` fields with JSON tags and `omitempty`. Update `WriteArtifact` to accept an optional `*ArtifactContext` parameter — marshal it into the `Envelope.Context` field if non-nil. Update existing callers (Muti-Mind's `GenerateBacklogItemArtifact`, `GenerateAcceptanceDecision`) to pass `nil` for backward compatibility.
- [ ] T003 Add `TestWriteArtifact_WithContext` to `internal/artifacts/artifacts_test.go`: write artifact with context, read it back via `ReadEnvelope`, verify context fields are populated (branch, workflow_id).
- [ ] T004 Add workflow commands and skills to `knownNonEmbeddedFiles` in `internal/scaffold/scaffold_test.go`: add `".opencode/command/workflow-start.md"`, `".opencode/command/workflow-status.md"`, `".opencode/command/workflow-list.md"`, `".opencode/command/workflow-advance.md"`. Also add `".opencode/skill/"` entries if the drift detection test scans that directory.

---

## Phase 2: Foundational — Orchestration Models + Store

**Purpose**: Create the data model structs and workflow state persistence layer. All user stories depend on these.

- [ ] T005 Create `internal/orchestration/models.go` with Go structs per data-model.md: `WorkflowInstance` (workflow_id, feature_branch, backlog_item_id, stages, current_stage, started_at, completed_at, status, available_heroes, iteration_count), `WorkflowStage` (stage_name, hero, status, artifacts_produced, artifacts_consumed, started_at, completed_at, skip_reason, error), `Decision` (type, hero, result, rationale, iteration, timestamp), `LearningFeedback` (id, source_hero, target_hero, pattern_observed, recommendation, supporting_data, status, created_at, workflow_ids), `WorkflowRecord` (workflow_id, backlog_item_id, stages, artifacts, decisions, total_elapsed_time, outcome, learning_feedback), `HeroStatus` (name, role, available, agent_file, detection_method). Include JSON struct tags on all fields. Add constants for stage names (`StageDefine`, `StageImplement`, `StageValidate`, `StageReview`, `StageAccept`, `StageMeasure`) and status values (`StatusPending`, `StatusActive`, `StatusCompleted`, `StatusSkipped`, `StatusFailed`, `StatusEscalated`).
- [ ] T006 Create `internal/orchestration/store.go` with `WorkflowStore` struct (wraps workflow dir path), `Save(wf *WorkflowInstance) error` (writes JSON to `.unbound-force/workflows/{workflow_id}.json`), `Load(workflowID string) (*WorkflowInstance, error)`, `List(statusFilter string) ([]WorkflowInstance, error)` (reads all workflow files, filters by status, sorts by started_at descending), `Latest(branch string) (*WorkflowInstance, error)` (returns most recent active workflow for a branch).
- [ ] T007 Create `internal/orchestration/store_test.go` with `TestWorkflowStore_SaveLoad_RoundTrip`, `TestWorkflowStore_List_FilterByStatus`, `TestWorkflowStore_Latest_ByBranch`, `TestWorkflowStore_Load_MissingFile`, `TestWorkflowStore_List_Empty`. All tests use `t.TempDir()`.

**Checkpoint**: Orchestration models defined. Workflow state persistence works with round-trip tests passing.

---

## Phase 3: User Story 1 — Feature Lifecycle Workflow (Priority: P1) MVP

**Goal**: Implement the core orchestration engine with Start, Advance, Skip, Escalate, and Complete methods. The 6-stage lifecycle workflow (define → implement → validate → review → accept → measure) with stage transitions.

**Independent Test**: Create a workflow, advance through all stages, verify each stage transitions correctly and the workflow completes.

### Hero Detection

- [ ] T008 [US1] Create `internal/orchestration/heroes.go` with `DetectHeroes(agentDir string, lookPath func(string) (string, error)) ([]HeroStatus, error)`: check for agent files per research.md R4 — `muti-mind-po.md`, `cobalt-crush-dev.md`, any `divisor-*.md`, `mx-f-coach.md`. For Gaze and Mx F CLI, use the injected `lookPath` function (defaults to `exec.LookPath` in production, stub in tests). Return `[]HeroStatus` with name, role, available, agent_file, detection_method. Add `StageHeroMap() map[string]string` that maps stage names to hero names.
- [ ] T009 [US1] Create `internal/orchestration/heroes_test.go` with `TestDetectHeroes_AllPresent` (create temp dir with all agent files, verify all detected), `TestDetectHeroes_MissingHeroes` (partial agent files, verify correct available/unavailable), `TestDetectHeroes_EmptyDir` (no agents, all unavailable). Use `t.TempDir()` with dummy agent files.

### Core Engine

- [ ] T010 [US1] Create `internal/orchestration/engine.go` with `Orchestrator` struct per contracts/workflow-commands.md Go API: `WorkflowDir`, `ArtifactDir`, `AgentDir`, `GHRunner`, `Now func() time.Time`, `Stdout io.Writer`. Implement:
  - `NewWorkflow(branch, backlogItemID string) *WorkflowInstance` — creates instance with 6 stages, detects heroes, marks unavailable stages as skipped
  - `Start(branch, backlogItemID string) (*WorkflowResult, error)` — creates workflow, saves it, returns result with warnings for skipped stages
  - `Advance(workflowID string) (*WorkflowResult, error)` — completes current stage, moves to next non-skipped stage. If review stage and iteration_count < 3, allow retry. If current stage is last, call Complete.
  - `Skip(workflowID string, stage int, reason string) error` — marks stage as skipped with reason
  - `Escalate(workflowID, reason string) error` — sets workflow status to escalated
  - `Complete(workflowID string) (*WorkflowRecord, error)` — finalizes workflow, produces workflow-record artifact via `artifacts.WriteArtifact`
  - `Status(workflowID string) (*WorkflowInstance, error)` — loads and returns workflow
  - `List(statusFilter string) ([]WorkflowInstance, error)` — delegates to store

### Engine Tests

- [ ] T011 [US1] Create `internal/orchestration/engine_test.go` with:
  - `TestOrchestrator_Start_AllHeroes`: all agents present, verify 6 stages created, none skipped, status active
  - `TestOrchestrator_Start_MissingHeroes`: some agents missing, verify skipped stages have skip_reason
  - `TestOrchestrator_Advance_ThroughAllStages`: start workflow, advance 6 times, verify completed status
  - `TestOrchestrator_Advance_SkipsUnavailable`: verify advance skips over skipped stages
  - `TestOrchestrator_Escalate`: verify workflow status changes to escalated
  - `TestOrchestrator_Complete_ProducesRecord`: verify workflow-record artifact is written
  - `TestOrchestrator_Start_NoHeroes`: empty agent dir, verify all stages skipped, warning produced

### Workflow Record

- [ ] T012 [US1] Create `internal/orchestration/record.go` with `GenerateWorkflowRecord(wf *WorkflowInstance, now time.Time) *WorkflowRecord` — extracts all stage timing, artifact paths, decisions, computes total elapsed time, determines outcome (shipped if all non-skipped completed, rejected if acceptance rejected, abandoned if failed/escalated).
- [ ] T013 [US1] Add `TestGenerateWorkflowRecord_CompletedWorkflow`, `TestGenerateWorkflowRecord_EscalatedWorkflow`, and `TestGenerateWorkflowRecord_RejectedWorkflow` to `internal/orchestration/engine_test.go`.
- [ ] T013b [US1] Add `TestOrchestrator_ConcurrentWorkflows_Isolated` to `internal/orchestration/engine_test.go`: create two workflows on different branches, advance both, verify each workflow's state is scoped to its branch. Verify `Latest(branch)` returns the correct workflow for each branch.

**Checkpoint**: `Orchestrator.Start()` creates workflows. `Advance()` transitions stages. `Complete()` produces workflow-record artifacts. All tests pass.

---

## Phase 4: User Story 2 — Artifact Handoff Protocol (Priority: P1)

**Goal**: Enhance artifact discovery with filtering by hero and time range. Add schema version compatibility checking. Ensure the orchestration engine records artifact production/consumption at each stage.

**Independent Test**: Write artifacts from different heroes, query by type/hero/time, verify discovery returns correct results.

- [ ] T014 [US2] Add `FindArtifactsByHero(dir, artifactType, hero string) ([]string, error)` to `internal/artifacts/artifacts.go`: extends `FindArtifacts` to also filter by `Envelope.Hero` field. Add `FindArtifactsSince(dir, artifactType string, since time.Time) ([]string, error)` for time-range filtering.
- [ ] T015 [US2] Add `CheckSchemaVersion(envelope *Envelope, expectedVersion string) (compatible bool, warning string)` to `internal/artifacts/artifacts.go`: compares `Envelope.SchemaVersion` with expected. Returns compatible=true if major version matches, warning if minor/patch differs.
- [ ] T016 [US2] Add tests to `internal/artifacts/artifacts_test.go`: `TestFindArtifactsByHero_FiltersCorrectly`, `TestFindArtifactsSince_FiltersByTime`, `TestCheckSchemaVersion_Compatible`, `TestCheckSchemaVersion_MajorMismatch`.
- [ ] T017 [US2] Update `Orchestrator.Advance()` in `internal/orchestration/engine.go` to record `artifacts_consumed` on the current stage (discover artifacts from previous stage's hero) and `artifacts_produced` (placeholder paths for what the hero should produce). This is metadata tracking, not actual hero invocation.

**Checkpoint**: Artifact discovery supports filtering by hero and time. Schema version checking works. Engine tracks artifact flow per stage.

---

## Phase 5: User Story 3 — Swarm Plugin Integration (Priority: P2)

**Goal**: Create the Swarm skills package and ensure the workflow commands work alongside the Swarm plugin.

**Independent Test**: Load the skills package content, verify it describes all 5 heroes with routing patterns.

- [ ] T018 [P] [US3] Create `.opencode/skill/unbound-force-heroes/SKILL.md` with YAML frontmatter (`name: unbound-force-heroes`, `description: "Unbound Force hero roles and workflow routing"`, `tags: [heroes, workflow, routing]`) and Markdown body describing: each hero (Muti-Mind, Cobalt-Crush, Gaze, The Divisor, Mx F) with role, agent file, and natural language routing patterns. Include the 6-stage workflow sequence. Include escalation rules (max 3 iterations, conflict → human).
- [ ] T019 [P] [US3] Create `.opencode/command/workflow-start.md` per contracts/workflow-commands.md: OpenCode command that reads the current git branch, reads `.unbound-force/workflows/` for existing workflows, checks `.opencode/agents/` for hero availability, creates a new workflow JSON file, and reports the workflow start with available/skipped heroes.
- [ ] T020 [P] [US3] Create `.opencode/command/workflow-status.md` per contracts/workflow-commands.md: OpenCode command that reads the most recent workflow for the current branch from `.unbound-force/workflows/`, displays stage status with checkmarks/dots, lists artifacts produced.
- [ ] T021 [P] [US3] Create `.opencode/command/workflow-list.md` per contracts/workflow-commands.md: OpenCode command that reads all workflow files from `.unbound-force/workflows/`, displays a table with ID, branch, status, started time.

- [ ] T021b [P] [US3] Create `.opencode/command/workflow-advance.md`: OpenCode command that reads the current workflow from `.unbound-force/workflows/`, identifies the current stage, validates the stage can advance (artifacts produced, no blocking failures), advances to the next non-skipped stage, updates the workflow JSON, and reports the new stage with the next hero action.

**Checkpoint**: Swarm skills package exists with hero routing. Four `/workflow` commands are functional (start, status, list, advance).

---

## Phase 6: User Story 4 — Learning Loop (Priority: P3)

**Goal**: Implement learning feedback extraction from completed workflows. Analyze patterns across multiple workflow records.

**Independent Test**: Complete 3 workflows with consistent review findings, verify learning feedback is generated.

- [ ] T022 [US4] Create `internal/orchestration/learning.go` with:
  - `AnalyzeWorkflows(records []WorkflowRecord) ([]LearningFeedback, error)` — analyzes completed workflow records for patterns: frequent Divisor findings (same category > 3x), declining quality trends (coverage dropping), velocity patterns.
  - `NextFeedbackID(existing []LearningFeedback) string` — auto-increment `LF-NNN`.
  - `SaveFeedback(dir string, feedback []LearningFeedback) error` — write to `.unbound-force/workflows/learning/`.
  - `LoadFeedback(dir string) ([]LearningFeedback, error)` — read all feedback.
- [ ] T023 [US4] Create `internal/orchestration/learning_test.go` with:
  - `TestAnalyzeWorkflows_FrequentDivisorFindings`: 3 records with same review category, verify feedback produced targeting cobalt-crush
  - `TestAnalyzeWorkflows_NoPatterns`: 2 records with no patterns, verify empty feedback
  - `TestNextFeedbackID`: verify auto-increment from existing feedback
  - `TestSaveFeedback_LoadFeedback_RoundTrip`

**Checkpoint**: Learning loop extracts patterns from workflow history. Feedback is persisted and loadable.

---

## Phase 7: User Story 5 — Failure Modes (Priority: P3)

**Goal**: Implement failure handling in the orchestration engine: hero unavailable (skip), max iterations (escalate), acceptance rejection (new backlog item), inter-hero contradiction (surface to human).

**Independent Test**: Trigger each failure mode and verify the correct fallback behavior.

- [ ] T024 [US5] Add failure mode handling to `Orchestrator` in `internal/orchestration/engine.go`:
  - `handleHeroUnavailable(wf *WorkflowInstance, stageIdx int)` — marks stage as skipped with "hero unavailable" reason (already done in Start, but also handle mid-workflow detection)
  - `handleMaxIterations(wf *WorkflowInstance)` — when iteration_count >= 3 on review stage, escalate with summary of unresolved findings
  - `handleAcceptanceRejection(wf *WorkflowInstance, decision Decision)` — set workflow status, record rejection rationale
  - `handleContradiction(wf *WorkflowInstance, conflict string) error` — set status to escalated, record both perspectives in workflow state
- [ ] T025 [US5] Add failure mode tests to `internal/orchestration/engine_test.go`:
  - `TestOrchestrator_Advance_MaxIterations_Escalates`: set iteration_count to 3, advance review stage, verify escalation
  - `TestOrchestrator_HandleAcceptanceRejection`: simulate acceptance rejection, verify workflow status
  - `TestOrchestrator_HandleContradiction`: simulate conflicting guidance, verify escalation with both perspectives

**Checkpoint**: All 5 failure modes handled with correct fallback behavior. Tests cover each mode.

---

## Phase 8: Polish & Cross-Cutting

**Purpose**: Documentation, status updates, test suite verification.

- [ ] T026 Run `go test -race -count=1 ./...` and verify all tests pass including new orchestration tests. Fix any failures.
- [ ] T027 Run `go build ./...` and verify the build succeeds.
- [ ] T028 [P] Update `AGENTS.md`: add `internal/orchestration/` to project structure, add Spec 008 to Recent Changes, add `.opencode/command/workflow-*.md` commands to project structure, add `.opencode/skill/unbound-force-heroes/` to project structure.
- [ ] T029 [P] Update `specs/008-swarm-orchestration/spec.md`: change `status: draft` to `status: complete` in frontmatter and body.
- [ ] T030 [P] Update `README.md` if needed: note the `/workflow` commands are available.
- [ ] T031 Verify SC-001 through SC-008 success criteria from spec.md are met.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Phase 1 — models and store needed by all
- **US1 (Phase 3)**: Depends on Phase 2 — engine uses models and store
- **US2 (Phase 4)**: Depends on Phase 2 — artifact extensions. Can run in parallel with US1.
- **US3 (Phase 5)**: Depends on Phase 3 — commands reference workflow state produced by engine
- **US4 (Phase 6)**: Depends on Phase 3 — learning analyzes workflow records produced by engine
- **US5 (Phase 7)**: Depends on Phase 3 — failure modes are engine behaviors
- **Polish (Phase 8)**: Depends on all phases

### Parallel Opportunities

- **Phase 1**: T001-T004 sequential (same packages)
- **Phase 4**: Can run in parallel with Phase 3 (different files)
- **Phase 5**: T018-T021 all parallel (different files)
- **Phase 8**: T028-T030 parallel (different files)

---

## Implementation Strategy

### MVP First (Phases 1-3)

1. Phase 1: Setup (artifact context, package directory)
2. Phase 2: Models + store (foundational)
3. Phase 3: US1 — Core engine (Start, Advance, Complete)
4. **STOP and VALIDATE**: Create workflow, advance through stages, verify JSON persistence and workflow-record artifact

### Incremental Delivery

1. Phases 1-3 → Core engine works (MVP)
2. Phase 4 → Enhanced artifact discovery + schema checking
3. Phase 5 → Swarm skills + `/workflow` commands
4. Phase 6 → Learning feedback extraction
5. Phase 7 → Failure mode handling
6. Phase 8 → Documentation + validation

### Estimated Time

| Phase | Tasks | Est. Time |
|-------|-------|-----------|
| Phase 1: Setup | 4 | 30 min |
| Phase 2: Foundational | 3 | 45 min |
| Phase 3: US1 Engine | 6 | 2 hrs |
| Phase 4: US2 Artifacts | 4 | 1 hr |
| Phase 5: US3 Swarm | 4 | 1 hr |
| Phase 6: US4 Learning | 2 | 45 min |
| Phase 7: US5 Failures | 2 | 45 min |
| Phase 8: Polish | 6 | 30 min |
| **Total** | **31** | **~7 hours** |
<!-- scaffolded by unbound vdev -->
