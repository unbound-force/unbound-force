# Architecture

## Why Unbound Force

Software development is shifting. Engineers increasingly
direct AI agents through specifications and rules rather
than writing every line by hand. But managing multiple AI
agents introduces its own problems: agents lack memory
between sessions, reviews are inconsistent, quality
drifts without metrics, and there is no structured way
to go from an idea to a reviewed pull request.

Unbound Force solves this by providing a complete agent
team -- each with a defined role, shared standards, and
artifact-based communication -- that works as a
coordinated swarm. Here is what that gives you
concretely:

**Consistent code review.** Instead of one reviewer
catching what they happen to notice, 9 specialized
personas audit every change from different angles --
security, architecture, testing, operations, intent
drift, documentation, and more. Findings come with
severity levels and actionable recommendations.

**Spec-driven development.** Changes start with a
specification, not a blank editor. The system enforces
a pipeline from requirements to implementation to
review, preventing the "I built the wrong thing"
problem. Two pipeline tiers handle everything from a
quick bug fix to a multi-story feature.

**Quality that only goes up.** CRAP scores, coverage
ratchets, and side effect classification give you
measurable quality baselines. Coverage can only
increase. High-risk code is surfaced automatically.

**Shared memory across sessions.** Dewey indexes your
specs, decisions, and learnings. When an agent starts
a new task, it queries prior context -- what patterns
worked, what gotchas exist, what the spec requires --
instead of starting from zero.

**Parallel execution.** Large tasks decompose into
subtasks that run concurrently with file-level
reservation to prevent conflicts. What takes one agent
an hour takes four agents fifteen minutes.

**Works alone or together.** Every tool is independently
useful. Install just the review council for better code
reviews. Add quality analysis when you're ready. Layer
on specification workflows when the team grows. Nothing
is mandatory except what you choose.

---

This document explains how the ecosystem fits together --
the components, how they connect, how data flows between
them, and how to use the system effectively.

## The Ecosystem at a Glance

```
                          You (Engineer)
                               |
                    +----------+-----------+
                    |                      |
                 Terminal               OpenCode
                    |                   (AI IDE)
                    v                      |
              +----------+     +-----------+-----------+
              |  uf CLI  |     |    Slash Commands      |
              |----------|     |   /review-council      |
              | init     |     |   /opsx-propose        |
              | doctor   |     |   /speckit.specify     |
              | setup    |     |   /gaze   /forge       |
              | sandbox  |     |   /unleash  ...        |
              | gateway  |     +-----------+-----------+
              | config   |                 |
              +----+-----+     +-----------+-----------+
                   |           |                       |
            +------+------+    |    Agent Personas     |
            |             |    |  (Cobalt-Crush, etc.) |
            v             v    +-----------+-----------+
    +------------+  +----------+           |
    | Replicator |  |  Dewey   |    +------+------+
    | (swarm)    |  | (search) |    |  Convention |
    +------------+  +----------+    |    Packs    |
                                    +-------------+
            +-------------+
            |    Gaze      |
            | (quality)    |
            +-------------+

    +--------------------------------------------+
    |            Specification Layer              |
    |  Speckit (strategic)  |  OpenSpec (tactical)|
    +--------------------------------------------+
```

### Components

**uf CLI** -- A Go binary (`unbound-force`, aliased as
`uf`). Available via Homebrew, dnf (Fedora/RHEL), or
direct download from GitHub releases. See
[QUICKSTART.md](../QUICKSTART.md) for installation
options. Provides seven commands: `init`, `doctor`,
`setup`, `sandbox`, `gateway`, `config`, and `version`.
This is the primary distribution mechanism for the
framework. When you run `uf init`, it scaffolds agents,
commands, convention packs, and configuration into your
project.

**OpenCode** -- An AI coding IDE (opencode.ai) that
provides the interactive environment where agents and
slash commands run. Unbound Force is designed for
OpenCode but the scaffolded files are portable Markdown
that can be adapted for other AI coding tools.

**Speckit** -- A strategic specification framework
(github/spec-kit) for features with 3+ user stories or
cross-repo scope. Provides a 9-phase pipeline from
constitution check through implementation.

