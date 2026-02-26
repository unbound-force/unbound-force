# Feature Specification: Hero Interface Contract

**Feature Branch**: `002-hero-interface-contract`
**Created**: 2026-02-24
**Status**: Complete
**Input**: User description: "Define the standard structure, contracts, and integration points that every hero repository must implement to be a member of the Unbound Force swarm. This is the 'plug shape' that ensures interoperability."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Standard Repository Structure (Priority: P1)

A maintainer bootstrapping a new hero repository follows the Hero Interface Contract to create a repository with the correct directory structure, configuration files, and metadata. The contract defines the minimum required files and directories that make a repository a valid Unbound Force hero.

**Why this priority**: P1 because without a standard structure, heroes cannot be composed, discovered, or validated by tooling or other heroes.

**Independent Test**: Can be tested by creating a new empty repository and running a validation script that checks for the presence of all required files and directories defined by the contract.

**Acceptance Scenarios**:

1. **Given** the Hero Interface Contract, **When** a maintainer creates a new hero repository, **Then** the contract specifies exactly which files and directories are required vs. optional.
2. **Given** a new hero repository following the contract, **When** a validation tool inspects it, **Then** it finds: `.specify/memory/constitution.md` (ratified), `.specify/templates/` (populated), `.opencode/` (configured), `specs/` (directory exists), `AGENTS.md` (populated), `LICENSE` (Apache 2.0), and `README.md`.
3. **Given** an existing hero repository (Gaze), **When** the contract is applied retroactively, **Then** Gaze already satisfies all required structure elements (or gaps are identified for remediation).
4. **Given** a repository missing a required element, **When** validation runs, **Then** it reports exactly which elements are missing and provides remediation guidance.

---

### User Story 2 - Inter-Hero Artifact Protocol (Priority: P1)

Heroes exchange information through well-defined artifacts (files, reports, schemas) rather than runtime coupling. The contract defines the standard artifact types, their formats, and the conventions for producing and consuming them. Any hero that produces an artifact type MUST conform to the schema; any hero that consumes an artifact type MUST handle valid instances of that schema.

**Why this priority**: P1 because inter-hero communication is the foundation of the swarm. Without a defined protocol, heroes cannot collaborate.

**Independent Test**: Can be tested by producing a sample artifact from one hero (e.g., a Gaze quality report) and verifying another hero (e.g., Mx F metrics collection) can parse it according to the contract schema.

**Acceptance Scenarios**:

1. **Given** the artifact protocol, **When** a hero produces a report, **Then** the report conforms to a registered JSON schema that includes: `hero` (producer name), `version` (producer version), `timestamp` (ISO 8601), `artifact_type` (registered type identifier), and `payload` (type-specific content).
2. **Given** a registered artifact type "quality-report", **When** Gaze produces a quality report, **Then** any hero configured to consume "quality-report" artifacts can parse the payload without knowledge of Gaze's internals.
3. **Given** a hero that does not recognize an artifact type, **When** it encounters that artifact, **Then** it MUST ignore it gracefully (no errors, optional warning log).
4. **Given** an artifact schema version bump (e.g., v1 to v2), **When** a consumer still expects v1, **Then** the artifact envelope includes a `schema_version` field enabling the consumer to detect and handle the mismatch.

---

### User Story 3 - Speckit Framework Integration (Priority: P2)

Every hero repository uses the speckit framework (`.specify/` templates and scripts, `.opencode/command/speckit.*.md` commands) for specification-driven development. The contract defines how the shared speckit framework is installed, configured, and kept in sync across hero repositories.

**Why this priority**: P2 because speckit is already deployed (with drift) across Gaze, Website, and unbound-force repos. Formalizing its integration prevents further drift and establishes the upgrade path.

**Independent Test**: Can be tested by comparing the speckit files across the three existing repositories and verifying the contract identifies all drift points and defines the canonical source.

**Acceptance Scenarios**:

1. **Given** the contract's speckit integration requirements, **When** a hero repository is bootstrapped, **Then** the speckit templates and scripts are installed from a single canonical source (not copy-pasted from another repo).
2. **Given** a speckit update is released, **When** a hero repository updates, **Then** the update process preserves any hero-specific customizations (extension points) while updating the canonical files.
3. **Given** the existing drift between Gaze and unbound-force `speckit.specify.md` files (integration pattern language differs), **When** the contract is applied, **Then** the contract defines which version is canonical and how project-specific variations are handled (via configuration, not file modification).

---

