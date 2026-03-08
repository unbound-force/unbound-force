# MCP Tool Interface Contract: Knowledge Graph

**Feature**: 010-knowledge-graph-integration
**Date**: 2026-03-08
**Protocol**: Model Context Protocol (MCP) via stdio transport
**Server name**: `knowledge-graph`

## Overview

This document defines the MCP tools exposed by the knowledge
graph service to hero agents via OpenCode. Tools are registered
with the `knowledge-graph` server name prefix. The service uses
graphthulhu's Obsidian backend, which exposes a subset of the
full 37-tool catalog (Logseq-only tools are excluded).

## Tool Catalog

### Navigate Tools

#### `get_page`

Retrieve a page with its full recursive block tree, parsed
links, tags, and properties.

| Parameter | Type   | Required | Description           |
|-----------|--------|----------|-----------------------|
| name      | string | Yes      | Page name (path-qualified, case-insensitive) |

**Returns**: Page entity with block tree, properties, links.
**Error**: Page not found returns MCP error.

#### `get_block`

Retrieve a block by UUID with its ancestor chain, children,
and sibling blocks.

| Parameter | Type   | Required | Description           |
|-----------|--------|----------|-----------------------|
| uuid      | string | Yes      | Block UUID            |

**Returns**: Block with parent chain and sibling context.

#### `list_pages`

List all indexed pages with optional filtering.

| Parameter | Type   | Required | Description           |
|-----------|--------|----------|-----------------------|
| namespace | string | No       | Filter by directory prefix |
| tag       | string | No       | Filter by tag         |
| sort      | string | No       | Sort: name, modified, created |

**Returns**: Array of page summaries (name, properties, block
count).

#### `get_links`

Get all forward and backward links for a page.

| Parameter | Type   | Required | Description           |
|-----------|--------|----------|-----------------------|
| name      | string | Yes      | Page name             |

**Returns**: Outbound links (with containing blocks) and inbound
backlinks (with source pages and blocks).

#### `traverse`

Find the shortest path between two pages through the link
graph using breadth-first search.

| Parameter | Type   | Required | Description           |
|-----------|--------|----------|-----------------------|
| from      | string | Yes      | Source page name      |
| to        | string | Yes      | Target page name      |

**Returns**: Ordered list of pages forming the shortest path.
**Error**: Returns empty path if pages are not connected.

### Search Tools

#### `search`

Full-text search across all indexed content with contextual
results.

| Parameter | Type   | Required | Description           |
|-----------|--------|----------|-----------------------|
| query     | string | Yes      | Search terms (AND logic) |
| limit     | int    | No       | Max results (default 20) |

**Returns**: Array of Search Results with parent chain and
sibling context for each match.

#### `query_properties`

Find pages by YAML frontmatter property values.

| Parameter | Type   | Required | Description           |
|-----------|--------|----------|-----------------------|
| property  | string | Yes      | Property name         |
| value     | string | Yes      | Value to match        |
| operator  | string | No       | eq (default), contains, gt, lt |

**Returns**: Array of pages matching the property query.

#### `find_by_tag`

Search for blocks and pages by tag.

| Parameter | Type   | Required | Description           |
|-----------|--------|----------|-----------------------|
| tag       | string | Yes      | Tag name (without `#`) |

**Returns**: Array of blocks containing the tag, with page
context.

### Analyze Tools

#### `graph_overview`

Get aggregate statistics of the knowledge graph.

| Parameter | Type | Required | Description              |
|-----------|------|----------|--------------------------|
| (none)    |      |          | No parameters required   |

**Returns**: Total pages, blocks, links, most-connected pages,
namespace breakdown, orphan count.

#### `find_connections`

Discover how two pages are connected.

| Parameter | Type   | Required | Description           |
|-----------|--------|----------|-----------------------|
| page_a    | string | Yes      | First page name       |
| page_b    | string | Yes      | Second page name      |

**Returns**: Direct links, shortest paths, and shared
connections between the pages.

#### `knowledge_gaps`

Identify structural deficiencies in the knowledge graph.

| Parameter | Type | Required | Description              |
|-----------|------|----------|--------------------------|
| (none)    |      |          | No parameters required   |

**Returns**: Orphan pages, dead-end pages, weakly-linked areas,
and missing references.

#### `list_orphans`

List pages with no inbound or outbound links.

| Parameter | Type | Required | Description              |
|-----------|------|----------|--------------------------|
| (none)    |      |          | No parameters required   |

**Returns**: Array of orphan page names with block counts and
property status.

#### `topic_clusters`

Identify connected components and topic groupings.

| Parameter | Type | Required | Description              |
|-----------|------|----------|--------------------------|
| (none)    |      |          | No parameters required   |

**Returns**: Connected components with hub identification.

### Operational Tools

#### `health`

Check server status.

| Parameter | Type | Required | Description              |
|-----------|------|----------|--------------------------|
| (none)    |      |          | No parameters required   |

**Returns**: Version, backend type, read-only mode status,
page count.

#### `reload`

Force a complete re-index of all files from disk.

| Parameter | Type | Required | Description              |
|-----------|------|----------|--------------------------|
| (none)    |      |          | No parameters required   |

**Returns**: Confirmation with number of pages re-indexed.

### Journal Tools

#### `journal_range`

Retrieve journal entries across a date range.

| Parameter | Type   | Required | Description           |
|-----------|--------|----------|-----------------------|
| from      | string | No       | Start date (YYYY-MM-DD) |
| to        | string | No       | End date (YYYY-MM-DD) |

**Returns**: Journal pages with full block trees within the
date range.

#### `journal_search`

Search within journal entries.

| Parameter | Type   | Required | Description           |
|-----------|--------|----------|-----------------------|
| query     | string | Yes      | Search terms          |
| from      | string | No       | Start date filter     |
| to        | string | No       | End date filter       |

**Returns**: Matching blocks within journal pages.

### Decision Tools

#### `decision_check`

Surface open, overdue, and resolved decisions.

| Parameter | Type | Required | Description              |
|-----------|------|----------|--------------------------|
| (none)    |      |          | No parameters required   |

**Returns**: Decision blocks with status, deadline, and
deferral history.

#### `decision_create`

Create a decision block with `#decision` tag, deadline,
options, and context. (Only available when not in read-only
mode.)

| Parameter  | Type   | Required | Description           |
|------------|--------|----------|-----------------------|
| page       | string | Yes      | Target page name      |
| question   | string | Yes      | Decision question     |
| deadline   | string | No       | Deadline (YYYY-MM-DD) |
| options    | string | No       | Comma-separated options |
| context    | string | No       | Additional context    |

**Returns**: Created block with UUID.

## Tools NOT Available (Logseq-only)

The following graphthulhu tools require the Logseq backend and
are not available with the Obsidian backend:

- `get_references` -- block reference resolution via `((uuid))`
- `query_datalog` -- raw DataScript/Datalog queries
- `flashcard_overview`, `flashcard_due`, `flashcard_create`
- `list_whiteboards`, `get_whiteboard`

## Response Format

All tool responses are JSON objects returned via MCP stdio.
Responses include the tool result payload; MCP protocol
framing (request ID, method) is handled by the MCP SDK.

## Error Handling

- **Page/block not found**: MCP error response with descriptive
  message.
- **Invalid parameters**: MCP error response with parameter
  validation details.
- **Internal error**: MCP error response; the service logs
  details to stderr (visible in OpenCode's MCP server logs).
