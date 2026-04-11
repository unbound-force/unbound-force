# Tasks: Documentation Curation

**Input**: Design documents from `/specs/026-documentation-curation/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/divisor-curator-contract.md, quickstart.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Verify Baseline)

**Purpose**: Confirm the project builds and tests pass before any changes

- [x] T001 Run `go test -race -count=1 ./...` and confirm all tests pass (baseline)
- [x] T002 Verify `go build ./...` succeeds
- [x] T003 Confirm current asset count: 54 files in `internal/scaffold/assets/`, 54 entries in `expectedAssetPaths` in `internal/scaffold/scaffold_test.go`, `"54 files processed"` assertion in `cmd/unbound-force/main_test.go`

**Checkpoint**: Baseline green. All changes from this point forward are additive.

---

## Phase 2: US1 + US2 + US3 — Create the Curator Agent (Priority: P1/P1/P2)

**Goal**: Create the new `divisor-curator.md` agent file that detects documentation gaps (US1), identifies blog opportunities (US2), and identifies tutorial opportunities (US3). This is the bulk of the work — a single ~250-line Markdown file.

**Independent Test**: Add `divisor-curator.md` to `.opencode/agents/`. Run the review council on a PR that adds a new CLI command without updating AGENTS.md. Verify the Curator flags the documentation gap and files a website issue.

### Implementation

- [x] T004 [US1] Create `.opencode/agents/divisor-curator.md` with YAML frontmatter: `description`, `mode: subagent`, `model: google-vertex-anthropic/claude-opus-4-6@default`, `temperature: 0.2`, `tools: { read: true, write: false, edit: false, bash: true, webfetch: false }` (per contract and DD-002, DD-003)
- [x] T005 [US1] Add `# Role: The Curator` section with exclusive domain declaration ("Documentation & Content Pipeline Triage") and mode announcement (Code Review / Spec Review)
- [x] T006 [US1] Add `## Bash Access Restriction` section documenting that bash is restricted to `gh issue create --repo unbound-force/website` and `gh issue list --repo unbound-force/website` only (per FR-010, DD-002)
- [x] T007 [US1] Add `## Step 0: Prior Learnings` section with Dewey integration (`dewey_semantic_search` queries for documentation patterns, content gap history) and 3-tier graceful degradation (per FR-014, established Divisor agent pattern from research R01)
- [x] T008 [US1] Add `## Source Documents` section listing numbered inputs: constitution, severity pack, content pack (optional), AGENTS.md, README.md, PR diff, existing website issues (per contract Inputs table)
- [x] T009 [US1] Add `## Code Review Mode` section with `### Review Scope` describing documentation gap detection, blog/tutorial opportunity identification, and cross-repo issue filing
- [x] T010 [US1] Add `#### 1. Documentation Gap Detection` audit checklist item: check whether AGENTS.md/README.md were updated when user-facing behavior changed, flag missing updates as MEDIUM (per FR-002, FR-004)
- [x] T011 [US1] Add `#### 2. Website Documentation Issue Check` audit checklist item: check whether a GitHub issue was filed in `unbound-force/website` with label `docs` for changes requiring website updates, file issue via `gh issue create` if missing, flag as HIGH (per FR-003, FR-005)
- [x] T012 [US1] Add user-facing change detection heuristic inline: user-facing paths (`cmd/`, `.opencode/agents/`, `.opencode/command/`, `internal/scaffold/`, `AGENTS.md`, `README.md`, `unbound-force.md`) vs internal paths (`internal/` excluding scaffold, `*_test.go`, `.github/`, `specs/`, `openspec/`) (per contract, research R08)
- [x] T013 [US1] Add `#### 3. Duplicate Issue Check` audit checklist item: before filing any issue, search existing open issues via `gh issue list --repo unbound-force/website --label <label> --search "<keyword>" --state open` (per FR-009, research R07)
- [x] T014 [US2] Add `#### 4. Blog Opportunity Identification` audit checklist item: detect blog-worthy changes (new agents, new CLI commands, architectural migrations, new hero capabilities), file `blog`-labeled issue in website repo with suggested topic/angle/key points, flag missing blog issue as MEDIUM (per FR-006, FR-008, contract significance thresholds)
- [x] T015 [US3] Add `#### 5. Tutorial Opportunity Identification` audit checklist item: detect tutorial-worthy changes (new slash commands with multi-step workflows, new tool integrations, new workflow patterns), file `tutorial`-labeled issue in website repo with suggested structure/target audience, flag missing tutorial issue as MEDIUM (per FR-007, FR-008, contract significance thresholds)
- [x] T016 [US1] Add `## Spec Review Mode` section with minimal scope: documentation completeness in specs, content coverage assessment (per research R01 pattern)
- [x] T017 [US1] Add `## Output Format` section with standard Divisor finding template (`### [SEVERITY] Finding Title` with File, Constraint, Description, Recommendation fields) (per contract Outputs)
- [x] T018 [US1] Add `## Decision Criteria` section: APPROVE when all documentation is current and all required website issues exist; REQUEST CHANGES when any documentation gap (MEDIUM+) or missing content issue (MEDIUM+) is found (per contract Decision Criteria)
- [x] T019 [US1] Add `## Graceful Degradation` section: `gh` not available → report failure as finding with issue text for manual filing; website repo inaccessible → same; Dewey not available → skip Step 0; no content pack → skip content quality checks (per contract Graceful Degradation table)
- [x] T020 [US1] Add `## Out of Scope` section with explicit ownership boundaries: does not write documentation (Scribe's domain), does not write blog posts (Herald's domain), does not write PR communications (Envoy's domain), does not check code quality (other Divisor agents' domain) (per DD-001, established Divisor pattern)
- [x] T021 [US1] Add internal-only change exemption: refactoring, test-only, CI-only changes MUST NOT trigger documentation or content findings (per FR-013, spec edge cases)

