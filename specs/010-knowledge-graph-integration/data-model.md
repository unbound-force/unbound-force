# Data Model: Knowledge Graph Integration

**Feature**: 010-knowledge-graph-integration
**Date**: 2026-03-08
**Source**: Spec Key Entities + graphthulhu Obsidian backend
internals

## Overview

The knowledge graph data model is an in-memory read-only index
built from Markdown files on disk. There is no persistent
database -- the index is rebuilt at service startup and
maintained via file system watching. All entities are derived
from the filesystem; the Markdown files are the source of truth.

## Entities

### Page

A single Markdown file in the project. The fundamental unit of
the knowledge graph.

| Attribute       | Type               | Source                    |
|-----------------|--------------------|---------------------------|
| name            | string             | Relative path without `.md` extension (e.g., `specs/004-muti-mind-architecture/spec`) |
| file_path       | string             | Absolute path to the `.md` file on disk |
| properties      | map[string]any     | YAML frontmatter parsed from `---` fences |
| aliases         | []string           | `aliases` key from YAML frontmatter (if present) |
| blocks          | []Block            | Hierarchical block tree (root-level blocks) |
| outbound_links  | []Link             | All `[[wikilinks]]` found in any block |
| inbound_links   | []Backlink         | Pages that link to this page |
| tags            | []string           | All `#tags` found in any block |
| is_journal      | bool               | True if path starts with the daily notes folder |
| created_at      | timestamp          | File creation time (from filesystem) |
| updated_at      | timestamp          | File modification time (from filesystem) |

**Identity**: Path-qualified name (case-insensitive). Two files
in different subdirectories with the same filename are distinct
pages (e.g., `specs/004/spec` vs `specs/009/spec`).

**Aliases**: If YAML frontmatter contains an `aliases` key with
a list value, each alias is registered as an additional lookup
key for the same page.

### Block

A section of a page defined by heading boundaries. Blocks form
a tree hierarchy based on heading levels (H1 > H2 > H3, etc.).

| Attribute        | Type             | Source                      |
|------------------|------------------|-----------------------------|
| uuid             | string (UUID v4) | Extracted from `<!-- id: UUID -->` HTML comment, or deterministic SHA-256 of `filepath:lineNumber` |
| content          | string           | Raw Markdown text of the section (heading line + body until next heading) |
| heading_level    | int (0-6)        | 0 = pre-heading content, 1-6 = H1-H6 |
| children         | []Block          | Sub-blocks (headings at a deeper level) |
| parent           | *Block           | Parent block (heading at a shallower level, or nil for root) |
| links            | []string         | `[[wikilink]]` targets extracted from content |
| tags             | []string         | `#tag` values extracted from content |
| inline_props     | map[string]string | Logseq-style `key:: value` properties from content |
| task_marker      | string           | TODO/DOING/DONE/LATER/NOW/WAITING/CANCELLED (if present) |
| priority         | string           | A/B/C from `[#X]` marker (if present) |

**Identity**: UUID. Stable across file modifications if the
`<!-- id: UUID -->` comment is embedded. Falls back to
deterministic generation from file path + line number, which
changes if content is reordered.

**Tree construction**: Lines are scanned sequentially. Each
heading starts a new section. A stack-based algorithm assigns
parent-child relationships: a heading becomes a child of the
nearest ancestor heading with a strictly lower level.

### Link

A directional reference from one page to another via a
`[[wikilink]]` in block content.

| Attribute      | Type   | Source                           |
|----------------|--------|----------------------------------|
| source_page    | string | Page name containing the link    |
| source_block   | string | Block UUID containing the link   |
| target_page    | string | Page name referenced by the link |

**Bidirectionality**: For every Link from A to B, page B's
inbound backlinks include a Backlink entry pointing back to A.
The backlink index is rebuilt from scratch on every file change.

### Backlink

A reverse reference indicating that another page links to this
page.

| Attribute      | Type   | Source                           |
|----------------|--------|----------------------------------|
| from_page      | string | Page name that contains the link |
| block_uuid     | string | UUID of the block with the link  |
| block_content  | string | Content summary of the block     |

### Search Result

A block matching a search query, enriched with navigational
context.

| Attribute      | Type      | Source                          |
|----------------|-----------|----------------------------------|
| page_name      | string    | Page containing the match        |
| block_uuid     | string    | UUID of the matching block       |
| content        | string    | Full content of the matching block |
| parent_chain   | []Block   | Ancestor blocks from root to parent |
| siblings       | []Block   | Sibling blocks at the same level |

**Search semantics**: Full-text search uses AND logic -- all
query terms must appear in a block. The inverted index maps
lowercase terms (minimum 2 characters) to block references.
Results are limited to 20 by default.

### Graph Overview

Aggregate statistics of the entire knowledge graph.

| Attribute          | Type             | Source              |
|--------------------|------------------|---------------------|
| total_pages        | int              | Count of indexed pages |
| total_blocks       | int              | Count of all blocks across pages |
| total_links        | int              | Count of all wikilink references |
| most_connected     | []PageSummary    | Pages with the most links |
| namespaces         | []NamespaceStat  | Breakdown by directory prefix |
| orphan_count       | int              | Pages with zero links |

### Knowledge Gap

A structural deficiency detected by graph analysis.

| Attribute    | Type   | Description                       |
|--------------|--------|-----------------------------------|
| type         | enum   | `orphan`, `dead_end`, `missing_reference` |
| page_name    | string | Page affected by the gap          |
| details      | string | Human-readable description        |

**Gap types**:

- `orphan`: Page with zero inbound and zero outbound links.
- `dead_end`: Page with outbound links but zero inbound links
  (nothing references it).
- `missing_reference`: A `[[wikilink]]` target that does not
  correspond to any indexed page.

## Relationships

```text
Page 1---* Block          (a page contains many blocks)
Block 1---* Block         (blocks nest hierarchically)
Block *---* Link          (blocks contain wikilinks)
Link *---1 Page           (links point to target pages)
Page 1---* Backlink       (pages have inbound backlinks)
Page *---* Page           (pages connect through links)
```

## State Lifecycle

The data model has no persistent state transitions. The index
exists in two states:

1. **Building**: Service startup. Files are walked, parsed, and
   indexed. Backlinks and search index are constructed. Duration:
   proportional to file count (target <10s for 100 files).
2. **Ready**: Serving queries. File watcher detects changes and
   triggers incremental re-indexing. Each file change triggers a
   full backlink rebuild (no incremental backlink updates).

There is no "degraded" state -- if the service process exits,
the index is lost and must be rebuilt on next startup.

## Scale Assumptions

| Metric              | Current    | Target       |
|---------------------|------------|--------------|
| Markdown files      | ~42        | 100-500      |
| Total content       | ~50K tokens| 100K-500K    |
| Wikilinks per file  | 0 (today)  | 5-20 (after enrichment) |
| Properties per file | 0 (today)  | 5-10 (after enrichment) |
| Index rebuild time  | <1s        | <10s         |
| Query response time | <100ms     | <2s          |
