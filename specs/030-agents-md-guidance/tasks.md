# Tasks: AGENTS.md Behavioral Guidance Injection

**Input**: Design documents from `specs/030-agents-md-guidance/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/uf-init-guidance.md, quickstart.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Scope

This feature modifies exactly **2 files** — both Markdown,
no Go code:

| # | File | Change |
|---|------|--------|
| 1 | `.opencode/command/uf-init.md` | Add Step 9, renumber Step 9→10, extend report |
| 2 | `internal/scaffold/assets/opencode/command/uf-init.md` | Byte-identical sync copy |

---

## Phase 1: Command Structure (Blocking Prerequisites)

**Purpose**: Renumber existing steps and establish the new
Step 9 skeleton before injecting guidance block content.

**⚠️ CRITICAL**: All guidance block tasks depend on the
skeleton being in place first.

- [x] T001 [US1] Renumber existing "Step 9: Report Results" to "Step 10: Report Results" in `.opencode/command/uf-init.md` — update the heading `### Step 9: Report Results` to `### Step 10: Report Results`; update any internal references to "Step 9" that refer to the report
- [x] T002 [US1] Insert new `### Step 9: AGENTS.md Behavioral Guidance` section skeleton in `.opencode/command/uf-init.md` between Step 8 (OpenSpec Command Guardrails, ends ~line 517) and the new Step 10 — include the preamble instructions: (1) check if `AGENTS.md` exists at repo root, (2) if not found report `⊘ AGENTS.md: not found (skipped)` and skip entire step, (3) if found read `AGENTS.md` and process each guidance block per the detection/injection pattern
- [x] T003 [US1] Add idempotency detection instructions to Step 9 skeleton in `.opencode/command/uf-init.md` — for each block: search for detection phrases (heading text + semantic fallback), if found report `⊘ <block>: already present (skipped)`, if not found find correct insertion point and inject, report `✅ <block>: injected`

**Checkpoint**: Step 9 skeleton is in place with the
AGENTS.md-not-found guard and detection/injection pattern.
Step 10 is the report. Ready for guidance block content.

---

## Phase 2: P1 Quality Gates (US2 — Consistent Quality Gates)

**Purpose**: Inject the 3 most critical behavioral rules
that prevent the most damaging agent behaviors.

**Goal**: Gatekeeping value protection, workflow phase
boundaries, and CI parity gate are present in every repo
after `/uf-init`.

**Independent Test**: Run `/uf-init` in the Gaze repo
(which has CI parity gate but not gatekeeping protection
or phase boundaries). Verify all 3 are present after.

- [x] T004 [US2] Add Block 2 (Gatekeeping Value Protection) to Step 9 in `.opencode/command/uf-init.md` — detection phrases: `Gatekeeping Value Protection` heading, `MUST NOT modify values that serve as quality`; placement: inside `## Behavioral Constraints` section; inject full text from data-model.md Block 2 (8 protected categories + stop-and-report instruction)
- [x] T005 [US2] Add Block 3 (Workflow Phase Boundaries) to Step 9 in `.opencode/command/uf-init.md` — detection phrases: `Workflow Phase Boundaries` heading, `MUST NOT cross workflow phase boundaries`; placement: inside `## Behavioral Constraints` after Gatekeeping; inject full text from data-model.md Block 3 (phase-to-output mapping + stop-and-report)
- [x] T006 [US2] Add Block 4 (CI Parity Gate) to Step 9 in `.opencode/command/uf-init.md` — detection phrases: `CI Parity Gate` heading or bold text, `replicate the CI checks locally`; placement: inside `## Behavioral Constraints` or `## Technical Guardrails`; inject full text from data-model.md Block 4 (read workflows, execute locally, blocking error)

**Checkpoint**: The 3 highest-impact behavioral controls
are defined in Step 9. These prevent threshold tampering,
out-of-phase code changes, and CI skipping.

---

