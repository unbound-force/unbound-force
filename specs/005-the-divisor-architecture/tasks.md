# Tasks: The Divisor Architecture (PR Reviewer Council)

**Input**: Design documents from `/specs/005-the-divisor-architecture/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/scaffold-cli.md

**Tests**: Tests are included — the plan specifies drift detection tests and unit tests for new Go functions as part of the constitution's Testability principle.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Create canonical source files for Divisor assets before embedding them in the scaffold engine.

- [x] T001 Create directory structure for convention packs at `.opencode/divisor/packs/`

---

## Phase 2: Foundational (Convention Packs)

**Purpose**: Convention packs are consumed by all persona agents (US1) and the scaffold engine (US4). They MUST exist before agents can reference them.

**CRITICAL**: No user story work can begin until these packs exist as canonical source files.

- [x] T002 [P] Create Go convention pack at `.opencode/divisor/packs/go.md` with YAML frontmatter (`pack_id: go`, `language: Go`, `version: 1.0.0`) and six H2 sections: Coding Style (CS-001 through CS-013), Architectural Patterns (AP-001 through AP-007), Security Checks (SC-001 through SC-006), Testing Conventions (TC-001 through TC-012), Documentation Requirements (DR-001 through DR-005), Custom Rules (empty placeholder). Extract Go-specific rules from existing `reviewer-*.md` agents per research.md R8.
- [x] T003 [P] Create language-agnostic default convention pack at `.opencode/divisor/packs/default.md` with YAML frontmatter (`pack_id: default`, `language: Any`, `version: 1.0.0`) and same six H2 sections with universal software engineering rules (SOLID, DRY, error handling, test coverage). Use CS-NNN/AP-NNN/SC-NNN/TC-NNN/DR-NNN identifiers with `[MUST]`/`[SHOULD]`/`[MAY]` severity indicators.
- [x] T004 [P] Create TypeScript convention pack at `.opencode/divisor/packs/typescript.md` with YAML frontmatter (`pack_id: typescript`, `language: TypeScript`, `version: 1.0.0`) and same six H2 sections with TypeScript-specific rules (ESLint, JSDoc, no `any`, import organization, error handling patterns).
- [x] T005 [P] Create Go custom rules stub at `.opencode/divisor/packs/go-custom.md` with empty Custom Rules section and `<!-- Add project-specific rules below this line -->` marker.
- [x] T006 [P] Create default custom rules stub at `.opencode/divisor/packs/default-custom.md` with empty Custom Rules section.
- [x] T007 [P] Create TypeScript custom rules stub at `.opencode/divisor/packs/typescript-custom.md` with empty Custom Rules section.

**Checkpoint**: All convention packs exist as canonical source files. Agents and scaffold engine can now reference them.

---

## Phase 3: User Story 1 — Dynamic Review Protocol (Priority: P1) MVP

**Goal**: Create five canonical `divisor-*.md` persona agents with dynamic convention pack loading and a `/review-council` command that discovers them via `divisor-*.md` pattern.

**Independent Test**: Run `/review-council` in this repo after deploying the new agents. Verify all five personas are discovered, produce structured verdicts, and the council decision follows the protocol (APPROVE only if no REQUEST CHANGES).

### Implementation for User Story 1

- [x] T008 [P] [US1] Create `divisor-guard.md` at `.opencode/agents/divisor-guard.md`. Structure: YAML frontmatter (description, mode: subagent, model, temperature: 0.1, tools: write/edit/bash false), H1 Role (generic intent drift detector — no project-specific references), H2 Source Documents (AGENTS.md, constitution, specs, convention pack loading instruction to read all `*.md` from `.opencode/divisor/packs/`), H2 Code Review Mode with universal checklist sections (Intent Drift Detection, Constitution Alignment, Zero-Waste Mandate, Neighborhood Rule [PACK]), H2 Spec Review Mode, H2 Output Format (structured finding template with severity/file/description/recommendation), H2 Decision Criteria (APPROVE/REQUEST CHANGES rules). Derive from existing `reviewer-guard.md` but remove all unbound-force-specific content per research.md R8 classification.
- [x] T009 [P] [US1] Create `divisor-architect.md` at `.opencode/agents/divisor-architect.md`. Same structural template as T008 but with Architect-specific focus: H2 Code Review Mode with universal sections (Architectural Alignment, DRY/Structural Integrity, Plan Alignment) plus `[PACK]` sections (Coding Convention Compliance referencing `coding_style`/`architectural_patterns`, Testing Convention Compliance referencing `testing_conventions`, Documentation Compliance referencing `documentation_requirements`). Include Architectural Alignment Score (1-10) per existing reviewer-architect.md pattern.
- [x] T010 [P] [US1] Create `divisor-adversary.md` at `.opencode/agents/divisor-adversary.md`. Same structural template but with Adversary-specific focus: universal sections (Zero-Waste, Error Handling/Resilience, Efficiency, Test Safety, Universal Security — hardcoded secrets, injection patterns per FR-020) plus `[PACK]` sections (Language-Specific Error Patterns referencing `security_checks`, Framework-Specific Resilience referencing `custom_rules`, Dependency Vulnerabilities). Include `read: true` in YAML frontmatter tools.
- [x] T011 [P] [US1] Create `divisor-sre.md` at `.opencode/agents/divisor-sre.md`. Same structural template but with SRE-specific focus: universal sections (Runtime Observability, Operational Documentation, Backup/Recovery, Upgrade/Migration Paths) plus `[PACK]` sections (Release Pipeline Integrity referencing `architectural_patterns`, Dependency Health, Configuration/Environment).
- [x] T012 [P] [US1] Create `divisor-testing.md` at `.opencode/agents/divisor-testing.md`. Same structural template but with Testing-specific focus: universal sections (Coverage Strategy, Assertion Depth, Test Isolation, Regression Protection) plus `[PACK]` sections (Test Architecture Conventions referencing `testing_conventions`, Assertion Convention Compliance, Convention Compliance for test execution).
- [x] T013 [US1] Update `/review-council` command at `.opencode/command/review-council.md`: change discovery pattern from `reviewer-*.md` to `divisor-*.md`, update Known Reviewer Roles reference table entries from `reviewer-adversary` to `divisor-adversary` (etc. for all 5), update guard clause error message to reference `divisor-*.md` pattern, update absent reviewer note to reference "Divisor persona roles".
- [x] T014 [US1] Update `/speckit.testreview` command at `.opencode/command/speckit.testreview.md`: change `reviewer-testing` reference to `divisor-testing`.

**Checkpoint**: Five `divisor-*.md` agents exist, `/review-council` discovers them via `divisor-*.md` pattern. Can test by running `/review-council` in this repo.

---

## Phase 4: User Story 2 — Convention Packs Integration (Priority: P1)

**Goal**: Verify that persona agents correctly load and apply convention pack content at review time. The Go pack produces review behavior equivalent to the existing Gaze prototype. The default pack provides meaningful universal feedback.

**Independent Test**: Deploy Go and default convention packs, run `/review-council` on a Go PR, verify The Architect cites Go-specific conventions from the pack (gofmt, GoDoc, error wrapping). Then remove the Go pack, re-run, verify personas fall back to default pack and note "no language-specific pack loaded."

### Implementation for User Story 2

- [x] T015 [US2] Validate Go convention pack (`.opencode/divisor/packs/go.md`) produces equivalent review behavior to existing `reviewer-architect.md` Go checks. Run `/review-council` on a sample change in this repo and compare findings against what `reviewer-architect.md` would produce. Adjust Go pack rules if gaps are found.
- [x] T016 [US2] Validate default convention pack (`.opencode/divisor/packs/default.md`) produces meaningful findings on a non-Go project. Verify that universal principles (SOLID, DRY, error handling, test coverage) are checked. Adjust default pack rules if gaps are found.
- [x] T017 [US2] Validate graceful degradation: temporarily remove all packs from `.opencode/divisor/packs/`, run `/review-council`, verify agents skip `[PACK]` sections and note "No convention pack found" in their findings. Restore packs after validation.

**Checkpoint**: Convention packs are loaded dynamically at review time. Go pack matches prototype behavior. Default pack provides universal coverage. Graceful degradation works when no pack exists.

---

## Phase 5: User Story 3 — Project-Aware Review Context (Priority: P2)

**Goal**: Persona agents read the target project's constitution, active spec, and AGENTS.md to produce context-sensitive reviews. The Guard detects intent drift, The Architect validates constitution compliance, The Adversary checks spec edge cases.

**Independent Test**: Run `/review-council` on a PR that violates a spec acceptance criterion. Verify The Guard flags the intent drift. Then remove the constitution file, re-run, verify agents fall back to convention-pack-only review and note "project context unavailable."

### Implementation for User Story 3

- [x] T018 [P] [US3] Verify `divisor-guard.md` Source Documents section instructs reading AGENTS.md, constitution, and active spec. Verify Code Review Mode checklist includes Intent Drift Detection that references the active spec's user stories and acceptance criteria. Adjust if needed.
- [x] T019 [P] [US3] Verify `divisor-architect.md` Source Documents section instructs reading constitution. Verify Code Review Mode checklist includes constitution principle compliance checks. Adjust if needed.
- [x] T020 [P] [US3] Verify `divisor-adversary.md` Source Documents section instructs reading spec edge cases. Verify Code Review Mode checklist includes edge case coverage verification. Adjust if needed.
- [x] T021 [US3] Validate fallback behavior: in a project with no constitution or spec files, verify all agents produce convention-pack-only reviews and note "project context unavailable." This should already work from the graceful degradation built into the agent templates in Phase 3.

**Checkpoint**: Project-aware review context works. Agents produce more targeted findings when constitution/spec are available, and gracefully degrade when they are not.

---

## Phase 6: User Story 4 — Deployment via `unbound init` (Priority: P2)

**Goal**: Extend the scaffold engine with `--divisor` and `--lang` flags. `unbound init` deploys all 45 files (including Divisor assets). `unbound init --divisor` deploys only the Divisor subset. Language auto-detection selects the correct convention pack.

**Independent Test**: Run `unbound init --divisor` in a temp Go project, verify only Divisor files are created. Run `unbound init --divisor --lang typescript` in a temp empty project, verify TypeScript pack is deployed instead of Go.

### Tests for User Story 4

- [x] T022 [US4] Add `isDivisorAsset()` function to `internal/scaffold/scaffold.go` and create `TestIsDivisorAsset` table-driven test in `internal/scaffold/scaffold_test.go` per contracts/scaffold-cli.md. Predicate matches: `opencode/agents/divisor-*.md` (prefix), `opencode/command/review-council.md` (exact), `opencode/divisor/packs/*` (prefix). Test both positive cases (all Divisor files) and negative cases (reviewer-*, constitution-check, speckit commands).
- [x] T023 [US4] Add `detectLang()` function to `internal/scaffold/scaffold.go` and create `TestDetectLang` table-driven test in `internal/scaffold/scaffold_test.go`. Check marker files in priority order: go.mod→"go", tsconfig.json→"typescript", package.json→"typescript", pyproject.toml→"python", Cargo.toml→"rust". Return "" if none found. Test with `t.TempDir()` and marker file creation.
- [x] T024 [US4] Add `isConventionPack()` and `shouldDeployPack()` functions to `internal/scaffold/scaffold.go` and create `TestShouldDeployPack` table-driven test in `internal/scaffold/scaffold_test.go`. `shouldDeployPack(relPath, lang)` returns true for: `{lang}.md`, `{lang}-custom.md`, `default.md`, `default-custom.md`. Returns false for other packs. Non-pack files always return true.

> **Note**: T022-T024 are logically independent but modify the same files (`scaffold.go`, `scaffold_test.go`). Execute sequentially to avoid merge conflicts.

### Implementation for User Story 4

- [x] T025 [US4] Add `DivisorOnly bool` and `Lang string` fields to `Options` struct in `internal/scaffold/scaffold.go`. Modify `Run()` function: add language resolution block (Lang flag → auto-detect → "default" fallback), add `isDivisorAsset` filter in `fs.WalkDir` callback when `DivisorOnly=true`, add `shouldDeployPack` filter for convention pack language selection, suppress `openspec/specs` and `openspec/changes` empty directory creation when `DivisorOnly=true`.
- [x] T026 [US4] Update `isToolOwned()` in `internal/scaffold/scaffold.go`: add convention pack ownership rules — `opencode/divisor/packs/*.md` files are tool-owned UNLESS filename contains `-custom` (those are user-owned).
- [x] T027 [US4] Update `printSummary()` in `internal/scaffold/scaffold.go`: accept `divisorOnly` parameter, show Divisor-specific hint line ("Run /review-council to start a code review.") when `divisorOnly=true`. Add informational note when language could not be detected and `--lang` was not provided.
- [x] T028 [US4] Add `--divisor` and `--lang` flags to CLI in `cmd/unbound/main.go`: add `divisorOnly bool` and `lang string` to `initParams` struct, register `--divisor` (bool, default false) and `--lang` (string, default "") flags on init command, pass through to `scaffold.Options`. Update Long description to document the new flags.
- [x] T029 [P] [US4] Copy all new canonical source files to embedded assets directory. Copy `.opencode/agents/divisor-guard.md` to `internal/scaffold/assets/opencode/agents/divisor-guard.md`. Repeat for `divisor-architect.md`, `divisor-adversary.md`, `divisor-sre.md`, `divisor-testing.md`. Copy `.opencode/command/review-council.md` to `internal/scaffold/assets/opencode/command/review-council.md`. Copy `.opencode/divisor/packs/go.md`, `default.md`, `typescript.md`, `go-custom.md`, `default-custom.md`, `typescript-custom.md` to `internal/scaffold/assets/opencode/divisor/packs/`.
- [x] T030 [US4] Update `expectedAssetPaths` in `internal/scaffold/scaffold_test.go`: add 12 new entries (5 divisor agents, 1 review-council command, 3 canonical packs, 3 custom pack stubs). Update `knownNonEmbeddedFiles`: remove `review-council.md` (now embedded). Add `.opencode/divisor/packs` to `expectedDirs` in `TestRun_CreatesFiles`. Add `.opencode/divisor` to `canonicalDirs` in `TestCanonicalSources_AreEmbedded`.
- [x] T031 [US4] Add `TestIsToolOwned` cases for convention packs in `internal/scaffold/scaffold_test.go`: `opencode/divisor/packs/go.md` → true (tool-owned), `opencode/divisor/packs/go-custom.md` → false (user-owned), `opencode/agents/divisor-guard.md` → false (user-owned agents).
- [x] T032 [US4] Create `TestRun_DivisorSubset` integration test in `internal/scaffold/scaffold_test.go`: create temp dir with `go.mod`, run `Run()` with `DivisorOnly=true`, verify only Divisor files created (agents, command, Go+default packs), verify no speckit/openspec files created, verify no openspec empty dirs created, verify summary mentions review-council.
- [x] T033 [US4] Create `TestRun_DivisorSubset_WithLangFlag` integration test in `internal/scaffold/scaffold_test.go`: run with `DivisorOnly=true, Lang="typescript"` in empty temp dir, verify TypeScript pack deployed and Go pack NOT deployed, verify all 5 agent files still created.
- [x] T034 [US4] Create `TestRun_DivisorSubset_DefaultFallback` integration test in `internal/scaffold/scaffold_test.go`: run with `DivisorOnly=true` in empty temp dir (no language markers), verify only `default.md` and `default-custom.md` packs deployed, verify informational message about no language detection in output.
- [x] T035 [US4] Run `go test -race -count=1 ./...` and verify all tests pass including new and existing drift detection tests. Fix any failures.
- [x] T036 [US4] Run `go build ./...` and verify the build succeeds with the new embedded assets and code changes.

**Checkpoint**: `unbound init --divisor` deploys the Divisor subset. `unbound init --divisor --lang typescript` deploys TypeScript pack. `unbound init` (full) deploys all 45 files including Divisor assets. All tests pass.

---

## Phase 7: User Story 5 — Review Report Artifact (Priority: P3)

**Goal**: The `/review-council` command produces a structured Markdown review report with all required sections: discovery summary, persona verdicts, council decision, iteration history, metadata.

**Independent Test**: Run `/review-council` on a code change in this repo. Verify the final report includes: discovery summary (which personas found, which absent), each persona's verdict with structured findings (severity, category, file, description, recommendation), the council decision (APPROVED/CHANGES_REQUESTED), iteration count, and metadata (convention pack used).

### Implementation for User Story 5

- [x] T037 [US5] Verify `/review-council` command (`.opencode/command/review-council.md`) final report section includes: discovery summary listing invoked and absent personas, per-persona verdict sections with structured findings, council decision with aggregation logic, iteration history (if iterations occurred), metadata (convention pack used, review timestamp). Adjust the report template in the command if any sections are missing per FR-017 and FR-018.
- [x] T038 [US5] Validate structured Markdown report is machine-parseable: verify finding format includes `### [SEVERITY] Finding Title`, `**File**: path`, `**Description**: text`, `**Recommendation**: text` consistently across all persona outputs. Verify discovery summary uses consistent format for invoked/absent personas.

**Checkpoint**: Review reports are structured, machine-parseable, and contain all required sections. JSON artifact output is deferred to Spec 009.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, migration support, and cleanup.

- [x] T039 [P] Update `AGENTS.md`: add The Divisor to the Heroes table (change status from "Spec only (005)" to "Implemented"), add `divisor-*.md` agents to the Project Structure tree, update Active Technologies section if needed, add convention pack directory to structure.
- [x] T040 [P] Update `unbound-force.md`: verify hero descriptions reflect The Divisor's implemented status and embedded distribution model.
- [x] T041 [P] Update `specs/005-the-divisor-architecture/spec.md`: change frontmatter `status: draft` to `status: complete`, change body `**Status**: Draft` to `**Status**: Complete`.
- [x] T042 Add migration documentation: create a section in the quickstart.md or release notes documenting the `reviewer-*` to `divisor-*` migration path (already outlined in `quickstart.md` — verify it is complete and accurate).
- [x] T043 Run `make check` (or `go build ./... && go test -race -count=1 ./... && golangci-lint run`) to verify all checks pass with the complete implementation.
- [x] T044 Verify SC-001 through SC-008 success criteria from spec.md are met. Document verification results.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — convention packs MUST exist before agents reference them
- **US1 (Phase 3)**: Depends on Phase 2 — agents load convention packs from `.opencode/divisor/packs/`
- **US2 (Phase 4)**: Depends on Phase 3 — validates pack integration with agents
- **US3 (Phase 5)**: Depends on Phase 3 — validates project-context features in agents
- **US4 (Phase 6)**: Depends on Phase 3 — embeds agent files created in Phase 3 into scaffold
- **US5 (Phase 7)**: Depends on Phase 3 — validates report output from `/review-council`
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: Depends on Foundational (Phase 2) only — core agents and protocol
- **US2 (P1)**: Depends on US1 — validates convention pack integration with agents created in US1
- **US3 (P2)**: Depends on US1 — validates project-context features in agents created in US1. Can run in parallel with US2.
- **US4 (P2)**: Depends on US1 — embeds files created in US1 into scaffold. Can run in parallel with US2/US3.
- **US5 (P3)**: Depends on US1 — validates report output. Can run in parallel with US2/US3/US4.

### Within Each User Story

- Tests written and verified to fail before implementation (US4 tests)
- Agent files can be created in parallel ([P] marked tasks)
- Integration/validation tasks are sequential within a story

### Parallel Opportunities

- **Phase 2**: All 6 convention pack tasks (T002-T007) can run in parallel
- **Phase 3**: All 5 agent creation tasks (T008-T012) can run in parallel
- **Phase 5**: All 3 context verification tasks (T018-T020) can run in parallel
- **Phase 6**: Test tasks T022-T024 are sequential (same files). Asset copy T029 can run in parallel with other non-file-conflicting tasks.
- **Phase 8**: Documentation tasks T039-T041 can run in parallel
- **US2, US3, US5**: Can proceed in parallel after US1 completes

---

## Parallel Example: User Story 1

```bash
# Launch all five agent creation tasks together (different files):
Task: "Create divisor-guard.md at .opencode/agents/divisor-guard.md"
Task: "Create divisor-architect.md at .opencode/agents/divisor-architect.md"
Task: "Create divisor-adversary.md at .opencode/agents/divisor-adversary.md"
Task: "Create divisor-sre.md at .opencode/agents/divisor-sre.md"
Task: "Create divisor-testing.md at .opencode/agents/divisor-testing.md"

# Then sequentially:
Task: "Update review-council.md discovery pattern"
Task: "Update speckit.testreview.md reference"
```

## Parallel Example: User Story 4

```bash
# Sequential: test function tasks (same files — scaffold.go, scaffold_test.go):
Task: "Add isDivisorAsset() and TestIsDivisorAsset"
Task: "Add detectLang() and TestDetectLang"
Task: "Add shouldDeployPack() and TestShouldDeployPack"

# Sequential: implementation tasks (depend on functions above):
Task: "Add DivisorOnly/Lang to Options, modify Run()"
Task: "Update isToolOwned() for convention packs"
Task: "Update printSummary() for divisor mode"

# Parallel: these touch different files:
Task: "Add --divisor and --lang flags to CLI"        # cmd/unbound/main.go
Task: "Copy canonical files to embedded assets"       # internal/scaffold/assets/

# Sequential: depends on all above:
Task: "Update expectedAssetPaths and drift detection"
Task: "Run all tests"
```

---

## Implementation Strategy

### MVP First (User Story 1 + 2)

1. Complete Phase 1: Setup (create directories)
2. Complete Phase 2: Foundational (convention packs)
3. Complete Phase 3: US1 — Dynamic Review Protocol (agents + command)
4. Complete Phase 4: US2 — Validate convention pack integration
5. **STOP and VALIDATE**: Run `/review-council` in this repo, verify 5 personas discovered, structured verdicts produced, convention packs loaded
6. The Divisor is now functional for local use

### Incremental Delivery

1. Setup + Foundational + US1 + US2 → The Divisor works locally (MVP)
2. Add US3 → Project-aware context (constitution/spec-driven reviews)
3. Add US4 → Distribution via `unbound init --divisor` (other projects can use it)
4. Add US5 → Structured review reports for inter-hero consumption
5. Polish → Documentation, migration support, success criteria validation

### Parallel Team Strategy

With multiple developers after Phase 2 + US1:
- Developer A: US2 (convention pack validation)
- Developer B: US3 (project-aware context)
- Developer C: US4 (scaffold engine integration)
- Developer D: US5 (review report artifact)
All stories complete and integrate independently.
<!-- scaffolded by unbound vdev -->
