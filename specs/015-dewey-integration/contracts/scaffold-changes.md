# Contract: Scaffold Changes for Dewey Integration

**Date**: 2026-03-22
**Branch**: `015-dewey-integration`

## MCP Configuration Change

### opencode.json (scaffold template + live copy)

**Before** (graphthulhu):
```json
{
  "mcp": {
    "knowledge-graph": {
      "type": "local",
      "command": ["graphthulhu", "serve", "--backend", "obsidian", "--vault", ".", "--include-hidden"],
      "enabled": true
    }
  }
}
```

**After** (Dewey):
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

## Agent File Tool Reference Changes

All agent persona files that reference knowledge
retrieval tools MUST be updated.

### Tool Name Mapping

| Old Tool Name | New Tool Name |
|--------------|---------------|
| `knowledge-graph_search` | `dewey_search` |
| `knowledge-graph_find_by_tag` | `dewey_find_by_tag` |
| `knowledge-graph_query_properties` | `dewey_query_properties` |
| `knowledge-graph_get_page` | `dewey_get_page` |
| `knowledge-graph_traverse` | `dewey_traverse` |
| `knowledge-graph_find_connections` | `dewey_find_connections` |
| (none) | `dewey_semantic_search` (NEW) |
| (none) | `dewey_similar` (NEW) |
| (none) | `dewey_semantic_search_filtered` (NEW) |

### Agent File Knowledge Retrieval Section

Each hero agent file gains a "Knowledge Retrieval"
section with the 3-tier pattern:

```markdown
## Knowledge Retrieval

When Dewey MCP tools are available, use them for
context retrieval. If Dewey is unavailable, fall back
to direct file operations.

**Tier 3 (Full Dewey)**:
- `dewey_semantic_search` for conceptual queries
- `dewey_search` for keyword queries
- `dewey_traverse` for relationship navigation
- [role-specific examples]

**Tier 2 (Graph-only, no embedding model)**:
- `dewey_search` for keyword queries
- `dewey_traverse` for relationship navigation
- Semantic search unavailable

**Tier 1 (No Dewey)**:
- Use Read tool for direct file access
- Use Grep for keyword search
- Reference convention packs for standards
```

## Doctor Health Check Group

### New Check Group: Dewey

| Check | Pass Condition | Fail Hint |
|-------|---------------|-----------|
| Dewey binary | `exec.LookPath("dewey")` succeeds | `Install Dewey: brew install unbound-force/tap/dewey` |
| Embedding model | Ollama model check succeeds | `Pull embedding model: ollama pull <model-name>` |
| Dewey workspace | `.dewey/` directory exists | `Initialize Dewey: dewey init` |

### Doctor Output Format

```
Dewey Knowledge Layer
  ✓ dewey binary found
  ✓ embedding model installed
  ✗ workspace not initialized
    Fix: Run `dewey init` in the project root
```

When Dewey binary is not found, the remaining checks
are skipped with a note:

```
Dewey Knowledge Layer
  ✗ dewey binary not found
    Install: brew install unbound-force/tap/dewey
  ⊘ embedding model (skipped: dewey not installed)
  ⊘ workspace (skipped: dewey not installed)
```

## Setup Step

### New Step: Dewey Installation

Position: After the Swarm plugin step, before the
scaffold step.

```
Installing Dewey...
  ✓ dewey installed via Homebrew
  ✓ embedding model pulled
```

If already installed:

```
Installing Dewey...
  ✓ dewey already installed (v0.1.0)
  ✓ embedding model already available
```

## Regression Test: No Stale References

Extend the existing
`TestScaffoldOutput_NoBareUnboundReferences` pattern
to also search for:
- `graphthulhu` (any occurrence)
- `knowledge-graph_` (tool name prefix)
- `knowledge-graph` (MCP server name)

Assert zero matches in all scaffold output files.