**OpenSpec** -- A tactical specification framework built
into `uf init`. Handles bug fixes and small changes
(under 3 user stories) with a lightweight propose/apply
workflow.

**Gaze** -- A Go quality analysis tool
(unbound-force/gaze). Computes CRAP scores, coverage
metrics, and side effect classifications. Generates
tests for weak spots. Separate repository, installed via
`uf setup` or directly from the Gaze repository.

**Dewey** -- A semantic knowledge layer
(unbound-force/dewey) that indexes project markdown
files and makes them searchable via graph traversal and
vector-based semantic search. Agents query Dewey through
MCP tools to find specs, patterns, and prior decisions
without loading entire files into context.

**Replicator** -- A swarm coordination tool
(unbound-force/replicator) that enables parallel task
execution. The Forge subsystem decomposes tasks and
spawns worker agents. The Comms subsystem handles
inter-agent messaging and file reservation. The Org
subsystem tracks work items.

**The Divisor** -- A review council embedded in `uf init`
with 9 agent personas. Each persona reviews code or specs
from a different angle (security, architecture, testing,
operations, intent drift, and more). Not a separate
repository -- distributed through the `uf` binary.

**Muti-Mind** -- The product owner agent. Manages
backlogs, generates user stories, syncs with GitHub
issues, and makes acceptance decisions. Exists as an
embedded agent persona plus a separate `mutimind` CLI
binary built from this repository.

**Cobalt-Crush** -- The developer agent persona. Writes
code following specs, convention packs, and engineering
principles. Embedded in `uf init` as an agent file -- no
separate binary.

**Mx F** -- The manager agent. Facilitates retrospectives,
tracks team metrics, coaches on process improvement, and
removes impediments. Exists as an embedded coaching agent. The `mxf` CLI
(unbound-force/mxf) is available separately but is not
installed by `uf setup`.


## What `uf init` Does

When you run `uf init` in a project directory, it scaffolds
files organized into two ownership categories:

**Tool-owned files** are auto-updated when you re-run
`uf init`. If the upstream content has changed, the file
is overwritten. These files carry a version marker
comment.

**User-owned files** are created once and never
overwritten on subsequent runs (unless you pass
`--force`, which resets everything). You customize
these freely.

### Scaffolded Files

**Slash commands** (tool-owned, 8 core files scaffolded
by `uf init`, plus more from companion tools):
```
.opencode/commands/
  review-council.md       review-pr.md
  cobalt-crush.md         agent-brief.md
  opsx-propose.md         opsx-apply.md
  opsx-explore.md         opsx-archive.md
  unleash.md              finale.md
  uf-init.md              constitution-check.md
```

Companion tools add their own commands during `uf init`:
Speckit adds the `/speckit.*` pipeline commands, Gaze
adds `/gaze` and `/gaze-fix`, Replicator adds `/forge`,
`/forge-status`, `/handoff`, `/inbox`, and `/org`, and
Muti-Mind adds the `/muti-mind.*` backlog commands.
A full scaffold produces ~46 command files total.

**Agent personas** (user-owned, 12 active agents):
```
.opencode/agents/
  cobalt-crush-dev.md     muti-mind-po.md
  mx-f-coach.md           gaze-reporter.md
  gaze-test-generator.md  divisor-guard.md
  divisor-architect.md    divisor-adversary.md
  divisor-sre.md          divisor-testing.md
  divisor-curator.md      divisor-herald.md
```

Plus specialized agents (coordinator, worker,
background-worker, divisor-scribe, divisor-envoy)
for swarm and review operations.

**Convention packs** (mixed ownership):
```
.opencode/uf/packs/
  default.md          (tool-owned, language-agnostic)
  default-custom.md   (user-owned, your extensions)
  severity.md         (tool-owned, finding severity)
  go.md               (tool-owned, Go conventions)
  go-custom.md        (user-owned, your Go extensions)
  typescript.md       (tool-owned, TS conventions)
  typescript-custom.md (user-owned, your TS extensions)
  content.md          (tool-owned, writing standards)
  content-custom.md   (user-owned, your content rules)
```

Tool-owned packs contain canonical rules. User-owned
`*-custom.md` files are where you add project-specific
rules. Language packs are selected by auto-detection
(checks for `go.mod`, `package.json`, etc.) or by the
`--lang` flag.

