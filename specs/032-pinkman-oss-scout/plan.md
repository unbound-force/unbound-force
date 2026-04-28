# Implementation Plan: Pinkman OSS Scout

**Branch**: `032-pinkman-oss-scout` | **Date**: 2026-04-22 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/032-pinkman-oss-scout/spec.md`

## Summary

Pinkman is a non-hero utility agent that discovers open
source projects by domain keyword, classifies their
licenses against the OSI-approved list, lists direct
dependencies with overlap detection across results,
quantifies industry trend signals, audits existing
dependency health, and generates structured adoption
recommendation reports. It is implemented as an OpenCode
agent file with a `/scout` slash command, following the
established agent + command pattern (Specs 006, 007,
031). Pinkman uses `webfetch` for public data retrieval
and Dewey for persistent scouting memory.

## Technical Context

**Language/Version**: Markdown (agent file, command
file), Go 1.24+ (scaffold engine -- asset embedding
and test updates only)
**Primary Dependencies**: OpenCode agent runtime,
`webfetch` tool (for public data source access),
`read`/`write`/`edit` tools (for manifest parsing and
report generation), Dewey MCP tools (optional -- for
persistent scouting memory)
**Storage**: Scouting results stored as Markdown files
with YAML frontmatter at `.uf/pinkman/reports/` for
local persistence. Optionally stored in Dewey knowledge
graph via `dewey_store_learning` for cross-session
retrieval.
**Testing**: Standard library `testing` package for
scaffold drift detection tests. Manual verification
for agent behavior (invoke agent, verify output format
and content).
**Target Platform**: macOS, Linux (same as existing
`uf` CLI)
**Project Type**: CLI tool + OpenCode agent ecosystem
**Performance Goals**: Discovery results returned within
60 seconds per SC-001. Report generation under 30
seconds for a single project.
**Constraints**: Must use only publicly available data
sources. Must not require API keys for basic operation
(API keys MAY be used for higher rate limits when
available).
**Scale/Scope**: Typical invocation returns 5-20
scouted projects. Dependency overlap analysis operates
on the result set (not the entire package ecosystem).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check
after Phase 1 design.*

### I. Autonomous Collaboration — PASS

Pinkman produces self-describing artifacts (scouting
reports with metadata: producer identity, version,
timestamp, artifact type) per the artifact envelope
format. It does not require synchronous interaction
with any other hero. It consumes no hero artifacts as
input -- it reads only public data sources and local
dependency manifests. Its outputs can be consumed by
Muti-Mind (for adoption prioritization), Cobalt-Crush
(for adoption implementation), and Mx F (for dependency
health metrics) without requiring Pinkman to be present.

### II. Composability First — PASS

Pinkman is independently installable and usable without
any other hero. It delivers its core value (OSS
discovery and license checking) when deployed alone. It
exposes extension points: Dewey integration is optional
(graceful degradation when unavailable), and its report
format uses the standard artifact envelope for
inter-hero consumption. It does not require any hero as
a prerequisite. When deployed alongside other heroes, it
produces additive value (e.g., Muti-Mind can prioritize
adoption candidates from Pinkman's reports).

### III. Observable Quality — PASS

All Pinkman outputs are structured Markdown with YAML
frontmatter (machine-parseable). Reports include
provenance metadata (producer: pinkman, version,
timestamp, query context). License verdicts reference
the OSI-approved list with SPDX identifiers. Trend
indicators are quantitative and comparable across runs.
Quality claims (SC-002: zero false positives for license
compatibility) are verifiable through automated tests
that compare Pinkman's license verdicts against the
canonical OSI list.

### IV. Testability — PASS

The agent file is tested through scaffold drift
detection (embedded asset matches canonical source).
License classification logic can be tested by invoking
the agent with known projects whose licenses are
established. Dependency overlap detection can be tested
with controlled manifest files. The coverage strategy
is:
- **Unit**: Scaffold drift detection tests (Go test)
- **Integration**: Manual invocation tests with known
  OSS projects
- **Acceptance**: Verify each acceptance scenario from
  the spec against actual agent output

No constitution violations. No complexity tracking
entries needed.

## Project Structure

### Documentation (this feature)

```text
specs/032-pinkman-oss-scout/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── agent-interface.md
└── checklists/
    └── requirements.md  # Spec quality checklist
```

### Source Code (repository root)

```text
# Agent and command files (deployed by uf init)
.opencode/
├── agents/
│   └── pinkman.md               # Pinkman agent file (user-owned)
└── command/
    └── scout.md                 # /scout slash command (tool-owned)

# Scaffold engine assets (embedded copies)
internal/scaffold/assets/
├── opencode/
│   ├── agents/
│   │   └── pinkman.md           # Embedded copy of agent
│   └── command/
│       └── scout.md             # Embedded copy of command

# Local storage (runtime data, gitignored)
.uf/pinkman/
└── reports/                     # Scouting report artifacts
    └── YYYY-MM-DDTHH-MM-SS-<query>.md
