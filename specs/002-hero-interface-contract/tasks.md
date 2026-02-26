# Tasks: Hero Interface Contract

**Input**: Design documents from
`/specs/002-hero-interface-contract/`
**Prerequisites**: plan.md (required), spec.md (required),
research.md, data-model.md, quickstart.md

This feature produces governance documents, a JSON Schema,
and a bash validation script — not compiled software. Tasks
create and validate these artifacts.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no deps)
- **[Story]**: Which user story (US1-US5) the task belongs to

## Path Conventions

- Contract document: `specs/002-hero-interface-contract/contract.md`
- Hero manifest schema: `schemas/hero-manifest/v1.0.0.schema.json`
- Validation script: `scripts/validate-hero-contract.sh`

---

## Phase 1: Setup

**Purpose**: Create directory structure and contract skeleton

- [x] T001 Create `schemas/hero-manifest/` directory at repository root for the hero manifest JSON Schema
- [x] T002 Create `scripts/` directory at repository root for the validation script
- [x] T003 Create the Hero Interface Contract document skeleton at `specs/002-hero-interface-contract/contract.md` with section headings for all five contract areas: Repository Structure, Artifact Protocol, Speckit Integration, OpenCode Conventions, Hero Manifest — include document header with version (1.0.0), date, and references to the org constitution and dependent specs

**Checkpoint**: Directory structure and contract skeleton ready.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Contract preamble and shared definitions that
all user story sections reference

**CRITICAL**: These sections must exist before individual
contract sections can be written, because they define the
terminology and principles referenced throughout.

- [x] T004 Write the contract preamble in `specs/002-hero-interface-contract/contract.md` — define the purpose of the Hero Interface Contract, its authority (per org constitution), scope (all hero repos), and relationship to Specs 001, 003, and 009
- [x] T005 Write the hero lifecycle section (FR-011) in `specs/002-hero-interface-contract/contract.md` — define the six lifecycle stages: bootstrap -> constitution -> specify -> implement -> deploy -> maintain, with a brief description of each stage and which contract requirements apply at each
- [x] T006 Write the version compatibility rules section (FR-013) in `specs/002-hero-interface-contract/contract.md` — define how heroes handle version incompatibilities when consuming artifacts: same MAJOR = compatible, different MAJOR = graceful degradation, unknown artifact type = ignore with optional warning

**Checkpoint**: Contract preamble, lifecycle, and versioning
rules are in place. User story sections can now reference
these shared definitions.

---

## Phase 3: User Story 1 - Standard Repository Structure (Priority: P1)

**Goal**: Define the minimum required directory structure for
every hero repository and create a validation script that
checks any repository against it.

**Independent Test**: Run the validation script against an
empty directory and verify it reports all missing required
elements. Then run it against the Gaze repo and verify it
identifies the known gaps (missing `hero.json`, missing
`parent_constitution` reference).

### Implementation for User Story 1

