## ADDED Requirements

### Requirement: /agent-brief Command (Create Mode)

The `/agent-brief` command MUST auto-detect create mode when no AGENTS.md
file exists at the repository root. In create mode, the command MUST scan
project files to determine language, build system, CI configuration,
directory structure, and spec-framework presence. The command MUST generate
AGENTS.md with Tier 1 sections (Project Overview, Build & Test Commands,
Project Structure, Code Conventions, Technology Stack) filled from the
project analysis. The command MUST generate Tier 1C sections (Constitution,
Spec Framework) when their trigger directories are detected. The command
MUST stub Tier 2 sections (Architecture, Testing Conventions, Git &
Workflow, Behavioral Constraints) with TODO guidance comments. The command
MUST present the generated content to the user before writing the file.

#### Scenario: Create AGENTS.md for Go project with speckit

- **GIVEN** a repository with `go.mod`, `Makefile`, `.specify/memory/constitution.md`, and `specs/001-feature/` but no AGENTS.md
- **WHEN** the user runs `/agent-brief`
- **THEN** the command creates AGENTS.md with all 5 Tier 1 sections filled from project analysis, a Constitution section summarizing the 4 principles, a Spec Framework section describing the speckit pipeline, and 4 Tier 2 sections with TODO stubs

#### Scenario: Create AGENTS.md for TypeScript project without spec framework

- **GIVEN** a repository with `package.json`, `tsconfig.json`, `.github/workflows/ci.yml` but no `.specify/` or `openspec/` directories and no AGENTS.md
- **WHEN** the user runs `/agent-brief`
- **THEN** the command creates AGENTS.md with all 5 Tier 1 sections filled, no Tier 1C sections (governance/spec sections omitted), and 4 Tier 2 stubs

#### Scenario: Force create mode when file exists

- **GIVEN** a repository with an existing AGENTS.md
- **WHEN** the user runs `/agent-brief create`
- **THEN** the command enters create mode, analyzes the project, generates new content, and presents it to the user for confirmation before overwriting

### Requirement: /agent-brief Command (Audit Mode)

The `/agent-brief` command MUST auto-detect audit mode when AGENTS.md
exists at the repository root. In audit mode, the command MUST read the
file and score it against the context-sensitive section taxonomy. The
command MUST report per-section status (present/absent/quality), tier
completion metrics, quality metrics (line count, code blocks in build
section), and an overall effectiveness label. The command MUST detect
staleness by comparing the Project Structure tree against actual
directories. The command MUST offer improvement suggestions with specific
content for missing sections. The command SHOULD offer to apply
improvements with user confirmation.

#### Scenario: Audit a complete AGENTS.md

- **GIVEN** a repository with AGENTS.md containing all Tier 1 and Tier 2 sections, under 300 lines
- **WHEN** the user runs `/agent-brief`
- **THEN** the command produces an audit report showing all sections present, quality metrics within thresholds, and overall label "Excellent"

#### Scenario: Audit AGENTS.md with missing sections

- **GIVEN** a repository with AGENTS.md missing the Testing Conventions section and the constitution reference, and `.specify/memory/constitution.md` exists
- **WHEN** the user runs `/agent-brief`
- **THEN** the command reports Testing Conventions as missing (Tier 2), Constitution reference as missing (Tier 1C), overall label downgraded, and generates specific improvement suggestions for both sections

#### Scenario: Detect stale project structure

- **GIVEN** AGENTS.md contains a Project Structure tree referencing `internal/foo/` which no longer exists on the filesystem
- **WHEN** the user runs `/agent-brief audit`
- **THEN** the command flags the Project Structure section as potentially stale with a suggestion to update or remove the reference

#### Scenario: Detect verbose AGENTS.md

- **GIVEN** AGENTS.md is 450 lines long
- **WHEN** the user runs `/agent-brief`
- **THEN** the audit report includes a line count warning suggesting condensing procedural content into skills or commands

### Requirement: /agent-brief Command (Bridge Files)

After creating or improving AGENTS.md, the command MUST verify that
cross-tool bridge files exist. If CLAUDE.md does not exist or does not
contain an `@AGENTS.md` import directive, the command MUST create or update
it. If `.cursorrules` does not exist or does not reference AGENTS.md, the
command MUST create or update it. Bridge file creation MUST follow the
existing patterns established by `uf init` (`ensureCLAUDEmd()` and
`ensureCursorrules()`).

#### Scenario: Create bridge files alongside new AGENTS.md

- **GIVEN** a repository with no AGENTS.md, no CLAUDE.md, no .cursorrules
- **WHEN** the user runs `/agent-brief` and confirms the generated content
- **THEN** the command writes AGENTS.md, creates CLAUDE.md with `@AGENTS.md` import, and creates .cursorrules with AGENTS.md reference

#### Scenario: Bridge files already exist