**OpenSpec schema** (tool-owned, 5 files):
```
openspec/
  config.yaml
  schemas/unbound-force/
    schema.yaml
    templates/
      proposal.md   design.md
      spec.md        tasks.md
```

The schema defines the structure of tactical changes --
what artifacts are required (proposal, design, specs,
tasks), their dependency order, and the templates agents
use to create them. Without the schema, agents would
generate inconsistent artifacts with arbitrary structure.
The `unbound-force` schema adds constitution alignment
checks that the generic OpenSpec format does not require,
ensuring every change is assessed against the org's four
governance principles.

**Skills** (tool-owned):
```
.opencode/skills/
  speckit-workflow/SKILL.md
  openspec-propose/SKILL.md
  openspec-apply-change/SKILL.md
  openspec-archive-change/SKILL.md
  openspec-explore/SKILL.md
```

Skills are reusable instruction packages that teach
agents complex multi-step workflows. Rather than
embedding long workflow instructions into every command
or agent file, skills are loaded on demand when a task
matches their description. For example, when you run
`/opsx-propose`, the system loads the `openspec-propose`
skill which contains the complete artifact creation
workflow -- how to read the schema, which artifacts to
create in what order, how to check dependencies, and
when to stop. Skills keep agent files focused on
identity and behavior while workflows stay modular and
independently updatable.

**Bridge files**:
```
CLAUDE.md          (tool-owned, references AGENTS.md +
                    agent + convention pack files)
.cursorrules       (tool-owned, Cursor IDE equivalent)
.gitignore block   (appended, ignores .uf/ runtime data)
```

### Sub-tool Delegation

During `uf init`, the scaffold engine detects and
delegates to companion tools when available:

- `dewey init` -- Indexes project markdown for semantic
  search
- `replicator init` -- Sets up swarm coordination
  database
- `specify init` -- Initializes Speckit templates and
  scripts
- `openspec init` -- Creates OpenSpec schema and config
- `gaze init` -- Configures quality analysis settings

Each sub-tool init is optional. If the tool is not
installed, that step is silently skipped.

### MCP Configuration

`uf init` creates or updates `opencode.json` to register
MCP servers:

```json
{
  "mcp": {
    "dewey": {
      "type": "local",
      "command": ["dewey", "serve", "--vault", "."]
    },
    "replicator": {
      "type": "local",
      "command": ["replicator", "serve"]
    }
  }
}
```

This gives OpenCode agents access to Dewey's semantic
search tools and Replicator's swarm coordination tools
via the Model Context Protocol.

### File Ownership Model

The scaffold engine tracks ownership per file:

| Ownership | On re-run | Customize? | Examples |
|-----------|-----------|------------|----------|
| Tool-owned | Updated if content changed | No (changes are overwritten) | Commands, canonical packs, schemas |
| User-owned | Skipped (never overwritten) | Yes | Agents, custom packs, templates |

Use `--force` to overwrite all files regardless of
ownership (destructive -- resets your customizations).


## The Hero Team

| Hero | Role | What It Does Today | Where It Lives |
|------|------|--------------------|----------------|
| **Muti-Mind** | Product Owner | Backlog management, user story generation, GitHub issue sync, acceptance decisions | Embedded agent + `mutimind` CLI in this repo |
| **Cobalt-Crush** | Developer | Spec-driven code implementation, convention pack adherence, test hook generation | Embedded agent in `uf init` |
| **Gaze** | Tester | CRAP scores, coverage metrics, side effect classification, test generation (Go) | Separate repo: unbound-force/gaze |
| **The Divisor** | Reviewer | 9-persona review council, auto-detected code vs spec mode, convention-based criteria | Embedded in `uf init` (9 agent files) |
| **Mx F** | Manager | Retrospectives, coaching, metrics tracking, impediment removal | Embedded agent in `uf init` |

See [docs/heroes.md](heroes.md) for full hero
descriptions with current vs planned capabilities.


## Specification Pipelines

All non-trivial changes follow a specification workflow
before implementation begins. Two pipelines handle
different scales of work.

