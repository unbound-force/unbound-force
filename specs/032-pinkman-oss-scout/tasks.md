# Tasks: Pinkman OSS Scout

**Input**: Design documents from `specs/032-pinkman-oss-scout/`
**Prerequisites**: plan.md (required), spec.md (required),
research.md, data-model.md, contracts/agent-interface.md

**Tests**: Scaffold drift detection tests (Go) are
required. Agent behavior tests are manual (invoke and
verify output format). No TDD requested.

**Organization**: Tasks are grouped by user story. All
four modes (discover, trend, audit, report) are
implemented within the single `pinkman.md` agent file
as instruction sections. The `/scout` command file
routes to the agent.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files)
- **[Story]**: Which user story (US1, US2, US3, US4)
- Exact file paths in all descriptions

---

## Phase 1: Setup

**Purpose**: Create the foundational file structure for
Pinkman agent and command.

- [x] T001 Create the Pinkman agent file at
  `.opencode/agents/pinkman.md` with YAML frontmatter
  (description: "OSS Scout — discovers open source
  projects, classifies licenses against the OSI-approved
  list, and tracks industry trends", mode: subagent,
  model: google-vertex-anthropic/claude-opus-4-6@default,
  temperature: 0.3, tools: read: true, write: true,
  edit: true, bash: false, webfetch: true). Include H1
  Role section with core identity (OSS Scout persona,
  Autonomous Collaboration compliance, non-hero utility
  agent status per Spec 031 pattern). Include H2 Core
  Constraint section stating FR-009 (must not replicate
  hero capabilities). Include H2 Source Documents
  section listing AGENTS.md, constitution, current spec.
  Per research.md R1, R6.

