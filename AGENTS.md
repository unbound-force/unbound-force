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

The org constitution at `.specify/memory/constitution.md` defines four core principles that govern all hero repositories:

1. **I. Autonomous Collaboration**: Heroes communicate through well-defined artifacts (files, reports, schemas), not runtime coupling. Every hero completes its primary function independently. Outputs are self-describing.
2. **II. Composability First**: Every hero is independently installable and usable alone. Heroes expose extension points for integration. Combining heroes produces additive value without mandatory dependencies.
3. **III. Observable Quality**: Every hero produces machine-parseable output (JSON minimum). Artifacts include provenance metadata. Quality claims are backed by automated, reproducible evidence.
4. **IV. Testability**: Every component MUST be testable in isolation without requiring external services or shared mutable state.

Hero constitutions extend (never contradict) the org constitution. See the constitution for the full MUST/SHOULD rules and governance model.

## The Heroes

| Hero | Role | Repo | Status |
|------|------|------|--------|
| Muti-Mind | Product Owner | `unbound-force/muti-mind` | Implemented (Spec 004) |
| Cobalt-Crush | Developer | Embedded in `unbound-force` binary | Implemented (Spec 006) |
| Gaze | Tester | `unbound-force/gaze` | Implemented |
| The Divisor | PR Reviewer (Council) | Embedded in `unbound-force` binary | Implemented (Spec 005) |
| Mx F | Manager | `cmd/mxf/` + OpenCode agent | Implemented (Spec 007) |

All five heroes have implementations. Gaze has the most mature standalone implementation (Go CLI + static analysis engine). Muti-Mind has a backend CLI (`cmd/mutimind/`) and OpenCode agent. The Divisor has 5 review personas and convention packs (embedded in the `unbound-force` binary). Cobalt-Crush has a developer persona agent. Mx F has a full CLI backend (`cmd/mxf/`) with 7 subcommands and a coaching agent.

## Project Structure