### Decision: Which Pipeline?

```
  How many user stories?
         |
    +----+----+
    |         |
  < 3       3+       Not sure?
    |         |         |
    v         v         v
  OpenSpec  Speckit   Start with
                      OpenSpec,
                      escalate if
                      scope grows
```

### Strategic Pipeline (Speckit)

For features with 3+ user stories or cross-repo impact.
Creates artifacts in `specs/NNN-feature-name/`.

```
/speckit.specify    Define the feature (spec.md)
       |
       v
/speckit.clarify    Refine requirements (optional)
       |
       v
/speckit.plan       Generate implementation plan
       |
       v
/speckit.tasks      Break into ordered task list
       |
       v
/speckit.analyze    Check spec consistency (optional)
       |
       v
/speckit.checklist  Quality validation (optional)
       |
       v
/speckit.implement  Execute the task list
       |
       v
/review-council     Run the Divisor review council
       |
       v
/finale             Commit, push, create PR, merge
```

### Tactical Pipeline (OpenSpec)

For bug fixes, minor enhancements, and tasks under 3
user stories. Creates artifacts in
`openspec/changes/change-name/`.

```
/opsx-propose    Create proposal + design + tasks
       |
       v
/opsx-apply      Implement tasks from the change
       |
       v
/review-council  Run the Divisor review council
       |
       v
/finale          Commit, push, create PR, merge
```

### Autonomous Mode

```
/unleash
```

Runs the full pipeline autonomously -- clarify, plan,
tasks, spec review, implement, code review, and
retrospective. Works with both Speckit (`NNN-*` branches)
and OpenSpec (`opsx/*` branches). Exits to human judgment
only when it encounters ambiguity, review failures, or
merge conflicts.


## The Knowledge Layer (Dewey)

Dewey indexes all markdown files in your project --
specs, plans, tasks, constitutions, convention packs --
into a searchable knowledge graph with vector embeddings.

### What Gets Indexed

- Specification documents (`specs/`, `openspec/`)
- Constitution and governance files
- Agent personas and command definitions
- Convention packs and design records
- Any other markdown in the project

### How Agents Use It

Agents query Dewey through MCP tools registered in
`opencode.json`:

| Tool | Purpose |
|------|---------|
| `dewey_semantic_search` | Find conceptually related content |
| `dewey_search` | Keyword search across all indexed files |
| `dewey_traverse` | Navigate relationships between documents |
| `dewey_find_by_tag` | Find content by metadata tags |
| `dewey_get_page` | Retrieve a specific indexed page |

Example: When Cobalt-Crush starts implementing a task,
it queries Dewey for prior learnings about the target
files, related specs governing the feature, and
architectural patterns from conventions -- all before
reading source code.

### 3-Tier Degradation

Dewey availability degrades gracefully:

| Tier | Availability | Capabilities |
|------|-------------|--------------|
| **Tier 3** (full) | Dewey running with embedding model | Semantic search + graph traversal + metadata queries |
| **Tier 2** (graph-only) | Dewey running, no embedding model | Keyword search + graph traversal (no semantic similarity) |
| **Tier 1** (no Dewey) | Dewey not installed | Agents fall back to direct file reads and grep |

Every agent works without Dewey. This follows
Constitution Principle II (Composability First) -- no
hero requires another hero to function.


## Code Review (The Divisor)

The Divisor operates as a council of specialized reviewer
personas. Each persona focuses on a different dimension
of code quality.

### The Council

| Persona | Focus Area |
|---------|-----------|
| **Guard** | Intent drift, constitution alignment, zero-waste |
| **Architect** | Structure, DRY/SOLID, conventions, tech debt |
| **Adversary** | Security, error handling, resilience, edge cases |
| **SRE** | Operations, performance, observability, scaling |
| **Testing** | Test quality, assertions, coverage strategy |
| **Curator** | Documentation, naming, API surface consistency |
| **Scribe** | Changelog entries, commit messages, release notes |
| **Herald** | Cross-repo impact, downstream effects, migration |
| **Envoy** | External dependency changes, license compliance |

### Auto-Detection

The council auto-detects whether to review code or
specs based on what changed on your branch. If the
branch contains spec artifacts (`.md` files in `specs/`
or `openspec/`), personas switch to spec review mode.

