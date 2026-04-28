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

### Gatekeeping Value Protection

Agents MUST NOT modify values that serve as quality or governance gates to make an implementation pass. The following categories are protected:

1. **Coverage thresholds and CRAP scores** -- minimum coverage percentages, CRAP score limits, coverage ratchets
2. **Severity definitions and auto-fix policies** -- CRITICAL/HIGH/MEDIUM/LOW boundaries, auto-fix eligibility rules
3. **Convention pack rule classifications** -- MUST/SHOULD/MAY designations on convention pack rules (downgrading MUST to SHOULD is prohibited)
4. **CI flags and linter configuration** -- `-race`, `-count=1`, `OSV-Scanner`, `Trivy`, `golangci-lint` rules, pinned action SHAs
5. **Agent temperature and tool-access settings** -- frontmatter `temperature`, `tools.write`, `tools.edit`, `tools.bash` restrictions
6. **Constitution MUST rules** -- any MUST rule in `.specify/memory/constitution.md` or hero constitutions
7. **Review iteration limits and worker concurrency** -- max review iterations, max concurrent Swarm workers, retry limits
8. **Workflow gate markers** -- `<!-- spec-review: passed -->`, task completion checkboxes used as gates, phase checkpoint requirements

**What to do instead**: When an implementation cannot meet a gate, the agent MUST stop, report which gate is blocking and why, and let the human decide whether to adjust the gate or rework the implementation. Modifying a gate without explicit human authorization is a constitution violation (CRITICAL severity).

### Workflow Phase Boundaries

Agents MUST NOT cross workflow phase boundaries:

- **Specify/Clarify/Plan/Tasks/Analyze/Checklist** phases:
  spec artifacts ONLY (`specs/NNN-*/` directory). No source
  code, test, agent, command, or config changes.
- **Implement** phase: source code changes allowed, guided
  by spec artifacts.
- **Review** phase: findings and minor fixes only. No new
  features.

A phase boundary violation is treated as a process error.
The agent MUST stop and report the violation rather than
proceeding with out-of-phase changes.

### CI Parity Gate

Before marking any implementation task complete or
declaring a PR ready, agents MUST replicate the CI checks
locally. Read `.github/workflows/` to identify the exact
commands CI runs, then execute those same commands. Any
failure is a blocking error — a task is not complete
until all CI-equivalent checks pass locally. Do not rely
on a memorized list of commands; always derive them from
the workflow files, which are the source of truth.

### Review Council as PR Prerequisite

Before submitting a pull request, agents **must** run
`/review-council` and resolve all REQUEST CHANGES
findings until all reviewers return APPROVE. There must
be **minimal to no code changes** between the council's
APPROVE verdict and the PR submission — the council
reviews the final code, not a draft that changes
afterward.

Workflow:

1. Complete all implementation tasks
2. Run CI checks locally (build, test, vet)
3. Run `/review-council` — fix any findings, re-run
   until APPROVE
4. Commit, push, and submit PR immediately after council
   APPROVE
5. Do NOT make further code changes between APPROVE and
   PR submission

Exempt from council review:

- Constitution amendments (governance documents, not code)
- Documentation-only changes (README, AGENTS.md, spec
  artifacts)
- Emergency hotfixes (must be retroactively reviewed)

### PR Review

Two review commands serve different workflow moments:

| Command | When | Scope | Agent Model |
|---------|------|-------|-------------|
| `/review-council` | Pre-PR (local tree) | Multi-agent Divisor council with Gaze integration | 5+ parallel agents |
| `/review-pr [N]` | Post-PR (GitHub PR) | Single-agent review with CI causality analysis | 1 agent, token-lean |

`/review-pr` reviews a GitHub PR by number (or
auto-detects from the current branch). It fetches CI
check results, classifies failures as PR-caused vs
pre-existing, runs local deterministic tools, and
applies AI judgment for alignment, security, and
constitution compliance. It can offer fix branches for
pre-existing CI failures and post in-line PR comments
(with human confirmation). Requires `gh` CLI.

