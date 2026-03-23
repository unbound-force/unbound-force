# Tasks: Dewey Integration

**Input**: Design documents from `/specs/015-dewey-integration/`
**Prerequisites**: plan.md, spec.md, research.md, contracts/scaffold-changes.md, quickstart.md

**Tests**: Tests are included -- the spec requires zero regressions (SC-005) and a regression test for stale graphthulhu/knowledge-graph references (SC-001). The constitution mandates testability (Principle IV).

**Organization**: Tasks are grouped by user story. US1-US3 are all P1 and tightly coupled (scaffold, agent files, and degradation must ship together), so they share phases but are labeled separately.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Phase 1: Setup (MCP Config Update)

**Purpose**: Replace the graphthulhu MCP server entry with Dewey in both the live config and the scaffold template.

- [x] T001 Update `opencode.json` at the repo root: replace the `knowledge-graph` MCP server entry with a `dewey` entry using command `["dewey", "serve", "--vault", "."]` per the contract in `specs/015-dewey-integration/contracts/scaffold-changes.md`
- [x] T002 Update the scaffold template `internal/scaffold/assets/opencode.json`: replace the `knowledge-graph` MCP server entry with the same `dewey` entry as T001
- [x] T003 Run `go build ./...` to verify the scaffold asset embeds correctly

---

## Phase 2: Foundational (Agent File Tool Reference Rename)

**Purpose**: Replace all `knowledge-graph_*` tool name references with `dewey_*` across agent files. MUST complete before user story work begins (which adds new semantic search content to these same files).

- [x] T004 Search all live agent files under `.opencode/agents/` for `knowledge-graph_` tool references. Record every file and line that matches
- [x] T005 Replace all `knowledge-graph_` tool name prefixes with `dewey_` in `.opencode/agents/muti-mind-po.md`
- [x] T006 [P] Replace all `knowledge-graph_` tool name prefixes with `dewey_` in `.opencode/agents/cobalt-crush-dev.md`
- [x] T007 [P] Replace all `knowledge-graph_` tool name prefixes with `dewey_` in `.opencode/agents/mx-f-coach.md`
- [x] T008 [P] Replace all `knowledge-graph_` tool name prefixes with `dewey_` in `.opencode/agents/constitution-check.md`
- [x] T009 Replace all `knowledge-graph_` or `graphthulhu` references in any remaining `.opencode/agents/divisor-*.md` and `.opencode/agents/gaze-reporter.md` files with `dewey_` equivalents
- [x] T010 Search all scaffold agent assets under `internal/scaffold/assets/opencode/agents/` for `knowledge-graph_` or `graphthulhu` references and replace with `dewey_` equivalents
- [x] T011 Search `.opencode/command/` files for `knowledge-graph_` or `graphthulhu` references and replace with `dewey_` equivalents
- [x] T012 Run `go build ./...` and `go test -race -count=1 ./internal/scaffold/...` to verify scaffold embeds and existing tests pass

**Checkpoint**: Zero `knowledge-graph_` or `graphthulhu` references remain in any live or scaffold agent file.

---

## Phase 3: User Story 1 - Scaffold Generates Dewey Config (Priority: P1)

**Goal**: `uf init` in a fresh directory produces 100% Dewey references with zero graphthulhu/knowledge-graph references.

**Independent Test**: Run `uf init` in a temp dir, grep all generated files for `graphthulhu` and `knowledge-graph_`, verify zero matches.

- [x] T013 [US1] Write `TestScaffoldOutput_NoGraphthulhuReferences` in `internal/scaffold/scaffold_test.go`: run scaffold in a temp dir, read all generated files, search for `graphthulhu` and `knowledge-graph_` patterns, assert zero matches (FR-001/FR-002/SC-001 regression guard)
- [x] T014 [US1] Run `go test -race -count=1 ./internal/scaffold/...` to verify T013 passes (all graphthulhu references already removed in Phase 2)
- [x] T015 [US1] Verify the scaffold file count test assertion in `cmd/unbound-force/main_test.go` still passes (no new files added, just content changes)

**Checkpoint**: Scaffold output is 100% Dewey. Regression test permanently guards against stale references.

---

## Phase 4: User Story 2 - Hero Agents Use Dewey Tools (Priority: P1)

**Goal**: All 5 hero agent persona files include Dewey semantic search instructions with role-specific query examples.

**Independent Test**: Read each hero's live agent file. Verify it references `dewey_semantic_search`. Verify it includes role-specific examples.

