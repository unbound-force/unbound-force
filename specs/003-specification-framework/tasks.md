# Tasks: Specification Framework

**Input**: Design documents from `/specs/003-specification-framework/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Initialize Go project and create the directory
structure for the `unbound` CLI binary

- [ ] T001 Initialize Go module with `go mod init github.com/unbound-force/unbound-force` in go.mod
- [ ] T002 Add Cobra dependency with `go get github.com/spf13/cobra@latest`
- [ ] T003 Create CLI entry point with root and version commands in cmd/unbound/main.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core scaffold package that ALL user stories
depend on. This implements the embed.FS scaffold engine
following the Gaze pattern.

**CRITICAL**: No user story work can begin until this
phase is complete.

- [ ] T004 Create scaffold Options and Result structs in internal/scaffold/scaffold.go per data-model.md entities 1-2
- [ ] T005 Implement `isToolOwned()` function in internal/scaffold/scaffold.go per contracts/installer-cli.md ownership rules
- [ ] T006 Implement `insertMarkerAfterFrontmatter()` function in internal/scaffold/scaffold.go for version marker injection (`<!-- scaffolded by unbound vX.Y.Z -->`)
- [ ] T007 Implement `Run()` function in internal/scaffold/scaffold.go: walk embedded assets, apply ownership rules, write files, return Result
- [ ] T008 Implement `printSummary()` function in internal/scaffold/scaffold.go to report created/skipped/updated/overwritten files
- [ ] T009 Wire `init` subcommand in cmd/unbound/main.go to call scaffold.Run() with Options from flags (--force)
- [ ] T010 Create empty embedded assets directory structure at internal/scaffold/assets/ with subdirectories: specify/templates/, specify/scripts/bash/, opencode/command/, opencode/agents/, openspec/schemas/unbound-force/templates/, openspec/

**Checkpoint**: Scaffold engine complete. `unbound init`
runs but has no assets to extract yet.

---

## Phase 3: User Story 1 - Single Source of Truth (Priority: P1)

**Goal**: Establish canonical copies of all framework files
in the embedded assets directory and verify the canonical
source files exist at the repo root.

**Independent Test**: Run `ls internal/scaffold/assets/` and
verify it contains copies of all 22 Speckit files plus the
OpenSpec schema files. Run `go test ./internal/scaffold/...`
and verify the drift detection test passes.

- [ ] T011 [P] [US1] Copy 6 Speckit templates from .specify/templates/ to internal/scaffold/assets/specify/templates/
- [ ] T012 [P] [US1] Copy 5 Speckit scripts from .specify/scripts/bash/ to internal/scaffold/assets/specify/scripts/bash/
- [ ] T013 [P] [US1] Copy 10 OpenCode commands from .opencode/command/ to internal/scaffold/assets/opencode/command/
- [ ] T014 [P] [US1] Copy 1 agent file from .opencode/agents/constitution-check.md to internal/scaffold/assets/opencode/agents/constitution-check.md
- [ ] T015 [US1] Add `//go:embed assets` directive in internal/scaffold/scaffold.go to embed the assets filesystem
- [ ] T016 [US1] Implement TestEmbeddedAssetsMatchSource drift detection test in internal/scaffold/scaffold_test.go: walk assets/, compare each file byte-for-byte against canonical source at repo root
- [ ] T017 [US1] Run `go test ./internal/scaffold/...` and verify drift detection test passes

**Checkpoint**: This repo contains exactly one authoritative
version of each framework file. Embedded copies are
byte-identical to canonical source files (SC-001, SC-004).

---

## Phase 4: User Story 2 - Distribution and Installation (Priority: P1)

**Goal**: `unbound init` scaffolds all framework files into
a fresh repository in under 5 seconds with correct file
ownership behavior.

**Independent Test**: Run `unbound init` in a fresh temp
directory. Verify all expected directories and files are
created. Modify a user-owned file, re-run `unbound init`,
verify it is skipped. Modify a tool-owned file's embedded
source, re-run, verify it is updated.

