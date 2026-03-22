# Tasks: Binary Rename

**Input**: Design documents from `/specs/013-binary-rename/`
**Prerequisites**: plan.md, spec.md, research.md, contracts/cli-schema.md

**Tests**: Tests are included -- the spec requires zero regressions (SC-005) and the constitution mandates testability (Principle IV). Existing tests must continue to pass after the directory rename.

**Organization**: Tasks are grouped by user story. US1 and US2 are both P1 and tightly coupled (the directory rename enables both), so they share the foundational phase. US3-US5 are independent.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Phase 1: Setup (Directory Rename)

**Purpose**: Rename the CLI source directory and update the Cobra root command. This is the structural change that enables everything else.

- [x] T001 Rename directory `cmd/unbound/` to `cmd/unbound-force/` (use `git mv cmd/unbound cmd/unbound-force`)
- [x] T002 Update Cobra root command `Use` field from `"unbound"` to `"unbound-force"` and add `(alias: uf)` to the Short description in `cmd/unbound-force/main.go`
- [x] T003 Update `cmd/unbound-force/main.go`: change `SetVersionTemplate` from `"unbound version {{.Version}}\n"` to `"unbound-force version {{.Version}}\n"`, change `newVersionCmd` output format from `"unbound v%s"` to `"unbound-force v%s"`, and update the version command `Short` description from `"Print the unbound version"` to `"Print the unbound-force version"`
- [x] T004 Run `go build ./...` to verify the rename compiles

---

## Phase 2: Foundational (Build and Test Infrastructure)

**Purpose**: Update the Makefile, GoReleaser config, and test assertions so the build and test pipeline works with the new name.

**CRITICAL**: Must complete before user story work begins.

- [x] T005 Add `install` target to `Makefile` that builds `$(GOPATH)/bin/unbound-force` from `./cmd/unbound-force/` and creates symlink `$(GOPATH)/bin/uf` pointing to it
- [x] T006 Update `.goreleaser.yaml`: change build `id` from `unbound` to `unbound-force`, change `main` from `./cmd/unbound` to `./cmd/unbound-force`, set `binary: unbound-force`, change `archives.name_template` from `unbound_{{ .Version }}...` to `unbound-force_{{ .Version }}...`, change `homebrew_casks.name` from `unbound` to `unbound-force`, update the quarantine removal hook from `"#{staged_path}/unbound"` to `"#{staged_path}/unbound-force"`
- [x] T006b Update `.github/workflows/release.yml`: replace all `unbound_*_darwin_*.tar.gz` glob patterns with `unbound-force_*_darwin_*.tar.gz`, replace all `$workdir/unbound` binary paths with `$workdir/unbound-force`, replace all `unbound.rb` cask file references with `unbound-force.rb`, update the final `git commit -m` message from `unbound cask` to `unbound-force cask`
- [x] T007 Update `cmd/unbound-force/main_test.go`: change `TestVersionCmd_Output` assertions from `"unbound v"` prefix to `"unbound-force v"` prefix, update format comment from `"unbound vX"` to `"unbound-force vX"`
- [x] T007b Update `internal/scaffold/scaffold_test.go`: change `printSummary` assertions from `"unbound init"` to `"uf init"` (4 locations: lines ~734, ~779, ~821, ~850)
- [x] T007c Update `internal/doctor/doctor_test.go`: change install hint assertions from `"unbound setup"` to `"uf setup"` (lines ~1067-1068, ~1159-1160)
- [x] T007d Update `internal/scaffold/scaffold.go`: change `printSummary` label from `"unbound init"` to `"uf init"` and `versionMarker` format from `"scaffolded by unbound v%s"` to `"scaffolded by uf v%s"` (line ~320 and ~376)
- [x] T007e Update all 6 template files under `internal/scaffold/assets/specify/templates/` to change `<!-- scaffolded by unbound vdev -->` markers to `<!-- scaffolded by uf vdev -->`
- [x] T008 Run `go test -race -count=1 ./...` and `go build ./...` to verify all tests pass with the renamed directory

**Checkpoint**: Build and test pipeline works. `go install ./cmd/unbound-force/` produces the correct binary name. `make install` creates both `unbound-force` and `uf`.

---

## Phase 3: User Story 1 + 2 - CLI Rename and Scaffold Assets (Priority: P1)

**Goal**: The binary is named `unbound-force` with `uf` alias, and all scaffold assets reference the correct name.

**Independent Test**: Run `go build -o /tmp/unbound-force ./cmd/unbound-force/` and verify `/tmp/unbound-force --help` shows `unbound-force (alias: uf)`. Run the init function in a temp dir and grep for bare `unbound ` references in output files.

### Implementation

