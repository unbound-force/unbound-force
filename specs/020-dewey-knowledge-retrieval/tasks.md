# Tasks: Dewey Knowledge Retrieval

**Input**: Design documents from `/specs/020-dewey-knowledge-retrieval/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, quickstart.md

**Tests**: No new tests required — existing scaffold drift detection tests (`TestScaffoldOutput_*`) verify all modified files are synced to `internal/scaffold/assets/`. No Go code changes.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story. All changes are Markdown edits + scaffold asset copies.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: User Story 1 — AGENTS.md Knowledge Retrieval Convention (Priority: P1) 🎯 MVP

**Goal**: Add a project-wide "Knowledge Retrieval" section to AGENTS.md that instructs all agents to prefer Dewey MCP tools over grep/glob for cross-repo context, design decisions, and architectural patterns.

**Independent Test**: Open AGENTS.md, search for "Knowledge Retrieval" section between "Coding Conventions" and "Testing Conventions". Verify it contains tool selection matrix, fallback guidance, and 3-tier graceful degradation pattern.

### Implementation for User Story 1

- [x] T001 [US1] Add "Knowledge Retrieval" section to `AGENTS.md` between "Coding Conventions" (line ~462) and "Testing Conventions" (line ~463) with: (a) behavioral instruction to prefer Dewey MCP tools over grep/glob for cross-repo context, (b) tool selection matrix mapping query types to Dewey tools (`dewey_semantic_search` for conceptual, `dewey_search` for keyword, `dewey_get_page` for specific pages, `dewey_find_connections` for relationships), (c) fallback criteria (when to use grep/glob/read instead), (d) 3-tier graceful degradation pattern (Full Dewey, Graph-only, No Dewey) per data-model.md
- [x] T002 [US1] Update "Recent Changes" section in `AGENTS.md` with a summary entry for 020-dewey-knowledge-retrieval describing all changes made by this spec

**Checkpoint**: AGENTS.md contains the Knowledge Retrieval convention. `grep -c "Knowledge Retrieval" AGENTS.md` returns at least 2 (section heading + recent changes entry).

---

## Phase 2: User Story 2 — Cobalt-Crush Knowledge Step (Priority: P1) 🎯 MVP

**Goal**: Enhance Cobalt-Crush's existing Knowledge Retrieval section with a "Step 0: Knowledge Retrieval" instruction that fires before code exploration, querying Dewey for prior learnings, related specs, and file-specific context.

**Independent Test**: Read `.opencode/agents/cobalt-crush-dev.md`, verify the Knowledge Retrieval section includes Step 0 with prior learnings queries, related spec queries, architectural pattern queries, and role-specific examples per data-model.md.

### Implementation for User Story 2

- [x] T003 [US2] Enhance the existing "Knowledge Retrieval" section in `.opencode/agents/cobalt-crush-dev.md` (line ~178) with: (a) "Step 0: Knowledge Retrieval" instruction that fires before reading source documents, (b) queries for prior learnings about target files (`dewey_semantic_search` for file-specific context), (c) queries for related specs governing the feature (`dewey_search` for spec references), (d) queries for architectural patterns from conventions (`dewey_find_by_tag` for convention-tagged content), (e) role-specific example queries per data-model.md ("scaffold.go patterns", "FR-001 implementation")
- [x] T004 [US2] Update the "Source Documents" section (line ~14) in `.opencode/agents/cobalt-crush-dev.md` to reference the Knowledge Retrieval step as a prerequisite step before reading source documents
- [x] T005 [US2] Sync `.opencode/agents/cobalt-crush-dev.md` to `internal/scaffold/assets/opencode/agents/cobalt-crush-dev.md` (copy live file to scaffold asset)

**Checkpoint**: `grep -c "dewey_semantic_search" .opencode/agents/cobalt-crush-dev.md` returns at least 2. Scaffold asset matches live file.

---

## Phase 3: User Story 3 — Speckit Pipeline Dewey Integration (Priority: P2)

**Goal**: Add Dewey query steps to three Speckit commands (`/speckit.specify`, `/speckit.plan`, `/speckit.tasks`) so they discover related existing specs, prior research decisions, and implementation patterns before generating new artifacts.

**Independent Test**: Read each Speckit command file and verify it contains a Dewey query step with graceful degradation. `grep -c "dewey_" .opencode/command/speckit.specify.md` returns at least 1.

### Implementation for User Story 3

- [x] T006 [P] [US3] Add Dewey query step to `.opencode/command/speckit.specify.md` in the Outline section (after line ~24): before generating the spec, query `dewey_semantic_search` for existing specs with similar topics and reference discovered specs in the Dependencies section. Include graceful degradation (skip if Dewey unavailable).
- [x] T007 [P] [US3] Add Dewey query step to `.opencode/command/speckit.plan.md` in Phase 0: Outline & Research (after line ~43): during research, query `dewey_search` for prior research decisions in related specs (research.md files). Include graceful degradation (skip if Dewey unavailable).
- [x] T008 [P] [US3] Add Dewey query step to `.opencode/command/speckit.tasks.md` in the Outline section (after line ~25): before generating tasks, query `dewey_semantic_search_filtered` for implementation patterns from completed specs. Include graceful degradation (skip if Dewey unavailable).
- [x] T009 [US3] Sync `.opencode/command/speckit.specify.md` to `internal/scaffold/assets/opencode/command/speckit.specify.md` (copy live file to scaffold asset)
- [x] T010 [US3] Sync `.opencode/command/speckit.plan.md` to `internal/scaffold/assets/opencode/command/speckit.plan.md` (copy live file to scaffold asset)
- [x] T011 [US3] Sync `.opencode/command/speckit.tasks.md` to `internal/scaffold/assets/opencode/command/speckit.tasks.md` (copy live file to scaffold asset)

**Checkpoint**: All three Speckit command files contain Dewey query steps. Scaffold assets match live files. `make test` passes.

---

## Phase 4: User Story 4 — All Hero Agents Knowledge Step (Priority: P3)

**Goal**: Enhance the existing Knowledge Retrieval sections in Muti-Mind, Mx F, and Gaze agent files with role-appropriate "prefer Dewey" behavioral instructions, extended tool examples, and the "Step 0" pattern.

**Independent Test**: Read each hero agent file and verify its Knowledge Retrieval section includes role-specific query examples per data-model.md and the "prefer Dewey" behavioral instruction.

### Implementation for User Story 4

- [x] T012 [P] [US4] Enhance the existing "Knowledge Retrieval" section in `.opencode/agents/muti-mind-po.md` (line ~79) with: (a) "prefer Dewey" behavioral instruction for backlog queries, (b) role-specific extended tool examples per data-model.md (backlog patterns, acceptance history, `dewey_find_by_tag` for backlog tags, `dewey_query_properties` for item status), (c) explicit "Step 0" instruction to query Dewey before rendering acceptance judgments
- [x] T013 [P] [US4] Enhance the existing "Knowledge Retrieval" section in `.opencode/agents/mx-f-coach.md` (line ~86) with: (a) "prefer Dewey" behavioral instruction for metrics and process queries, (b) role-specific extended tool examples per data-model.md (velocity trends, retrospective outcomes, coaching patterns, `dewey_find_by_tag` for retrospective tags), (c) explicit "Step 0" instruction to query Dewey before coaching sessions
- [x] T014 [P] [US4] Enhance the existing "Knowledge Retrieval" section in `.opencode/agents/gaze-reporter.md` (line ~353) with: (a) "prefer Dewey" behavioral instruction for quality and test queries, (b) role-specific extended tool examples per data-model.md (CRAP score patterns, quality baselines, test patterns, `dewey_find_by_tag` for quality tags), (c) explicit "Step 0" instruction to query Dewey before producing quality reports
- [x] T015 [US4] Sync `.opencode/agents/muti-mind-po.md` to `internal/scaffold/assets/opencode/agents/muti-mind-po.md` (copy live file to scaffold asset)
- [x] T016 [US4] Sync `.opencode/agents/mx-f-coach.md` to `internal/scaffold/assets/opencode/agents/mx-f-coach.md` (copy live file to scaffold asset)
- [x] T017 [US4] Sync `.opencode/agents/gaze-reporter.md` to `internal/scaffold/assets/opencode/agents/gaze-reporter.md` (copy live file to scaffold asset)

**Checkpoint**: All three hero agent files have enhanced Knowledge Retrieval sections. Scaffold assets match live files. `make test` passes.

---

## Phase 5: User Story 5 — Unbound Force Heroes Skill Update (Priority: P3)

**Goal**: Update the `unbound-force-heroes` skill to include Dewey as the knowledge retrieval layer in the hero lifecycle workflow documentation, instructing Swarm coordinators to query Dewey for context before each stage.

**Independent Test**: Read `.opencode/skill/unbound-force-heroes/SKILL.md` and verify it mentions Dewey knowledge retrieval as a step in the hero lifecycle, with instructions for Swarm coordinators.

### Implementation for User Story 5

- [x] T018 [US5] Update `.opencode/skill/unbound-force-heroes/SKILL.md` to add a "Knowledge Retrieval" subsection in the hero lifecycle documentation that: (a) instructs Swarm coordinators to query Dewey for relevant context before delegating to hero agents at each stage, (b) specifies which Dewey tools to use per stage (semantic search for define, keyword search for implement, etc.), (c) includes graceful degradation (skip if Dewey unavailable)
- [x] T019 [US5] Sync `.opencode/skill/unbound-force-heroes/SKILL.md` to `internal/scaffold/assets/opencode/skill/unbound-force-heroes/SKILL.md` (copy live file to scaffold asset)

**Checkpoint**: Heroes skill mentions Dewey knowledge retrieval in the lifecycle. Scaffold asset matches live file.

---

## Phase 6: Final Validation & Documentation

**Purpose**: Cross-cutting validation that all changes are consistent and synced.

- [x] T020 Run `make test` (or `go test -race -count=1 ./...`) to verify all scaffold drift detection tests pass — confirms all 8 scaffold asset copies match their live files
- [x] T021 Verify scaffold file count is unchanged: `find internal/scaffold/assets -type f | wc -l` must equal 49 (no files added or removed per FR-010)
- [x] T022 Run quickstart.md smoke test sequence: (a) `grep -c "Knowledge Retrieval" AGENTS.md` ≥ 2, (b) `grep -c "dewey_semantic_search" .opencode/agents/cobalt-crush-dev.md` ≥ 2, (c) `grep -c "dewey_" .opencode/command/speckit.specify.md` ≥ 1, (d) `diff` each live file against its scaffold asset copy shows no differences

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (US1 — AGENTS.md)**: No dependencies — can start immediately. This is the MVP.
- **Phase 2 (US2 — Cobalt-Crush)**: No dependency on Phase 1 (different file). Can run in parallel with Phase 1.
- **Phase 3 (US3 — Speckit commands)**: No dependency on Phase 1 or 2 (different files). Can run in parallel.
- **Phase 4 (US4 — All hero agents)**: No dependency on Phase 1-3 (different files). Can run in parallel.
- **Phase 5 (US5 — Heroes skill)**: No dependency on Phase 1-4 (different file). Can run in parallel.
- **Phase 6 (Validation)**: Depends on ALL previous phases — runs last.

### Within Each Phase

- Tasks marked [P] within a phase can run in parallel (they modify different files).
- Scaffold sync tasks (T005, T009-T011, T015-T017, T019) MUST run after their corresponding edit tasks.
- T002 (AGENTS.md Recent Changes) SHOULD run last within Phase 1 to capture the final summary.

### Parallel Opportunities

**Maximum parallelism**: Phases 1-5 can ALL run in parallel since they modify entirely different files:
- Phase 1: `AGENTS.md`
- Phase 2: `.opencode/agents/cobalt-crush-dev.md`
- Phase 3: `.opencode/command/speckit.{specify,plan,tasks}.md`
- Phase 4: `.opencode/agents/{muti-mind-po,mx-f-coach,gaze-reporter}.md`
- Phase 5: `.opencode/skill/unbound-force-heroes/SKILL.md`

Within Phase 3, T006/T007/T008 are [P] (different files). Within Phase 4, T012/T013/T014 are [P] (different files).

---

## Implementation Strategy

### MVP First (User Stories 1 + 2 Only)

1. Complete Phase 1: AGENTS.md Knowledge Retrieval section
2. Complete Phase 2: Cobalt-Crush enhanced Knowledge Step
3. **STOP and VALIDATE**: Run `make test`, verify scaffold sync
4. Deploy/demo — the two highest-leverage integration points are live

### Incremental Delivery

1. Phase 1 + 2 → MVP (AGENTS.md + Cobalt-Crush) → Validate
2. Add Phase 3 → Speckit commands → Validate
3. Add Phase 4 + 5 → All heroes + skill → Validate
4. Phase 6 → Final validation → Done

### Full Parallel Strategy

With multiple agents:
1. Agent A: Phase 1 (AGENTS.md)
2. Agent B: Phase 2 (Cobalt-Crush)
3. Agent C: Phase 3 (Speckit commands — 3 files)
4. Agent D: Phase 4 (Hero agents — 3 files)
5. Agent E: Phase 5 (Heroes skill)
6. All complete → Phase 6 validation

---

## Notes

- All changes are Markdown edits — no Go code, no new files, no file count changes
- Scaffold file count MUST remain 49 (FR-010)
- Every modified live file has a corresponding scaffold asset copy that must be synced
- The "prefer Dewey" instruction is SHOULD (soft preference), not MUST (hard requirement)
- Dewey and Hivemind are complementary — this spec adds Dewey alongside Hivemind, not replacing it
- Divisor agents already have "Prior Learnings" (Spec 019) — this spec does NOT modify Divisor agents

<!-- spec-review: passed -->
