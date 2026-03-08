# Quickstart: Knowledge Graph Integration

**Feature**: 010-knowledge-graph-integration
**Date**: 2026-03-08

## Prerequisites

- Go toolchain installed (for `go install`), OR willingness
  to download a pre-built binary
- OpenCode installed and configured for the project
- A project repository with Markdown files to index

## Step 1: Install graphthulhu

```bash
# Option A: Install via Go (recommended for developers)
go install github.com/skridlevsky/graphthulhu@latest

# Option B: Download binary from GitHub Releases
# Visit https://github.com/skridlevsky/graphthulhu/releases
# Download the binary for your platform (macOS/Linux/Windows)
# Extract and move to a directory in your PATH
```

Verify installation:

```bash
graphthulhu version
```

## Step 2: Verify It Works Standalone

Before configuring OpenCode, test graphthulhu directly against
your project:

```bash
# From the project root directory
graphthulhu serve \
  --backend obsidian \
  --vault . \
  --include-hidden \
  --read-only \
  --http :12315
```

In a separate terminal, verify the service is running:

```bash
# Check health (requires curl + JSON parsing)
curl -s http://127.0.0.1:12315/health
```

Stop the server (Ctrl+C) once verified. The OpenCode
integration will manage the service lifecycle automatically.

**Note**: The `--include-hidden` flag requires the upstream PR
to graphthulhu (see research.md R-002). Until the PR is merged,
content in `.specify/` and `.opencode/` will not be indexed.
Full-text search and graph analysis will still work for all
non-hidden directories.

## Step 3: Configure OpenCode

Create or update `opencode.json` in the project root:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "knowledge-graph": {
      "type": "local",
      "command": [
        "graphthulhu",
        "serve",
        "--backend", "obsidian",
        "--vault", ".",
        "--include-hidden",
        "--read-only"
      ],
      "enabled": true
    }
  }
}
```

**Configuration notes**:

- `"type": "local"` -- graphthulhu runs as a subprocess,
  communicating via stdio. OpenCode manages its lifecycle.
- `"--vault", "."` -- indexes from the project root. OpenCode
  sets the working directory to the project root when spawning
  the subprocess.
- `"--include-hidden"` -- indexes `.specify/`, `.opencode/`,
  and other hidden directories. Remove this flag if you only
  want non-hidden content indexed.
- `"--read-only"` -- disables all write operations. The
  knowledge graph is a read-only index over project files.

## Step 4: Use from an Agent

Start OpenCode. The knowledge graph service starts
automatically as a subprocess. All agents in the project can
now use the `knowledge-graph` MCP tools.

Example agent interactions:

```
# Search across all project files
"Search the knowledge graph for 'acceptance criteria'"

# Get a specific spec
"Get the page for specs/004-muti-mind-architecture/spec"

# Analyze project structure
"Run graph_overview to see project statistics"

# Find orphaned documents
"Use knowledge_gaps to find documents with no links"

# List all specs
"List pages in the specs namespace"
```

## Step 5: Verify Agent Access

Ask any agent to invoke the `health` tool:

```
"Check the knowledge graph health status"
```

Expected response includes: version, backend type (`obsidian`),
read-only status (`true`), and page count.

## Troubleshooting

**"graphthulhu: command not found"**

The binary is not in your PATH. Either:
- Run `go install github.com/skridlevsky/graphthulhu@latest`
  and ensure `$GOPATH/bin` is in your PATH
- Download the binary and move it to `/usr/local/bin/` or
  another directory in your PATH

**MCP server not starting**

Check OpenCode's MCP server logs for error output from the
graphthulhu subprocess. Common issues:
- Invalid `--vault` path (ensure the project root exists)
- Port conflict (only applies to `--http` mode, not stdio)
- Missing `--backend obsidian` flag

**Files not appearing in search results**

- Non-Markdown files are not indexed (only `.md` files)
- Files in hidden directories (`.specify/`, `.opencode/`)
  require the `--include-hidden` flag
- After creating a new file, wait up to 5 seconds for the
  file watcher to detect and index it
- Use the `reload` tool to force a complete re-index

**Search returns no results for known content**

- The inverted index requires terms of at least 2 characters
- Search uses AND logic -- all terms must appear in the same
  block
- Single-character terms and common Markdown syntax (`#`, `*`,
  `-`) are stripped during tokenization