- [ ] T018 [US2] Implement scaffold test for fresh repo: create temp dir, run scaffold.Run(), verify all files created in internal/scaffold/scaffold_test.go
- [ ] T019 [US2] Implement scaffold test for re-run: run scaffold.Run() twice, verify user-owned files skipped, tool-owned files unchanged in internal/scaffold/scaffold_test.go
- [ ] T020 [US2] Implement scaffold test for --force flag: run scaffold.Run() with Force=true, verify all files overwritten in internal/scaffold/scaffold_test.go
- [ ] T021 [US2] Implement scaffold test for version marker: verify scaffolded files contain `<!-- scaffolded by unbound ... -->` after YAML frontmatter in internal/scaffold/scaffold_test.go
- [ ] T022 [US2] Implement scaffold test for tool-owned update: modify a tool-owned file on disk, re-run scaffold.Run(), verify it is overwritten in internal/scaffold/scaffold_test.go
- [ ] T023 [US2] Run `go test ./...` and verify all scaffold tests pass
- [ ] T024 [US2] Build binary with `go build -o unbound ./cmd/unbound` and run `./unbound init` in a temp directory to verify end-to-end

**Checkpoint**: `unbound init` works in a fresh repo (SC-002),
re-run skips user-owned files (SC-003), and version markers
are present in all scaffolded files.

---

## Phase 5: User Story 5 - Tactical Change Workflow (Priority: P1)

**Goal**: OpenSpec directory structure and custom schema are
scaffolded by `unbound init`, enabling tactical workflows.

**Independent Test**: Run `unbound init` in a fresh repo,
verify `openspec/` directory exists with schemas, config,
specs, and changes subdirectories. Install OpenSpec CLI and
verify `openspec schemas` lists `unbound-force`.

- [ ] T025 [P] [US5] Create OpenSpec schema.yaml file at openspec/schemas/unbound-force/schema.yaml per contracts/openspec-schema.md
- [ ] T026 [P] [US5] Create OpenSpec proposal template at openspec/schemas/unbound-force/templates/proposal.md per contracts/openspec-schema.md (including Constitution Alignment section)
- [ ] T027 [P] [US5] Create OpenSpec spec template at openspec/schemas/unbound-force/templates/spec.md per contracts/openspec-schema.md (ADDED/MODIFIED/REMOVED with RFC 2119)
- [ ] T028 [P] [US5] Create OpenSpec design template at openspec/schemas/unbound-force/templates/design.md per contracts/openspec-schema.md
- [ ] T029 [P] [US5] Create OpenSpec tasks template at openspec/schemas/unbound-force/templates/tasks.md per contracts/openspec-schema.md
- [ ] T030 [US5] Create default OpenSpec config at openspec/config.yaml per contracts/openspec-schema.md (with constitution context and per-artifact rules)
- [ ] T031 [US5] Copy all OpenSpec files (T025-T030) to internal/scaffold/assets/openspec/ for embedding
- [ ] T032 [US5] Run `go test ./internal/scaffold/...` to verify drift detection passes for new OpenSpec files
- [ ] T033 [US5] Build binary and run `unbound init` in temp dir, verify openspec/ directory structure is created with schema, templates, config, specs/, and changes/ subdirectories

**Checkpoint**: OpenSpec tactical workflow is scaffolded.
`openspec/schemas/unbound-force/` exists with all templates.

---

## Phase 6: User Story 6 - Constitution Governance Bridge (Priority: P1)

**Goal**: Every OpenSpec proposal created with the
`unbound-force` schema includes a constitution alignment
section with assessments for all three principles.

**Independent Test**: Inspect the scaffolded
`openspec/schemas/unbound-force/templates/proposal.md` and
verify it contains Constitution Alignment section headings
for all three principles. Inspect `openspec/config.yaml`
and verify the context field references the constitution.

