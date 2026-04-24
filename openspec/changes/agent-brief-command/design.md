## Context

AGENTS.md is the primary context document that all AI coding agents consume at
session start. The uf ecosystem modifies this file through multiple components
(`uf init`, `/uf-init`, Divisor review agents) but never creates it. This leaves
a bootstrap gap for new adopters and no ongoing quality validation beyond manual
code review.

The proposal establishes `/agent-brief` as the lifecycle command for AGENTS.md
(create, audit, improve) and a new doctor check group for structural validation.

This design covers two deliverables:
1. An LLM-driven OpenCode command (Markdown instruction file)
2. Deterministic Go checks in the doctor binary

## Goals / Non-Goals

### Goals
- Close the AGENTS.md bootstrap gap: any project can generate an effective
  AGENTS.md from project analysis
- Provide ongoing quality feedback: audit mode scores the file against a
  context-sensitive section taxonomy
- Surface spec-framework governance gaps: detect when constitution or
  cross-framework references are missing
- Integrate structural checks into `uf doctor` for CI-compatible validation
- Distribute the command via scaffold so every `uf init` project receives it

### Non-Goals
- Replacing `/uf-init` guidance blocks: `/agent-brief` owns project context
  (build commands, structure, conventions), `/uf-init` owns uf-specific
  behavioral guidance (gatekeeping, phase boundaries, CI parity)
- Enforcing a rigid template: AGENTS.md must accommodate project-specific
  sections; the command validates presence of essential sections, not exact
  structure
- Content quality analysis in doctor: the Go binary checks structural signals
  (headers, code blocks, line count); semantic quality analysis is the LLM
  command's job
- Automating AGENTS.md updates on every spec change: the Divisor agents
  (Guard, Curator, Scribe) handle that during code review

## Decisions

### D1: Context-Sensitive Section Taxonomy

**Decision**: The section taxonomy is not a static checklist. Section
requirements are promoted based on detected project characteristics.

**Rationale**: A Go CLI project with speckit and openspec should require
constitution and spec-framework sections, but a simple npm package should
not. Static requirements would produce false negatives for simple projects
and false positives for complex ones.

**Tier structure**:

| Tier | Sections | When Required |
|------|----------|---------------|
| Tier 1 (Essential) | Project Overview, Build & Test Commands, Project Structure, Code Conventions, Technology Stack | Always |
| Tier 1C (Contextual) | Constitution / Governance, Spec Framework | When `.specify/memory/`, `specs/`, or `openspec/` detected |
| Tier 2 (Recommended) | Architecture, Testing Conventions, Git & Workflow, Behavioral Constraints | Always (warn if missing) |
| Tier 3 (Advanced) | Knowledge Retrieval, Recent Changes, Convention Packs | Informational only |

Constitution alignment: Composability First (Principle II) -- the command
adapts to each project's complexity rather than imposing a one-size-fits-all
structure.

### D2: Hybrid Creation Approach (Template Skeleton + LLM Analysis)

**Decision**: Create mode uses a fixed Markdown skeleton with section headers
and guidance comments, then the LLM fills Tier 1 sections by reading project
files (README, go.mod/package.json, Makefile, CI config, directory listing).
Tier 2 sections get stubs with TODO guidance. Tier 1C sections are generated
only when their triggers are detected.

**Rationale**: Pure template approach leaves too much manual work. Pure LLM
approach is slow and model-dependent. The hybrid gets the most impactful
content (build commands, tech stack) filled automatically while leaving
project-specific nuance (architecture patterns, behavioral constraints) for
the human to complete.

**Scan order for project analysis**:
1. `README.md` -- project description
2. `go.mod` / `package.json` / `Cargo.toml` / `pyproject.toml` -- language,
   module name, dependencies
3. `Makefile` / `justfile` -- build/test/lint commands
4. `.github/workflows/` -- CI commands (source of truth)
5. Directory tree (top-level only) -- project layout
6. `.golangci.yml` / `ruff.toml` / `.eslintrc` -- linter config
7. `.specify/memory/constitution.md` -- governance (Tier 1C trigger)
8. `specs/` / `openspec/` -- spec framework presence (Tier 1C trigger)
9. `.opencode/uf/packs/` -- convention packs (Tier 3 trigger)
10. `opencode.json` -- MCP server configuration (Tier 3 trigger)
11. `.git/config` -- remote URL for project name
12. `LICENSE` -- license type

### D3: Audit Scoring System

**Decision**: Use tier completion as the primary metric with a summary label.

**Labels**:
- **Excellent**: 5/5 Tier 1 + 4/4 Tier 2 (+ all Tier 1C if applicable)
- **Strong**: 5/5 Tier 1 + 2-3/4 Tier 2
- **Adequate**: 4-5/5 Tier 1
- **Weak**: 2-3/5 Tier 1
- **Missing**: 0-1/5 Tier 1

