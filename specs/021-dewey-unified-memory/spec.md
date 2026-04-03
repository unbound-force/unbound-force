---
spec_id: "021"
title: "Dewey Unified Memory"
status: draft
created: 2026-03-30
branch: 021-dewey-unified-memory
phase: 2
depends_on:
  - "[[specs/014-dewey-architecture/spec]]"
  - "[[specs/015-dewey-integration/spec]]"
  - "[[specs/018-unleash-command/spec]]"
  - "[[specs/019-divisor-council-refinement/spec]]"
  - "[[specs/020-dewey-knowledge-retrieval/spec]]"
supersedes:
  - "Spec 020 data-model.md 'Relationship to Hivemind'
    section (Dewey now replaces Hivemind, not
    complements it)"
---

# Feature Specification: Dewey Unified Memory

**Feature Branch**: `021-dewey-unified-memory`
**Created**: 2026-03-30
**Status**: Draft
**Input**: Dewey as unified embedding and semantic memory
layer: manage Ollama lifecycle, store learnings, replace
Hivemind, hard fork Swarm plugin.

## User Scenarios & Testing *(mandatory)*

### User Story 1 -- Dewey Manages Ollama (Priority: P1)

A developer starts an OpenCode session. The MCP config
launches `dewey serve`, which checks if Ollama is running
at the expected endpoint. If Ollama is not running but
is installed, Dewey starts it automatically as a managed
subprocess. The developer never needs to run
`ollama serve` manually. If Ollama is already running
(started by the user or another tool), Dewey uses the
existing instance without interfering.

**Why this priority**: This eliminates the root cause of
Issue #73 (silent degradation across Hivemind, Dewey,
and Swarm when Ollama is not serving). Every other user
story depends on Ollama being available for embeddings.

**Independent Test**: Stop Ollama if running. Start
`dewey serve`. Verify Ollama is now running at the
expected endpoint. Verify `dewey_semantic_search`
returns embedding-based results.

**Acceptance Scenarios**:

1. **Given** Ollama is installed but not running,
   **When** `dewey serve` starts,
   **Then** Dewey starts Ollama as a managed subprocess
   and waits for it to be ready before accepting MCP
   tool calls.

2. **Given** Ollama is already running (user-started),
   **When** `dewey serve` starts,
   **Then** Dewey uses the existing Ollama instance
   without starting a new one or modifying its config.

3. **Given** Ollama is not installed (not in PATH),
   **When** `dewey serve` starts,
   **Then** Dewey operates in keyword-only mode
   (no embeddings) and logs an informational message
   about the missing Ollama binary.

4. **Given** `dewey serve` started Ollama as a managed
   subprocess,
   **When** `dewey serve` exits,
   **Then** Ollama is left running (other consumers may
   need it).

---

### User Story 2 -- Dewey Stores Learnings (Priority: P1)

An AI agent stores a learning from a session (e.g., a
retrospective insight, a gotcha about a specific file)
using a Dewey MCP tool. The learning is persisted in
Dewey's index alongside specs, code documentation, and
web documentation. Future sessions can discover the
learning through the same `dewey_semantic_search` that
finds specs and code -- no separate memory query needed.

**Why this priority**: This is the core "unified memory"
value proposition. Without learning storage, Dewey is
a read-only index. With it, Dewey becomes the swarm's
persistent memory that grows smarter with every session.

**Independent Test**: Store a learning via the new
Dewey MCP tool. Run `dewey_semantic_search` with a
query related to the learning. Verify the learning
appears in results alongside other indexed content.

**Acceptance Scenarios**:

1. **Given** a Dewey MCP tool for storing learnings
   exists,
   **When** an agent stores a learning with text, tags,
   and metadata,
   **Then** the learning is persisted in Dewey's index
   and is immediately searchable via
   `dewey_semantic_search`.

2. **Given** a learning was stored in a previous
   session,
   **When** a new session queries Dewey for related
   content,
   **Then** the learning appears in search results with
   provenance metadata indicating it is a learning
   (not a spec or code document).

3. **Given** a learning about `scaffold.go` was stored,
   **When** an agent searches for "scaffold patterns",
   **Then** the learning appears alongside the scaffold
   spec, scaffold source code documentation, and any
   related web documentation -- all ranked by relevance.