### Convention Packs as Review Criteria

Convention packs define the rules that Divisor agents
enforce. Rules are tagged by severity:

- `[MUST]` -- Mandatory. Violation blocks the merge.
- `[SHOULD]` -- Strong recommendation. Requires
  justification to skip.
- `[MAY]` -- Optional improvement. Noted but not
  blocking.

### Review Commands

| Command | When | What Happens |
|---------|------|--------------|
| `/review-council` | Pre-PR (local) | Discovers available personas, runs them in parallel, produces APPROVE or REQUEST CHANGES |
| `/review-pr [N]` | Post-PR (GitHub) | Single agent reviews a specific PR, analyzes CI results |
| `/address-feedback [N]` | Post-PR (GitHub) | Triage + address reviewer feedback |

### Addressing Feedback

The `/address-feedback` command closes the gap between
PR creation and merge. After reviewers comment on a PR,
the author runs `/address-feedback` to systematically
ingest, assess, triage, and respond to every piece of
feedback.

**Four-phase workflow**:
1. **Ingest** -- Fetch all unresolved review threads
   from GitHub, filtering out resolved and outdated
   items.
2. **Assess** -- Classify each item as data-driven or
   subjective using convention packs and constitution
   principles as evidence.
3. **Triage** -- Present each item to the author with a
   recommendation. The author chooses one of four
   decisions: accept, modify, reject, or ask.
4. **Execute** -- Commit code fixes (one per item), run
   `/review-council` on the changes, push, and post
   reply comments to each thread.

**Tiered assessment architecture**:

- **Tier 1 (direct)** -- The command agent assesses
  simple items itself: formatting, naming, typos, clear
  convention matches, single-file scope. Fast (~2-3s
  per item).
- **Tier 2 (Divisor escalation)** -- Complex items are
  delegated to the relevant Divisor persona via Task
  tool: security concerns to the Adversary,
  architectural changes to the Architect, test strategy
  to Testing, and so on. Slower (~10-20s) but uses
  specialized expertise.

If Divisor agents are not available, all items fall
back to Tier 1. This follows Constitution Principle II
(Composability First).