```

**Structure Decision**: Single-project layout with two
new Markdown files (agent + command) deployed through
the existing scaffold engine pattern. No new Go packages
needed -- Pinkman's logic is expressed entirely in agent
instructions. Local report storage follows the `.uf/`
convention (Spec 025). The scaffold engine's `fs.WalkDir`
automatically picks up new files under
`internal/scaffold/assets/`.

## Design Decisions

### D1: Agent-Only Implementation (No Go CLI Backend)

Pinkman is implemented entirely as an OpenCode agent
file + slash command, with no Go CLI backend package.
The agent uses `webfetch` for data retrieval and
`read`/`write` for file operations.

**Rationale**: Pinkman's core operations (web scraping,
license lookup, trend analysis, report generation) are
inherently AI reasoning tasks that benefit from the
agent's language understanding capabilities. A Go CLI
backend would add complexity without significant value
-- unlike Muti-Mind (which needs structured backlog
parsing) or Mx F (which needs metrics aggregation),
Pinkman's operations are better suited to the agent
runtime. This also keeps the implementation scope small
(2 Markdown files vs. a full Go package).

### D2: OSI-Approved License List via WebFetch

The agent retrieves the current OSI-approved license
list from opensource.org at invocation time rather than
maintaining a static embedded list.

**Rationale**: The OSI list is the authoritative source
per FR-003. Fetching it live ensures Pinkman always
uses the current list without requiring agent file
updates when OSI approves new licenses. Graceful
degradation: if the OSI site is unreachable, the agent
uses a well-known fallback set of commonly approved
licenses (MIT, Apache-2.0, BSD-2-Clause, BSD-3-Clause,
ISC, MPL-2.0, etc.) and notes the fallback in results.

### D3: Public Data Sources -- GitHub as Primary

GitHub's public web interface (via `webfetch`) is the
primary data source for project discovery, trend
metrics, and dependency manifest reading. Package
registries (pkg.go.dev, npmjs.com, crates.io) are
secondary sources for dependency resolution.

**Rationale**: GitHub hosts the majority of open source
projects and provides publicly accessible project
metadata (stars, forks, contributors, releases,
license files, dependency manifests). Using `webfetch`
avoids requiring API keys for basic operation. Rate
limiting is handled by the agent reporting partial
results when throttled (FR-012).

### D4: Report Storage and Dewey Integration

Reports are persisted locally as Markdown files at
`.uf/pinkman/reports/` and optionally stored in Dewey
via `dewey_store_learning` for cross-session retrieval.

**Rationale**: Local file storage ensures reports are
available even when Dewey is not deployed (Composability
First). Dewey integration enables semantic search across
past scouting results ("have we evaluated this project
before?"). The dual-storage approach follows the
graceful degradation pattern established by other agents
(Cobalt-Crush, Divisor personas).

### D5: Dependency Overlap as Post-Processing

Dependency overlap detection operates on the completed
result set after all projects have been scouted, rather
than during individual project analysis.

**Rationale**: Overlap is a cross-project property that
requires comparing dependency lists across all scouted
projects in a single invocation. Computing it during
individual project analysis would require either
multiple passes or mutable shared state, both of which
add complexity. Post-processing the complete result set
is simpler and produces the overlap matrix in a single
pass.

### D6: File Ownership Model

- `pinkman.md` (agent file): **user-owned** -- allows
  users to customize scouting behavior, data sources,
  and report format.
- `scout.md` (command file): **tool-owned** -- ensures
  the invocation interface stays canonical across
  scaffold updates.

**Rationale**: This follows the exact pattern used by
every prior agent + command pair (Specs 006, 007, 031).
User ownership for the agent enables customization of
scouting domains and report preferences. Tool ownership
for the command ensures consistent invocation.

## Coverage Strategy

### Unit Tests (Go)
- Scaffold drift detection: embedded asset at
  `internal/scaffold/assets/opencode/agents/pinkman.md`
  must match canonical `.opencode/agents/pinkman.md`
- Scaffold drift detection: embedded asset at
  `internal/scaffold/assets/opencode/command/scout.md`
  must match canonical `.opencode/command/scout.md`
- `expectedAssetPaths` count updated (35 → 37)
- `isToolOwned` returns false for `pinkman.md`, true
  for `scout.md`

### Integration Tests (Manual)
- Invoke `/scout "static analysis Go"` and verify:
  - Results contain only OSI-approved licensed projects
  - Each result includes direct dependency list
  - Shared dependencies are highlighted with counts
  - Trend indicators present (stars, forks, releases)
- Invoke `/scout --audit go.mod` and verify:
  - Dependencies listed with current vs. latest version
  - License changes detected between versions
  - Maintenance risk flags present where applicable
- Invoke `/scout --report <project-url>` and verify:
  - Report contains all required sections
  - Report stored at `.uf/pinkman/reports/`
  - Report can be stored in Dewey (if available)

### Acceptance Tests (Per Spec)
- SC-001: Time discovery invocation, verify < 60s
- SC-002: Cross-reference results against OSI list,
  verify zero false positives
- SC-003: Count trend indicators per project, verify ≥ 3
- SC-004: Compare audit results against known dependency
  updates, verify ≥ 95% detection rate
- SC-005: Verify report structure contains all sections
- SC-005a: Verify 100% shared dependency detection
- SC-006: Constitution check confirms no hero overlap

## Complexity Tracking

No constitution violations to justify. All four
principles pass cleanly.
<!-- scaffolded by uf vdev -->
