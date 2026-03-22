# Research: Dewey Architecture

**Date**: 2026-03-22
**Branch**: `014-dewey-architecture`

## R1: Fork vs Clean-Room Reimplementation

**Decision**: Hard fork of graphthulhu, import full git
history into a new `unbound-force/dewey` repo (not a
GitHub fork).

**Rationale**: graphthulhu's codebase has strong
architectural fundamentals (backend interface with
marker interfaces for capability detection, clean MCP
tool registration, well-structured Markdown parser) and
~3,500 lines of tests. A clean-room rewrite would take
4-6 weeks to reach feature parity and would re-solve
problems graphthulhu already handles (Markdown parsing
edge cases, UUID stability, file watcher race
conditions, atomic writes). The fork gets a working MCP
server on day one.

Creating a new repo (not a GitHub fork) avoids the
"forked from" label and GitHub fork limitations
(can't make private, PRs default to upstream). Full
git history is preserved by pushing graphthulhu's
commits into the new repo.

**Alternatives considered**:
- GitHub fork (rename): Rejected because GitHub forks
  have restrictions and the "forked from" label is
  confusing for an intentionally diverging project.
- Clean-room rewrite: Rejected because graphthulhu's
  architecture is sound and 4-6 weeks of feature-parity
  work is unnecessary.

## R2: Persistent Storage Backend

**Decision**: SQLite for both knowledge graph index and
vector embeddings.

**Rationale**: SQLite is the natural choice for a
per-repo, per-developer tool:
- Zero-configuration (no server process)
- Single-file database (easy to inspect, backup, delete)
- Excellent Go support via `modernc.org/sqlite` (pure Go,
  no CGO required)
- Vector extension options (`sqlite-vec`) for native
  vector operations
- ACID transactions for index integrity
- Concurrent read support (multiple queries during a
  session)

The knowledge graph index (pages, blocks, links,
properties) and vector embeddings can share a single
database file or use separate files. A single file
(`.dewey/index.db`) is simpler.

**Alternatives considered**:
- Flat JSON files: Rejected because query performance
  degrades with document count and concurrent access is
  unsafe.
- BoltDB/bbolt: Rejected because it lacks built-in
  vector support and SQL query flexibility.
- Dedicated vector database (Qdrant, Milvus): Rejected
  because it requires a separate server process,
  violating the zero-configuration design goal.

## R3: Embedding Model Selection

**Decision**: IBM Granite Embedding (`granite-embedding:30m`)
as the default, configurable to any Ollama-compatible
model.

**Rationale**: Granite meets three critical requirements:
1. **Enterprise licensing provenance**: Apache 2.0
   license, training data is "carefully curated,
   permissibly licensed public datasets" with full
   transparency. IBM Research backing.
2. **Local execution**: 63 MB model runs locally via
   Ollama. No data leaves the developer's machine.
3. **Practical performance**: 30M parameters produce
   embeddings at ~10-50ms per chunk on Apple Silicon.
   For 700 chunks (typical repo), total embedding time
   is 7-35 seconds.

The 30M English model is recommended for code repos
and English documentation. The 278M multilingual model
is available for repos with non-English content. Teams
can configure any Ollama-compatible model.

**Alternatives considered**:
- mxbai-embed-large (mixedbread.ai): Rejected due to
  limited training data disclosure and smaller company
  backing. 670 MB model vs 63 MB.
- OpenAI text-embedding-3: Rejected because it requires
  API keys and sends data to a cloud service, violating
  the local-only constraint.
- No default model: Rejected because a sensible default
  reduces setup friction. Teams can override.

## R4: Content Source Interface Design

**Decision**: Pluggable interface with four methods:
List, Fetch, Diff, Meta.

**Rationale**: Each content source (disk, GitHub, web)
has different access patterns, authentication
requirements, and update semantics. A common interface
abstracts these differences while allowing each source
to optimize independently.

- `List()` returns all available documents from the
  source.
- `Fetch(id)` retrieves a specific document's content.
- `Diff()` returns what changed since the last index
  (for incremental updates).
- `Meta()` returns source metadata (name, type,
  freshness).

This design allows adding new source types (Confluence,
Notion, S3) by implementing the interface, without
modifying core indexing logic.

**Alternatives considered**:
- Single `Sync()` method: Rejected because it doesn't
  support incremental updates or selective fetching.
- Event-based (pub/sub): Rejected because most sources
  are poll-based (GitHub API, web crawl). Only the disk
  source can push events (fsnotify).

## R5: Chunking Strategy

**Decision**: Heading-based chunking aligned with
graphthulhu's block model.

**Rationale**: graphthulhu already parses Markdown into
a hierarchical block model (headings, paragraphs, code
blocks, lists). Chunks aligned with this hierarchy have
two advantages:
1. Each chunk has natural semantic boundaries (a heading
   and its content form a complete thought).
2. Search results can link to specific blocks within a
   document, not just the document as a whole.

Chunks are created at heading boundaries (H1, H2, H3).
If a heading's content exceeds the embedding model's
context window (512 tokens), it is split into
overlapping sub-chunks with 50-token overlap to
preserve context.

**Alternatives considered**:
- Fixed-size overlapping windows: Rejected because
  they split content at arbitrary positions, breaking
  semantic coherence.
- Sentence-level chunking: Rejected because individual
  sentences lack sufficient context for meaningful
  embeddings. Heading-level chunks are richer.

## R6: graphthulhu Code Quality Assessment

**Decision**: Fork and incrementally improve. Do not
rewrite.

**Rationale**: Based on a thorough code review of
graphthulhu:
- **Architecture**: Strong. Backend interface with
  marker interfaces for capability detection is exactly
  what Dewey needs for extensibility.
- **Code quality**: Solid. ~6,700 lines of production
  code, ~3,500 lines of tests (52% ratio). Clean Go
  idioms, proper error handling, mutex discipline.
- **Dependencies**: Lean. Only 1 direct dependency
  (MCP Go SDK).

Known issues to address during the fork:
- `vault/vault.go` is 1,409 lines (should be split)
- `tools/` package has minimal test coverage
- Some DataScript query interpolation uses `fmt.Sprintf`
  (injection risk for non-Obsidian backends)

These are manageable improvements, not architectural
problems. The fork cleanup is Phase 0.2 in the
orchestration plan.

## R7: MCP Tool Naming Convention

**Decision**: All Dewey MCP tools use the `dewey_`
prefix. Structured query tools preserve graphthulhu's
tool names with the prefix change. New semantic tools
use descriptive names.

**Rationale**: The MCP protocol namespaces tools by
server name. Agent prompts reference tools by their
full name (e.g., `dewey_search`, `dewey_semantic_search`).
Using a consistent prefix makes tool discovery
predictable and avoids collisions with other MCP
servers.

Existing graphthulhu tools map as:
- `knowledge-graph_search` → `dewey_search`
- `knowledge-graph_find_by_tag` → `dewey_find_by_tag`
- `knowledge-graph_query_properties` → `dewey_query_properties`
- `knowledge-graph_get_page` → `dewey_get_page`
- `knowledge-graph_traverse` → `dewey_traverse`
- `knowledge-graph_find_connections` → `dewey_find_connections`

New semantic tools:
- `dewey_semantic_search` -- natural language query
- `dewey_similar` -- find similar documents
- `dewey_semantic_search_filtered` -- semantic + structural filter

**Alternatives considered**:
- Keep `knowledge-graph_` prefix: Rejected because
  Dewey is not just a knowledge graph -- it includes
  semantic search and external sources. The name would
  be misleading.
- No prefix (use MCP server name): The MCP protocol
  already namespaces by server. However, explicit
  prefixes in agent prompts improve discoverability.

## R8: Dewey Workspace Layout

**Decision**: `.dewey/` directory in the project root.

**Rationale**: Consistent with other tool workspace
patterns in the ecosystem (`.opencode/`, `.specify/`,
`.muti-mind/`, `.mx-f/`, `.unbound-force/`). The
directory contains:

```
.dewey/
├── config.yaml          # Source configuration
├── index.db             # Persistent SQLite index
├── meta.json            # Index metadata (timestamps, checksums)
└── cache/               # Crawl cache (fetched HTML/Markdown)
```

The `.dewey/` directory SHOULD be added to `.gitignore`
(it contains machine-local data: index, cache,
potentially sensitive GitHub issue content). The
`config.yaml` MAY be committed if the team wants shared
source configuration.

**Alternatives considered**:
- `~/.dewey/` (user-level): Rejected because per-repo
  configuration is more flexible and aligns with the
  per-repo MCP server model.
- `.graphthulhu/`: Rejected because Dewey is a separate
  tool, not a graphthulhu extension.

## R9: Graceful Degradation Pattern

**Decision**: Three tiers of degradation.

**Rationale**: Aligns with Constitution Principle II
(Composability First). Each tier provides a functional
experience with progressively richer context:

**Tier 1** (Dewey unavailable): Agents use direct file
reads, CLI queries (`uf doctor`, `gh api`), and
convention packs. Core function works. Context quality
is limited to the current repository and what the LLM's
training data knows.

**Tier 2** (Dewey available, no embedding model): Agents
use structured queries (search, traverse, tags,
properties). This is equivalent to graphthulhu's
functionality with persistence. Semantic search is
unavailable. The `dewey status` command reports the
missing model.

**Tier 3** (Dewey available, embedding model installed):
Full capability. Agents use both structured and semantic
queries across local and external sources.

Agent files include instructions for all three tiers:
```
IF dewey_semantic_search available:
  Use semantic + structured queries (Tier 3)
ELSE IF dewey_search available:
  Use structured queries only (Tier 2)
ELSE:
  Use direct file reads + CLI queries (Tier 1)
```

**Alternatives considered**:
- Binary degradation (Dewey or nothing): Rejected
  because it doesn't account for the common case where
  Dewey is installed but the embedding model isn't.
- Automatic model download: Rejected because it
  introduces network dependency and potential licensing
  concerns without user consent.
