# Tasks: Shared Data Model

**Input**: Design documents from `/specs/009-shared-data-model/`

## Phase 1: Setup

- [x] T001 Run `go get github.com/invopop/jsonschema` to add the schema generation dependency.
- [x] T002 Create `internal/schemas/` package directory.
- [x] T003 Create schema registry directory structure: `schemas/{envelope,quality-report,review-verdict,backlog-item,acceptance-decision,metrics-snapshot,coaching-record,workflow-record,convention-pack}/` with `samples/` subdirectories.

## Phase 2: Foundational — Schema Generator + Validator

- [x] T004 Create `internal/schemas/generate.go` with `GenerateSchema(v interface{}) (*jsonschema.Schema, error)` using `invopop/jsonschema.Reflector`. Configure for draft 2020-12. Add `WriteSchema(schema *jsonschema.Schema, path string) error` to write formatted JSON.
- [x] T005 Create `internal/schemas/validate.go` with `ValidateArtifact(schemaPath, artifactPath string) error` that loads a JSON Schema file and validates a JSON artifact against it. Use `santhosh-tekuri/jsonschema` for validation (generation uses `invopop`, validation uses `santhosh-tekuri` — different concerns).
- [x] T006 Create `internal/schemas/packvalidator.go` with `ValidateConventionPack(packPath string) error` that reads a Markdown file, parses YAML frontmatter (requires `pack_id`, `language`, `version`), and checks for required H2 sections (Coding Style, Architectural Patterns, Security Checks, Testing Conventions, Documentation Requirements, Custom Rules).

## Phase 3: US1+US2 — Generate Schemas + Samples (Priority: P1) MVP

- [x] T007 [US1] Create `internal/schemas/types.go` with Go struct definitions for all artifact payloads that mirror the existing structs but with `jsonschema` tags for description/title metadata. Import and alias existing types where possible (e.g., `type MetricsSnapshotPayload = metrics.MetricsSnapshot`). Define `EnvelopeSchema` struct matching `artifacts.Envelope`.
- [x] T008 [US1] [US2] Create `internal/schemas/registry.go` with `GenerateAll(outputDir string) error` that generates all 8 schemas (envelope + 7 types) and writes them to `schemas/{type}/v1.0.0.schema.json`. Use `invopop/jsonschema.Reflector` on each payload struct.
- [x] T009 [P] [US2] Create sample artifacts: `schemas/envelope/samples/sample-envelope.json`, one sample per artifact type (7 samples total). Each sample must be a valid JSON artifact that conforms to the envelope + its type schema. Use realistic data from existing tests.
- [x] T010 [P] [US2] Create README.md files for each schema directory (8 READMEs): describe the artifact type, its producer hero, consumer heroes, required fields, and version history.
- [x] T011 [US1] [US2] Create `internal/schemas/generate_test.go` with `TestGenerateAll_ProducesValidSchemas` (generate all schemas to t.TempDir, verify 8 files exist), `TestValidateArtifact_AllSamples` (validate each sample against its schema — SC-001, SC-002), `TestValidateArtifact_InvalidArtifact` (missing required field produces error — US1-AS2).

## Phase 4: US3 — Versioning + Compatibility (Priority: P2)

- [x] T012 [US3] Create `internal/schemas/version.go` with `CheckCompatibility(producerVersion, consumerVersion string) (compatible bool, err error)` — returns compatible=true if MAJOR versions match (MINOR/PATCH differences are backward compatible per FR-006). Returns error with migration guidance if MAJOR differs (FR-007).
- [x] T013 [US3] Create `internal/schemas/version_test.go` with `TestCheckCompatibility_SameVersion`, `TestCheckCompatibility_MinorBump_Compatible` (SC-003), `TestCheckCompatibility_MajorBump_Incompatible` (SC-004), `TestCheckCompatibility_PatchBump_Compatible`.

## Phase 5: US4 — Schema Registry CI (Priority: P2)

- [x] T014 [US4] Create `internal/schemas/ci_test.go` with `TestSchemaRegistry_AllSchemasValid` (load every .schema.json file under `schemas/`, verify each is valid JSON Schema draft 2020-12 — SC-006), `TestSchemaRegistry_AllSamplesValidate` (validate every sample artifact against its corresponding schema — SC-006), `TestSchemaRegistry_DirectoryStructure` (verify expected directory structure exists — SC-005).

## Phase 6: US5 — Convention Pack Validation (Priority: P3)

- [x] T015 [US5] Create `internal/schemas/packvalidator_test.go` with `TestValidateConventionPack_GoPackValid` (validate `.opencode/unbound/packs/go.md` passes — SC-007), `TestValidateConventionPack_MissingSection` (pack without Coding Style section fails), `TestValidateConventionPack_MissingFrontmatter` (pack without pack_id fails).

## Phase 7: Polish

- [x] T016 Run `go test -race -count=1 ./...` and verify all tests pass.
- [x] T017 Run `go build ./...` and verify build succeeds.
- [x] T018 [P] Update `AGENTS.md`: add `internal/schemas/` to project structure, add Spec 009 to Recent Changes, update schema registry structure.
- [x] T019 [P] Update `specs/009-shared-data-model/spec.md`: change status to complete.
- [x] T020 Verify SC-001 through SC-008 success criteria.
<!-- scaffolded by unbound vdev -->
