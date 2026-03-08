# Implementation Plan: Knowledge Graph Integration

**Branch**: `010-knowledge-graph-integration` | **Date**: 2026-03-08 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/010-knowledge-graph-integration/spec.md`

## Summary

Integrate graphthulhu's Obsidian backend as an MCP-based
knowledge graph service for the Unbound Force swarm. The service
reads Markdown files directly from project repositories, builds
an in-memory index with full-text search, wikilink graph, and
heading-based block trees, and exposes 30+ query/analysis tools
via MCP stdio transport. Hero agents (starting with Muti-Mind)
query the service to retrieve relevant project knowledge without
exhausting their context windows. The service runs as a
subprocess managed by OpenCode, starting at session init and
persisting across agent invocations.

## Technical Context

**Language/Version**: Go (graphthulhu is a Go project; this
integration consumes it as a pre-built binary -- no Go
development required in this repo)
**Primary Dependencies**: graphthulhu v0.4.0+ (MIT license,
`github.com/skridlevsky/graphthulhu`), OpenCode (agent runtime
with MCP client support)
**Storage**: In-memory index built from filesystem Markdown
files. No database. File watching via fsnotify for live
re-indexing.
**Testing**: Manual verification via MCP tool invocations from
OpenCode agents. Automated validation via the Gaze hero (when
deployed) for contract compliance.
**Target Platform**: macOS (primary development), Linux
(CI/CD). graphthulhu provides cross-platform Go binaries.
**Project Type**: Integration/configuration -- this feature
configures an external tool (graphthulhu) for use within the
Unbound Force ecosystem. Deliverables are configuration files,
documentation, and potentially an upstream PR to graphthulhu.
**Performance Goals**: Index 100 Markdown files in <10 seconds
at startup. Search queries return in <2 seconds. File system
changes reflected in <5 seconds.
**Constraints**: Must run headless (no GUI/desktop app). Must
use MIT-compatible license. Must maintain process boundary
(no importing graphthulhu source). Must support indexing
directories starting with `.` (requires upstream PR or fork).
**Scale/Scope**: Current meta repo has ~42 Markdown files.
Target scale is 100+ files per hero repo as backlogs grow.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after
Phase 1 design.*

### I. Autonomous Collaboration — PASS

- The knowledge graph service is internal tooling for each
  hero, not a cross-hero communication channel. Each hero repo
  runs its own instance; heroes do not share a knowledge graph
  instance (FR-020).
- The service does not introduce runtime coupling between
  heroes. It indexes local files and exposes them via MCP to
  the local agent runtime.
- The service does not modify the artifact-based communication
  model. Artifacts remain files in git. The knowledge graph is
  a read-only index over those files.
- No hero blocks waiting for another hero's knowledge graph.
  Each instance operates independently.

### II. Composability First — PASS

- The knowledge graph is an optional enhancement (FR-021).
  Heroes MUST function without it. When the service is not
  running, agents fall back to direct file reading (existing
  behavior).
- The service is independently installable (`go install` or
  binary download). No hero needs to be modified to support
  it -- it is additive tooling.
- When deployed alongside a hero, the service produces
  additive value (faster retrieval, graph analysis, gap
  detection) without degrading standalone functionality.
- The service auto-activates when configured in OpenCode's
  MCP settings. No manual wiring required beyond initial
  configuration.

### III. Observable Quality — PASS

- All service responses are machine-parseable MCP tool results
  (JSON over stdio). FR-022 requires this explicitly.
- Index statistics (page count, block count, link count) are
  available via the `health` and `graph_overview` tools,
  providing provenance and observability.
- The service is backed by graphthulhu's test suite (`go test
  ./...`). Search accuracy claims (SC-002, SC-003, SC-005) are
  verifiable through automated test scenarios.
- The service does not make quality claims that cannot be
  reproduced. All analysis results (orphans, clusters, gaps)
  are deterministic given the same input files.

**Gate evaluation**: All three principles PASS. No violations.
Proceeding to Phase 0.

## Project Structure

### Documentation (this feature)

```text
specs/010-knowledge-graph-integration/
├── plan.md              # This file
├── research.md          # Phase 0: upstream PR feasibility,
│                        #   OpenCode MCP config, alternatives
├── data-model.md        # Phase 1: entity definitions
├── quickstart.md        # Phase 1: setup guide
├── contracts/           # Phase 1: MCP tool interface contract
│   └── mcp-tools.md     #   Tool catalog and expected behavior
├── checklists/
│   └── requirements.md  # Spec quality checklist (existing)
└── tasks.md             # Phase 2 (/speckit.tasks - not this cmd)
```

### Source Code (repository root)

```text
opencode.json            # MCP server configuration for
                         #   graphthulhu (new file at project root)
```

**Structure Decision**: This is an integration/configuration
feature, not a source code feature. The deliverables are:
(1) OpenCode MCP server configuration (`opencode.json` at
project root) pointing to graphthulhu, (2) documentation for
setup and usage, and (3) potentially an upstream PR to
graphthulhu for hidden directory support. No new source code
directories are created in this repo.

## Constitution Check — Post-Design Re-evaluation

*Re-check after Phase 1 design decisions.*

### I. Autonomous Collaboration — PASS (confirmed)

Post-design, the stdio transport model reinforces autonomy.
Each hero's OpenCode instance spawns its own graphthulhu
subprocess. The `opencode.json` configuration is per-project,
so each hero repo has an independent knowledge graph instance.
No shared state, no cross-hero communication through the
knowledge graph.

The MCP tool contract (contracts/mcp-tools.md) defines a
read-only interface by default (`--read-only` flag). The
service does not produce artifacts -- it indexes existing
ones. This preserves the artifact-based communication model.

### II. Composability First — PASS (confirmed)

Post-design, the optional nature is explicit: the
`opencode.json` configuration can be removed or
`"enabled": false` set, and agents revert to direct file
reading. No hero code imports graphthulhu. The process
boundary is maintained via MCP stdio.

The quickstart.md documents installation as a single `go
install` command or binary download. No modification to
existing hero repos is required -- only adding an
`opencode.json` file.

### III. Observable Quality — PASS (confirmed)

Post-design, the MCP tool contract documents all tool inputs
and outputs. All responses are machine-parseable JSON via MCP
stdio. The `health` tool provides runtime observability
(version, backend, page count). The `graph_overview` tool
provides index statistics. These satisfy the Observable
Quality principle's requirement for machine-parseable output
and provenance metadata.

**Post-design gate evaluation**: All three principles PASS.
No violations introduced by design decisions. No complexity
tracking entries needed.
