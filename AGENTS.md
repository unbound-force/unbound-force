# AGENTS.md

## Project Overview

Unbound Force is an organization of AI agent personas and roles for a software agent swarm, themed as a superhero team. Each hero is a repository in the [unbound-force](https://github.com/unbound-force) GitHub organization. This repo (`unbound-force/unbound-force`) is the meta/organizational repository -- it defines the team, the org-level constitution, the architectural specs for all heroes, and the shared standards that every hero repo must follow.

- **Type**: Meta repository (specifications, governance, standards)
- **Heroes**: Muti-Mind (PO), Cobalt-Crush (Dev), Gaze (Tester), The Divisor (Reviewer), Mx F (Manager)
- **Tooling**: [Speckit](https://github.com/github/spec-kit) + [OpenCode](https://opencode.ai) + [Swarm](https://www.swarmtools.ai/)
- **License**: Apache 2.0

## Core Mission

- **Strategic Architecture**: Engineers shift from manual coding to directing an "infinite supply of junior developers" (AI agents).
- **Outcome Orientation**: Focus on conveying business value and user intent rather than low-level technical sub-tasks.
- **Intent-to-Context**: Treat specs and rules as the medium through which human intent is manifested into code.

## Behavioral Constraints

- **Zero-Waste Mandate**: No orphaned specs, unused standards, or aspirational documents that do not map to actionable work.
- **Neighborhood Rule**: Changes to org-level standards must be audited for impact on all hero repos (Gaze, Website, future heroes).
- **Intent Drift Detection**: Specs must faithfully capture the product vision from `unbound-force.md`. Implementation specs must trace back to architectural specs.
- **Automated Governance**: The constitution and Hero Interface Contract are enforced through Constitution Check gates, not ad-hoc review.

## Constitution (Highest Authority)

The org constitution at `.specify/memory/constitution.md` (v1.0.0) defines three core principles that govern all hero repositories:

1. **I. Autonomous Collaboration**: Heroes communicate through well-defined artifacts (files, reports, schemas), not runtime coupling. Every hero completes its primary function independently. Outputs are self-describing.
2. **II. Composability First**: Every hero is independently installable and usable alone. Heroes expose extension points for integration. Combining heroes produces additive value without mandatory dependencies.
3. **III. Observable Quality**: Every hero produces machine-parseable output (JSON minimum). Artifacts include provenance metadata. Quality claims are backed by automated, reproducible evidence.

Hero constitutions extend (never contradict) the org constitution. See the constitution for the full MUST/SHOULD rules and governance model.

## The Heroes

| Hero | Role | Repo | Status |
|------|------|------|--------|
| Muti-Mind | Product Owner | `unbound-force/muti-mind` | Spec only (004) |
| Cobalt-Crush | Developer | `unbound-force/cobalt-crush` | Spec only (006) |
| Gaze | Tester | `unbound-force/gaze` | Implemented |
| The Divisor | PR Reviewer (Council) | `unbound-force/the-divisor` | Spec only (005) |
| Mx F | Manager | `unbound-force/mx-f` | Spec only (007) |

Gaze is the only hero with a functional implementation. The Divisor has a prototype deployment (reviewer agents) inside the Gaze repo.

## Project Structure

```text
unbound-force/
├── .specify/
│   ├── memory/
│   │   └── constitution.md          # Org constitution v1.0.0 (highest authority)
│   ├── templates/                   # Speckit templates (6 files)
│   │   ├── spec-template.md
│   │   ├── plan-template.md
│   │   ├── tasks-template.md
│   │   ├── checklist-template.md
│   │   ├── constitution-template.md
│   │   └── agent-file-template.md
│   ├── scripts/bash/               # Speckit automation scripts (5 files)
│   │   ├── common.sh
│   │   ├── check-prerequisites.sh
│   │   ├── setup-plan.sh
│   │   ├── create-new-feature.sh
│   │   └── update-agent-context.sh
│   └── config.yaml                  # Project-specific configuration
├── .opencode/
│   ├── agents/
│   │   ├── constitution-check.md    # Alignment checking agent (subagent)
│   │   ├── reviewer-guard.md        # The Guard: intent drift detection
│   │   ├── reviewer-architect.md    # The Architect: structural review
│   │   └── reviewer-adversary.md    # The Adversary: resilience audit
│   └── command/                     # Speckit pipeline commands + utilities
│       ├── speckit.constitution.md
│       ├── speckit.specify.md
│       ├── speckit.clarify.md
│       ├── speckit.plan.md
│       ├── speckit.tasks.md
│       ├── speckit.analyze.md
│       ├── speckit.checklist.md
│       ├── speckit.implement.md
│       ├── speckit.taskstoissues.md
│       └── constitution-check.md    # /constitution-check command
├── cmd/unbound/
│   └── main.go                      # Cobra CLI entry point
├── internal/scaffold/
│   ├── scaffold.go                  # Core scaffold engine
│   ├── scaffold_test.go             # Tests + drift detection
│   └── assets/                      # Embedded files (go:embed)
├── openspec/                        # OpenSpec tactical workflow
│   ├── specs/                       # Living behavior contracts
│   ├── changes/                     # Active tactical changes
│   ├── schemas/
│   │   └── unbound-force/           # Custom schema + templates
│   └── config.yaml                  # OpenSpec configuration
├── specs/                           # Architectural specifications
│   ├── 001-org-constitution/        # Org constitution ratification
│   ├── 002-hero-interface-contract/ # Standard hero repo structure & protocols
│   ├── 003-specification-framework/ # Specification framework (Speckit + OpenSpec)
│   ├── 004-muti-mind-architecture/  # Product Owner hero design
│   ├── 005-the-divisor-architecture/# PR Reviewer Council design
│   ├── 006-cobalt-crush-architecture/# Developer hero design
│   ├── 007-mx-f-architecture/       # Manager hero design
│   ├── 008-swarm-orchestration/     # End-to-end workflow & Swarm plugin
│   ├── 009-shared-data-model/       # JSON schemas for inter-hero artifacts
│   └── 010-knowledge-graph-integration/ # MCP knowledge graph via graphthulhu
├── schemas/                         # JSON Schemas for validation
│   ├── hero-manifest/
│   │   ├── v1.0.0.schema.json       # Hero manifest JSON Schema
│   │   └── samples/
│   │       └── gaze-hero.json       # Sample Gaze hero manifest
│   └── samples/
│       └── sample-quality-report-envelope.json  # Sample artifact envelope
├── scripts/
│   └── validate-hero-contract.sh    # Contract compliance validation
├── go.mod                           # Go module definition
├── opencode.json                    # MCP server configuration (knowledge graph)
├── .goreleaser.yaml                 # GoReleaser release configuration
├── unbound-force.md                 # Hero descriptions and team vision
├── AGENTS.md                        # This file
├── README.md
└── LICENSE
```

## Spec Organization

This repo contains **architectural design specs** that define each hero's capabilities, interfaces, and integration points. These are not implementation specs (those live in each hero's own repo). The specs are organized in three phases:

### Phase 0: Foundation (001-003)

Must be finalized before hero repos are built.

- **001-org-constitution**: Three core principles, governance, hero alignment
- **002-hero-interface-contract**: Standard repo structure, artifact envelope, naming conventions, hero manifest
- **003-specification-framework**: Unified specification framework (Speckit strategic + OpenSpec tactical), define extension points

### Phase 1: Hero Architectures (004-007)

Each hero's design. Can proceed in parallel once Phase 0 is done.

- **004-muti-mind-architecture**: AI persona, backlog CLI, priority scoring, GitHub sync, acceptance authority
- **005-the-divisor-architecture**: Three-persona review protocol, convention packs, deployment generator
- **006-cobalt-crush-architecture**: Dev persona, coding standards, Gaze/Divisor feedback loops
- **007-mx-f-architecture**: Metrics platform, coaching engine, impediment tracking, sprint management

### Phase 2: Cross-Cutting (008-009)

Depends on all Phase 1 specs.

- **008-swarm-orchestration**: Feature lifecycle workflow, artifact handoff, Swarm plugin, learning loop
- **009-shared-data-model**: JSON schemas for all artifact types, versioning, schema registry

### Phase 3: Infrastructure (010)

Meta-repo operational spec (not a hero architecture).

- **010-knowledge-graph-integration**: MCP knowledge graph via graphthulhu, Obsidian backend, spec search/traversal

### Dependency Graph

```text
Phase 0 (Foundation)
  001-org-constitution
    └─> 002-hero-interface-contract
    └─> 003-specification-framework

Phase 1 (Heroes) — all depend on 001 + 002
  004-muti-mind-architecture
  005-the-divisor-architecture
  006-cobalt-crush-architecture
  007-mx-f-architecture

Phase 2 (Cross-Cutting) — depends on Phase 1
  008-swarm-orchestration
  009-shared-data-model

Phase 3 (Infrastructure) — meta-repo operational
  010-knowledge-graph-integration
```

## Inter-Hero Artifact Types

These artifact types are defined across the specs for inter-hero communication:

| Artifact Type | Producer | Consumers | Spec |
|--------------|----------|-----------|------|
| `quality-report` | Gaze | Mx F, Muti-Mind, Cobalt-Crush | 009 |
| `review-verdict` | The Divisor | Mx F, Cobalt-Crush, Muti-Mind | 009 |
| `backlog-item` | Muti-Mind | Mx F, Cobalt-Crush | 009 |
| `acceptance-decision` | Muti-Mind | Mx F, Cobalt-Crush | 009 |
| `metrics-snapshot` | Mx F | Muti-Mind | 009 |
| `coaching-record` | Mx F | All heroes | 009 |
| `workflow-record` | Swarm Orchestration | Mx F, Muti-Mind | 009 |

All artifacts use the standard envelope format: `hero`, `version`, `timestamp`, `artifact_type`, `schema_version`, `context`, `payload`.

## Sibling Repositories

| Repo | Purpose | Constitution | Status |
|------|---------|-------------|--------|
| `unbound-force/gaze` | Go static analysis (tester hero) | v1.0.0 (Accuracy, Minimal Assumptions, Actionable Output) | Active, 5 specs complete |
| `unbound-force/website` | Public website (Hugo + Doks) | v1.0.0 (Content Accuracy, Minimal Footprint, Visitor Clarity) | Active, 1 spec complete |
| `unbound-force/homebrew-tap` | Homebrew formula distribution | N/A | Active |

## Specification Framework

This repo uses a unified two-tier specification framework
distributed via the `unbound` CLI binary. Install with
`brew install unbound-force/tap/unbound` and run
`unbound init` to scaffold into any repository.

### Dual-Tier Overview

| Tier | Tool | When to Use | Artifacts |
|------|------|-------------|-----------|
| Strategic | Speckit | 3+ stories, cross-repo, architecture | `specs/NNN-*/` |
| Tactical | OpenSpec | <3 stories, bug fix, maintenance | `openspec/` |

Both tiers share the org constitution as their governance
bridge. See the "Strategic vs Tactical" section for
selection criteria.

### Speckit Pipeline (Strategic -- Mandatory)

All non-trivial feature work **must** go through the
Speckit pipeline. The constitution
(`.specify/memory/constitution.md`) is the
highest-authority document -- all work must align with it.

The workflow is a strict, sequential pipeline:

```text
constitution → specify → clarify → plan → tasks
  → analyze → checklist → implement
```

| Phase | Command | Purpose | Prerequisites | Inputs | Outputs | Required? |
|-------|---------|---------|---------------|--------|---------|:---------:|
| 1 | `/speckit.constitution` | Create/update project constitution | None | User description | `.specify/memory/constitution.md` | Yes (once) |
| 2 | `/speckit.specify` | Create feature specification | Constitution | User description | `specs/NNN-*/spec.md` | Yes |
| 3 | `/speckit.clarify` | Reduce spec ambiguity | spec.md | spec.md | Updated spec.md | Recommended |
| 4 | `/speckit.plan` | Generate implementation plan | spec.md | spec.md, research | `plan.md`, `contracts/`, `data-model.md` | Yes |
| 5 | `/speckit.tasks` | Generate task list | plan.md | plan.md, spec.md | `tasks.md` | Yes |
| 6 | `/speckit.analyze` | Consistency analysis | tasks.md | All artifacts | Analysis report | Recommended |
| 7 | `/speckit.checklist` | Quality validation | spec.md | spec.md | `checklists/*.md` | Yes |
| 8 | `/speckit.implement` | Execute tasks | tasks.md, checklists | All artifacts | Implementation | Yes |
| 9 | `/speckit.taskstoissues` | Convert to GitHub Issues | tasks.md | tasks.md | GitHub Issues | Optional |

### OpenSpec Workflow (Tactical)

For small changes: bug fixes, minor enhancements (<3
stories), maintenance tasks, single-repo refactoring.

Requires Node.js >= 20.19.0 and the OpenSpec CLI:
`npm install -g @fission-ai/openspec@latest`

| Action | Command | Purpose | Prerequisites | Inputs | Outputs |
|--------|---------|---------|---------------|--------|---------|
| Propose | `/opsx:propose` | Create change proposal + plan | Constitution | User description | `openspec/changes/*/proposal.md` |
| Explore | `/opsx:explore` | Analyze specs and design | proposal.md | proposal.md | `specs/*.md`, `design.md` |
| Apply | `/opsx:apply` | Implement from task list | tasks.md | All artifacts | Implementation |
| Archive | `/opsx:archive` | Archive completed change | Completed tasks | Change directory | Archived change |

### Ordering Constraints

1. Constitution must exist before specs.
2. Spec must exist before plan.
3. Plan must exist before tasks.
4. Tasks must exist before implementation and analysis.
5. Clarify should run before plan (skipping increases rework risk).
6. Analyze should run after tasks but before implementation.
7. All checklists must pass before implementation (or user must explicitly override).

### Task Completion Bookkeeping

When a task from `tasks.md` is completed during implementation, its checkbox **must** be updated from `- [ ]` to `- [x]` immediately. Do not defer this -- mark tasks complete as they are finished, not in a batch after all work is done.

### Documentation Validation Gate

Before marking any task complete, validate whether the change requires documentation updates:

- `README.md` -- project description changes
- `AGENTS.md` -- new specs, changed hero status, updated conventions
- `unbound-force.md` -- hero description changes
- Spec artifacts under `specs/` -- if the change affects planned behavior
- `.specify/memory/constitution.md` -- if the change affects governance

A task is not complete until its documentation impact has been assessed and any necessary updates have been made.

### Spec Commit Gate

All spec artifacts (`spec.md`, `plan.md`, `tasks.md`, and any other files under `specs/`) **must** be committed and pushed before implementation begins. Run `/speckit.implement` only after the spec commit is on the remote.

### Constitution Check

A mandatory gate at the planning phase. The constitution's three core principles -- Autonomous Collaboration, Composability First, and Observable Quality -- must each receive a PASS before proceeding. Constitution violations are automatically CRITICAL severity and non-negotiable.

For hero constitution alignment validation, use the `/constitution-check` command. This invokes a dedicated OpenCode agent that compares a hero constitution against the org constitution and produces a structured alignment report with per-principle findings and an overall ALIGNED/NON-ALIGNED verdict. See `.opencode/agents/constitution-check.md` and `.opencode/command/constitution-check.md` for implementation details.

## Strategic vs Tactical: Boundary Guidelines

This project uses a two-tier specification framework:

- **Speckit** (strategic): Full pipeline for architectural
  work. Specs live under `specs/` as numbered directories.
- **OpenSpec** (tactical): Lightweight workflow for small
  changes. Artifacts live under `openspec/specs/` and
  `openspec/changes/`.

### Decision Criteria Matrix

| Criterion | Speckit (Strategic) | OpenSpec (Tactical) |
|-----------|:------------------:|:-------------------:|
| User stories | >= 3 | < 3 |
| Cross-repo impact | Yes | No |
| Constitution changes | Always | Never |
| New hero architecture | Always | Never |
| New inter-hero artifact types | Always | Never |
| Bug fix | Never | Always |
| Single-repo maintenance | Never | Always |
| Refactoring (non-architectural) | Rarely | Usually |

### Default Heuristic

When in doubt, start with OpenSpec. If the scope grows
beyond 3 stories or crosses repo boundaries, escalate to
Speckit by extracting the proposal into a new numbered
spec directory under `specs/`.

### Escalation Path

1. Start with `/opsx:propose` for the initial change.
2. During exploration, if the scope expands beyond 3 user
   stories or affects multiple repositories, stop.
3. Run `/speckit.specify` to create a full spec under
   `specs/NNN-feature-name/`.
4. Archive the OpenSpec proposal with `/opsx:archive`.
5. Continue with the Speckit pipeline from the new spec.

### Directory Boundary Enforcement

These boundaries are enforced by convention and code
review, not automated gates (v1.0.0):

- OpenSpec changes MUST NOT modify files under `specs/`.
  OpenSpec artifacts belong exclusively in `openspec/`.
- Speckit specs MUST NOT be created under `openspec/`.
  Strategic specs belong exclusively in `specs/NNN-*/`.

**Violation examples**: If a developer attempts to create
a Speckit-style numbered spec directory under `openspec/`,
or an OpenSpec delta spec under `specs/`, the expected
outcome is code review rejection per this documented
convention. The reviewer SHOULD point the author to this
section and recommend the correct directory.

## Architecture

Single binary CLI with layered internal packages:

```text
cmd/unbound/           CLI layer (Cobra commands, flag handling)
internal/
  scaffold/            Core scaffold engine (embed.FS, file ownership, version markers)
```

All business logic lives under `internal/` and MUST NOT be imported externally by other repositories.

### Key Patterns

- **Scaffold pattern (Gaze-derived)**: Configurable behavior uses `Options`/`Result` structs, a core `Run()` function, file ownership classification (`isToolOwned`), and version markers (`insertMarkerAfterFrontmatter`).
- **Testable CLI pattern**: Commands delegate to `runXxx(params)` functions. Params structs include `io.Writer` for stdout/stderr, enabling unit testing without subprocess execution or `os.Stdout` mocking.

## Coding Conventions

- **Formatting**: `gofmt` and `goimports` (enforced by golangci-lint).
- **Naming**: Standard Go conventions. PascalCase for exported, camelCase for unexported.
- **Comments**: GoDoc-style comments on all exported functions and types.
- **Error handling**: Return `error` values. Wrap with `fmt.Errorf("context: %w", err)`.
- **Import grouping**: Standard library, then third-party, then internal packages (separated by blank lines).
- **No global state**: Prefer functional style and dependency injection.
- **Logging**: Use `github.com/charmbracelet/log` for all application logging. Avoid standard library `log` or `fmt.Println` for operational logs.
- **CLI Framework**: Use `github.com/spf13/cobra` for command routing and flag parsing.

## Testing Conventions

- **Framework**: Standard library `testing` package only. No testify, gomega, or other external assertion libraries.
- **Assertions**: Use `t.Errorf` / `t.Fatalf` directly. No assertion helpers from third-party packages.
- **Test naming**: `TestXxx_Description` (e.g., `TestRun_CreatesFiles`, `TestIsToolOwned_ToolFiles`).
- **Test isolation**: Use `t.TempDir()` for all filesystem tests. No shared mutable state between tests.
- **Drift detection**: Tests MUST exist to ensure embedded assets (`internal/scaffold/assets`) perfectly match their canonical sources.

## Build & Test Commands

```bash
# Build
make build
# or: go build ./...

# Run tests
make test
# or: go test -race -count=1 ./...

# Run all checks (build, test, vet, lint)
make check

# Lint
golangci-lint run
```

Always run tests with `-race -count=1`. CI enforces this.

## Writing Style for Specs

This repo is primarily specifications and governance documents. Follow these conventions:

- **RFC-style language**: Use MUST, SHOULD, MAY, MUST NOT per RFC 2119 semantics in all requirement statements.
- **Acceptance scenarios**: Use Given/When/Then format for all acceptance criteria.
- **Functional requirements**: Number as FR-NNN with MUST/SHOULD/MAY prefix.
- **Success criteria**: Number as SC-NNN with measurable, testable outcomes.
- **User stories**: Prioritize as P1, P2, P3. Each must be independently testable.
- **Cross-references**: Reference other specs by number (e.g., "per Spec 002").
- **Line length**: Keep prose under 72 characters for readability in terminals and diffs.

## Git & Workflow

- **Commit format**: Conventional Commits -- `type: description` (e.g., `feat:`, `fix:`, `docs:`, `chore:`, `refactor:`).
- **Branching**: Feature branches required. No direct commits to `main` except trivial doc fixes.
- **Code review**: Required before merge.
- **Semantic versioning**: For releases and constitution amendments.

## Active Technologies
- Go 1.24+ (unbound CLI binary, Cobra CLI framework, embed.FS scaffold)
- GoReleaser v2 (cross-platform release pipeline, Homebrew cask publishing)
- Markdown (specifications, governance, templates, commands)
- YAML (OpenSpec schema, configuration files)
- Bash (speckit scripts)
- JSON Schema draft 2020-12 (hero manifest schema, sample artifact envelope)
- OpenCode + Speckit + OpenSpec (development workflow)
- Node.js >= 20.19.0 (OpenSpec CLI, `@fission-ai/openspec`)

## Recent Changes

- 003-specification-framework: Implemented unified two-tier specification framework -- `unbound` CLI binary (Go + Cobra + embed.FS), scaffolds 33 files (22 Speckit + 6 OpenSpec + 5 agents) via `unbound init`, custom `unbound-force` OpenSpec schema with constitution alignment in proposals, boundary guidelines documented, GoReleaser v2 release pipeline, all 8 user stories (US1-US8) and 67 tasks completed, format-aware version markers, drift detection tests
- 010-knowledge-graph-integration: Implemented knowledge graph integration -- graphthulhu installed with `--include-hidden` support (upstream PR submitted), OpenCode MCP configuration created, all 5 user stories verified (search, analysis, live sync, link traversal, property queries), YAML frontmatter added to specs 001/002/004, 50 pages indexed across all directories including `.specify/` and `.opencode/`
- 002-hero-interface-contract: Completed spec implementation -- Hero Interface Contract v1.0.0 ratified, hero manifest JSON Schema created and validated, contract compliance validation script created and tested against Gaze and Website repos, sample artifact envelope and Gaze manifest produced, all FRs (001-015) and SCs (001-007) validated, spec status set to Complete
- 001-org-constitution: Completed spec implementation -- constitution ratified v1.0.0, alignment agent and `/constitution-check` command created, all FRs and SCs validated, spec status set to Complete
- 001-org-constitution: Ratified org constitution v1.0.0 with three principles (Autonomous Collaboration, Composability First, Observable Quality)
- Specs 001-009: Added architectural design specs for all heroes (Muti-Mind, The Divisor, Cobalt-Crush, Mx F), infrastructure (org constitution, hero interface contract, speckit framework), and cross-cutting concerns (swarm orchestration, shared data model)
