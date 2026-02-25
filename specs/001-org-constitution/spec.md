# Feature Specification: Unbound Force Organization Constitution

**Feature Branch**: `001-org-constitution`
**Created**: 2026-02-24
**Status**: Draft
**Input**: User description: "Establish the overarching constitution for the Unbound Force organization that governs all hero repositories. Define the three core principles (Autonomous Collaboration, Composability First, Observable Quality) and the governance model that individual hero constitutions must align with."

## Clarifications

### Session 2026-02-25

- Q: Should the constitution define the Hero Interface Contract concept inline, or just reference Spec 002? A: Reference Spec 002 only. The constitution names the concept and defers full definition to Spec 002 to avoid duplication and maintain separation of concerns.
- Q: Should the spec include an explicit out-of-scope section? A: Yes. List what is explicitly deferred: artifact schemas (Spec 009), repo structure details (Spec 002), speckit pipeline mechanics (Spec 003), and hero-specific implementation details (Specs 004-007).
- Q: Should alignment checking be manual, automated, or agent-assisted? A: Agent-assisted. An OpenCode agent performs the alignment check using the constitution as context, producing a structured report.
- Q: Should the spec add an FR requiring principle conflicts are resolved via documented tradeoffs? A: Yes. Add a new FR requiring that principle conflicts are resolved via explicit documented tradeoffs, with no implicit priority between principles.

## Out of Scope

The following concerns are explicitly deferred to other specs and MUST NOT be defined in the constitution:

- **Artifact schemas and envelope formats**: Defined by Spec 009 (Shared Data Model). The constitution references the requirement for machine-parseable output but does not define the JSON schemas.
- **Repository structure and hero manifest**: Defined by Spec 002 (Hero Interface Contract). The constitution references the Hero Interface Contract concept but does not define its contents.
- **Speckit pipeline mechanics**: Defined by Spec 003 (Speckit Framework). The constitution recommends spec-driven development but does not define the pipeline stages or commands.
- **Hero-specific implementation details**: Defined by Specs 004-007. The constitution constrains hero behavior via principles but does not prescribe architecture, technology, or implementation approach.
- **Convention packs and coding standards**: Defined by Specs 005 and 006. The constitution does not govern language-specific coding conventions.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Constitution Ratification (Priority: P1)

A contributor or maintainer ratifies the Unbound Force organization constitution by defining the three core principles, governance rules, and alignment requirements. Once ratified, every hero repository constitution must demonstrate alignment with these org-level principles. The constitution becomes the highest-authority document in the organization.

**Why this priority**: P1 because every subsequent spec, hero repo, and design decision references the constitution. Nothing can be finalized without ratified principles.

**Independent Test**: Can be tested by reviewing the constitution document against the three stated principles, verifying each principle has MUST/SHOULD/MAY rules, and confirming the governance section defines amendment and alignment processes.

**Acceptance Scenarios**:

1. **Given** the constitution template exists in `.specify/memory/constitution.md`, **When** the constitution is ratified, **Then** all placeholder tokens (`[PROJECT_NAME]`, `[PRINCIPLE_*]`, etc.) are replaced with concrete content.
2. **Given** the ratified constitution, **When** a reviewer inspects Principle I (Autonomous Collaboration), **Then** it contains at least three MUST rules and at least one SHOULD rule governing how heroes operate independently while producing swarm-compatible artifacts.
3. **Given** the ratified constitution, **When** a reviewer inspects Principle II (Composability First), **Then** it contains at least three MUST rules defining that every hero is usable standalone or in combination without hard coupling.
4. **Given** the ratified constitution, **When** a reviewer inspects Principle III (Observable Quality), **Then** it contains at least three MUST rules requiring measurable, traceable, and auditable outputs from every hero.
5. **Given** the ratified constitution, **When** a reviewer inspects the Governance section, **Then** it defines: amendment process (PR-based), versioning (semantic), supremacy clause (constitution overrides all other documents), and alignment requirement (hero constitutions must not contradict org constitution).

---

### User Story 2 - Hero Constitution Alignment Validation (Priority: P2)