- **GIVEN** a repository with AGENTS.md, CLAUDE.md (containing `@AGENTS.md`), and .cursorrules (referencing AGENTS.md)
- **WHEN** the user runs `/agent-brief` in audit mode
- **THEN** the bridge file checks report "present" with no modifications

### Requirement: /agent-brief Command (Cross-Framework Governance)

When `.specify/memory/constitution.md` exists, the command MUST include a
Constitution section that summarizes the principles and states their
cross-framework applicability. When both `specs/` (Speckit) and `openspec/`
(OpenSpec) directories exist, the command MUST include a Spec Framework
section that describes both frameworks and explicitly states the
constitution governs both. The audit mode MUST check for this cross-
framework governance bridge and flag its absence as a warning.

#### Scenario: Both spec frameworks present without governance bridge

- **GIVEN** AGENTS.md describes Speckit and OpenSpec separately but does not state the constitution applies to both, and both framework directories exist
- **WHEN** the user runs `/agent-brief`
- **THEN** the audit flags the missing cross-framework governance bridge and suggests adding "Both tiers share the org constitution as their governance bridge"

#### Scenario: Only one spec framework present

- **GIVEN** a repository with `specs/` but no `openspec/` directory
- **WHEN** the user runs `/agent-brief`
- **THEN** the spec framework section describes Speckit only, and the cross-framework bridge check is skipped

### Requirement: Doctor "Agent Context" Check Group

The `uf doctor` command MUST include a new "Agent Context" check group that
replaces the existing AGENTS.md existence check in `checkScaffoldedFiles()`.
The check group MUST contain 12 checks covering file existence, Tier 1
section presence (5 checks), build section code blocks, line count,
constitution reference (context-sensitive), spec framework description
(context-sensitive), and bridge file validation (2 checks). Context-
sensitive checks MUST be skipped when their trigger directories do not
exist. The install hint for a missing AGENTS.md MUST be changed from
"Run: uf init" to "Run: /agent-brief in OpenCode". All checks MUST use
the injectable `Options` pattern (`opts.ReadFile`) for test isolation.

#### Scenario: Doctor reports full agent context checks

- **GIVEN** a project with AGENTS.md containing all Tier 1 sections, `.specify/memory/constitution.md`, CLAUDE.md with @AGENTS.md, and .cursorrules
- **WHEN** the user runs `uf doctor`
- **THEN** the "Agent Context" group shows 12 checks, all Pass, including constitution reference and spec framework description

#### Scenario: Doctor skips context-sensitive checks when not applicable

- **GIVEN** a project with AGENTS.md but no `.specify/` or `openspec/` directories
- **WHEN** the user runs `uf doctor`
- **THEN** the "Agent Context" group shows 10 checks (constitution and spec framework checks omitted), and only Tier 1 sections, code blocks, line count, and bridge files are validated

#### Scenario: Doctor reports missing AGENTS.md with correct hint

- **GIVEN** a project with no AGENTS.md
- **WHEN** the user runs `uf doctor`
- **THEN** the "Agent Context" group shows a single Fail result for "AGENTS.md" with InstallHint "Run: /agent-brief in OpenCode"

#### Scenario: Doctor reports missing Tier 1 section

- **GIVEN** a project with AGENTS.md that has no Build & Test Commands section
- **WHEN** the user runs `uf doctor`
- **THEN** the "Agent Context" group shows Fail for "Tier 1: Build Commands" with a message indicating the section was not found

### Requirement: Scaffold Asset Distribution

The `/agent-brief` command file MUST be embedded as a scaffold asset at
`internal/scaffold/assets/opencode/command/agent-brief.md`. The file MUST
be registered in `expectedAssetPaths` in the scaffold test. The file MUST
be classified as tool-owned by `isToolOwned()`. The `uf init` command MUST
deploy the file to `.opencode/command/agent-brief.md` in target projects.

#### Scenario: New project receives /agent-brief via uf init

- **GIVEN** a new project directory with no `.opencode/` directory
- **WHEN** the user runs `uf init`
- **THEN** `.opencode/command/agent-brief.md` is created among the scaffolded files

#### Scenario: Scaffold asset drift detection

- **GIVEN** the scaffold asset at `internal/scaffold/assets/opencode/command/agent-brief.md`
- **WHEN** the scaffold drift detection test runs
- **THEN** the asset is included in the expected paths and the live copy at `.opencode/command/agent-brief.md` matches the embedded asset

## MODIFIED Requirements

### Requirement: checkScaffoldedFiles AGENTS.md Check Removal

The AGENTS.md existence check MUST be removed from `checkScaffoldedFiles()`
in `internal/doctor/checks.go`. Previously, this function checked for
AGENTS.md existence alongside other scaffolded files. The check moves to
the new `checkAgentContext()` function with deeper validation.

(Previously: `checkScaffoldedFiles()` at checks.go:483-498 contained a
simple `os.Stat()` existence check for AGENTS.md with InstallHint
"Run: uf init".)

## REMOVED Requirements

None.