- [x] T009 [P] [US1] Update `internal/doctor/checks.go`: replace all hint strings referencing `unbound setup` → `uf setup` and `unbound init` → `uf init`
- [x] T010 [P] [US1] Update `internal/doctor/format.go`: replace any `unbound` display text in formatted output with `uf` or `unbound-force` as appropriate
- [x] T011 [P] [US1] Update `internal/setup/setup.go`: replace `unbound init` references in progress messages with `uf init`
- [x] T012 [P] [US2] Update scaffold asset `internal/scaffold/assets/opencode/agents/cobalt-crush-dev.md`: replace `unbound init --divisor` with `uf init --divisor`
- [x] T013 [P] [US2] Update scaffold asset `internal/scaffold/assets/opencode/agents/reviewer-architect.md`: replace any `cmd/unbound` path references with `cmd/unbound-force`
- [x] T014 [US2] Search all remaining scaffold assets under `internal/scaffold/assets/` for bare `unbound init`, `unbound doctor`, `unbound setup`, `unbound version`, and `cmd/unbound` path references. Known files to check: `reviewer-adversary.md` (3 refs), `reviewer-guard.md` (1 ref), `reviewer-sre.md` (5 refs), `specify/config.yaml` (1 ref). Replace all with `uf` or `unbound-force` equivalents
- [x] T015 [US1] Run `go test -race -count=1 ./...` to verify all doctor, setup, and scaffold tests pass with updated strings

**Checkpoint**: The CLI binary is named correctly, help output shows `(alias: uf)`, doctor/setup hints reference `uf`, and scaffold assets reference `uf`/`unbound-force`.

---

## Phase 4: User Story 3 - Homebrew Distribution (Priority: P2)

**Goal**: The GoReleaser config and Homebrew cask produce the correct binary name. The `uf` alias is handled via a cask `binary` stanza or post-install hook.

**Independent Test**: Run `goreleaser check` to validate the config. Verify the `homebrew_casks` section has the correct cask name `unbound-force`.

- [x] T016 [US3] Verify `.goreleaser.yaml` `homebrew_casks` section has the correct cask name `unbound-force`, quarantine hook references `unbound-force`, and `dependencies` section references are updated (already done in T006 -- this is a verification task)
- [x] T017 [US3] Research and implement the `uf` symlink mechanism for Homebrew casks: either add a `binary "unbound-force", target: "uf"` stanza in the cask, or add a post-install hook that creates the `uf` symlink. Update `.goreleaser.yaml` `homebrew_casks` section accordingly
- [x] T017b [US3] Verify the `release.yml` workflow references are all updated to `unbound-force` (already done in T006b -- this is a verification task)

**Checkpoint**: GoReleaser config produces correct release artifacts. Homebrew cask installs `unbound-force` and creates `uf` alias.

---

## Phase 5: User Story 4 - Doctor and Setup Output (Priority: P2)

**Goal**: All doctor and setup output strings reference `uf`.

**Independent Test**: Run doctor/setup tests with injected dependencies and verify hint strings.

- [x] T018 [US4] Search `internal/doctor/` for any remaining `unbound ` references (case-sensitive, excluding `unbound-force` and `unbound_force`) and replace with `uf`
- [x] T019 [US4] Search `internal/setup/` for any remaining `unbound ` references and replace with `uf`
- [x] T020 [US4] Run `go test -race -count=1 ./internal/doctor/... ./internal/setup/...` to verify hint string tests pass

**Checkpoint**: `uf doctor` and `uf setup` output reference `uf` in all hints and progress messages.

---

## Phase 6: User Story 5 - Living Documentation (Priority: P3)

**Goal**: All living documentation in the meta repo references `uf` or `unbound-force`. Completed specs are not modified.

**Independent Test**: `grep -rn 'unbound init\|unbound doctor\|unbound setup' AGENTS.md README.md .opencode/ internal/scaffold/assets/` returns zero matches.

- [x] T021 [P] [US5] Update `AGENTS.md`: replace all bare `unbound init`, `unbound doctor`, `unbound setup`, `unbound version` references with `uf` equivalents in living sections (Build & Test Commands, Coding Conventions, etc.). Replace `brew install unbound-force/tap/unbound` with `brew install unbound-force/tap/unbound-force`. Replace `unbound init` with `uf init` in scaffold references. Do NOT modify completed spec references in the Recent Changes section that describe historical work.
- [x] T022 [P] [US5] Update `README.md`: replace all bare `unbound` CLI references with `uf` or `unbound-force` as appropriate
- [x] T023 [P] [US5] Update `.opencode/agents/cobalt-crush-dev.md` (the live copy, not scaffold asset): replace `unbound init --divisor` with `uf init --divisor`
- [x] T024 [P] [US5] Update `.opencode/agents/reviewer-architect.md` (the live copy): replace any `cmd/unbound` path references with `cmd/unbound-force`
- [x] T025 [US5] Search all files under `.opencode/` (excluding `skill/` which has no CLI refs) for remaining bare `unbound init`, `unbound doctor`, `unbound setup` references and replace
- [x] T026 [US5] Search `unbound-force.md` for any bare `unbound` CLI references and replace
- [x] T027 [US5] Verify that all completed specs under `specs/` (excluding `specs/013-binary-rename/`) are NOT modified: run `git diff` and confirm no changes in `specs/001-*` through `specs/012-*`. Archived OpenSpec changes under `openspec/changes/archive/` are also excluded as historical records

