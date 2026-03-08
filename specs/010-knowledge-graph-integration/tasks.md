# Tasks: Knowledge Graph Integration

**Input**: Design documents from `/specs/010-knowledge-graph-integration/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: No automated tests requested. Verification is manual
via MCP tool invocations from OpenCode agents.

**Organization**: Tasks are grouped by user story to enable
independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

This is an integration/configuration feature. Deliverables are
configuration files at the project root, an upstream PR to the
graphthulhu repository, and content enrichment of existing
Markdown files. No new source code directories are created.

---

## Phase 1: Setup

**Purpose**: Install graphthulhu and establish the MCP server
configuration so OpenCode can launch the knowledge graph
service.

- [x] T001 Install graphthulhu binary via `go install github.com/skridlevsky/graphthulhu@latest` and verify with `graphthulhu version`
- [x] T002 Create OpenCode MCP server configuration file at opencode.json in the project root, declaring the `knowledge-graph` server as `"type": "local"` with command `["graphthulhu", "serve", "--backend", "obsidian", "--vault", ".", "--read-only"]` per research.md R-001
- [x] T003 Start OpenCode with the new configuration and verify the knowledge graph MCP server launches successfully as a subprocess (check `health` tool response returns version, backend type, and page count)

**Checkpoint**: graphthulhu is installed, configured in
OpenCode, and launches successfully. Agents can invoke the
`health` tool. Note: `--include-hidden` flag is omitted until
the upstream PR (Phase 2) is complete. Hidden directories
(`.specify/`, `.opencode/`) are not yet indexed.

---

## Phase 2: Foundational (Hidden Directory Support)

**Purpose**: Contribute the `--include-hidden` flag to
graphthulhu upstream so that content in `.specify/` and
`.opencode/` can be indexed (FR-012, SC-008).

- [x] T004 Fork graphthulhu repository to the `unbound-force` GitHub organization (or personal fork) as a working branch for the upstream PR
- [x] T005 Add `includeHidden bool` field to `vault.Client` struct and `WithIncludeHidden(bool) Option` constructor in vault/vault.go, following the existing `WithDailyFolder` pattern
- [x] T006 Guard the hidden directory skip in `vault.Client.Load()` method in vault/vault.go with `&& !c.includeHidden` on the `strings.HasPrefix(info.Name(), ".")` check, while always skipping `.git` directories regardless of flag
- [x] T007 [P] Guard the hidden directory skip in `vault.Client.addWatcherDirs()` method in vault/vault.go with the same `&& !c.includeHidden` guard
- [x] T008 [P] Guard the hidden directory skip in `vault.Client.handleEvent()` method in vault/vault.go by adding `&& !c.includeHidden` to the `strings.Contains(event.Name, "/.")` check
- [x] T009 Add `--include-hidden` CLI flag in main.go `runServe()` function, passing it to `vault.New()` via `vault.WithIncludeHidden(*includeHidden)`
- [x] T010 Add test in vault/vault_test.go that creates a temporary vault with a hidden directory containing `.md` files, verifies they are skipped by default, and included when `WithIncludeHidden(true)` is set
- [x] T011 Run `go test ./...` and `go vet ./...` in the graphthulhu fork to verify all existing tests pass with the new changes
- [x] T012 Submit upstream pull request to `skridlevsky/graphthulhu` with the `--include-hidden` changes, referencing the use case of indexing project configuration directories like `.specify/` and `.opencode/`
- [x] T013 Update opencode.json to add `"--include-hidden"` to the graphthulhu command array (either after upstream PR is merged and a new release is available, or by installing from the fork branch)
- [x] T014 Verify that content in `.specify/memory/constitution.md` and `.opencode/agents/constitution-check.md` appears in search results after enabling `--include-hidden`

**Checkpoint**: Hidden directory support is contributed
upstream (or available via fork). The OpenCode configuration
includes `--include-hidden`. All project Markdown files,
including those in `.specify/` and `.opencode/`, are indexed
and searchable.

---

## Phase 3: User Story 1 - Agent Knowledge Retrieval via MCP (Priority: P1) MVP

**Goal**: Hero agents can search, retrieve, and list project
knowledge via MCP without loading files directly into their
context windows.

**Independent Test**: Start OpenCode, invoke `search`,
`get_page`, and `list_pages` tools, and verify results match
expected content from the repo's Markdown files.

### Implementation for User Story 1

- [x] T015 [US1] Invoke the `search` MCP tool with a query term that appears in multiple spec files (e.g., "constitution") and verify results return matches from all relevant files with parent block chain context
- [x] T016 [US1] Invoke the `get_page` MCP tool with a known spec page name (e.g., "specs/004-muti-mind-architecture/spec") and verify the response contains the full hierarchical block tree with parsed links, tags, and properties
- [x] T017 [US1] Invoke the `list_pages` MCP tool and verify the response lists all indexed Markdown files with their path-qualified names and metadata
- [x] T018 [US1] Verify that search results return contextual information (parent block chain and sibling blocks) by searching for a term that appears within a nested heading section and confirming ancestor headings are included in the response
- [x] T019 [US1] Verify that the `get_block` MCP tool retrieves a specific block by UUID (obtain a UUID from a `get_page` response) and returns the block with its ancestor chain and sibling context
- [x] T020 [US1] Verify read-only mode enforcement (FR-015): invoke a write tool (e.g., `create_page` or `append_blocks`) and confirm the service returns an MCP error rejecting the operation because `--read-only` is enabled
- [x] T021 [US1] Verify file non-modification (FR-019): record file timestamps of 3 indexed Markdown files before a query session, invoke `search`, `get_page`, and `list_pages`, then confirm all file timestamps are unchanged after the queries
- [x] T022 [US1] Document any discrepancies between expected and actual MCP tool behavior in a verification log at specs/010-knowledge-graph-integration/verification-log.md

**Checkpoint**: Agents can search across all project files,
retrieve specific pages by name, list all indexed content,
and get individual blocks by UUID. Read-only mode is verified.
This is the MVP -- the core value proposition is delivered.

---

## Phase 4: User Story 2 - Knowledge Graph Analysis (Priority: P2)

**Goal**: Agents can analyze the structure and health of the
project's knowledge base, identifying orphans, clusters, and
connection patterns.

**Independent Test**: Invoke `graph_overview`, `knowledge_gaps`,
`list_orphans`, and `topic_clusters` tools and verify results
reflect the known structure of the unbound-force meta repo.

### Implementation for User Story 2

- [x] T023 [US2] Invoke the `graph_overview` MCP tool and verify it returns accurate statistics: total pages (matching the known file count), total blocks, and namespace breakdown (specs/, .specify/, .opencode/ etc.)
- [x] T024 [US2] Invoke the `knowledge_gaps` MCP tool and verify it correctly identifies orphan pages (pages with no wikilinks in or out) -- note: most current pages will be orphans since wikilinks have not been added yet
- [x] T025 [US2] Invoke the `list_orphans` MCP tool and verify the orphan list matches expectations (all pages should be orphans until content enrichment adds wikilinks)
- [x] T026 [US2] Invoke the `topic_clusters` MCP tool and verify it returns connected components -- note: with no wikilinks, each page will be its own isolated cluster
- [x] T027 [US2] Document analysis tool results and any discrepancies in specs/010-knowledge-graph-integration/verification-log.md

**Checkpoint**: Graph analysis tools work correctly. Results
accurately reflect the current (un-enriched) state of the
repo. The tools will become more valuable after content
enrichment adds wikilinks (Phase 6).

---

## Phase 5: User Story 3 - Live Content Synchronization (Priority: P2)

**Goal**: The knowledge graph automatically re-indexes when
files change, ensuring agents always query current content.

**Independent Test**: Create a new Markdown file, modify an
existing one, and delete a file, verifying search results
update within 5 seconds each time without restarting the
service.

### Implementation for User Story 3

- [x] T028 [US3] Create a temporary test Markdown file (e.g., specs/010-knowledge-graph-integration/test-live-sync.md) with unique content, wait 5 seconds, then search for that content and verify it appears in results
- [x] T029 [US3] Modify the test file to change its content, wait 5 seconds, then search for the new content and verify the updated content appears (and old content does not)
- [x] T030 [US3] Delete the test file, wait 5 seconds, then search for its content and verify it no longer appears in results
- [x] T031 [US3] Invoke the `reload` MCP tool and verify it returns a confirmation with the correct number of re-indexed pages
- [x] T032 [US3] Document live sync verification results in specs/010-knowledge-graph-integration/verification-log.md

**Checkpoint**: File watching and live re-indexing work
correctly. Agents always see current content without manual
intervention.

---

## Phase 6: User Story 4 - Cross-Spec Link Traversal (Priority: P3)

**Goal**: Agents can follow wikilink relationships between
project artifacts, enabling traversal from backlog items to
specs to constitution principles.

**Independent Test**: Add wikilinks to a few spec files, then
use `get_links`, `find_connections`, and `traverse` tools to
navigate the link graph.

### Implementation for User Story 4

- [x] T033 [P] [US4] Add YAML frontmatter and wikilinks to specs/001-org-constitution/spec.md, adding frontmatter with `spec_id`, `title`, `phase`, `status`, and `depends_on` fields, and converting prose cross-references to `[[wikilinks]]` where specs reference each other
- [x] T034 [P] [US4] Add YAML frontmatter and wikilinks to specs/002-hero-interface-contract/spec.md with the same frontmatter schema, including `depends_on: ["[[001-org-constitution]]"]` and converting cross-references
- [x] T035 [P] [US4] Add YAML frontmatter and wikilinks to specs/004-muti-mind-architecture/spec.md with frontmatter including `depends_on` references to specs 001 and 002
- [x] T036 [US4] Invoke the `get_links` MCP tool on one of the enriched specs and verify it returns both outbound links (specs this page references) and inbound backlinks (specs that reference this page) with containing block context
- [x] T037 [US4] Invoke the `find_connections` MCP tool with two enriched spec page names and verify it returns direct links, shortest paths, and shared connections between them
- [x] T038 [US4] Invoke the `traverse` MCP tool between two specs connected through an intermediate page and verify it returns the shortest path through the link graph
- [x] T039 [US4] Document link traversal verification results in specs/010-knowledge-graph-integration/verification-log.md

**Checkpoint**: Wikilink-based navigation works. Agents can
traverse from one spec to related specs through the link
graph. The foundation is laid for richer content enrichment
across all specs.

---

## Phase 7: User Story 5 - Property-Based Querying (Priority: P3)

**Goal**: Agents can find artifacts by structured metadata
(YAML frontmatter properties) such as status, phase, priority,
or hero assignment.

**Independent Test**: Query for specs by `status`, `phase`, or
`depends_on` properties using the `query_properties` tool and
verify correct results.

### Implementation for User Story 5

- [x] T040 [US5] Verify that the YAML frontmatter added in Phase 6 (T033-T035) is correctly parsed by invoking `get_page` on an enriched spec and confirming properties appear in the response
- [x] T041 [US5] Invoke the `query_properties` MCP tool with property `status` equal to `draft` and verify it returns all spec files with that status in their frontmatter
- [x] T042 [US5] Invoke the `query_properties` MCP tool with property `phase` to filter specs by implementation phase and verify correct results
- [x] T043 [US5] Invoke the `find_by_tag` MCP tool with a tag present in enriched content and verify it returns the correct blocks and pages
- [x] T044 [US5] Document property-based querying verification results in specs/010-knowledge-graph-integration/verification-log.md

**Checkpoint**: Property-based querying works for enriched
files. Agents can find specs by metadata without full-text
search. Further content enrichment (adding frontmatter to all
specs and future backlog items) will increase the value of
this capability.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Documentation updates, AGENTS.md alignment, and
cleanup.

- [x] T045 [P] Update AGENTS.md to add spec 010 to the Recent Changes section and update the spec organization tables to include 010-knowledge-graph-integration in the dependency graph
- [x] T046 [P] Update README.md to mention the knowledge graph integration as available tooling for the Unbound Force swarm
- [x] T047 Add remaining YAML frontmatter to specs 003, 005, 006, 007, 008, and 009 following the same schema used in T033-T035 (spec_id, title, phase, status, depends_on) in their respective spec.md files
- [x] T048 Review and finalize specs/010-knowledge-graph-integration/verification-log.md, ensuring all user stories have documented verification results
- [x] T049 Run the `graph_overview` and `knowledge_gaps` MCP tools one final time to produce a baseline knowledge graph health report for the fully enriched meta repo

**Checkpoint**: All documentation is updated. All spec files
have YAML frontmatter. The knowledge graph reflects the full
enriched state of the meta repo.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- can start immediately
- **Foundational (Phase 2)**: Depends on Setup (T001-T003) -- BLOCKS full hidden directory indexing but does NOT block user story work for non-hidden content
- **User Story 1 (Phase 3)**: Can start after Setup (Phase 1) -- works with non-hidden content even before Phase 2 completes
- **User Story 2 (Phase 4)**: Can start after Setup (Phase 1) -- analysis tools work on any indexed content
- **User Story 3 (Phase 5)**: Can start after Setup (Phase 1) -- file watching works immediately
- **User Story 4 (Phase 6)**: Can start after Setup (Phase 1) -- but benefits most after Phase 2 (hidden dirs indexed)
- **User Story 5 (Phase 7)**: Depends on User Story 4 (Phase 6) -- uses the frontmatter added in T033-T035
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Independent -- works immediately after Setup
- **User Story 2 (P2)**: Independent -- works immediately after Setup (results improve after content enrichment)
- **User Story 3 (P2)**: Independent -- works immediately after Setup
- **User Story 4 (P3)**: Independent of other stories -- adds content enrichment that US5 and US2 benefit from
- **User Story 5 (P3)**: Depends on US4's content enrichment (T033-T035) for YAML frontmatter to query

### Within Each User Story

- Verification tasks are sequential (each builds on prior tool invocations)
- Content enrichment tasks (T033-T035) are parallelizable across different files
- Documentation tasks are always last within a phase

### Parallel Opportunities

- T007 and T008 can run in parallel (different methods in same file, independent guards)
- T033, T034, T035 can run in parallel (different spec files, independent edits)
- T045 and T046 can run in parallel (different documentation files)
- US1, US2, and US3 can all proceed in parallel after Phase 1 Setup completes
- US4 and US5 should run sequentially (US5 depends on US4's content enrichment)

---

## Parallel Example: User Story 4

```text
# Launch content enrichment for multiple specs in parallel:
Task: "Add YAML frontmatter and wikilinks to specs/001-org-constitution/spec.md"
Task: "Add YAML frontmatter and wikilinks to specs/002-hero-interface-contract/spec.md"
Task: "Add YAML frontmatter and wikilinks to specs/004-muti-mind-architecture/spec.md"
```

## Parallel Example: Phase 2 (Upstream PR)

```text
# These guard changes are in different methods, can be done in parallel:
Task: "Guard hidden dir skip in addWatcherDirs() in vault/vault.go"
Task: "Guard hidden dir skip in handleEvent() in vault/vault.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 3: User Story 1 (T015-T022)
3. **STOP and VALIDATE**: Verify agents can search, retrieve
   pages, list content, and confirm read-only enforcement
4. This delivers the core value: agents can query project
   knowledge via MCP without exhausting context windows

### Incremental Delivery

1. Setup → US1 verified → **MVP delivered**
2. Add US2 (analysis) → Verify graph stats and gap detection
3. Add US3 (live sync) → Verify file watching works
4. Complete Phase 2 (upstream PR) → Hidden dirs indexed
5. Add US4 (link traversal) → Enrich 3 specs with wikilinks
6. Add US5 (property queries) → Query by frontmatter properties
7. Polish → Enrich remaining specs, update documentation

### Note on Phase 2 Timing

The upstream PR (Phase 2) can proceed in parallel with user
story work. US1-US3 work fine without hidden directory support
-- they just index fewer files. The PR can be submitted early
and merged on its own timeline. Once merged, T013 adds the
flag to the config and T014 verifies hidden content is indexed.

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- No automated tests are generated (not requested in spec)
- Verification is manual via MCP tool invocations
- The upstream PR (Phase 2) is the only task requiring Go
  development -- all other tasks are configuration, content
  editing, and verification
- Content enrichment (adding frontmatter/wikilinks) is scoped
  to 3 specs initially (T033-T035). Full enrichment of all 9
  specs is in Polish (T047).