## Phase 3: P2 Workflow Rules (US3 — Workflow and Documentation Rules)

**Purpose**: Inject the 3 workflow/documentation rules
that enforce process discipline.

**Goal**: Review council prerequisite, website doc sync
gate, and spec-first development are present in every
repo after `/uf-init`.

**Independent Test**: Run `/uf-init` in the Website repo
(which has minimal behavioral guidance). Verify all 3
workflow rules are present after.

- [x] T007 [US3] Add Block 5 (Review Council PR Prerequisite) to Step 9 in `.opencode/command/uf-init.md` — detection phrases: `Review Council` + `PR Prerequisite`, `/review-council`; placement: after behavioral constraints, before build commands; inject full text from data-model.md Block 5 (5-step workflow + exemptions list)
- [x] T008 [US3] Add Block 6 (Website Documentation Sync Gate) to Step 9 in `.opencode/command/uf-init.md` — detection phrases: `Website Documentation` + `Gate`, `gh issue create --repo`, `unbound-force/website`; placement: near documentation validation gate or spec commit gate; inject full text from data-model.md Block 6 (`gh issue create` template + exempt/required lists)
- [x] T009 [US3] Add Block 7 (Spec-First Development) to Step 9 in `.opencode/command/uf-init.md` — detection phrases: `Spec-First Development`, `preceded by a spec workflow`; placement: after behavioral constraints, before build commands; inject full text from data-model.md Block 7 (what requires spec + what is exempt + when-unsure rule)

**Checkpoint**: All 6 behavioral gate blocks (P1 + P2)
are defined in Step 9. Process discipline rules are
complete.

---

## Phase 4: P3 Contextual Guidance (US4 — Knowledge Retrieval and Core Mission)

**Purpose**: Inject the 2 contextual guidance sections
that improve agent output quality.

**Goal**: Knowledge retrieval (Dewey tool matrix + 3-tier
degradation) and core mission (3 strategic framing
bullets) are present in every repo after `/uf-init`.

**Independent Test**: Run `/uf-init` in the Replicator
repo (which has knowledge retrieval but no core mission).
Verify the core mission is added and the existing
knowledge retrieval is preserved.

- [x] T010 [US4] Add Block 1 (Core Mission) to Step 9 in `.opencode/command/uf-init.md` — detection phrases: `## Core Mission`, `Strategic Architecture`, `Outcome Orientation`; placement: after `## Project Overview`, before `## Behavioral Constraints`; inject full text from data-model.md Block 1 (3 strategic framing bullets)
- [x] T011 [US4] Add Block 8 (Knowledge Retrieval) to Step 9 in `.opencode/command/uf-init.md` — detection phrases: `## Knowledge Retrieval`, `dewey_semantic_search`, `Tool Selection Matrix`; placement: after coding conventions, before testing conventions; inject full text from data-model.md Block 8 (tool selection matrix + fallback criteria + 3-tier degradation pattern)

**Checkpoint**: All 8 guidance blocks are defined in
Step 9. The full injection catalog is complete.

---

## Phase 5: Report Extension and Finalization (US1)

**Purpose**: Extend the report template, sync the scaffold
asset copy, and verify the build.