**Quality metrics** (supplemental):
- Line count: flag if >300, suggest condensing if >500
- Build commands: check for code blocks (fenced ``` blocks) in the build
  section
- Staleness: check if directories in the Project Structure tree still exist
- Bridge files: verify CLAUDE.md imports AGENTS.md and .cursorrules
  references it

Constitution alignment: Observable Quality (Principle III) -- the scoring
system produces quantifiable, reproducible metrics.

### D4: Doctor Check Group Architecture

**Decision**: Create a new `checkAgentContext(opts *Options) CheckGroup`
function that replaces the existing AGENTS.md existence check in
`checkScaffoldedFiles()`. The new function reads the file content and
performs regex-based section detection.

**Check inventory** (12 checks):

| # | Check | Detection | Severity |
|---|-------|-----------|----------|
| 1 | File exists | `os.Stat()` | Fail |
| 2 | Tier 1: Project Overview | Regex: `(?i)^##\s+.*overview\|about` | Fail |
| 3 | Tier 1: Build Commands | Regex: `(?i)^##\s+.*build` | Fail |
| 4 | Tier 1: Project Structure | Regex: `(?i)^##\s+.*structure\|layout` | Fail |
| 5 | Tier 1: Code Conventions | Regex: `(?i)^##\s+.*convention\|coding.*standard\|style` | Fail |
| 6 | Tier 1: Technology Stack | Regex: `(?i)^##\s+.*technolog\|tech.*stack\|stack` | Fail |
| 7 | Build has code blocks | Regex: triple-backtick within build section | Warn |
| 8 | Line count | `bytes.Count(content, '\n')` | Warn if >300 |
| 9 | Constitution reference | Scan for `constitution` near `.specify/memory/` (only when `.specify/memory/constitution.md` exists) | Warn |
| 10 | Spec framework described | Scan for `speckit`/`openspec`/`spec.*framework` (only when `specs/` or `openspec/` exists) | Warn |
| 11 | Bridge: CLAUDE.md | Read file, check for `@AGENTS.md` or `AGENTS.md` | Warn |
| 12 | Bridge: .cursorrules | Read file, check for `AGENTS.md` | Warn |

**Injectable dependencies**: The function uses `opts.ReadFile()` for all file
reads, enabling test isolation per Constitution Principle IV (Testability).

**Context-sensitive checks**: Checks 9-10 are skipped entirely (not shown in
output) when their trigger directories do not exist. This prevents false
negatives for projects that do not use spec-driven development.

Constitution alignment: Testability (Principle IV) -- all checks use the
injectable `Options` pattern, enabling unit tests without filesystem access.

### D5: Command Placement and Distribution

**Decision**: The command file lives at
`internal/scaffold/assets/opencode/command/agent-brief.md` (scaffold asset,
tool-owned) and deploys to `.opencode/command/agent-brief.md` in target
projects via `uf init`.

**Rationale**: This command has cross-project utility -- every project
benefits from AGENTS.md lifecycle management. Scaffold distribution ensures
it is available wherever uf is used, matching the pattern of other cross-
project commands (`review-council.md`, `unleash.md`, `uf-init.md`).

### D6: Responsibility Boundary with /uf-init

**Decision**: `/agent-brief` owns project context (Tier 1 + Tier 2 sections:
build commands, project structure, conventions, tech stack, architecture,
testing, git workflow). `/uf-init` owns uf-specific behavioral guidance
(the 8 guidance blocks: gatekeeping, phase boundaries, CI parity, review
council, spec-first, website gate, knowledge retrieval, core mission).

**Interaction flow**:
- New project: `/agent-brief` creates AGENTS.md → `uf init` appends
  Convention Packs → `/uf-init` injects behavioral guidance
- Existing project: `/agent-brief audit` validates and suggests improvements

**Rationale**: Keeping responsibilities separate prevents both commands from
fighting over the same sections. `/agent-brief` is project-generic (works
on any codebase), while `/uf-init` is uf-specific (only relevant for uf-
managed projects).

### D7: Section Header Detection Pattern Strategy

**Decision**: Use flexible regex patterns that match common section naming
conventions across projects. Each required section has 2-4 detection
patterns covering likely header variations.

**Example for Build Commands**:
`(?i)^##\s+(build|build\s*[&+]\s*test|development|build.*command)`

This matches:
- `## Build`
- `## Build & Test Commands`
- `## Development`
- `## Build Commands`

**Rationale**: Projects name their sections differently. Rigid exact-match
would produce false negatives. Overly broad patterns would produce false
positives. The selected patterns cover >95% of observed naming conventions
without matching unrelated sections.

### D8: Cross-Framework Governance Bridge

**Decision**: When both `specs/` (Speckit) and `openspec/` (OpenSpec)
directories exist, the audit checks for an explicit statement that the
constitution governs both frameworks. A project might describe both
frameworks separately but never state that the constitution applies to
OpenSpec work -- this is a real governance gap.

**Detection**: Scan for co-occurrence of `constitution` with
`both`/`all`/`regardless`/`openspec` within the spec framework section.

**Severity**: Warn -- the information is strongly recommended but the
project may have legitimate reasons for framework-specific governance.

## Risks / Trade-offs

### R1: Regex-Based Section Detection May Produce False Matches

**Risk**: A section titled `## Project Overview and API Structure` would
match both Project Overview and Project Structure patterns.

**Mitigation**: Patterns are ordered and first-match wins. The detection
function scans line-by-line and assigns each section header to at most one
category. Overlap risk is low because the patterns target distinct keywords.

### R2: LLM-Generated Content Quality Varies by Model

**Risk**: The create mode depends on LLM analysis of project files. Lower-
capability models may produce less accurate build commands or miss nuances.

**Mitigation**: The hybrid approach limits LLM responsibility to Tier 1
sections where the source data is structured (go.mod, Makefile, CI YAML).
Tier 2 sections use guidance stubs that the human fills. The user reviews
all generated content before it is written.

### R3: Doctor Check Group Increases Output Length

**Risk**: Adding 12 checks to doctor output may overwhelm users with
information.

**Mitigation**: Context-sensitive checks (9-12) are skipped when not
applicable, reducing typical output to 8-10 checks. The check group
follows the same formatting as existing groups and collapses visually
when all checks pass (single-line summary in the terminal formatter).

### R4: Command Size (~500 Lines)

**Risk**: Large command files consume more context when loaded.

**Mitigation**: The command is only loaded when explicitly invoked via
`/agent-brief`. It does not load at session start. This matches the
pattern of other large commands (`uf-init.md` is 1026 lines,
`unleash.md` is 664 lines).
