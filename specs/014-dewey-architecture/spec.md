---
spec_id: "014"
title: "Dewey Architecture"
phase: 3
status: draft
depends_on:
  - "[[specs/002-hero-interface-contract/spec]]"
  - "[[specs/010-knowledge-graph-integration/spec]]"
---

# Feature Specification: Dewey Architecture

**Feature Branch**: `014-dewey-architecture`
**Created**: 2026-03-22
**Status**: Draft
**Input**: Dewey design paper
(`../dewey-design-paper.md`) -- a semantic knowledge
layer for AI agent swarms combining knowledge graph
traversal with vector-based semantic search.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Persistent Knowledge Graph (Priority: P1)

A developer starts a coding session in a project with
~200 indexed documents. Instead of rebuilding the
entire knowledge graph from scratch (as graphthulhu
does today), Dewey loads a persisted index from disk
in under 1 second and incrementally updates only the
files that changed since the last session. Structured
queries (search, tag lookup, wikilink traversal,
property queries) work identically to graphthulhu.

**Why this priority**: Persistence is the foundation
for everything else. Without it, vector embeddings and
external content sources would need to be rebuilt
every session, making them impractical. Persistence
also makes Dewey a viable graphthulhu replacement on
day one.

**Independent Test**: Initialize Dewey in a project
with 200 Markdown files. Start the server, run
structured queries, stop the server. Modify 3 files.
Restart the server and verify it loads in under 1
second and the 3 modified files are re-indexed while
the other 197 are loaded from the persistent index.

**Acceptance Scenarios**:

1. **Given** a repository with 200 Markdown files and
   no existing Dewey index, **When** `dewey serve` is
   started for the first time, **Then** a persistent
   index is created on disk and all structured query
   tools (search, find_by_tag, query_properties,
   get_page, traverse, find_connections) return correct
   results.
2. **Given** a repository with an existing Dewey index
   from a previous session, **When** `dewey serve` is
   started, **Then** the index loads from disk in under
   1 second and only files modified since the last
   session are re-indexed.
3. **Given** a running Dewey server monitoring the
   filesystem, **When** a Markdown file is created,
   modified, or deleted, **Then** the index is updated
   incrementally in real time without requiring a
   server restart.
4. **Given** a project using graphthulhu today, **When**
   the developer replaces graphthulhu with Dewey in
   their MCP configuration, **Then** all existing
   structured query tools produce identical results
   with no agent prompt changes required.

---

### User Story 2 - Semantic Search (Priority: P1)

An AI agent persona (e.g., the Product Owner drafting
a specification) queries Dewey with a natural language
concept like "authentication timeout issues." Dewey
returns semantically related documents even when they
use different terminology (e.g., "login session expiry,"
"token refresh failure"). The agent gets richer,
more complete context than keyword search alone could
provide.

**Why this priority**: Semantic search is the core
differentiator between Dewey and graphthulhu. It solves
the terminology gap that makes keyword-only retrieval
insufficient for autonomous agent decision-making.
Without it, agents still need human help to find
relevant context.

**Independent Test**: Index a project containing
documents about "authentication timeout." Query with
"login session expiry" and verify the authentication
timeout document appears in the results. Verify that
traditional keyword search for "login session expiry"
would NOT find this document.

**Acceptance Scenarios**:

1. **Given** a project with indexed documents about
   "authentication timeout handling," **When** an agent
   queries for "login session expiry," **Then** the
   semantic search returns the authentication timeout
   document with a relevance score above the threshold.
2. **Given** a running Dewey server with vector
   embeddings built, **When** an agent calls the
   semantic search tool with a natural language query,
   **Then** results are returned in under 100
   milliseconds per query.
3. **Given** a document in the index, **When** an agent
   calls the "similar documents" tool with that
   document's identifier, **Then** Dewey returns the
   most semantically similar documents ranked by
   relevance.
4. **Given** an agent that needs both conceptual
   relevance and structured filtering, **When** the
   agent calls the filtered semantic search tool with
   a query and a source filter (e.g., source=github,
   repo=gaze), **Then** only documents matching both
   the semantic relevance and the structural filter are
   returned.

---

### User Story 3 - Content Sources (Priority: P2)

A developer configures Dewey to index content from
multiple sources beyond the local repository: GitHub
issues and pull requests from whitelisted repositories
in the organization, and web-crawled documentation for
the project's toolstack (e.g., framework APIs, standard
library docs). When an agent queries Dewey, it
discovers cross-repository context and toolstack
knowledge alongside local documentation.

**Why this priority**: Content sources are what make
Dewey org-aware. Without them, Dewey is a better
graphthulhu (persistent, semantic) but still single-
repo. With them, agents can discover cross-repo
dependencies, find related issues across the org, and
reference current toolstack documentation instead of
relying on potentially stale training data.

