---

description: "Task list for Divisor Council Refinement"
---
<!-- scaffolded by uf vdev -->

# Tasks: Divisor Council Refinement

**Input**: Design documents from `/specs/019-divisor-council-refinement/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, quickstart.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Scaffold Asset Cleanup (US1 — Zero-Waste, P1) 🎯 MVP

**Goal**: Remove legacy `reviewer-*.md` files from the scaffold asset set, add legacy file detection warning to `uf init`, and update all test expectations.

**Independent Test**: Run `go test -race -count=1 ./internal/scaffold/...` — all tests pass with updated file counts. Run `uf init` in a temp dir with pre-existing `reviewer-*.md` files — warning is printed.

### Implementation

- [x] T001 [P] [US1] Delete scaffold asset `internal/scaffold/assets/opencode/agents/reviewer-adversary.md`
- [x] T002 [P] [US1] Delete scaffold asset `internal/scaffold/assets/opencode/agents/reviewer-architect.md`
- [x] T003 [P] [US1] Delete scaffold asset `internal/scaffold/assets/opencode/agents/reviewer-guard.md`
- [x] T004 [P] [US1] Delete scaffold asset `internal/scaffold/assets/opencode/agents/reviewer-sre.md`
- [x] T005 [US1] Remove 4 `reviewer-*` entries from `expectedAssetPaths` in `internal/scaffold/scaffold_test.go` (lines 127-130: `reviewer-adversary.md`, `reviewer-architect.md`, `reviewer-guard.md`, `reviewer-sre.md`) and update the comment on line 118 from "legacy reviewers (4)" to remove the legacy reviewer count
- [x] T006 [US1] Remove `reviewer-testing.md` entry from `knownNonEmbeddedFiles` in `internal/scaffold/scaffold_test.go` (line 876) — no longer relevant since the canonical source is no longer expected
- [x] T007 [US1] Update `isDivisorAsset` test cases in `internal/scaffold/scaffold_test.go` (lines 1006-1007): remove the two `reviewer-guard.md` and `reviewer-architect.md` test entries that assert `false`, since those assets no longer exist
- [x] T008 [US1] Add legacy file detection to `scaffold.Run()` in `internal/scaffold/scaffold.go`: after the main file-writing loop and before `printSummary()`, use `filepath.Glob` to check for `reviewer-*.md` files in the target's `.opencode/agents/` directory. If found, print a warning listing each file and suggest removal command `rm .opencode/agents/reviewer-*.md`. Do NOT delete the files (per Spec 019 FR-003a)
- [x] T009 [US1] Add test `TestRun_LegacyFileWarning` in `internal/scaffold/scaffold_test.go`: create `reviewer-*.md` files in `t.TempDir()/.opencode/agents/`, run scaffold, verify warning output contains file names and removal command
- [x] T010 [US1] Verify existing `TestScaffoldOutput_NoReviewerFiles` regression test (line 1124) still passes — this test already asserts no `reviewer-` files in scaffold output

**Checkpoint**: `go test -race -count=1 ./internal/scaffold/...` passes. Asset count is 48 (52 - 4 reviewer files). Legacy file warning works.

---

## Phase 2: Severity Convention Pack (US3 — Severity, P2)

**Goal**: Create the shared severity definitions convention pack and integrate it into the scaffold engine.

**Independent Test**: Run `uf init` — `severity.md` is deployed to `.opencode/unbound/packs/`. All scaffold tests pass with the new asset.

### Implementation

- [x] T011 [US3] Create severity convention pack at `.opencode/unbound/packs/severity.md` with CRITICAL/HIGH/MEDIUM/LOW definitions, boundary criteria, and per-persona domain-specific example tables (content from `data-model.md` Severity Level Definitions section). Include auto-fix policy reference (LOW/MEDIUM = auto-fix, HIGH/CRITICAL = report only). Mark as tool-owned with `<!-- scaffolded by uf vdev -->` marker
- [x] T012 [US3] Copy severity pack to scaffold assets at `internal/scaffold/assets/opencode/unbound/packs/severity.md`
- [x] T013 [US3] Add `"opencode/unbound/packs/severity.md"` to `expectedAssetPaths` in `internal/scaffold/scaffold_test.go` (in the convention packs section, after `typescript.md`)
- [x] T014 [US3] Verify `isToolOwned` in `internal/scaffold/scaffold.go` returns `true` for `severity.md` — the existing pattern matching on `packs/` prefix should already cover this; if not, add explicit match
- [x] T015 [US3] Verify `isDivisorAsset` in `internal/scaffold/scaffold.go` returns `true` for `severity.md` — the existing pattern matching on `packs/` prefix should already cover this; if not, add explicit match
- [x] T016 [US3] Verify `shouldDeployPack` in `internal/scaffold/scaffold.go` returns `true` for `severity.md` regardless of language — severity is language-agnostic like `default.md`; add explicit handling if needed
- [x] T017 [US3] Add `isDivisorAsset` test case for `severity.md` → `true` in `internal/scaffold/scaffold_test.go`
- [x] T018 [US3] Add `shouldDeployPack` test case for `severity.md` → `true` (all languages) in `internal/scaffold/scaffold_test.go`

**Checkpoint**: `go test -race -count=1 ./internal/scaffold/...` passes. Asset count is 49 (48 + 1 severity pack). `severity.md` deploys with `uf init` and `uf init --divisor`.

---

## Phase 3: Agent De-duplication (US2 — De-duplication, P1) 🎯 MVP

**Goal**: Rewrite all 5 Divisor agent files with exclusive ownership boundaries, out-of-scope sections, and no cross-persona duplication.

**Independent Test**: `grep -l "Out of Scope" .opencode/agents/divisor-*.md` returns 5 files. `grep -l "hardcoded secrets" .opencode/agents/divisor-*.md` returns only `divisor-adversary.md`.

### Implementation

- [x] T019 [P] [US2] Rewrite `divisor-adversary.md` at `.opencode/agents/divisor-adversary.md`: audit checklist covers only Security & Resilience domain (secrets/credentials, dependency CVEs/supply chain, error handling/resilience, path/injection safety). Remove efficiency/performance checks (→ SRE), file permissions (→ SRE), test isolation (→ Tester), zero-waste (→ Guard). Add "Out of Scope" section per data-model.md
- [x] T020 [P] [US2] Rewrite `divisor-architect.md` at `.opencode/agents/divisor-architect.md`: audit checklist covers only Structure & Conventions domain (architectural alignment, key pattern adherence, coding/testing/documentation convention compliance [PACK], DRY/structural integrity). Remove plan alignment (→ Guard). Add "Out of Scope" section per data-model.md
- [x] T021 [P] [US2] Rewrite `divisor-guard.md` at `.opencode/agents/divisor-guard.md`: audit checklist covers only Intent & Governance domain (plan alignment/intent drift, zero-waste mandate, constitution alignment, cross-component value). Add "Out of Scope" section per data-model.md
- [x] T022 [P] [US2] Rewrite `divisor-sre.md` at `.opencode/agents/divisor-sre.md`: audit checklist covers only Operations & Efficiency domain (file permissions/hardcoded config, efficiency/performance, release pipeline integrity, dependency health, runtime observability, upgrade/migration paths, operational documentation, backup/recovery). Add efficiency checks moved from Adversary. Add "Out of Scope" section per data-model.md
- [x] T023 [P] [US2] Rewrite `divisor-testing.md` at `.opencode/agents/divisor-testing.md`: audit checklist covers only Test Quality & Coverage domain (test architecture, coverage strategy, assertion depth, test isolation, regression protection). Add "Out of Scope" section per data-model.md
- [x] T024 [US2] Copy all 5 updated `divisor-*.md` agent files to scaffold assets at `internal/scaffold/assets/opencode/agents/divisor-*.md` (overwrite existing)
- [x] T025 [US2] Verify drift detection test `TestEmbeddedAssets_MatchSource` passes — canonical sources must match embedded assets

**Checkpoint**: `go test -race -count=1 ./internal/scaffold/...` passes. Each review dimension is owned by exactly one persona. No cross-persona duplication.

---

## Phase 4: Severity References in Agents (US3 — Severity, P2)

**Goal**: All 5 Divisor agents reference the shared severity convention pack instead of inline severity definitions.

**Independent Test**: `grep -l "severity.md" .opencode/agents/divisor-*.md` returns 5 files.

### Implementation

- [x] T026 [P] [US3] Update `divisor-adversary.md`: replace inline severity definitions with reference to `.opencode/unbound/packs/severity.md`. Add instruction to load severity pack at review start
- [x] T027 [P] [US3] Update `divisor-architect.md`: replace inline severity definitions (including 1-10 alignment score mapping) with reference to `.opencode/unbound/packs/severity.md`
- [x] T028 [P] [US3] Update `divisor-guard.md`: replace inline severity definitions with reference to `.opencode/unbound/packs/severity.md`
- [x] T029 [P] [US3] Update `divisor-sre.md`: replace inline severity definitions with reference to `.opencode/unbound/packs/severity.md`
- [x] T030 [P] [US3] Update `divisor-testing.md`: replace inline severity definitions with reference to `.opencode/unbound/packs/severity.md`
- [x] T031 [US3] Copy all 5 updated agent files to scaffold assets (overwrite) and verify drift detection passes

**Checkpoint**: All 5 agents reference the shared severity standard. Drift detection passes.

---

## Phase 5: Qualified FR References (US4 — Qualified FRs, P2)

**Goal**: All functional requirement references in Divisor agent files use the fully qualified "per Spec NNN FR-XXX" format.

**Independent Test**: `grep -n "FR-[0-9]" .opencode/agents/divisor-*.md | grep -v "per Spec [0-9]"` returns no output.

### Implementation

- [x] T032 [P] [US4] Update `divisor-adversary.md`: qualify all bare "FR-" references with spec numbers (e.g., "FR-020" → "per Spec 005 FR-020")
- [x] T033 [P] [US4] Update `divisor-architect.md`: qualify all bare "FR-" references with spec numbers
- [x] T034 [P] [US4] Update `divisor-guard.md`: qualify all bare "FR-" references with spec numbers
- [x] T035 [P] [US4] Update `divisor-sre.md`: qualify all bare "FR-" references with spec numbers
- [x] T036 [P] [US4] Update `divisor-testing.md`: qualify all bare "FR-" references with spec numbers
- [x] T037 [US4] Copy all 5 updated agent files to scaffold assets (overwrite) and verify drift detection passes
- [x] T038 [US4] Add regression test `TestDivisorAgents_NoBareFRReferences` in `internal/scaffold/scaffold_test.go`: scan all `divisor-*.md` assets for bare "FR-NNN" patterns without "per Spec" qualifier

**Checkpoint**: Zero bare FR references in any Divisor agent file. Regression test passes.

---

## Phase 6: Learning Loop Integration (US5 — Learning Loop, P3)

**Goal**: All 5 Divisor agents include a Prior Learnings step that queries Hivemind before beginning review, with graceful degradation.

**Independent Test**: `grep -l "Prior Learnings" .opencode/agents/divisor-*.md` returns 5 files. `grep -l "hivemind_find" .opencode/agents/divisor-*.md` returns 5 files.

### Implementation

- [x] T039 [P] [US5] Update `divisor-adversary.md`: add "Step 0: Prior Learnings" section before the Source Documents step. Query `hivemind_find` with file paths from the diff. Include graceful degradation when Hivemind is unavailable (skip with informational note)
- [x] T040 [P] [US5] Update `divisor-architect.md`: add "Step 0: Prior Learnings" section with same pattern
- [x] T041 [P] [US5] Update `divisor-guard.md`: add "Step 0: Prior Learnings" section with same pattern
- [x] T042 [P] [US5] Update `divisor-sre.md`: add "Step 0: Prior Learnings" section with same pattern
- [x] T043 [P] [US5] Update `divisor-testing.md`: add "Step 0: Prior Learnings" section with same pattern
- [x] T044 [US5] Copy all 5 updated agent files to scaffold assets (overwrite) and verify drift detection passes

**Checkpoint**: All 5 agents have Prior Learnings step. Graceful degradation documented. Drift detection passes.

---

## Phase 7: Review Council & Severity Pack Update (US3 + US6, P2/P3)

**Goal**: Update `/review-council` command to reference severity pack and ensure Phase 1a picks up static analysis tools from CI workflow.

**Independent Test**: `grep "severity" .opencode/command/review-council.md` returns matches. Phase 1a derives `golangci-lint` and `govulncheck` from `.github/workflows/test.yml`.

### Implementation

- [x] T045 [US3] Update `review-council.md` at `.opencode/command/review-council.md`: add reference to shared severity convention pack in the council output format section. Ensure the auto-fix policy (LOW/MEDIUM auto-fix, HIGH/CRITICAL report) references the shared severity definitions
- [x] T046 [US3] Copy updated `review-council.md` to scaffold asset at `internal/scaffold/assets/opencode/command/review-council.md` and verify drift detection passes

**Checkpoint**: Review council references shared severity standard. Phase 1a will naturally pick up golangci-lint/govulncheck once they appear in CI workflow (Phase 8).

---

## Phase 8: Static Analysis in CI (US6 — Static Analysis, P3)

**Goal**: Add `golangci-lint` and `govulncheck` to the CI workflow and `uf setup`.

**Independent Test**: `grep "golangci-lint" .github/workflows/test.yml` returns matches. `grep "govulncheck" .github/workflows/test.yml` returns matches. `grep "golangci-lint" internal/setup/setup.go` returns matches.

### Implementation

- [x] T047 [US6] Update `.github/workflows/test.yml`: add `golangci-lint` installation step (via `golangci/golangci-lint-action` or `go install`) and `golangci-lint run` step after `go vet` and before `go test`
- [x] T048 [US6] Update `.github/workflows/test.yml`: add `govulncheck` installation step (`go install golang.org/x/vuln/cmd/govulncheck@latest`) and `govulncheck ./...` step after `go test`
- [x] T049 [US6] Add `installGolangciLint` function in `internal/setup/setup.go` following the `installGaze` pattern: check if `golangci-lint` is in PATH, if not install via `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` with Homebrew fallback
- [x] T050 [US6] Add `installGovulncheck` function in `internal/setup/setup.go` following the `installGaze` pattern: check if `govulncheck` is in PATH, if not install via `go install golang.org/x/vuln/cmd/govulncheck@latest`
- [x] T051 [US6] Wire `installGolangciLint` and `installGovulncheck` into the `Run()` function in `internal/setup/setup.go` — add to the step sequence after existing tool installations. Update step count constant if applicable
- [x] T052 [US6] Fix any existing `golangci-lint` findings in the codebase that would cause CI to fail with the new gate (expected per spec assumptions — run `golangci-lint run` locally and fix all reported issues)

**Checkpoint**: CI workflow includes both tools. `uf setup` installs both tools. `golangci-lint run` exits 0 locally. `govulncheck ./...` exits 0 locally.

---

## Phase 9: Documentation & Polish

**Purpose**: Update living documentation, AGENTS.md, and run final validation.

- [x] T053 [P] Update `AGENTS.md`: add Spec 019 to Recent Changes section with summary of all 6 user stories. Update file count references (52 → 49). Update Active Technologies if needed
- [x] T054 [P] Update `specs/019-divisor-council-refinement/spec.md`: change status from `draft` to `complete`
- [x] T055 Run full test suite: `make check` (build + vet + lint + test). All tests must pass
- [x] T056 Run quickstart.md verification steps from `specs/019-divisor-council-refinement/quickstart.md` — all 6 user story verification checks must pass

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Scaffold Cleanup)**: No dependencies — can start immediately. BLOCKS Phase 3 (agents reference updated asset set)
- **Phase 2 (Severity Pack)**: No dependencies — can start immediately. Can run in PARALLEL with Phase 1
- **Phase 3 (De-duplication)**: Depends on Phase 1 completion (asset paths must be correct). Can run in PARALLEL with Phase 2
- **Phase 4 (Severity Refs)**: Depends on Phase 2 (severity pack must exist) AND Phase 3 (agents must be rewritten first)
- **Phase 5 (Qualified FRs)**: Depends on Phase 3 (agents must be rewritten first). Can run in PARALLEL with Phase 4
- **Phase 6 (Learning Loop)**: Depends on Phase 3 (agents must be rewritten first). Can run in PARALLEL with Phases 4 and 5
- **Phase 7 (Review Council)**: Depends on Phase 2 (severity pack must exist)
- **Phase 8 (Static Analysis)**: No dependencies on other phases — can start any time. Independent CI/setup changes
- **Phase 9 (Documentation)**: Depends on ALL previous phases

### Parallel Opportunities

```text
Time →