4. **Given** learnings are stored with tags (branch
   name, date, category),
   **When** an agent uses `dewey_semantic_search_filtered`
   with a tag filter,
   **Then** only learnings matching the filter are
   returned.

---

### User Story 3 -- /unleash Uses Dewey for Memory (Priority: P1)

The `/unleash` command's retrospective step stores
learnings via Dewey instead of Hivemind. The Divisor
agents' Prior Learnings step searches Dewey for
learnings instead of Hivemind. This is a seamless
migration -- the workflow is identical, only the
storage backend changes.

**Why this priority**: `/unleash` and the Divisor
agents are the primary producers and consumers of
learnings. Migrating them to Dewey is the essential
integration that makes unified memory real.

**Independent Test**: Run `/unleash` on a spec branch.
Verify the retrospective step calls the Dewey learning
tool (not `hivemind_store`). Start a new session, run
`/review-council`, verify the Prior Learnings step
queries Dewey (not `hivemind_find`).

**Acceptance Scenarios**:

1. **Given** `/unleash` reaches the retrospective step,
   **When** it stores session learnings,
   **Then** it uses the Dewey learning storage tool
   (not `hivemind_store`).

2. **Given** a Divisor agent begins its review,
   **When** the Prior Learnings step fires,
   **Then** it queries Dewey via `dewey_semantic_search`
   (not `hivemind_find`) for learnings about the files
   under review.

3. **Given** Dewey is unavailable,
   **When** `/unleash` or a Divisor agent attempts
   learning storage or retrieval,
   **Then** the step is skipped with an informational
   note (graceful degradation preserved).

---

### User Story 4 -- Fork Swarm Plugin (Priority: P2)

The Swarm plugin is forked from the upstream repository
into the `unbound-force` GitHub organization, giving
the project full control over the plugin's behavior.
The fork enables modifications to Hivemind's memory
layer -- either deprecating it in favor of Dewey or
configuring it to proxy through Dewey.

**Why this priority**: The upstream Swarm plugin
controls Hivemind's implementation. Without a fork,
the project cannot replace Hivemind with Dewey or
modify the plugin's behavior. The fork is a
prerequisite for full memory unification.

**Independent Test**: Verify the forked plugin is
installable via `npm install` or `bun install` from
the `unbound-force/swarm` repository. Verify all
existing Swarm tools still function after the fork.

**Acceptance Scenarios**:

1. **Given** the Swarm plugin is forked to
   `unbound-force/swarm`,
   **When** a developer installs it via the package
   manager,
   **Then** all existing Swarm MCP tools function
   identically to the upstream version.

2. **Given** the forked Swarm plugin is installed,
   **When** `uf setup` runs,
   **Then** it installs the forked version from
   `unbound-force/swarm` instead of the upstream.

3. **Given** the fork includes Hivemind modifications,
   **When** Hivemind tools are called,
   **Then** they either proxy through Dewey or are
   deprecated with clear migration messages pointing
   to the equivalent Dewey tools.

---

### User Story 5 -- Doctor and Setup Updates (Priority: P3)

`uf doctor` checks Dewey's health comprehensively,
including whether Dewey can generate embeddings (which
implies Ollama is serving). `uf setup` installs the
forked Swarm plugin. The Ollama serving check in doctor
becomes redundant since Dewey manages Ollama -- doctor
checks Dewey, Dewey checks Ollama.

**Why this priority**: These are operational updates
that complete the integration but are not required for
the core memory functionality to work.

**Independent Test**: Run `uf doctor` and verify the
Dewey health check includes embedding capability. Run
`uf setup` and verify it installs from the forked
Swarm repo.

**Acceptance Scenarios**:

1. **Given** Dewey is running and Ollama is serving,
   **When** `uf doctor` runs,
   **Then** the Dewey health check shows PASS with
   embedding capability confirmed.

2. **Given** Dewey is running but Ollama is NOT serving
   (Dewey failed to start it),
   **When** `uf doctor` runs,
   **Then** the Dewey health check shows WARN with a
   message: "Dewey running but embeddings unavailable.
   Semantic search is keyword-only."