- [ ] T034 [US6] Verify proposal.md template (from T026) contains "Constitution Alignment" section with I. Autonomous Collaboration, II. Composability First, III. Observable Quality subsections and PASS/N/A assessment fields
- [ ] T035 [US6] Verify openspec/config.yaml (from T030) context field references `.specify/memory/constitution.md` and includes three principle summaries
- [ ] T036 [US6] Verify openspec/config.yaml (from T030) rules.proposal includes "MUST include Constitution Alignment section"
- [ ] T037 [US6] Verify schema.yaml (from T025) proposal artifact instruction includes "MUST include a Constitution Alignment section" and "Read the constitution from .specify/memory/constitution.md"

**Checkpoint**: Constitution governance bridge is structurally
enforced via template, context injection, and per-artifact
rules (SC-008).

---

## Phase 7: User Story 7 - Strategic/Tactical Boundary Guidelines (Priority: P1)

**Goal**: Clear, documented boundary guidelines enable
contributors to determine whether to use Speckit or
OpenSpec for any given piece of work.

**Independent Test**: Read the boundary guidelines
documentation and classify 10 representative work items.
At least 8 of 10 should produce a clear recommendation.

- [ ] T038 [US7] Add boundary guidelines section to AGENTS.md with decision criteria matrix (story count, cross-repo impact, constitution changes, bug fixes, maintenance) per research.md R4
- [ ] T039 [US7] Add escalation path guidance to AGENTS.md: "when in doubt, start with OpenSpec and escalate to Speckit if scope grows beyond 3 stories or crosses repo boundaries"
- [ ] T040 [US7] Add boundary guidelines to quickstart.md Decision Guide table (already partially present, verify completeness) in specs/003-specification-framework/quickstart.md

**Checkpoint**: Boundary guidelines documented with decision
matrix, heuristic, and escalation path (SC-009).

---

## Phase 8: User Story 4 - Pipeline Documentation (Priority: P1)

**Goal**: A new contributor understands both specification
workflows through clear documentation covering all phases,
actions, inputs, outputs, and prerequisites.

**Independent Test**: A new contributor follows the
quickstart documentation to scaffold a repo, run the first
three Speckit phases, and create an OpenSpec proposal.

- [ ] T041 [P] [US4] Document all 9 Speckit pipeline phases in AGENTS.md with purpose, prerequisites, inputs, outputs, and command for each phase
- [ ] T042 [P] [US4] Document all 4 core OpenSpec actions (propose, explore, apply, archive) in AGENTS.md with purpose, prerequisites, inputs, outputs, and command for each action
- [ ] T043 [US4] Add pipeline overview section to AGENTS.md showing the dual-tier framework with Speckit and OpenSpec side by side, including which phases are mandatory vs optional
- [ ] T044 [US4] Verify quickstart.md (specs/003-specification-framework/quickstart.md) covers end-to-end workflow from installation through first spec and first proposal

**Checkpoint**: Pipeline documentation covers all 9 Speckit
phases and 4 OpenSpec actions (SC-006).

---

## Phase 9: User Story 3 - Project-Specific Extension Points (Priority: P2)

**Goal**: Hero repos can customize framework behavior via
`.specify/config.yaml` without modifying canonical files.

**Independent Test**: Create a `.specify/config.yaml` with
`language: go` in a test repo. Run a speckit command and
verify it uses the Go-specific patterns instead of defaults.

- [ ] T045 [US3] Create default .specify/config.yaml template with all fields (language, framework, build_command, test_command, integration_patterns, project_type) and sensible defaults
- [ ] T046 [US3] Copy .specify/config.yaml to internal/scaffold/assets/specify/config.yaml for embedding
- [ ] T047 [US3] Update speckit.specify.md command in .opencode/command/speckit.specify.md to read integration_patterns from .specify/config.yaml instead of hardcoding patterns
- [ ] T048 [US3] Update speckit.plan.md command in .opencode/command/speckit.plan.md to read project_type from .specify/config.yaml to determine contract type (API/CLI/library)
- [ ] T049 [US3] Copy updated speckit.specify.md and speckit.plan.md to internal/scaffold/assets/opencode/command/ and verify drift test passes
- [ ] T050 [US3] Run `go test ./internal/scaffold/...` to verify all drift detection tests pass after config and command updates

