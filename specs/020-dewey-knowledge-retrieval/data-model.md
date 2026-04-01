# Data Model: Dewey Knowledge Retrieval

**Branch**: `020-dewey-knowledge-retrieval`
**Date**: 2026-04-01

## Overview

This spec introduces no new data entities, schemas, or
artifacts. It adds behavioral instructions (Markdown text)
to existing files. This data model document captures the
two key conceptual models that govern the implementation:
the tool selection matrix and the graceful degradation
pattern.

## Tool Selection Matrix

The core "data model" of this spec is the decision matrix
that maps query intents to Dewey MCP tools. This matrix
is the authoritative reference for all Knowledge Retrieval
sections across agent files, command files, and AGENTS.md.

### Primary Tools (used in AGENTS.md convention)

| Query Intent | Dewey Tool | Input | Output | When to Prefer |
|-------------|-----------|-------|--------|----------------|
| Conceptual understanding | `dewey_semantic_search` | Natural language query | Ranked results with similarity scores | "How does X work?", "Patterns for Y" |
| Keyword lookup | `dewey_search` | Exact text query | Matching blocks with context | Known terms, file names, FR numbers |
| Read specific page | `dewey_get_page` | Page name | Full page with block tree | Known spec or document path |
| Relationship discovery | `dewey_find_connections` | Two page names | Paths, shared connections | "How are X and Y related?" |

### Extended Tools (used in role-specific sections)

| Query Intent | Dewey Tool | Best For |
|-------------|-----------|----------|
| Find similar docs | `dewey_similar` | "Find specs like this one" |
| Tag-based discovery | `dewey_find_by_tag` | "All #decision items" |
| Property queries | `dewey_query_properties` | "All specs with status: draft" |
| Filtered semantic | `dewey_semantic_search_filtered` | Semantic search within source type |
| Graph navigation | `dewey_traverse` | Dependency chain walking |

### Fallback Tools (when Dewey unavailable)

| Dewey Tool | Fallback | Notes |
|-----------|----------|-------|
| `dewey_semantic_search` | Grep + Read | Loses semantic matching; keyword only |
| `dewey_search` | Grep | Equivalent for exact matches |
| `dewey_get_page` | Read | Direct file path required |
| `dewey_find_connections` | Manual traversal | Read specs, follow `depends_on` links |
| `dewey_traverse` | Manual traversal | Read specs, follow `depends_on` links |

## Graceful Degradation Pattern

Every Knowledge Retrieval section follows this 3-tier
pattern. The tiers are ordered by capability (highest
first) and each tier is a complete fallback for the
tier above it.

### Tier 3: Full Dewey (semantic + structured search)

**Availability**: Dewey MCP server running with embedding
model loaded (`granite-embedding:30m`).

**Capabilities**:
- `dewey_semantic_search` — natural language queries
- `dewey_search` — keyword queries
- `dewey_get_page` — specific page reads
- `dewey_find_connections` — relationship discovery
- `dewey_traverse` — graph navigation
- `dewey_similar` — document similarity
- `dewey_find_by_tag` — tag-based discovery
- `dewey_query_properties` — property-based queries
- `dewey_semantic_search_filtered` — filtered semantic

### Tier 2: Graph-only (no embedding model)

**Availability**: Dewey MCP server running but embedding
model not loaded or not configured.

**Capabilities**:
- `dewey_search` — keyword queries (works without embeddings)
- `dewey_get_page` — specific page reads
- `dewey_traverse` — graph navigation
- `dewey_find_connections` — relationship discovery
- `dewey_find_by_tag` — tag-based discovery
- `dewey_query_properties` — property-based queries

**Not available**: `dewey_semantic_search`,
`dewey_semantic_search_filtered`, `dewey_similar`
(all require embedding model)

### Tier 1: No Dewey (direct file access)

**Availability**: Dewey MCP server not running or not
configured.

**Capabilities**:
- Read tool — direct file access (requires known path)
- Grep tool — keyword search across codebase
- Glob tool — file pattern matching

**Not available**: Any semantic or graph-based queries.
Agent must rely on its context window and explicit file
reads.

## Role-Specific Query Patterns

Each hero agent uses Dewey differently based on its role.
These patterns are documented in each agent's Knowledge
Retrieval section.

| Hero | Primary Query Pattern | Example Queries |
|------|----------------------|-----------------|
| Cobalt-Crush | Prior learnings + related specs | "scaffold.go patterns", "FR-001 implementation" |
| Muti-Mind | Backlog patterns + acceptance history | "past acceptance criteria", "backlog priorities" |
| Mx F | Velocity trends + retrospective outcomes | "velocity trends across repos", "coaching patterns" |
| Gaze | Quality baselines + test patterns | "CRAP score patterns", "quality baselines" |
| Divisor (all) | Review patterns + convention violations | "recurring review findings", "security patterns" |

## Speckit Command Query Patterns

Each Speckit command uses Dewey at a specific point in
its workflow.

| Command | Query Point | Query Type | Purpose |
|---------|------------|------------|---------|
| `/speckit.specify` | Before generating spec | `dewey_semantic_search` | Find existing specs with similar topics |
| `/speckit.plan` | During Phase 0 research | `dewey_search` | Find prior research decisions |
| `/speckit.tasks` | Before generating tasks | `dewey_semantic_search_filtered` | Find implementation patterns from completed specs |

## Relationship to Hivemind

Dewey and Hivemind serve complementary purposes. This spec
adds Dewey alongside Hivemind, not replacing it.

| Aspect | Dewey | Hivemind |
|--------|-------|----------|
| Scope | Cross-repo architectural context | Session-specific learnings |
| Data source | Indexed Markdown files, web docs, GitHub | Agent session memories |
| Query type | Semantic + structured search | Semantic search only |
| Used by | All agents (this spec) | Divisor agents (Spec 019) |
| Persistence | Disk-based index | SQLite + JSONL |
