# Tasks: Hivemind-to-Dewey Memory Migration

**Input**: Design documents from `/specs/022-hivemind-dewey-migration/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, quickstart.md

**Tests**: One regression test is explicitly required (FR-007, data-model.md §Regression Test Specification).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Project type**: CLI / meta-repository (Go + Markdown)
- **Canonical agent files**: `.opencode/agents/`
- **Canonical command files**: `.opencode/command/`
- **Scaffold asset copies**: `internal/scaffold/assets/opencode/`
- **Regression tests**: `internal/scaffold/scaffold_test.go`
- **Documentation**: `AGENTS.md` (project root)

---

## Phase 1: Setup

**Purpose**: No project initialization needed — all target files and test infrastructure already exist. This phase verifies the working environment.

- [x] T001 Verify branch is `022-hivemind-dewey-migration` and all tests pass via `make check`

**⚠️ CRITICAL**: No user story work can begin until Phase 1 verification passes.

**Checkpoint**: Foundation ready — user story implementation can now begin in parallel

---

## Phase 2: User Story 1 — Unified Learning Storage in Autonomous Pipeline (Priority: P1) 🎯 MVP

**Goal**: Replace `hivemind_store` with `dewey_store_learning` in the `/unleash` retrospective step so learnings flow into Dewey.

**Independent Test**: `grep -c "hivemind_store" .opencode/command/unleash.md` returns 0; `grep -c "dewey_store_learning" .opencode/command/unleash.md` returns ≥ 1.

### Implementation for User Story 1

- [x] T002 [US1] Replace all `hivemind_store` tool references with `dewey_store_learning` in `.opencode/command/unleash.md` (retrospective step, ~lines 493-519 per data-model.md)
- [x] T003 [US1] Replace all `Hivemind` prose references with `Dewey` equivalents in `.opencode/command/unleash.md` (section headings, availability checks, hint messages — 7 replacements per data-model.md §Text Replacement Patterns)
- [x] T004 [US1] Verify graceful degradation wording: "If Dewey is available" / "If Dewey is NOT available" sections read correctly after replacement in `.opencode/command/unleash.md`

**Checkpoint**: `/unleash` retrospective step references Dewey exclusively. Zero `hivemind_store` or `Hivemind` references remain in unleash.md.

---

## Phase 3: User Story 2 — Unified Learning Retrieval in Review Agents (Priority: P1)

**Goal**: Replace `hivemind_find` with `dewey_semantic_search` in the Prior Learnings step of all five Divisor review agents.

**Independent Test**: `grep -c "hivemind_find" .opencode/agents/divisor-*.md` returns 0 for each file; `grep -c "dewey_semantic_search" .opencode/agents/divisor-*.md` returns ≥ 1 for each file.

### Implementation for User Story 2

- [x] T005 [P] [US2] Replace `hivemind_find` with `dewey_semantic_search` and `Hivemind` with `Dewey` in Step 0: Prior Learnings of `.opencode/agents/divisor-adversary.md` (3 replacements per data-model.md §Divisor Agent Prior Learnings Step)
- [x] T006 [P] [US2] Replace `hivemind_find` with `dewey_semantic_search` and `Hivemind` with `Dewey` in Step 0: Prior Learnings of `.opencode/agents/divisor-architect.md` (3 replacements)
- [x] T007 [P] [US2] Replace `hivemind_find` with `dewey_semantic_search` and `Hivemind` with `Dewey` in Step 0: Prior Learnings of `.opencode/agents/divisor-guard.md` (3 replacements)
- [x] T008 [P] [US2] Replace `hivemind_find` with `dewey_semantic_search` and `Hivemind` with `Dewey` in Step 0: Prior Learnings of `.opencode/agents/divisor-sre.md` (3 replacements)
- [x] T009 [P] [US2] Replace `hivemind_find` with `dewey_semantic_search` and `Hivemind` with `Dewey` in Step 0: Prior Learnings of `.opencode/agents/divisor-testing.md` (3 replacements)

**Checkpoint**: All five Divisor agents reference Dewey exclusively in Prior Learnings. Zero `hivemind_find` or `Hivemind` references remain in any `divisor-*.md` file.

---

## Phase 4: User Story 3 — Graceful Degradation (Priority: P2)

**Goal**: Verify that the replacement text in US1 and US2 preserves the 3-tier graceful degradation pattern (Full Dewey, Dewey without learning storage, No Dewey).

**Independent Test**: Read each modified file and confirm the degradation tiers are present and correctly worded.

### Implementation for User Story 3

- [x] T010 [US3] Verify `/unleash` retrospective step in `.opencode/command/unleash.md` has correct 3-tier degradation: (1) `dewey_store_learning` available → store, (2) Dewey available but tool missing → warn and display, (3) Dewey unavailable → warn and display
- [x] T011 [US3] Verify each Divisor agent's Prior Learnings step has correct 2-tier degradation: (1) `dewey_semantic_search` available → query, (2) Dewey unavailable → skip with informational note

**Checkpoint**: Graceful degradation wording is correct in all 6 modified files.

---

## Phase 5: User Story 4 — Documentation Accuracy (Priority: P2)

**Goal**: Update project documentation to describe Dewey as the unified memory layer, superseding the "Dewey complements Hivemind" framing.

**Independent Test**: `grep -n "Hivemind" AGENTS.md` returns only historical Recent Changes entries. `grep -n "Hivemind" internal/setup/setup.go` returns 0 matches.

### Implementation for User Story 4

- [x] T012 [P] [US4] Update "Embedding Model Alignment" section in `AGENTS.md` (~line 556): replace "To ensure Swarm's Hivemind uses the same model" with "To ensure all tools use the same model" (per data-model.md §AGENTS.md Documentation)
- [x] T013 [P] [US4] Update Spec 020 Recent Changes entry in `AGENTS.md` (~line 628): replace "Dewey complements Hivemind (Spec 019), not replaces it" with "Dewey is the unified memory layer (superseded by Spec 022)" (per data-model.md §AGENTS.md Documentation)
- [x] T014 [P] [US4] Update comment on ~line 112 in `internal/setup/setup.go`: replace "Setting these env vars aligns Swarm's Hivemind" with "Setting these env vars aligns all tools" (per data-model.md §setup.go Comments)
- [x] T015 [P] [US4] Update comment on ~line 128 in `internal/setup/setup.go`: replace "Set Ollama env vars so Swarm's Hivemind uses the same" with "Set Ollama env vars so all embedding consumers use the same" (per data-model.md §setup.go Comments)
- [x] T016 [P] [US4] Update Hivemind prose reference in `.opencode/agents/cobalt-crush-dev.md` (~line 190): replace "complementing Hivemind's session-specific learnings" with Dewey-unified framing (per data-model.md §Cobalt-Crush Agent Reference)
- [x] T017 [US4] Add Spec 022 Recent Changes entry to `AGENTS.md` summarizing the migration

**Checkpoint**: Documentation accurately describes Dewey as the unified memory layer. No active Hivemind references remain outside historical Recent Changes entries.

---

## Phase 6: User Story 5 — Scaffold Consistency (Priority: P3)

**Goal**: Synchronize all 6 scaffold asset copies with their canonical sources and add a regression test to prevent Hivemind tool references from reappearing.

**Independent Test**: `make check` passes — drift detection test (`TestEmbeddedAssets_MatchSource`) and new regression test (`TestScaffoldOutput_NoHivemindReferences`) both pass.

### Implementation for User Story 5

- [x] T018 [P] [US5] Copy canonical `.opencode/command/unleash.md` to `internal/scaffold/assets/opencode/command/unleash.md` (scaffold asset sync for FR-006)
- [x] T019 [P] [US5] Copy canonical `.opencode/agents/divisor-adversary.md` to `internal/scaffold/assets/opencode/agents/divisor-adversary.md` (scaffold asset sync for FR-006)
- [x] T020 [P] [US5] Copy canonical `.opencode/agents/divisor-architect.md` to `internal/scaffold/assets/opencode/agents/divisor-architect.md` (scaffold asset sync for FR-006)
- [x] T021 [P] [US5] Copy canonical `.opencode/agents/divisor-guard.md` to `internal/scaffold/assets/opencode/agents/divisor-guard.md` (scaffold asset sync for FR-006)
- [x] T022 [P] [US5] Copy canonical `.opencode/agents/divisor-sre.md` to `internal/scaffold/assets/opencode/agents/divisor-sre.md` (scaffold asset sync for FR-006)
- [x] T023 [P] [US5] Copy canonical `.opencode/agents/divisor-testing.md` to `internal/scaffold/assets/opencode/agents/divisor-testing.md` (scaffold asset sync for FR-006)
- [x] T024 [US5] Copy canonical `.opencode/agents/cobalt-crush-dev.md` to `internal/scaffold/assets/opencode/agents/cobalt-crush-dev.md` (scaffold asset sync — prose reference updated in T016)
- [x] T025 [US5] Add `TestScaffoldOutput_NoHivemindReferences` regression test in `internal/scaffold/scaffold_test.go` following the `TestScaffoldOutput_NoGraphthulhuReferences` pattern — stale patterns: `hivemind_store`, `hivemind_find`, `hivemind_validate`, `hivemind_remove`, `hivemind_get` (per data-model.md §Regression Test Specification)
- [x] T026 [US5] Run `make check` to verify all tests pass: drift detection, new regression test, and existing regression guards

**Checkpoint**: All scaffold assets synchronized. Regression test prevents Hivemind tool references from reappearing in scaffolded output.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final validation across all user stories

- [x] T027 Run full test suite via `make check` — all tests must pass including drift detection and both regression guards (NoGraphthulhuReferences, NoHivemindReferences)
- [x] T028 Run quickstart.md Phase B manual verification steps (B1–B4) to confirm zero stale references across all modified files
- [x] T029 Validate documentation impact: verify `AGENTS.md`, `README.md`, `unbound-force.md`, and spec artifacts are up to date per Documentation Validation Gate

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **US1 (Phase 2)** and **US2 (Phase 3)**: Can proceed in parallel after Phase 1 — they modify different files
- **US3 (Phase 4)**: Depends on US1 and US2 completion (verifies their output)
- **US4 (Phase 5)**: Can proceed in parallel with US1/US2/US3 — modifies different files (AGENTS.md, setup.go, cobalt-crush-dev.md)
- **US5 (Phase 6)**: Depends on US1, US2, and US4 completion (copies canonical files that were modified in those phases)
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: Independent — modifies only `unleash.md`
- **US2 (P1)**: Independent — modifies only `divisor-*.md` files
- **US3 (P2)**: Depends on US1 + US2 (verification of their output)
- **US4 (P2)**: Independent — modifies `AGENTS.md`, `setup.go`, `cobalt-crush-dev.md`
- **US5 (P3)**: Depends on US1 + US2 + US4 (copies their modified files to scaffold assets)

### Within Each User Story

- US2 tasks (T005–T009) are all [P] — they modify different files with no cross-dependencies
- US4 tasks (T012–T016) are all [P] — they modify different files; T017 is sequential (depends on knowing what changed)
- US5 tasks (T018–T024) are all [P] for file copies; T025 (regression test) is sequential; T026 (verification) depends on all prior US5 tasks

### Parallel Opportunities

```text
After Phase 1:
  ┌─ US1 (Phase 2): T002–T004 (unleash.md)
  │
  ├─ US2 (Phase 3): T005–T009 in parallel (5 divisor-*.md files)
  │
  └─ US4 (Phase 5): T012–T016 in parallel (AGENTS.md, setup.go, cobalt-crush-dev.md)