**Authority matrix**: The command adjusts
recommendations based on reviewer role (derived from
GitHub's `author_association` field). Data-driven
feedback from any reviewer role produces an ACCEPT
recommendation. Subjective feedback is always
AUTHOR-DECIDES regardless of role. Bot and external
reviewer findings are cross-referenced against
convention packs before recommending acceptance.

**Schema relationship**: The Divisor's `review-verdict`
schema captures outbound review findings (what
reviewers flag). The `feedback-triage` schema captures
the inbound response (how the author triaged each
finding). Together they form the complete review
feedback loop: `review-verdict` → reviewer comments →
`feedback-triage`.

The command is re-entrant -- run it again after
reviewers respond to your ASK items or post new
feedback. Local cache under `.uf/feedback/` accelerates
subsequent rounds by skipping unchanged threads.


## Quality Analysis (Gaze)

Gaze provides automated quality analysis for Go projects.

### What It Measures

- **CRAP scores** -- Change Risk Anti-Patterns. Combines
  cyclomatic complexity with test coverage. Score > 30
  indicates high-risk code.
- **Coverage metrics** -- Line and branch coverage with
  ratchets (coverage can only go up, never down).
- **Side effect classification** -- Categorizes functions
  by their I/O behavior (pure, reads, writes, network)
  to guide testing strategy.

### Commands

| Command | What It Does |
|---------|--------------|
| `/gaze` | Run quality analysis -- produces scores, metrics, and health assessment |
| `/gaze-fix` | Generate tests targeting the weakest spots identified by analysis |

### Integration with Review

Gaze quality data feeds into the review council's
context. When Divisor agents review code, they can
reference CRAP scores and coverage metrics to ground
their findings in evidence rather than opinion.


## Parallel Execution (Forge / Replicator)

For large tasks, the Forge subsystem decomposes work and
runs multiple agents in parallel.

### How It Works

```
  /forge
    |
    v
  Coordinator analyzes the task
    |
    v
  Decomposes into subtasks
  (file-based, feature-based, or risk-based strategy)
    |
    +-----+-----+-----+
    |     |     |     |
    v     v     v     v
  Worker Worker Worker Worker
  (agent) (agent) (agent) (agent)
    |     |     |     |
    v     v     v     v
  Results merge back to coordinator
```

### File Reservation

When multiple workers edit files in parallel, the Comms
subsystem prevents conflicts through file reservation:

1. Before editing a file, a worker calls
   `reserve(paths)`.
2. If another worker already reserved that file, the
   request is denied.
3. After completing edits, the worker releases the
   reservation.

### Worktree Isolation

Workers can operate in separate git worktrees for full
isolation. Each worker gets its own checkout of the
repository, and commits are cherry-picked back to the
main branch when complete.

### Commands

| Command | What It Does |
|---------|--------------|
| `/forge` | Decompose a task and spawn parallel workers |
| `/forge-status` | Check progress of running workers |


## Convention Pack System

Convention packs define coding and writing standards that
agents follow when implementing code and reviewing
changes.

### Pack Types

**Tool-owned** (updated by `uf init`):
- `default.md` -- Language-agnostic coding standards
- `go.md` -- Go-specific conventions
- `typescript.md` -- TypeScript-specific conventions
- `content.md` -- Writing and documentation standards
- `severity.md` -- Finding severity definitions

**User-owned** (never overwritten):
- `default-custom.md` -- Your project-specific rules
- `go-custom.md` -- Your Go extensions
- `typescript-custom.md` -- Your TypeScript extensions
- `content-custom.md` -- Your content extensions

### Who Consumes Them

- **Cobalt-Crush** reads packs before writing code to
  follow the correct patterns and conventions.
- **All Divisor agents** read packs before reviewing to
  know what rules to enforce.

### Language Selection

The language pack is selected automatically by detecting
project files (`go.mod` for Go, `package.json` for
TypeScript, etc.). Override with `uf init --lang go` or
`uf init --lang typescript`.


## Artifact Flow

Heroes communicate through well-defined artifacts --
files with metadata that any other hero can interpret
without direct coordination.

### Well-Known Paths

```
.uf/
  artifacts/
    quality-report/    Gaze analysis output
    review-verdict/    Divisor review decisions
    feedback-triage/   Author response to review feedback
    acceptance-decision/  Muti-Mind accept/reject
  muti-mind/
    backlog.yaml       Product backlog
    stories/           Generated user stories
  mx-f/
    metrics/           Sprint and velocity data
    coaching/          Coaching session records
```

### Artifact Envelope Format

Every artifact includes metadata so consumers can
interpret it without consulting the producer:

```json
{
  "hero": "gaze",
  "version": "1.2.0",
  "timestamp": "2026-05-06T10:30:00Z",
  "type": "quality-report",
  "payload": { ... }
}
```

### Constitution Principle I

Artifact-based communication is mandated by the
constitution's Autonomous Collaboration principle.
Heroes must be able to complete their primary function
without synchronous interaction. Artifacts are
asynchronous, auditable, and resilient to individual
hero unavailability.


## Configuration

Unbound Force uses a layered configuration system.
Values are resolved from most specific to least specific:

```
CLI flags
   |
   v
Environment variables
   |
   v
.uf/config.yaml (repo-level)
   |
   v
~/.config/uf/config.yaml (user-level)
   |
   v
Compiled defaults
```

### Config Sections

The `.uf/config.yaml` file has 7 sections:

| Section | Controls |
|---------|----------|
| `scaffold` | Language selection, scaffold behavior |
| `doctor` | Which checks to skip, tool severities |
| `setup` | Package manager preference, tool install methods |
| `sandbox` | Backend (podman/devpod), image, resources, mode |
| `gateway` | Port, provider override |
| `embedding` | Model name, dimensions (used by Dewey and Replicator) |
| `workflow` | Workflow stage definitions and transitions |

### Getting Started with Config

```bash
# Create a commented config file with all defaults
uf config init

# View the effective config after all layers merge
uf config show

# Validate your config file
uf config validate
```

See [docs/configuration.md](configuration.md) for full
reference with all fields, defaults, and common
scenarios.