```text
unbound-force/
├── .specify/
│   ├── memory/
│   │   └── constitution.md          # Org constitution (highest authority)
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
│   │   ├── cobalt-crush-dev.md       # Developer persona (Cobalt-Crush)
│   │   ├── constitution-check.md    # Alignment checking agent (subagent)
│   │   ├── mx-f-coach.md            # Coaching persona (Mx F)
│   │   ├── divisor-adversary.md     # The Adversary: resilience/security (Divisor)
│   │   ├── divisor-architect.md     # The Architect: structural review (Divisor)
│   │   ├── divisor-guard.md         # The Guard: intent drift detection (Divisor)
│   │   ├── divisor-sre.md           # The Operator: operational readiness (Divisor)
│   │   ├── divisor-testing.md       # The Tester: test quality (Divisor)
│   │   ├── gaze-reporter.md         # Gaze report agent
│   │   ├── muti-mind-po.md          # Muti-Mind Product Owner agent
│   │   ├── reviewer-adversary.md    # Legacy reviewer (superseded by divisor-*)
│   │   ├── reviewer-architect.md    # Legacy reviewer (superseded by divisor-*)
│   │   ├── reviewer-guard.md        # Legacy reviewer (superseded by divisor-*)
│   │   ├── reviewer-sre.md          # Legacy reviewer (superseded by divisor-*)
│   │   └── reviewer-testing.md      # Legacy reviewer (superseded by divisor-*)
│   ├── command/                     # Speckit pipeline commands + utilities
│   │   ├── speckit.constitution.md
│   │   ├── speckit.specify.md
│   │   ├── speckit.clarify.md
│   │   ├── speckit.plan.md
│   │   ├── speckit.tasks.md
│   │   ├── speckit.analyze.md
│   │   ├── speckit.checklist.md
│   │   ├── speckit.implement.md
│   │   ├── speckit.taskstoissues.md
│   │   ├── review-council.md        # /review-council command (Divisor)
│   │   ├── constitution-check.md    # /constitution-check command
│   │   ├── unleash.md               # /unleash autonomous pipeline (Spec 018)
│   │   ├── workflow-start.md        # /workflow start command (Spec 008)
│   │   ├── workflow-status.md       # /workflow status command (Spec 008)
│   │   ├── workflow-list.md         # /workflow list command (Spec 008)
│   │   └── workflow-advance.md      # /workflow advance command (Spec 008)
│   ├── skill/                       # Swarm skills packages
│   │   └── unbound-force-heroes/
│   │       └── SKILL.md             # Hero roles, routing, workflow stages (Spec 008)
│   └── unbound/
│       └── packs/                   # Convention packs (shared by all heroes)
│           ├── go.md                # Go convention pack (tool-owned)
│           ├── go-custom.md         # Go custom rules (user-owned)
│           ├── default.md           # Language-agnostic default (tool-owned)
│           ├── default-custom.md    # Default custom rules (user-owned)
│           ├── typescript.md        # TypeScript convention pack (tool-owned)
│           └── typescript-custom.md # TypeScript custom rules (user-owned)
├── cmd/unbound-force/
│   └── main.go                      # Cobra CLI entry point
├── cmd/mutimind/
│   └── main.go                      # Muti-Mind backend CLI entry point
├── internal/scaffold/
│   ├── scaffold.go                  # Core scaffold engine
│   ├── scaffold_test.go             # Tests + drift detection
│   └── assets/                      # Embedded files (go:embed)
├── internal/backlog/                # Muti-Mind local backlog parsing
├── internal/sync/                   # Muti-Mind GitHub issue sync
├── internal/artifacts/              # Artifact envelope I/O (WriteArtifact, ReadEnvelope, FindArtifacts)
├── internal/schemas/               # JSON Schema generation, validation, convention pack validation (Spec 009)
├── internal/orchestration/          # Swarm orchestration engine (Spec 008)
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
│   ├── 010-knowledge-graph-integration/ # MCP knowledge graph via graphthulhu
│   ├── 011-doctor-setup/            # Environment health checking & automated setup
│   ├── 012-swarm-delegation/        # Swarm delegation workflow
│   ├── 013-binary-rename/           # CLI binary rename (unbound → unbound-force/uf)
│   ├── 014-dewey-architecture/      # Dewey semantic knowledge layer design
│   ├── 015-dewey-integration/       # Dewey integration with agents, scaffold, doctor, setup
│   ├── 016-autonomous-define/       # Autonomous define with Dewey
│   └── 018-unleash-command/         # Autonomous Speckit pipeline (/unleash)
├── schemas/                         # JSON Schema registry for inter-hero artifacts
│   ├── envelope/                    # Artifact envelope schema + samples
│   ├── quality-report/              # Gaze quality report payload schema
│   ├── review-verdict/              # Divisor review verdict payload schema
│   ├── backlog-item/                # Muti-Mind backlog item payload schema
│   ├── acceptance-decision/         # Muti-Mind acceptance decision payload schema
│   ├── metrics-snapshot/            # Mx F metrics snapshot payload schema
│   ├── coaching-record/             # Mx F coaching record payload schema
│   ├── workflow-record/             # Orchestration workflow record payload schema
│   ├── convention-pack/             # Convention pack structural schema
│   ├── hero-manifest/               # Hero manifest JSON Schema
│   └── samples/                     # Legacy sample artifacts
├── scripts/
│   └── validate-hero-contract.sh    # Contract compliance validation
├── go.mod                           # Go module definition
├── opencode.json                    # MCP server configuration (Dewey)
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

- **001-org-constitution**: Four core principles, governance, hero alignment
- **002-hero-interface-contract**: Standard repo structure, artifact envelope, naming conventions, hero manifest
- **003-specification-framework**: Unified specification framework (Speckit strategic + OpenSpec tactical), define extension points

### Phase 1: Hero Architectures (004-007)

Each hero's design. Can proceed in parallel once Phase 0 is done.

- **004-muti-mind-architecture**: AI persona, backlog CLI, priority scoring, GitHub sync, acceptance authority
- **005-the-divisor-architecture**: Five-persona review council, convention packs, deployment generator
- **006-cobalt-crush-architecture**: Dev persona, coding standards, Gaze/Divisor feedback loops
- **007-mx-f-architecture**: Metrics platform, coaching engine, impediment tracking, sprint management

### Phase 2: Cross-Cutting (008-009, 016, 018)

Depends on all Phase 1 specs.

- **008-swarm-orchestration**: Feature lifecycle workflow, artifact handoff, Swarm plugin, learning loop
- **009-shared-data-model**: JSON schemas for all artifact types, versioning, schema registry
- **016-autonomous-define**: Autonomous define with Dewey -- configurable execution modes, seed command, spec review checkpoint
- **018-unleash-command**: Autonomous Speckit pipeline (`/unleash`) -- single command from spec to demo-ready code

### Phase 3: Infrastructure (010-013)

Meta-repo operational specs (not hero architectures).

- **010-knowledge-graph-integration**: MCP knowledge graph via graphthulhu, Obsidian backend, spec search/traversal
- **011-doctor-setup**: Environment health checking (`uf doctor`) and automated tool chain setup (`uf setup`)
- **012-swarm-delegation**: Swarm delegation workflow with execution mode awareness
- **013-binary-rename**: CLI binary rename (`unbound` → `unbound-force` with `uf` alias)

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
  016-autonomous-define (depends on 012, 014, 015)
  018-unleash-command (depends on 003, 008, 012, 014, 016)

Phase 3 (Infrastructure) — meta-repo operational
  010-knowledge-graph-integration
  011-doctor-setup (depends on 001, 003, 008)
  012-swarm-delegation
  013-binary-rename (depends on 003, 011)
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
| `unbound-force/dewey` | Semantic knowledge layer (MCP server) | N/A | Active |
| `unbound-force/homebrew-tap` | Homebrew formula distribution | N/A | Active |

## Specification Framework

This repo uses a unified two-tier specification framework
distributed via the `unbound-force` CLI binary (alias: `uf`). Install with
`brew install unbound-force/tap/unbound-force` and run
`uf init` to scaffold into any repository.

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

A mandatory gate at the planning phase. The constitution's four core principles -- Autonomous Collaboration, Composability First, Observable Quality, and Testability -- must each receive a PASS before proceeding. Constitution violations are automatically CRITICAL severity and non-negotiable.

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

### Branch Conventions

Both tiers enforce branch-based workflows:

- **Speckit** branches: `NNN-<short-name>`
  (e.g., `013-binary-rename`). Created automatically by
  `/speckit.specify`. Validated by `check-prerequisites.sh`
  at every pipeline step (hard gate).
- **OpenSpec** branches: `opsx/<change-name>`
  (e.g., `opsx/doctor-ux-improvement`). Created by
  `/opsx-propose`. Validated by `/opsx-apply` before
  implementation (hard gate).

The `opsx/` prefix namespace ensures OpenSpec branches
are visually distinct from Speckit branches in
`git branch` output and do not collide with the
`NNN-*` numbering pattern.

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
cmd/unbound-force/     CLI layer (Cobra commands, flag handling)
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

## Knowledge Retrieval

Agents SHOULD prefer Dewey MCP tools over grep/glob/read
for cross-repo context, design decisions, and
architectural patterns. Dewey provides semantic search
across all indexed Markdown files, specs, and web
documentation — returning ranked results with provenance
metadata that grep cannot match.

### Tool Selection Matrix

| Query Intent | Dewey Tool | When to Use |
|-------------|-----------|-------------|
| Conceptual understanding | `dewey_semantic_search` | "How does X work?", "Patterns for Y" |
| Keyword lookup | `dewey_search` | Known terms, file names, FR numbers |
| Read specific page | `dewey_get_page` | Known spec or document path |
| Relationship discovery | `dewey_find_connections` | "How are X and Y related?" |
| Similar documents | `dewey_similar` | "Find specs like this one" |
| Tag-based discovery | `dewey_find_by_tag` | "All pages tagged #decision" |
| Property queries | `dewey_query_properties` | "All specs with status: draft" |
| Filtered semantic | `dewey_semantic_search_filtered` | Semantic search within source type |
| Graph navigation | `dewey_traverse` | Dependency chain walking |

### When to Fall Back to grep/glob/read

Use direct file operations instead of Dewey when:
- **Dewey is unavailable** — MCP tools return errors or
  are not configured
- **Exact string matching is needed** — searching for a
  specific error message, variable name, or code pattern
- **Specific file path is known** — reading a file you
  already know the path to (use Read directly)
- **Binary/non-Markdown content** — Dewey indexes
  Markdown; use grep for Go source, JSON, YAML, etc.

### Graceful Degradation (3-Tier Pattern)

All Knowledge Retrieval steps follow this pattern:

**Tier 3 (Full Dewey)** — semantic + structured search:
- `dewey_semantic_search` — natural language queries
- `dewey_search` — keyword queries
- `dewey_get_page`, `dewey_find_connections`,
  `dewey_traverse` — structured navigation
- `dewey_find_by_tag`, `dewey_query_properties` —
  metadata queries

**Tier 2 (Graph-only, no embedding model)** — structured
search only:
- `dewey_search` — keyword queries (no embeddings needed)
- `dewey_get_page`, `dewey_traverse`,
  `dewey_find_connections` — graph navigation
- `dewey_find_by_tag`, `dewey_query_properties` —
  metadata queries
- Semantic search unavailable — use exact keyword matches

**Tier 1 (No Dewey)** — direct file access:
- Use Read tool for direct file access
- Use Grep for keyword search across the codebase
- Use Glob for file pattern matching

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

### Embedding Model Alignment

Both Dewey and Swarm use IBM Granite Embedding
(`granite-embedding:30m`, Apache 2.0) for semantic
search. To ensure all tools use the same model,
add these to your shell profile:

```bash
export OLLAMA_MODEL=granite-embedding:30m
export OLLAMA_EMBED_DIM=256
```

`uf setup` sets these automatically for child processes.
Setting them in your shell profile ensures consistency
when running Swarm commands directly.

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
- Go 1.24+ (for tooling/MCP if any, though primarily OpenCode agents/commands) + OpenCode runtime, GitHub CLI (`gh`) or GitHub API (004-muti-mind-architecture)
- Local Markdown files (YAML frontmatter) in `.muti-mind/backlog/` indexed by Dewey (004-muti-mind-architecture)
- Go 1.24+ (CLI backend), OpenCode Agents (AI runtime) + `github.com/spf13/cobra`, `github.com/charmbracelet/log`, OpenCode Runtime, Dewey MCP Server (004-muti-mind-architecture)
- Local Markdown files with YAML frontmatter in `.muti-mind/backlog/` (004-muti-mind-architecture)
- Go 1.24+ (CLI/scaffold engine), Markdown (agents, packs, commands) + `github.com/spf13/cobra` (CLI), `embed.FS` (asset embedding), `github.com/charmbracelet/log` (logging) (005-the-divisor-architecture)
- Filesystem only (embedded assets deployed to target directory) (005-the-divisor-architecture)
- Markdown (agent file), Go 1.24+ (scaffold engine refactor) + `embed.FS` (asset embedding), existing scaffold engine (006-cobalt-crush-architecture)
- Filesystem only (Markdown files deployed to target directory) (006-cobalt-crush-architecture)
- Go 1.24+ (CLI backend), Markdown (coaching agent) + `github.com/spf13/cobra` (CLI), `github.com/charmbracelet/log` (logging), `github.com/charmbracelet/lipgloss` (terminal styling), `embed.FS` (agent embedding) (007-mx-f-architecture)
- JSON files in `.mx-f/data/{source}/{timestamp}.json` for metrics, Markdown+YAML frontmatter in `.mx-f/impediments/` for impediments, `.mx-f/retros/` for retrospective records (007-mx-f-architecture)
- Go 1.24+ (orchestration engine), Markdown (commands, skills) + `internal/artifacts` (envelope, FindArtifacts, WriteArtifact, ReadEnvelope — already exist from Spec 007), `internal/sync` (GHRunner), `github.com/charmbracelet/log` (008-swarm-orchestration)
- JSON files at `.unbound-force/workflows/{workflow_id}.json` (workflow state), `.unbound-force/artifacts/{type}/{timestamp}-{hero}.json` (artifacts) (008-swarm-orchestration)
- Go 1.24+ + `github.com/spf13/cobra` (CLI), `github.com/charmbracelet/lipgloss` (terminal styling), `gopkg.in/yaml.v3` (frontmatter parsing) (011-doctor-setup)
- N/A (reads filesystem and subprocess output, writes only to `opencode.json`) (011-doctor-setup)
- Go 1.24 + `github.com/spf13/cobra` (012-swarm-delegation)
- Go 1.24+ + `github.com/modelcontextprotocol/go-sdk/mcp` (014-dewey-architecture)
- SQLite for persistent indexes (knowledge (014-dewey-architecture)
- N/A (configuration and documentation changes; Dewey MCP tools: `dewey_search`, `dewey_semantic_search`, `dewey_traverse`, `dewey_get_page`, `dewey_find_by_tag`, `dewey_query_properties`, `dewey_find_connections`, `dewey_similar`, `dewey_semantic_search_filtered`) (015-dewey-integration)
- JSON workflow files at `.unbound-force/workflows/` (016-autonomous-define)
- `opencode.json` at repo root (JSON file) (017-init-opencode-config)
- Markdown (OpenCode command file) + Existing slash commands (018-unleash-command)
- N/A (orchestrates existing tools) (018-unleash-command)
- Go 1.24+ (scaffold engine, setup), Markdown (agents, packs, commands), YAML (CI workflow) + `github.com/spf13/cobra` (CLI), `embed.FS` (asset embedding), `github.com/charmbracelet/log` (logging) (019-divisor-council-refinement)
- Filesystem only (Markdown files deployed to target directory, CI workflow YAML) (019-divisor-council-refinement)
- N/A (configuration and documentation changes; Dewey MCP tools: `dewey_search`, `dewey_semantic_search`, `dewey_traverse`, `dewey_get_page`, `dewey_find_by_tag`, `dewey_query_properties`, `dewey_find_connections`, `dewey_similar`, `dewey_semantic_search_filtered`) (020-dewey-knowledge-retrieval)
- Go 1.24+ (Dewey repo: Ollama + Dewey MCP server, Ollama HTTP (021-dewey-unified-memory)
- Dewey `graph.db` (SQLite — learnings (021-dewey-unified-memory)
- Go 1.24+ (scaffold engine, regression tests), Markdown (agent files, command files, documentation) + `embed.FS` (scaffold engine), `testing` (regression tests), Dewey MCP tools (`dewey_store_learning`, `dewey_semantic_search`) (022-hivemind-dewey-migration)
- N/A (Markdown file edits; Dewey persists to `graph.db` — out of scope for this repo) (022-hivemind-dewey-migration)

## Recent Changes

- 022-hivemind-dewey-migration: Migrated all Hivemind tool references to Dewey equivalents across agent files and commands (Spec 021 Phase 2, FR-012 through FR-015). Replaced `hivemind_store` with `dewey_store_learning` in `/unleash` retrospective step. Replaced `hivemind_find` with `dewey_semantic_search` in all 5 Divisor agents' Prior Learnings step (`divisor-adversary.md`, `divisor-architect.md`, `divisor-guard.md`, `divisor-sre.md`, `divisor-testing.md`). Updated Cobalt-Crush agent prose from "complementing Hivemind" to "Dewey is the unified memory layer." Updated AGENTS.md Embedding Model Alignment section and Spec 020 Recent Changes entry to reflect Dewey as unified (not complementary) memory system. Updated `internal/setup/setup.go` comments to remove Hivemind references. Synchronized 7 scaffold asset copies (`internal/scaffold/assets/`). Added `TestScaffoldOutput_NoHivemindReferences` regression test (5 stale patterns: `hivemind_store`, `hivemind_find`, `hivemind_validate`, `hivemind_remove`, `hivemind_get`). All learning storage/retrieval operations gracefully degrade when Dewey is unavailable. All 5 user stories and 29 tasks completed.
- 020-dewey-knowledge-retrieval: Added Dewey knowledge retrieval behavioral instructions across the agent ecosystem. Added "Knowledge Retrieval" top-level section to AGENTS.md with tool selection matrix (9 Dewey tools mapped to query intents), fallback criteria (when to use grep/glob/read instead), and 3-tier graceful degradation pattern (Full Dewey, Graph-only, No Dewey). Enhanced Cobalt-Crush with "Step 0: Knowledge Retrieval" that fires before code exploration (prior learnings, related specs, architectural patterns). Added Dewey query steps to 3 Speckit commands: `/speckit.specify` (search for similar specs), `/speckit.plan` (search for prior research decisions), `/speckit.tasks` (search for implementation patterns). Enhanced 3 hero agents with role-specific "Step 0" and "prefer Dewey" instructions: Muti-Mind (backlog patterns, acceptance history), Mx F (velocity trends, retrospective outcomes), Gaze (CRAP score patterns, quality baselines). Updated `unbound-force-heroes` SKILL.md with per-stage Dewey query table for Swarm coordinators. All Dewey instructions are SHOULD (soft preference) with graceful degradation — Dewey is the unified memory layer (superseded by Spec 022). All 5 user stories and 22 tasks completed.
- 019-divisor-council-refinement: Refined the Divisor Council review system -- removed 4 legacy `reviewer-*.md` scaffold assets (net -3 files: 52 → 49), added legacy file detection warning to `uf init`, de-duplicated cross-persona review responsibilities with exclusive ownership boundaries and "Out of Scope" sections in all 5 `divisor-*.md` agents, created shared `severity.md` convention pack (tool-owned, language-agnostic, always-deploy), qualified all FR references with "per Spec NNN" format, added Prior Learnings step (Hivemind `hivemind_find` with graceful degradation) to all 5 agents, updated `/review-council` reference table and severity pack reference, added `golangci-lint` and `govulncheck` to CI workflow and `uf setup` (step count 13 → 15), fixed 28+ pre-existing lint findings, added `.golangci.yml` with `version: "2"` and `fmt.Fprint*` exclusions. All 6 user stories and 56 tasks completed.
- 018-unleash-command: Created `/unleash` autonomous Speckit pipeline command -- single Markdown command file (`.opencode/command/unleash.md`, ~600 lines) that orchestrates 8 steps (clarify, plan, tasks, spec review, implement, code review, retrospective, demo) with Dewey-powered clarification (auto-resolve with provenance annotations), parallel Swarm worker execution (max 4 concurrent, worktree isolation, cherry-pick merge), filesystem-based resumability (no state file -- probes spec.md markers, plan.md/tasks.md existence, `<!-- spec-review: passed -->` marker, task checkbox state), 6 exit points (clarify unanswerable, spec review HIGH/CRITICAL, worker failure, merge conflict, build checkpoint, code review 3 iterations), graceful degradation for all optional tools (Dewey, Gaze, Swarm worktrees, Hivemind, SwarmMail), CI commands derived from `.github/workflows/` (not hardcoded). Scaffold asset copy at `internal/scaffold/assets/opencode/command/unleash.md`, file count updated 51 → 52, expectedAssetPaths updated (13 → 14 commands). Pre-existing scaffold drift synced (23 assets). All 6 user stories and 31 tasks completed.
- 017-init-opencode-config: Moved `opencode.json` management from `uf setup` to `uf init`. `uf init` now creates/updates `opencode.json` with Dewey MCP server entry (`mcp.dewey` with `type: local`, `command: ["dewey", "serve", "--vault", "."]`, `enabled: true`) when `dewey` is in PATH, and Swarm plugin entry (`opencode-swarm-plugin` in `plugin` array) when `.hive/` exists. Idempotent by default (skips when both entries present, preserves custom MCP servers and plugins). `--force` overwrites stale `mcp.dewey` entries. `scaffold.Options` expanded with `ReadFile`, `WriteFile`, and `DryRun` fields for injectable file I/O. `printSummary()` updated with new action symbols (`✓`/`—`/`✗`). `uf setup` step count reduced from 16 to 15 (opencode.json step removed, now handled transparently by `uf init` at final step). `uf doctor` `checkMCPConfig()` fixed to check canonical `"mcp"` key first with `"mcpServers"` fallback, and to extract binary names from both string-style and array-style command fields. Legacy `mcpServers.dewey` treated as already configured. All 4 user stories and 37 tasks completed.
- 016-autonomous-define: Enabled autonomous define stage -- configurable execution mode overrides on `NewWorkflow()` (accepts `overrides map[string]string` for per-stage mode customization, validated against `StageOrder()` and `ModeHuman`/`ModeSwarm`), `SpecReviewEnabled` field on `WorkflowInstance` (instance-only, not on `WorkflowRecord`), spec review checkpoint in `Advance()` (fires when define=swarm + specReview=true, silently skipped when define=human), `Start()` updated to forward overrides and specReview, `/workflow seed` command (creates backlog item + starts workflow with define=swarm in one operation), `--define-mode` and `--spec-review` flags on `/workflow start`, Muti-Mind autonomous specification workflow instructions (6-step Dewey-powered spec drafting with Tier 1 fallback), SKILL.md updated with seed workflow section and comparison table, workflow command docs updated for spec review checkpoint output. All 5 user stories and 28 tasks completed.
- setup-init-full-stack: Extended `uf setup` to install ALL ecosystem tools and `uf init` to initialize sub-tools automatically. `uf setup` now installs 3 new tools (Mx F via `brew install unbound-force/tap/mxf`, GitHub CLI via `brew install gh`, OpenSpec CLI via `bun add -g @fission-ai/openspec@latest` with npm fallback) and runs 2 new initialization steps (`dewey init` to create `.dewey/` workspace, `dewey index` to build initial search index). Step count increased from 11 to 16. `uf init` now initializes Dewey workspace after scaffolding via `initSubTools()` (idempotent — skips if `.dewey/` exists or if called after `uf setup`). `scaffold.Options` expanded with `LookPath` and `ExecCmd` fields for testability. `printSummary` updated with context-aware next-step guidance (constitution, doctor, speckit, opsx when tools available; `uf setup` when tools missing). `runUnboundInit` forwards `LookPath`/`ExecCmd` to scaffold for injection chain. All changes follow existing patterns (installGaze for Homebrew tools, installSwarmPlugin for npm/bun tools, runSwarmSetup for subprocess init). 43 tasks completed.
- 015-dewey-integration: Integrated Dewey as the semantic knowledge layer replacing graphthulhu. Updated `opencode.json` MCP config (live + scaffold) from `knowledge-graph`/`graphthulhu` to `dewey`. Replaced all `knowledge-graph_*` tool references with `dewey_*` across agent files and commands. Added "Knowledge Retrieval" sections with role-specific Dewey usage and 3-tier graceful degradation pattern (Full Dewey, Graph-only, No Dewey) to all 5 hero agent files and 5 Divisor persona files. Added "Dewey Knowledge Layer" health check group to `uf doctor` (dewey binary, embedding model, workspace). Added Dewey installation step to `uf setup` (Homebrew install, embedding model pull). Updated embedding model from `mxbai-embed-large` to `granite-embedding:30m`. Added `TestScaffoldOutput_NoGraphthulhuReferences` regression test. All 5 user stories and 40 tasks completed.
- 013-binary-rename: Renamed CLI binary from `unbound` to `unbound-force` with `uf` symlink alias to resolve NLnet Labs Unbound DNS resolver name collision. Directory rename `cmd/unbound/` → `cmd/unbound-force/`, Cobra root command `Use` field updated, Makefile `install` target creates both binaries, GoReleaser config updated (build id, binary name, archive template, cask name, quarantine hook, `uf` symlink via post-install hook), release workflow updated (20+ references), scaffold engine `versionMarker()` and `printSummary()` updated to `uf`, all embedded asset markers changed from `scaffolded by unbound` to `scaffolded by uf`, doctor hint strings updated (`uf init`/`uf setup`), setup progress messages updated, scaffold assets updated (reviewer-adversary, reviewer-guard, reviewer-sre, reviewer-architect, cobalt-crush-dev, specify/config.yaml), living documentation updated (AGENTS.md, README.md, unbound-force.md), 3 new regression tests (TestRootCmd_HelpOutput, TestScaffoldOutput_NoBareUnboundReferences, TestDoctorHints_NoBareUnboundReferences). Completed specs (001-012) and archived OpenSpec changes preserved as historical records. All 5 user stories and 35 tasks completed.
- 012-swarm-delegation: Added swarm delegation workflow -- execution mode awareness (`ModeHuman`/`ModeSwarm`) on each `WorkflowStage`, automatic checkpoint pausing at swarm→human boundaries (`StatusAwaitingHuman`), resume via `Advance()`, `StageExecutionModeMap()` with default assignments (define=human, implement=swarm, validate=swarm, review=swarm, accept=human, reflect=swarm), renamed `StageMeasure` to `StageReflect` with enriched reflect stage documentation (metrics + learning + retrospective), `WorkflowStore.Latest()` discovers both active and awaiting_human workflows, backward compatible with legacy JSON (empty execution_mode treated as human), workflow-record schema v1.1.0 with `execution_mode` field, updated `/workflow` command docs with `[human]`/`[swarm]` indicators and `⏸` awaiting_human display, SKILL.md updated with execution modes section and reflect stage documentation. All 4 user stories and 42 tasks completed.
- 011-doctor-setup: Implemented Doctor and Setup commands -- two new packages `internal/doctor/` (5 files: models.go, doctor.go, environ.go, checks.go, format.go) and `internal/setup/` (1 file: setup.go). `unbound doctor` checks 7 groups (Detected Environment, Core Tools, Swarm Plugin, Scaffolded Files, Hero Availability, MCP Server Config, Agent/Skill Integrity) with environment-aware install hints, colored text output (lipgloss), and JSON output (`--format=json`). `unbound setup` installs missing tools through detected version managers (goenv, nvm, fnm, mise, bun, Homebrew), configures opencode.json atomically, initializes .hive/, and runs unbound init. Supports `--dry-run` and `--yes` flags. Platform guard rejects Windows. All external dependencies injected for testability. Reuses `orchestration.DetectHeroes()` for hero availability. Promoted lipgloss from indirect to direct dependency. All 5 user stories and 79 tasks completed.
- 009-shared-data-model: Implemented shared data model -- Go package `internal/schemas/` with JSON Schema generation from Go structs (`invopop/jsonschema`), runtime validation (`santhosh-tekuri/jsonschema/v6`), convention pack structural validation, schema versioning with semver compatibility checking. Schema registry at `schemas/` with 9 artifact type directories (envelope, quality-report, review-verdict, backlog-item, acceptance-decision, metrics-snapshot, coaching-record, workflow-record, convention-pack), each containing v1.0.0.schema.json, samples/, and README.md. Type aliases reuse existing Go structs (MetricsSnapshot, WorkflowRecord, AcceptanceDecision); new structs defined for QualityReportPayload and ReviewVerdictPayload. CI tests validate all schemas are draft 2020-12, all samples pass validation, and directory structure is complete. All 5 user stories and 20 tasks completed.
- 008-swarm-orchestration: Implemented swarm orchestration engine -- Go package `internal/orchestration/` with Orchestrator struct (Start, Advance, Skip, Escalate, Complete, Status, List), 6-stage hero lifecycle workflow (define, implement, validate, review, accept, measure), hero availability detection via agent files and exec.LookPath, workflow state persistence as JSON at `.unbound-force/workflows/`, workflow-record artifact production via `internal/artifacts`, learning feedback extraction (AnalyzeWorkflows, SaveFeedback, LoadFeedback), failure mode handling (max iterations escalation, acceptance rejection, inter-hero contradiction), ArtifactContext struct for workflow metadata in envelopes, FindArtifactsByHero and FindArtifactsSince for enhanced artifact discovery, CheckSchemaVersion for compatibility checking, 4 `/workflow` commands (start, status, list, advance), Swarm skills package at `.opencode/skill/unbound-force-heroes/SKILL.md`. All 5 user stories and 31 tasks completed.
- 007-mx-f-architecture: Implemented Mx F Manager hero -- Go CLI backend (`cmd/mxf/`) with 7 subcommands (collect, metrics, impediment, dashboard, sprint, standup, retro), 5 domain packages (`internal/metrics/`, `internal/impediment/`, `internal/coaching/`, `internal/dashboard/`, `internal/sprint/`), OpenCode coaching agent (`mx-f-coach.md`), artifact package generalization (WriteArtifact, ReadEnvelope, FindArtifacts), GoReleaser multi-binary configuration, 47 total embedded files (was 46). All 6 user stories and 68 tasks completed.
- 006-cobalt-crush-architecture: Implemented Cobalt-Crush Developer persona -- single `cobalt-crush-dev.md` agent file with engineering philosophy (clean code, SOLID, TDD, spec-driven development), convention pack adherence via shared `.opencode/unbound/packs/`, Gaze feedback loop (quality report consumption), Divisor review preparation (finding resolution, learned patterns), speckit integration (task processing, phase checkpoints), graphthulhu MCP integration (optional). Prerequisite refactor relocated convention packs from `.opencode/divisor/packs/` to `.opencode/unbound/packs/` (shared neutral location), updating scaffold engine, tests, Divisor agents, and documentation. 46 total embedded files (was 45).
- 005-the-divisor-architecture: Implemented The Divisor PR Reviewer Council -- 5 canonical `divisor-*.md` persona agents (Guard, Architect, Adversary, SRE, Testing) with dynamic discovery, 6 convention pack files (Go, TypeScript, default + custom stubs) loaded at review time, `/review-council` command updated to `divisor-*` pattern, `unbound init --divisor` subset deployment with `--lang` language override, scaffold engine extended with `isDivisorAsset()`, `detectLang()`, `shouldDeployPack()`, convention pack ownership model (tool-owned canonical, user-owned custom), 45 total embedded files (was 33), all 5 user stories and 44 tasks completed.
- graphthulhu-homebrew-distribution: `brew install unbound-force/tap/unbound` now installs graphthulhu automatically as a cask dependency. Added `Casks/graphthulhu.rb` to `unbound-force/homebrew-tap` (v0.4.0, all four platform checksums), added `dependencies: [{cask: graphthulhu}]` and macOS quarantine-removal hook to `.goreleaser.yaml`, opened upstream PR to `skridlevsky/graphthulhu` (#5) for eventual tap ownership transfer.
- 004-muti-mind-architecture: Implemented Muti-Mind backend CLI (`cmd/mutimind`), OpenCode agent (`muti-mind-po`), commands, local MD backlog parsing, GitHub bidirectional sync, and JSON artifact generation (`backlog-item`, `acceptance-decision`).
- 003-specification-framework: Implemented unified two-tier specification framework -- `unbound` CLI binary (Go + Cobra + embed.FS), scaffolds files via `unbound init` (33 initial, extended to 45 by Spec 005), custom `unbound-force` OpenSpec schema with constitution alignment in proposals, boundary guidelines documented, GoReleaser v2 release pipeline, all 8 user stories (US1-US8) and 67 tasks completed, format-aware version markers, drift detection tests
- 010-knowledge-graph-integration: Implemented knowledge graph integration -- graphthulhu installed with `--include-hidden` support (upstream PR submitted), OpenCode MCP configuration created, all 5 user stories verified (search, analysis, live sync, link traversal, property queries), YAML frontmatter added to specs 001/002/004, 50 pages indexed across all directories including `.specify/` and `.opencode/`
- 002-hero-interface-contract: Completed spec implementation -- Hero Interface Contract v1.0.0 ratified, hero manifest JSON Schema created and validated, contract compliance validation script created and tested against Gaze and Website repos, sample artifact envelope and Gaze manifest produced, all FRs (001-015) and SCs (001-007) validated, spec status set to Complete
- 001-org-constitution: Completed spec implementation -- constitution ratified v1.0.0, alignment agent and `/constitution-check` command created, all FRs and SCs validated, spec status set to Complete
- 001-org-constitution: Ratified org constitution v1.0.0 with three principles (Autonomous Collaboration, Composability First, Observable Quality)
- Specs 001-009: Added architectural design specs for all heroes (Muti-Mind, The Divisor, Cobalt-Crush, Mx F), infrastructure (org constitution, hero interface contract, speckit framework), and cross-cutting concerns (swarm orchestration, shared data model)