- [x] T012 [US1] Add "AGENTS.md Guidance" section to the report template in Step 10 of `.opencode/command/uf-init.md` — insert between "OpenSpec Command Guardrails" and "Summary" sections; include 8 status lines (one per block: Core Mission, Gatekeeping Value Protection, Workflow Phase Boundaries, CI Parity Gate, Review Council PR Prerequisite, Website Documentation Sync Gate, Spec-First Development, Knowledge Retrieval) using `[status] <block>: [action]` format; update Summary counters to include AGENTS.md guidance results
- [x] T013 [US1] Add injection order instructions to Step 9 in `.opencode/command/uf-init.md` — when multiple blocks are missing, inject in this order per data-model.md: (1) Core Mission, (2) Gatekeeping Value Protection, (3) Workflow Phase Boundaries, (4) CI Parity Gate, (5) Review Council PR Prerequisite, (6) Spec-First Development, (7) Website Documentation Sync Gate, (8) Knowledge Retrieval
- [x] T014 [US1] Sync scaffold asset copy — copy `.opencode/command/uf-init.md` to `internal/scaffold/assets/opencode/command/uf-init.md` (must be byte-identical per contract)
- [x] T015 [US1] Add content-presence assertions to an existing scaffold test — verify the `/uf-init` scaffold asset contains all 8 guidance block detection phrases (one per block: `Core Mission`, `Gatekeeping Value Protection`, `Workflow Phase Boundaries`, `CI Parity Gate`, `Review Council`, `Website Documentation`, `Spec-First Development`, `Knowledge Retrieval`); this catches missing or truncated blocks without new test infrastructure
- [x] T016 [US1] Verify build and drift detection — run `go build ./...` and `go test -race -count=1 ./internal/scaffold/...` to confirm `TestScaffoldAssetDrift` passes, content-presence assertions pass, and no existing tests are broken

**Checkpoint**: Implementation complete. Both files are
modified, asset copy is synchronized, and all tests pass.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Command Structure)**: No dependencies — can
  start immediately. BLOCKS all subsequent phases.
- **Phase 2 (P1 Quality Gates)**: Depends on Phase 1
  completion (Step 9 skeleton must exist).
- **Phase 3 (P2 Workflow Rules)**: Depends on Phase 1
  completion. Can run in parallel with Phase 2 (all tasks
  modify the same file but different sections within
  Step 9).
- **Phase 4 (P3 Contextual Guidance)**: Depends on Phase 1
  completion. Can run in parallel with Phases 2-3.
- **Phase 5 (Report & Finalization)**: Depends on Phases
  2-4 completion (all 8 blocks must be defined before
  the report can reference them and the asset can be
  synced).

### Within Each Phase

- Phase 1: T001 → T002 → T003 (sequential — each builds
  on the previous)
- Phase 2: T004, T005, T006 modify the same Step 9
  section but are independent blocks — execute
  sequentially to avoid edit conflicts in the same file
- Phase 3: T007, T008, T009 — same pattern as Phase 2
- Phase 4: T010, T011 — same pattern
- Phase 5: T012 → T013 → T014 → T015 → T016 (sequential —
  report before sync, sync before test, test before verify)

### File Conflict Note

All tasks in Phases 1-5 modify the same file
(`.opencode/command/uf-init.md`). While the blocks are
conceptually independent, they MUST be executed
sequentially to avoid concurrent edit conflicts. The
`[P]` marker is intentionally omitted from all tasks.

---

## Implementation Strategy

### Single-File Sequential

This feature modifies 1 primary file with a sync copy.
All tasks are sequential within the primary file:

1. Complete Phase 1: Skeleton → Step 9 exists
2. Complete Phase 2: P1 blocks → 3 critical gates defined
3. Complete Phase 3: P2 blocks → 3 workflow rules defined
4. Complete Phase 4: P3 blocks → 2 contextual sections
   defined
5. Complete Phase 5: Report + sync + verify → Done

### Verification

After Phase 5, the implementation is verified by:
- `TestScaffoldAssetDrift` — asset copy matches live file
- `go build ./...` — no compilation errors
- `go test -race -count=1 ./...` — all tests pass
- Manual: only 2 files modified (`git diff --stat`)

---

## Notes

- All 8 guidance block texts come from `data-model.md`
  (the canonical content reference for this spec)
- No Go code is modified — this is a Markdown-only change
- No new embedded assets — `expectedAssetPaths` count
  does not change
- No new test files — existing `TestScaffoldAssetDrift`
  covers the sync requirement
- The injection order in T013 differs slightly from the
  block numbering (Blocks 1-8) because the order is
  optimized for document flow, not priority

<!-- spec-review: passed -->
