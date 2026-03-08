---
spec_id: "003"
title: "Specification Framework"
phase: 0
status: draft
depends_on:
  - "[[specs/001-org-constitution/spec]]"
---

# Feature Specification: Specification Framework

**Feature Branch**: `003-specification-framework`
**Created**: 2026-02-24
**Updated**: 2026-03-08
**Status**: Draft
**Input**: User description: "Establish a unified specification
framework with two tiers: Speckit for strategic, architectural
work (new heroes, cross-cutting specs, constitution changes) and
OpenSpec for tactical work (bugs, quick fixes, small enhancements,
maintenance tasks). Both tiers share a single governance bridge
through the org constitution. This repo is the canonical source
for all framework artifacts."

## Clarifications

### Session 2026-02-24

- Q: Should speckit become its own repository or remain part of
  the unbound-force meta repo? A: This repo
  (`unbound-force/unbound-force`) is the canonical source for
  all specification framework artifacts. No separate repository
  is needed.
- Q: Known drift points: Gaze `speckit.specify.md` says
  "RESTful APIs unless specified otherwise" while unbound-force
  says "Use project-appropriate patterns." Gaze
  `speckit.plan.md` says "Generate API contracts" while
  unbound-force says "Define interface contracts." Which is
  canonical? A: The unbound-force (broader) versions are
  canonical. The Gaze versions are project-specific
  customizations that should be handled via configuration, not
  file modification.
- Q: What distribution mechanism should speckit use? A: Define
  multiple options (git submodule, npm package, homebrew,
  standalone CLI) and recommend the best fit during the plan
  phase.

### Session 2026-03-08

- Q: Should OpenSpec be added as a tactical specification tool
  alongside Speckit? A: Yes. Speckit handles strategic changes
  (architectural specs, cross-cutting concerns, constitution
  changes, work with 3 or more user stories). OpenSpec handles
  tactical changes (bugs, quick fixes, small enhancements,
  standalone maintenance tasks, work with fewer than 3 stories).
- Q: Where should OpenSpec specs live relative to Speckit specs?
  A: Separate directories. Speckit specs stay in `./specs/`
  (numbered). OpenSpec uses `openspec/specs/` and
  `openspec/changes/` as its own namespace. Two distinct spec
  trees in all repos.
- Q: Should OpenSpec proposals pass constitution checks? A: Yes,
  all proposals must pass constitution alignment. Enforcement
  via a custom OpenSpec schema (`unbound-force`) that injects
  the constitution into the proposal template with a mandatory
  alignment section.
- Q: Where should the custom OpenSpec schema live? A: In this
  repo at `openspec/schemas/unbound-force/`, distributed
  alongside Speckit artifacts.
- Q: Is the boundary between strategic and tactical enforced or
  advisory? A: Advisory. The boundary criteria are documented
  guidelines (SHOULD), not enforced gates. Teams use judgment.
- Q: Where should OpenSpec specs live in hero repos? A: Same
  pattern everywhere. All repos use `./specs/` for Speckit and
  `openspec/specs/` for OpenSpec.
- Q: What should the distribution binary be named? A: `unbound`
  (not `speckit`, which would collide with the existing Speckit
  CLI whose binary is `specify`). Distributed via
  `brew install unbound-force/tap/unbound` and `go install`.
- Q: What distribution mechanism should the framework use?
  A: Go binary with `embed.FS` scaffold pattern, matching the
  Gaze project's architecture. GoReleaser for cross-platform
  builds, Homebrew cask for installation. Replaces the earlier
  Bash copy script decision from Session 2026-02-24.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Single Source of Truth (Priority: P1)

A maintainer establishes this repository
(`unbound-force/unbound-force`) as the canonical source for
all specification framework artifacts: Speckit templates,
scripts, and OpenCode commands for the strategic tier, plus
the custom OpenSpec schema and configuration templates for
the tactical tier. All hero repositories consume both tiers
from this single source rather than maintaining independent
copies.

**Why this priority**: P1 because drift is already happening
between the three existing repos (Gaze, Website,
unbound-force). Every day without a canonical source adds
more divergence. Adding OpenSpec as a second tier makes a
single source of truth even more critical.

**Independent Test**: Can be tested by comparing this
repository's contents against the files currently in each
hero repo's `.specify/`, `.opencode/command/speckit.*.md`,
and `openspec/` directories, verifying the canonical version
is a valid superset.

**Acceptance Scenarios**:

