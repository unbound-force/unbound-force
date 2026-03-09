# Feature Specification: Knowledge Graph Integration

**Feature Branch**: `010-knowledge-graph-integration`
**Created**: 2026-03-08
**Status**: Complete
**Input**: User description: "Integrate graphthulhu (MIT-licensed Go MCP server) with its Obsidian backend as the knowledge management layer for the Unbound Force swarm. graphthulhu reads Markdown files directly from the repo, builds an in-memory knowledge graph with wikilink indexing, full-text search, and graph analysis, and exposes it via MCP so that hero agents (starting with Muti-Mind) can query project knowledge without exceeding context window limits. This addresses the architectural gap where no spec currently defines how heroes retrieve, search, or traverse accumulated project knowledge at scale."

## Clarifications

### Session 2026-03-08

- Q: Which MCP transport should be used (stdio vs HTTP vs both)? → A: Stdio -- service launched as subprocess by OpenCode, lifecycle managed by the agent runtime.
- Q: What is the service lifecycle model (per-invocation, per-session, or persistent daemon)? → A: Per-session -- service starts with OpenCode, persists across agent invocations within the session, stops when OpenCode exits.
- Q: Should the spec capture the rationale for choosing graphthulhu's Obsidian backend over alternatives? → A: Yes -- add a tradeoffs section documenting the decision and alternatives considered.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Agent Knowledge Retrieval via MCP (Priority: P1)

A hero agent (e.g., Muti-Mind, Cobalt-Crush, or any OpenCode
agent operating within a hero repo) needs to answer a question
that requires knowledge spread across multiple project
artifacts -- specs, backlog items, constitution, past decisions,
and related documents. Instead of loading dozens of Markdown
files into its context window, the agent queries a knowledge
graph service via MCP. The service searches, filters, and
returns only the relevant content, keeping the agent's context
focused and within token limits.

**Why this priority**: P1 because this is the core value
proposition. Without MCP-based retrieval, agents must read all
relevant files directly, which becomes untenable as project
knowledge grows. Every other user story depends on the
knowledge graph being queryable by agents.

**Independent Test**: Can be tested by starting the knowledge
graph service pointed at an Unbound Force repo, connecting an
OpenCode agent to it via MCP, and verifying the agent can
search for content across specs, retrieve specific pages, and
get accurate results without reading files directly.

**Acceptance Scenarios**:

1. **Given** a knowledge graph service running against a hero
   repo with at least 10 Markdown files, **When** an agent
   invokes the MCP `search` tool with a query term that
   appears in 3 of those files, **Then** the service returns
   results from all 3 files with surrounding context (parent
   block chain and sibling blocks).
2. **Given** a knowledge graph service, **When** an agent
   invokes the MCP `get_page` tool with a spec name (e.g.,
   "004-muti-mind-architecture/spec"), **Then** the service
   returns the full hierarchical block tree of that document
   with parsed links, tags, and properties.
3. **Given** a knowledge graph service, **When** an agent
   invokes the MCP `list_pages` tool, **Then** the service
   returns a list of all indexed Markdown files with their
   names and metadata.
4. **Given** a repo with 50+ Markdown files totaling over
   100,000 tokens of content, **When** an agent queries for a
   specific topic, **Then** the MCP response contains only the
   relevant blocks (not the entire corpus), keeping the
   returned content under 5,000 tokens for a typical query.

---

### User Story 2 - Knowledge Graph Analysis (Priority: P2)

A hero agent or human operator needs to understand the
structure and health of a project's knowledge base. The
knowledge graph service provides analysis tools that reveal how
documents relate to each other, identify orphaned or
poorly-connected content, and surface topic clusters. This
enables Muti-Mind to identify gaps in specifications, Mx F to
assess documentation health, and human operators to understand
project structure at a glance.

**Why this priority**: P2 because graph analysis provides
strategic value beyond simple retrieval. It enables proactive
quality management and gap detection, but the project can
function without it as long as basic retrieval (US1) works.

**Independent Test**: Can be tested by pointing the knowledge
graph at a repo with known structure (e.g., the unbound-force
meta repo with its 9 spec directories), running analysis tools,
and verifying the results match the known document
relationships and structural properties.

**Acceptance Scenarios**:

1. **Given** a knowledge graph built from a repo where specs
   reference each other via wikilinks, **When** an agent
   invokes the `graph_overview` tool, **Then** the service
   returns aggregate statistics: total pages, total blocks,
   total links, most-connected pages, and namespace breakdown.
2. **Given** a knowledge graph with at least 2 pages linked
   via wikilinks, **When** an agent invokes the
   `find_connections` tool for those pages, **Then** the
   service returns direct links, shortest paths, and shared
   connections between them.
3. **Given** a knowledge graph containing pages with no
   inbound or outbound links, **When** an agent invokes the
   `knowledge_gaps` tool, **Then** the service identifies
   orphan pages, dead-end pages, and weakly-linked areas.
4. **Given** a knowledge graph with interconnected pages,
   **When** an agent invokes the `topic_clusters` tool,
   **Then** the service returns connected components with hub
   identification, revealing natural groupings of related
   content.

---

### User Story 3 - Live Content Synchronization (Priority: P2)

The knowledge graph stays current as project files change. When
a developer or agent modifies a spec, adds a backlog item, or
updates any Markdown file in the repo, the knowledge graph
automatically detects the change and re-indexes the affected
content without requiring a manual restart or reload. This
ensures agents always query up-to-date knowledge.

**Why this priority**: P2 because stale knowledge leads to
incorrect agent decisions. However, a manual reload mechanism
provides an acceptable fallback if automatic detection is not
yet available, making this important but not blocking.

**Independent Test**: Can be tested by starting the knowledge
graph service, modifying a Markdown file in the repo, and
verifying the search index reflects the change within a short
interval without restarting the service.

**Acceptance Scenarios**:

1. **Given** a running knowledge graph service indexing a repo,
   **When** a Markdown file is created in the repo, **Then**
   the new file appears in search results and page listings
   within 5 seconds without restarting the service.
2. **Given** a running knowledge graph service, **When** an
   existing Markdown file is modified, **Then** the updated
   content is reflected in search results within 5 seconds.
3. **Given** a running knowledge graph service, **When** a
   Markdown file is deleted, **Then** it no longer appears in
   search results or page listings within 5 seconds.
4. **Given** a running knowledge graph service, **When** an
   agent invokes the `reload` tool, **Then** the entire index
   is rebuilt from disk and the response confirms the number of
   pages re-indexed.

---

### User Story 4 - Cross-Spec Link Traversal (Priority: P3)

Hero agents need to follow relationships between project
artifacts. When Muti-Mind evaluates a backlog item, it may need
to traverse links from the item to the related spec, from that
spec to the constitution principle it implements, and from there
to other specs governed by the same principle. The knowledge
graph provides link traversal and backlink discovery so agents
can navigate these chains without loading every file.

**Why this priority**: P3 because link traversal requires the
content to use wikilinks, which requires enrichment of existing
content (see Assumptions). The retrieval and analysis
capabilities from US1 and US2 deliver value before link
traversal is fully useful.

**Independent Test**: Can be tested by creating two Markdown
files where file A contains a `[[wikilink]]` to file B,
querying backlinks of file B, and verifying file A appears
as a source.

**Acceptance Scenarios**:

1. **Given** a spec file containing `[[002-hero-interface-contract]]`
   as a wikilink, **When** an agent invokes the `get_links`
   tool on that spec, **Then** the service returns the outbound
   link to `002-hero-interface-contract` along with the block
   containing the reference.
2. **Given** multiple specs referencing a common page via
   wikilinks, **When** an agent invokes `get_links` on the
   common page, **Then** the service returns all inbound
   backlinks with their source pages and containing blocks.
3. **Given** two pages connected through intermediate pages,
   **When** an agent invokes the `traverse` tool with both page
   names, **Then** the service returns the shortest path of
   pages connecting them through the link graph.

---

### User Story 5 - Property-Based Querying (Priority: P3)

Hero agents need to find artifacts by structured metadata. When
Muti-Mind needs all P1 backlog items, or all specs in Phase 1,
or all documents related to a specific hero, it queries the
knowledge graph by property values rather than full-text search.
This requires Markdown files to include YAML frontmatter with
structured metadata that the knowledge graph indexes.

**Why this priority**: P3 because property querying depends on
content enrichment (adding YAML frontmatter to existing files).
Full-text search (US1) provides an adequate fallback for most
queries until frontmatter is widely adopted.

