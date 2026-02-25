# Feature Specification: Speckit Framework Centralization

**Feature Branch**: `003-speckit-framework`
**Created**: 2026-02-24
**Status**: Draft
**Input**: User description: "Centralize the speckit framework (templates, scripts, OpenCode commands) into a single source of truth with a distribution mechanism that eliminates cross-repo copy-paste drift. Define extension points so individual hero repos can customize without forking."

## Clarifications

### Session 2026-02-24

- Q: Should speckit become its own repository or remain part of the unbound-force meta repo? A: Speckit should become its own repository (`unbound-force/speckit`) since it is an independent tool used by all heroes and potentially by external projects.
- Q: Known drift points: Gaze `speckit.specify.md` says "RESTful APIs unless specified otherwise" while unbound-force says "Use project-appropriate patterns." Gaze `speckit.plan.md` says "Generate API contracts" while unbound-force says "Define interface contracts." Which is canonical? A: The unbound-force (broader) versions are canonical. The Gaze versions are project-specific customizations that should be handled via configuration, not file modification.
- Q: What distribution mechanism should speckit use? A: Define multiple options (git submodule, npm package, homebrew, standalone CLI) and recommend the best fit during the plan phase.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Single Source of Truth (Priority: P1)

A maintainer establishes a canonical speckit repository (`unbound-force/speckit`) containing the authoritative versions of all templates, scripts, and OpenCode command files. All hero repositories consume speckit from this single source rather than maintaining independent copies.

**Why this priority**: P1 because drift is already happening between the three existing repos (Gaze, Website, unbound-force). Every day without a canonical source adds more divergence.

**Independent Test**: Can be tested by comparing the canonical speckit repository contents against the files currently in each hero repo's `.specify/` and `.opencode/command/speckit.*.md` directories, verifying the canonical version is a valid superset.

**Acceptance Scenarios**:

1. **Given** the canonical speckit repository, **When** a maintainer inspects its contents, **Then** it contains exactly one authoritative version of each template (`spec-template.md`, `plan-template.md`, `tasks-template.md`, `checklist-template.md`, `constitution-template.md`, `agent-file-template.md`), each script (`common.sh`, `check-prerequisites.sh`, `setup-plan.sh`, `create-new-feature.sh`, `update-agent-context.sh`), and each OpenCode command (`speckit.constitution.md`, `speckit.specify.md`, `speckit.clarify.md`, `speckit.plan.md`, `speckit.tasks.md`, `speckit.analyze.md`, `speckit.checklist.md`, `speckit.implement.md`, `speckit.taskstoissues.md`).
2. **Given** the canonical versions, **When** they are compared to the Gaze repo copies, **Then** all differences are identified and categorized as either (a) drift to be corrected or (b) legitimate project-specific customizations to be handled via extension points.
3. **Given** the canonical versions, **When** they are compared to the Website repo copies, **Then** the same drift analysis is performed.
4. **Given** the canonical versions, **When** they are compared to the unbound-force repo copies, **Then** the unbound-force versions are identical to canonical (since they are the most general).

---

### User Story 2 - Distribution and Installation (Priority: P1)

A maintainer of a new or existing hero repository installs speckit from the canonical source. The installation process places the correct files in the correct directories (`.specify/templates/`, `.specify/scripts/`, `.opencode/command/`) and records the installed version for future upgrade detection.

**Why this priority**: P1 because without a distribution mechanism, the single source of truth is useless — people will continue to copy-paste.

**Independent Test**: Can be tested by running the installation command in a fresh repository and verifying the correct files are placed in the correct locations with the correct version metadata.

**Acceptance Scenarios**:

1. **Given** a fresh repository with no speckit files, **When** the installation command runs, **Then** it creates `.specify/templates/` with all 6 templates, `.specify/scripts/bash/` with all 5 scripts, and `.opencode/command/` with all 9 speckit commands.
2. **Given** a repository with an older speckit version installed, **When** the upgrade command runs, **Then** it updates only the canonical files (not user-modified files) and reports which files were updated.
3. **Given** a repository with local modifications to a speckit file, **When** the upgrade command runs, **Then** it detects the modification, skips that file, and warns the user with a diff summary.
4. **Given** the installation completes, **When** a maintainer inspects `.specify/speckit.version`, **Then** it contains the installed speckit version and installation timestamp.