1. **Given** this repository as the canonical source, **When**
   a maintainer inspects its contents, **Then** it contains
   exactly one authoritative version of each Speckit template
   (`spec-template.md`, `plan-template.md`,
   `tasks-template.md`, `checklist-template.md`,
   `constitution-template.md`, `agent-file-template.md`),
   each script (`common.sh`, `check-prerequisites.sh`,
   `setup-plan.sh`, `create-new-feature.sh`,
   `update-agent-context.sh`), and each OpenCode command
   (`speckit.constitution.md`, `speckit.specify.md`,
   `speckit.clarify.md`, `speckit.plan.md`,
   `speckit.tasks.md`, `speckit.analyze.md`,
   `speckit.checklist.md`, `speckit.implement.md`,
   `speckit.taskstoissues.md`).
2. **Given** this repository as the canonical source, **When**
   a maintainer inspects the `openspec/schemas/unbound-force/`
   directory, **Then** it contains the custom OpenSpec schema
   definition and proposal template with a mandatory
   constitution alignment section.
3. **Given** the canonical versions, **When** they are compared
   to the Gaze repo copies, **Then** all differences are
   identified and categorized as either (a) drift to be
   corrected or (b) legitimate project-specific customizations
   to be handled via extension points.
4. **Given** the canonical versions, **When** they are compared
   to the Website repo copies, **Then** the same drift
   analysis is performed.

---

### User Story 2 - Distribution and Installation (Priority: P1)

A maintainer of a new or existing hero repository installs
the specification framework from the canonical source by
running `unbound init`. The scaffold process places Speckit
files in the correct directories (`.specify/templates/`,
`.specify/scripts/`, `.opencode/command/`), initializes
the OpenSpec directory structure with the custom
`unbound-force` schema and configuration, and marks each
scaffolded file with a version comment for provenance.

**Why this priority**: P1 because without a distribution
mechanism, the single source of truth is useless -- people
will continue to copy-paste. Both tiers must be installable
in one step.

**Independent Test**: Can be tested by running the
installation command in a fresh repository and verifying
Speckit files, OpenSpec schema, and OpenSpec configuration
are placed in the correct locations with the correct version
metadata.

**Acceptance Scenarios**:

1. **Given** a fresh repository with no framework files,
   **When** the installation command runs, **Then** it creates
   `.specify/templates/` with all 6 templates,
   `.specify/scripts/bash/` with all 5 scripts,
   `.opencode/command/` with all 9 speckit commands, and
   `openspec/schemas/unbound-force/` with the custom schema
   and templates.
2. **Given** a repository with an older framework version
   installed, **When** the upgrade command runs, **Then** it
   updates only the canonical files (not user-modified files)
   for both Speckit and OpenSpec artifacts and reports which
   files were updated.
3. **Given** a repository with local modifications to a
   framework file, **When** the upgrade command runs, **Then**
   it detects the modification, skips that file, and warns the
   user with a diff summary.
4. **Given** the scaffold completes, **When** a maintainer
   inspects any scaffolded file, **Then** it contains a
   version marker comment identifying which version of the
   `unbound` binary created it.

---

### User Story 3 - Project-Specific Extension Points
(Priority: P2)

A hero repository customizes specification framework behavior
for its specific technology stack and domain without modifying
the canonical files. Extension points allow projects to inject
project-specific context (e.g., language conventions, build
commands, testing frameworks) that both Speckit commands and
OpenSpec proposals consume during execution.

**Why this priority**: P2 because the existing drift between
repos is partly caused by legitimate project-specific needs
(Gaze needs Go-specific patterns, Website needs Hugo-specific
patterns). Extension points prevent future drift by providing
a sanctioned customization mechanism for both tiers.

**Independent Test**: Can be tested by configuring
project-specific extensions for both Speckit and OpenSpec and
verifying that commands and proposals use these values instead
of defaults.

**Acceptance Scenarios**:

1. **Given** a `.specify/config.yaml` extension file in a hero
   repo, **When** a speckit command (e.g., `/specify`) runs,
   **Then** it reads project-specific values (language,
   framework, build_command, test_command,
   integration_patterns) from the config and injects them into
   the template filling process.
2. **Given** no `.specify/config.yaml` exists, **When** a
   speckit command runs, **Then** it uses sensible defaults
   (language-agnostic patterns, generic build/test
   placeholders).
3. **Given** the Gaze project has `config.yaml` with
   `language: go` and `integration_patterns: "RESTful APIs
   unless specified otherwise"`, **When** `/specify` runs,
   **Then** it produces the same Go-specific output currently
   hardcoded in Gaze's forked `speckit.specify.md`.