**Independent Test**: Can be tested by creating Markdown files
with YAML frontmatter containing known property values, then
querying for specific property-value combinations and verifying
correct results.

**Acceptance Scenarios**:

1. **Given** spec files with YAML frontmatter containing
   `status: "draft"`, **When** an agent invokes the
   `query_properties` tool with property `status` equal to
   `draft`, **Then** the service returns all spec files with
   that status.
2. **Given** backlog item files with YAML frontmatter
   containing `priority: "P1"`, **When** an agent invokes the
   `query_properties` tool with property `priority` equal to
   `P1`, **Then** the service returns only P1 items.
3. **Given** spec files with YAML frontmatter containing
   `depends_on` as a list of wikilinks, **When** an agent
   invokes `query_properties` with `depends_on` containing a
   specific spec name, **Then** the service returns all specs
   that depend on it.

---

### Edge Cases

- What happens when the knowledge graph service is started
  against a repo with no Markdown files? The service MUST start
  successfully with zero indexed pages and return empty results
  for all queries.
- What happens when a Markdown file contains malformed YAML
  frontmatter? The service MUST skip the frontmatter (treating
  the file as having no properties) and still index the body
  content. A warning SHOULD be logged.
- What happens when a wikilink references a page that does not
  exist? The link MUST still be indexed and reported by link
  tools. The target page SHOULD appear in `knowledge_gaps`
  results as a missing reference.
- What happens when two files have the same name in different
  subdirectories? Each MUST be indexed as a separate page with
  its path-qualified name (e.g., `specs/004/spec` vs
  `specs/009/spec`).
- What happens when the repo contains directories starting
  with `.` (e.g., `.specify/`, `.opencode/`)? Content in these
  directories MUST be indexable. If the chosen tool excludes
  hidden directories by default, a configuration mechanism or
  upstream contribution MUST provide access to this content.
- What happens when the service process crashes while an agent
  is querying? The agent MUST receive an MCP error response and
  SHOULD fall back to direct file reading for the current
  operation.
- What happens when a Markdown file exceeds 100KB? The service
  MUST index it without truncation. Search results from large
  files SHOULD include surrounding context to help the agent
  assess relevance without loading the entire file.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The knowledge graph service MUST index all
  Markdown files in a configured project directory, building
  a queryable in-memory representation of pages, blocks,
  links, tags, and properties.
- **FR-002**: The service MUST expose its capabilities via the
  Model Context Protocol (MCP) using stdio transport. The
  service is launched as a subprocess by the agent runtime
  (OpenCode) at session start, persists across agent
  invocations within the session, and stops when the runtime
  exits. The in-memory index is built once at startup and
  kept current by the file watcher (FR-008) for the duration
  of the session. The service SHOULD also support streamable
  HTTP transport as an alternative for deployments that
  require an independently running service.
- **FR-003**: The service MUST provide full-text search across
  all indexed content with contextual results (parent block
  chain and sibling blocks for each match).
- **FR-004**: The service MUST parse Markdown files into
  hierarchical block trees based on headings (H1 through H6),
  with pre-heading content as root-level blocks.
- **FR-005**: The service MUST extract and index wikilinks
  (`[[page name]]`) from block content, building a bidirectional
  link graph with backlink resolution.
- **FR-006**: The service MUST parse YAML frontmatter
  (delimited by `---` fences) from Markdown files and expose
  properties for structured queries.
- **FR-007**: The service MUST provide graph analysis
  capabilities: overview statistics, connection discovery,
  knowledge gap detection, orphan identification, and topic
  cluster analysis.
- **FR-008**: The service MUST detect file system changes
  (create, modify, delete, rename) and re-index affected files
  automatically without requiring a restart.
- **FR-009**: The service MUST provide link traversal tools:
  outbound links, inbound backlinks, and shortest-path
  traversal between pages through the link graph.
- **FR-010**: The service MUST provide property-based querying
  with operators (equality, contains, greater-than, less-than)
  against YAML frontmatter values.
- **FR-011**: The service MUST be runnable as a standalone
  process with no dependency on a desktop application, GUI, or
  external database. It MUST read Markdown files directly from
  the filesystem.