**Checkpoint**: Extension points functional. Config-driven
customization replaces hardcoded patterns (SC-005).

---

## Phase 10: User Story 8 - Schema and Template Distribution (Priority: P2)

**Goal**: OpenSpec schema is versioned and distributed via
`unbound init`, with tool-owned schema files auto-updated
on re-run.

**Independent Test**: Modify a schema template in the
canonical source, rebuild the binary, re-run `unbound init`
in a repo that already has the schema. Verify the schema
file is updated while `openspec/config.yaml` is preserved.

- [ ] T051 [US8] Verify isToolOwned() classifies all files under openspec/schemas/ as tool-owned in internal/scaffold/scaffold.go
- [ ] T052 [US8] Verify isToolOwned() classifies openspec/config.yaml as user-owned in internal/scaffold/scaffold.go
- [ ] T053 [US8] Add scaffold test: modify schema file on disk, re-run scaffold.Run(), verify schema is updated while config.yaml is preserved in internal/scaffold/scaffold_test.go
- [ ] T054 [US8] Run `go test ./internal/scaffold/...` to verify schema distribution test passes

**Checkpoint**: Schema distributed via unbound init with
correct ownership behavior (SC-003 for OpenSpec files).

---

## Phase 11: Polish & Cross-Cutting Concerns

**Purpose**: Release pipeline, documentation updates, and
final validation across all user stories.

