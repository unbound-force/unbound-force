# Contract: Dewey MCP Tools

**Date**: 2026-03-22
**Branch**: `014-dewey-architecture`

## Overview

Dewey exposes its capabilities as MCP tools via stdio.
All tools are prefixed with `dewey_`. Structured query
tools are behavioral supersets of graphthulhu's
equivalent tools.

## Structured Query Tools (graphthulhu superset)

These tools provide identical behavior to graphthulhu's
equivalents, with the `dewey_` prefix replacing the
`knowledge-graph_` prefix.

### dewey_search

Full-text keyword search across all indexed content.

**Parameters**:
- `query` (string, required): Search terms
- `limit` (int, optional): Max results (default: 10)

**Returns**: Array of search results with:
- `document_id`: Unique document identifier
- `title`: Document title
- `snippet`: Relevant text snippet
- `score`: Relevance score
- `source`: Source provenance metadata

### dewey_find_by_tag

Find documents by tag values from YAML frontmatter.

**Parameters**:
- `tag` (string, required): Tag to search for

**Returns**: Array of matching documents with metadata.

### dewey_query_properties

Structured queries against YAML frontmatter properties.

**Parameters**:
- `property` (string, required): Property name
- `value` (string, optional): Expected value
- `operator` (string, optional): Comparison operator

**Returns**: Array of documents matching the property
query.

### dewey_get_page

Retrieve the full block tree of a specific document.

**Parameters**:
- `path` (string, required): Document path or ID

**Returns**: Full document content with block hierarchy.

### dewey_traverse

Follow wikilinks from a document to discover
relationships.

**Parameters**:
- `path` (string, required): Starting document
- `depth` (int, optional): Traversal depth (default: 1)

**Returns**: Connected documents and their link types.

### dewey_find_connections

Discover implicit connections between documents.

**Parameters**:
- `path` (string, required): Document path or ID

**Returns**: Documents connected by shared tags,
properties, or wikilinks.

## Semantic Search Tools (new in Dewey)

These tools require an embedding model to be configured
and installed. If the model is unavailable, these tools
return an error indicating "semantic search unavailable,
install embedding model."

### dewey_semantic_search

Find documents by conceptual similarity using vector
embeddings.

**Parameters**:
- `query` (string, required): Natural language query
- `limit` (int, optional): Max results (default: 10)
- `threshold` (float, optional): Minimum similarity
  score (0.0-1.0, default: 0.5)

**Returns**: Array of semantically similar documents:
- `document_id`: Unique document identifier
- `title`: Document title
- `snippet`: Most relevant chunk text
- `similarity`: Cosine similarity score (0.0-1.0)
- `source`: Source provenance metadata

### dewey_similar

Given a document, find the most semantically similar
documents in the index.

**Parameters**:
- `document_id` (string, required): Source document ID
- `limit` (int, optional): Max results (default: 5)

**Returns**: Array of similar documents ranked by
similarity score.

### dewey_semantic_search_filtered

Semantic search with structural filters. Combines
conceptual similarity with source-type, repository,
or property constraints.

**Parameters**:
- `query` (string, required): Natural language query
- `source_type` (string, optional): Filter by source
  type (`"disk"`, `"github"`, `"web"`)
- `source_name` (string, optional): Filter by source
  name
- `property` (string, optional): Filter by property
  name
- `property_value` (string, optional): Filter by
  property value
- `limit` (int, optional): Max results (default: 10)

**Returns**: Same format as `dewey_semantic_search`,
filtered by the structural constraints.

## Source Provenance Metadata

Every search result (structured or semantic) includes
source provenance:

```json
{
  "source": {
    "type": "github",
    "name": "gaze-issues",
    "repo": "unbound-force/gaze",
    "url": "https://github.com/unbound-force/gaze/issues/142",
    "fetched_at": "2026-03-21T14:30:00Z"
  }
}
```

For local disk sources:

```json
{
  "source": {
    "type": "disk",
    "name": "local",
    "path": "specs/008-swarm-orchestration/spec.md",
    "fetched_at": "2026-03-22T10:00:00Z"
  }
}
```

## CLI Commands

### dewey init

Initialize a Dewey workspace in the current directory.

```bash
dewey init
```

Creates `.dewey/` directory with default `config.yaml`.

### dewey index

Fetch and index content from all configured sources.

```bash
dewey index                  # all sources
dewey index --source github  # specific source type
```

### dewey status

Show index health and freshness.

```bash
dewey status
```

Output:

```
Dewey Index Status

Documents: 342 (local: 188, github: 127, web: 27)
Index size: 23 MB
Last indexed: 2026-03-22T10:00:00Z

Sources:
  local          ready    188 docs  10s ago
  gaze-issues    ready     89 docs  2h ago
  website-issues ready     38 docs  2h ago
  go-stdlib      ready     27 docs  3d ago

Embedding: granite-embedding:30m (342 docs embedded)
```

### dewey serve

Start the MCP server (stdio).

```bash
dewey serve --vault .
```

## graphthulhu Migration

Replacing graphthulhu with Dewey requires only an
MCP configuration change:

**Before** (`opencode.json`):
```json
{
  "mcp": {
    "knowledge-graph": {
      "command": ["graphthulhu", "serve", "--backend", "obsidian", "--vault", "."]
    }
  }
}
```

**After**:
```json
{
  "mcp": {
    "dewey": {
      "command": ["dewey", "serve", "--vault", "."]
    }
  }
}
```

Agent tool references change from `knowledge-graph_*`
to `dewey_*`. No other agent prompt changes needed for
existing structured query functionality.