- **FR-012**: The service MUST be configurable to index
  content in directories that start with `.` (e.g., `.specify/`,
  `.opencode/`), either through configuration options or by
  default inclusion.
- **FR-013**: The service MUST be licensed under a permissive
  open-source license (MIT, Apache 2.0, or BSD) that is
  compatible with Unbound Force's Apache 2.0 license without
  imposing copyleft obligations.
- **FR-014**: The service MUST be invocable as a separate
  process communicating over MCP, maintaining a clear process
  boundary with hero codebases. Hero code MUST NOT import,
  link against, or embed the service's source code.
- **FR-015**: The service MUST support read-only mode where all
  write operations are disabled, suitable for shared or
  production knowledge bases.
- **FR-016**: The service SHOULD extract `#tags` and Logseq-style
  `key:: value` inline properties from block content for
  tag-based filtering and search.
- **FR-017**: The service SHOULD provide a manual reload tool
  that forces a complete re-index of all files from disk.
- **FR-018**: Each hero repo that adopts the knowledge graph
  SHOULD declare the MCP server dependency in its hero manifest
  (per Spec 002) using the `mcp_server` field.
- **FR-019**: The service MUST NOT modify project files during
  read or query operations. Write operations, if supported,
  MUST be gated behind explicit configuration.
- **FR-020**: The knowledge graph integration MUST comply with
  Org Constitution Principle I (Autonomous Collaboration):
  the service is internal tooling for each hero, not a
  cross-hero communication channel. Heroes MUST NOT depend on
  another hero's knowledge graph instance.
- **FR-021**: The knowledge graph integration MUST comply with
  Org Constitution Principle II (Composability First): heroes
  MUST function without the knowledge graph service. The
  service MUST be an optional enhancement that improves
  retrieval performance, not a mandatory dependency.
- **FR-022**: The knowledge graph integration MUST comply with
  Org Constitution Principle III (Observable Quality): the
  service MUST produce machine-parseable MCP responses. Index
  statistics SHOULD be reportable for quality auditing.

### Key Entities

- **Page**: A single Markdown file in the project, identified
  by its path-qualified name (relative path without `.md`
  extension). Attributes: name, file path, YAML frontmatter
  properties, child blocks, outbound links, inbound backlinks,
  tags, journal status (boolean).
- **Block**: A section of a page defined by heading boundaries.
  Attributes: UUID (embedded or deterministic), content (raw
  Markdown text), heading level (0-6), child blocks, parent
  block, parsed links, parsed tags, parsed inline properties.
- **Link**: A wikilink reference (`[[target]]`) within a block.
  Attributes: source page, source block UUID, target page name.
  Links are bidirectional: the target page's backlinks include
  the source.
- **Search Result**: A block matching a query, returned with
  context. Attributes: page name, block UUID, matching content,
  parent block chain (ancestor hierarchy), sibling blocks.
- **Graph Overview**: Aggregate statistics of the knowledge
  graph. Attributes: total pages, total blocks, total links,
  most-connected pages, namespace breakdown, orphan count.
- **Knowledge Gap**: A structural deficiency in the knowledge
  graph. Types: orphan page (no inbound or outbound links),
  dead-end page (links out but nothing links to it), missing
  reference (wikilink target that does not exist as a page).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An agent can retrieve relevant content from a
  repo with 50+ Markdown files in under 2 seconds per MCP
  query, without loading more than 5,000 tokens of content
  per typical search result.
- **SC-002**: Full-text search returns results from all files
  containing the query term with zero false negatives (no
  relevant files missed) for exact-match queries.
- **SC-003**: The knowledge graph correctly identifies all
  wikilink relationships: for every `[[target]]` in the
  indexed content, the target page's backlinks include the
  source page.
- **SC-004**: File system changes (create, modify, delete) are
  reflected in search results within 5 seconds without
  restarting the service.
- **SC-005**: Graph analysis correctly identifies orphan pages
  (pages with zero links) and missing references (wikilink
  targets with no corresponding file) with zero false
  negatives.
- **SC-006**: The service starts and indexes a repo with 100
  Markdown files in under 10 seconds.
- **SC-007**: The service operates as a standalone process
  with no desktop application, GUI, or external database
  dependency. Installation requires only a single binary or
  package install command.