**Checkpoint**: `divisor-curator.md` is complete (~250 lines). US1, US2, US3 acceptance scenarios are addressed in the agent file. The Curator is discoverable by the review council via `divisor-*.md` glob.

---

## Phase 3: US4 — Guard Documentation Completeness Enhancement (Priority: P2)

**Goal**: Add a "Documentation Completeness" checklist item to the Guard's Code Review audit, supplementing the Curator's cross-repo checks with a focused in-repo check.

**Independent Test**: Submit a PR that changes `uf setup` behavior without updating AGENTS.md. Verify the Guard flags the missing AGENTS.md update as a MEDIUM finding.

### Implementation

- [x] T022 [US4] Add `#### 6. Documentation Completeness` audit checklist item to `.opencode/agents/divisor-guard.md` after section 5 ("Gatekeeping Integrity") in the Code Review Mode section: verify AGENTS.md was updated when user-facing behavior changed, verify README.md was updated if project description changed, skip for internal-only changes, severity MEDIUM (per FR-011, FR-012, FR-013, DD-004, research R04)

**Checkpoint**: Guard has 6 Code Review audit items (was 5). US4 acceptance scenarios are addressed.

---

## Phase 4: Review Council Integration

**Goal**: Update the review council reference table to include the Curator's focus areas for targeted delegation context.

**Independent Test**: Read `review-council.md` and verify the Curator row exists in the Known Divisor Persona Roles table.

### Implementation

- [x] T023 [US1] Add Curator row to the Known Divisor Persona Roles reference table in `.opencode/command/review-council.md`: `| divisor-curator | The Curator | Documentation gaps, blog/tutorial opportunities, website issue filing | Documentation completeness in specs, content coverage |` (per DD-005 correction, research R02)

**Checkpoint**: Reference table has 6 review-agent rows (was 5). The Curator is documented for both Code Review and Spec Review focus areas.

---

## Phase 5: Scaffold Assets + Test Updates

**Purpose**: Synchronize scaffold asset copies and update test assertions to account for the new file.

### Scaffold Asset Sync

- [x] T024 [P] [US1] Copy `.opencode/agents/divisor-curator.md` to `internal/scaffold/assets/opencode/agents/divisor-curator.md` (new scaffold asset)
- [x] T025 [P] [US4] Copy `.opencode/agents/divisor-guard.md` to `internal/scaffold/assets/opencode/agents/divisor-guard.md` (sync modified Guard)
- [x] T026 [P] [US1] Copy `.opencode/command/review-council.md` to `internal/scaffold/assets/opencode/command/review-council.md` (sync modified review-council)

### Test Updates

- [x] T027 [US1] Add `"opencode/agents/divisor-curator.md"` entry to `expectedAssetPaths` in `internal/scaffold/scaffold_test.go` (after `divisor-architect.md`, before `divisor-envoy.md` — alphabetical order). Update comment from `// Divisor personas (5)` to `// Divisor personas (6)` (line ~118)
- [x] T028 [US1] Update `cmd/unbound-force/main_test.go` file count assertion from `"54 files processed"` to `"55 files processed"` (line ~34)

**Checkpoint**: Asset count is 55 (was 54). All scaffold assets are synchronized. Test assertions match new count.

---

## Phase 6: Documentation + Verification

**Purpose**: Update living documentation and verify everything passes.

### Documentation

- [x] T029 [P] Update `.opencode/agents/` listing in AGENTS.md Project Structure section to include `divisor-curator.md` with comment `# NEW: The Curator — documentation & content pipeline triage`
- [x] T030 [P] Update the agent count comment in AGENTS.md Project Structure section (if applicable — update `agents/` directory listing)
- [x] T031 [P] Add Recent Changes entry to AGENTS.md: summarize the Curator agent creation, Guard enhancement, review council update, scaffold sync, and test count updates. Include task count and user story count.
- [x] T032 [P] Update Active Technologies section in AGENTS.md to add `(026-documentation-curation)` annotation for the relevant technology entry: `Go 1.24+ (scaffold engine, tests only — no new Go logic) + Markdown (agent files), embed.FS (scaffold engine)`

