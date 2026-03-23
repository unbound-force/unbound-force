---
spec_id: "015"
title: "Dewey Integration"
phase: 3
status: draft
depends_on:
  - "[[specs/014-dewey-architecture/spec]]"
  - "[[specs/013-binary-rename/spec]]"
---

# Feature Specification: Dewey Integration

**Feature Branch**: `015-dewey-integration`
**Created**: 2026-03-22
**Status**: Draft
**Input**: User description: "hero integration,
scaffold updates"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Scaffold Generates Dewey Config (Priority: P1)

A developer runs `uf init` in a new project. The
scaffolded MCP configuration references Dewey instead
of graphthulhu. The generated agent persona files
reference `dewey_*` tool names. The developer can
immediately use Dewey for knowledge retrieval after
installing it, with no manual configuration changes.

**Why this priority**: Scaffold output is how new
projects adopt the toolchain. If `uf init` still
generates graphthulhu references, every new project
starts with an outdated configuration. This is the
highest-impact change because it affects all future
projects.

**Independent Test**: Run `uf init` in a fresh temp
directory. Verify the generated MCP configuration
references `dewey` (not `graphthulhu` or
`knowledge-graph`). Verify all generated agent files
reference `dewey_*` tools. Search for any remaining
`graphthulhu` or `knowledge-graph_` references and
confirm zero matches.

**Acceptance Scenarios**:

1. **Given** a fresh project directory with no existing
   scaffold files, **When** the developer runs
   `uf init`, **Then** the generated MCP configuration
   file references Dewey as the knowledge server with
   the correct command to start it.
2. **Given** a fresh project directory, **When** the
   developer runs `uf init`, **Then** all generated
   agent persona files that reference knowledge
   retrieval tools use `dewey_*` tool names (e.g.,
   `dewey_search`, `dewey_semantic_search`,
   `dewey_traverse`), not `knowledge-graph_*`.
3. **Given** a project that was previously scaffolded
   with graphthulhu references, **When** the developer
   runs `uf init` again, **Then** tool-owned files are
   updated with Dewey references while user-owned
   custom files are preserved.

---

### User Story 2 - Hero Agents Use Dewey Tools (Priority: P1)

All five hero agent personas (Muti-Mind, Cobalt-Crush,
Gaze, The Divisor, Mx F) can use Dewey's knowledge
retrieval tools when available. Each hero benefits from
both structured queries (keyword search, tag lookup,
wikilink traversal) and semantic queries (conceptual
similarity, similar documents). The agent files include
clear instructions for when and how to use each Dewey
tool in the hero's context.

**Why this priority**: Hero agent files are the primary
interface between the AI coding environment and the
knowledge layer. Without updated agent files, heroes
cannot discover or use Dewey's capabilities. This is
tied with US1 because both are needed for a functional
integration.

**Independent Test**: Read each hero's agent file.
Verify it references `dewey_search` and
`dewey_semantic_search` tools. Verify it includes
role-specific guidance (e.g., Product Owner uses
semantic search for backlog patterns, Developer uses
it for toolstack docs). Verify each file includes a
fallback path for when Dewey is unavailable.

**Acceptance Scenarios**:

1. **Given** a hero agent file for each of the 5 heroes,
   **When** Dewey is running and indexed, **Then** each
   hero can call `dewey_search`,
   `dewey_semantic_search`, and `dewey_traverse` tools
   to retrieve context relevant to their role.
2. **Given** a hero agent file for the Product Owner
   (Muti-Mind), **When** the agent file is read, **Then**
   it includes instructions to use semantic search for
   cross-repo backlog patterns and issue discovery.
3. **Given** a hero agent file for the Developer
   (Cobalt-Crush), **When** the agent file is read,
   **Then** it includes instructions to use semantic
   search for toolstack API documentation and
   implementation patterns from other repos.
4. **Given** a hero agent file for the Manager (Mx F),
   **When** the agent file is read, **Then** it includes
   instructions to use semantic search for cross-repo
   velocity trends and historical retrospective data.

---

### User Story 3 - Graceful Degradation in Agent Files (Priority: P1)

Every hero agent file includes fallback instructions
for when Dewey is unavailable. The fallback uses direct
file reads, CLI queries, and convention packs. No hero's
core function breaks when Dewey is absent. The fallback
pattern follows a clear 3-tier structure: full Dewey
(semantic + structured), graph-only (no embedding
model), and no Dewey (file reads only).

**Why this priority**: The constitution (Principle II:
Composability First) requires that every hero works
independently. If agent files are updated to reference
Dewey tools without fallback instructions, heroes will
fail when Dewey is not installed. This must ship with
the tool references, not as a follow-up.