3. **Given** a fresh environment,
   **When** `uf setup` runs,
   **Then** it installs the Swarm plugin from
   `unbound-force/swarm` (the fork).

---

### User Story 6 -- AGENTS.md Documents Unified Memory (Priority: P3)

AGENTS.md is updated to document that Dewey is the
unified knowledge and memory layer for the swarm. The
previous "Dewey complements Hivemind" framing (from
Spec 020) is replaced with "Dewey replaces Hivemind
as the semantic memory layer." All references to
Hivemind as the memory backend are updated.

**Why this priority**: Documentation ensures future
sessions and new contributors understand the
architecture. Without this, agents may still attempt
to use Hivemind tools.

**Independent Test**: Read AGENTS.md and verify it
describes Dewey as the unified memory layer with no
references to Hivemind as a primary memory backend.

**Acceptance Scenarios**:

1. **Given** AGENTS.md is updated,
   **When** an agent reads the Knowledge Retrieval and
   Embedding Model sections,
   **Then** it understands that Dewey is the single
   memory layer for learnings, specs, code, and web
   documentation.

---

### Edge Cases

- What happens when Ollama is running on a non-default
  port? Dewey's Ollama endpoint SHOULD be configurable
  (environment variable or config file). Default is
  `localhost:11434`.
- What happens when two `dewey serve` instances both
  detect Ollama is not running? Both may attempt to
  start Ollama. The second attempt should detect that
  Ollama is now running (started by the first) and
  use it. No coordination needed beyond health checks.
- What happens when Dewey's graph.db is corrupted?
  Learning storage fails. The error is reported to the
  agent. `dewey reindex` rebuilds the index from
  sources (disk, web, GitHub) but learnings are lost
  (they exist only in graph.db). Future consideration:
  backup learnings to a JSONL file.
- What happens when the forked Swarm diverges
  significantly from upstream? The fork is the project's
  responsibility. Upstream changes can be cherry-picked
  as needed but are not automatically merged.
- What happens when an agent tries to use `hivemind_store`
  after the migration? If Hivemind is deprecated in the
  fork, the tool returns an error with a migration
  message: "Use dewey_store_learning instead." If
  Hivemind proxies through Dewey, the call succeeds
  transparently.
- What happens when an MCP tool call requiring
  embeddings arrives while Dewey is still waiting for
  Ollama to start? Dewey serves the request in
  keyword-only mode (degraded but functional). Once
  Ollama is ready, subsequent requests use embeddings.
  No queuing or rejection — the caller gets the best
  available result at the time of the request.
- What happens when Ollama runs out of GPU memory?
  Dewey's embedding calls fail. Dewey falls back to
  keyword-only search. The error is logged but does
  not crash Dewey.

## Requirements *(mandatory)*

### Functional Requirements

**Ollama Lifecycle (Dewey repo):**

- **FR-001**: `dewey serve` MUST check if Ollama is
  running at the configured endpoint on startup.
- **FR-002**: If Ollama is not running and the `ollama`
  binary is in PATH, `dewey serve` MUST start
  `ollama serve` as a managed subprocess.
- **FR-003**: `dewey serve` MUST wait for Ollama to be
  ready (health check passes) before accepting MCP
  tool calls that require embeddings.
- **FR-004**: If Ollama was started by Dewey,
  `dewey serve` MUST leave it running on exit.
- **FR-005**: If Ollama is already running (not started
  by Dewey), `dewey serve` MUST use it without
  modification.
- **FR-006**: If Ollama is not available (not installed
  and not running), `dewey serve` MUST operate in
  keyword-only mode with an informational log message.

**Learning Storage (Dewey repo):**

- **FR-007**: Dewey MUST provide an MCP tool for storing
  learnings with text content, tags, and metadata.
- **FR-008**: Stored learnings MUST be persisted in
  Dewey's index and be immediately searchable via
  `dewey_semantic_search`.
- **FR-009**: Learnings MUST include provenance metadata
  distinguishing them from other document types
  (e.g., `source_type: "learning"`).
- **FR-010**: Learnings MUST support tags for filtering
  via `dewey_semantic_search_filtered`.
- **FR-011**: The learning storage tool MUST generate
  embeddings for the learning text using the same
  embedding model as other indexed content.