- **SC-008**: All content in directories starting with `.`
  (specifically `.specify/` and `.opencode/`) is indexed and
  searchable when the service is properly configured.

## Assumptions

- Existing spec files do not currently use YAML frontmatter
  or wikilinks. Property-based querying (US5) and link
  traversal (US4) will deliver full value only after content
  enrichment is performed in a follow-up effort. Full-text
  search (US1) provides immediate value without content
  changes.
- The graphthulhu project (MIT-licensed, Go, actively
  maintained) is the intended implementation vehicle. If
  graphthulhu's Obsidian backend does not support indexing
  hidden directories (directories starting with `.`), an
  upstream contribution or fork will be required to address
  FR-012.
- The knowledge graph service runs locally alongside the
  agent runtime (OpenCode). It does not need to be deployed
  as a shared or remote service.
- Write operations through the knowledge graph are not a
  priority for the initial integration. The primary use case
  is read-only retrieval. Write operations (creating pages,
  appending blocks) are a future enhancement.
- The MCP integration pattern follows the same approach used
  by other OpenCode MCP servers: the service is configured
  in the project's OpenCode MCP configuration and its tools
  become available to all agents operating in that project.

## Tradeoffs & Rejected Alternatives

The following alternatives were evaluated before selecting
graphthulhu's Obsidian backend as the implementation vehicle:

- **Logseq desktop app + community MCP server (e.g.,
  ergut/mcp-logseq, graphthulhu Logseq backend)**: Provides
  richer structured data (typed properties, Datalog queries,
  semantic search) but requires the Logseq Electron desktop
  app running as an HTTP API server. Rejected because it
  introduces a heavyweight desktop dependency for what should
  be a headless, CLI-driven agent workflow. The Electron
  process consumes significant resources and cannot run in
  CI/CD or headless server environments.

- **Logseq built-in MCP server (via @logseq/cli)**: Can run
  headless against a SQLite file without the desktop app.
  However, the tool set is incomplete (explicit TODOs for
  editing, property handling, block children), and the
  codebase is AGPL-3.0 licensed. While the AGPL license is
  safe when invoked as a separate process, the MIT license of
  graphthulhu eliminates all license boundary concerns.

- **Custom SQLite index + custom MCP server**: Purpose-built
  for Unbound Force's artifact schemas with no external
  dependency risk. Rejected because it requires building and
  maintaining indexing, search, graph analysis, and MCP
  server infrastructure from scratch. graphthulhu already
  provides 30+ tools covering these capabilities.

- **graphthulhu Obsidian backend (selected)**: Reads Markdown
  files directly from disk with no desktop app, database, or
  external service dependency. MIT license is fully compatible
  with Apache 2.0. Single Go binary, easy to install and
  distribute. Provides 30+ MCP tools including full-text
  search, graph analysis, link traversal, property querying,
  and file watching. The primary limitation -- inability to
  index hidden directories (directories starting with `.`) --
  is addressable via an upstream contribution or fork since
  the project is MIT-licensed and actively maintained.

## Dependencies

### Prerequisites

- **Spec 001** (Org Constitution): The knowledge graph
  integration must comply with the three core principles.
- **Spec 002** (Hero Interface Contract): Heroes adopting the
  knowledge graph declare the MCP server in their manifest.

### Downstream Dependents

- **Spec 004** (Muti-Mind Architecture): Muti-Mind is the
  primary consumer. The knowledge graph enables Muti-Mind to
  query backlogs, specs, and decisions without context window
  exhaustion.
- **Spec 006** (Cobalt-Crush Architecture): Cobalt-Crush
  FR-007 requires maintaining awareness of past review
  feedback. The knowledge graph provides the retrieval
  mechanism.
- **Spec 007** (Mx F Architecture): Mx F can query the
  knowledge graph for documentation health metrics and
  coaching record history.
- **Spec 008** (Swarm Orchestration): The learning loop
  requires accumulated knowledge across sessions. The
  knowledge graph provides persistent, queryable access to
  workflow records and coaching records.

### External Dependencies

- **graphthulhu**: MIT-licensed Go MCP server
  (github.com/skridlevsky/graphthulhu). The Obsidian backend
  is used for direct Markdown file access without a desktop
  application dependency. An upstream contribution may be
  required for hidden directory support (FR-012).