- [x] T016 [US2] Add a "Knowledge Retrieval" section to `.opencode/agents/muti-mind-po.md` with `dewey_semantic_search` usage and Product Owner-specific examples: cross-repo backlog patterns, past acceptance criteria, issue discovery across whitelisted repos
- [x] T017 [P] [US2] Add a "Knowledge Retrieval" section to `.opencode/agents/cobalt-crush-dev.md` with `dewey_semantic_search` usage and Developer-specific examples: toolstack API docs, implementation patterns from other repos, similar code in the org
- [x] T018 [P] [US2] Add a "Knowledge Retrieval" section to `.opencode/agents/gaze-reporter.md` with `dewey_semantic_search` usage and Tester-specific examples: test quality patterns, CRAP score baselines from other repos, known failure modes
- [x] T019 [P] [US2] Add a "Knowledge Retrieval" section to `.opencode/agents/mx-f-coach.md` with `dewey_semantic_search` usage and Manager-specific examples: cross-repo velocity trends, retrospective outcomes, coaching patterns
- [x] T020 [US2] Add Dewey tool awareness to all 5 `.opencode/agents/divisor-*.md` files: add a brief note in each persona's context section that `dewey_semantic_search` can be used to find cross-repo review patterns and convention violations

**Checkpoint**: All 5 hero roles have Dewey semantic search instructions with role-specific examples.

---

## Phase 5: User Story 3 - Graceful Degradation (Priority: P1)

**Goal**: Every hero agent file includes a 3-tier fallback pattern. No hero breaks when Dewey is unavailable.

**Independent Test**: Read each hero's agent file. Verify it contains the 3-tier pattern (full Dewey, graph-only, no Dewey).

- [x] T021 [US3] Add the 3-tier fallback pattern to the "Knowledge Retrieval" section in `.opencode/agents/muti-mind-po.md`: Tier 3 (semantic + structured), Tier 2 (structured only, no embedding model), Tier 1 (Read + Grep + convention packs)
- [x] T022 [P] [US3] Add the 3-tier fallback pattern to `.opencode/agents/cobalt-crush-dev.md`
- [x] T023 [P] [US3] Add the 3-tier fallback pattern to `.opencode/agents/gaze-reporter.md`
- [x] T024 [P] [US3] Add the 3-tier fallback pattern to `.opencode/agents/mx-f-coach.md`
- [x] T025 [US3] Add the 3-tier fallback pattern to all 5 `.opencode/agents/divisor-*.md` files (brief version since reviewers are less context-dependent)
- [x] T026 [US3] Update the scaffold template copies under `internal/scaffold/assets/opencode/agents/` to match the live agent files updated in T016-T025 (keep scaffold assets in sync)

**Checkpoint**: All hero agent files have 3-tier degradation. Constitution Principle II (Composability First) is satisfied.

---

## Phase 6: User Story 4 - Environment Health Checks (Priority: P2)

**Goal**: `uf doctor` includes a Dewey health check group. `uf setup` installs Dewey.

**Independent Test**: Run `uf doctor` with Dewey installed vs not installed. Verify correct pass/fail with fix hints.

- [x] T027 [US4] Add a Dewey health check group to `internal/doctor/checks.go`: check for `dewey` binary via `LookPath`, check for the embedding model via Ollama, check for `.dewey/` workspace directory. Include fix hints per the contract in `specs/015-dewey-integration/contracts/scaffold-changes.md`
- [x] T028 [US4] Add Dewey health check tests to `internal/doctor/doctor_test.go`: test all 3 checks passing, test dewey binary missing (with hint), test embedding model missing (with graph-only note), test workspace missing (with `dewey init` hint). Verify skipped checks when dewey binary is absent
- [x] T029 [US4] Add a Dewey installation step to `internal/setup/setup.go`: install `dewey` via Homebrew if not present, pull the embedding model via Ollama if not present. Position after the Swarm plugin step and before `uf init`. Include the same interactive guard used for swarm setup (`!opts.YesFlag && !opts.IsTTY()` skip pattern)
- [x] T030 [US4] Add Dewey setup tests to `internal/setup/setup_test.go`: test installation when missing, test skip when already installed, test embedding model pull
- [x] T031 [US4] Run `go test -race -count=1 ./internal/doctor/... ./internal/setup/...` to verify all new tests pass

**Checkpoint**: `uf doctor` reports Dewey health. `uf setup` installs Dewey components.

---

## Phase 7: User Story 5 - Swarm Embedding Model Update (Priority: P2)

**Goal**: Doctor and setup reference the enterprise-grade embedding model instead of the previous default.

**Independent Test**: Run `uf doctor` and verify the Ollama check references the correct model name.

