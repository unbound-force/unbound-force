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
│   └── scripts/bash/               # Speckit automation scripts (5 files)
│       ├── common.sh
│       ├── check-prerequisites.sh
│       ├── setup-plan.sh
│       ├── create-new-feature.sh
│       └── update-agent-context.sh
├── .opencode/
│   └── command/                     # Speckit pipeline commands (9 files)
│       ├── speckit.constitution.md
│       ├── speckit.specify.md
│       ├── speckit.clarify.md
│       ├── speckit.plan.md
│       ├── speckit.tasks.md
│       ├── speckit.analyze.md
│       ├── speckit.checklist.md
│       ├── speckit.implement.md
│       └── speckit.taskstoissues.md
├── specs/                           # Architectural specifications
│   ├── 001-org-constitution/        # Org constitution ratification
│   ├── 002-hero-interface-contract/ # Standard hero repo structure & protocols
│   ├── 003-speckit-framework/       # Speckit centralization & distribution
│   ├── 004-muti-mind-architecture/  # Product Owner hero design
│   ├── 005-the-divisor-architecture/# PR Reviewer Council design
│   ├── 006-cobalt-crush-architecture/# Developer hero design
│   ├── 007-mx-f-architecture/       # Manager hero design
│   ├── 008-swarm-orchestration/     # End-to-end workflow & Swarm plugin
│   └── 009-shared-data-model/       # JSON schemas for inter-hero artifacts
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
- **003-speckit-framework**: Centralize speckit, eliminate cross-repo drift, define extension points

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

### Dependency Graph

```text
Phase 0 (Foundation)
  001-org-constitution
    └─> 002-hero-interface-contract
    └─> 003-speckit-framework

Phase 1 (Heroes) — all depend on 001 + 002
  004-muti-mind-architecture
  005-the-divisor-architecture
  006-cobalt-crush-architecture
  007-mx-f-architecture

Phase 2 (Cross-Cutting) — depends on Phase 1
  008-swarm-orchestration
  009-shared-data-model
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

## Speckit Workflow (Mandatory)

All non-trivial feature work **must** go through the Speckit pipeline. The constitution (`.specify/memory/constitution.md`) is the highest-authority document in this project -- all work must align with it.

### Pipeline

The workflow is a strict, sequential pipeline. Each stage has a corresponding `/speckit.*` command:

```text
constitution → specify → clarify → plan → tasks → analyze → checklist → implement
```

| Command | Purpose |
|---------|---------|
| `/speckit.constitution` | Create or update the project constitution |
| `/speckit.specify` | Create a feature specification from a description |
| `/speckit.clarify` | Reduce ambiguity in the spec before planning |
| `/speckit.plan` | Generate the technical implementation plan |
| `/speckit.tasks` | Generate actionable, dependency-ordered task list |
| `/speckit.analyze` | Non-destructive cross-artifact consistency analysis |
| `/speckit.checklist` | Generate requirement quality validation checklists |
| `/speckit.implement` | Execute the implementation plan task by task |
| `/speckit.taskstoissues` | Convert tasks.md into GitHub Issues |

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
- Markdown (constitution document) + OpenCode agent configuration (Markdown agent files) + OpenCode (agent runtime), speckit (pipeline integration) (001-org-constitution)
- Filesystem only -- constitution at `.specify/memory/constitution.md`, agent files in `.opencode/agents/` and `.opencode/command/` (001-org-constitution)

- Markdown (specifications, governance)
- Bash (speckit scripts)
- JSON Schema draft 2020-12 (planned, for shared data model)
- OpenCode + Speckit (development workflow)

## Recent Changes

- 001-org-constitution: Ratified org constitution v1.0.0 with three principles (Autonomous Collaboration, Composability First, Observable Quality)
- Specs 001-009: Added architectural design specs for all heroes (Muti-Mind, The Divisor, Cobalt-Crush, Mx F), infrastructure (org constitution, hero interface contract, speckit framework), and cross-cutting concerns (swarm orchestration, shared data model)