- [x] T007 [US1] Write the "Repository Structure" section in `specs/002-hero-interface-contract/contract.md` — define all required and optional elements from data-model.md Hero Repository Structure table (FR-001): `.specify/memory/constitution.md`, `.specify/templates/`, `.specify/scripts/bash/`, `.opencode/`, `.opencode/command/`, `specs/`, `AGENTS.md`, `LICENSE`, `README.md`, `.unbound-force/hero.json`. Mark each as MUST (required) or SHOULD (optional) per RFC 2119.
- [x] T008 [US1] Write the "Constitution Requirements" subsection in `specs/002-hero-interface-contract/contract.md` — define that hero constitutions MUST include a `parent_constitution` field referencing the org constitution version (FR-009), MUST NOT contradict org principles, and MUST be ratified before a hero is considered contract-compliant
- [x] T009 [US1] Write the "License Requirements" subsection in `specs/002-hero-interface-contract/contract.md` — define that Apache 2.0 is RECOMMENDED, alternatives MAY be used with explicit justification in the hero's constitution (per edge case)
- [x] T010 [US1] Write the "README Template" subsection in `specs/002-hero-interface-contract/contract.md` — define the SHOULD-level hero README template structure per FR-015: hero name, role, installation, quick start, integration with other heroes, link to org README
- [x] T011 [US1] Create the validation script at `scripts/validate-hero-contract.sh` — implement bash script per FR-008 and research.md Task 4: accept a repository path as argument, check all required elements from T007, check constitution `parent_constitution` reference from T008, check `.unbound-force/hero.json` existence and basic JSON validity, categorize checks as required (errors) or optional (warnings), output structured pass/fail results per quickstart.md expected output format, exit 0 on PASS, exit 1 on FAIL
- [x] T012 [US1] Run `scripts/validate-hero-contract.sh` against the Gaze repo at `/Users/jflowers/Projects/github/unbound-force/gaze` — verify it produces the expected results from quickstart.md Gaze compliance table (PASS for most items, FAIL for `.unbound-force/hero.json` and `parent_constitution` ref, WARN for agent naming). Document the actual output as validation evidence.
- [x] T013 [US1] Run `scripts/validate-hero-contract.sh` against the Website repo at `/Users/jflowers/Projects/github/unbound-force/website` — verify it produces the expected results from quickstart.md Website compliance table (PASS for most items, FAIL for `.unbound-force/hero.json` and `parent_constitution` ref). Document the actual output as validation evidence.

**Checkpoint**: Contract defines repository structure.
Validation script runs and correctly identifies gaps in
Gaze and Website repos. SC-001, SC-002, SC-003 satisfied.

---

## Phase 4: User Story 2 - Inter-Hero Artifact Protocol (Priority: P1)

**Goal**: Define the artifact envelope format conceptually
and the registry of standard artifact types, so heroes know
how to produce and consume inter-hero artifacts.

**Independent Test**: Create a sample artifact JSON matching
the envelope format, verify it contains all required fields
from the contract, and confirm a Gaze quality report could
be wrapped in the envelope.

### Implementation for User Story 2

- [x] T014 [US2] Write the "Artifact Envelope" section in `specs/002-hero-interface-contract/contract.md` — define the conceptual envelope format per FR-003 and data-model.md: required fields (`hero`, `version`, `timestamp`, `artifact_type`, `schema_version`, `context`, `payload`), context sub-fields (`branch`, `commit`, `backlog_item_id`, `correlation_id`), field types and constraints (semver, ISO 8601, UUID). Note that the formal JSON Schema is deferred to Spec 009.
- [x] T015 [US2] Write the "Artifact Type Registry" section in `specs/002-hero-interface-contract/contract.md` — define the initial set of registered artifact types per FR-004 and data-model.md registry table: `quality-report`, `review-verdict`, `backlog-item`, `acceptance-decision`, `metrics-snapshot`, `coaching-record`, `workflow-record`. For each, list producer hero, consumer heroes, and brief description. Note that JSON Schemas for payloads are deferred to Spec 009.
- [x] T016 [US2] Write the "Artifact Versioning" subsection in `specs/002-hero-interface-contract/contract.md` — define the schema versioning rules already captured in T006 but now specific to artifact types: MAJOR bump for breaking payload changes, MINOR for backward-compatible additions, PATCH for documentation. Include the migration requirement from FR-013: consumers MUST check `schema_version`, same MAJOR = proceed, different MAJOR = graceful degradation or rejection.
- [x] T017 [US2] Write the "Producing and Consuming Artifacts" subsection in `specs/002-hero-interface-contract/contract.md` — define producer obligations (MUST conform to registered schema, MUST include all envelope fields, MUST use registered `artifact_type` values) and consumer obligations (MUST handle valid instances, MUST ignore unknown types gracefully, SHOULD log warnings for unknown types). Reference quickstart.md producing/consuming sections.
- [x] T018 [US2] Create a sample artifact file at `schemas/samples/sample-quality-report-envelope.json` — a valid quality-report artifact wrapped in the envelope format using Gaze as producer, with realistic payload matching Gaze's current quality report structure. Verify all envelope fields are present and correctly typed. This validates SC-004.

