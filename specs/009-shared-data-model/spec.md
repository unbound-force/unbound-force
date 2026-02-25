# Feature Specification: Shared Data Model

**Feature Branch**: `009-shared-data-model`
**Created**: 2026-02-24
**Status**: Draft
**Input**: User description: "Define the shared data structures and JSON schemas used for inter-hero communication across the Unbound Force swarm. This includes all artifact type schemas, versioning strategy, backward compatibility rules, and the event/notification model."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Artifact Envelope Schema (Priority: P1)

The shared data model defines the artifact envelope — the standard JSON wrapper that every inter-hero artifact conforms to. The envelope provides metadata (producer, version, timestamp, type) that enables any consumer to identify, route, and version-check an artifact without understanding its payload.

**Why this priority**: P1 because the envelope is the foundation of all inter-hero communication. Every artifact type (quality-report, review-verdict, backlog-item, etc.) is wrapped in this envelope.

**Independent Test**: Can be tested by creating sample artifacts for each type, wrapping them in the envelope, and validating each against the JSON Schema.

**Acceptance Scenarios**:

1. **Given** the artifact envelope JSON Schema, **When** a valid artifact is validated against it, **Then** validation passes with zero errors.
2. **Given** an artifact missing the `hero` field, **When** validated, **Then** the schema reports a required property error.
3. **Given** an artifact with an unregistered `artifact_type`, **When** validated against the envelope schema alone, **Then** validation passes (the envelope does not restrict payload types — type-specific schemas validate the payload).
4. **Given** the envelope schema, **When** a developer inspects it, **Then** it defines these required fields: `hero` (string), `version` (semver string), `timestamp` (ISO 8601 string), `artifact_type` (string), `schema_version` (semver string), `context` (object with `branch` and `backlog_item_id`), and `payload` (object, type-specific).

---

### User Story 2 - Artifact Type Schemas (Priority: P1)

The shared data model defines JSON Schemas for each registered artifact type. Each schema validates the `payload` object within the artifact envelope for its specific type. The initial artifact types are: `quality-report` (Gaze), `review-verdict` (The Divisor), `backlog-item` (Muti-Mind), `acceptance-decision` (Muti-Mind), `metrics-snapshot` (Mx F), `coaching-record` (Mx F), and `workflow-record` (Swarm Orchestration).

**Why this priority**: P1 because without type-specific schemas, consumers cannot reliably parse payloads. Schemas are the contract between producers and consumers.

**Independent Test**: Can be tested by producing a sample artifact for each type from its producing hero, validating the payload against the type schema, and verifying a consuming hero can parse all required fields.

**Acceptance Scenarios**:

1. **Given** the `quality-report` schema, **When** Gaze produces a quality report payload, **Then** it validates against the schema and contains at minimum: `summary` (overall scores), `functions[]` (per-function metrics), `coverage` (aggregate coverage data), and `recommendations[]`.
2. **Given** the `review-verdict` schema, **When** The Divisor produces a review verdict payload, **Then** it validates and contains: `persona_verdicts[]` (each with persona, verdict, findings[]), `council_decision`, `iteration_count`, and `pr_url`.
3. **Given** the `backlog-item` schema, **When** Muti-Mind produces a backlog item payload, **Then** it validates and contains: `id`, `title`, `type`, `priority`, `status`, `acceptance_criteria[]`, `sprint`, and `effort_estimate`.
4. **Given** the `acceptance-decision` schema, **When** Muti-Mind produces an acceptance decision payload, **Then** it validates and contains: `item_id`, `decision`, `rationale`, `criteria_met[]`, `criteria_failed[]`, and `report_ref`.
5. **Given** the `metrics-snapshot` schema, **When** Mx F produces a metrics snapshot payload, **Then** it validates and contains: `velocity`, `cycle_time`, `lead_time`, `defect_rate`, `review_iterations`, `ci_pass_rate`, `backlog_health`, and `health_indicators[]`.
6. **Given** the `coaching-record` schema, **When** Mx F produces a coaching record payload, **Then** it validates and contains: `retrospective` (date, participants, patterns, action_items) or `coaching_interaction` (topic, questions, insights, outcome).
7. **Given** the `workflow-record` schema, **When** the swarm orchestration produces a workflow record payload, **Then** it validates and contains: `workflow_id`, `backlog_item_id`, `stages[]`, `artifacts[]`, `decisions[]`, `total_elapsed_time`, and `outcome`.