**Checkpoint**: Zero bare `unbound` CLI references in living documentation. Completed specs unchanged.

---

## Phase 7: New Regression Tests

**Purpose**: Permanent automated guards for the rename requirements (FR-004, FR-015, SC-003).

- [x] T028 [P] Write `TestRootCmd_HelpOutput` in `cmd/unbound-force/main_test.go`: create the root command, capture `--help` output, assert it contains `"(alias: uf)"` and `"unbound-force [command]"` in usage line (FR-004 regression guard)
- [x] T029 [P] Write `TestScaffoldOutput_NoBareUnboundReferences` in `internal/scaffold/scaffold_test.go`: run scaffold in a temp dir, read all generated files, search for bare `unbound init`, `unbound doctor`, `unbound setup`, `unbound version` patterns, assert zero matches (FR-015/SC-003 regression guard)
- [x] T030 [P] Write `TestDoctorHints_NoBareUnboundReferences` in `internal/doctor/doctor_test.go`: run the full doctor check suite with injected dependencies, search all `InstallHint` fields across results for bare `unbound ` references (excluding `unbound-force`), assert zero matches (FR-006 regression guard)

---

## Phase 8: Polish & Final Validation

**Purpose**: Final validation, full test suite, cross-repo notes.

- [x] T031 Run full test suite `go test -race -count=1 ./...` and `go build ./...` to verify zero regressions
- [x] T032 Build the binary and verify `--help` output matches the contract in `contracts/cli-schema.md`: `unbound-force (alias: uf)` in description, `unbound-force [command]` in usage
- [x] T033 Run `make install` and verify both `$(GOPATH)/bin/unbound-force` and `$(GOPATH)/bin/uf` exist and produce identical `--help` output
- [x] T034 [P] Grep living documentation and CI workflows for stale references: `grep -rn 'unbound init\|unbound doctor\|unbound setup\|unbound version' AGENTS.md README.md .opencode/ internal/scaffold/assets/ .github/workflows/` -- verify zero matches (excluding completed specs)
- [x] T035 [P] Register spec 013 in AGENTS.md's Project Structure, Spec Organization (Phase 3: Infrastructure), and Dependency Graph sections

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- directory rename first
- **Foundational (Phase 2)**: Depends on Phase 1 -- build/test infrastructure
- **US1+US2 (Phase 3)**: Depends on Phase 2 -- string replacements in Go code and scaffold assets
- **US3 (Phase 4)**: Depends on Phase 2 -- GoReleaser verification
- **US4 (Phase 5)**: Can run in parallel with Phase 3 (same files but different sections; sequence if concerned about conflicts)
- **US5 (Phase 6)**: Can run in parallel with Phase 3-5 (different files -- Markdown docs vs Go code)
- **Regression Tests (Phase 7)**: Depends on Phase 3 and Phase 5 (need production code updated before writing sweep tests)
- **Polish (Phase 8)**: Depends on all phases complete

### User Story Dependencies

- **US1 (P1)** + **US2 (P1)**: Tightly coupled -- the directory rename (Phase 1) and Cobra update enable both. Combined in Phase 3.
- **US3 (P2)**: Independent -- GoReleaser config only. Can run after Phase 2.
- **US4 (P2)**: Overlaps with US1 (doctor/setup strings). Sequenced after Phase 3 to catch any missed references.
- **US5 (P3)**: Independent -- Markdown documentation only. Can run in parallel with US1-US4.

### Parallel Opportunities

- T009, T010, T011, T012, T013 can all run in parallel (different files)
- T021, T022, T023, T024 can all run in parallel (different Markdown files)
- Phase 4 (US3) and Phase 6 (US5) can run in parallel with Phase 3 (US1+US2)

---

## Implementation Strategy

### MVP First (US1 + US2 Only)

1. Phase 1: Directory rename (T001-T004)
2. Phase 2: Build infrastructure (T005-T008)
3. Phase 3: CLI + scaffold assets (T009-T015)
4. **STOP and VALIDATE**: `go test` passes, `--help` shows `unbound-force (alias: uf)`, scaffold output is clean
5. Core value delivered -- the name collision is resolved

### Incremental Delivery

1. Setup + Foundational → Build works with new name
2. US1+US2 → CLI and scaffold correct → Core problem solved
3. US3 → GoReleaser ready → Distribution solved
4. US4 → Doctor/setup hints correct → Onboarding polished
5. US5 → All docs updated → Full consistency
6. Polish → Final validation → Ship it

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story
- The rename is primarily string replacements after the directory rename
- Completed specs are explicitly excluded from modification (FR-014)
- Cross-repo updates (gaze, website, homebrew-tap) are separate work items outside this tasks.md -- they depend on this meta repo work being merged first
- The `uf` alias is preferred in user-facing hints; `unbound-force` is used in formal contexts (Homebrew, `go install`, build commands)