---

### User Story 3 - Project-Specific Extension Points (Priority: P2)

A hero repository customizes speckit behavior for its specific technology stack and domain without modifying the canonical speckit files. Extension points allow projects to inject project-specific context (e.g., language conventions, build commands, testing frameworks) that speckit commands consume during execution.

**Why this priority**: P2 because the existing drift between repos is partly caused by legitimate project-specific needs (Gaze needs Go-specific patterns, Website needs Hugo-specific patterns). Extension points prevent future drift by providing a sanctioned customization mechanism.

**Independent Test**: Can be tested by configuring a project-specific extension (e.g., specifying Go as the language and `go test` as the test command) and verifying that speckit commands use these values instead of defaults.

**Acceptance Scenarios**:

1. **Given** a `.specify/config.yaml` extension file in a hero repo, **When** a speckit command (e.g., `/specify`) runs, **Then** it reads project-specific values (language, framework, build_command, test_command, integration_patterns) from the config and injects them into the template filling process.
2. **Given** no `.specify/config.yaml` exists, **When** a speckit command runs, **Then** it uses sensible defaults (language-agnostic patterns, generic build/test placeholders).
3. **Given** the Gaze project has `config.yaml` with `language: go` and `integration_patterns: "RESTful APIs unless specified otherwise"`, **When** `/specify` runs, **Then** it produces the same Go-specific output currently hardcoded in Gaze's forked `speckit.specify.md`.
4. **Given** a project-specific override for one section of a template, **When** the template is filled, **Then** only the overridden section uses the custom content; all other sections use the canonical template.

---

### User Story 4 - Speckit Pipeline Documentation (Priority: P3)

A new contributor to any Unbound Force hero repository understands the complete speckit pipeline (constitution -> specify -> clarify -> plan -> tasks -> analyze -> checklist -> implement -> taskstoissues) through clear documentation that explains each phase, its inputs, outputs, and relationship to other phases.

**Why this priority**: P3 because the pipeline is already functional — documentation improves adoption and reduces onboarding friction but does not block development.

**Independent Test**: Can be tested by having a new contributor follow the documentation to create a feature spec from scratch in a test repository, verifying each pipeline phase produces the expected output.

**Acceptance Scenarios**:

1. **Given** the speckit documentation, **When** a contributor reads the pipeline overview, **Then** they can identify all 9 phases, their order, and which are optional vs. mandatory.
2. **Given** the documentation, **When** a contributor reads a phase description, **Then** it includes: purpose, prerequisites, inputs (which files must exist), outputs (which files are created/modified), and the OpenCode command to invoke it.
3. **Given** the documentation, **When** a contributor follows the quickstart guide, **Then** they can run the first three phases (constitution, specify, clarify) and produce a valid spec with clarifications.

---

### Edge Cases