### User Story 4 - OpenCode Plugin and Agent Standards (Priority: P2)

Heroes that provide OpenCode agents or commands follow a standard convention for naming, discovery, configuration, and documentation. The contract defines the agent file format, naming conventions, tool permissions, and model selection guidelines.

**Why this priority**: P2 because multiple heroes will provide OpenCode agents (The Divisor's review council, Gaze's reporter, Muti-Mind's PO agent, etc.) and consistency is critical for the swarm plugin to orchestrate them.

**Independent Test**: Can be tested by inspecting the existing Gaze OpenCode agents against the proposed standard and verifying compliance or identifying gaps.

**Acceptance Scenarios**:

1. **Given** the agent standard, **When** a hero provides an OpenCode agent, **Then** the agent file includes: a descriptive header, explicit tool permissions (`read`, `edit`, `bash`, etc.), model specification, and behavioral constraints section.
2. **Given** the naming convention, **When** agents are installed across multiple hero repos, **Then** agent names are globally unique within a project's `.opencode/agents/` directory (prefixed by hero name, e.g., `gaze-reporter`, `divisor-guard`, `muti-mind-po`).
3. **Given** the command standard, **When** a hero provides an OpenCode command, **Then** the command file defines: trigger syntax, argument parsing, agent delegation, and error handling.
4. **Given** a project using multiple heroes, **When** all heroes' `init` commands have been run, **Then** no agent or command files collide (unique names, no overwrites without `--force`).

---

### User Story 5 - Hero Metadata and Discovery (Priority: P3)

Each hero repository publishes metadata that enables discovery and composition. The contract defines a hero manifest file that describes the hero's role, capabilities, artifact types produced/consumed, and integration points. This manifest enables tools (like the Swarm plugin or Mx F's metrics platform) to discover and configure heroes automatically.

**Why this priority**: P3 because discovery and auto-configuration are valuable but not required for initial hero development. Heroes can be manually configured initially.

**Independent Test**: Can be tested by creating a sample manifest for Gaze and verifying a hypothetical discovery tool can parse it and understand Gaze's capabilities.

**Acceptance Scenarios**:

1. **Given** the manifest standard, **When** a hero repository includes a `.unbound-force/hero.json` manifest, **Then** the manifest contains: `name`, `role` (tester/developer/reviewer/product-owner/manager), `version`, `description`, `artifacts_produced[]`, `artifacts_consumed[]`, `opencode_agents[]`, `opencode_commands[]`, and `dependencies[]`.
2. **Given** manifests for all five heroes, **When** a composition tool reads them, **Then** it can construct a dependency graph showing which heroes produce artifacts consumed by other heroes.
3. **Given** a manifest with `artifacts_consumed: ["quality-report"]`, **When** no installed hero produces "quality-report", **Then** the discovery tool warns that a dependency is unmet (but does not prevent the hero from functioning standalone per Principle II).

---

### Edge Cases