---

### User Story 3 - Schema Versioning and Backward Compatibility (Priority: P2)

The shared data model defines a versioning strategy for schemas. Schemas use semantic versioning (MAJOR.MINOR.PATCH). Minor and patch updates are backward compatible. Major updates may break compatibility. Consumers MUST handle artifacts with a schema version they understand and gracefully handle artifacts with incompatible versions.

**Why this priority**: P2 because version incompatibilities will inevitably arise as heroes evolve independently. Without a versioning strategy, schema changes break the swarm.

**Independent Test**: Can be tested by creating a v1 artifact, bumping the schema to v2 with a new optional field, validating a v1 consumer can still parse the v2 artifact, and then bumping to v3 with a removed required field and verifying the v1 consumer rejects it.

**Acceptance Scenarios**:

1. **Given** a `quality-report` v1.0.0 artifact, **When** the schema is updated to v1.1.0 (new optional field `coverage_trend`), **Then** a consumer expecting v1.0.0 can still parse the artifact (the new field is ignored).
2. **Given** a `quality-report` v1.0.0 artifact, **When** a consumer expects v2.0.0 (which renamed `functions[]` to `analysis_results[]`), **Then** the consumer detects the version mismatch and either applies a migration or reports "incompatible schema version."
3. **Given** the versioning strategy, **When** a hero wants to add a required field to an artifact type, **Then** the strategy requires a MAJOR version bump and a migration guide for consumers.
4. **Given** the versioning strategy, **When** a hero wants to add an optional field, **Then** a MINOR version bump suffices and no consumer changes are needed.

---

### User Story 4 - Schema Registry and Documentation (Priority: P2)

The shared data model is published as a schema registry — a collection of JSON Schema files organized by artifact type and version, with human-readable documentation for each schema. The registry lives in the canonical speckit or unbound-force repository and is the single source of truth for all artifact type definitions.

**Why this priority**: P2 because a centralized registry prevents heroes from defining incompatible schemas independently. It is the single source of truth for inter-hero communication contracts.

**Independent Test**: Can be tested by cloning the registry, running a schema validation tool against all sample artifacts, and verifying 100% pass rate.

**Acceptance Scenarios**:

1. **Given** the schema registry, **When** a developer browses it, **Then** they find: one directory per artifact type, each containing versioned schema files (e.g., `quality-report/v1.0.0.schema.json`) and a `README.md` with human-readable documentation.
2. **Given** the registry, **When** a hero team wants to register a new artifact type, **Then** they create a new directory with the schema file and README, submit a PR, and it is reviewed against the artifact envelope compatibility rules.
3. **Given** the registry, **When** a CI pipeline runs, **Then** it validates all schema files are syntactically valid JSON Schema (draft 2020-12) and all sample artifacts validate against their respective schemas.
4. **Given** the registry, **When** a developer generates code from a schema, **Then** the schema is machine-readable and compatible with standard JSON Schema code generators (for Go, TypeScript, Python, etc.).

---

### User Story 5 - Convention Pack Schema (Priority: P3)

The shared data model defines the schema for convention packs — the pluggable coding convention configurations shared between Cobalt-Crush (developer) and The Divisor (reviewer). This schema ensures convention packs are interchangeable and machine-readable.

**Why this priority**: P3 because convention packs are a secondary data model (not inter-hero artifacts) but critical for developer-reviewer alignment. Standardizing their schema prevents drift between the two heroes.

**Independent Test**: Can be tested by creating a Go convention pack and a TypeScript convention pack, validating both against the schema, and verifying both Cobalt-Crush and The Divisor can parse them.

**Acceptance Scenarios**:

1. **Given** the convention pack schema, **When** the Go convention pack is validated, **Then** it passes and contains all required sections: `coding_style`, `architectural_patterns`, `security_checks`, `testing_conventions`, `documentation_requirements`.
2. **Given** the convention pack schema, **When** a developer creates a new pack for Python, **Then** the schema guides them to fill in all required sections and validates the result.
3. **Given** a convention pack consumed by both Cobalt-Crush and The Divisor, **When** the pack is updated, **Then** both heroes pick up the same update (from the single source of truth).

---

### Edge Cases

