# Implementation Plan: Dewey Architecture

**Branch**: `014-dewey-architecture` | **Date**: 2026-03-22 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/014-dewey-architecture/spec.md`

## Summary

Define the architectural design for Dewey, a per-repo
MCP knowledge server that combines graphthulhu's
structured knowledge graph with vector-based semantic
search and pluggable content sources. Dewey is a hard
fork of graphthulhu, extended with persistent storage,
embedding generation via a locally-run model, and
content sources (local disk, GitHub API, web crawl).

This spec lives in the meta repo and defines the
architecture. The actual implementation lives in
`unbound-force/dewey`. This plan covers:

1. The architectural design artifacts produced in this
   repo (data model, contracts, research decisions)
2. The integration work in this repo (scaffold config
   update, agent file updates, doctor/setup checks)
3. The phased implementation roadmap for the dewey repo
   (documented here, executed there)

## Technical Context

**Language/Version**: Go 1.24+ (same as graphthulhu)
**Primary Dependencies**: `github.com/modelcontextprotocol/go-sdk/mcp`
(MCP SDK, inherited from graphthulhu), SQLite (new, for
persistence), Ollama API (new, for embeddings)
**Storage**: SQLite for persistent indexes (knowledge
graph + vector embeddings), local filesystem for crawl
cache
**Testing**: Standard library `testing` package,
`go test -race -count=1`
**Target Platform**: macOS, Linux (MCP server binary)
**Project Type**: MCP server (CLI binary that speaks
MCP via stdio)
**Performance Goals**: <1s warm start from persistent
index (SC-001 scenario 2), <3s cold start including
initial index build for 200 documents (SC-001 overall),
<100ms per semantic query, <500ms per structured query
at 10k documents
**Constraints**: All data processed locally (no cloud
services). Embedding model runs locally via Ollama. No
data leaves the developer's machine. Enterprise-grade
licensing provenance required for default embedding
model.
**Scale/Scope**: 200 local documents (typical repo) +
moderate external sources (5 GitHub repos, 3 web crawl
targets). Index size <50MB. This architectural spec
produces ~10 files in this repo; the dewey repo
implementation is ~20 Go source files across 6 packages.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check
after Phase 1 design.*

### I. Autonomous Collaboration -- PASS

Dewey communicates through MCP tool calls (a
well-defined protocol), not runtime coupling. Agent
personas query Dewey through standard MCP tools; they
do not couple to Dewey's internals. Dewey's index is
a persistent artifact on disk, not an in-memory
ephemeral state. If Dewey is unavailable, agents
continue functioning (FR-025).

### II. Composability First -- PASS

Dewey is independently installable
(`brew install unbound-force/tap/dewey`). No hero
requires Dewey for its core function (FR-025). Heroes
auto-detect Dewey's presence via MCP tool availability
and activate enhanced functionality when it's present
(FR-024). Dewey can be used by any project, not just
Unbound Force repos.

### III. Observable Quality -- PASS

Dewey produces machine-parseable output (MCP tool
responses are JSON). The `dewey status` command (FR-020)
provides observable index health metrics. Search results
include source provenance metadata (FR-017). The index
itself is a persistent, inspectable artifact on disk.

### IV. Testability -- PASS

Dewey's architecture inherits graphthulhu's testable
patterns (backend interface with marker interfaces,
dependency injection). All new components (persistence
layer, embedding client, content sources) implement
interfaces that can be stubbed for testing. The source
architecture uses a common interface (list, fetch, diff,
metadata) that is independently testable per source
type. No external services required for core tests
(SQLite is embedded, Ollama is optional and mockable).

**Coverage strategy**: Unit tests for each package
(persistence, embedding, sources, CLI). Integration
tests for the MCP tool surface (structured + semantic
queries). No e2e tests needed -- the MCP protocol is
the contract boundary. Coverage target: maintain
graphthulhu's existing test coverage (~52% line
coverage) and improve to 70%+ for new packages.

## Project Structure

### Documentation (this feature)

```text
specs/014-dewey-architecture/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── mcp-tools.md     # MCP tool interface contract
└── checklists/
    └── requirements.md  # Already created
```

### Dewey Source Code (unbound-force/dewey repo)

```text
unbound-force/dewey/
├── main.go              # CLI entry point (renamed from graphthulhu)
├── server.go            # MCP server (inherited + extended)
├── cli.go               # CLI commands (inherited + extended)
├── store/               # NEW: persistent storage package
│   ├── sqlite.go        # SQLite backend for KG + vectors
│   ├── sqlite_test.go
│   └── migration.go     # Schema versioning
├── embed/               # NEW: embedding generation package
│   ├── ollama.go        # Ollama API client
│   ├── ollama_test.go
│   └── chunker.go       # Document chunking strategy
├── source/              # NEW: content source package
│   ├── source.go        # Source interface
│   ├── disk.go          # Local filesystem source
│   ├── github.go        # GitHub API source
│   ├── web.go           # Web crawl source
│   └── *_test.go
├── tools/               # Extended MCP tools
│   ├── semantic.go      # NEW: semantic search tools
│   └── ... (inherited)
├── graph/               # Inherited graph package
├── parser/              # Inherited parser package
├── backend/             # Extended backend package
├── vault/               # Inherited vault package (refactored)
├── types/               # Inherited types package
└── config/              # NEW: configuration package
    ├── config.go
    └── config_test.go
```

### Integration Points (this repo)

```text
unbound-force/unbound-force/
├── opencode.json                    # MCP config: dewey replaces graphthulhu
├── .opencode/agents/*.md            # Agent files: dewey_* tool usage
├── internal/scaffold/assets/        # Scaffold templates: dewey config
├── internal/doctor/checks.go        # Dewey health check
├── internal/setup/setup.go          # Dewey installation
└── AGENTS.md                        # Documentation update
```

**Structure Decision**: Dewey is a separate repo
(`unbound-force/dewey`) with its own Go module. This
repo produces the architectural design artifacts and
will later receive integration updates (scaffold, agent
files, doctor/setup). The integration work is a
separate spec (Phase 4 in the Dewey orchestration plan).

## Complexity Tracking

No constitution violations. No complexity justifications
needed.