**Checkpoint**: Contract defines artifact protocol. Sample
artifact demonstrates the envelope format. SC-004 satisfied.

---

## Phase 5: User Story 3 - Speckit Framework Integration (Priority: P2)

**Goal**: Define speckit integration requirements — canonical
source, drift resolution approach, and extension point
mechanism — without duplicating Spec 003's distribution
mechanism.

**Independent Test**: Compare the contract's speckit
requirements against the known drift points from the speckit
drift report (research.md Task 2) and verify the contract
addresses each point.

### Implementation for User Story 3

- [x] T019 [US3] Write the "Speckit Integration" section in `specs/002-hero-interface-contract/contract.md` — define that hero repos MUST install speckit from the canonical source per FR-007, referencing Spec 003 for the distribution mechanism. List the canonical file inventory: 6 templates in `.specify/templates/`, 5 scripts in `.specify/scripts/bash/`, 9+ speckit commands in `.opencode/command/`. State that the `unbound-force` (org repo) versions are canonical where drift exists.
- [x] T020 [US3] Write the "Speckit Extension Points" subsection in `specs/002-hero-interface-contract/contract.md` — define that project-specific customizations MUST be handled via `.specify/config.yaml` configuration (per Spec 003 FR-007), NOT by modifying canonical speckit files. Reference the known drift points: `speckit.specify.md` integration patterns, `speckit.plan.md` contract generation, `speckit.tasks.md` terminology.
- [x] T021 [US3] Write the "Speckit Update Protocol" subsection in `specs/002-hero-interface-contract/contract.md` — define that updates MUST preserve local modifications (detect via checksum, skip with warning), new files MUST be installed, the update tool MUST NOT overwrite without `--force`. Reference edge case about uncommitted local changes.

**Checkpoint**: Contract defines speckit integration
requirements. SC-007 satisfied (drift mechanism defined,
canonical source identified).

---

## Phase 6: User Story 4 - OpenCode Plugin and Agent Standards (Priority: P2)

**Goal**: Define naming conventions and format standards for
OpenCode agents and commands so heroes can provide agents
that coexist without collisions.

**Independent Test**: Inspect the Gaze repo's `.opencode/`
directory against the conventions defined in this section and
verify compliance or identify deviations.

### Implementation for User Story 4

- [x] T022 [US4] Write the "OpenCode Agent Standards" section in `specs/002-hero-interface-contract/contract.md` — define agent naming convention per FR-005: `{hero-name}-{agent-function}.md`, mandatory file sections (descriptive header, tool permissions, model specification, behavioral constraints), and collision prevention rules (hero prefix required, last `init` wins without `--force`). Include examples from quickstart.md.
- [x] T023 [US4] Write the "OpenCode Command Standards" subsection in `specs/002-hero-interface-contract/contract.md` — define command naming convention per FR-006: `/{hero-name}` for primary, `/{hero-name}-{subfunction}` for secondary. Define mandatory command file sections: trigger syntax, argument parsing, agent delegation, error handling. Include examples.
- [x] T024 [US4] Write the "Init Command Convention" subsection in `specs/002-hero-interface-contract/contract.md` — define the SHOULD-level `{hero-name} init` convention per FR-012: scaffolds agent files and command files into target project's `.opencode/` directory, detects existing files and warns on collision, supports `--force` flag for overwrite.
- [x] T025 [US4] Write the "MCP Server Conventions" subsection in `specs/002-hero-interface-contract/contract.md` — define the SHOULD-level naming conventions per FR-014: `{hero-name}-mcp` for server name, `{hero-name}_{tool_name}` for tool names. Note that full MCP interface requirements are deferred to hero-specific architecture specs.
- [x] T026 [US4] Audit Gaze OpenCode agents against the conventions defined in T022-T023 — inspect all files in Gaze's `.opencode/agents/` and `.opencode/command/` directories, compare naming against the `{hero-name}-{agent-function}` convention, identify deviations (expected: `doc-classifier.md` does not follow `gaze-` prefix), and document compliance findings. This validates SC-005.