**Agent Migration (this repo):**

- **FR-012**: The `/unleash` retrospective step MUST
  use the Dewey learning storage tool instead of
  `hivemind_store`.
- **FR-013**: All 5 Divisor agents' Prior Learnings
  step MUST use `dewey_semantic_search` instead of
  `hivemind_find`.
- **FR-014**: All learning storage and retrieval steps
  MUST gracefully degrade when Dewey is unavailable.
- **FR-015**: AGENTS.md MUST document Dewey as the
  unified memory layer, superseding the "complements
  Hivemind" framing from Spec 020.

**Swarm Fork:**

- **FR-016**: The Swarm plugin MUST be forked to the
  `unbound-force` GitHub organization.
- **FR-017**: The forked plugin MUST function
  identically to the upstream for all non-Hivemind
  tools.
- **FR-018**: Hivemind tools in the fork MUST either
  proxy through Dewey or be deprecated with migration
  messages.
- **FR-019**: `uf setup` MUST install the forked Swarm
  plugin instead of the upstream version.

**Doctor and Setup:**

- **FR-020**: `uf doctor` MUST check Dewey's embedding
  capability as part of the Dewey health check group.
- **FR-021**: The direct Ollama serving check in
  `uf doctor` MAY be removed or demoted since Dewey
  now manages the Ollama lifecycle.

### Key Entities

- **Learning**: A semantic memory entry stored via
  Dewey with text content, tags (branch name, date,
  category), and provenance metadata
  (`source_type: "learning"`). Searchable alongside
  specs, code, and web documentation.
- **Managed Subprocess**: An Ollama server process
  started by `dewey serve` when no existing instance
  is detected. Left running on Dewey exit.
- **Swarm Fork**: A copy of the upstream Swarm plugin
  under `unbound-force/swarm` with full control over
  Hivemind behavior.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can start an OpenCode session
  without manually running `ollama serve` -- Dewey
  handles it transparently.
- **SC-002**: A learning stored via Dewey in session N
  is discoverable via `dewey_semantic_search` in
  session N+1, appearing alongside specs and code
  results.
- **SC-003**: A single `dewey_semantic_search` query
  returns learnings, specs, code documentation, and
  web documentation -- all ranked by relevance. No
  separate memory query is needed.
- **SC-004**: The `/unleash` retrospective successfully
  stores learnings that persist across sessions,
  ending the "Failed to generate embedding" errors
  from Issue #73.
- **SC-005**: All existing Swarm tools (worktrees,
  subtasks, hive, swarmmail) function identically after
  the fork. Zero regressions.
- **SC-006**: `uf doctor` provides a single
  comprehensive health check for the embedding
  infrastructure (Dewey + Ollama), replacing the
  current fragmented checks.

## Assumptions

- Dewey's `graph.db` is the appropriate storage layer
  for learnings. Learnings are treated as documents
  with `source_type: "learning"` in the existing index
  infrastructure.
- The Swarm plugin can be forked without legal or
  licensing issues (the upstream license permits forks).
- The Ollama auto-start subprocess uses the default
  port (11434) and the default model configuration.
  Users with custom Ollama setups (different ports, GPU
  configs) should start Ollama manually before starting
  Dewey -- Dewey will detect and use the existing
  instance.
- No Hivemind data migration is needed. Existing
  Hivemind learnings are abandoned. New learnings go
  to Dewey from day one.
- This spec supersedes the "Dewey complements Hivemind"
  framing in Spec 020's data-model.md. Dewey now
  replaces Hivemind as the semantic memory layer.
- Implementation is phased: Dewey repo changes first
  (Ollama lifecycle + learning storage), then this repo
  (agent migration + doctor/setup), then Swarm fork
  (Hivemind deprecation).
- The Dewey repo changes require separate permission
  to execute.
- **Swarm fork maintenance**: The project owner is
  responsible for monitoring the upstream Swarm plugin
  for security patches and significant bug fixes.
  Recommended upstream sync cadence: monthly review of
  upstream commits, cherry-pick as needed. Sunset
  criteria: if upstream adds native Dewey/pluggable
  memory support, evaluate merging back to upstream
  and deprecating the fork.