A maintainer adding a new hero repository or amending an existing hero constitution validates that the hero-level constitution aligns with the org constitution. The alignment check is performed by an OpenCode agent that compares the hero constitution against the org constitution, producing a structured report of findings. The agent verifies that no hero principle contradicts an org principle and that the hero constitution references its parent.

**Why this priority**: P2 because this is the enforcement mechanism that ensures the org constitution is not just aspirational. It depends on the org constitution being ratified first (US1).

**Independent Test**: Can be tested by running the alignment agent against the existing Gaze and Website constitutions and verifying the produced report shows zero contradictions.

**Acceptance Scenarios**:

1. **Given** a ratified org constitution and the Gaze constitution (Accuracy, Minimal Assumptions, Actionable Output), **When** the alignment agent checks alignment, **Then** the report shows no Gaze principle contradicts any org principle, and each Gaze principle can be mapped to at least one org principle it supports.
2. **Given** a ratified org constitution and the Website constitution (Content Accuracy, Minimal Footprint, Visitor Clarity), **When** the alignment agent checks alignment, **Then** the report shows no Website principle contradicts any org principle.
3. **Given** a new hero repository is bootstrapped, **When** its constitution is drafted, **Then** it MUST include a "Parent Constitution" reference to the org constitution version it aligns with.
4. **Given** an org constitution amendment, **When** the amendment changes a MUST rule, **Then** all hero constitutions MUST be reviewed for continued alignment within one release cycle.
5. **Given** a hero constitution submitted for alignment checking, **When** the OpenCode alignment agent runs, **Then** it produces a structured report containing: hero_name, org_constitution_version, findings[] (each with principle, status, rationale), and overall status (aligned/non-aligned).

---

### User Story 3 - Constitution-Aware Development Workflow (Priority: P3)

Developers and agents working within any Unbound Force repository can reference the org constitution to resolve ambiguity about cross-cutting concerns (inter-hero communication, output formats, standalone usability). The constitution provides clear, citable rules for design decisions that affect multiple heroes.

**Why this priority**: P3 because this is the day-to-day usage of the constitution. It becomes valuable once heroes are actively being developed (Phase 1+).

**Independent Test**: Can be tested by presenting a hypothetical cross-cutting design decision (e.g., "Should hero X require hero Y to be installed?") and verifying the constitution provides a clear, citable answer.

**Acceptance Scenarios**:

1. **Given** a design decision about whether Muti-Mind should require Gaze to be installed, **When** the developer consults Principle II (Composability First), **Then** the constitution provides a clear MUST rule that answers the question (heroes MUST be usable standalone).
2. **Given** a design decision about output format for a new hero, **When** the developer consults Principle III (Observable Quality), **Then** the constitution provides MUST rules about measurable and machine-parseable outputs.
3. **Given** a design decision about how two heroes communicate, **When** the developer consults Principle I (Autonomous Collaboration), **Then** the constitution provides MUST rules about artifact-based communication (no tight runtime coupling).

---

### Edge Cases