- What happens when a project uses a speckit version that is incompatible with the latest canonical version? The version file MUST include a minimum compatible speckit version, and the upgrade command MUST refuse to upgrade if the project's speckit usage relies on removed features.
- What happens when two speckit commands are run concurrently in the same repo? Speckit commands are designed for serial execution. Concurrent execution SHOULD be detected and one SHOULD yield with a warning.
- What happens when speckit is installed in a non-hero repository (e.g., an external project)? Speckit MUST work in any repository, not just Unbound Force hero repos. The Hero Interface Contract features are additive.
- What happens when a script (`common.sh`, etc.) is modified locally and an upgrade is attempted? The upgrade tool MUST detect the modification via checksum comparison and skip the file with a warning.
- What happens when a new speckit command is added in a later version? The upgrade command MUST install new files that do not exist locally, even when other files are skipped due to modifications.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Speckit MUST have a canonical repository (`unbound-force/speckit`) containing all authoritative templates, scripts, and OpenCode commands.
- **FR-002**: The canonical repository MUST include a versioning mechanism (semantic versioning) tracked in a manifest file.
- **FR-003**: Speckit MUST provide an installation mechanism that places files into the correct directories: `.specify/templates/`, `.specify/scripts/bash/`, `.opencode/command/`.
- **FR-004**: Speckit MUST provide an upgrade mechanism that updates canonical files while preserving local modifications to previously installed files.
- **FR-005**: The upgrade mechanism MUST detect local modifications via content checksum comparison and skip modified files with a warning.
- **FR-006**: Speckit MUST record the installed version in `.specify/speckit.version` after installation or upgrade.
- **FR-007**: Speckit MUST support project-specific configuration via `.specify/config.yaml` with at minimum: `language`, `framework`, `build_command`, `test_command`, `integration_patterns`, and `project_type` (library/cli/web/mobile) fields.
- **FR-008**: Speckit commands MUST read `.specify/config.yaml` (if present) and use its values to fill templates and guide output generation.
- **FR-009**: Speckit MUST define the complete pipeline: constitution -> specify -> clarify -> plan -> tasks -> analyze -> checklist -> implement -> taskstoissues, with each phase documented.
- **FR-010**: The `speckit.specify.md` command MUST use `integration_patterns` from config.yaml instead of hardcoding language-specific patterns.
- **FR-011**: The `speckit.plan.md` command MUST use `project_type` from config.yaml to determine whether to generate API contracts, CLI contracts, or library interface contracts.
- **FR-012**: Speckit MUST work in any Git repository, not only Unbound Force hero repositories.
- **FR-013**: Speckit SHOULD provide a `speckit init` command (or equivalent) that initializes a fresh repository with the correct directory structure and optionally runs the constitution phase.
- **FR-014**: Speckit MUST include a drift detection mechanism that compares installed files against the canonical versions and reports differences.
- **FR-015**: The canonical repository MUST include automated tests that verify all templates are syntactically valid and all scripts execute without errors on macOS and Linux.

### Key Entities

- **Speckit Manifest**: Metadata about the canonical speckit distribution. Attributes: version (semver), files[] (list of files with paths and checksums), minimum_compatible_version, changelog_url.
- **Speckit Installation Record**: Per-repository tracking of installed speckit version. Attributes: speckit_version, installed_at (ISO 8601), files_installed[], files_skipped[], canonical_checksums{}.
- **Project Configuration**: Per-repository speckit customization. Attributes: language, framework, build_command, test_command, integration_patterns, project_type, custom_sections{}.
- **Pipeline Phase**: One step in the speckit pipeline. Attributes: name, order, command, prerequisites[], inputs[], outputs[], is_optional (bool).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A canonical speckit repository exists with exactly 6 templates, 5 scripts, and 9 OpenCode commands.
- **SC-002**: The installation mechanism places files into a fresh repository in under 10 seconds with zero errors.
- **SC-003**: The upgrade mechanism correctly detects and skips locally modified files (verified by modifying a template, running upgrade, and confirming the modified file is preserved).
- **SC-004**: The drift detection mechanism identifies all known differences between the Gaze, Website, and unbound-force speckit copies.
- **SC-005**: A project-specific `config.yaml` with `language: go` produces the same output from `/specify` that the current hardcoded Gaze `speckit.specify.md` produces.
- **SC-006**: The pipeline documentation covers all 9 phases with inputs, outputs, and prerequisites for each.
- **SC-007**: Speckit installs and functions correctly in a non-Unbound Force repository (verified by testing in a fresh, unrelated project).

## Dependencies

### Prerequisites

- **Spec 001** (Org Constitution): Speckit includes a constitution phase that must align with the org constitution process.
- **Spec 002** (Hero Interface Contract): Speckit distribution is part of the hero bootstrapping process defined by the contract.

### Downstream Dependents

- **Specs 004-007** (Hero Architectures): All hero repos consume speckit.
- **All future specs**: Use speckit for their own specification process (meta-dependency).

```
Spec 001 (Constitution)
  └─> Spec 002 (Interface Contract)
       └─> Spec 003 (Speckit Framework)
            └─> All hero repos consume speckit
```