**Independent Test**: Remove Dewey from the MCP
configuration. Invoke each hero agent's primary
workflow. Verify each hero completes its function
without errors. Verify no error messages reference
missing Dewey tools.

**Acceptance Scenarios**:

1. **Given** a project with no Dewey configured, **When**
   any hero agent is invoked, **Then** it operates using
   direct file reads and convention packs without errors.
2. **Given** a hero agent file, **When** the agent file
   is read, **Then** it contains a clear fallback pattern
   distinguishing 3 tiers: full Dewey, graph-only (no
   embedding model), and no Dewey (file reads only).
3. **Given** a project where Dewey is configured but the
   embedding model is not installed, **When** a hero
   attempts a semantic search, **Then** the agent falls
   back to structured keyword search instead.

---

### User Story 4 - Environment Health Checks (Priority: P2)

The `uf doctor` command includes a Dewey health check
that verifies Dewey is installed, the embedding model
is available, and the index is healthy. The `uf setup`
command can install Dewey and pull the embedding model
as part of the environment setup. Both commands provide
clear, actionable feedback when Dewey components are
missing or unhealthy.

**Why this priority**: Doctor and setup are the primary
onboarding tools. Without Dewey health checks,
developers won't know if their knowledge layer is
working. This is P2 because the integration itself
(US1-US3) must be complete before health checks make
sense.

**Independent Test**: Run `uf doctor` in a project
with Dewey configured. Verify it checks for Dewey
binary, embedding model, and index health. Deliberately
remove the embedding model and re-run `uf doctor`.
Verify it reports the missing model with a clear fix
instruction.

**Acceptance Scenarios**:

1. **Given** a project with Dewey installed, indexed,
   and the embedding model available, **When** the
   developer runs `uf doctor`, **Then** the Dewey health
   check group shows all items passing.
2. **Given** a project where Dewey is not installed,
   **When** the developer runs `uf doctor`, **Then** the
   Dewey check reports "not found" with a fix hint
   showing the install command.
3. **Given** a project where Dewey is installed but the
   embedding model is missing, **When** the developer
   runs `uf doctor`, **Then** the check reports the
   missing model with a hint to install it, and notes
   that Dewey works in "graph-only" mode without it.
4. **Given** a developer running `uf setup`, **When**
   setup processes the Dewey installation step, **Then**
   it installs the Dewey binary and pulls the embedding
   model if not already present.

---

### User Story 5 - Swarm Embedding Model Update (Priority: P2)

The Swarm plugin's default embedding model configuration
is updated to use the same enterprise-grade embedding
model that Dewey uses. The `uf doctor` check for the
embedding model references the correct model name. The
`uf setup` command pulls the correct model.

**Why this priority**: Consistency between Swarm's
semantic memory and Dewey's embedding model reduces
developer confusion and ensures both tools use a model
with acceptable licensing provenance. This is P2
because it can be done alongside the doctor/setup
updates.

**Independent Test**: Run `uf doctor` and verify the
Ollama model check references the correct embedding
model name. Run `uf setup` and verify it pulls the
correct model.

**Acceptance Scenarios**:

1. **Given** the meta repo's doctor configuration,
   **When** `uf doctor` checks for the Ollama embedding
   model, **Then** it looks for the enterprise-grade
   model (not the previous default).
2. **Given** the meta repo's setup configuration,
   **When** `uf setup` pulls the embedding model,
   **Then** it pulls the enterprise-grade model.

---

### Edge Cases

- What happens when `uf init` is run in a project that
  already has a graphthulhu MCP configuration? The
  scaffold engine detects the existing config as
  tool-owned and updates it to reference Dewey. The
  developer may need to run `dewey init` and
  `dewey index` to set up the Dewey workspace.
- What happens when a hero agent file references both
  old `knowledge-graph_*` tools and new `dewey_*` tools?
  This should not occur -- the scaffold engine replaces
  tool-owned files entirely. If it does occur in a
  user-owned custom agent file, the old tool names will
  fail gracefully (MCP server not found) and the agent
  will use the fallback path.
- What happens when Dewey is installed but not
  initialized in the project (no `.dewey/` directory)?
  The `uf doctor` check detects this and suggests
  running `dewey init`. The hero agents fall back to
  Tier 1 (file reads) since the MCP server won't start
  without a workspace.
- What happens when the developer has both graphthulhu
  and Dewey configured as separate MCP servers? Both
  servers start. Tools are namespaced by server name, so
  `knowledge-graph_search` and `dewey_search` coexist.
  Agent files reference `dewey_*` tools, so graphthulhu
  is effectively unused. The developer should remove the
  graphthulhu configuration.

