# Tasks: Unleash OpenSpec Support

**Input**: Design documents from `specs/031-unleash-openspec/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/unleash-command-contract.md

**Scope**: 1 live Markdown command file + 1 scaffold asset copy. No Go code. ~50 lines changed.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to

---

## Phase 1: US1+US3 — Branch Detection & Skip Logic (Priority: P1+P2) 🎯 MVP

**Goal**: Replace the `opsx/*` hard stop with OpenSpec detection, change name extraction, prerequisite gating, and clarify/plan/tasks skip logic. This is the core routing change.

**Independent Test**: Run `/unleash` on an `opsx/*` branch with completed `/opsx-propose` artifacts. Verify it detects the OpenSpec change, announces skip of clarify/plan/tasks, and proceeds to spec review. Run on a `NNN-*` branch and verify unchanged behavior.

### Implementation

- [x] T001 [US1] Update frontmatter description in `.opencode/command/unleash.md` (lines 2-7): change `Run the full Speckit pipeline autonomously` to `Run the full Speckit or OpenSpec pipeline autonomously` per data-model.md §1 (FR-001)
- [x] T002 [US1] Update body description in `.opencode/command/unleash.md` (lines 16-21): change `Autonomous Speckit pipeline execution` to `Autonomous pipeline execution for both Speckit (strategic) and OpenSpec (tactical) changes` per data-model.md §2 (FR-001)
- [x] T003 [US1] In Step 1 — Branch Safety Gate (lines 59-61): replace the `opsx/*` STOP block with OpenSpec detection logic — extract `<name>` from `opsx/<name>`, set `FEATURE_DIR = openspec/changes/<name>/`, set `WORKFLOW_TIER = openspec`, announce `Detected OpenSpec change: <name>` per contracts §After and data-model.md §3 (FR-001, FR-002)
- [x] T004 [US1] In Step 1 — Branch Safety Gate: add OpenSpec prerequisite check — verify `FEATURE_DIR/tasks.md` exists; if not, STOP with `No tasks.md found for change '<name>'. Run '/opsx-propose' first.` per contracts §Error Messages (FR-003)
- [x] T005 [US1] In Step 1 — Branch Safety Gate (lines 69-82): wrap the existing `check-prerequisites.sh` call in a Speckit-only conditional — run only when `WORKFLOW_TIER = speckit`; set `FEATURE_DIR` from JSON output and `WORKFLOW_TIER = speckit` per data-model.md §3 branch 3 (FR-008)
- [x] T006 [US3] Add skip logic between Step 3 (Tasks) and Step 4 (Spec Review): if `WORKFLOW_TIER = openspec`, announce `OpenSpec mode — artifacts from /opsx-propose, skipping clarify/plan/tasks` and skip directly to Step 4 per data-model.md §5 (FR-004)
- [x] T007 [US1] Update guardrails section (lines 581-582): change `the command is for Speckit feature branches only` to `the command is for Speckit ('NNN-*') and OpenSpec ('opsx/*') feature branches` per data-model.md §8 (FR-008)

**Checkpoint**: At this point, `/unleash` should detect `opsx/*` branches, extract the change name, verify tasks.md exists, skip clarify/plan/tasks, and fall through to spec review. Speckit branches should behave identically to before.

---

## Phase 2: US2 — Resumability for OpenSpec (Priority: P1)

**Goal**: Ensure the resumability detection in Step 2 works correctly for OpenSpec changes — clarify/plan/tasks are always "done", and spec-review/implementation/code-review markers use `FEATURE_DIR`.

**Independent Test**: Run `/unleash` on an `opsx/*` branch. After spec review passes, interrupt. Re-run `/unleash`. Verify it skips spec review and resumes at implementation.

### Implementation

- [x] T008 [US2] In Step 2 — Resumability Detection (lines 84-124): add OpenSpec conditional — when `WORKFLOW_TIER = openspec`, report clarify/plan/tasks as always "done" (skip checks 1-3) per data-model.md §4 and contracts §Resumability Markers (FR-006)
- [x] T009 [US2] In Step 2 — Resumability Detection: ensure checks 4-6 (spec-review marker, all `[x]`, code-review marker) read from `FEATURE_DIR/tasks.md` for both workflow tiers per data-model.md §4 (FR-006, FR-007)

**Checkpoint**: Resumability should work end-to-end for OpenSpec — interrupted runs resume from the correct step.

---

## Phase 3: US4 — Spec Review & Remaining Steps with FEATURE_DIR (Priority: P2)

**Goal**: Ensure Steps 4-8 use `FEATURE_DIR` for artifact paths so both Speckit and OpenSpec modes work identically through the review/implement/demo pipeline.

**Independent Test**: Run `/unleash` on an `opsx/*` branch. Verify the review council receives `openspec/changes/<name>/` as the review scope and announces OpenSpec workflow tier.

### Implementation

- [x] T010 [US4] In Step 4 — Spec Review (lines 230-279): update the review council delegation to pass `FEATURE_DIR` as the review scope instead of implicit `specs/` references; remove explicit "Spec Review Mode" Speckit-specific artifact assumptions per data-model.md §6 (FR-005, FR-007)
- [x] T011 [US4] In Step 5 — Implement (lines 281-424): ensure `tasks.md` is read from `FEATURE_DIR/tasks.md` per data-model.md §7 (FR-007)
- [x] T012 [US4] In Step 6 — Code Review (lines 426-487): ensure the review council receives `FEATURE_DIR` as the review scope per data-model.md §7 (FR-007)
- [x] T013 [US4] In Step 8 — Demo (lines 525-577): add conditional for OpenSpec — read `proposal.md` instead of `spec.md` for "What Was Built", and skip `quickstart.md` reference (OpenSpec changes don't have one) per data-model.md §7 and artifact mapping table (FR-007)

**Checkpoint**: The full pipeline (spec review → implement → code review → retrospective → demo) should work for both Speckit and OpenSpec modes using the unified `FEATURE_DIR` variable.

---

## Phase 4: Scaffold Sync & Validation

**Purpose**: Synchronize the scaffold asset copy and verify all tests pass.

- [x] T014 [P] [US1] Copy the modified `.opencode/command/unleash.md` to `internal/scaffold/assets/opencode/command/unleash.md` — both files must be byte-identical per scaffold sync pattern (FR-009)
- [x] T015 Run `go test -race -count=1 ./internal/scaffold/...` to verify `TestEmbeddedAssets_MatchSource` passes (scaffold drift detection) (FR-010)
- [x] T016 Run `go test -race -count=1 ./...` to verify all existing tests pass (FR-010)
- [x] T017 Run quickstart.md verification scenarios: test `/unleash` on `opsx/*` branch (US1), test resumability (US2), test skip announcement (US3), test backward compatibility on `NNN-*` branch (SC-003)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1** (US1+US3): No dependencies — can start immediately. This is the core change.
- **Phase 2** (US2): Depends on T003/T005 from Phase 1 (needs `WORKFLOW_TIER` variable established in Step 1).
- **Phase 3** (US4): Depends on T003/T005 from Phase 1 (needs `FEATURE_DIR` variable established in Step 1).
- **Phase 4** (Sync): Depends on all previous phases — scaffold copy must reflect final state.

### Within Phase 1

- T001 and T002 are independent description edits (could be [P] but same file region).
- T003 must precede T004 (detection before prerequisite check).
- T005 depends on T003 (wraps existing code in conditional).
- T006 depends on T003 (references `WORKFLOW_TIER`).
- T007 is independent of T003-T006 (different section of file).

### Within Phase 3

- T010, T011, T012, T013 modify different sections of the same file but are logically independent. Execute sequentially to avoid edit conflicts.

### Parallel Opportunities

- T014 is marked [P] because it is a file copy operation independent of the test runs.
- Phases 2 and 3 could run in parallel (different sections of the file, no overlapping edits) but are sequenced for clarity since it's a single file.

---

## FR Traceability

| FR | Tasks | Verified By |
|----|-------|-------------|
| FR-001 | T001, T002, T003, T007 | SC-001 |
| FR-002 | T003 | SC-001 |
| FR-003 | T004 | SC-001 |
| FR-004 | T006 | SC-001 |
| FR-005 | T010 | SC-002 |
| FR-006 | T008, T009 | SC-004 |
| FR-007 | T009, T010, T011, T012, T013 | SC-001 |
| FR-008 | T005, T007 | SC-003 |
| FR-009 | T014 | SC-005 |
| FR-010 | T015, T016 | SC-005 |

<!-- spec-review: passed -->