**Checkpoint**: Contract defines OpenCode conventions.
Gaze compliance audit completed. SC-005 satisfied.

---

## Phase 7: User Story 5 - Hero Metadata and Discovery (Priority: P3)

**Goal**: Define the hero manifest file format, create a
JSON Schema for validation, and produce a sample manifest
for Gaze.

**Independent Test**: Validate the sample Gaze manifest
against the JSON Schema and verify all required fields are
present and correctly typed.

### Implementation for User Story 5

- [x] T027 [US5] Write the "Hero Manifest" section in `specs/002-hero-interface-contract/contract.md` — define the `.unbound-force/hero.json` file per FR-002 and data-model.md Hero Manifest table: all required fields (name, display_name, role, version, description, repository, parent_constitution_version, artifacts_produced, artifacts_consumed, opencode_agents, opencode_commands, dependencies) and optional fields (mcp_server). Define the sub-entity schemas (Artifact Reference, Agent Reference, Command Reference, Dependency, MCP Reference) inline.
- [x] T028 [US5] Create the hero manifest JSON Schema at `schemas/hero-manifest/v1.0.0.schema.json` — translate the manifest field definitions from T027 and data-model.md into a JSON Schema draft 2020-12 file. Include: required fields validation, `role` enum constraint (`tester`, `developer`, `reviewer`, `product-owner`, `manager`), semver string patterns, URL format for `repository`, array type constraints for `artifacts_produced`, `artifacts_consumed`, etc. Include `$id` and `$schema` fields.
- [x] T029 [US5] Create a sample Gaze hero manifest at `schemas/hero-manifest/samples/gaze-hero.json` — populate with Gaze's actual data: name "gaze", role "tester", current version, two artifacts_produced (quality-report, analysis-report with schema_version "1.0.0"), five agents (gaze-reporter, doc-classifier, reviewer-adversary, reviewer-architect, reviewer-guard), three commands (/gaze, /classify-docs, /review-council), no dependencies, parent_constitution_version "1.0.0"
- [x] T030 [US5] Validate the sample Gaze manifest (`schemas/hero-manifest/samples/gaze-hero.json`) against the JSON Schema (`schemas/hero-manifest/v1.0.0.schema.json`) — use available JSON Schema validation tooling (python3 with jsonschema, or node with ajv, or jq for basic syntax check). Document validation result as evidence for SC-006.
- [x] T031 [US5] Update the validation script `scripts/validate-hero-contract.sh` to add hero manifest schema validation — extend the script to validate `.unbound-force/hero.json` against the manifest schema at `schemas/hero-manifest/v1.0.0.schema.json` when both files exist. Use best-available validator (python3 jsonschema, ajv, or fallback to required-field checking via jq).

**Checkpoint**: Contract defines hero manifest. JSON Schema
created and validated. Sample Gaze manifest validates. SC-006
satisfied.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, documentation updates, and
spec completion

- [x] T032 Review the complete Hero Interface Contract document at `specs/002-hero-interface-contract/contract.md` — verify all 15 FRs are addressed, all MUST/SHOULD/MAY language is consistent with RFC 2119, all cross-references to Specs 001/003/009 are accurate, and no section contradicts another
- [x] T033 Verify the contract is consistent with the org constitution at `.specify/memory/constitution.md` — confirm Principle I (artifact-based communication) is reflected in the artifact protocol sections, Principle II (composability) is reflected in the standalone requirement and optional dependencies, Principle III (observable quality) is reflected in JSON output and provenance metadata requirements
- [x] T034 Update `AGENTS.md` to reflect spec 002 completion — add entry to Recent Changes documenting the Hero Interface Contract ratification, update the Active Technologies section if needed, ensure the Project Structure section reflects new `schemas/` and `scripts/` directories
- [x] T035 Verify `specs/002-hero-interface-contract/spec.md` status can be updated from Draft to Complete — confirm all SCs are satisfied: SC-001 (repo structure + validation script), SC-002 (Gaze validation), SC-003 (Website validation), SC-004 (sample artifact), SC-005 (agent naming audit), SC-006 (manifest schema + sample), SC-007 (speckit mechanism defined)
- [x] T036 Run `scripts/validate-hero-contract.sh` against the `unbound-force/unbound-force` meta repo itself — verify the meta repo's own compliance status (expected: partial compliance since meta repo is not a hero repo, but validation script should run without errors)
- [x] T037 Run quickstart.md validation — follow the quickstart guide steps and verify each step produces the expected outcome: bootstrapping workflow makes sense, validation output matches expected format, sample artifact is well-formed, naming conventions are clear

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup (T001-T003)
  — BLOCKS all user story phases