4. **Given** a project-specific override for one section of a
   template, **When** the template is filled, **Then** only
   the overridden section uses the custom content; all other
   sections use the canonical template.
5. **Given** an `openspec/config.yaml` with project-specific
   context and rules, **When** an OpenSpec proposal is
   created, **Then** the proposal includes the project-specific
   context alongside the constitution alignment section.

---

### User Story 4 - Specification Pipeline Documentation
(Priority: P1)

A new contributor to any Unbound Force hero repository
understands both specification workflows -- the Speckit
strategic pipeline and the OpenSpec tactical workflow --
through clear documentation that explains each tool's
purpose, its phases, inputs, outputs, and the guidelines for
choosing between them.

**Why this priority**: P1 because with two specification
tools, clear documentation of both pipelines and the boundary
guidelines is essential to prevent confusion and misuse.
Contributors must know when to use which tool.

**Independent Test**: Can be tested by having a new
contributor follow the documentation to create both a
strategic spec (via Speckit) and a tactical change (via
OpenSpec) in a test repository, verifying each workflow
produces the expected output.

**Acceptance Scenarios**:

1. **Given** the framework documentation, **When** a
   contributor reads the pipeline overview, **Then** they can
   identify all 9 Speckit phases, the 4 core OpenSpec actions,
   their order, and which are optional vs. mandatory.
2. **Given** the documentation, **When** a contributor reads a
   phase or action description, **Then** it includes: purpose,
   prerequisites, inputs (which files must exist), outputs
   (which files are created/modified), and the command to
   invoke it.
3. **Given** the documentation, **When** a contributor follows
   the quickstart guide, **Then** they can run the first three
   Speckit phases (constitution, specify, clarify) and produce
   a valid spec with clarifications.
4. **Given** the documentation, **When** a contributor reads
   the boundary guidelines, **Then** they can determine
   whether a given piece of work should use Speckit or
   OpenSpec based on scope, story count, and change type.

---

### User Story 5 - Tactical Change Workflow (Priority: P1)

A developer uses OpenSpec for feature-level changes, bug
fixes, small enhancements, and standalone maintenance tasks
within any Unbound Force repository. The tactical workflow
provides a lightweight, fluid process -- propose, apply,
archive -- that produces just enough specification to guide
implementation without the overhead of the full Speckit
pipeline.

**Why this priority**: P1 because the full Speckit pipeline
is too heavyweight for small changes. Developers currently
skip specification entirely for tactical work, leading to
undocumented changes and context loss.

**Independent Test**: Can be tested by creating a tactical
change (e.g., "fix artifact envelope validation") using the
OpenSpec workflow, verifying the proposal includes
constitution alignment, tasks are checkable, and archiving
preserves the change history.

**Acceptance Scenarios**:

1. **Given** a repository with OpenSpec installed and
   configured, **When** a developer runs the propose command,
   **Then** it creates a change directory under
   `openspec/changes/` with a proposal, delta specs, design,
   and tasks.
2. **Given** an active OpenSpec change with tasks, **When** a
   developer completes implementation, **Then** tasks are
   checked off and the change can be applied.
3. **Given** a completed OpenSpec change, **When** the
   developer archives it, **Then** delta specs merge into
   `openspec/specs/`, the change folder moves to
   `openspec/changes/archive/`, and full context is preserved.
4. **Given** an OpenSpec change that modifies behavior
   documented in `openspec/specs/`, **When** the delta spec
   is reviewed, **Then** it clearly shows what requirements
   are being added, modified, or removed.

---

### User Story 6 - Constitution Governance Bridge
(Priority: P1)

Every OpenSpec proposal in any Unbound Force repository
automatically includes a constitution alignment check as
part of the proposal creation process. The custom
`unbound-force` OpenSpec schema injects the org constitution
into the proposal context and requires explicit alignment
assessment against all three principles before tasks can
proceed.

**Why this priority**: P1 because constitution alignment is
non-negotiable per the org constitution. Tactical changes
must not bypass governance -- a lightweight change process
does not mean a governance-free process.

**Independent Test**: Can be tested by creating an OpenSpec
proposal and verifying that the proposal template includes a
constitution alignment section, the constitution content is
available in the proposal context, and all three principles
are addressed.

**Acceptance Scenarios**:

1. **Given** a repository with the `unbound-force` OpenSpec
   schema installed, **When** a developer creates a new
   proposal, **Then** the proposal template includes a
   "Constitution Alignment" section with fields for each of
   the three principles (Autonomous Collaboration,
   Composability First, Observable Quality).