**Independent Test**: Configure a GitHub source for 2
whitelisted repos and a web crawl source for one
toolstack docs URL. Run `dewey index`. Verify that
GitHub issues and crawled docs appear in search results
alongside local documents. Verify source provenance
metadata is attached to each result.

**Acceptance Scenarios**:

1. **Given** a Dewey configuration with a GitHub source
   pointing to 2 whitelisted repos, **When**
   `dewey index` is run, **Then** issues, pull request
   descriptions, and READMEs from those repos are
   indexed and searchable (both keyword and semantic).
2. **Given** a Dewey configuration with a web crawl
   source pointing to a toolstack documentation URL,
   **When** `dewey index` is run, **Then** the crawled
   pages are converted to text, indexed, and searchable.
3. **Given** a search result from an external source,
   **When** the agent inspects the result metadata,
   **Then** it includes the source type (github, web,
   disk), the source repository or URL, the fetch
   timestamp, and a link back to the original content.
4. **Given** a previously indexed external source,
   **When** `dewey index` is run again, **Then** only
   content updated since the last fetch is re-indexed
   (incremental update based on timestamps).
5. **Given** a web crawl source, **When** Dewey crawls
   the configured URLs, **Then** it respects
   `robots.txt` rules, imposes a configurable delay
   between requests (default: 1 second), and caches
   crawled content locally to avoid re-fetching
   unchanged pages.

---

### User Story 4 - CLI and Configuration (Priority: P2)

A developer installs Dewey and configures it for their
project using a simple CLI. They initialize a Dewey
workspace with `dewey init`, configure content sources
in a YAML file, build the initial index with
`dewey index`, and start the MCP server with
`dewey serve`. A status command shows what's indexed,
from which sources, and how fresh the data is.

**Why this priority**: The CLI is the developer's
interface to Dewey. Without it, configuration and index
management are manual and error-prone. The CLI makes
Dewey self-service and scriptable.

**Independent Test**: Run `dewey init` in a fresh
directory. Edit the generated configuration file to
add a GitHub source. Run `dewey index`. Run
`dewey status` and verify it shows the indexed
content summary. Run `dewey serve` and verify the
MCP server starts and responds to tool calls.

**Acceptance Scenarios**:

1. **Given** a fresh project directory, **When** the
   developer runs `dewey init`, **Then** a `.dewey/`
   directory is created with a default configuration
   file and an empty index.
2. **Given** a configured Dewey workspace, **When** the
   developer runs `dewey index`, **Then** all configured
   sources are fetched and indexed, with progress
   reported to the terminal.
3. **Given** a Dewey workspace with an existing index,
   **When** the developer runs `dewey status`, **Then**
   the output shows: total documents indexed, breakdown
   by source type, last index timestamp, and index
   freshness for each source.
4. **Given** a configured Dewey workspace, **When** the
   developer runs `dewey serve`, **Then** the MCP
   server starts and accepts tool calls via stdio.
5. **Given** a Dewey configuration file, **When** the
   developer specifies an alternative embedding model,
   **Then** Dewey uses that model for all embedding
   operations instead of the default.

---

### User Story 5 - Hero Integration (Priority: P3)

All five AI agent personas in the Unbound Force swarm
(Muti-Mind, Cobalt-Crush, Gaze, The Divisor, Mx F)
can query Dewey for context relevant to their role.
Each hero benefits differently -- the Product Owner
finds cross-repo backlog patterns, the Developer finds
toolstack API documentation, the Tester finds quality
baselines from other repos, the Reviewer finds
recurring review patterns, and the Manager finds
cross-repo velocity trends. If Dewey is unavailable,
all heroes fall back to direct file reads and CLI
queries with reduced but functional capability.

**Why this priority**: Hero integration is the
ecosystem payoff. Without it, Dewey is a standalone
tool. With it, every hero agent makes better decisions.
This is P3 because it depends on Dewey being fully
functional (US1-US4) and requires updates to all hero
agent files across the org.

**Independent Test**: Configure Dewey with local content
and at least one external source. Query Dewey as each
hero persona and verify role-relevant results. Stop
Dewey and verify each hero still functions (falls back
to file reads).

**Acceptance Scenarios**:

1. **Given** all hero agent files are updated to use
   Dewey tools, **When** Dewey is running and indexed,
   **Then** each hero can call `dewey_search`,
   `dewey_semantic_search`, and `dewey_traverse` tools
   to retrieve context relevant to their role.
2. **Given** Dewey is configured and running, **When**
   the Product Owner agent (Muti-Mind) queries for
   backlog patterns across the org, **Then** it
   receives results from both local backlog items and
   GitHub issues from whitelisted repos.
