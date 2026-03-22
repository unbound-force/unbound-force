# Tasks: Dewey Architecture

**Input**: Design documents from `/specs/014-dewey-architecture/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/mcp-tools.md

**Tests**: Tests are included for the meta-repo integration work. The dewey repo implementation tasks describe what needs to be built but are executed in that repo's own speckit pipeline, not here.

**Organization**: This is an architectural spec. Tasks cover two scopes:
1. **Meta-repo tasks** (executed here): Design artifact validation, AGENTS.md updates, scaffold/doctor/setup integration
2. **Dewey-repo roadmap** (executed in `unbound-force/dewey`): Documented here for planning but implemented via separate speckit specs in the dewey repo

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Phase 1: Setup (Design Artifact Validation)

**Purpose**: Validate the architectural design artifacts are complete and internally consistent before proceeding to integration work.

- [ ] T001 Verify `specs/014-dewey-architecture/data-model.md` entity definitions (Document, Block, Link, Embedding, Source, Workspace) are consistent with the MCP tool contract parameters in `specs/014-dewey-architecture/contracts/mcp-tools.md`
- [ ] T002 [P] Verify `specs/014-dewey-architecture/research.md` decisions (R1-R9) are all reflected in the plan's Technical Context and Project Structure sections in `specs/014-dewey-architecture/plan.md`
- [ ] T003 [P] Verify `specs/014-dewey-architecture/quickstart.md` CLI commands match the contract's CLI Commands section in `specs/014-dewey-architecture/contracts/mcp-tools.md`

---

## Phase 2: Foundational (AGENTS.md Registration)

**Purpose**: Register Spec 014 in the meta repo's living documentation.

- [ ] T004 Update `AGENTS.md` Project Structure section: add `specs/014-dewey-architecture/` to the directory tree with description "Dewey semantic knowledge layer design"
- [ ] T005 [P] Update `AGENTS.md` Spec Organization section: add Spec 014 to Phase 3 (Infrastructure) with description "Dewey: semantic knowledge layer combining KG + vector search, pluggable content sources"
- [ ] T006 [P] Update `AGENTS.md` Dependency Graph: add `014-dewey-architecture` under Phase 3 with dependencies on 002 and 010
- [ ] T007 Update `AGENTS.md` Recent Changes section: add entry for `014-dewey-architecture` describing the architectural design (design artifacts only, no implementation yet)
- [ ] T008 Update `AGENTS.md` Active Technologies section: add `SQLite for persistent indexes (knowledge graph + vector embeddings)` and `Ollama API for embedding generation` under Spec 014 context

---

## Phase 3: User Story 1 - Persistent Knowledge Graph (Priority: P1)

**Goal**: Validate that the persistent storage design (SQLite backend, incremental updates, graphthulhu compatibility) is architecturally sound and documented.

**Independent Test**: Review the data model's Index entity, the research decision R2 (SQLite), and the contract's structured query tools. Verify all graphthulhu tools are accounted for.

**Note**: This is a design validation phase. The Go implementation happens in the dewey repo.

- [ ] T009 [US1] Verify `specs/014-dewey-architecture/data-model.md` Index entity lifecycle (building → ready → updating → rebuilding) covers all acceptance scenarios from US1 in `specs/014-dewey-architecture/spec.md`, including the "updating" state for real-time filesystem monitoring (FR-004)
- [ ] T010 [US1] Verify `specs/014-dewey-architecture/contracts/mcp-tools.md` lists all 6 graphthulhu-equivalent tools (search, find_by_tag, query_properties, get_page, traverse, find_connections) with the `dewey_` prefix
- [ ] T011 [US1] Verify `specs/014-dewey-architecture/research.md` R2 (SQLite) rationale addresses the warm-start performance target (<1s for 200 documents) from SC-001

---

## Phase 4: User Story 2 - Semantic Search (Priority: P1)

**Goal**: Validate that the semantic search design (embedding model, chunking, vector queries) is architecturally sound and documented.

**Independent Test**: Review the Embedding entity, the research decisions R3 (Granite) and R5 (chunking), and the contract's semantic tools. Verify the 3 new tools are fully specified.

- [ ] T012 [US2] Verify `specs/014-dewey-architecture/data-model.md` Embedding entity includes `model_name` field for invalidation on model change (FR-009 configurability), and that the Embedding lifecycle covers the "invalidated → re-embed all" flow when the model changes (FR-008 persistence)
- [ ] T013 [US2] Verify `specs/014-dewey-architecture/contracts/mcp-tools.md` specifies all 3 semantic tools (dewey_semantic_search, dewey_similar, dewey_semantic_search_filtered) with complete parameter lists
- [ ] T014 [US2] Verify `specs/014-dewey-architecture/research.md` R5 (chunking) describes heading-based strategy aligned with graphthulhu's block model, with sub-chunk splitting for content exceeding the 512-token context window

---

## Phase 5: User Story 3 - Content Sources (Priority: P2)

**Goal**: Validate that the pluggable content source design (interface, 3 source types, incremental updates) is architecturally sound and documented.

**Independent Test**: Review the Source entity, the research decision R4, and the contract's provenance metadata. Verify each source type is specified.

- [ ] T015 [US3] Verify `specs/014-dewey-architecture/data-model.md` Source entity state machine (ready → indexing → ready/failed) covers all acceptance scenarios from US3 in `specs/014-dewey-architecture/spec.md`, including incremental update flow via the `Diff()` interface method (FR-016)
- [ ] T016 [US3] Verify `specs/014-dewey-architecture/data-model.md` Workspace configuration schema includes all 3 source types (disk, github, web) with their specific config fields
- [ ] T017 [US3] Verify `specs/014-dewey-architecture/contracts/mcp-tools.md` Source Provenance Metadata section shows both GitHub and disk provenance examples with all required fields (source type, name, URL/path, fetched_at)

---

## Phase 6: User Story 4 - CLI and Configuration (Priority: P2)

**Goal**: Validate that the CLI commands and configuration format are fully specified in the contract.

**Independent Test**: Review the contract's CLI Commands section. Verify all 4 commands (init, index, status, serve) are specified with their parameters and output formats.

- [ ] T018 [US4] Verify `specs/014-dewey-architecture/contracts/mcp-tools.md` CLI Commands section specifies all 4 commands: `dewey init`, `dewey index`, `dewey status`, `dewey serve` with expected behavior
- [ ] T019 [US4] Verify `specs/014-dewey-architecture/contracts/mcp-tools.md` `dewey status` output format includes: total documents, source breakdown, last index timestamp, freshness per source, embedding model status
- [ ] T020 [US4] Verify `specs/014-dewey-architecture/quickstart.md` installation sequence (brew install → ollama pull → dewey init → dewey index → dewey serve) is consistent with the contract

---

## Phase 7: User Story 5 + 6 - Hero Integration and Graceful Degradation (Priority: P3)

**Goal**: Validate that the hero integration design and 3-tier degradation pattern are documented and constitutionally compliant.

**Independent Test**: Review research decision R9 (degradation tiers), the contract's migration path, and the spec's FR-024/FR-025 for constitutional compliance.

- [ ] T021 [US5] Verify `specs/014-dewey-architecture/research.md` R9 describes all 3 degradation tiers (no Dewey → graph-only → full) and the agent fallback pattern
- [ ] T022 [US5] Verify `specs/014-dewey-architecture/contracts/mcp-tools.md` graphthulhu Migration section specifies the exact `opencode.json` config change needed to switch from graphthulhu to Dewey
- [ ] T023 [US6] Confirm `specs/014-dewey-architecture/spec.md` FR-024 (SHOULD use Dewey) and FR-025 (MUST NOT be a hard dependency) are complementary per RFC 2119 semantics: SHOULD recommends usage when available, MUST NOT prohibits making it mandatory
- [ ] T024 [US6] Verify the 3-tier degradation pattern in `specs/014-dewey-architecture/quickstart.md` matches the Tier 1/2/3 descriptions in `specs/014-dewey-architecture/research.md` R9

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, constitution alignment verification, and documentation completeness.

- [ ] T025 Run constitution alignment check: verify all 4 principles (Autonomous Collaboration, Composability First, Observable Quality, Testability) pass against the design in `specs/014-dewey-architecture/plan.md`
- [ ] T026 [P] Verify `specs/014-dewey-architecture/spec.md` has all 25 FRs mapped to at least one acceptance scenario across the 6 user stories
- [ ] T027 [P] Verify `specs/014-dewey-architecture/spec.md` has all 8 SCs testable without implementation details
- [ ] T028 Verify the Dewey repo project structure in `specs/014-dewey-architecture/plan.md` is consistent with the data model entities and contract tools (store/ maps to Index/Embedding, embed/ maps to Embedding, source/ maps to Source, tools/ maps to MCP tools)
- [ ] T029 [P] Commit and push all spec artifacts under `specs/014-dewey-architecture/` to the remote before implementation begins (spec commit gate)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- design artifact validation first
- **Foundational (Phase 2)**: Can run in parallel with Phase 1
- **US1 (Phase 3)**: Depends on Phase 1 (design artifacts validated)
- **US2 (Phase 4)**: Can run in parallel with Phase 3
- **US3 (Phase 5)**: Can run in parallel with Phase 3-4
- **US4 (Phase 6)**: Can run in parallel with Phase 3-5
- **US5+US6 (Phase 7)**: Depends on Phase 3-6 (all designs validated before integration review)
- **Polish (Phase 8)**: Depends on all phases complete

### User Story Dependencies

- **US1 (P1)**: Independent -- persistence design stands alone
- **US2 (P1)**: Independent -- semantic search design stands alone (but builds on US1's persistence for embedding storage)
- **US3 (P2)**: Independent -- source design stands alone
- **US4 (P2)**: Depends on US1-US3 (CLI commands reference all capabilities)
- **US5 (P3)**: Depends on US1-US4 (integration requires all capabilities designed)
- **US6 (P3)**: Depends on US5 (degradation is defined relative to full capability)

### Parallel Opportunities

- T002, T003 can run in parallel (different artifacts)
- T005, T006 can run in parallel (different AGENTS.md sections)
- Phases 3, 4, 5, 6 can all run in parallel (independent design validations)
- T026, T027, T029 can run in parallel (different validation checks)

---

## Implementation Strategy

### MVP First (US1 Only)

1. Phase 1: Validate design artifacts
2. Phase 2: Register in AGENTS.md
3. Phase 3: Validate persistence design
4. **STOP and VALIDATE**: The persistence design is sound and the spec is registered
5. Core architectural decision validated -- Dewey can proceed as a graphthulhu replacement

### Incremental Delivery

1. Setup + Foundational → Design artifacts validated, AGENTS.md updated
2. US1 → Persistence design validated → Can start dewey repo Phase 2.1
3. US2 → Semantic search design validated → Can start dewey repo Phase 2.2
4. US3 → Source design validated → Can start dewey repo Phase 2.3
5. US4 → CLI design validated → Can start dewey repo Phase 2.4
6. US5+US6 → Integration design validated → Can start meta repo Phase 4 integration
7. Polish → Full spec committed → Ready for dewey repo implementation

### Relationship to Dewey Repo Work

This spec's tasks are **design validation and documentation** in the meta repo. The actual Go implementation happens in `unbound-force/dewey` via separate speckit specs:

| Meta Repo Task | Dewey Repo Spec |
|---------------|----------------|
| US1 validated | → dewey repo: Persistence spec |
| US2 validated | → dewey repo: Vector Search spec |
| US3 validated | → dewey repo: Content Sources spec |
| US4 validated | → dewey repo: CLI/Config spec |
| US5+US6 validated | → meta repo: Dewey Integration spec (Phase 4 of orchestration plan) |

---

## Notes

- This is an **architectural spec** in the meta repo -- tasks are design validation, not Go implementation
- The dewey repo has its own speckit pipeline for implementation tasks
- Completed specs in this repo are not modified (FR-014 from Spec 013 applies)
- The design paper (`../dewey-design-paper.md`) is the primary source for architectural decisions
- The orchestration plan (`../dewey-orchestration-plan.md`) sequences this spec within the broader Dewey delivery
