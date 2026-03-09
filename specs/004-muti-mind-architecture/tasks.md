# Implementation Tasks: Muti-Mind Architecture (Product Owner)

## Phase 1: Setup & Infrastructure
**Goal**: Initialize the project structure and shared infrastructure.

- [x] T001 Create `.muti-mind/` and `.muti-mind/backlog/` directory structure
- [x] T001a Create hero manifest `schemas/hero-manifest/muti-mind-hero.json` defining the PO role, produced artifacts, and commands (FR-018)
- [x] T002 Create initial `.muti-mind/config.yaml` with default settings
- [x] T003 Create `cmd/mutimind/main.go` entrypoint as the backend CLI for data logic
- [x] T004 Create `internal/backlog` package for MD file parsing, writing, and synchronization logic

## Phase 2: User Story 1 - AI Persona and Decision Framework (P1)
**Goal**: Establish the Muti-Mind AI persona and decision-making framework.
**Independent Test**: Persona answers product questions consistently based on vision without contradicting prior decisions.

- [x] T005 [P] [US1] Create `.opencode/agents/muti-mind-po.md` defining the AI persona and decision framework
- [x] T006 [US1] Implement `/muti-mind.init` command to install/configure the agent locally

## Phase 3: User Story 2 - Backlog Management Commands (P1)
**Goal**: Create, read, update, and delete backlog items via OpenCode commands.
**Independent Test**: Can create, update, and list backlog items with correct priority ordering and details.

- [x] T007 [P] [US2] Implement `/muti-mind.backlog-add` command
- [x] T008 [P] [US2] Implement `/muti-mind.backlog-list` command
- [x] T009 [P] [US2] Implement `/muti-mind.backlog-update` command
- [x] T010 [P] [US2] Implement `/muti-mind.backlog-show` command

## Phase 4: User Story 3 - Priority Scoring Engine (P2)
**Goal**: Provide AI-assisted priority scoring algorithm.
**Independent Test**: Engine ranks items and provides transparent score breakdowns across 5 dimensions.

- [x] T011 [US3] Implement `/muti-mind.prioritize` command to delegate scoring to the AI agent
- [x] T012 [US3] Update `muti-mind-po.md` agent prompt to include priority scoring logic (business value, risk, dependencies, urgency, effort)

## Phase 5: User Story 4 - GitHub Issues Synchronization (P2)
**Goal**: Two-way sync between local backlog MD files and GitHub Issues/Projects.
**Independent Test**: Round-trip a backlog item (create local -> push -> modify on GitHub -> pull) with zero data loss.

- [x] T013 [P] [US4] Implement `/muti-mind.sync-push` command (local -> GitHub)
- [x] T014 [P] [US4] Implement `/muti-mind.sync-pull` command (GitHub -> local)
- [x] T015 [P] [US4] Implement `/muti-mind.sync-status` command to report sync state
- [x] T016 [US4] Implement `/muti-mind.sync` (bidirectional) including interactive conflict detection and resolution prompts
- [x] T016a [US4] Implement `/muti-mind.sync-project` command

## Phase 6: User Story 5 - Speckit Integration and Acceptance Authority (P3)
**Goal**: Drive speckit pipeline and act as acceptance authority on Gaze reports.
**Independent Test**: Muti-Mind reviews a Gaze report and produces an accept/reject decision with rationale.

- [x] T017 [US5] Update `muti-mind-po.md` agent to handle speckit `/specify` and `/clarify` invocation workflows
- [x] T018 [US5] Implement acceptance logic in `muti-mind-po.md` to evaluate Gaze reports against backlog item acceptance criteria
- [x] T019 [US5] Implement generation of the `acceptance-decision` JSON artifact
- [x] T019a [US5] Implement automated generation of the `backlog-item` JSON artifact representation (FR-015)

## Phase 7: User Story 6 - User Story Generation (P3)
**Goal**: Generate structured user stories from high-level goals.
**Independent Test**: Generate well-formed stories with Given/When/Then criteria from a brief description.

- [x] T020 [US6] Implement `/muti-mind.generate-stories` command
- [x] T021 [US6] Update `muti-mind-po.md` agent to support the story generation prompt and output format
- [x] T021a [US6] Implement interactive approval workflow to confirm generated story proposals before adding them to the backlog (FR-014)

## Final Phase: Polish & Cross-Cutting Concerns
**Goal**: Ensure quality, formatting, documentation, and compliance.

- [x] T022 Ensure all commands support `--format json` and `--format text` (FR-017)
- [x] T023 Validate all outputs against Hero Interface Contract artifact envelopes
- [x] T024 Write integration tests verifying OpenCode shell delegation to Go binary
- [x] T025 Setup CI coverage ratchets (80% global, 90% internal/backlog) to enforce Principle IV testability
