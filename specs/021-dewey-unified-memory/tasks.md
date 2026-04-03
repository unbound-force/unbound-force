# Tasks: Dewey Unified Memory

**Input**: Design documents from `specs/021-dewey-unified-memory/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, quickstart.md

**Cross-Repo Note**: This spec spans 3 repositories.
Phase 1 (Dewey repo) and Phase 3 (Swarm fork) are
documented as prerequisite/follow-up phases with
specification tasks — not implementation tasks. Phase 2
(this repo) contains the implementation tasks that
modify files in this repository.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Dewey Repo Prerequisites (External — Specification Only)

**Purpose**: Document what the Dewey repo must implement
before Phase 2 can begin. These are NOT implementation
tasks — they describe issues/specs to create in the
`unbound-force/dewey` repository.

**⚠️ CRITICAL**: Phase 2 CANNOT begin until the Phase 1
gate passes: `dewey serve` auto-starts Ollama AND
`dewey_store_learning` persists a learning retrievable
via `dewey_semantic_search`.

- [ ] T001 [US1] Create issue in `unbound-force/dewey`: Ollama lifecycle management — `internal/ollama/` package with detect-then-start logic, health check polling (30s max, 500ms intervals), fire-and-forget exit semantics (FR-001 through FR-006, per Spec 021 research.md R1)
- [ ] T002 [US2] Create issue in `unbound-force/dewey`: Learning storage MCP tool — `dewey_store_learning` accepting `{ text, tags }`, persisting to `graph.db` with `source_type: "learning"`, generating embeddings via granite-embedding:30m, immediately searchable via `dewey_semantic_search` (FR-007 through FR-011, per Spec 021 data-model.md)
- [ ] T003 [US1] Create issue in `unbound-force/dewey`: Integration of Ollama auto-start into `dewey serve` startup — `cmd/dewey/serve.go` calls Ollama lifecycle check before accepting MCP tool calls (FR-003)

**Gate**: `dewey serve` auto-starts Ollama. `dewey_store_learning` stores a learning retrievable via `dewey_semantic_search`. Verified per quickstart.md V1–V5.

---

## Phase 2: Agent Migration — /unleash Command (US3, Priority: P1) 🎯 MVP

**Goal**: The `/unleash` retrospective step stores learnings via Dewey instead of Hivemind.

**Independent Test**: `grep -c "hivemind_store" .opencode/command/unleash.md` returns 0. `grep -c "dewey_store_learning" .opencode/command/unleash.md` returns ≥ 1.

### Implementation

- [ ] T004 [US3] Update `/unleash` retrospective step in `.opencode/command/unleash.md` — replace `hivemind_store` with `dewey_store_learning` in the learning storage call (FR-012). Replace `hivemind_store` tool existence check with `dewey_store_learning`. Update the learning payload to include `text` and `tags` fields per data-model.md schema.
- [ ] T005 [US3] Update `/unleash` graceful degradation in `.opencode/command/unleash.md` — replace "If Hivemind is available" / "If Hivemind is NOT available" / "Hivemind not available" messaging with Dewey equivalents (FR-014). Preserve the skip-with-note pattern.
- [ ] T006 [US3] Sync scaffold asset copy: copy updated `.opencode/command/unleash.md` to `internal/scaffold/assets/opencode/command/unleash.md` — both files MUST be identical (drift detection enforced by `TestEmbeddedAssets_MatchSource`).

**Checkpoint**: `grep -r "hivemind_store" .opencode/command/unleash.md` returns 0. Scaffold drift test passes.

---

## Phase 3: Agent Migration — Divisor Agents (US3, Priority: P1)

**Goal**: All 5 Divisor agents' Prior Learnings step queries Dewey instead of Hivemind.

**Independent Test**: `grep -c "hivemind_find" .opencode/agents/divisor-*.md` returns 0. Each file contains `dewey_semantic_search`.

### Implementation

- [ ] T007 [P] [US3] Update Prior Learnings step in `.opencode/agents/divisor-adversary.md` — replace `hivemind_find` with `dewey_semantic_search` for file-specific learning queries (FR-013). Replace "If Hivemind MCP tools are available" with "If Dewey MCP tools are available". Replace "If Hivemind is not available" with "If Dewey is not available".
- [ ] T008 [P] [US3] Update Prior Learnings step in `.opencode/agents/divisor-architect.md` — same migration as T007 (FR-013).
- [ ] T009 [P] [US3] Update Prior Learnings step in `.opencode/agents/divisor-guard.md` — same migration as T007 (FR-013).
- [ ] T010 [P] [US3] Update Prior Learnings step in `.opencode/agents/divisor-sre.md` — same migration as T007 (FR-013).
- [ ] T011 [P] [US3] Update Prior Learnings step in `.opencode/agents/divisor-testing.md` — same migration as T007 (FR-013).

### Scaffold Asset Sync

- [ ] T012 [P] [US3] Sync scaffold asset: copy `.opencode/agents/divisor-adversary.md` to `internal/scaffold/assets/opencode/agents/divisor-adversary.md`.
- [ ] T013 [P] [US3] Sync scaffold asset: copy `.opencode/agents/divisor-architect.md` to `internal/scaffold/assets/opencode/agents/divisor-architect.md`.
- [ ] T014 [P] [US3] Sync scaffold asset: copy `.opencode/agents/divisor-guard.md` to `internal/scaffold/assets/opencode/agents/divisor-guard.md`.
- [ ] T015 [P] [US3] Sync scaffold asset: copy `.opencode/agents/divisor-sre.md` to `internal/scaffold/assets/opencode/agents/divisor-sre.md`.
- [ ] T016 [P] [US3] Sync scaffold asset: copy `.opencode/agents/divisor-testing.md` to `internal/scaffold/assets/opencode/agents/divisor-testing.md`.

**Checkpoint**: `grep -r "hivemind_find" .opencode/agents/` returns 0. All 5 scaffold asset copies match live files. `go test -race -count=1 ./internal/scaffold/...` passes.

---

## Phase 4: Agent Migration — Cobalt-Crush (US3, Priority: P1)

**Goal**: Cobalt-Crush agent references Dewey as the unified memory layer, not "complementing Hivemind."

**Independent Test**: `grep -c "Hivemind" .opencode/agents/cobalt-crush-dev.md` returns 0.

### Implementation

- [ ] T017 [US3] Update Hivemind reference in `.opencode/agents/cobalt-crush-dev.md` line 190 — replace "complementing Hivemind's session-specific learnings" with Dewey-appropriate framing (e.g., "complementing Dewey's persistent learnings"). No tool call changes needed — Cobalt-Crush does not call Hivemind tools directly.
- [ ] T018 [US3] Sync scaffold asset: copy `.opencode/agents/cobalt-crush-dev.md` to `internal/scaffold/assets/opencode/agents/cobalt-crush-dev.md`.

**Checkpoint**: No Hivemind references in cobalt-crush-dev.md. Scaffold drift test passes.

---

## Phase 5: AGENTS.md Documentation (US6, Priority: P3)

**Goal**: AGENTS.md documents Dewey as the unified memory layer, superseding the "complements Hivemind" framing.

**Independent Test**: `grep -c "complements Hivemind" AGENTS.md` returns 0. `grep -c "unified memory" AGENTS.md` returns ≥ 1.

### Implementation

- [ ] T019 [US6] Update "Embedding Model Alignment" section in `AGENTS.md` (line ~552) — replace "To ensure Swarm's Hivemind uses the same model" framing with Dewey-centric framing. Dewey is now the unified embedding layer; Swarm's Hivemind is deprecated. Update the section to explain that Dewey manages Ollama lifecycle and provides unified semantic memory (learnings + specs + code + web docs).
- [ ] T020 [US6] Update Spec 020 entry in "Recent Changes" section of `AGENTS.md` (line ~626) — the phrase "Dewey complements Hivemind (Spec 019), not replaces it" is now superseded by Spec 021. Add a note or update the entry to reflect that Spec 021 supersedes this framing.
- [ ] T021 [US6] Add Spec 021 entry to "Recent Changes" section of `AGENTS.md` — document the unified memory migration: Dewey replaces Hivemind as semantic memory layer, Ollama lifecycle managed by Dewey, agent files migrated from `hivemind_store`/`hivemind_find` to Dewey equivalents, Swarm plugin forked.
- [ ] T022 [US6] Update "Active Technologies" section in `AGENTS.md` — add Dewey unified memory entry documenting `dewey_store_learning` MCP tool and learning persistence in `graph.db`. Update any Hivemind references in technology entries.

**Checkpoint**: No "complements Hivemind" in AGENTS.md. Unified memory documented. `grep -c "unified memory\|replaces Hivemind" AGENTS.md` returns ≥ 1.

---

## Phase 6: Doctor Health Check Enhancement (US5, Priority: P3)

**Goal**: `uf doctor` checks Dewey's embedding capability as part of the Dewey health check group.

**Independent Test**: `uf doctor` shows "embedding capability" check in the Dewey Knowledge Layer group.

### Implementation

- [ ] T023 [US5] Add `embedding capability` check to `checkDewey()` in `internal/doctor/checks.go` — after the existing `embedding model` check, add a new check that verifies Dewey can generate embeddings (FR-020). When Dewey binary is found AND Ollama is serving AND the model is pulled: PASS with "available — Dewey manages Ollama lifecycle". When Dewey is found but Ollama is not serving: WARN with "Dewey running but embeddings unavailable. Semantic search is keyword-only." Per research.md R5 decision.
- [ ] T024 [US5] Update `embedding model` check messaging in `checkDewey()` in `internal/doctor/checks.go` — reframe the existing check to note "Dewey manages Ollama lifecycle" in the pass message, per research.md R5 decision.
- [ ] T025 [US5] Add test for `checkDewey` embedding capability in `internal/doctor/doctor_test.go` — test PASS case (Dewey + Ollama + model all available), WARN case (Dewey available but Ollama not serving), and skip case (Dewey not found). Follow existing `TestCheckDewey_*` test patterns.

**Checkpoint**: `go test -race -count=1 ./internal/doctor/...` passes. New embedding capability check verified.

---

## Phase 7: Setup — Forked Swarm Plugin (US4/US5, Priority: P2/P3)

**Goal**: `uf setup` installs the forked Swarm plugin from `unbound-force/swarm`.

**Independent Test**: `grep -c "unbound-force" internal/setup/setup.go` shows forked package references.

### Implementation

- [ ] T026 [US4] Update Swarm plugin install command in `internal/setup/setup.go` — replace `opencode-swarm-plugin@latest` with `@unbound-force/opencode-swarm-plugin@latest` in both the bun and npm install paths (FR-019). Update dry-run messages to match.
- [ ] T027 [US4] Update Hivemind comment in `internal/setup/setup.go` (lines ~112, ~128) — replace "Swarm's Hivemind" references with updated framing reflecting Dewey as the unified memory layer.
- [ ] T028 [US4] Update Swarm install hints in `internal/doctor/environ.go` — replace `opencode-swarm-plugin@latest` with `@unbound-force/opencode-swarm-plugin@latest` in all install hint strings (3 occurrences at lines ~248, ~272, ~293) (FR-019).
- [ ] T029 [US4] Update Swarm plugin check in `internal/doctor/checks.go` — replace `opencode-swarm-plugin` string in the plugin array check and install hint (lines ~337, ~417, ~426, ~432) with `@unbound-force/opencode-swarm-plugin` (FR-019).
- [ ] T030 [US4] Update scaffolded `opencode.json` plugin entry in `internal/scaffold/scaffold.go` — replace `opencode-swarm-plugin` with `@unbound-force/opencode-swarm-plugin` in the plugin array write logic (lines ~572, ~579) (FR-019).
- [ ] T031 [US4] Update setup tests in `internal/setup/setup_test.go` — update all `opencode-swarm-plugin` references to `@unbound-force/opencode-swarm-plugin` in test assertions, mock commands, and fixture data.
- [ ] T032 [US4] Update doctor tests in `internal/doctor/doctor_test.go` — update all `opencode-swarm-plugin` references to `@unbound-force/opencode-swarm-plugin` in test assertions, fixture data, and install hint checks.
- [ ] T033 [US4] Update scaffold tests in `internal/scaffold/scaffold_test.go` — update all `opencode-swarm-plugin` references to `@unbound-force/opencode-swarm-plugin` in test assertions and fixture data.

**Checkpoint**: `go test -race -count=1 ./internal/setup/... ./internal/doctor/... ./internal/scaffold/...` passes. All references point to forked package.

---

## Phase 8: Regression Tests & Validation

**Purpose**: Add regression guard and run full validation.

- [ ] T034 [US3] Add `TestScaffoldOutput_NoHivemindReferences` regression test in `internal/scaffold/scaffold_test.go` — pattern from `TestScaffoldOutput_NoGraphthulhuReferences`. Walk all embedded scaffold assets and verify no `hivemind_store` or `hivemind_find` tool references remain. Allow `Hivemind` as a word in historical/contextual prose but reject active tool call references.
- [ ] T035 Run `make check` — full build, test, vet, lint validation. All tests must pass with zero regressions.
- [ ] T036 Run quickstart.md Phase B verification steps (V6–V10) — verify unleash.md, Divisor agents, graceful degradation, AGENTS.md, and scaffold drift detection.

**Checkpoint**: `make check` passes. Zero `hivemind_store` or `hivemind_find` references in scaffold assets. All quickstart.md Phase B verifications pass.

---

## Phase 9: Swarm Fork Follow-Up (External — Specification Only)

**Purpose**: Document what the Swarm fork must implement
after Phase 2 is complete. These are NOT implementation
tasks — they describe work to be done in the
`unbound-force/swarm` repository.

- [ ] T037 [US4] Fork upstream Swarm plugin repository to `unbound-force/swarm` via `gh repo fork` (FR-016).
- [ ] T038 [US4] Update `package.json` in fork: set name to `@unbound-force/opencode-swarm-plugin` (scoped package per data-model.md).
- [ ] T039 [US4] Modify Hivemind tools in fork to proxy through Dewey — `hivemind_store` calls `dewey_store_learning`, `hivemind_find` calls `dewey_semantic_search` (FR-018, per research.md R3 proxy approach).
- [ ] T040 [US4] Verify all existing Swarm tools function identically after fork (FR-017). Run upstream test suite. Zero regressions (SC-005).
- [ ] T041 [US4] Publish forked package or configure GitHub-based npm install (per research.md R4).

**Gate**: `swarm doctor` passes. All Swarm MCP tools functional. Forked plugin installable via npm/bun. Verified per quickstart.md V11–V12.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Dewey Repo)**: External prerequisite — BLOCKS all Phase 2+ work
- **Phase 2 (Unleash)**: Depends on Phase 1 gate — can start immediately after
- **Phase 3 (Divisor Agents)**: Depends on Phase 1 gate — can run in PARALLEL with Phase 2
- **Phase 4 (Cobalt-Crush)**: Depends on Phase 1 gate — can run in PARALLEL with Phases 2–3
- **Phase 5 (AGENTS.md)**: Depends on Phase 1 gate — can run in PARALLEL with Phases 2–4
- **Phase 6 (Doctor)**: No dependency on Phase 1 gate — can run in PARALLEL with Phases 2–5
- **Phase 7 (Setup/Fork refs)**: No dependency on Phase 1 gate — can run in PARALLEL with Phases 2–6
- **Phase 8 (Regression)**: Depends on ALL of Phases 2–7 being complete
- **Phase 9 (Swarm Fork)**: External follow-up — depends on Phase 8 completion

### Within-Phase Parallel Opportunities

- **Phase 3**: T007–T011 are [P] (5 different files). T012–T016 are [P] (5 different files). All 10 tasks can run in parallel.
- **Phase 5**: T019–T022 are sequential (same file: AGENTS.md).
- **Phase 6**: T023–T024 are sequential (same file: checks.go). T025 depends on T023–T024.
- **Phase 7**: T026–T027 sequential (setup.go), T028 parallel (environ.go), T029 parallel (checks.go), T030 parallel (scaffold.go). T031–T033 depend on T026–T030.

### Cross-Phase Parallel Opportunities

Phases 2, 3, 4, 5, 6, and 7 can ALL run in parallel since they modify different files:
- Phase 2: `.opencode/command/unleash.md` + scaffold copy
- Phase 3: `.opencode/agents/divisor-*.md` + scaffold copies
- Phase 4: `.opencode/agents/cobalt-crush-dev.md` + scaffold copy
- Phase 5: `AGENTS.md`
- Phase 6: `internal/doctor/checks.go`, `internal/doctor/doctor_test.go`
- Phase 7: `internal/setup/setup.go`, `internal/doctor/environ.go`, `internal/scaffold/scaffold.go`, `*_test.go`

**Exception**: Phase 7 T029 modifies `internal/doctor/checks.go` which Phase 6 T023–T024 also modify. These MUST be sequential or coordinated.

---

## Summary

| Metric | Count |
|--------|-------|
| **Total tasks** | 41 |
| **Phase 1 (Dewey repo — spec only)** | 3 |
| **Phase 2 (Unleash migration)** | 3 |
| **Phase 3 (Divisor migration)** | 10 |
| **Phase 4 (Cobalt-Crush migration)** | 2 |
| **Phase 5 (AGENTS.md docs)** | 4 |
| **Phase 6 (Doctor enhancement)** | 3 |
| **Phase 7 (Setup/fork refs)** | 8 |
| **Phase 8 (Regression/validation)** | 3 |
| **Phase 9 (Swarm fork — spec only)** | 5 |
| **This-repo implementation tasks** | 33 (T004–T036) |
| **External specification tasks** | 8 (T001–T003, T037–T041) |

### Per-Story Breakdown

| Story | Tasks | Priority |
|-------|-------|----------|
| US1 — Dewey Manages Ollama | 2 (T001, T003) | P1 (external) |
| US2 — Dewey Stores Learnings | 1 (T002) | P1 (external) |
| US3 — /unleash Uses Dewey | 21 (T004–T018, T034) | P1 |
| US4 — Fork Swarm Plugin | 13 (T026–T033, T037–T041) | P2 |
| US5 — Doctor/Setup Updates | 3 (T023–T025) | P3 |
| US6 — AGENTS.md Documentation | 4 (T019–T022) | P3 |

### Parallel Opportunities

- **Max parallelism (Phases 2–7)**: 6 phases can run concurrently
- **Phase 3 internal**: 10 tasks across 10 different files — full parallelism
- **Phase 7 internal**: T028, T029, T030 touch different files — 3-way parallel
- **Cross-phase conflict**: Phase 6 T023–T024 and Phase 7 T029 both modify `internal/doctor/checks.go` — must be sequential

<!-- spec-review: passed -->