Use `/review-council` before pushing to validate your
own work. Use `/review-pr` after creating a PR to
review any PR (yours or others').

### Branch Protection

Agents MUST NOT commit directly to `main`. All changes
-- including documentation, archival, chores, and CI
fixes -- MUST be committed on a feature branch and
submitted via pull request.

Before running `git commit`, verify the current branch:

```bash
git branch --show-current
```

If on `main`, create a branch first:

```bash
git checkout -b <type>/<short-description>
```

A direct commit to `main` is a process violation. If it
occurs, immediately move the commit to a branch (via
`git branch <name>` + `git reset`), restore `main`, and
submit a PR instead.

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

**Agent model resolution**: Agent files do not hardcode a
`model:` field in their frontmatter. Instead, they inherit the
model from OpenCode's own configuration hierarchy: project-level
`opencode.json` `"model"` field > user-level
`~/.config/opencode/opencode.json` > OpenCode's built-in default.
Subagents inherit from their invoking primary agent. To change the
model for all agents, set the `"model"` field in `opencode.json`.

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
│   │   ├── divisor-curator.md       # The Curator: documentation & content pipeline triage (Divisor)
│   │   ├── divisor-guard.md         # The Guard: intent drift detection (Divisor)
│   │   ├── divisor-sre.md           # The Operator: operational readiness (Divisor)
│   │   ├── divisor-testing.md       # The Tester: test quality (Divisor)
│   │   ├── divisor-scribe.md        # The Scribe: technical documentation (Divisor)
│   │   ├── divisor-herald.md        # The Herald: blog/announcements (Divisor)
│   │   ├── divisor-envoy.md         # The Envoy: PR/communications (Divisor)
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
│   │   ├── agent-brief.md            # /agent-brief command (AGENTS.md lifecycle)
│   │   ├── review-council.md        # /review-council command (Divisor)
│   │   ├── review-pr.md            # /review-pr command (GitHub PR review)
│   │   ├── constitution-check.md    # /constitution-check command
│   │   ├── unleash.md               # /unleash autonomous pipeline (Spec 018)
│   │   ├── workflow-start.md        # /workflow start command (Spec 008)
│   │   ├── workflow-status.md       # /workflow status command (Spec 008)
│   │   ├── workflow-list.md         # /workflow list command (Spec 008)
│   │   └── workflow-advance.md      # /workflow advance command (Spec 008)
│   ├── skill/                       # Swarm skills packages
│   │   └── unbound-force-heroes/
│   │       └── SKILL.md             # Hero roles, routing, workflow stages (Spec 008)
│   └── uf/
│       └── packs/                   # Convention packs (shared by all heroes)
│           ├── go.md                # Go convention pack (tool-owned)
│           ├── go-custom.md         # Go custom rules (user-owned)
│           ├── content.md           # Content writing standards (tool-owned)
│           ├── content-custom.md    # Content custom rules (user-owned)
│           ├── default.md           # Language-agnostic default (tool-owned)
│           ├── default-custom.md    # Default custom rules (user-owned)
│           ├── typescript.md        # TypeScript convention pack (tool-owned)
│           └── typescript-custom.md # TypeScript custom rules (user-owned)
├── cmd/unbound-force/
│   ├── main.go                      # Cobra CLI entry point
│   ├── config.go                    # Config command group (init, show, validate)
│   ├── gateway.go                   # Gateway command group (start, stop, status)
│   └── sandbox.go                   # Sandbox command group (start, stop, attach, extract, status)
├── cmd/mutimind/
│   └── main.go                      # Muti-Mind backend CLI entry point
├── internal/scaffold/
│   ├── scaffold.go                  # Core scaffold engine
│   ├── scaffold_test.go             # Tests + drift detection
│   └── assets/                      # Embedded files (go:embed)
├── internal/backlog/                # Muti-Mind local backlog parsing
├── internal/sync/                   # Muti-Mind GitHub issue sync
├── internal/artifacts/              # Artifact envelope I/O (WriteArtifact, ReadEnvelope, FindArtifacts)
├── internal/sandbox/                 # Containerized OpenCode session management (Spec 028, 029)
│   ├── sandbox.go                   # Core orchestration (Create, Start, Stop, Destroy, Attach, Extract, Status)
│   ├── backend.go                   # Backend interface, ResolveBackend(), constants (Spec 029)
│   ├── podman.go                    # PodmanBackend: named volumes, persistent lifecycle (Spec 029)
│   ├── che.go                       # CheBackend: Eclipse Che/Dev Spaces provisioning (Spec 029)
│   ├── workspace.go                 # WorkspaceStatus, SandboxConfig, LoadConfig, git sync (Spec 029)
│   ├── detect.go                    # Platform detection (DetectPlatform, SELinux)
│   ├── config.go                    # Container config (buildRunArgs, env vars, volumes)
│   └── sandbox_test.go              # Tests for all sandbox functions (82 tests)
├── internal/gateway/                 # LLM reverse proxy gateway (Spec 033)
│   ├── gateway.go                   # Core gateway: Options, Start, Stop, Status, health endpoint, reverse proxy
│   ├── provider.go                  # Provider interface, DetectProvider, Anthropic/Vertex/Bedrock implementations
│   ├── refresh.go                   # Token refresh loop, SigV4 signing, credential refresh helpers
│   ├── pid.go                       # PID file management (WritePID, ReadPID, IsAlive, CleanupStale)
│   ├── signal_unix.go               # Unix signal constants (signal 0 liveness check)
│   └── gateway_test.go              # Tests for all gateway functions
├── internal/config/                  # Unified configuration loading and validation
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
| `unbound-force/homebrew-tap` | Homebrew cask + formula distribution | N/A | Active |

## Spec-First Development

All changes that modify production code, test code, agent
prompts, embedded assets, or CI configuration **must** be
preceded by a spec workflow. The constitution
(`.specify/memory/constitution.md`) is the highest-
authority document in this project — all work must align
with it.

Two spec workflows are available:

| Workflow | Location | Best For |
|----------|----------|----------|
| **Speckit** | `specs/NNN-name/` | Numbered feature specs with the full pipeline |
| **OpenSpec** | `openspec/changes/name/` | Targeted changes with lightweight artifacts |

**What requires a spec** (no exceptions without explicit
user override):

- New features or capabilities
- Refactoring that changes function signatures, extracts
  helpers, or moves code between packages
- Test additions or assertion strengthening across
  multiple functions
- Agent prompt changes
- CI workflow modifications
- Data model changes (new struct fields, schema updates)

**What is exempt** (may be done directly):

- Constitution amendments (governed by the constitution's
  own Governance section)
- Typo corrections, comment-only changes, single-line
  formatting fixes
- Emergency hotfixes for critical production bugs (must
  be retroactively documented)

When an agent is unsure whether a change is trivial, it
**must** ask the user rather than proceeding without a
spec. The cost of an unnecessary spec is minutes; the
cost of an unplanned change is rework, drift, and broken
CI.

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

### Website Documentation Gate

When a change affects user-facing behavior, hero capabilities, CLI commands, or workflows, a GitHub issue **MUST** be created in the `unbound-force/website` repository to track required documentation or website updates. The issue must be created before the implementing PR is merged.

```bash
gh issue create --repo unbound-force/website \
  --title "docs: <brief description of what changed>" \
  --body "<what changed, why it matters, which pages need updating>"
```

**Exempt changes** (no website issue needed):
- Internal refactoring with no user-facing behavior change
- Test-only changes
- CI/CD pipeline changes
- Spec artifacts (specs are internal planning documents)

**Examples requiring a website issue**:
- New CLI command or flag added
- Hero capabilities changed (new agent, removed feature)
- Installation steps changed (`uf setup` flow)
- New convention pack added
- Breaking changes to any user-facing workflow

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

### CI Workflow Structure

| Workflow | File | Triggers | Purpose |
|---|---|---|---|
| Local CI | `ci_local.yml` | push/PR to main | Build, test (`-race -count=1`), coverage ratchets |
| CI Checks | `ci_checks.yml` | push/PR to main | MegaLinter (Go, Actions, Bash, Gitleaks) + commitlint |
| Security | `ci_security.yml` | push/PR to main | OSV-Scanner, Trivy source scan, OpenSSF Scorecards |
| Dependencies | `ci_dependencies.yml` | push/PR to main | Dependency review + dependabot auto-approval |
| CRAP Load | `ci_crapload.yml` | PR to main | CRAP/GazeCRAP regression analysis + PR comment |
| Scheduled | `ci_scheduled.yml` | daily midnight UTC | OSV-Scanner + OpenSSF Scorecards |
| Release | `release.yml` | tag push (`v*`) | GoReleaser + macOS signing + Homebrew tap |

All workflows consuming org-infra reusable workflows are
SHA-pinned to `v0.1.0`. All workflows include SPDX headers,
explicit permissions blocks, and concurrency groups.

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

The embedding model defaults can be overridden in
`.uf/config.yaml`:

```yaml
embedding:
  model: granite-embedding:30m
  dimensions: 256
```

Run `uf config init` to create the config file.

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

Commit format, branching strategy, code review requirements,
and semantic versioning rules are defined in the org
constitution (`.specify/memory/constitution.md`, Development
Workflow section). This repo follows all org-level rules
without exception.

## Active Technologies
- Go 1.24+ (unbound CLI binary, Cobra CLI framework, embed.FS scaffold)
- GoReleaser v2 (cross-platform release pipeline, Homebrew cask + formula publishing, RPM generation via nfpms)
- Markdown (specifications, governance, templates, commands)
- YAML (OpenSpec schema, configuration files)
- Bash (speckit scripts)
- JSON Schema draft 2020-12 (hero manifest schema, sample artifact envelope)
- OpenCode + Speckit + OpenSpec (development workflow)
- Node.js >= 20.19.0 (OpenSpec CLI, `@fission-ai/openspec`)
- Go 1.24+ (for tooling/MCP if any, though primarily OpenCode agents/commands) + OpenCode runtime, GitHub CLI (`gh`) or GitHub API (004-muti-mind-architecture)
- Local Markdown files (YAML frontmatter) in `.uf/muti-mind/backlog/` indexed by Dewey (004-muti-mind-architecture)
- Go 1.24+ (CLI backend), OpenCode Agents (AI runtime) + `github.com/spf13/cobra`, `github.com/charmbracelet/log`, OpenCode Runtime, Dewey MCP Server (004-muti-mind-architecture)
- Local Markdown files with YAML frontmatter in `.uf/muti-mind/backlog/` (004-muti-mind-architecture)
- Go 1.24+ (CLI/scaffold engine), Markdown (agents, packs, commands) + `github.com/spf13/cobra` (CLI), `embed.FS` (asset embedding), `github.com/charmbracelet/log` (logging) (005-the-divisor-architecture)
- Filesystem only (embedded assets deployed to target directory) (005-the-divisor-architecture)
- Markdown (agent file), Go 1.24+ (scaffold engine refactor) + `embed.FS` (asset embedding), existing scaffold engine (006-cobalt-crush-architecture)
- Filesystem only (Markdown files deployed to target directory) (006-cobalt-crush-architecture)
- Go 1.24+ (CLI backend), Markdown (coaching agent) + `github.com/spf13/cobra` (CLI), `github.com/charmbracelet/log` (logging), `github.com/charmbracelet/lipgloss` (terminal styling), `embed.FS` (agent embedding) (007-mx-f-architecture)
- JSON files in `.uf/mx-f/data/{source}/{timestamp}.json` for metrics, Markdown+YAML frontmatter in `.uf/mx-f/impediments/` for impediments, `.uf/mx-f/retros/` for retrospective records (007-mx-f-architecture)
- Go 1.24+ (orchestration engine), Markdown (commands, skills) + `internal/artifacts` (envelope, FindArtifacts, WriteArtifact, ReadEnvelope — already exist from Spec 007), `internal/sync` (GHRunner), `github.com/charmbracelet/log` (008-swarm-orchestration)
- JSON files at `.uf/workflows/{workflow_id}.json` (workflow state), `.uf/artifacts/{type}/{timestamp}-{hero}.json` (artifacts) (008-swarm-orchestration)
- Go 1.24+ + `github.com/spf13/cobra` (CLI), `github.com/charmbracelet/lipgloss` (terminal styling), `gopkg.in/yaml.v3` (frontmatter parsing) (011-doctor-setup)
- N/A (reads filesystem and subprocess output, writes only to `opencode.json`) (011-doctor-setup)
- Go 1.24 + `github.com/spf13/cobra` (012-swarm-delegation)
- Go 1.24+ + `github.com/modelcontextprotocol/go-sdk/mcp` (014-dewey-architecture)
- SQLite for persistent indexes (knowledge (014-dewey-architecture)
- N/A (configuration and documentation changes; Dewey MCP tools: `dewey_search`, `dewey_semantic_search`, `dewey_traverse`, `dewey_get_page`, `dewey_find_by_tag`, `dewey_query_properties`, `dewey_find_connections`, `dewey_similar`, `dewey_semantic_search_filtered`) (015-dewey-integration)
- JSON workflow files at `.uf/workflows/` (016-autonomous-define)
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
- Go 1.24+ + `github.com/spf13/cobra` (CLI), `github.com/charmbracelet/lipgloss` (terminal styling), `gopkg.in/yaml.v3` (frontmatter parsing), `net/http` (standard library — new import in `checks.go`) (023-doctor-setup-dewey)
- N/A (reads environment state, writes no persistent data) (023-doctor-setup-dewey)
- Go 1.24+ + `github.com/spf13/cobra` (CLI), `embed.FS` (scaffold), `github.com/charmbracelet/log` (logging), `encoding/json` (opencode.json manipulation) (024-replicator-migration)
- Filesystem only (`opencode.json`, `.uf/replicator/`) (024-replicator-migration)
- Go 1.24+ (scaffold engine, doctor, orchestration, setup), Markdown (agents, commands, skills), Bash (hero contract script) + `github.com/spf13/cobra` (CLI), `embed.FS` (scaffold), `gopkg.in/yaml.v3` (config parsing) (025-uf-directory-convention)
- Filesystem only (path renames across `.uf/`, `.opencode/uf/packs/`, `opencode.json`, `.gitignore`) (025-uf-directory-convention)
- Go 1.24+ (scaffold engine, tests only — no new Go logic) + Markdown (agent files), `embed.FS` (scaffold engine) (026-documentation-curation)
- N/A (Markdown files deployed to target directory) (026-documentation-curation)
- Go 1.24+ + `github.com/spf13/cobra` (CLI), `github.com/charmbracelet/log` (logging), `net/http` (health check), `os/exec` (subprocess) (028-sandbox-command)
- N/A (no persistent state — container lifecycle is transient) (028-sandbox-command)
- Go 1.24+ + `github.com/spf13/cobra` (CLI), `gopkg.in/yaml.v3` (config), `net/http` (Che REST API), `os/exec` (subprocess) (029-sandbox-cde-lifecycle)
- Podman named volumes for persistent state, Eclipse Che / Dev Spaces for CDE workspaces (029-sandbox-cde-lifecycle)
- Markdown (OpenCode command files) + OpenCode runtime, `/review-council` command, `/opsx-propose` artifacts (031-unleash-openspec)
- YAML (GitHub Actions, Peribolos config, Settings App config), Markdown + uwu-tools/peribolos (Go binary, Apache 2.0), Repository Settings App (hosted GitHub App, Probot-based) (032-org-gitops)
- Git repositories (`.github` org repo + per-repo `.github/settings.yml`) (032-org-gitops)
- Go 1.24+ + `net/http`, `net/http/httputil` (stdlib), `github.com/spf13/cobra` (CLI), `github.com/charmbracelet/log` (logging) (033-gateway-command)
- PID file at `.uf/gateway.pid` (plain text) (033-gateway-command)

## Recent Changes

- opsx/gateway-token-refresh-fix: Fixed silent token expiry failures in the gateway's Vertex AI and Bedrock providers. Added token/credential expiry tracking (`tokenExpiry time.Time` on `VertexProvider`, `credExpiry time.Time` on `BedrockProvider`) with 55-minute/50-minute safety margins. When a background refresh fails, the stored token/credentials are now cleared atomically under the write lock, causing `PrepareRequest` to return a clear "Re-authenticate" error instead of silently forwarding stale credentials that Vertex rejects with `ACCESS_TOKEN_TYPE_UNSUPPORTED`. Added proactive refresh in `PrepareRequest` — when a token is within 5 minutes of expiry, a synchronous refresh is attempted with `sync.Mutex.TryLock()` deduplication to prevent thundering herd. Detached gateway child process now redirects stdout/stderr to `.uf/gateway.log` (0600 permissions) instead of discarding logs. `uf gateway status` displays the log file path. Named constants extracted (`vertexTokenLifetime`, `proactiveRefreshWindow`, `bedrockCredLifetime`). Added 14 new test functions including regression test (`TestVertexPrepareRequest_StaleTokenRegression` per TC-006), concurrent deduplication test with `sync.WaitGroup` barrier and `atomic.Int32` counter, and foreground negative case. Updated 8 existing test call sites. Synced 18 pre-existing scaffold asset drifts. Modified files: `internal/gateway/provider.go`, `internal/gateway/gateway.go`, `internal/gateway/gateway_test.go`. No new dependencies. 7 task groups and 42 tasks completed.
- opsx/review-pr-reliability: Fixed 6 reliability issues in `/review-pr` command that caused 12 wasted tool calls and an incomplete review during PR #139. Added argument-first parsing gate (prevents auto-detection when PR number provided), execution mode check in Step 0 (stops early when local tools can't run in plan/read-only mode), mandatory CI coverage matrix in Step 4 (makes CI→local tool dedup decision visible and auditable), save-and-navigate diff handling in Step 5 (replaces nonexistent `gh pr diff -- <path>` file-filter syntax with concrete temp-file technique), GitHub API guidance for PR branch file access (replaces failing `git show`/`git fetch` with `gh api`), and PR-introduced spec detection in Step 6 (finds specs in the PR's changed file list when not on base branch). Fixed pre-existing step numbering issue in Step 10. Modified files: `.opencode/command/review-pr.md`, `internal/scaffold/assets/opencode/command/review-pr.md`, `AGENTS.md`. No Go code changes. 8 task groups and 28 tasks completed.
- opsx/review-pr-command: Added `/review-pr` scaffold command for post-PR GitHub review. Single-agent, token-lean command adapted from org-infra's `review_pr.md` with standardization deltas: kebab-case naming, optional PR number with auto-detection, generic constitution reference, convention pack awareness with graceful degradation, severity pack integration, shell injection mitigations (`--body-file`, JSON input files, branch-name sanitization). Fetches CI check results with causality analysis (PR-caused vs pre-existing), runs local deterministic tools, performs AI review (alignment, security, constitution compliance), offers fix branches for pre-existing CI failures (with dirty-tree guard, user confirmation, collision handling), and offers in-line PR comments (15-comment cap, human confirmation gate). Requires `gh` CLI. New files: `.opencode/command/review-pr.md`, `internal/scaffold/assets/opencode/command/review-pr.md`. Modified files: `internal/scaffold/scaffold_test.go` (expectedAssetPaths 7→8 commands), `cmd/unbound-force/main_test.go` (file count 34→35), `AGENTS.md` (project structure + PR Review section). No Go logic changes. All tasks completed.
- opsx/agent-brief-command: Added `/agent-brief` slash command for AGENTS.md lifecycle management (create, audit, improve). The command auto-detects mode: creates AGENTS.md from project analysis when none exists (hybrid template + LLM approach filling Tier 1 sections from go.mod/package.json, Makefile, CI config, README), or audits existing files against a context-sensitive section taxonomy with scoring (Excellent/Strong/Adequate/Weak/Missing). Context-sensitive Tier 1C sections (Constitution, Spec Framework) are generated or checked only when `.specify/memory/constitution.md`, `specs/`, or `openspec/` are detected. Cross-framework governance bridge check verifies the constitution is stated as governing both Speckit and OpenSpec when both frameworks exist. Bridge file handling ensures CLAUDE.md (`@AGENTS.md` import) and .cursorrules exist. New "Agent Context" doctor check group replaces the single AGENTS.md existence check with 12 deterministic structural checks: file existence, 5 Tier 1 section headers (Project Overview, Build Commands, Project Structure, Code Conventions, Technology Stack), build section code blocks, line count (warn >300), constitution reference (context-sensitive), spec framework description (context-sensitive), CLAUDE.md bridge, .cursorrules bridge. InstallHint changed from "Run: uf init" to "Run: /agent-brief in OpenCode". Scaffold asset distributed via `uf init`. New file: `.opencode/command/agent-brief.md`. Modified files: `internal/doctor/checks.go`, `internal/doctor/doctor.go`, `internal/doctor/doctor_test.go`, `internal/scaffold/scaffold_test.go`. Added 12 new doctor test functions and updated 3 existing tests. All 7 task groups and 51 tasks completed.
- opsx/unified-config: Created `internal/config/` package with unified `Config` struct covering 7 sections (setup, scaffold, embedding, sandbox, gateway, doctor, workflow) and layered `Load()` function (user `~/.config/uf/config.yaml` > repo `.uf/config.yaml` > env vars). Added `uf config` command group with 3 subcommands: `init` (creates/updates `.uf/config.yaml` with commented template), `show` (displays merged config as YAML), `validate` (schema validation with actionable error messages). Added `github.com/goccy/go-yaml` dependency (replacing archived `gopkg.in/yaml.v3` for new code). Setup integration: config-driven `package_manager`, `skip` list, per-tool install methods, embedding model from config. Scaffold integration: removed `workflowConfigContent` constant, config-based language selection via `scaffold.language` field. Gateway integration: config-driven `port` and `provider` defaults. Agent model cleanup: removed hardcoded `model` frontmatter from 24 agent files (12 live + 12 scaffold copies). Absorbs `.uf/sandbox.yaml` into unified config with backward-compatible fallback. Go 1.24+ (CLI), `github.com/goccy/go-yaml` (YAML parsing). All tasks completed.
- opsx/gateway-global-region-error: Replaced silent `"global"` → `"us-east5"` region fallback in `newVertexProvider()` with a clear, actionable error. When `VERTEX_LOCATION` or `CLOUD_ML_REGION` resolves to `"global"`, the gateway now returns an error at startup explaining that Vertex AI `rawPredict`/`streamRawPredict` requires a specific regional endpoint and recommending `ANTHROPIC_VERTEX_REGION` as the override. Changed `newVertexProvider()` signature from `*VertexProvider` to `(*VertexProvider, error)`, propagated through `DetectProvider()` and `NewProviderByName()`. Preserved `"us-east5"` default for the empty-region case. Added 6 new test functions (`TestNewVertexProvider_GlobalRegionError`, `TestNewVertexProvider_CloudMLRegionGlobalAlone`, `TestNewVertexProvider_GlobalOverriddenBySpecificRegion`, `TestNewVertexProvider_EmptyRegionDefault`, `TestDetectProvider_GlobalRegionError`, `TestNewProviderByName_VertexGlobalRegionError`). Updated 3 existing test call sites. Modified files: `internal/gateway/provider.go`, `internal/gateway/gateway_test.go`. No new dependencies. 16 tasks completed.
- 034-gateway-vertex-translation: Extended `uf gateway` Vertex provider with full request/response translation. Request body transformation: removes `model` field (Vertex uses URL path), injects `anthropic_version: "vertex-2023-10-16"` if absent, preserves existing `anthropic_version`, updates `Content-Length` after body modification. Streaming endpoint detection: selects `streamRawPredict` when `"stream": true`, `rawPredict` otherwise (`count_tokens` always uses `rawPredict`). Header stripping: removes `anthropic-beta` and `anthropic-version` HTTP headers before forwarding to Vertex. SSE response filtering: created `internal/gateway/sse.go` with `sseFilterReader` (io.ReadCloser wrapper applied via `ModifyResponse`) that drops `vertex_event` and `ping` SSE events from Vertex streaming responses while forwarding all standard Anthropic events (`message_start`, `content_block_delta`, `content_block_stop`, `message_delta`, `message_stop`) unchanged. Non-streaming responses pass through unfiltered. Synthetic model catalog: added `/v1/models` endpoint returning 9 Vertex-available Claude models (Opus 4.7, Sonnet 4.6, Opus 4.6, Opus 4.5, Sonnet 4.5, Opus 4.1, Haiku 4.5, Opus 4, Sonnet 4) with `capabilities` metadata (vision, extended_thinking, pdf_input). Added `/v1/models/{model_id}` for single-model lookup with 404 for unknown models. Sandbox provider isolation verified: `gatewaySkippedKeys` prevents Vertex env vars (`CLAUDE_CODE_USE_VERTEX`, `ANTHROPIC_VERTEX_PROJECT_ID`, `VERTEX_LOCATION`, `GOOGLE_CLOUD_PROJECT`) from leaking into container when gateway active. New file: `internal/gateway/sse.go`. Modified files: `internal/gateway/provider.go`, `internal/gateway/gateway.go`, `internal/gateway/gateway_test.go`. All stdlib — no new dependencies (`net/http`, `net/http/httputil`, `bufio`, `bytes`). Backward compatible: Anthropic and Bedrock providers unchanged (FR-010). All 5 user stories and 65 tasks completed.
- opsx/quickstart-usage-docs: Created QUICKSTART.md and USAGE.md onboarding documentation. QUICKSTART.md (~110 lines) covers installation for macOS (Homebrew) and Fedora/RHEL (Homebrew recommended, dnf minimal with dynamic RPM version via GitHub API), maintainer journey (`uf init`), contributor journey (`uf setup` + `uf doctor`), and first-use walkthrough (`/review-council`). USAGE.md (~150 lines) covers OpenCode modes and agents orientation (primary modes vs subagents, Tab switching vs @mention vs slash commands), 5 task-oriented workflow recipes (review, propose, feature, unleash, quality), workflow decision table, convention pack customization, and 13-command quick reference. Updated README.md to replace install section with pointer to QUICKSTART.md. Filed upstream issues for agent mode fixes: unbound-force/gaze#91 (add `mode: subagent` to gaze-reporter.md and gaze-test-generator.md) and unbound-force/replicator#12 (add `mode: subagent` to coordinator.md, worker.md, background-worker.md). No Go code changes -- documentation only. 29 tasks completed.
- 033-gateway-command: Added `uf gateway` command with 3 subcommands (`start`, `stop`, `status`) for LLM reverse proxy gateway. Created new `internal/gateway/` package (5 source files: `gateway.go`, `provider.go`, `refresh.go`, `pid.go`, `signal_unix.go`) following the established `Options`/injectable-function pattern. Gateway auto-detects cloud provider from environment variables (Vertex AI → Bedrock → Anthropic priority) and injects host-side credentials into upstream requests. Three provider implementations: `AnthropicProvider` (API key injection), `VertexProvider` (gcloud token refresh, rawPredict URL rewriting), `BedrockProvider` (AWS credential refresh, SigV4 request signing). Background mode via re-exec with `_UF_GATEWAY_CHILD` sentinel. PID file management at `.uf/gateway.pid` with atomic writes and stale cleanup. Health endpoint at `GET /health` returns JSON status. Reverse proxy routes `POST /v1/messages` and `POST /v1/messages/count_tokens` to upstream with credential injection. Token refresh goroutines for Vertex (gcloud) and Bedrock (aws CLI) with configurable intervals. Minimal SigV4 signing implementation using `crypto/hmac` + `crypto/sha256` (~200 lines, no AWS SDK dependency). Sandbox integration: `autoStartGateway()` in `internal/sandbox/sandbox.go` auto-starts gateway when cloud provider detected, passes `ANTHROPIC_BASE_URL` and `ANTHROPIC_AUTH_TOKEN` to container, skips credential mounts when gateway active, falls back to credential mounts when no provider detected. Created `cmd/unbound-force/gateway.go` with Cobra command registration (`--port`, `--provider`, `--detach` flags). Added 29 gateway test functions and 12 sandbox integration test functions. Go 1.24+ (CLI), `net/http`, `net/http/httputil` (stdlib), `github.com/charmbracelet/log`. All 5 user stories and 87 tasks completed.
- 031-unleash-openspec: Extended `/unleash` command (`.opencode/command/unleash.md`) to support OpenSpec (`opsx/*`) branches alongside existing Speckit (`NNN-*`) branches. Replaced the `opsx/*` hard STOP with OpenSpec mode detection — extracts change name from `opsx/<name>`, sets `FEATURE_DIR = openspec/changes/<name>/`, sets `WORKFLOW_TIER = openspec`, verifies `tasks.md` exists (gates on `/opsx-propose` completion). Added skip logic for clarify/plan/tasks steps (handled by `/opsx-propose`). Updated resumability detection to mark checks 1-3 as always "done" for OpenSpec mode, with checks 4-6 (spec-review/implementation/code-review markers) reading from `FEATURE_DIR/tasks.md` for both tiers. Updated spec review and code review steps to pass `FEATURE_DIR` to the review council. Updated demo step to read `proposal.md` instead of `spec.md` for OpenSpec changes. Updated guardrails to reference both branch patterns. Synchronized scaffold asset copy. No Go code changes — Markdown only (~58 lines added). All 4 user stories and 17 tasks completed.
- 029-sandbox-cde-lifecycle: Extended `internal/sandbox/` package with persistent workspace lifecycle (`create`/`destroy`) and CDE backend (Eclipse Che / Dev Spaces). Introduced `Backend` interface (Strategy pattern per SOLID Open/Closed) with two implementations: `PodmanBackend` (named volumes `uf-sandbox-<project>` for persistent state, per-project container names) and `CheBackend` (workspace provisioning via `chectl` CLI primary path with REST API fallback). Created 4 new source files: `backend.go` (interface + `ResolveBackend()` with resolution matrix: flag > env > config > auto-detect), `podman.go` (Create/Start/Stop/Destroy/Status/Attach with named volume lifecycle and partial failure cleanup), `che.go` (Create/Start/Stop/Destroy/Status/Attach via chectl or REST API with devfile-based provisioning), `workspace.go` (WorkspaceStatus, SandboxConfig with YAML parsing, DemoEndpoint, projectName sanitization, LoadConfig with env var overrides, FormatWorkspaceStatus, bidirectional git sync via setupGitSync/checkGitSync). Extended `Options` struct with `BackendName`, `WorkspaceName`, `DemoPorts`, `ConfigPath`, `CheURL`, `HTTPDo` fields (all zero-value defaults preserve Spec 028 behavior). Updated `sandbox.go` with `Create()` and `Destroy()` dispatch functions, persistent workspace detection in `Start()`/`Stop()`/`Extract()`, `WorkspaceStatusCheck()`. Updated `cmd/unbound-force/sandbox.go` with 2 new subcommands (`create` with `--backend`/`--demo-ports`/`--name` flags, `destroy` with `--yes`/`--force` and confirmation prompt), updated `start` with `--backend` flag, updated `stop`/`status` for persistent workspace support. Added 43 new test functions (82 total) covering backend resolution, PodmanBackend lifecycle, CheBackend chectl/REST paths, workspace helpers, config parsing, backward compatibility, git sync, demo endpoints. All existing Spec 028 tests pass (backward compatible). Go 1.24+ (CLI), `gopkg.in/yaml.v3` (config), `net/http` (Che REST API). All 4 user stories and 111 tasks completed.
- opsx/workflow-phase-boundaries: Added workflow phase boundary enforcement (closes #92 + #94). Externalized 9 `speckit.*.md` command files from scaffold assets to `specify init` + `/uf-init` (embedded asset count 42 → 33). Added "Phase Discipline" MUST rule to constitution Development Workflow section. Added "Workflow Phase Boundaries" subsection to AGENTS.md Behavioral Constraints. Added `<!-- code-review: passed -->` marker to `/unleash` Step 6 (code review) with matching resumability detection in Step 2. Added 3 new sections to `/uf-init`: "Speckit Custom Commands" (creates 4 UF-custom commands: analyze, checklist, clarify, taskstoissues), "Speckit Command Guardrails" (injects `## Guardrails` into all 9 speckit commands), "Speckit UF Customizations" (verifies Dewey/constitution/review-council references). Deleted stray tool directories from `internal/scaffold/` and `cmd/unbound-force/` (~74 files). Added `.gitignore` patterns to prevent stray recurrence (`cmd/**/.opencode/`, `internal/**/.uf/`, etc.). Updated `expectedAssetPaths` (42 → 33), added 9 entries to `knownNonEmbeddedFiles`, updated file count assertion (42 → 33). Synchronized 2 scaffold asset copies (unleash.md, uf-init.md). No new Go logic -- Markdown + test assertion changes only. 27 tasks completed.
- 028-sandbox-command: Added `uf sandbox` command with 5 subcommands (`start`, `stop`, `attach`, `extract`, `status`) for containerized OpenCode session management via Podman. Created new `internal/sandbox/` package (3 source files: `sandbox.go`, `detect.go`, `config.go`) following the established `Options`/`ExecCmd` injection pattern. `Start()` checks prerequisites (Podman, OpenCode), detects platform (macOS/Linux, SELinux), pulls image if needed, starts container with API key forwarding and resource limits, waits for health check with exponential backoff (500ms→5s, 60s timeout), and attaches TUI (or prints URL in `--detach` mode). `Extract()` generates patches via `git format-patch` inside container, displays summary, and applies via `git am` on confirmation. `Stop()` is idempotent. `Status()` parses `podman inspect` JSON. `Attach()` delegates to `opencode attach`. Two mount modes: isolated (read-only, default) and direct (read-write). SELinux `:Z` volume flag auto-detection on Linux. Dead container cleanup before start. Created `cmd/unbound-force/sandbox.go` with Cobra command registration (`--mode`, `--detach`, `--image`, `--memory`, `--cpus`, `--yes` flags). Added 39 test functions covering all public functions, error paths, platform detection, health check polling, and configuration builders. No persistent state -- container existence is the state. Go 1.24+ (CLI), `net/http` (health check), `os/exec` (subprocess). All 4 user stories and 87 tasks completed.
- 027-externalize-tool-init: Externalized Speckit, OpenSpec, and Gaze initialization from embedded scaffold assets to CLI delegation. Removed 13 embedded assets (12 Speckit files from `internal/scaffold/assets/specify/` + 1 `openspec/config.yaml`), reducing embedded asset count from 55 to 42. Added 3 tool delegations to `initSubTools()` in `internal/scaffold/scaffold.go`: `specify init` (gated on `.specify/` absence), `openspec init --tools opencode` (gated on `openspec/config.yaml` absence), `gaze init` (gated on `gaze-reporter.md` absence). Added 2 new `uf setup` steps: `installUV()` (Homebrew-first with curl fallback) and `installSpecify()` (via `uv tool install specify-cli`, gated on uv availability), increasing step count from 12 to 14. Removed `"specify/"` from `knownAssetPrefixes` and `mapAssetPath()`. Updated `expectedAssetPaths` (55 → 42), added 13 entries to `knownNonEmbeddedFiles`, updated `TestRunInit_FreshDir` file count (55 → 42). Added 18 new test functions (7 setup + 11 scaffold delegation). Updated `/uf-init` command with tool delegation documentation. All 4 user stories and 41 tasks completed.
- 026-documentation-curation: Added The Curator Divisor agent (`divisor-curator.md`) for documentation & content pipeline triage -- detects documentation gaps (AGENTS.md/README.md not updated for user-facing changes), identifies blog opportunities (new agents, CLI commands, migrations), identifies tutorial opportunities (new slash commands, workflow patterns), and files GitHub issues in `unbound-force/website` with labels `docs`/`blog`/`tutorial`. First Divisor agent with `bash: true` (restricted to `gh issue create` and `gh issue list` only). Temperature 0.2 (judgment-based decisions). Added "Documentation Completeness" checklist item (#6) to Guard's Code Review audit. Added Curator row to review council reference table. Synchronized 3 scaffold asset copies. Updated `expectedAssetPaths` (54 → 55 files). Updated Divisor agent count assertion (8 → 9). All 4 user stories and 38 tasks completed.
- opsx/gitignore-init: Added `ensureGitignore()` function to the scaffold engine (`internal/scaffold/scaffold.go`) that appends a standard Unbound Force ignore block to `.gitignore` during `uf init`. The block covers `.uf/` runtime data (databases, caches, locks, logs) and legacy tool directories (`.dewey/`, `.hive/`, `.unbound-force/`, `.muti-mind/`, `.mx-f/`). Idempotent via marker comment detection (`# Unbound Force — managed by uf init`). Creates `.gitignore` if it does not exist. Called from `Run()` after file scaffolding but before sub-tool delegation. Result included in scaffold summary output. Added 4 test functions (`TestEnsureGitignore_FreshDir`, `TestEnsureGitignore_ExistingNoBlock`, `TestEnsureGitignore_ExistingWithBlock`, `TestEnsureGitignore_Idempotent`). Updated `TestScaffoldOutput_NoOldPathReferences` to exclude `.gitignore` from stale pattern check (legacy directories are intentional ignore patterns). No scaffold asset changes. 9 tasks completed.
- 025-uf-directory-convention: Unified all tool workspace directories under `.uf/` -- renamed `.dewey/` to `.uf/dewey/`, `.hive/` to `.uf/replicator/`, `.unbound-force/` to `.uf/`, `.muti-mind/` to `.uf/muti-mind/`, `.mx-f/` to `.uf/mx-f/`, and `.opencode/unbound/packs/` to `.opencode/uf/packs/`. Updated scaffold engine (`internal/scaffold/scaffold.go`), doctor checks (`internal/doctor/checks.go`), orchestration engine (`internal/orchestration/`), Muti-Mind CLI (`cmd/mutimind/main.go`), metrics store comment, all 13 scaffold asset Markdown files, all 21 live agent/command Markdown files, `.gitignore`, `scripts/validate-hero-contract.sh`, hero-manifest schema, acceptance-decision sample, pack validator test comment, and AGENTS.md Project Structure + Active Technologies sections. Added `TestScaffoldOutput_NoOldPathReferences` regression test (6 stale patterns). No new Go logic -- purely mechanical path rename (~260 references). All 5 user stories and 79 tasks completed.
- opsx/gatekeeping-protection: Added gatekeeping value protection rules to prevent AI agents from modifying quality/governance gates to make implementations pass. Added "Gatekeeping Integrity" MUST rule to constitution Development Workflow section (CRITICAL severity violation). Added "Gatekeeping Value Protection" subsection to AGENTS.md Behavioral Constraints with 8 protected value categories (coverage thresholds, severity definitions, convention pack classifications, CI flags, agent settings, constitution MUST rules, review limits, workflow markers) and "what to do instead" instruction. Added "Gatekeeping Integrity" checklist item to `divisor-guard.md` in both Code Review and Spec Review audit sections. Added "Gate Tampering" checklist item to `divisor-adversary.md` Security Checks section. Added "Gatekeeping Integrity" behavioral constraint to `cobalt-crush-dev.md` Engineering Philosophy section. Synchronized 3 scaffold asset copies. No Go code changes -- behavioral rules only. 13 tasks completed.
- opsx/divisor-content-agents: Added 3 content-creation Divisor agents -- `divisor-scribe.md` (The Scribe: technical documentation, temp 0.1), `divisor-herald.md` (The Herald: blog/announcements, temp 0.4), `divisor-envoy.md` (The Envoy: PR/communications, temp 0.5). Created `content.md` convention pack (tool-owned, language-agnostic, always-deploy) with 6 sections: Voice & Brand (VB), Technical Documentation (TD), Blog & Announcements (BA), Public Relations (PR), Fact-Checking (FA), Formatting (FT). Created `content-custom.md` user-owned stub. Updated `shouldDeployPack()` to always deploy content/content-custom alongside default and severity. Added 5 scaffold asset copies. Updated `expectedAssetPaths` (added 5 files). All 3 agents have write+edit access, Dewey integration, and Out of Scope boundaries. 18 tasks completed.
- 024-replicator-migration: Replaced Node.js Swarm plugin (`opencode-swarm-plugin`) with Replicator Go binary across all CLI commands. `uf setup` now installs Replicator via Homebrew (`brew install unbound-force/tap/replicator`) and runs `replicator setup` for per-machine init -- step count reduced from 15 to 12. Removed `installSwarmPlugin()`, `ensureBun()`, `runSwarmSetup()`, `initializeHive()`, `swarmForkSource` constant, and `uf init` call from setup. Removed all bun references -- OpenSpec CLI installs via npm only. `uf init` now adds `mcp.replicator` server entry to `opencode.json` (instead of `opencode-swarm-plugin` plugin array entry), detects via `LookPath("replicator")` (instead of `.hive/` directory), delegates to `replicator init` for per-repo setup, and migrates legacy `opencode-swarm-plugin` plugin entries. `uf doctor` replaced "Swarm Plugin" check group with "Replicator" group -- binary check, `replicator doctor` delegation, `.hive/` existence, `mcp.replicator` config verification. Updated all install hints from npm/bun to Homebrew. Updated `opencode.json` (removed `plugin` array, added `mcp.replicator`). Updated `/unleash` install hint text. Added 22 new test functions, removed 10 old ones, updated ~40 existing. Go 1.24+ (CLI), Markdown (commands). All 4 user stories and 63 tasks completed.
- 023-doctor-setup-dewey: Updated `uf doctor` and `uf setup` for Dewey unified memory (Spec 021 FR-019 through FR-021). Added Dewey embedding capability health check to `uf doctor` -- probes Ollama `/api/embed` endpoint to verify embeddings work end-to-end, with categorized error hints (connection refused, model not found). Added `EmbedCheck` injectable function field on `Options` struct following established `LookPath`/`ExecCmd` pattern. Changed Swarm plugin install source in `uf setup` from upstream `opencode-swarm-plugin@latest` to forked `github:unbound-force/swarm-tools` (removed early-return skip, always installs to ensure fork version). Updated all install hints in `environ.go` and `checks.go` to reference fork. Annotated embedding model check with "(Dewey manages Ollama lifecycle)" when Dewey is installed. Added 8 new test functions across `doctor_test.go` and `setup_test.go`. Updated existing test assertions for fork source. `net/http` (standard library -- new import in `checks.go`). All 3 user stories and 38 tasks completed.
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

## Convention Packs

This repository uses convention packs scaffolded by
unbound-force. Agents MUST read the applicable pack(s)
before writing or reviewing code.

- `.opencode/uf/packs/default.md`
- `.opencode/uf/packs/default-custom.md`
- `.opencode/uf/packs/severity.md`
- `.opencode/uf/packs/content.md`
- `.opencode/uf/packs/content-custom.md`
- `.opencode/uf/packs/go.md`
- `.opencode/uf/packs/go-custom.md`
