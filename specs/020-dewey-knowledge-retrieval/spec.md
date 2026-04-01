---
spec_id: "020"
title: "Dewey Knowledge Retrieval"
status: draft
created: 2026-03-30
branch: 020-dewey-knowledge-retrieval
phase: 2
depends_on:
  - "[[specs/014-dewey-architecture/spec]]"
  - "[[specs/015-dewey-integration/spec]]"
---

# Feature Specification: Dewey Knowledge Retrieval

**Feature Branch**: `020-dewey-knowledge-retrieval`
**Created**: 2026-03-30
**Status**: Draft
**Input**: Add Dewey knowledge retrieval behavioral
instructions to AGENTS.md and all hero agent files so
AI agents prefer Dewey MCP tools over grep/glob for
cross-repo context, design decisions, and architectural
patterns.

## User Scenarios & Testing *(mandatory)*

### User Story 1 -- AGENTS.md Knowledge Retrieval Convention (Priority: P1)

A developer opens a new OpenCode session in a repo
scaffolded by `uf init`. The agent reads AGENTS.md and
encounters a "Knowledge Retrieval" section that
instructs it to prefer Dewey MCP tools for cross-repo
context, design decisions, and architectural patterns.
When the agent needs to understand how a feature works
across repos, it uses `dewey_semantic_search` instead
of spawning an `explore` agent or running `grep`.

**Why this priority**: AGENTS.md is the primary context
file injected into every OpenCode agent session. Adding
a behavioral instruction here is the highest-leverage
change -- it affects every agent, every command, every
session.

**Independent Test**: Open a new session, ask "how does
the scaffold system work?" Verify the agent uses
`dewey_semantic_search` or `dewey_search` before falling
back to file reads. Check AGENTS.md for the Knowledge
Retrieval section.

**Acceptance Scenarios**:

1. **Given** AGENTS.md contains a "Knowledge Retrieval"
   section,
   **When** an agent needs cross-repo context or
   architectural understanding,
   **Then** it queries Dewey MCP tools first and only
   falls back to grep/glob/read when Dewey is
   unavailable or the query requires exact string
   matching.

2. **Given** AGENTS.md's Knowledge Retrieval section
   defines when to use which Dewey tool,
   **When** an agent reads the section,
   **Then** it has clear guidance: `dewey_semantic_search`
   for conceptual queries, `dewey_search` for keyword
   queries, `dewey_get_page` for reading specific pages,
   `dewey_find_connections` for relationship discovery.

3. **Given** Dewey MCP server is not running,
   **When** an agent attempts a Dewey query,
   **Then** it falls back gracefully to grep/glob/read
   with no error or interruption.

---

### User Story 2 -- Cobalt-Crush Knowledge Step (Priority: P1)

Cobalt-Crush (the developer persona) includes a
"Knowledge Retrieval" step in its workflow that fires
before code exploration. When implementing a feature,
Cobalt-Crush queries Dewey for prior decisions, related
specs, and learned patterns about the files being
modified, producing more context-aware implementations.

**Why this priority**: Cobalt-Crush is the primary
implementation agent. Adding Dewey awareness to its
workflow directly improves code quality by grounding
implementations in project history and conventions.

**Independent Test**: Invoke `/cobalt-crush` on a task
involving `scaffold.go`. Verify the agent queries Dewey
for learnings and prior decisions about scaffold.go
before writing code.

**Acceptance Scenarios**:

1. **Given** Cobalt-Crush is about to implement a task
   involving files that have Dewey-indexed history,
   **When** the Knowledge Retrieval step fires,
   **Then** the agent queries Dewey for: prior learnings
   about the target files, related design decisions from
   specs, and architectural patterns from conventions.

