# Quickstart: Dewey Integration

**Date**: 2026-03-22
**Branch**: `015-dewey-integration`

## Overview

After this change, the Unbound Force meta repo and all
projects scaffolded with `uf init` use Dewey as the
knowledge retrieval server instead of graphthulhu. All
hero agent personas gain semantic search capabilities
with graceful degradation when Dewey is unavailable.

## For Developers (Using the CLI)

### New Projects

```bash
# Scaffold a new project
uf init

# The generated opencode.json now references Dewey:
# "dewey": { "command": ["dewey", "serve", "--vault", "."] }

# Install Dewey (if not already)
brew install unbound-force/tap/dewey
ollama pull granite-embedding:30m

# Initialize the Dewey workspace
dewey init
dewey index

# Start coding -- agents now use Dewey for context
```

### Existing Projects (Migration)

```bash
# Re-run uf init to update tool-owned files
uf init

# This updates:
# - opencode.json (graphthulhu → Dewey)
# - Agent files (knowledge-graph_* → dewey_*)
# - User-owned custom files are preserved

# Install Dewey
brew install unbound-force/tap/dewey
ollama pull granite-embedding:30m
dewey init
dewey index
```

### Verify Setup

```bash
uf doctor
```

Output when fully configured:

```
Dewey Knowledge Layer
  ✓ dewey binary found
  ✓ embedding model installed
  ✓ workspace initialized
```

### Without Dewey

Everything still works. Heroes fall back to direct
file reads and convention packs. You get less context,
not broken functionality.

## For Contributors (Meta Repo)

### What Changed

| Category | Files | Change |
|----------|-------|--------|
| MCP config | `opencode.json` | `knowledge-graph` → `dewey` |
| Scaffold template | `internal/scaffold/assets/opencode.json` | Same change (embedded copy) |
| Agent files (live) | `.opencode/agents/*.md` | Dewey tools + 3-tier fallback |
| Agent files (scaffold) | `internal/scaffold/assets/opencode/agents/*.md` | Same changes (embedded copies) |
| Doctor | `internal/doctor/checks.go` | New Dewey health check group |
| Setup | `internal/setup/setup.go` | New Dewey installation step |
| Docs | `AGENTS.md` | Updated references |

### Testing

```bash
# Run all tests
go test -race -count=1 ./...

# Verify no stale graphthulhu references in scaffold
grep -rn 'graphthulhu\|knowledge-graph' \
  internal/scaffold/assets/ opencode.json .opencode/
# Should return zero matches
```

### What's NOT Changed

- Completed specs (`specs/001-*` through `specs/014-*`)
- Archived OpenSpec changes
- User-owned custom files (convention packs, custom
  agent files)
- The Dewey binary itself (lives in `unbound-force/dewey`)