After US1 + US2:
  └─ US3 (Phase 4): T010–T011 (verification)

After US1 + US2 + US4:
  └─ US5 (Phase 6): T018–T024 in parallel (6 file copies + 1 cobalt-crush copy)
                     T025 sequential (regression test)
                     T026 sequential (make check)

After all:
  └─ Polish (Phase 7): T027–T029
```

---

## Implementation Strategy

### MVP First (US1 + US2)

1. Complete Phase 1: Setup verification
2. Complete Phase 2: US1 — `/unleash` retrospective migration
3. Complete Phase 3: US2 — Divisor agent Prior Learnings migration
4. **STOP and VALIDATE**: Verify zero `hivemind_store`/`hivemind_find` in canonical files
5. These two stories deliver the core value — unified learning storage and retrieval

### Incremental Delivery

1. US1 + US2 → Core migration complete (P1 stories)
2. US3 → Graceful degradation verified (P2)
3. US4 → Documentation accurate (P2)
4. US5 → Scaffold synchronized + regression guard (P3)
5. Each story adds confidence without breaking previous stories

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- This is a Markdown-only migration (plus 1 Go test and 2 Go comment edits) — no runtime behavior changes
- The regression test (T025) follows the exact pattern of `TestScaffoldOutput_NoGraphthulhuReferences` in `internal/scaffold/scaffold_test.go`
- Scaffold drift detection (`TestEmbeddedAssets_MatchSource`) will automatically catch any unsynchronized assets — T026 is the verification gate
- Commit after each phase or logical group
- Stop at any checkpoint to validate independently

<!-- spec-review: passed -->
