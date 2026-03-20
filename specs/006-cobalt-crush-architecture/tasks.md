# Tasks: Cobalt-Crush Architecture (Developer)

**Input**: Design documents from `/specs/006-cobalt-crush-architecture/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

**Tests**: Tests are included — the plan specifies drift detection tests and scaffold integration tests as part of the constitution's Testability principle.

**Organization**: Tasks are grouped by user story. A prerequisite refactor (convention pack relocation) must complete before user story work begins.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Create target directories for the convention pack relocation.

- [x] T001 Create directory `.opencode/unbound/packs/`
- [x] T002 Create directory `internal/scaffold/assets/opencode/unbound/packs/`

---

## Phase 2: Foundational — Convention Pack Relocation

**Purpose**: Move convention packs from `.opencode/divisor/packs/` to `.opencode/unbound/packs/` per research.md R1. This is a prerequisite for all user stories — both Cobalt-Crush and Divisor agents must reference the new shared location.

**CRITICAL**: All user story work depends on this phase completing. The relocation touches Go source, tests, agents, embedded assets, and documentation.

### Move Files

- [x] T003 [P] Move 6 canonical pack files from `.opencode/divisor/packs/` to `.opencode/unbound/packs/` (go.md, go-custom.md, default.md, default-custom.md, typescript.md, typescript-custom.md). Use `git mv` to preserve history.
- [x] T004 [P] Move 6 embedded pack files from `internal/scaffold/assets/opencode/divisor/packs/` to `internal/scaffold/assets/opencode/unbound/packs/`. Use `git mv` to preserve history.
- [x] T005 Remove empty directories `.opencode/divisor/` and `internal/scaffold/assets/opencode/divisor/` after moves are complete.

### Update Go Source Code

- [x] T006 Update `isConventionPack()` in `internal/scaffold/scaffold.go`: change prefix string from `"opencode/divisor/packs/"` to `"opencode/unbound/packs/"`.
- [x] T007 Update `isDivisorAsset()` in `internal/scaffold/scaffold.go`: remove `strings.HasPrefix(relPath, "opencode/divisor/")` check and replace with `isConventionPack(relPath)` call per research.md R2. This ensures `--divisor` mode still deploys convention packs from the new location.
- [x] T008 Update `isConventionPack()` GoDoc comment in `internal/scaffold/scaffold.go` to reference `opencode/unbound/packs/`.

### Update Go Test Code

- [x] T009 Update `expectedAssetPaths` in `internal/scaffold/scaffold_test.go`: change all 6 `opencode/divisor/packs/*` entries to `opencode/unbound/packs/*`. Update comment from "Divisor convention packs" to "Convention packs (shared)".
- [x] T010 Update `TestRun_CreatesFiles` expectedDirs in `internal/scaffold/scaffold_test.go`: change `.opencode/divisor/packs` to `.opencode/unbound/packs`.
- [x] T011 Update `TestIsToolOwned` test cases in `internal/scaffold/scaffold_test.go`: change all 6 `opencode/divisor/packs/*` paths to `opencode/unbound/packs/*`.
- [x] T012 Update `TestIsDivisorAsset` test cases in `internal/scaffold/scaffold_test.go`: change `opencode/divisor/packs/*` paths to `opencode/unbound/packs/*`. Add a test case verifying `isConventionPack` return values for new paths.
- [x] T013 Update `TestShouldDeployPack` test cases in `internal/scaffold/scaffold_test.go`: change all `opencode/divisor/packs/*` paths to `opencode/unbound/packs/*`.
- [x] T014 Update `TestRun_DivisorSubset` in `internal/scaffold/scaffold_test.go`: change `"divisor/packs"` substring assertions to `"unbound/packs"`.
- [x] T015 Update `TestRun_DivisorSubset_WithLangFlag` in `internal/scaffold/scaffold_test.go`: change `"divisor/packs"` substring assertions to `"unbound/packs"`.
- [x] T016 Update `TestRun_DivisorSubset_DefaultFallback` in `internal/scaffold/scaffold_test.go`: change `"divisor/packs"` substring assertions to `"unbound/packs"`.
- [x] T017 Update `TestCanonicalSources_AreEmbedded` canonicalDirs in `internal/scaffold/scaffold_test.go`: change `.opencode/divisor/packs` to `.opencode/unbound/packs`.

### Update Divisor Agent Files

- [x] T018 [P] Update `.opencode/agents/divisor-guard.md`: replace all `.opencode/divisor/packs/` references with `.opencode/unbound/packs/`.
- [x] T019 [P] Update `.opencode/agents/divisor-architect.md`: replace all `.opencode/divisor/packs/` references with `.opencode/unbound/packs/`.
- [x] T020 [P] Update `.opencode/agents/divisor-adversary.md`: replace all `.opencode/divisor/packs/` references (3 occurrences) with `.opencode/unbound/packs/`.
- [x] T021 [P] Update `.opencode/agents/divisor-sre.md`: replace all `.opencode/divisor/packs/` references with `.opencode/unbound/packs/`.
- [x] T022 [P] Update `.opencode/agents/divisor-testing.md`: replace all `.opencode/divisor/packs/` references with `.opencode/unbound/packs/`.

### Sync Embedded Agent Copies

- [x] T023 Copy updated canonical Divisor agents to embedded assets: copy `.opencode/agents/divisor-{guard,architect,adversary,sre,testing}.md` to `internal/scaffold/assets/opencode/agents/` (5 files).

### Update Documentation

- [x] T024 [P] Update `AGENTS.md`: change project structure tree from `.opencode/divisor/packs/` to `.opencode/unbound/packs/`. Update comment from "Convention packs for Divisor personas" to "Convention packs (shared by all heroes)". Update Recent Changes entry for Spec 005 to note the pack relocation.
- [x] T025 [P] Update Spec 005 artifacts: find-and-replace `.opencode/divisor/packs/` → `.opencode/unbound/packs/` and `opencode/divisor/packs/` → `opencode/unbound/packs/` across `specs/005-the-divisor-architecture/spec.md`, `plan.md`, `tasks.md`, `data-model.md`, `research.md`, `quickstart.md`, and `contracts/scaffold-cli.md` (~50 references).

### Verify

- [x] T026 Run `go test -race -count=1 ./...` and verify all tests pass including drift detection. Fix any failures from the relocation.
- [x] T027 Run `go build ./...` and verify the build succeeds.

**Checkpoint**: Convention packs are at `.opencode/unbound/packs/`. All Divisor agents reference the new path. All tests pass. Build succeeds.

---

## Phase 3: User Story 1 + 2 — AI Persona and Coding Standards (Priority: P1) MVP

**Goal**: Create the `cobalt-crush-dev.md` agent file with the engineering philosophy (US1) and convention pack loading for coding standards (US2). These are combined because the persona and the coding standards framework are both expressed in the same agent file.

**Independent Test**: Deploy the agent, ask it to implement a feature from a spec. Verify output follows clean code principles, loads convention packs, adheres to language-specific conventions, documents design decisions, and produces testable code.

- [x] T028 [US1] [US2] Create `cobalt-crush-dev.md` at `.opencode/agents/cobalt-crush-dev.md`. Structure per research.md R3: YAML frontmatter (description: "Adaptive implementation engine and coding persona", mode: agent, model, temperature: 0.4, tools: all enabled), H1 Role (engineering philosophy — clean code, SOLID, TDD awareness, CI/CD focus, spec-driven development), H2 Source Documents (AGENTS.md, constitution, specs, tasks.md, convention packs at `.opencode/unbound/packs/`, artifacts at `.unbound-force/artifacts/`, graphthulhu MCP conditional), H2 Engineering Philosophy (clean code principles, SOLID, DRY, YAGNI, separation of concerns), H2 Code Implementation Checklist (convention pack adherence with `[PACK]` sections, test hook generation, documentation patterns, error handling), H2 Decision Framework (ambiguity resolution, pattern selection, rationale documentation), H2 Output Standards (code quality expectations).

**Checkpoint**: Agent file exists. When deployed, it reads convention packs and produces code following the engineering philosophy.

---

## Phase 4: User Story 3 — Gaze Feedback Loop (Priority: P2)

**Goal**: Add Gaze feedback loop instructions to the agent file. Cobalt-Crush reads Gaze reports and addresses quality issues.

**Independent Test**: Have Cobalt-Crush implement a feature, run Gaze, feed the report back, verify Cobalt-Crush addresses issues.

- [x] T029 [US3] Add H2 Gaze Feedback Loop section to `.opencode/agents/cobalt-crush-dev.md`: instructions to read quality reports from `.unbound-force/artifacts/quality-report/` (or fallback to `coverage.out`, Gaze CLI output), identify CRAP scores > 30, coverage gaps, testability issues, and produce corrective changes. Include step-by-step workflow: (1) check for Gaze artifacts, (2) parse findings, (3) address each finding, (4) re-run validation, (5) proceed to review when clean.

---

## Phase 5: User Story 4 — Divisor Feedback Loop (Priority: P2)

**Goal**: Add Divisor feedback loop instructions to the agent file. Cobalt-Crush reads review findings and addresses them.

**Independent Test**: Simulate a Divisor review with findings, have Cobalt-Crush address them, verify fixes resolve issues.

- [x] T030 [US4] Add H2 Divisor Review Preparation section to `.opencode/agents/cobalt-crush-dev.md`: instructions to read review verdicts from `.unbound-force/artifacts/review-verdict/` (or recent `/review-council` output), address each REQUEST CHANGES finding systematically, re-run Gaze after fixes, learn from past review patterns (read previous findings and proactively apply conventions). Include step-by-step workflow: (1) check for review artifacts, (2) categorize findings by persona and severity, (3) address CRITICAL/HIGH first, (4) re-validate with Gaze, (5) re-submit for review.

---

## Phase 6: User Story 5 — Speckit Integration (Priority: P3)

**Goal**: Add speckit integration instructions to the agent file. Cobalt-Crush serves as the coding persona for `/speckit.implement`.

**Independent Test**: Run `/speckit.implement` with Cobalt-Crush active, verify it processes tasks in order and follows conventions.

- [x] T031 [US5] Add H2 Speckit Integration section to `.opencode/agents/cobalt-crush-dev.md`: instructions for working with tasks.md (read phase structure, process tasks in dependency order, respect `[P]` markers, mark `[x]` on completion, run test suite at phase checkpoints, skip dependent tasks until prerequisites are met). Reference the existing `/speckit.implement` command as the orchestration layer. Include guidance on mapping `[US1]` tags to spec acceptance criteria.

---

## Phase 7: User Story 6 — Deployment via `unbound init` (Priority: P3)

**Goal**: Embed the agent file in the scaffold engine and update tests so `unbound init` deploys `cobalt-crush-dev.md`.

**Independent Test**: Run `unbound init` in a temp directory, verify `cobalt-crush-dev.md` exists and references `.opencode/unbound/packs/`.

- [x] T032 [US6] Copy `.opencode/agents/cobalt-crush-dev.md` to `internal/scaffold/assets/opencode/agents/cobalt-crush-dev.md`.
- [x] T033 [US6] Add `"opencode/agents/cobalt-crush-dev.md"` to `expectedAssetPaths` in `internal/scaffold/scaffold_test.go`. Update comment to reflect new agent count.
- [x] T034 [US6] Update `cmd/unbound/main_test.go`: change "45 files processed" to "46 files processed" in `TestRunInit_FreshDir`.
- [x] T035 [US6] Run `go test -race -count=1 ./...` and verify all tests pass including drift detection for the new agent file.
- [x] T036 [US6] Run `go build ./...` and verify the build succeeds.

**Checkpoint**: `unbound init` deploys 46 files including `cobalt-crush-dev.md`. All tests pass.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, status updates, and final validation.

- [x] T037 [P] Update `AGENTS.md`: change Cobalt-Crush status in Heroes table from "Spec only (006)" to "Implemented (Spec 006)". Add `cobalt-crush-dev.md` to the project structure tree under `.opencode/agents/`. Add Spec 006 to Recent Changes.
- [x] T038 [P] Update `README.md`: change Cobalt-Crush status from "Spec only" to "Implemented". Update file count from "45 files" to "46 files".
- [x] T039 [P] Update `specs/006-cobalt-crush-architecture/spec.md`: change `status: draft` to `status: complete` in frontmatter and body.
- [x] T040 Run `make check` (or `go build ./... && go test -race -count=1 ./... && go vet ./...`) and verify all checks pass.
- [x] T041 Verify SC-001 through SC-007 success criteria from spec.md are met.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Phase 1 — pack relocation is the prerequisite refactor
- **US1+US2 (Phase 3)**: Depends on Phase 2 — agent needs `.opencode/unbound/packs/` path
- **US3 (Phase 4)**: Depends on Phase 3 — adds section to agent file
- **US4 (Phase 5)**: Depends on Phase 3 — adds section to agent file. Can run in parallel with US3.
- **US5 (Phase 6)**: Depends on Phase 3 — adds section to agent file. Can run in parallel with US3/US4.
- **US6 (Phase 7)**: Depends on Phases 3-6 — embeds completed agent file
- **Polish (Phase 8)**: Depends on all phases

### Parallel Opportunities

- **Phase 1**: T001 and T002 can run in parallel
- **Phase 2**: T003 and T004 (file moves) can run in parallel. T018-T022 (agent updates) can run in parallel. T024-T025 (doc updates) can run in parallel.
- **Phases 4-6**: US3, US4, US5 can run in parallel (different sections of same agent file — but sequential if single agent editing)
- **Phase 8**: T037-T039 (doc updates) can run in parallel

---

## Implementation Strategy

### MVP First (Phases 1-3)

1. Phase 1: Create directories
2. Phase 2: Convention pack relocation (the main engineering work)
3. Phase 3: Create `cobalt-crush-dev.md` with persona + coding standards
4. **STOP and VALIDATE**: Run tests, verify agent deploys, verify convention packs at new location

### Incremental Delivery

1. Phases 1-3 → Cobalt-Crush exists with engineering philosophy + convention pack loading (MVP)
2. Phase 4 → Gaze feedback loop added
3. Phase 5 → Divisor review preparation added
4. Phase 6 → Speckit integration added
5. Phase 7 → Scaffold engine integration + tests
6. Phase 8 → Documentation + validation

### Estimated Time

| Phase | Tasks | Est. Time |
|-------|-------|-----------|
| Phase 1: Setup | 2 | 1 min |
| Phase 2: Pack relocation | 25 | 45 min |
| Phase 3: US1+US2 Agent | 1 | 20 min |
| Phase 4: US3 Gaze loop | 1 | 10 min |
| Phase 5: US4 Divisor loop | 1 | 10 min |
| Phase 6: US5 Speckit | 1 | 10 min |
| Phase 7: US6 Scaffold | 5 | 15 min |
| Phase 8: Polish | 5 | 10 min |
| **Total** | **41** | **~2 hours** |
<!-- scaffolded by unbound vdev -->