## Requirements *(mandatory)*

### Functional Requirements

**Scaffold Updates:**

- **FR-001**: The `uf init` command MUST generate MCP
  configuration that references Dewey as the knowledge
  server, replacing all graphthulhu references.
- **FR-002**: The `uf init` command MUST generate agent
  persona files that reference `dewey_*` tool names,
  not `knowledge-graph_*` tool names.
- **FR-003**: All scaffold assets embedded in the binary
  MUST reference `dewey_*` tools where knowledge
  retrieval is used.
- **FR-004**: Tool-owned scaffold files MUST be updated
  on re-scaffold (`uf init`). User-owned custom files
  MUST NOT be modified.

**Hero Agent Updates:**

- **FR-005**: Each of the 5 hero agent persona files
  (Muti-Mind, Cobalt-Crush, Gaze, The Divisor, Mx F)
  MUST include instructions for using Dewey's semantic
  search tool (`dewey_semantic_search`) alongside the
  existing structured search tools.
- **FR-006**: Each hero agent file MUST include
  role-specific guidance for what to query Dewey for
  (e.g., Product Owner queries for backlog patterns,
  Developer queries for toolstack docs).
- **FR-007**: Each hero agent file MUST include a 3-tier
  fallback pattern: full Dewey (semantic + structured),
  graph-only (structured queries, no embedding model),
  and no Dewey (direct file reads + CLI queries).
- **FR-008**: No hero agent's core function MUST depend
  on Dewey being available. Dewey is an enhancement,
  not a hard dependency.

**Doctor and Setup:**

- **FR-009**: The `uf doctor` command MUST include a
  Dewey health check group that verifies: Dewey binary
  installed, embedding model available, and index
  initialized.
- **FR-010**: The `uf doctor` Dewey checks MUST provide
  clear fix hints when components are missing (e.g.,
  "Install Dewey: brew install unbound-force/tap/dewey").
- **FR-011**: The `uf setup` command MUST install the
  Dewey binary and pull the embedding model if not
  already present.
- **FR-012**: The embedding model referenced in doctor
  and setup MUST be the enterprise-grade model with
  permissible licensing provenance.

**MCP Configuration:**

- **FR-013**: The MCP configuration generated by
  `uf init` MUST specify the Dewey server command and
  arguments needed to start it in the project directory.
- **FR-014**: The MCP configuration MUST NOT include
  graphthulhu as a configured server. Dewey is a
  complete replacement.

### Key Entities

- **MCP Configuration**: The `opencode.json` file (or
  equivalent) that defines which MCP servers to start.
  Contains the server name, command, and arguments for
  Dewey.
- **Agent Persona File**: A Markdown file under
  `.opencode/agents/` that defines a hero's behavior,
  tool usage, and context retrieval strategy. Each
  hero has one persona file.
- **Scaffold Asset**: An embedded file in the CLI binary
  that is deployed to a project directory during
  `uf init`. Tool-owned assets are overwritten on
  re-scaffold; user-owned assets are preserved.
- **Health Check**: A diagnostic check run by
  `uf doctor` that verifies a tool or component is
  installed, configured, and healthy. Reports pass/fail
  with fix hints.

## Assumptions

- Dewey is built and distributed via Homebrew before
  this integration work begins. The `dewey` binary
  exists at `brew install unbound-force/tap/dewey`.
- The Dewey MCP tool names (`dewey_search`,
  `dewey_semantic_search`, `dewey_traverse`, etc.) are
  stable and match the contract defined in Spec 014.
- The binary rename (Spec 013) is complete. All
  references use `uf` or `unbound-force`, not the bare
  `unbound`.
- The existing scaffold ownership model (tool-owned vs
  user-owned files) continues to work correctly. No
  changes to the ownership classification logic are
  needed.
- Cross-repo updates (gaze, website) are separate work
  items. This spec covers the meta repo only.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `uf init` in a fresh directory produces
  zero `graphthulhu` or `knowledge-graph_` references in
  any generated file -- 100% Dewey references.
- **SC-002**: All 5 hero agent persona files include
  `dewey_semantic_search` tool usage instructions and a
  3-tier fallback pattern.
- **SC-003**: `uf doctor` in a project with Dewey fully
  configured shows all Dewey health checks passing. In
  a project without Dewey, the checks report clear fix
  hints.
- **SC-004**: Every hero agent completes its primary
  function without errors when Dewey is not installed --
  zero hard dependencies.
- **SC-005**: All existing tests pass with the updated
  scaffold assets -- zero regressions.
- **SC-006**: The scaffold file count test assertion is
  updated to reflect any new or changed files and passes.