- [ ] T055 [P] Create .goreleaser.yaml at repo root with cross-platform build config (darwin/amd64, darwin/arm64, linux/amd64, linux/arm64), CGO_ENABLED=0, ldflags for version/commit/date injection, Homebrew cask publishing to unbound-force/homebrew-tap
- [ ] T056 [P] Create .github/workflows/release.yml for GoReleaser release pipeline triggered by v* tags (matching Gaze's release.yml pattern)
- [ ] T057 [P] Update AGENTS.md Active Technologies section with Go, Cobra, embed, GoReleaser, OpenSpec
- [ ] T058 [P] Update AGENTS.md Recent Changes section with spec 003 implementation summary
- [ ] T059 Update AGENTS.md Project Structure tree to include cmd/unbound/, internal/scaffold/, openspec/, go.mod, .goreleaser.yaml
- [ ] T060 Update README.md with Specification Framework section covering installation (`brew install unbound-force/tap/unbound`), usage (`unbound init`), and both workflows
- [ ] T061 Run `go vet ./...` and fix any issues
- [ ] T062 Run `go test ./...` and verify all tests pass
- [ ] T063 Build final binary with `go build -o unbound ./cmd/unbound` and run end-to-end validation: `unbound init` in fresh repo, verify all files created, re-run to verify skip/update behavior, run with `--force` to verify overwrite
- [ ] T064 Run quickstart.md validation: follow steps 1-6 in specs/003-specification-framework/quickstart.md in a fresh test repo and verify each step produces expected output
- [ ] T065 [US2] Verify `unbound init` works in a non-Unbound Force repository: run in a fresh directory with no `.specify/memory/constitution.md` or hero manifest, verify all files are created without errors and no hero-specific assumptions cause failures (SC-007, FR-012)
- [ ] T066 [US7] Document directory boundary enforcement in boundary guidelines (AGENTS.md): OpenSpec changes MUST NOT modify files under `specs/`, Speckit specs MUST NOT be created under `openspec/`. Enforcement is by convention and code review, not automated gates in v1.0.0 (FR-021, FR-022, SC-010)
- [ ] T067 [US7] Add boundary violation examples to quickstart.md or AGENTS.md showing what happens when a developer attempts to create a Speckit spec under `openspec/` or an OpenSpec delta spec under `specs/` -- expected outcome is code review rejection per documented convention (SC-010)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 -- BLOCKS
  all user stories
- **US1 (Phase 3)**: Depends on Phase 2 -- BLOCKS US2
  (assets must exist before scaffold tests)
- **US2 (Phase 4)**: Depends on Phase 3 -- scaffold tests
  need assets
- **US5 (Phase 5)**: Depends on Phase 2 -- creates OpenSpec
  schema files (independent of US1/US2)
- **US6 (Phase 6)**: Depends on Phase 5 -- verifies schema
  content
- **US7 (Phase 7)**: Depends on Phase 2 -- documentation
  only, no code dependency
- **US4 (Phase 8)**: Depends on Phase 7 -- documents both
  pipelines including boundary guidelines
- **US3 (Phase 9)**: Depends on Phase 4 -- needs working
  scaffold to distribute config
- **US8 (Phase 10)**: Depends on Phase 5 + Phase 4 -- needs
  schema files + working scaffold
- **Polish (Phase 11)**: Depends on all user stories

### User Story Dependencies

```text
Phase 1 (Setup)
  +-> Phase 2 (Foundational)
       +-> Phase 3 (US1: Canonical Source)
       |    +-> Phase 4 (US2: Distribution)
       |         +-> Phase 9 (US3: Extension Points)
       |         +-> Phase 10 (US8: Schema Distribution)
       +-> Phase 5 (US5: Tactical Workflow)
       |    +-> Phase 6 (US6: Constitution Bridge)
       |    +-> Phase 10 (US8: Schema Distribution)
       +-> Phase 7 (US7: Boundary Guidelines)
            +-> Phase 8 (US4: Pipeline Documentation)

Phase 11 (Polish) -- depends on all above
```

### Parallel Opportunities

- T011, T012, T013, T014 can run in parallel (different
  asset directories)
- T025, T026, T027, T028, T029 can run in parallel
  (different schema files)
- T041, T042 can run in parallel (different doc sections)
- T055, T056, T057, T058 can run in parallel (different
  files)
- Phases 5+6+7 can start in parallel with Phase 3+4 (after
  Phase 2)

---

## Parallel Example: User Story 1

```bash
# Copy all asset categories in parallel:
Task: "Copy 6 Speckit templates to internal/scaffold/assets/"
Task: "Copy 5 Speckit scripts to internal/scaffold/assets/"
Task: "Copy 10 OpenCode commands to internal/scaffold/assets/"
Task: "Copy 1 agent file to internal/scaffold/assets/"
```

## Parallel Example: User Story 5

```bash
# Create all OpenSpec schema files in parallel:
Task: "Create schema.yaml"
Task: "Create proposal.md template"
Task: "Create spec.md template"
Task: "Create design.md template"
Task: "Create tasks.md template"
```

---

## Implementation Strategy

### MVP First (US1 + US2 Only)

1. Complete Phase 1: Setup (Go project)
2. Complete Phase 2: Foundational (scaffold engine)
3. Complete Phase 3: US1 (canonical assets + drift test)
4. Complete Phase 4: US2 (scaffold tests + end-to-end)
5. **STOP and VALIDATE**: `unbound init` works in fresh repo
6. Binary is usable for Speckit-only distribution

### Incremental Delivery

1. Setup + Foundational + US1 + US2 → Working `unbound init`
2. Add US5 + US6 → OpenSpec schema + constitution bridge
3. Add US7 + US4 → Documentation complete
4. Add US3 → Extension points for project customization
5. Add US8 → Schema auto-update on re-run verified
6. Polish → Release pipeline, final docs, end-to-end

### Parallel Team Strategy

With two developers after Phase 2:

- **Developer A**: US1 → US2 → US3 → US8 (binary + scaffold)
- **Developer B**: US5 → US6 → US7 → US4 (schema + docs)
- Both converge at Phase 11 (Polish)

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story
- Each user story is independently testable at its checkpoint
- Commit after each task or logical group
- Run `go test ./...` frequently to catch drift early
- The drift detection test is the primary safety net for
  canonical source consistency