- What happens when an existing hero constitution predates the org constitution? The hero constitution MUST be reviewed for alignment but is not automatically invalidated. Contradictions MUST be resolved by amending the hero constitution.
- What happens when two org principles appear to conflict in a specific scenario? The conflict MUST be resolved via explicit documented tradeoffs in the relevant spec or plan. No principle has implicit priority over another.
- What happens when a hero needs a principle that the org constitution does not address? The hero MAY add principles beyond the org constitution's scope, provided they do not contradict any org principle.
- What happens when the org constitution is amended and a hero constitution becomes non-aligned? The hero repository MUST open an alignment issue within one release cycle and resolve it before the next major version.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The org constitution MUST define exactly three core principles: Autonomous Collaboration, Composability First, and Observable Quality.
- **FR-002**: Each principle MUST contain at least three MUST rules and at least one SHOULD rule.
- **FR-003**: The constitution MUST include a Governance section defining: amendment process, versioning scheme, supremacy clause, and hero alignment requirements.
- **FR-004**: The constitution MUST use semantic versioning (MAJOR.MINOR.PATCH) where MAJOR indicates breaking changes to MUST rules.
- **FR-005**: The constitution MUST reference the "Hero Interface Contract" concept — the minimum requirements every hero repository must satisfy to be a member of the Unbound Force swarm. The full definition of the Hero Interface Contract is deferred to Spec 002.
- **FR-006**: The Governance section MUST state that the org constitution is the highest-authority document in the organization, superseding all hero-level constitutions on matters of cross-cutting concern.
- **FR-007**: The constitution MUST define the relationship between the org constitution and hero constitutions: hero constitutions extend (not contradict) the org constitution.
- **FR-008**: The constitution SHOULD define a "Constitution Check" process that can be applied during the speckit plan phase to validate compliance.
- **FR-009**: Principle I (Autonomous Collaboration) MUST require that heroes communicate through well-defined artifacts (files, reports, schemas) rather than runtime coupling.
- **FR-010**: Principle I MUST require that each hero can complete its primary function without requiring synchronous interaction with another hero.
- **FR-011**: Principle I MUST require that hero outputs are self-describing (contain enough metadata to be consumed without consulting the producing hero).
- **FR-012**: Principle II (Composability First) MUST require that every hero is independently installable and usable without any other hero being present.
- **FR-013**: Principle II MUST require that heroes expose well-defined extension points for integration rather than requiring modification of their internals.
- **FR-014**: Principle II MUST require that combining heroes produces additive value (the combination is more useful than the sum of parts) without introducing mandatory dependencies.
- **FR-015**: Principle III (Observable Quality) MUST require that every hero produces machine-parseable output (at minimum JSON) alongside any human-readable output.
- **FR-016**: Principle III MUST require that hero outputs include provenance metadata (which hero, which version, when, against what input).
- **FR-017**: Principle III MUST require that quality claims are backed by automated, reproducible evidence (tests, benchmarks, reports).
- **FR-018**: The constitution MUST include a "Development Workflow" section defining: feature branches required, code review mandatory (via The Divisor or equivalent), CI must pass, semantic versioning, conventional commits.
- **FR-019**: The Governance section MUST require that when two or more org principles appear to conflict in a specific scenario, the tradeoff MUST be explicitly documented in the relevant spec or plan. No principle has implicit priority over another; resolution is context-dependent and requires written justification.

### Key Entities

- **Org Constitution**: The root governance document for the Unbound Force organization. Attributes: version, ratification_date, last_amended_date, principles[], governance_rules[], development_workflow.
- **Core Principle**: A named, numbered principle with description, MUST rules, SHOULD rules, and MAY rules. Attributes: number (I/II/III), name, description, must_rules[], should_rules[], may_rules[].
- **Hero Constitution**: A per-repository constitution that extends the org constitution. Attributes: parent_constitution_version, hero_name, principles[], governance_rules[].
- **Alignment Check**: A validation that a hero constitution does not contradict the org constitution. Performed by an OpenCode agent that reads both constitutions and produces a structured report. Attributes: hero_name, org_constitution_version, findings[] (each with principle, status, rationale), overall_status (aligned/non-aligned).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The ratified constitution contains exactly three principles, each with at least three MUST rules and at least one SHOULD rule.
- **SC-002**: The Governance section defines all four required elements: amendment process, versioning, supremacy clause, hero alignment requirements.
- **SC-003**: The existing Gaze constitution (v1.0.0) passes an alignment check against the org constitution with zero contradictions.
- **SC-004**: The existing Website constitution (v1.0.0) passes an alignment check against the org constitution with zero contradictions.
- **SC-005**: A developer presented with three cross-cutting design questions can find a clear, citable answer in the constitution for each one.
- **SC-006**: The constitution document is under 500 lines (concise enough to be read and internalized by both humans and agents).

## Dependencies

### Prerequisites

- None. This is the foundational spec.

### Downstream Dependents

- **Spec 002** (Hero Interface Contract): Depends on the constitution for principle definitions. Defines the Hero Interface Contract concept referenced by FR-005.
- **Spec 003** (Speckit Framework): Depends on the constitution for the Constitution Check process.
- **Specs 004-007** (Hero Architectures): All depend on the constitution for design constraints.
- **Spec 008** (Swarm Orchestration): Depends on the constitution for inter-hero communication rules.
- **Spec 009** (Shared Data Model): Depends on the constitution for output format requirements.