- What happens when a schema file is syntactically invalid JSON? The CI pipeline MUST catch this and prevent merging. Schema files MUST be validated before registration.
- What happens when two heroes define the same artifact type name? The registry MUST enforce unique artifact type names. Attempting to register a duplicate MUST be rejected.
- What happens when a consumer receives an artifact with a `schema_version` it has never seen? The consumer MUST check if the MAJOR version matches. If MAJOR matches, proceed (minor/patch are backward compatible). If MAJOR differs, reject with a version mismatch warning.
- What happens when the artifact envelope schema itself is updated? Envelope schema changes follow the same versioning rules. All consumers MUST handle the envelope version in addition to the payload schema version.
- What happens when a convention pack has optional sections that a consumer expects to be present? The convention pack schema MUST clearly mark required vs. optional sections. Consumers MUST provide defaults for optional sections.
- What happens when a hero produces an artifact that does not validate against the registered schema? This is a producer bug. Consuming heroes SHOULD reject the artifact with a validation error rather than attempting to parse invalid data.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The shared data model MUST define a JSON Schema (draft 2020-12) for the artifact envelope with required fields: `hero`, `version`, `timestamp`, `artifact_type`, `schema_version`, `context`, and `payload`.
- **FR-002**: The `context` field MUST contain at minimum: `branch` (string, git branch name) and `backlog_item_id` (string, optional, the originating backlog item).
- **FR-003**: The shared data model MUST define JSON Schemas for all registered artifact types: `quality-report`, `review-verdict`, `backlog-item`, `acceptance-decision`, `metrics-snapshot`, `coaching-record`, `workflow-record`.
- **FR-004**: Each artifact type schema MUST validate the `payload` field of the artifact envelope for that type.
- **FR-005**: All schemas MUST use semantic versioning (MAJOR.MINOR.PATCH) where: MAJOR = breaking change, MINOR = backward-compatible addition, PATCH = documentation or non-functional fix.
- **FR-006**: Minor version updates MUST be backward compatible: a consumer expecting v1.0.0 MUST successfully parse a v1.1.0 artifact (new optional fields are ignored).
- **FR-007**: Major version updates MAY break compatibility. A migration guide MUST be provided for each major version bump.
- **FR-008**: The schema registry MUST be organized as: `schemas/{artifact_type}/v{MAJOR}.{MINOR}.{PATCH}.schema.json` with a `README.md` per type.
- **FR-009**: The schema registry MUST include sample artifacts for each type (`schemas/{artifact_type}/samples/`) that validate against the schema.
- **FR-010**: CI MUST validate all schemas are syntactically valid JSON Schema and all samples validate against their respective schemas.
- **FR-011**: The shared data model MUST define the convention pack schema with required sections: `pack_id`, `language`, `coding_style`, `architectural_patterns`, `security_checks`, `testing_conventions`, `documentation_requirements`, and optional `custom_rules[]` and `framework` fields.
- **FR-012**: The convention pack schema MUST be shared between Cobalt-Crush and The Divisor (single source of truth, not duplicated).
- **FR-013**: Artifact type names MUST be unique across the registry. Duplicate registrations MUST be rejected.
- **FR-014**: The envelope MUST include an optional `correlation_id` field (UUID) for linking related artifacts across workflow stages.
- **FR-015**: The shared data model SHOULD define an event model for hero-to-hero notifications (e.g., "quality-report available for PR #42") as an optional enhancement over polling-based artifact discovery.

### Key Entities