- What happens when two heroes provide agents with the same name? The contract MUST require hero-prefixed agent names to prevent collisions. If a collision still occurs (e.g., two heroes claim the same prefix), the last `init` run wins unless `--force` is used, and a warning is emitted.
- What happens when a hero repository predates the contract? Existing repos (Gaze, Website) MUST be updated to comply. A compliance checklist SHOULD be provided for migration.
- What happens when a hero needs to extend the artifact protocol with a new type? The hero MUST register the new artifact type in the shared data model (Spec 009) and provide a JSON schema for validation.
- What happens when the speckit framework is updated but a hero has uncommitted local changes to speckit files? The update tool MUST detect modifications and refuse to overwrite without `--force`, listing the conflicting files.
- What happens when a hero's constitution contradicts the org constitution? The hero MUST NOT be considered a valid swarm member until the contradiction is resolved. The validation tool MUST flag this as a blocking error.
- What happens when a hero wants to use a different license than Apache 2.0? The contract SHOULD recommend Apache 2.0 but MAY allow alternatives with explicit justification documented in the hero's constitution.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The contract MUST define a minimum required directory structure for every hero repository: `.specify/memory/constitution.md`, `.specify/templates/`, `.specify/scripts/`, `.opencode/`, `specs/`, `AGENTS.md`, `LICENSE`, `README.md`.
- **FR-002**: The contract MUST define a hero manifest file (`.unbound-force/hero.json`) that describes the hero's role, capabilities, and integration points.
- **FR-003**: The contract MUST define the artifact envelope format conceptually: a JSON object with `hero`, `version`, `timestamp`, `artifact_type`, `schema_version`, `context`, and `payload` fields. The formal JSON Schema for the envelope is deferred to Spec 009 (Shared Data Model).
- **FR-004**: The contract MUST define a registry of standard artifact types: `quality-report`, `review-verdict`, `backlog-item`, `acceptance-decision`, `metrics-snapshot`, `coaching-record`, `workflow-record`. JSON Schemas for each type are deferred to Spec 009.
- **FR-005**: The contract MUST define OpenCode agent naming conventions: `{hero-name}-{agent-function}` (e.g., `gaze-reporter`, `divisor-guard`).
- **FR-006**: The contract MUST define OpenCode command naming conventions: `/{hero-name}` for the primary command, `/{hero-name}-{subfunction}` for secondary commands.
- **FR-007**: The contract MUST require that hero repositories install speckit from the canonical source defined by Spec 003 (Speckit Framework). The distribution mechanism itself is defined by Spec 003.
- **FR-008**: The contract MUST define a validation process — a bash script that verifies a repository conforms to the Hero Interface Contract by checking directory structure, required files, constitution presence, and manifest validity.
- **FR-009**: The contract MUST require that hero constitutions include a `parent_constitution` field referencing the org constitution version they align with.
- **FR-010**: The contract MUST require that heroes producing machine-parseable output support at minimum JSON format, with human-readable format as a SHOULD.
- **FR-011**: The contract MUST define a hero lifecycle: bootstrap -> constitution -> specify -> implement -> deploy -> maintain.
- **FR-012**: The contract SHOULD define an `init` command convention: `{hero-name} init` scaffolds the hero's OpenCode integration files into the target project.
- **FR-013**: The contract MUST define how heroes handle version incompatibilities when consuming artifacts from other heroes (graceful degradation, not hard failure).
- **FR-014**: The contract SHOULD define minimal MCP server naming conventions for heroes that expose capabilities via MCP: `{hero-name}-mcp` for the server name, standardized tool naming with hero prefix. Full MCP interface requirements are deferred to hero-specific architecture specs.
- **FR-015**: The contract SHOULD define a hero README template that includes: hero name, role, installation, quick start, integration with other heroes, and link to the org README.

### Key Entities

- **Hero Manifest**: Machine-readable description of a hero's capabilities and integration points. Attributes: name, role, version, description, artifacts_produced[], artifacts_consumed[], opencode_agents[], opencode_commands[], dependencies[], parent_constitution_version.
- **Artifact Envelope**: Standard wrapper for all inter-hero artifacts. Attributes: hero (string), version (semver string), timestamp (ISO 8601), artifact_type (registered type), schema_version (semver string), context (object with branch, backlog_item_id, correlation_id), payload (type-specific JSON object). Conceptual definition in this spec; JSON Schema in Spec 009.
- **Artifact Type Registration**: Entry in the shared artifact type registry. Attributes: type_id (string), description, producing_heroes[], consuming_heroes[]. JSON Schemas for each type defined in Spec 009.
- **Validation Result**: Output of a contract compliance check. Attributes: hero_name, contract_version, required_checks[], optional_checks[], findings[], overall_status (pass/fail).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The contract document defines the complete minimum required directory structure and a validation script can check any repository against it.
- **SC-002**: The existing Gaze repository passes the contract validation with zero missing required elements (or a clear remediation list is produced).
- **SC-003**: The existing Website repository passes the contract validation with zero missing required elements (or a clear remediation list is produced).
- **SC-004**: A sample artifact produced by Gaze (quality report JSON) validates against the artifact envelope format defined by the contract.
- **SC-005**: Agent and command naming conventions are documented and the existing Gaze agents comply (or deviations are identified).
- **SC-006**: The hero manifest schema is defined as a JSON Schema and a sample manifest for Gaze validates against it.
- **SC-007**: The speckit distribution mechanism is defined and eliminates the known drift between Gaze and unbound-force `speckit.specify.md` files.

## Dependencies

### Prerequisites

- **Spec 001** (Org Constitution): The contract references org constitution principles for its rules.

### Downstream Dependents

- **Spec 003** (Speckit Framework): Implements the speckit distribution mechanism referenced by FR-007.
- **Specs 004-007** (Hero Architectures): Each hero architecture must conform to the contract.
- **Spec 008** (Swarm Orchestration): Uses the artifact protocol and discovery mechanism.
- **Spec 009** (Shared Data Model): Defines the JSON schemas referenced by the artifact type registry.