2. **Given** Dewey returns relevant prior learnings
   (e.g., "scaffold.go requires initSubTools nil guard
   for Stdout"),
   **When** Cobalt-Crush implements the task,
   **Then** the implementation accounts for the learned
   pattern without the developer having to remind the
   agent.

---

### User Story 3 -- Speckit Pipeline Dewey Integration (Priority: P2)

The Speckit pipeline commands (`/speckit.specify`,
`/speckit.plan`, `/speckit.tasks`) query Dewey for
existing specs, decisions, and conventions before
generating new artifacts. This prevents duplicate specs,
ensures consistency with prior decisions, and discovers
related work across the organization.

**Why this priority**: The Speckit pipeline generates
foundational artifacts (specs, plans, tasks). Grounding
these in project history reduces rework and inconsistency.

**Independent Test**: Run `/speckit.specify` for a new
feature. Verify the agent queries Dewey for existing
specs with similar topics before creating the new spec.

**Acceptance Scenarios**:

1. **Given** a developer runs `/speckit.specify` for a
   feature related to an existing spec topic,
   **When** the specify command generates the spec,
   **Then** it queries Dewey for similar existing specs
   and references them in the Dependencies or
   Assumptions section.

2. **Given** a developer runs `/speckit.plan` for a
   feature,
   **When** the plan command generates research.md,
   **Then** it queries Dewey for prior research
   decisions in related specs rather than starting
   from scratch.

---

### User Story 4 -- All Hero Agents Knowledge Step (Priority: P3)

All hero agent personas (Muti-Mind, Cobalt-Crush,
Mx F, Divisor agents, Gaze reporter) include a
knowledge retrieval step that queries Dewey before
performing their primary function. This transforms
every hero from a stateless operator into a
context-aware participant that leverages the
organization's accumulated knowledge.

**Why this priority**: While Cobalt-Crush (US2) and
Speckit (US3) are the highest-leverage integration
points, extending to all heroes completes the knowledge
layer vision (per Spec 014/015).

**Independent Test**: Invoke Muti-Mind for a backlog
decision. Verify it queries Dewey for prior acceptance
decisions and related backlog items before rendering
its judgment.

**Acceptance Scenarios**:

1. **Given** any hero agent is invoked,
   **When** it begins its primary workflow,
   **Then** it includes a Dewey knowledge retrieval
   step appropriate to its role (e.g., Muti-Mind
   searches for backlog patterns, Mx F searches for
   velocity trends, Gaze searches for quality history).

2. **Given** Dewey is not available,
   **When** any hero agent begins its workflow,
   **Then** the knowledge retrieval step is skipped
   with an informational note and the agent proceeds
   with its primary function.

---

### User Story 5 -- Unbound Force Heroes Skill Update (Priority: P3)

The `unbound-force-heroes` skill
(`.opencode/skill/unbound-force-heroes/SKILL.md`)
includes Dewey as the knowledge retrieval layer in
the hero lifecycle workflow documentation. The skill
teaches Swarm coordinators to use Dewey for context
before delegating to hero agents.

**Why this priority**: The heroes skill is the routing
layer for swarm coordination. Adding Dewey awareness
here ensures that even swarm-coordinated workflows
benefit from knowledge retrieval.

**Independent Test**: Read the heroes skill and verify
it mentions Dewey knowledge retrieval as a step in the
hero lifecycle.

**Acceptance Scenarios**:

1. **Given** the heroes skill describes the 6-stage
   hero lifecycle,
   **When** a Swarm coordinator reads the skill,
   **Then** the skill instructs the coordinator to
   query Dewey for relevant context before each stage.

---

### Edge Cases

- What happens when Dewey returns irrelevant results
  for a query? The agent uses its judgment to determine
  relevance (same pattern as `/unleash` clarify step).
  Irrelevant results are silently discarded.
- What happens when Dewey is slow (>10s response)? The
  agent proceeds without Dewey results after a
  reasonable timeout. No hard timeout is enforced --
  the agent decides when to move on.
- What happens when Dewey returns too many results? The
  agent focuses on the top 3-5 results by similarity
  score and ignores the rest.
- What happens when the agent already knows the answer
  from its context window (e.g., AGENTS.md was already
  loaded)? The agent SHOULD still query Dewey for
  cross-repo context that isn't in its context window,
  but MAY skip the query if the answer is already
  clearly available.
- What happens when a query returns results from web
  sources (e.g., Swarm docs)? The agent treats web
  source results the same as disk/GitHub results --
  the provenance metadata indicates the source type.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: AGENTS.md MUST include a "Knowledge
  Retrieval" section that instructs agents to prefer
  Dewey MCP tools over grep/glob for cross-repo
  context, design decisions, and architectural patterns.
- **FR-002**: The Knowledge Retrieval section MUST
  specify which Dewey tool to use for which query type:
  `dewey_semantic_search` for conceptual queries,
  `dewey_search` for keyword queries,
  `dewey_get_page` for specific pages,
  `dewey_find_connections` for relationship discovery.
- **FR-003**: The Knowledge Retrieval section MUST
  specify when to fall back to grep/glob/read: when
  Dewey is unavailable, when exact string matching is
  needed, or when a specific file path is already known.
- **FR-004**: The `cobalt-crush-dev.md` agent MUST
  include a "Knowledge Retrieval" step that fires
  before code exploration, querying Dewey for prior
  learnings, related specs, and file-specific context.
- **FR-005**: The Speckit commands (`speckit.specify.md`,
  `speckit.plan.md`, `speckit.tasks.md`) SHOULD include
  Dewey queries for related existing specs and prior
  decisions before generating new artifacts.
- **FR-006**: All hero agent personas (`muti-mind-po.md`,
  `mx-f-coach.md`, `gaze-reporter.md`) SHOULD include
  a role-appropriate Dewey knowledge retrieval step.
- **FR-007**: The `unbound-force-heroes` skill SHOULD
  mention Dewey as the knowledge retrieval layer in the
  hero lifecycle documentation.
- **FR-008**: All Dewey knowledge retrieval steps MUST
  gracefully degrade when Dewey is unavailable (skip
  with informational note, never block the workflow).
- **FR-009**: All changes to agent and command files
  MUST be synced to their scaffold asset copies in
  `internal/scaffold/assets/`.
- **FR-010**: The scaffold file count and expected asset
  paths MUST remain unchanged (no new files added or
  removed).

### Key Entities

- **Knowledge Retrieval Step**: A behavioral instruction
  in an agent file that directs the agent to query Dewey
  before performing its primary function.
- **Dewey Tool Selection**: A decision matrix mapping
  query types to Dewey MCP tools (semantic search for
  concepts, keyword search for terms, get_page for
  known pages, find_connections for relationships).
- **Graceful Degradation**: The pattern of attempting a
  Dewey query, catching unavailability, and proceeding
  without it.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An agent reading AGENTS.md encounters
  clear instructions to prefer Dewey over grep for
  cross-repo context, with specific tool selection
  guidance.
- **SC-002**: Cobalt-Crush queries Dewey for prior
  learnings when implementing tasks on files with
  indexed history, referencing discovered patterns in
  its implementation.
- **SC-003**: Speckit commands discover related existing
  specs via Dewey before creating new artifacts,
  reducing duplicate or conflicting specifications.
- **SC-004**: All Dewey knowledge retrieval steps
  degrade gracefully when Dewey is unavailable -- zero
  workflow interruptions from missing Dewey.
- **SC-005**: All modified agent and command files are
  synced to scaffold assets with no drift.

## Assumptions

- Dewey MCP server is functional and returning results
  (the issues from earlier sessions have been partially
  addressed).
- Web documentation sources (Swarm, OpenCode, Cobra,
  etc.) have been manually added to this repo's
  `.dewey/sources.yaml` and indexed.
- The "prefer Dewey" instruction is a SHOULD (soft
  preference), not a MUST (hard requirement). Agents
  retain discretion to use grep/glob/read when it's
  more appropriate.
- The Divisor agents already have a "Prior Learnings"
  step (from Spec 019) that uses `hivemind_find`. This
  spec adds Dewey search alongside Hivemind, not
  replacing it. Hivemind stores session-specific
  learnings; Dewey provides cross-repo architectural
  context.
- Changes are all Markdown (agent files, command files,
  skill file, AGENTS.md). No Go code changes. Scaffold
  asset syncing is the only "build" step.