- **Artifact Envelope**: Standard JSON wrapper. Fields: `hero` (string), `version` (semver), `timestamp` (ISO 8601), `artifact_type` (string), `schema_version` (semver), `context` ({branch, backlog_item_id, correlation_id}), `payload` (object).
- **Schema Registry Entry**: One artifact type's schema collection. Attributes: artifact_type, current_version, versions[] (each with schema file, changelog, migration_guide if major), samples[], producing_heroes[], consuming_heroes[].
- **Quality Report Payload**: Gaze's quality output. Fields: summary (crap_load, avg_crap, avg_coverage, total_functions), functions[] (name, crap_score, complexity, coverage, contract_coverage, classification), coverage (aggregate stats), recommendations[] (priority, description, target).
- **Review Verdict Payload**: The Divisor's review output. Fields: persona_verdicts[] (persona, verdict, findings[], summary), council_decision (APPROVED/CHANGES_REQUESTED/ESCALATED), iteration_count, unresolved_findings[], pr_url, convention_pack_used.
- **Backlog Item Payload**: Muti-Mind's backlog output. Fields: id, title, description, type, priority, status, acceptance_criteria[] (given, when, then), sprint, effort_estimate, dependencies[], related_specs[].
- **Acceptance Decision Payload**: Muti-Mind's acceptance output. Fields: item_id, decision (accept/reject/conditional), rationale, criteria_met[], criteria_failed[], report_ref.
- **Metrics Snapshot Payload**: Mx F's metrics output. Fields: velocity, cycle_time (avg, median, p90, p99), lead_time, defect_rate, review_iterations, ci_pass_rate, backlog_health (total, ready, stale), health_indicators[] (dimension, status, value, trend).
- **Coaching Record Payload**: Mx F's coaching output. Fields: record_type (retrospective/coaching), retrospective (date, patterns[], root_causes[], action_items[]), coaching_interaction (topic, questions[], insights[], outcome).
- **Workflow Record Payload**: Orchestration lifecycle output. Fields: workflow_id, backlog_item_id, stages[] (name, hero, status, started_at, completed_at, artifacts[]), total_elapsed_time, outcome.
- **Convention Pack**: Shared developer/reviewer config. Fields: pack_id, language, framework, coding_style{}, architectural_patterns{}, security_checks{}, testing_conventions{}, documentation_requirements{}, custom_rules[].

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The artifact envelope JSON Schema is valid JSON Schema (draft 2020-12) and validates 7 sample artifacts (one per type) without errors.
- **SC-002**: All 7 artifact type schemas are valid JSON Schema and each validates at least one sample artifact.
- **SC-003**: Backward compatibility is demonstrated: a v1.0.0 consumer parses a v1.1.0 artifact successfully (optional field ignored).
- **SC-004**: Forward incompatibility is demonstrated: a v1.0.0 consumer rejects a v2.0.0 artifact with a clear version mismatch message.
- **SC-005**: The schema registry directory structure is created with all types, schemas, samples, and READMEs.
- **SC-006**: CI validation runs against all schemas and samples with 100% pass rate.
- **SC-007**: The convention pack schema validates the Go convention pack and is parseable by both a hypothetical Cobalt-Crush consumer and a hypothetical Divisor consumer.
- **SC-008**: A developer can generate Go and TypeScript type definitions from any schema using standard JSON Schema code generation tools.

## Dependencies

### Prerequisites

- **Spec 002** (Hero Interface Contract): Defines the artifact envelope concept that this spec provides the schema for.
- **Specs 004-007** (Hero Architectures): Define the specific artifact types each hero produces/consumes.
- **Spec 008** (Swarm Orchestration): Defines the `workflow-record` artifact type and the correlation/context model.

### Downstream Dependents

- All hero implementations reference these schemas when producing or consuming artifacts.

```
Schema Registry Structure

schemas/
├── envelope/
│   ├── v1.0.0.schema.json      (artifact envelope)
│   ├── samples/
│   │   └── sample-envelope.json
│   └── README.md
├── quality-report/
│   ├── v1.0.0.schema.json      (Gaze payload)
│   ├── samples/
│   │   └── sample-quality-report.json
│   └── README.md
├── review-verdict/
│   ├── v1.0.0.schema.json      (Divisor payload)
│   ├── samples/
│   │   └── sample-review-verdict.json
│   └── README.md
├── backlog-item/
│   ├── v1.0.0.schema.json      (Muti-Mind payload)
│   ├── samples/
│   │   └── sample-backlog-item.json
│   └── README.md
├── acceptance-decision/
│   ├── v1.0.0.schema.json      (Muti-Mind payload)
│   ├── samples/
│   │   └── sample-acceptance-decision.json
│   └── README.md
├── metrics-snapshot/
│   ├── v1.0.0.schema.json      (Mx F payload)
│   ├── samples/
│   │   └── sample-metrics-snapshot.json
│   └── README.md
├── coaching-record/
│   ├── v1.0.0.schema.json      (Mx F payload)
│   ├── samples/
│   │   └── sample-coaching-record.json
│   └── README.md
├── workflow-record/
│   ├── v1.0.0.schema.json      (Orchestration payload)
│   ├── samples/
│   │   └── sample-workflow-record.json
│   └── README.md
└── convention-pack/
    ├── v1.0.0.schema.json      (Shared: Cobalt-Crush + Divisor)
    ├── packs/
    │   ├── go.yaml
    │   ├── typescript.yaml
    │   └── default.yaml
    ├── samples/
    │   └── sample-convention-pack.json
    └── README.md
```