[Phase 1: Scaffold Cleanup]  ──┐
[Phase 2: Severity Pack]    ──┤──→ [Phase 3: De-dup] ──┬──→ [Phase 4: Severity Refs] ──┐
[Phase 8: Static Analysis]  ──┤                        ├──→ [Phase 5: Qualified FRs]  ──┤
                              │                        └──→ [Phase 6: Learning Loop]  ──┤
                              └──→ [Phase 7: Review Council] ──────────────────────────┤
                                                                                       └──→ [Phase 9: Docs]
```

- **T001-T004** (asset deletions) can all run in parallel
- **T019-T023** (agent rewrites) can all run in parallel (different files)
- **T026-T030** (severity refs) can all run in parallel (different files)
- **T032-T036** (FR qualification) can all run in parallel (different files)
- **T039-T043** (learning loop) can all run in parallel (different files)
- **T047-T048** (CI workflow) are sequential (same file)
- **T049-T050** (setup functions) can run in parallel (different functions, same file — but sequential is safer)

### Within Each Agent File

Phases 3, 4, 5, and 6 all modify the same 5 agent files. They MUST be executed in order:
1. Phase 3: Rewrite with ownership boundaries (structural change)
2. Phase 4: Add severity references (content addition)
3. Phase 5: Qualify FR references (content modification)
4. Phase 6: Add Prior Learnings step (content addition)

After each phase, copy updated agents to scaffold assets (T024/T031/T037/T044).

---

## Implementation Strategy

### MVP First (US1 + US2 — P1 Stories)

1. Complete Phase 1: Scaffold Cleanup (US1)
2. Complete Phase 3: Agent De-duplication (US2)
3. **STOP and VALIDATE**: Run `make check`. All tests pass. No duplicate review dimensions.
4. This delivers the two highest-value changes: zero-waste cleanup and de-duplication.

### Incremental Delivery

1. Phase 1 + Phase 2 (parallel) → Scaffold clean, severity pack ready
2. Phase 3 → Agents rewritten with ownership boundaries (MVP!)
3. Phase 4 + Phase 5 + Phase 6 (sequential per-agent, parallel across agents) → Severity refs, qualified FRs, learning loop
4. Phase 7 + Phase 8 (parallel) → Review council + CI updates
5. Phase 9 → Documentation and final validation

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Agent file modifications in Phases 3-6 are cumulative — each phase builds on the previous
- The scaffold asset copy tasks (T024, T031, T037, T044, T046) are synchronization points — they ensure drift detection passes after each batch of changes
- File count: 52 (current) → 49 (after removing 4 reviewer assets, adding 1 severity pack, net -3). SC-003 updated to match. The `reviewer-testing.md` was never in `expectedAssetPaths` (it was in `knownNonEmbeddedFiles`), so removing it from that list doesn't affect the count
- `golangci-lint` may produce findings in the existing codebase — T052 handles this explicitly
- File count note updated: SC-003 corrected from "52 → 47" to "52 → 49" during spec review

<!-- spec-review: passed -->