3. **Given** Dewey is configured and running, **When**
   the Developer agent (Cobalt-Crush) queries for
   toolstack API documentation, **Then** it receives
   results from crawled documentation alongside local
   specs and code comments.
4. **Given** Dewey is unavailable (server not running
   or not configured), **When** any hero agent attempts
   to use Dewey tools, **Then** the agent falls back to
   direct file reads and CLI queries. The agent's core
   function is degraded but not broken.

---

### User Story 6 - Graceful Degradation (Priority: P3)

Every hero agent in the swarm can function without
Dewey. The agent files include fallback instructions
that use direct file reads, CLI queries, and convention
packs when Dewey is unavailable. No hero's core function
depends on Dewey being present. Dewey is an enhancement
that improves context quality, not a hard dependency.

**Why this priority**: The Unbound Force constitution
(Principle II: Composability First) requires that every
hero is independently installable and usable alone.
Making Dewey a hard dependency would violate this
principle. This story ensures constitutional compliance.

**Independent Test**: Remove Dewey from the MCP
configuration. Run each hero agent's primary workflow
(specify, implement, analyze, review, collect). Verify
each hero completes its function successfully, though
with reduced context quality.

**Acceptance Scenarios**:

1. **Given** a project with no Dewey configured in the
   MCP configuration, **When** any hero agent is
   invoked, **Then** it operates using direct file reads
   and convention packs without errors.
2. **Given** a project where Dewey was previously
   configured but the server is not running, **When**
   an agent attempts to call a Dewey tool, **Then** the
   tool call fails gracefully and the agent falls back
   to alternative context sources.
3. **Given** the Muti-Mind agent (Product Owner),
   **When** it runs without Dewey, **Then** it still
   generates specifications using local backlog items
   and convention packs, but without cross-repo context
   or toolstack documentation.

---

### Edge Cases

- What happens when the embedding model is not
  installed? Dewey starts in "graph-only" mode --
  structured queries work, but semantic search is
  unavailable. The status command reports the missing
  model and suggests installation.
- What happens when a GitHub API source returns
  authentication errors? The source is marked as failed
  in the index status. Other sources continue to work.
  The status command reports the failure with the error
  message and suggests checking credentials.
- What happens when a web crawl source encounters a
  `robots.txt` disallow? The disallowed paths are
  skipped. The status command reports which paths were
  skipped and why.
