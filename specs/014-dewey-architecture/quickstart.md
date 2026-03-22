# Quickstart: Dewey

**Date**: 2026-03-22
**Branch**: `014-dewey-architecture`

## Overview

Dewey is a per-repository knowledge server that
combines structured knowledge graph traversal with
semantic search. It replaces graphthulhu, adding
persistent indexes, vector embeddings, and pluggable
content sources (local files, GitHub issues, web docs).

## Installation

```bash
# Install Dewey
brew install unbound-force/tap/dewey

# Install the embedding model (optional but recommended)
brew install ollama
ollama pull granite-embedding:30m
```

Without the embedding model, Dewey works in "graph-only"
mode -- structured queries work, semantic search is
unavailable.

## First Use

```bash
# Initialize in your project
cd ~/my-project
dewey init

# Build the initial index
dewey index

# Check what's indexed
dewey status

# Start the MCP server (OpenCode does this automatically)
dewey serve
```

## Configuring Sources

Edit `.dewey/config.yaml` to add external sources:

```yaml
embedding:
  provider: ollama
  model: granite-embedding:30m

sources:
  - name: local
    type: disk

  - name: org-issues
    type: github
    org: unbound-force
    repos:
      - gaze
      - website
    content:
      - issues
      - pulls
      - readme
    refresh: daily

  - name: go-docs
    type: web
    urls:
      - https://pkg.go.dev/std
    depth: 2
    refresh: weekly
```

Then rebuild the index:

```bash
dewey index
```

## OpenCode Integration

Update `opencode.json` to use Dewey instead of
graphthulhu:

```json
{
  "mcp": {
    "dewey": {
      "type": "local",
      "command": ["dewey", "serve", "--vault", "."],
      "enabled": true
    }
  }
}
```

The `uf init` scaffold command will generate this
configuration automatically in future versions.

## Querying Dewey

### From Agent Prompts

Agent personas use Dewey tools in their MCP tool calls:

```
dewey_search "authentication timeout"
dewey_semantic_search "login session expiry"
dewey_find_by_tag "priority:P1"
dewey_traverse "specs/008-swarm-orchestration/spec.md"
dewey_similar "specs/012-swarm-delegation/spec.md"
```

### Semantic vs Keyword Search

**Keyword** (`dewey_search`): Finds documents containing
the exact terms. "login session expiry" only matches
documents with those words.

**Semantic** (`dewey_semantic_search`): Finds documents
about the same concept. "login session expiry" also
matches documents about "authentication timeout" and
"token refresh failure."

### Filtered Semantic Search

Combine semantic relevance with structural filters:

```
dewey_semantic_search_filtered
  query: "test quality patterns"
  source_type: "github"
  source_name: "gaze-issues"
```

This finds conceptually relevant GitHub issues from the
Gaze repo, not just keyword matches.

## Graceful Degradation

Dewey is an enhancement, not a dependency. Each hero
agent works at three tiers:

| Tier | Dewey Status | Agent Behavior |
|------|-------------|----------------|
| 1 | Not installed | Direct file reads + CLI queries |
| 2 | Installed, no embedding model | Structured queries only (search, traverse, tags) |
| 3 | Full (model installed) | Structured + semantic queries across all sources |

All heroes function at every tier. Higher tiers provide
richer context, not essential capabilities.

## Common Operations

### Refresh External Sources

```bash
dewey index                  # all sources
dewey index --source github  # just GitHub
dewey index --source web     # just web crawls
```

### Check Index Health

```bash
dewey status
```

Shows document count, source breakdown, freshness per
source, index size, and embedding model status.

### Change Embedding Model

Edit `.dewey/config.yaml`:

```yaml
embedding:
  provider: ollama
  model: granite-embedding:278m  # multilingual
```

Then rebuild embeddings:

```bash
dewey index
```

Dewey detects the model change and re-embeds all
documents automatically.
