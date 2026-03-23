# Implementation Plan: Dewey Integration

**Branch**: `015-dewey-integration` | **Date**: 2026-03-22 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/015-dewey-integration/spec.md`

## Summary

Integrate Dewey into the Unbound Force meta repo by
updating scaffold assets, hero agent persona files,
MCP configuration templates, and doctor/setup commands
to reference Dewey instead of graphthulhu. All hero
agents gain semantic search capabilities with a 3-tier
graceful degradation pattern. The Swarm plugin's default
embedding model is updated to match Dewey's
enterprise-grade model.

This is a cross-cutting change touching scaffold
assets (embedded in the CLI binary), live agent files,
Go source code (doctor checks, setup steps), and
living documentation (AGENTS.md).

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: `github.com/spf13/cobra`
(CLI), `github.com/charmbracelet/log` (logging),
`github.com/charmbracelet/lipgloss` (terminal styling)
**Storage**: N/A (configuration and documentation
changes)
**Testing**: Standard library `testing` package,
`go test -race -count=1`
**Target Platform**: macOS, Linux (CLI binary)
**Project Type**: CLI
**Performance Goals**: N/A (no runtime behavior changes
-- scaffold and agent file updates only)
**Constraints**: Backward compatible with projects that
have not installed Dewey. Graceful degradation is
mandatory (Constitution Principle II).
**Scale/Scope**: ~30 files changed across scaffold
assets, agent files, Go source, and documentation.
~15 scaffold asset files updated (tool references),
~10 live agent/command files updated, ~5 Go source
files changed (doctor, setup), ~2 documentation files
(AGENTS.md, opencode.json).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check
after Phase 1 design.*

### I. Autonomous Collaboration -- PASS

The integration updates agent files and scaffold
templates to reference Dewey tools. Dewey communicates
via MCP (well-defined protocol). No runtime coupling
is introduced. Agent files describe tool usage
patterns, not synchronous dependencies.

### II. Composability First -- PASS

The 3-tier degradation pattern (FR-007, FR-008) ensures
every hero works without Dewey. Dewey is an enhancement
that improves context quality, not a hard dependency.
The doctor check reports Dewey as optional with
"graph-only" and "file reads" fallback tiers.

### III. Observable Quality -- PASS

The doctor health check (FR-009) provides observable
status for Dewey components. Agent files produce the
same artifact formats regardless of Dewey availability.
No output format changes.

### IV. Testability -- PASS

All changes are testable: scaffold asset content can
be searched for stale references (regression test),
doctor checks use injected dependencies, agent files
are Markdown (content-searchable). No external services
needed for testing.

**Coverage strategy**: Update existing regression tests
(`TestScaffoldOutput_NoBareUnboundReferences` pattern)
to also check for stale `graphthulhu` and
`knowledge-graph_` references. Add doctor test
assertions for new Dewey health check group. Coverage
target: maintain existing coverage percentage.

## Project Structure

### Documentation (this feature)

```text
specs/015-dewey-integration/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── scaffold-changes.md
└── checklists/
    └── requirements.md  # Already created
```

### Source Code (repository root)

```text
opencode.json                          # MCP config: dewey replaces graphthulhu

.opencode/agents/
├── muti-mind-po.md                    # Dewey tool usage + fallback
├── cobalt-crush-dev.md                # Dewey tool usage + fallback
├── gaze-reporter.md                   # Dewey tool usage + fallback
├── divisor-*.md (5 files)             # Dewey tool usage + fallback
├── mx-f-coach.md                      # Dewey tool usage + fallback
└── constitution-check.md              # Dewey tool usage + fallback

internal/scaffold/assets/
├── opencode.json                      # Scaffold template: dewey config
├── opencode/agents/*.md               # Scaffold agent templates
└── ...                                # Other scaffold assets with tool refs

internal/doctor/
└── checks.go                          # New Dewey health check group

internal/setup/
└── setup.go                           # New Dewey installation step

AGENTS.md                              # Updated references
```

**Structure Decision**: No new packages or directories.
All changes are within existing files and the existing
scaffold asset structure. The doctor and setup changes
add new check functions and setup steps to existing
packages.

## Complexity Tracking

No constitution violations. No complexity justifications
needed.