2. **Given** a proposal being created with the `unbound-force`
   schema, **When** the schema context is loaded, **Then** the
   full org constitution from
   `.specify/memory/constitution.md` is available to the agent
   generating the proposal.
3. **Given** a completed proposal, **When** a reviewer
   inspects the constitution alignment section, **Then** each
   principle has an explicit PASS or N/A assessment with a
   brief justification.

---

### User Story 7 - Strategic/Tactical Boundary Guidelines
(Priority: P1)

A developer or maintainer can quickly determine whether a
given piece of work should use the Speckit strategic pipeline
or the OpenSpec tactical workflow by consulting clear,
documented boundary guidelines. The guidelines are advisory,
not enforced gates, and are based on objective criteria such
as scope, story count, and change type.

**Why this priority**: P1 because with two specification
tools, confusion about which to use will lead to inconsistent
practices, over-engineering small changes, or
under-specifying large ones.

**Independent Test**: Can be tested by presenting a set of
representative work items (ranging from a typo fix to a new
hero architecture) and verifying the guidelines produce a
clear, consistent recommendation for each.

**Acceptance Scenarios**:

1. **Given** the boundary guidelines, **When** a developer
   evaluates work involving 3 or more user stories, **Then**
   the guidelines recommend Speckit.
2. **Given** the boundary guidelines, **When** a developer
   evaluates a single bug fix or small enhancement, **Then**
   the guidelines recommend OpenSpec.
3. **Given** the boundary guidelines, **When** a developer
   evaluates work that crosses repository boundaries or
   affects the org constitution, **Then** the guidelines
   recommend Speckit regardless of story count.