- [x] T032 [US5] Update the Ollama model name constant in `internal/doctor/checks.go`: change the expected model from the previous default to the enterprise-grade model name (same model Dewey uses as default)
- [x] T033 [US5] Update the Ollama model pull command in `internal/setup/setup.go`: change the model name in the Ollama tip message (line ~235) from the previous default to the enterprise-grade model
- [x] T034 [US5] Update any doctor or setup test assertions that reference the old model name in `internal/doctor/doctor_test.go` and `internal/setup/setup_test.go`
- [x] T035 [US5] Run `go test -race -count=1 ./internal/doctor/... ./internal/setup/...` to verify model name tests pass

**Checkpoint**: Doctor and setup consistently reference the enterprise-grade embedding model.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, final validation, regression sweep.

- [x] T036 Update `AGENTS.md` Recent Changes section: add entry for `015-dewey-integration` describing the Dewey integration (scaffold, agent files, doctor/setup, embedding model update)
- [x] T037 [P] Update `AGENTS.md` Active Technologies section: replace graphthulhu references with Dewey, add Dewey MCP tool names
- [x] T038 [P] Update `AGENTS.md` Project Structure section: update `opencode.json` description from "MCP server configuration (knowledge graph)" to "MCP server configuration (Dewey)"
- [x] T039 Run full test suite `go test -race -count=1 ./...` and `go build ./...` to verify zero regressions across all packages
- [x] T040 Run comprehensive stale reference sweep: `grep -rn 'graphthulhu\|knowledge-graph' opencode.json .opencode/ internal/scaffold/assets/ internal/doctor/ internal/setup/ AGENTS.md` -- verify zero matches (excluding completed specs and archived changes)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- MCP config first
- **Foundational (Phase 2)**: Depends on Phase 1 -- tool name renames before new content
- **US1 (Phase 3)**: Depends on Phase 2 -- regression test verifies Phase 2 work
- **US2 (Phase 4)**: Depends on Phase 2 -- adds new content to files already renamed
- **US3 (Phase 5)**: Depends on Phase 4 -- fallback pattern added to the "Knowledge Retrieval" sections created in Phase 4
- **US4 (Phase 6)**: Can run in parallel with Phases 3-5 -- different files (doctor/setup vs agent files)
- **US5 (Phase 7)**: Depends on Phase 6 -- modifies the same doctor/setup files
- **Polish (Phase 8)**: Depends on all phases complete

### User Story Dependencies

- **US1 (P1)**: Depends on Phase 2 (tool renames)
- **US2 (P1)**: Depends on Phase 2 (tool renames)
- **US3 (P1)**: Depends on US2 (adds fallback to sections created in US2)
- **US4 (P2)**: Independent of US1-US3 (different files)
- **US5 (P2)**: Depends on US4 (same files)

### Parallel Opportunities

- T006, T007, T008 can run in parallel (different agent files)
- T017, T018, T019 can run in parallel (different agent files)
- T022, T023, T024 can run in parallel (different agent files)
- Phase 6 (US4) can run in parallel with Phases 3-5 (different packages)
- T037, T038 can run in parallel (different AGENTS.md sections)

---

## Implementation Strategy

### MVP First (US1 + US2 + US3)

1. Phase 1: MCP config update (T001-T003)
2. Phase 2: Tool name renames (T004-T012)
3. Phase 3: Scaffold regression test (T013-T015)
4. Phase 4: Hero semantic search instructions (T016-T020)
5. Phase 5: 3-tier degradation (T021-T026)
6. **STOP and VALIDATE**: `go test` passes, zero graphthulhu refs, all heroes have Dewey tools + fallback
7. Core integration delivered -- heroes can use Dewey with graceful degradation

### Incremental Delivery

1. Setup + Foundational → Config and tool renames done
2. US1 → Scaffold verified clean → New projects get Dewey config
3. US2 → Heroes have semantic search → Agents can use Dewey
4. US3 → Degradation pattern complete → Constitution compliant
5. US4 → Doctor/setup checks → Onboarding polished
6. US5 → Embedding model aligned → Consistency with Dewey
7. Polish → Full regression sweep → Ship it

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story
- The fix for `swarm setup`/`swarm init` stdin hang (interactive guard) is already applied in the working tree and will be included in the commit
- Completed specs and archived OpenSpec changes are NOT modified
- Cross-repo updates (gaze, website) are separate work items per spec Assumptions
- The scaffold asset copies under `internal/scaffold/assets/` must stay in sync with the live copies under `.opencode/` -- T026 handles this
