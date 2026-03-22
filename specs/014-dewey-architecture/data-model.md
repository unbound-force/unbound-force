# Data Model: Dewey Architecture

**Date**: 2026-03-22
**Branch**: `014-dewey-architecture`

## Entities

### Document

A unit of indexed content. May originate from local
disk, GitHub API, or web crawl.

| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique identifier (source-prefixed, e.g., `disk:specs/001/spec.md`, `gh:gaze:issue:142`, `web:pkg.go.dev:fmt`) |
| source_type | string | `"disk"`, `"github"`, `"web"` |
| source_name | string | Source configuration name (e.g., `"local"`, `"gaze-issues"`, `"go-stdlib"`) |
| title | string | Document title (filename, issue title, page title) |
| content | string | Raw text content (Markdown or converted-from-HTML) |
| url | string | Original URL or file path for attribution |
| fetched_at | timestamp | When the content was last fetched/indexed |
| checksum | string | Content hash for change detection |
| properties | map | YAML frontmatter or metadata properties (tags, status, priority) |

**Relationships**:
- A Document belongs to one Source.
- A Document has zero or more Embeddings (one per chunk).
- A Document has zero or more Blocks (from the KG parser).

### Block

A structural unit within a Document, inherited from
graphthulhu's parser. Represents a heading, paragraph,
code block, or list.

| Field | Type | Description |
|-------|------|-------------|
| id | string | UUID (stable across re-indexes per graphthulhu's design) |
| document_id | string | Parent document reference |
| type | string | `"heading"`, `"paragraph"`, `"code"`, `"list"`, etc. |
| level | int | Heading level (1-6) or nesting depth |
| content | string | Raw text of the block |
| parent_id | string | Parent block ID (for hierarchical nesting) |

**Relationships**:
- A Block belongs to one Document.
- A Block may have a parent Block (hierarchy).
- Blocks reference other Blocks via wikilinks (Link entity).

### Link

A connection between two Blocks or Documents, inherited
from graphthulhu's wikilink parser.

| Field | Type | Description |
|-------|------|-------------|
| source_id | string | Source block or document ID |
| target_id | string | Target block or document ID |
| type | string | `"wikilink"`, `"reference"`, `"backlink"` |

### Embedding

A vector representation of a Document chunk, generated
by the configured embedding model.

| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique identifier |
| document_id | string | Parent document reference |
| chunk_index | int | Position within the document's chunks |
| chunk_text | string | The text that was embedded |
| vector | float[] | Dense vector representation (dimension varies by configured model) |
| model_name | string | Embedding model used (for invalidation on model change) |

**Relationships**:
- An Embedding belongs to one Document.
- An Embedding is generated from a contiguous section
  of the Document's content (a chunk).

### Source

A pluggable content provider. Each source is
independently configurable and refreshable.

| Field | Type | Description |
|-------|------|-------------|
| name | string | Unique source name (e.g., `"local"`, `"gaze-issues"`, `"go-stdlib"`) |
| type | string | `"disk"`, `"github"`, `"web"` |
| config | map | Source-specific configuration (paths, repos, URLs, auth, refresh interval) |
| status | string | `"ready"`, `"indexing"`, `"failed"`, `"stale"` |
| last_indexed_at | timestamp | When this source was last successfully indexed |
| document_count | int | Number of documents from this source currently in the index |
| error | string | Error message if status is `"failed"` |

**Relationships**:
- A Source produces zero or more Documents.
- Sources are defined in the Workspace configuration.

### Workspace

The `.dewey/` directory and its configuration.

| Field | Type | Description |
|-------|------|-------------|
| root | path | Absolute path to the `.dewey/` directory |
| config_file | path | Path to `config.yaml` |
| index_file | path | Path to `index.db` (SQLite) |
| cache_dir | path | Path to `cache/` directory |

**Configuration schema** (`.dewey/config.yaml`):

```yaml
embedding:
  provider: ollama          # required
  model: granite-embedding:30m  # configurable

sources:
  - name: local
    type: disk
    # No additional config needed for the default disk source

  - name: gaze-issues
    type: github
    org: unbound-force
    repos:
      - gaze
    content:
      - issues
      - pulls
      - readme
    refresh: daily

  - name: go-stdlib
    type: web
    urls:
      - https://pkg.go.dev/std
    depth: 2
    refresh: weekly
```

## State Transitions

### Source Lifecycle

```text
ready ──index──▶ indexing ──success──▶ ready
  │                  │
  │                  └──failure──▶ failed
  │
  └──config removed──▶ (documents purged from index)
```

### Index Lifecycle

```text
(no index) ──first serve──▶ building ──complete──▶ ready
  ready ──serve start──▶ loading ──complete──▶ ready
  ready ──file change──▶ updating ──complete──▶ ready
  ready ──corruption──▶ rebuilding ──complete──▶ ready
```

### Embedding Lifecycle

```text
(no model) ──model installed──▶ embedding all
  embedding all ──complete──▶ ready
  ready ──document changed──▶ re-embedding
  ready ──model changed──▶ invalidated ──re-embed all──▶ ready
```

## Identity and Uniqueness

- **Document IDs** are source-prefixed to ensure
  uniqueness across sources: `{source_type}:{source_name}:{path_or_id}`
- **Block IDs** are UUIDs, stable across re-indexes
  (inherited from graphthulhu's deterministic UUID
  generation based on content position)
- **Embedding IDs** are `{document_id}:chunk:{index}`
- **Source names** are unique within a workspace
  configuration