4. **Given** the boundary guidelines, **When** a developer
   encounters an edge case not clearly covered, **Then** the
   guidelines provide a decision heuristic (e.g., "when in
   doubt, start with OpenSpec and escalate to Speckit if
   scope grows").

---

### User Story 8 - Schema and Template Distribution
(Priority: P2)

The custom `unbound-force` OpenSpec schema -- including its
workflow definition, proposal template with constitution
alignment, and delta spec templates -- is versioned and
distributed alongside Speckit artifacts through the same
installation and upgrade mechanism. Schema updates are
tracked and applied consistently across all repositories.

**Why this priority**: P2 because the schema must exist
before tactical workflows can be used (addressed by US5 and
US6), but the distribution mechanism can initially be manual.
Automated distribution improves consistency over time.

**Independent Test**: Can be tested by modifying the
canonical schema in this repository, running an upgrade in
a hero repo, and verifying the schema is updated while
preserving any local configuration.

**Acceptance Scenarios**:

1. **Given** a new version of the `unbound-force` schema in
   this repository, **When** the upgrade command runs in a
   hero repo, **Then** the schema files in
   `openspec/schemas/unbound-force/` are updated.
2. **Given** a hero repo with a locally modified
   `openspec/config.yaml`, **When** the upgrade command runs,
   **Then** the config file is preserved (not overwritten)
   while schema files are updated.
3. **Given** a new template added to the schema in a later
   version, **When** the upgrade command runs, **Then** the
   new template is installed alongside existing files.

---

### Edge Cases

- What happens when a project uses a framework version that is
  incompatible with the latest canonical version? The version
  marker in scaffolded files identifies the installed version.
  Re-running `unbound init` overwrites tool-owned files with
  the new version; user-owned files are preserved.
- What happens when two speckit commands are run concurrently
  in the same repo? Speckit commands are designed for serial
  execution within a single OpenCode session. Concurrent
  execution across sessions is outside the scope of this spec
  — it is an OpenCode runtime concern, not a framework
  distribution concern. Concurrent writes to spec artifacts
  produce undefined behavior.
- What happens when speckit is installed in a non-hero
  repository (e.g., an external project)? The framework MUST
  work in any repository, not just Unbound Force hero repos.
  The Hero Interface Contract features and OpenSpec schema are
  additive.
- What happens when a script (`common.sh`, etc.) is modified
  locally and `unbound init` is re-run? Scripts are
  user-owned files. Re-running `unbound init` skips
  user-owned files that already exist (unless `--force` is
  used).
- What happens when a new speckit command is added in a later
  version? Re-running `unbound init` creates files that do
  not exist locally (new files are always created regardless
  of ownership classification).
- What happens when a developer creates an OpenSpec change
  that should have been a Speckit spec? The boundary
  guidelines SHOULD provide an escalation path: convert the
  OpenSpec change into a Speckit spec by extracting the
  proposal content into a new numbered spec directory under
  `specs/`.
- What happens when an OpenSpec change touches files under
  `specs/` (Speckit's domain)? The framework MUST prevent
  OpenSpec changes from modifying files under `specs/`.
  OpenSpec delta specs apply only to `openspec/specs/`.
- What happens when the constitution is amended after OpenSpec
  proposals were created? Active proposals SHOULD be
  re-evaluated against the updated constitution. The schema
  context references the live constitution file, so new
  proposals automatically use the latest version.
- What happens when OpenSpec is not installed but Speckit is?
  Speckit MUST function independently. OpenSpec is additive
  and optional for repos that choose not to adopt the
  tactical tier.

## Requirements *(mandatory)*

### Functional Requirements

#### Speckit (Strategic Tier)

- **FR-001**: This repository (`unbound-force/unbound-force`)
  MUST serve as the canonical source for all Speckit
  templates, scripts, OpenCode commands, and the custom
  OpenSpec `unbound-force` schema.
- **FR-002**: The canonical source MUST include a versioning
  mechanism (semantic versioning) injected at build time via
  ldflags and visible in scaffolded file markers.
- **FR-003**: The framework MUST provide a Go binary
  (`unbound`) that scaffolds files into the correct
  directories (`.specify/templates/`,
  `.specify/scripts/bash/`, `.opencode/command/`,
  `.opencode/agents/`) and OpenSpec schema files into
  `openspec/schemas/unbound-force/`, using Go's `embed.FS`
  to compile all distributable files into the binary.
- **FR-004**: The framework MUST classify scaffolded files as
  user-owned or tool-owned. On re-run, user-owned files that
  already exist MUST be skipped. Tool-owned files MUST be
  updated if their content has changed.
- **FR-005**: The framework MUST support a `--force` flag
  that overrides the user-owned skip behavior and overwrites
  all files.
- **FR-006**: Each scaffolded file MUST include a version
  marker comment identifying the unbound version that
  created it.
- **FR-007**: The framework MUST support project-specific
  configuration via `.specify/config.yaml` with at minimum:
  `language`, `framework`, `build_command`, `test_command`,
  `integration_patterns`, and `project_type`
  (library/cli/web/mobile) fields.
- **FR-008**: Speckit commands MUST read
  `.specify/config.yaml` (if present) and use its values to
  fill templates and guide output generation.
- **FR-009**: The framework MUST define the complete Speckit
  pipeline (constitution -> specify -> clarify -> plan ->
  tasks -> analyze -> checklist -> implement -> taskstoissues)
  and the OpenSpec workflow (propose -> apply -> archive),
  with each phase or action documented.
- **FR-010**: The `speckit.specify.md` command MUST use
  `integration_patterns` from config.yaml instead of
  hardcoding language-specific patterns.
- **FR-011**: The `speckit.plan.md` command MUST use
  `project_type` from config.yaml to determine whether to
  generate API contracts, CLI contracts, or library interface
  contracts.
- **FR-012**: The framework MUST work in any Git repository,
  not only Unbound Force hero repositories.
- **FR-013**: The framework MUST provide an `unbound init`
  command that initializes a fresh repository with the
  correct directory structure for both tiers.
- **FR-014**: The framework MUST include a drift detection
  test that verifies embedded asset copies are byte-identical
  to the canonical source files in the repository.
- **FR-015**: The framework MUST be distributed via Homebrew
  cask (`unbound-force/tap/unbound`) and `go install`, using
  GoReleaser for cross-platform builds (darwin/amd64,
  darwin/arm64, linux/amd64, linux/arm64).

#### OpenSpec (Tactical Tier)

- **FR-016**: All Unbound Force repositories MUST have the
  `openspec/` directory initialized with the
  `unbound-force` custom schema.
- **FR-017**: The `unbound-force` OpenSpec schema MUST
  include a constitution alignment section in the proposal
  template requiring explicit PASS or N/A assessment for
  each of the three org constitution principles.
- **FR-018**: The OpenSpec `config.yaml` in each repository
  MUST reference `.specify/memory/constitution.md` in its
  context field so that the constitution content is available
  to agents generating proposals.
- **FR-019**: The `unbound init` command MUST also scaffold
  the custom OpenSpec schema and provide a default
  `openspec/config.yaml` template.

#### Governance and Boundaries

- **FR-020**: Boundary guidelines SHOULD be documented
  recommending Speckit for architectural and cross-cutting
  work (3 or more user stories, cross-repo impact,
  constitution changes) and OpenSpec for tactical work (bugs,
  small enhancements, maintenance tasks, fewer than 3
  stories). These guidelines are advisory, not enforced.
- **FR-021**: OpenSpec changes MUST NOT modify files under
  `specs/` (Speckit's domain). OpenSpec delta specs apply
  only to files under `openspec/specs/`.
- **FR-022**: Speckit specs MUST NOT be created inside the
  `openspec/` directory. Speckit specs live exclusively
  under `specs/`.

### Key Entities

- **Scaffold Options**: Configuration for the scaffold run.
  Attributes: TargetDir (path), Version (semver string),
  Force (boolean).
- **Scaffold Result**: Outcome of a scaffold run. Attributes:
  Created[] (new files), Skipped[] (existing user-owned),
  Overwritten[] (force-replaced), Updated[] (tool-owned
  content changes).
- **Speckit Project Configuration**: Per-repository Speckit
  customization. Attributes: language, framework,
  build_command, test_command, integration_patterns,
  project_type, custom_sections{}.
- **OpenSpec Project Configuration**: Per-repository OpenSpec
  customization. Attributes: schema (default workflow),
  context (project-specific context including constitution
  reference), rules (per-artifact rules).
- **Pipeline Phase**: One step in the Speckit pipeline.
  Attributes: name, order, command, prerequisites[], inputs[],
  outputs[], is_optional (bool).
- **OpenSpec Action**: One action in the OpenSpec workflow.
  Attributes: name, command, prerequisites[], inputs[],
  outputs[].
- **Custom Schema**: An OpenSpec schema definition tailored
  to the organization. Attributes: name, artifacts[] (with
  dependency DAG), templates{}, context_injection{}.
- **Boundary Guideline**: A documented criterion for choosing
  between specification tiers. Attributes: criterion,
  recommendation (speckit/openspec), examples[].

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: This repository contains exactly 6 Speckit
  templates, 5 scripts, 10 OpenCode command files
  (9 speckit pipeline commands + 1 constitution-check),
  1 agent file, and 1 custom OpenSpec schema with at least
  4 template files.
- **SC-002**: The `unbound init` command scaffolds all
  framework files into a fresh repository in under 5 seconds
  with zero errors.
- **SC-003**: Re-running `unbound init` correctly skips
  user-owned files that already exist and updates tool-owned
  files with changed content (verified by modifying a
  template, running init, and confirming the modified file
  is preserved while tool-owned files are updated).
- **SC-004**: The drift detection test identifies when
  embedded asset copies differ from canonical source files
  in the repository.
- **SC-005**: A project-specific `.specify/config.yaml` with
  `language: go` produces the same output from `/specify`
  that the current hardcoded Gaze `speckit.specify.md`
  produces.
- **SC-006**: The pipeline documentation covers all 9 Speckit
  phases and all 4 core OpenSpec actions with inputs, outputs,
  and prerequisites for each.
- **SC-007**: The `unbound init` command works correctly in a
  non-Unbound Force repository (verified by running in a
  fresh, unrelated project).
- **SC-008**: An OpenSpec proposal created with the
  `unbound-force` schema contains a constitution alignment
  section with assessments for all three principles.
- **SC-009**: The boundary guidelines documentation enables a
  new contributor to correctly classify at least 8 out of 10
  representative work items as strategic or tactical.
- **SC-010**: OpenSpec changes cannot create or modify files
  under `specs/`, and Speckit specs cannot be created under
  `openspec/` (verified by attempting both and confirming
  rejection).

## Dependencies

### Prerequisites

- **Spec 001** (Org Constitution): The framework includes a
  constitution phase (Speckit) and constitution injection
  (OpenSpec) that must align with the org constitution
  process.
- **Spec 002** (Hero Interface Contract): Framework
  distribution is part of the hero bootstrapping process
  defined by the contract.

### Downstream Dependents

- **Specs 004-007** (Hero Architectures): All hero repos
  consume the specification framework.
- **All future specs**: Use Speckit for their own
  specification process (meta-dependency).
- **All tactical changes**: Use OpenSpec with the
  `unbound-force` schema for bug fixes, small enhancements,
  and maintenance.

```
Spec 001 (Constitution)
  +-> Spec 002 (Interface Contract)
       +-> Spec 003 (Specification Framework)
            +-> All hero repos consume framework
            +-> Speckit for strategic specs
            +-> OpenSpec for tactical changes
```