- What happens when the persistent index file is
  corrupted? Dewey detects the corruption at startup,
  logs a warning, and rebuilds the index from scratch
  (falling back to graphthulhu's cold-start behavior).
- What happens when two Dewey instances index the same
  repository concurrently? File locking on the index
  database prevents concurrent writes. The second
  instance waits or fails gracefully with a clear error
  message.
- What happens when the index grows very large (10,000+
  documents including external sources)? Performance
  degrades gracefully. Queries remain under 500ms for
  the structured graph and under 200ms for semantic
  search. The status command shows index size and
  resource usage.
- What happens when a content source is removed from
  the configuration? The next `dewey index` run removes
  the orphaned documents from the index. Queries no
  longer return results from the removed source.

## Requirements *(mandatory)*

### Functional Requirements

**Knowledge Graph (graphthulhu superset):**

- **FR-001**: Dewey MUST provide all MCP tools that
  graphthulhu currently exposes (search, find_by_tag,
  query_properties, get_page, traverse,
  find_connections) with identical behavior.
- **FR-002**: Dewey MUST persist the knowledge graph
  index to disk so it survives across server restarts
  and coding sessions.
- **FR-003**: Dewey MUST support incremental index
  updates -- only files modified since the last session
  are re-indexed at startup.
- **FR-004**: Dewey MUST monitor the filesystem for
  changes and update the index in real time during a
  session.

**Semantic Search:**

- **FR-005**: Dewey MUST provide a semantic search tool
  that finds documents by conceptual similarity, not
  just keyword matching.
- **FR-006**: Dewey MUST provide a "similar documents"
  tool that, given a document identifier, returns the
  most semantically similar documents ranked by
  relevance.
- **FR-007**: Dewey MUST provide a filtered semantic
  search tool that combines conceptual similarity with
  structural filters (source type, repository, property
  values).
- **FR-008**: Dewey MUST generate and persist vector
  embeddings for all indexed content using a locally-run
  embedding model via a local model runtime.
- **FR-009**: The embedding model MUST be configurable
  so teams can choose models that meet their licensing
  and provenance requirements.
- **FR-010**: The default embedding model MUST use
  permissibly licensed training data with full
  provenance transparency.

**Content Sources:**

- **FR-011**: Dewey MUST support a pluggable content
  source architecture where each source implements a
  common interface (list, fetch, diff, metadata).
- **FR-012**: Dewey MUST include a local disk source
  that indexes all Markdown files in the repository,
  including hidden directories.
- **FR-013**: Dewey MUST include a GitHub API source
  that fetches issues, pull request descriptions, and
  READMEs from whitelisted repositories.
- **FR-014**: Dewey MUST include a web crawl source
  that fetches, converts to text, and indexes
  documentation from configured URLs.
- **FR-015**: The web crawl source MUST respect
  `robots.txt` rules and impose a configurable delay
  between requests.
- **FR-016**: Each content source MUST support
  incremental updates based on timestamps or change
  detection.
- **FR-017**: Each search result MUST include source
  provenance metadata: source type, source repository
  or URL, fetch timestamp, and a link to the original
  content.

**CLI:**

- **FR-018**: Dewey MUST provide a `dewey init` command
  that creates a workspace directory with a default
  configuration file.
- **FR-019**: Dewey MUST provide a `dewey index` command
  that fetches and indexes content from all configured
  sources.
- **FR-020**: Dewey MUST provide a `dewey status` command
  that shows index health: total documents, breakdown by
  source, last index time, and freshness per source.
- **FR-021**: Dewey MUST provide a `dewey serve` command
  that starts the MCP server and accepts tool calls via
  stdio.
- **FR-022**: Content source configuration MUST be
  defined in a YAML file within the Dewey workspace
  directory.

**Integration:**

- **FR-023**: Dewey MUST be usable as a drop-in
  replacement for graphthulhu in MCP configurations --
  changing the server name and command is sufficient.
- **FR-024**: All hero agent personas SHOULD use Dewey
  tools when available but MUST fall back to direct file
  reads and CLI queries when Dewey is unavailable.
- **FR-025**: Dewey MUST NOT be a hard dependency for
  any hero agent's core function.

### Key Entities

- **Document**: A unit of indexed content (a Markdown
  file, a GitHub issue, a crawled web page). Has a
  unique identifier, source provenance, raw content,
  and an optional vector embedding.
- **Source**: A pluggable content provider that
  implements the list/fetch/diff/metadata interface.
  Types: disk, github, web. Each source is
  independently configurable and refreshable.
- **Index**: The persistent store containing the
  knowledge graph (pages, blocks, links, properties)
  and vector embeddings. Survives across sessions.
- **Embedding**: A dense vector representation of a
  document chunk, generated by the configured embedding
  model. Used for semantic similarity search.
- **Workspace**: The `.dewey/` directory in a project,
  containing configuration, persistent indexes, and
  crawl cache.

## Assumptions

- graphthulhu's codebase provides a sound architectural
  foundation (backend interface with marker interfaces,
  clean MCP tool registration, well-structured parser).
  Dewey extends this architecture rather than replacing
  it.
- The embedding model runs locally via a model runtime
  on the developer's machine. No data leaves the
  machine. No cloud services, API keys, or network
  dependency for core functionality.
- The default embedding model is chosen for enterprise
  licensing provenance (permissively licensed training
  data, full transparency, Apache 2.0 license). Teams
  can configure alternative models.
- The GitHub API source uses existing CLI authentication
  (the same mechanism other tools in the ecosystem use
  for GitHub operations). No new authentication setup
  is required.
- Dewey is distributed as a standalone binary via
  Homebrew, independently installable from the rest of
  the Unbound Force ecosystem. Any project can use it,
  not just Unbound Force hero repos.
- The web crawl source caches content locally to avoid
  redundant network requests. Toolstack documentation
  is re-crawled on explicit `dewey index` runs, not on
  every server start.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A Dewey server with a persistent index of
  200 documents starts and is ready for queries in
  under 3 seconds (compared to full rebuild every
  session with graphthulhu).
- **SC-002**: Semantic search returns conceptually
  relevant results for queries where keyword search
  returns zero results -- specifically, at least 80%
  of queries using synonyms or related concepts find
  the correct documents.
- **SC-003**: All existing graphthulhu MCP tools produce
  identical results when served by Dewey -- 100%
  behavioral compatibility for structured queries.
- **SC-004**: Content from at least 3 source types
  (disk, GitHub, web) is queryable through a unified
  search interface with source provenance attached to
  every result.
- **SC-005**: All 5 hero agents can use Dewey tools
  when available and fall back gracefully when Dewey
  is unavailable -- zero hard dependencies.
- **SC-006**: `dewey index` completes in under 60
  seconds for an organization with 5 whitelisted repos
  and 3 web crawl sources on a standard developer
  machine.
- **SC-007**: The persistent index adds no more than
  50 MB of disk overhead for a typical repository
  (200 local documents + moderate external sources).
- **SC-008**: `brew install unbound-force/tap/dewey`
  successfully installs Dewey and `dewey serve --help`
  produces correct output.
