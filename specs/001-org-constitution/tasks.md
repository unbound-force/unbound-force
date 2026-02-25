# Tasks: Unbound Force Organization Constitution

**Input**: Design documents from `/specs/001-org-constitution/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, quickstart.md

**Tests**: No test tasks included -- not requested in the feature specification.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

This feature produces governance documents and OpenCode agent configurations, not compiled software:

- **Constitution**: `.specify/memory/constitution.md`
- **Agent files**: `.opencode/agents/`, `.opencode/command/`
- **Documentation**: `AGENTS.md`, `README.md`

## Phase 1: Setup

**Purpose**: Verify project structure and prerequisites are in place

- [x] T001 Verify `.specify/memory/constitution.md` exists and contains the ratified v1.0.0 constitution (not the blank template)
- [x] T002 Verify `.opencode/agents/` directory exists for agent file placement
- [x] T003 Verify `.opencode/command/` directory exists for command file placement

**Checkpoint**: Project structure verified -- ready for constitution validation and agent creation.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Ensure the ratified constitution meets all spec requirements before building the alignment agent

- [x] T004 Validate the ratified constitution at `.specify/memory/constitution.md` against FR-001: verify exactly three core principles are defined (Autonomous Collaboration, Composability First, Observable Quality)
- [x] T005 Validate the constitution against FR-002: verify each principle contains at least three MUST rules and at least one SHOULD rule
- [x] T006 Validate the constitution against FR-003: verify the Governance section defines amendment process, versioning scheme, supremacy clause, and hero alignment requirements
- [x] T007 Validate the constitution against FR-018: verify the Development Workflow section defines feature branches, code review, CI, semantic versioning, and conventional commits
- [x] T008 Validate the constitution against FR-019: verify the Governance section includes the conflict resolution clause (principle conflicts resolved via documented tradeoffs, no implicit priority)
- [x] T009 Validate the constitution against FR-005: verify the constitution references the Hero Interface Contract concept (deferred to Spec 002)
- [x] T010 Validate the constitution against SC-006: verify the document is under 500 lines (excluding the Sync Impact Report HTML comment)

**Checkpoint**: Constitution validated against all functional requirements. User story work can begin.

---

## Phase 3: User Story 1 - Constitution Ratification (Priority: P1) MVP

**Goal**: Ratify the org constitution with three core principles, governance model, development workflow, and hero alignment section. Ensure the document is complete and all placeholder tokens are replaced.

**Independent Test**: Review `.specify/memory/constitution.md` and verify: no bracket placeholder tokens remain, three principles each have 3+ MUST rules and 1+ SHOULD rule, Governance section has all four required elements (amendment, versioning, supremacy, alignment).

### Implementation for User Story 1

- [x] T011 [US1] Verify all placeholder tokens (`[PROJECT_NAME]`, `[PRINCIPLE_*]`, etc.) are replaced with concrete content in `.specify/memory/constitution.md` (FR-001, acceptance scenario 1)
- [x] T012 [US1] Verify Principle I (Autonomous Collaboration) contains MUST rules for artifact-based communication (FR-009), independent primary function (FR-010), and self-describing outputs (FR-011) in `.specify/memory/constitution.md`
- [x] T013 [US1] Verify Principle II (Composability First) contains MUST rules for standalone installability (FR-012), extension points (FR-013), and additive combination value (FR-014) in `.specify/memory/constitution.md`
- [x] T014 [US1] Verify Principle III (Observable Quality) contains MUST rules for machine-parseable output (FR-015), provenance metadata (FR-016), and automated evidence (FR-017) in `.specify/memory/constitution.md`
- [x] T015 [US1] Verify the Governance section includes the supremacy clause (FR-006) and the org-hero relationship definition (FR-007) in `.specify/memory/constitution.md`
- [x] T016 [US1] Verify the constitution includes the Constitution Check process reference (FR-008) in `.specify/memory/constitution.md`
- [x] T017 [US1] Update `AGENTS.md` to reflect the ratified constitution principles in the "Constitution (Highest Authority)" section
- [x] T018 [US1] Update `README.md` with a project description that references the constitution and its three principles

**Checkpoint**: Constitution ratification complete. SC-001 and SC-002 are satisfied. The constitution is the authoritative governance document for the org.

---

## Phase 4: User Story 2 - Hero Constitution Alignment Validation (Priority: P2)

**Goal**: Create an OpenCode agent and command that performs agent-assisted alignment checking of hero constitutions against the org constitution, producing a structured report.

**Independent Test**: Run `/constitution-check` against the Gaze constitution at `/Users/jflowers/Projects/github/unbound-force/gaze/.specify/memory/constitution.md` and verify the report shows ALIGNED with zero contradictions (SC-003).

### Implementation for User Story 2

- [x] T019 [US2] Create the alignment checking agent at `.opencode/agents/constitution-check.md` with: `mode: subagent`, `model: google-vertex-anthropic/claude-sonnet-4-6@default`, `temperature: 0.1`, read-only tool permissions (`read: true`, all others false)
- [x] T020 [US2] Write the agent system prompt in `.opencode/agents/constitution-check.md` with instructions to: read the org constitution, read the hero constitution, compare each org principle against hero principles, produce structured findings per the data-model.md Alignment Finding entity
- [x] T021 [US2] Define the agent output format in `.opencode/agents/constitution-check.md` using the report template from research.md: header (hero name, versions, timestamp, overall status), findings section (one `### [STATUS] Org Principle ↔ Hero Principle` block per org principle), summary section (counts of aligned/gap/contradiction, parent reference status)
- [x] T022 [US2] Define the agent decision criteria in `.opencode/agents/constitution-check.md`: ALIGNED requires all findings ALIGNED or GAP and parent reference PRESENT; NON-ALIGNED if any CONTRADICTION or parent reference MISSING
- [x] T023 [US2] Create the `/constitution-check` command at `.opencode/command/constitution-check.md` with: description, agent delegation to `constitution-check`, instructions to locate both constitution files and pass them to the agent
- [x] T024 [US2] Add usage instructions to the command in `.opencode/command/constitution-check.md`: how to specify the org constitution path (default: read from unbound-force meta repo or local `.specify/memory/constitution.md`), how to specify the hero constitution path (default: current repo's `.specify/memory/constitution.md`)
- [x] T025 [US2] Validate the agent by running `/constitution-check` against the Gaze constitution and verifying the report shows ALIGNED with zero contradictions (SC-003)
- [x] T026 [US2] Validate the agent by running `/constitution-check` against the Website constitution and verifying the report shows ALIGNED with zero contradictions (SC-004)

**Checkpoint**: Alignment agent and command functional. SC-003 and SC-004 are satisfied. Agent produces structured reports for any hero constitution.

---

## Phase 5: User Story 3 - Constitution-Aware Development Workflow (Priority: P3)

**Goal**: Ensure the constitution provides clear, citable answers to cross-cutting design questions. Validate that developers and agents can use it to resolve ambiguity.

**Independent Test**: Present three cross-cutting design questions and verify the constitution provides a clear, citable answer for each (SC-005).

### Implementation for User Story 3

- [ ] T027 [US3] Validate design question 1: "Should Muti-Mind require Gaze to be installed?" -- verify Principle II (Composability First) provides a clear MUST rule answering this (heroes MUST be usable standalone) in `.specify/memory/constitution.md`
- [ ] T028 [US3] Validate design question 2: "What output format should a new hero use?" -- verify Principle III (Observable Quality) provides MUST rules about machine-parseable output (JSON minimum) and provenance metadata in `.specify/memory/constitution.md`
- [ ] T029 [US3] Validate design question 3: "How should two heroes communicate?" -- verify Principle I (Autonomous Collaboration) provides MUST rules about artifact-based communication (no runtime coupling) in `.specify/memory/constitution.md`
- [ ] T030 [US3] Update `AGENTS.md` "Constitution Check" section to reference the `/constitution-check` command as the recommended alignment verification tool
- [ ] T031 [US3] Verify that the `AGENTS.md` "Constitution (Highest Authority)" section accurately summarizes all three principles and references the constitution file path

**Checkpoint**: All three user stories complete. SC-005 satisfied. The constitution is citable, enforceable, and agent-validated.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final documentation alignment and validation

- [ ] T032 Verify `AGENTS.md` is consistent with the ratified constitution -- all principle names, rule counts, and governance details match
- [ ] T033 Verify `specs/001-org-constitution/spec.md` status can be updated from Draft to Complete (all FRs and SCs satisfied)
- [ ] T034 Run quickstart.md validation: follow the quickstart guide steps and verify each step produces the expected outcome
- [ ] T035 Verify no orphaned or contradictory documentation exists across `AGENTS.md`, `README.md`, `unbound-force.md`, and `.specify/memory/constitution.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- can start immediately
- **Foundational (Phase 2)**: Depends on Setup -- BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational -- constitution ratification validation
- **User Story 2 (Phase 4)**: Depends on Phase 3 -- alignment agent needs the ratified constitution to exist
- **User Story 3 (Phase 5)**: Depends on Phase 3 -- design questions reference the ratified constitution; can run in parallel with Phase 4
- **Polish (Phase 6)**: Depends on Phases 3, 4, and 5

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) -- no dependencies on other stories
- **User Story 2 (P2)**: Depends on US1 (needs ratified constitution as input for the alignment agent)
- **User Story 3 (P3)**: Depends on US1 (validates the ratified constitution's citability). Can run in parallel with US2.

### Within Each User Story

- Validation tasks before documentation tasks
- Agent definition before command definition (US2)
- Agent creation before agent validation (US2)
- Commit after each task or logical group

### Parallel Opportunities

- T001, T002, T003 (Setup) can all run in parallel
- T004-T010 (Foundational validation) can all run in parallel
- T011-T016 (US1 constitution validation) can all run in parallel
- T019-T024 (US2 agent creation) are sequential (each builds on the previous)
- T027-T029 (US3 design question validation) can all run in parallel
- US2 (Phase 4) and US3 (Phase 5) can run in parallel after US1 completes

---

## Parallel Example: Foundational Validation

```bash
# Launch all foundational validation tasks together:
Task: "Validate constitution against FR-001 (three principles)"
Task: "Validate constitution against FR-002 (MUST/SHOULD rule counts)"
Task: "Validate constitution against FR-003 (Governance section)"
Task: "Validate constitution against FR-018 (Development Workflow)"
Task: "Validate constitution against FR-019 (conflict resolution)"
Task: "Validate constitution against FR-005 (HIC reference)"
Task: "Validate constitution against SC-006 (under 500 lines)"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (verify structure)
2. Complete Phase 2: Foundational (validate constitution against FRs)
3. Complete Phase 3: User Story 1 (constitution ratification validation)
4. **STOP and VALIDATE**: Constitution is ratified, validated, and documented
5. This MVP delivers the governance foundation for all other specs

### Incremental Delivery

1. Complete Setup + Foundational + US1 → Ratified constitution (MVP)
2. Add US2 → Alignment agent for hero validation
3. Add US3 → Constitution citability validation + workflow documentation
4. Polish → Final documentation consistency check

### Note on Pre-Existing Work

The constitution at `.specify/memory/constitution.md` has already been ratified as v1.0.0. Many US1 tasks are validation tasks confirming the existing document meets all FRs, rather than creation tasks. US2 (alignment agent) and US3 (citability validation) are the primary new implementation work.

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- The constitution is already ratified -- US1 is primarily validation, US2 and US3 are new work