### Verification

- [x] T033 Run `go build ./...` and confirm build succeeds
- [x] T034 Run `go test -race -count=1 ./...` and confirm all tests pass (key tests: `TestAssetPaths_MatchExpected`, `TestRun_CreatesFiles`, `TestScaffoldOutput_NoOldPathReferences`, `TestScaffoldOutput_NoGraphthulhuReferences`, `TestScaffoldOutput_NoHivemindReferences`)
- [x] T035 Run `golangci-lint run` and confirm no new lint findings
- [x] T036 Verify `divisor-curator.md` exists in both `.opencode/agents/` (live) and `internal/scaffold/assets/opencode/agents/` (asset)
- [x] T037 Verify Guard Code Review section has 6 audit items (was 5)
- [x] T038 Verify `review-council.md` reference table has 6 review-agent rows (was 5)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — verify baseline first
- **Phase 2 (Curator Agent)**: Depends on Phase 1 — the bulk of the work
- **Phase 3 (Guard Enhancement)**: Depends on Phase 1 — can run in parallel with Phase 2
- **Phase 4 (Review Council)**: Depends on Phase 1 — can run in parallel with Phase 2 and Phase 3
- **Phase 5 (Scaffold + Tests)**: Depends on Phases 2, 3, and 4 — copies final versions of all modified files
- **Phase 6 (Docs + Verification)**: Depends on Phase 5 — final documentation and full test run

### Parallel Opportunities

- **Phase 2 + Phase 3 + Phase 4**: Can run in parallel (different files: `divisor-curator.md`, `divisor-guard.md`, `review-council.md`)
- **T024 + T025 + T026**: Scaffold copies can run in parallel (different target files)
- **T029 + T030 + T031 + T032**: Documentation updates can run in parallel (all in AGENTS.md but different sections)

### Within Phase 2

- T004 through T021 are sequential within `divisor-curator.md` (same file, building up sections)
- T004 (frontmatter) must come first
- T005-T008 (role, bash restriction, Step 0, source docs) form the preamble
- T009-T015 (Code Review Mode with audit items) form the core
- T016-T021 (Spec Review, output, decisions, degradation, out of scope) form the closing

---

## Implementation Strategy

### Recommended Order (Single Developer)

1. Phase 1: Verify baseline (5 min)
2. Phase 2: Create Curator agent — T004 through T021 (main work, ~45 min)
3. Phase 3: Guard enhancement — T022 (5 min)
4. Phase 4: Review council update — T023 (5 min)
5. Phase 5: Scaffold sync + test updates — T024 through T028 (10 min)
6. Phase 6: Documentation + verification — T029 through T038 (15 min)

### File Inventory

| File | Action | Phase |
|------|--------|-------|
| `.opencode/agents/divisor-curator.md` | CREATE | 2 |
| `.opencode/agents/divisor-guard.md` | MODIFY | 3 |
| `.opencode/command/review-council.md` | MODIFY | 4 |
| `internal/scaffold/assets/opencode/agents/divisor-curator.md` | CREATE (copy) | 5 |
| `internal/scaffold/assets/opencode/agents/divisor-guard.md` | SYNC (copy) | 5 |
| `internal/scaffold/assets/opencode/command/review-council.md` | SYNC (copy) | 5 |
| `internal/scaffold/scaffold_test.go` | MODIFY | 5 |
| `cmd/unbound-force/main_test.go` | MODIFY | 5 |
| `AGENTS.md` | MODIFY | 6 |

### FR Traceability

| FR | Task(s) | Story |
|----|---------|-------|
| FR-001 | T004, T005 | US1 |
| FR-002 | T010 | US1 |
| FR-003 | T011 | US1 |
| FR-004 | T010 | US1 |
| FR-005 | T011 | US1 |
| FR-006 | T014 | US2 |
| FR-007 | T015 | US3 |
| FR-008 | T014, T015 | US2, US3 |
| FR-009 | T013 | US1 |
| FR-010 | T004, T006 | US1 |
| FR-011 | T022 | US4 |
| FR-012 | T022 | US4 |
| FR-013 | T021, T022 | US1, US4 |
| FR-014 | T007 | US1 |
| FR-015 | T024, T025, T026 | US1, US4 |
| FR-016 | T034 | All |

---

## Notes

- This is a Markdown-only change — no new Go logic, no new Go functions
- The scaffold engine's `isDivisorAsset()` already matches `divisor-*` — no Go code change needed
- The Curator is the first Divisor agent with `bash: true` — this is a documented exception (DD-002)
- Temperature 0.2 is a documented exception from the standard 0.1 (DD-003)
- Total task count: 38 tasks across 6 phases
- User stories covered: US1 (P1), US2 (P1), US3 (P2), US4 (P2)

<!-- spec-review: passed -->