- [x] T002 Create the `/scout` slash command file at
  `.opencode/command/scout.md` with YAML frontmatter
  (description: "Invoke Pinkman OSS Scout to discover,
  trend-scan, audit, or report on open source projects").
  Include mode routing instructions per
  contracts/agent-interface.md: default mode is
  `discover`, `--trend` flag routes to trend mode,
  `--audit` flag routes to audit mode (default manifest:
  `go.mod`), `--report` flag routes to report mode.
  Command delegates to the `pinkman` agent with the
  parsed arguments. Per research.md R1.

**Checkpoint**: Two empty-shell Markdown files exist.
The `/scout` command can be invoked but the agent has
no operational logic yet.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core agent infrastructure that all user
story modes depend on -- OSI license retrieval, fallback
list, output formatting, error handling, Dewey
integration, and report persistence.

**CRITICAL**: No user story work can begin until this
phase is complete.

- [x] T003 Add H2 "OSI License Retrieval" section to
  `.opencode/agents/pinkman.md`: Instructions to fetch
  the current OSI-approved license list from
  `https://opensource.org/licenses/` via `webfetch` at
  every invocation. Parse the page to extract SPDX
  identifiers of all approved licenses. Per research.md
  R2, FR-003.

- [x] T004 Add H2 "Fallback License List" section to
  `.opencode/agents/pinkman.md`: Define a hardcoded
  fallback set of well-known OSI-approved licenses (MIT,
  Apache-2.0, BSD-2-Clause, BSD-3-Clause, ISC, MPL-2.0,
  LGPL-2.1, LGPL-3.0, GPL-2.0, GPL-3.0, AGPL-3.0,
  Unlicense, Artistic-2.0, EPL-2.0, EUPL-1.2, 0BSD,
  Zlib, BSL-1.0) for use when the OSI site is
  unreachable. Instruct the agent to note "using
  fallback license list, live OSI verification
  unavailable" in results when fallback is used. Per
  research.md R2, FR-012.

- [x] T005 Add H2 "License Classification" section to
  `.opencode/agents/pinkman.md`: Instructions to classify
  a project's license by: (1) detecting the license file
  in the repository via webfetch, (2) identifying the
  SPDX identifier, (3) checking against the retrieved
  OSI list (or fallback), (4) assigning a verdict enum
  (approved, not_approved, unknown, manual_review,
  dual_approved per data-model.md License Verdict). Handle
  edge cases: dual-license (FR-011 -- evaluate each,
  report most favorable), custom/non-standard (flag as
  manual_review), no license detected (flag as unknown).
  Per FR-002, FR-003, FR-011.

- [x] T006 Add H2 "Output Formatting" section to
  `.opencode/agents/pinkman.md`: Define the standard
  output format for all modes per
  contracts/agent-interface.md. Include the Scouted
  Project display template (name, URL, license, language,
  stars, releases, contributors, dependencies,
  description). Include the Shared Dependencies table
  template. Include the Incompatible Projects section
  template. Per FR-010, data-model.md.

- [x] T007 Add H2 "Error Handling" section to
  `.opencode/agents/pinkman.md`: Define error handling
  behavior per contracts/agent-interface.md Error
  Handling Contract table: OSI site unreachable (use
  fallback), GitHub rate-limited (report partial),
  no manifest found (skip dependency listing), unknown
  license (classify as unknown), custom license (classify
  as manual_review), no results (report "no projects
  found" with search criteria), Dewey unavailable (skip
  silently). Per FR-012.

- [x] T008 Add H2 "Report Persistence" section to
  `.opencode/agents/pinkman.md`: Instructions to save
  scouting reports as Markdown files with YAML
  frontmatter at `.uf/pinkman/reports/` using the naming
  convention `YYYY-MM-DDTHH-MM-SS-<sanitized-query>.md`.
  YAML frontmatter schema per data-model.md Scouting
  Report section (producer, version, timestamp, query,
  mode, result_count, etc.). Create directory if it does
  not exist. Per research.md R5,
  contracts/agent-interface.md Persistence Contract.

- [x] T009 Add H2 "Dewey Integration" section to
  `.opencode/agents/pinkman.md`: Instructions for
  optional Dewey integration per
  contracts/agent-interface.md Dewey Integration section.
  Before scouting, query `dewey_semantic_search` for
  past evaluations of the same domain or project URL.
  After scouting, store a condensed summary via
  `dewey_store_learning` with tag "pinkman", category
  "reference". Graceful degradation: if Dewey tools
  return errors, skip silently and proceed with local
  storage only. Per FR-013.

**Checkpoint**: Agent has all shared infrastructure.
OSI license list can be retrieved and classified.
Reports can be formatted and persisted. Error handling
is defined. Ready for user story mode implementations.

---

## Phase 3: User Story 1 — Discover License-Compatible OSS Projects (Priority: P1) MVP

**Goal**: Implement the `discover` mode: search for
open source projects by domain keyword, classify
licenses against the OSI list, list direct dependencies
per project, and detect shared dependencies across
results.

**Independent Test**: Invoke `/scout static analysis Go`
and verify: (1) all results have OSI-approved licenses,
(2) each result lists direct dependencies, (3) shared
dependencies are highlighted with counts, (4) non-approved
projects appear in a separate section.

- [x] T010 [US1] Add H2 "Discover Mode" section to
  `.opencode/agents/pinkman.md`: Instructions for the
  default discover mode. When invoked with a domain
  keyword: (1) use `webfetch` to search GitHub for
  repositories matching the keyword, (2) for each
  discovered project, fetch repository metadata (stars,
  forks, description, primary language, license file),
  (3) classify the license using the License
  Classification section, (4) separate results into
  compatible (OSI-approved) and incompatible lists. Per
  FR-001, FR-002, research.md R3.

- [x] T011 [US1] Add H2 "Dependency Listing" section to
  `.opencode/agents/pinkman.md`: Instructions to list
  direct dependencies for each scouted project. For each
  compatible project: (1) detect the dependency manifest
  file (go.mod, package.json, Cargo.toml,
  requirements.txt, pyproject.toml) via `webfetch`,
  (2) parse dependency names and versions per
  research.md R4, (3) include the dependency list in
  the project's output per data-model.md Dependency
  Reference entity. Handle missing manifests: report
  "dependencies unknown -- no manifest detected" and set
  has_manifest to false. Per FR-014, research.md R4.

- [x] T012 [US1] Add H2 "Dependency Overlap Detection"
  section to `.opencode/agents/pinkman.md`: Instructions
  for post-processing dependency overlap. After all
  projects are scouted: (1) collect all dependency lists,
  (2) identify dependencies that appear in 2+ projects,
  (3) for each shared dependency, report the dependency
  name, list of projects using it, version used by each,
  and whether versions conflict, (4) format as a Shared
  Dependencies table per contracts/agent-interface.md.
  Per FR-015, FR-016, data-model.md Dependency Overlap
  entity, plan.md D5.

- [x] T013 [US1] Add H2 "Incompatible Projects Section"
  to `.opencode/agents/pinkman.md`: Instructions to
  present non-OSI-approved projects in a separate
  "Incompatible Projects (for awareness)" section with
  the project name, URL, detected license, and an
  explanation of why the license is not OSI-approved.
  Projects with unknown licenses appear with "license
  unknown -- manual review required" note. Per spec.md
  acceptance scenarios 2 and 3.

**Checkpoint**: US1 is fully functional. `/scout <keyword>`
returns a curated list with license verdicts, dependency
lists, overlap detection, and incompatible flagging.

---

## Phase 4: User Story 2 — Track Industry Trends (Priority: P2)

**Goal**: Implement the `trend` mode: identify trending
projects with quantitative indicators, filter by
OSI-approved license, present non-approved trending
projects separately.

**Independent Test**: Invoke `/scout --trend MCP servers`
and verify: (1) results are ranked by trend strength,
(2) each project shows at least three quantitative
indicators, (3) non-OSI-approved trending projects
appear in a separate section.

- [x] T014 [US2] Add H2 "Trend Mode" section to
  `.opencode/agents/pinkman.md`: Instructions for trend
  scanning. When invoked with `--trend <category>`:
  (1) use `webfetch` to search GitHub for repositories
  in the category sorted by recent activity, (2) for
  each project, compute three primary trend indicators
  per research.md R7: star growth rate (stars gained in
  last 90 days as % of total), release velocity (releases
  in last 6 months), contributor activity (unique
  contributors in last 90 days), (3) rank projects by
  composite trend strength, (4) classify licenses and
  separate into "trending and compatible" vs. "trending
  but not OSI-approved" sections. Per FR-004,
  research.md R7.

- [x] T015 [US2] Add trend indicator display to the
  Output Formatting section in
  `.opencode/agents/pinkman.md`: Extend the Scouted
  Project display template to include trend indicators
  (star growth ↑X% in 90d, releases in 6mo, active
  contributors in 90d). Add secondary indicators when
  available (fork trajectory, issue response time). Per
  data-model.md Scouted Project entity fields
  (star_growth_rate, release_velocity,
  contributor_activity).

- [x] T016 [US2] Add "no trends detected" handling to
  Trend Mode section in `.opencode/agents/pinkman.md`:
  When no projects show significant trend signals in the
  requested category, report "no significant trends
  detected" with the date range consulted and sources
  checked. Per spec.md US2 acceptance scenario 3.

**Checkpoint**: US2 is functional. `/scout --trend`
returns projects ranked by trend strength with
quantitative indicators.

---

## Phase 5: User Story 3 — Monitor Existing Dependencies (Priority: P3)

**Goal**: Implement the `audit` mode: read a local
dependency manifest, check each dependency for updates,
license changes, and maintenance health.

**Independent Test**: Invoke `/scout --audit go.mod` and
verify: (1) each dependency shows current vs. latest
version, (2) license changes between versions are
detected, (3) maintenance risk levels are assigned with
specific indicators.

- [x] T017 [US3] Add H2 "Audit Mode" section to
  `.opencode/agents/pinkman.md`: Instructions for
  dependency auditing. When invoked with
  `--audit [manifest-path]`: (1) read the local manifest
  file using the `read` tool (default: `go.mod`),
  (2) parse dependency names and versions per
  research.md R4, (3) for each dependency, use `webfetch`
  to check the package registry for the latest available
  version and current license, (4) compare versions and
  licenses between current and latest. Per FR-005,
  FR-006, research.md R4.

- [x] T018 [US3] Add H2 "License Change Detection"
  section to `.opencode/agents/pinkman.md`: Instructions
  to detect license changes between dependency versions.
  For each dependency with an available update: (1) fetch
  the license of the currently used version, (2) fetch
  the license of the latest version, (3) if they differ,
  prominently warn about the change, (4) if the new
  license is not OSI-approved, recommend staying on the
  current version or finding an alternative. Per FR-006,
  data-model.md Dependency Health Report entity.

- [x] T019 [US3] Add H2 "Maintenance Risk Assessment"
  section to `.opencode/agents/pinkman.md`: Instructions
  to assess maintenance health for each dependency.
  Check for: last commit date (healthy: <6mo, warning:
  6-12mo, critical: >12mo), archived repository status,
  owner/organization changes, unresolved critical issues,
  growing issue backlog. Assign risk level (healthy,
  warning, critical) per data-model.md Maintenance Risk
  Enum. Report specific risk indicators per data-model.md
  Risk Indicators list. Per FR-007.

- [x] T020 [US3] Add audit output format to the Output
  Formatting section in `.opencode/agents/pinkman.md`:
  Define the audit result table format per
  contracts/agent-interface.md Audit Result Table
  (columns: Dependency, Current, Latest, Update?,
  License Changed?, Risk). Include a Risk Details
  subsection with explanations for warning and critical
  entries. Per contracts/agent-interface.md.

**Checkpoint**: US3 is functional. `/scout --audit`
reads a manifest and reports dependency health with
version, license, and maintenance analysis.

---

## Phase 6: User Story 4 — Generate Adoption Reports (Priority: P3)

**Goal**: Implement the `report` mode: generate a
comprehensive recommendation report for a specific
project with all required sections.

**Independent Test**: Invoke
`/scout --report https://github.com/example/project`
and verify: (1) report contains all sections (license,
health, trends, deps, overlap, recommendation),
(2) report is saved to `.uf/pinkman/reports/`,
(3) report includes YAML frontmatter with provenance.

- [x] T021 [US4] Add H2 "Report Mode" section to
  `.opencode/agents/pinkman.md`: Instructions for
  generating adoption recommendation reports. When
  invoked with `--report <project-url>`: (1) use
  `webfetch` to fetch comprehensive project metadata
  (license, stars, forks, contributors, releases,
  commit history, dependency manifest), (2) classify
  license using License Classification section,
  (3) compute trend indicators per Trend Mode,
  (4) assess maintenance health per Maintenance Risk
  Assessment, (5) list direct dependencies per
  Dependency Listing section, (6) check for overlap
  with previously evaluated projects (query Dewey if
  available), (7) check relationship to existing Unbound
  Force dependencies (read local go.mod), (8) assign
  overall recommendation verdict (adopt, evaluate,
  defer, avoid per data-model.md Recommendation Enum).
  Per FR-008, data-model.md Adoption Recommendation
  entity.

- [x] T022 [US4] Add report output format to the Output
  Formatting section in `.opencode/agents/pinkman.md`:
  Define the full recommendation report format per
  contracts/agent-interface.md Recommendation Report
  template. Include YAML frontmatter with provenance
  metadata per data-model.md Scouting Report schema.
  Sections: License Analysis, Community Health,
  Maintenance Signals, Trend Trajectory, Dependencies,
  Dependency Overlap, Relationship to Existing
  Dependencies, Recommendation verdict with reason.
  Per FR-008, FR-010.

- [x] T023 [US4] Add recommendation verdict logic to
  Report Mode section in `.opencode/agents/pinkman.md`:
  Define the decision criteria for each verdict: adopt
  (OSI-approved, healthy maintenance, positive trend,
  no conflicts), evaluate (OSI-approved but has concerns),
  defer (OSI-approved but significant concerns), avoid
  (not OSI-approved or critical risks). Per data-model.md
  Recommendation Enum.

**Checkpoint**: US4 is functional. `/scout --report`
generates a comprehensive adoption recommendation saved
to `.uf/pinkman/reports/`.

---

## Phase 7: Scaffold Integration & Tests

**Purpose**: Embed Pinkman files in the scaffold engine,
update tests, and update documentation.

- [x] T024 [P] Copy `.opencode/agents/pinkman.md` to
  `internal/scaffold/assets/opencode/agents/pinkman.md`
  as the embedded scaffold asset copy. Ensure exact
  byte-for-byte match with the canonical file. Per
  research.md R6.

- [x] T025 [P] Copy `.opencode/command/scout.md` to
  `internal/scaffold/assets/opencode/command/scout.md`
  as the embedded scaffold asset copy. Ensure exact
  byte-for-byte match with the canonical file. Per
  research.md R6.

- [x] T026 Update `expectedAssetPaths` in
  `internal/scaffold/scaffold_test.go`: add
  `"opencode/agents/pinkman.md"` to the agents section
  (alphabetically after `onboarding.md`) and
  `"opencode/command/scout.md"` to the commands section
  (alphabetically after `review-council.md`). Total
  count changes from 35 to 37. Per research.md R6,
  Spec 031 pattern.

- [x] T027 Update `cmd/unbound-force/main_test.go`:
  change the `"35 files processed"` assertion to
  `"37 files processed"` to match the new embedded
  asset count. Per Spec 031 pattern.

- [x] T028 Verify `isToolOwned` behavior in
  `internal/scaffold/scaffold.go`: confirm that
  `pinkman.md` returns false (user-owned) and
  `scout.md` returns true (tool-owned) based on
  existing classification logic. If the existing
  `isToolOwned` function does not correctly classify
  these files, update it. Per plan.md D6.

- [x] T029 Run `go build ./...` and
  `go test -race -count=1 ./...` to verify all scaffold
  drift detection tests pass, file count assertions are
  correct, and no compilation errors exist.

**Checkpoint**: Scaffold engine deploys both files during
`uf init`. All Go tests pass.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, AGENTS.md updates, and
validation.

- [x] T030 [P] Update AGENTS.md: add Pinkman to the
  "Utility Agents (Non-Hero)" table with columns: Agent
  = "Pinkman", Role = "OSS project scouting and license
  compatibility", File =
  `.opencode/agents/pinkman.md`, Status = "Implemented
  (Spec 032)". Add Spec 032 to the "Recent Changes"
  section with a summary of what was implemented. Update
  the Project Structure section to include `pinkman.md`
  in the agents listing and `scout.md` in the commands
  listing. Update the `expectedAssetPaths` count in the
  Active Technologies section if referenced.

- [x] T031 [P] Update `specs/032-pinkman-oss-scout/spec.md`:
  change **Status** from "Draft" to "Complete".

- [x] T032 Run the quickstart.md validation: invoke each
  command from `specs/032-pinkman-oss-scout/quickstart.md`
  (`/scout static analysis Go`, `/scout --trend MCP
  servers`, `/scout --audit`, `/scout --report <url>`)
  and verify each produces output matching the expected
  format from contracts/agent-interface.md.

- [x] T033 Assess documentation impact per AGENTS.md
  Documentation Validation Gate. Determine whether a
  GitHub issue is needed in `unbound-force/website` for
  the new `/scout` command and Pinkman agent. If user-
  facing behavior is added (it is -- new slash command),
  create the issue with `gh issue create --repo
  unbound-force/website`.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- start
  immediately
- **Foundational (Phase 2)**: Depends on Phase 1
  completion -- BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Phase 2 -- no
  dependencies on other stories
- **US2 (Phase 4)**: Depends on Phase 2 -- no
  dependencies on US1 (uses same license classification
  infrastructure)
- **US3 (Phase 5)**: Depends on Phase 2 -- no
  dependencies on US1 or US2
- **US4 (Phase 6)**: Depends on Phase 2 -- benefits from
  US1 (dependency listing), US2 (trend indicators), and
  US3 (maintenance risk) sections existing in the agent
  file, but can be implemented independently by
  duplicating relevant instructions
- **Scaffold (Phase 7)**: Depends on Phases 1-6 (all
  agent content must be finalized before copying)
- **Polish (Phase 8)**: Depends on Phase 7 (scaffold
  tests must pass before documentation)

### User Story Dependencies

- **US1 (P1)**: Independent -- MVP target
- **US2 (P2)**: Independent -- can run parallel with US1
- **US3 (P3)**: Independent -- can run parallel with
  US1/US2
- **US4 (P3)**: Soft dependency on US1/US2/US3 (reuses
  their instruction sections) but can be implemented
  standalone by including all needed logic in the Report
  Mode section

### Within Each User Story

All US tasks operate on the same file
(`.opencode/agents/pinkman.md`), so they MUST be
sequential within each story. However, stories add
non-overlapping sections to the file, so different
stories can be implemented sequentially without conflict.

### Parallel Opportunities

- T024 and T025 (scaffold copies) can run in parallel
- T030 and T031 (documentation updates) can run in
  parallel
- US1 through US4 implementation is sequential (same
  file) but each story's tasks are independent of other
  stories' tasks

---

## Parallel Example: Phase 7

```text
# Launch scaffold copies in parallel:
Task: "Copy pinkman.md to internal/scaffold/assets/"
Task: "Copy scout.md to internal/scaffold/assets/"

# Then sequential:
Task: "Update expectedAssetPaths in scaffold_test.go"
Task: "Update file count assertion in main_test.go"
Task: "Run tests"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: Foundational (T003-T009)
3. Complete Phase 3: US1 — Discover (T010-T013)
4. **STOP and VALIDATE**: Test `/scout <keyword>`
   independently
5. Complete Phase 7: Scaffold (T024-T029)
6. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational → Agent shell ready
2. Add US1 (Discover) → Test → Deploy (MVP!)
3. Add US2 (Trend) → Test → Deploy
4. Add US3 (Audit) → Test → Deploy
5. Add US4 (Report) → Test → Deploy
6. Scaffold + Polish → Final release

### Single-File Advantage

All four user stories add sections to a single agent
file (`pinkman.md`). This means:
- No merge conflicts between stories
- Each story adds non-overlapping H2 sections
- The file grows incrementally as stories are completed
- Scaffold copy (Phase 7) only needs to happen once,
  after all content is finalized

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story
- All US tasks modify `.opencode/agents/pinkman.md`
  (single file) so they are sequential within a story
- Total: 33 tasks across 8 phases
- Commit after each phase completion
- Stop at any checkpoint to validate independently