- **US1 (Phase 3)**: Depends on Foundational (T004-T006)
- **US2 (Phase 4)**: Depends on Foundational (T004-T006).
  Independent of US1 (different contract sections).
- **US3 (Phase 5)**: Depends on Foundational. Independent of
  US1/US2.
- **US4 (Phase 6)**: Depends on Foundational. Independent of
  US1/US2/US3.
- **US5 (Phase 7)**: Depends on Foundational. The validation
  script update (T031) depends on T011 (US1 validation script)
  and T028 (US5 schema).
- **Polish (Phase 8)**: Depends on all user stories complete.

### User Story Dependencies

- **US1 (P1)**: Independent after Foundational. Creates
  validation script that US5 extends.
- **US2 (P1)**: Independent after Foundational. Different
  contract section than US1.
- **US3 (P2)**: Independent after Foundational.
- **US4 (P2)**: Independent after Foundational.
- **US5 (P3)**: Mostly independent. T031 depends on T011
  (validation script exists) and T028 (schema exists).

### Within Each User Story

- Contract document sections can be written in sequence
- Validation/audit tasks depend on contract sections being
  written first
- Schema tasks (US5) must precede sample validation tasks

### Parallel Opportunities

- T001 and T002 can run in parallel (different directories)
- US1 contract sections (T007-T010) can run in parallel
  (different subsections of same document)
- US2 contract sections (T014-T017) can run in parallel
- US3 and US4 contract sections can run in parallel with
  each other (different contract areas)
- T029 and T030 are sequential (create then validate)

---

## Parallel Examples

### Setup Phase

```text
# These create different directories:
Task T001: Create schemas/hero-manifest/ directory
Task T002: Create scripts/ directory
```

### User Story 1 — Contract Subsections

```text
# These write different subsections of the contract:
Task T007: Repository Structure section
Task T008: Constitution Requirements subsection
Task T009: License Requirements subsection
Task T010: README Template subsection
```

### User Stories 3 & 4 (can run in parallel)

```text
# Different contract areas, no dependencies:
Task T019-T021: Speckit Integration (US3)
Task T022-T026: OpenCode Conventions (US4)
```

---

## Implementation Strategy

### MVP First (User Stories 1 & 2)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T006)
3. Complete Phase 3: US1 — Repository Structure + Validation
   Script (T007-T013)
4. Complete Phase 4: US2 — Artifact Protocol (T014-T018)
5. **STOP and VALIDATE**: The contract defines the two most
   critical areas (structure and communication). The validation
   script can check any repo.

### Incremental Delivery

1. Setup + Foundational -> contract skeleton ready
2. Add US1 (repo structure + validation) -> can validate repos
3. Add US2 (artifact protocol) -> heroes know how to communicate
4. Add US3 (speckit integration) -> drift elimination defined
5. Add US4 (OpenCode conventions) -> agent collisions prevented
6. Add US5 (hero manifest + schema) -> discovery enabled
7. Polish -> full cross-validation and documentation

### Notes

- All contract sections are written into a single document
  (`contract.md`) but are independently valuable
- The validation script (T011) is a key deliverable — test it
  against real repos (T012, T013) before proceeding
- The hero manifest schema (T028) is the only JSON Schema this
  spec produces — the envelope schema is deferred to Spec 009
