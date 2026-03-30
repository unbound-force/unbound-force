# Tasks: Unleash Command

**Input**: Design documents from `specs/018-unleash-command/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: No project initialization needed -- this is
a Markdown command file added to an existing scaffold
system. Setup verifies the existing infrastructure.

- [x] T001 [P] Read the existing `/finale` command at `.opencode/command/finale.md` as a reference for command file structure (frontmatter, sections, step numbering, guardrails pattern)
- [x] T002 [P] Read the existing `/cobalt-crush` command at `.opencode/command/cobalt-crush.md` as a reference for workflow auto-detection and Task tool delegation patterns

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Create the command file with the core
structure, branch safety gate, and resumability engine.
All user stories depend on this foundation.

**CRITICAL**: No user story work can begin until the
command file exists with the branch gate and resume
detection logic.

- [x] T003 Create `.opencode/command/unleash.md` with frontmatter (`description` field), command title, description section, and usage examples in `.opencode/command/unleash.md`
- [x] T004 Write the Branch Safety Gate section (step 1): get current branch via `git rev-parse --abbrev-ref HEAD`, refuse on `main` with error message, refuse on `opsx/*` with error suggesting `/opsx:apply`, refuse on unrecognized branches, validate spec.md exists via `check-prerequisites.sh` in `.opencode/command/unleash.md`
- [x] T005 Write the Resumability Detection section: probe filesystem state in order (spec.md for NEEDS CLARIFICATION markers, plan.md existence, tasks.md existence, task completion state, build/test pass), announce which steps are skipped and which step is resuming from, per the detection table in data-model.md, in `.opencode/command/unleash.md`

**Checkpoint**: Command file exists with branch gate and resume detection. Ready for pipeline step implementation.

---

## Phase 3: User Story 1 -- Happy Path Pipeline (Priority: P1)

**Goal**: The full 8-step autonomous pipeline works end-to-end when no human intervention is needed.

**Independent Test**: Run `/unleash` on a spec with no ambiguities, verify all 8 steps execute and demo instructions are presented.

- [x] T006 [US1] Write Step 1 (Clarify): scan spec.md for `[NEEDS CLARIFICATION]` markers, for each marker extract the question and 3-5 surrounding lines of context, formulate a Dewey semantic search query using `dewey_semantic_search`, evaluate results using agent judgment (per FR-005), auto-resolve if Dewey answers sufficiently (write answer to spec, add to Clarifications section), accumulate unanswerable questions in `.opencode/command/unleash.md`
- [x] T007 [US1] Write Step 2 (Plan): read the full contents of `.opencode/command/speckit.plan.md`, delegate to the `cobalt-crush-dev` agent via Task tool with the plan command's instructions as the prompt, verify plan.md was created in the feature directory in `.opencode/command/unleash.md`
- [x] T008 [US1] Write Step 3 (Tasks): read the full contents of `.opencode/command/speckit.tasks.md`, delegate to the `cobalt-crush-dev` agent via Task tool with the tasks command's instructions as the prompt, verify tasks.md was created in the feature directory in `.opencode/command/unleash.md`
- [x] T009 [US1] Write Step 4 (Spec Review): invoke `/review-council` in spec review mode by reading `.opencode/command/review-council.md` and delegating to the `cobalt-crush-dev` agent, process results -- if all APPROVE continue, if LOW/MEDIUM auto-fix, if HIGH/CRITICAL exit with findings in `.opencode/command/unleash.md`
- [x] T010 [US1] Write Step 5 (Implement) sequential path: parse tasks.md for phases, for each phase process non-`[P]` tasks sequentially via a single `cobalt-crush-dev` agent delegation, mark tasks `[x]` after completion, run build+test checkpoint after each phase in `.opencode/command/unleash.md`
- [x] T011 [US1] Write Step 6 (Code Review): invoke `/review-council` in code review mode (Phase 1a CI hard gate + Phase 1b Gaze + Divisor agents), if findings exist attempt fixes and re-run up to 3 iterations, if all APPROVE continue, if 3 iterations exhausted exit with persistent findings in `.opencode/command/unleash.md`
- [x] T012 [US1] Write Step 7 (Retrospective): analyze the session (tasks completed, review findings, fixes applied, patterns discovered), store at least one learning via `hivemind_store` with tags including the branch name and date, categorize learnings as patterns/gotchas/review-insights/file-specific per research.md R5 in `.opencode/command/unleash.md`
- [x] T013 [US1] Write Step 8 (Demo): read spec user story titles, run `git diff --name-only main...HEAD` for changed files, summarize test results, read quickstart.md if exists, present structured demo output with sections (What Was Built, How to Verify, Key Files Changed, Test Results, Next Steps: `/finale` or `/speckit.clarify`) in `.opencode/command/unleash.md`

**Checkpoint**: US1 complete. Full 8-step pipeline works for the happy path (no exits).

---

## Phase 4: User Story 2 -- Exit on Unanswerable Question (Priority: P1)

**Goal**: The clarify step exits gracefully when Dewey can't answer questions.

**Independent Test**: Spec with unanswerable ambiguity exits at clarify with the question presented.

- [x] T014 [US2] Update Step 1 (Clarify) to add the exit logic: after processing all markers, if unanswerable questions remain present them all at once in a structured format with the exit message "Answer these questions in the spec, then re-run `/unleash`" per research.md R6 exit point format in `.opencode/command/unleash.md`
- [x] T015 [US2] Update the resumability detection to handle clarify-done state: skip clarify when no `[NEEDS CLARIFICATION]` markers exist AND (Clarifications section exists OR plan.md exists) in `.opencode/command/unleash.md`

**Checkpoint**: US2 complete. Clarify exit and resume works.

---

## Phase 5: User Story 3 -- Exit on Spec Review Failure (Priority: P2)

**Goal**: Spec review exits on HIGH/CRITICAL findings.

**Independent Test**: Spec with deliberate HIGH issue exits after spec review.

- [x] T016 [US3] Update Step 4 (Spec Review) to add the exit logic: when HIGH/CRITICAL findings remain after auto-fixing LOW/MEDIUM, present findings with context and suggest "/speckit.clarify to address the findings, then re-run /unleash" per the exit point format in `.opencode/command/unleash.md`

**Checkpoint**: US3 complete. Spec review exit works.

---

## Phase 6: User Story 4 -- Parallel Implementation (Priority: P2)

**Goal**: `[P]`-marked tasks execute in parallel via Swarm workers with worktrees.

**Independent Test**: Phase with 4 `[P]` tasks spawns 4 parallel workers.

- [x] T017 [US4] Update Step 5 (Implement) to add parallel execution: for each phase, separate tasks into sequential (no `[P]`) and parallel (`[P]`) groups, run sequential tasks first, then for parallel tasks check if `swarm_worktree_create` is available in `.opencode/command/unleash.md`
- [x] T018 [US4] Write the parallel worker spawning logic: for each `[P]` task call `swarm_worktree_create` to create a dedicated worktree, then `swarm_spawn_subtask` to spawn a worker with the task description, wait for all workers to complete in `.opencode/command/unleash.md`
- [x] T019 [US4] Write the worktree merge logic: after all parallel workers complete, call `swarm_worktree_merge` for each worktree, attempt auto-resolution on merge conflicts (accept both changes), if auto-resolution fails exit to human with conflict details, call `swarm_worktree_cleanup` after merge in `.opencode/command/unleash.md`
- [x] T020 [US4] Write the parallel worker failure handling: if any worker fails stop all other workers (instruct agent to not spawn more), report the failure with error context, exit to human per FR-011 in `.opencode/command/unleash.md`
- [x] T021 [US4] Write the graceful degradation for missing Swarm worktree tools: if `swarm_worktree_create` is not available, fall back to sequential execution for `[P]` tasks with an informational note, per FR-017 in `.opencode/command/unleash.md`

**Checkpoint**: US4 complete. Parallel execution with worktrees and fallback works.

---

## Phase 7: User Story 5 -- Exit on Code Review Failure (Priority: P2)

**Goal**: Code review exits after 3 failed iterations.

**Independent Test**: Persistent circular findings exit after 3 iterations.

- [x] T022 [US5] Update Step 6 (Code Review) exit logic: after 3 iterations with remaining findings, present the persistent issues with context showing which findings were fixed vs. which persist, include detail on any circular dependencies between reviewers, suggest the human fix manually then re-run `/unleash` in `.opencode/command/unleash.md`

**Checkpoint**: US5 complete. Code review exit works.

---

## Phase 8: User Story 6 -- Resumability (Priority: P3)

**Goal**: Re-running `/unleash` skips completed steps.

**Independent Test**: Run with plan.md + tasks.md existing, verify skips to spec review.

- [x] T023 [US6] Write the step progress announcement: at the start of each `/unleash` run, display which steps are detected as complete (✓) and which step it's resuming from, e.g., "Detected: clarify ✓ plan ✓ tasks ✗ — Resuming at step 3/8: Generating tasks..." in `.opencode/command/unleash.md`
- [x] T024 [US6] Write graceful degradation notes for all optional tools: Dewey (fall back to human questions), Gaze (skip Phase 1b in code review), Swarm worktrees (fall back to sequential), Hivemind (skip retrospective), SwarmMail (proceed without locks) -- add these as inline checks at each step that uses the tool in `.opencode/command/unleash.md`

**Checkpoint**: US6 complete. Resumability and degradation work.

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Scaffold integration, guardrails, test updates, documentation.

- [x] T025 Write the Guardrails section at the end of `.opencode/command/unleash.md`: NEVER run on main, NEVER skip spec review exit on HIGH/CRITICAL, NEVER merge worktrees with unresolved semantic conflicts, ALWAYS present exit messages with actionable next steps, ALWAYS store at least one learning in retrospective
- [x] T026 Copy `.opencode/command/unleash.md` to `internal/scaffold/assets/opencode/command/unleash.md` to include in the scaffold embed
- [x] T027 [P] Update the expected file count in `cmd/unbound-force/main_test.go` from 51 to 52 to account for the new `unleash.md` command file
- [x] T028 [P] Add `"opencode/command/unleash.md"` to the `expectedAssetPaths` slice in `internal/scaffold/scaffold_test.go` (maintain alphabetical order within the commands section, update the comment count from 13 to 14)
- [x] T029 Run `go build ./...` to verify the build succeeds with the new embedded asset
- [x] T030 Run `go test -race -count=1 ./...` to verify all tests pass including the updated file count and scaffold drift detection
- [x] T031 Update AGENTS.md: add spec 018 to project structure tree, spec organization (Phase 2), dependency graph, and "Recent Changes" section with this spec's summary

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- read reference files
- **Foundational (Phase 2)**: Depends on Phase 1 -- creates the command file skeleton
- **US1 (Phase 3)**: Depends on Phase 2 -- writes the 8 pipeline steps
- **US2 (Phase 4)**: Depends on Phase 3 T006 -- adds exit logic to clarify step
- **US3 (Phase 5)**: Depends on Phase 3 T009 -- adds exit logic to spec review step
- **US4 (Phase 6)**: Depends on Phase 3 T010 -- adds parallel execution to implement step
- **US5 (Phase 7)**: Depends on Phase 3 T011 -- adds exit logic to code review step
- **US6 (Phase 8)**: Depends on Phase 3 -- adds resume detection and degradation
- **Polish (Phase 9)**: Depends on all phases complete

### User Story Dependencies

- **US1 (P1)**: Independent -- writes the full pipeline
- **US2 (P1)**: Depends on US1 T006 (clarify step must exist to add exit logic)
- **US3 (P2)**: Depends on US1 T009 (spec review step must exist)
- **US4 (P2)**: Depends on US1 T010 (implement step must exist)
- **US5 (P2)**: Depends on US1 T011 (code review step must exist)
- **US6 (P3)**: Depends on US1 (all steps must exist for resume detection)

### Parallel Opportunities

- T001, T002 can run in parallel (reading reference files)
- T027, T028 can run in parallel (different test files)
- US3, US4, US5 can run in parallel after US1 (different steps in the same file, but different sections)
- Phase 9 T026-T028 can partially parallelize (different files)

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T002) -- read references
2. Complete Phase 2: Foundational (T003-T005) -- skeleton
3. Complete Phase 3: US1 (T006-T013) -- full pipeline
4. **STOP and VALIDATE**: Test `/unleash` on a simple spec
5. Happy path works end-to-end

### Incremental Delivery

1. Setup + Foundational → Command file skeleton
2. US1 → Full 8-step pipeline (happy path)
3. US2 → Clarify exit and resume
4. US3 → Spec review exit
5. US4 → Parallel implementation with worktrees
6. US5 → Code review exit
7. US6 → Resume detection and degradation
8. Polish → Scaffold, tests, docs
