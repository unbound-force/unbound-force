# Implementation Plan: [FEATURE]

**Branch**: `[###-feature-name]` | **Date**: [DATE] | **Spec**: [link]
**Input**: Feature specification from `/specs/[###-feature-name]/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Muti-Mind is the Product Owner hero, responsible for maintaining the product backlog, prioritizing work, and serving as the acceptance authority. It is implemented as a hybrid architecture: a Go-based CLI backend (`cmd/mutimind`) handles robust data manipulation and GitHub synchronization, while the AI capabilities (prioritization, story generation, acceptance review) are delegated to OpenCode commands and agents. Backlog items are stored locally as Markdown files with YAML frontmatter, enabling the `graphthulhu` MCP server to natively index them as a knowledge graph for semantic search and context retrieval by the Swarm.

## Technical Context

**Language/Version**: Go 1.24+ (CLI backend), OpenCode Agents (AI runtime)
**Primary Dependencies**: `github.com/spf13/cobra`, `github.com/charmbracelet/log`, OpenCode Runtime, `graphthulhu` MCP Server
**Storage**: Local Markdown files with YAML frontmatter in `.muti-mind/backlog/`
**Testing**: Go standard library `testing`, OpenCode command functional tests. **Coverage Strategy**: 80% global unit test coverage minimum, 90% unit test coverage for `internal/backlog` and `internal/artifacts` parsing logic. Integration tests required for all GitHub API interactions in `internal/sync`. Functional tests via OpenCode scenarios for agent interactions.
**Target Platform**: CLI / OpenCode runtime
**Project Type**: CLI Tool + Agent Persona
**Performance Goals**: Sub-second local data manipulation; graceful pagination for MCP knowledge graph queries
**Constraints**: The OpenCode agent MUST exclusively use MCP tools for all read operations (hard dependency), reserving the CLI solely for write/sync operations. GitHub Issues/Projects is the ultimate source of truth for priority.
**Scale/Scope**: Project-level backlog management, single repository scope

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Autonomous Collaboration**: PASS. Muti-Mind produces and consumes standard artifact envelopes (e.g., `backlog-item`, `acceptance-decision`) without synchronous blocking.
- **II. Composability First**: PASS. Muti-Mind can be used standalone via OpenCode commands or the `mutimind` CLI without requiring other heroes.
- **III. Observable Quality**: PASS. Output commands support `--format json` alongside human-readable text. Artifacts include provenance metadata.
- **IV. Testability**: PASS. The Go backend enforces test coverage ratchets. OpenCode commands are testable via standard scenarios.

## Project Structure

### Documentation (this feature)

```text
specs/004-muti-mind-architecture/
├── spec.md              # Feature specification
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
cmd/mutimind/
└── main.go              # Go CLI entry point for backlog manipulation

internal/
├── backlog/             # Backlog MD parsing/storage logic
├── sync/                # GitHub CLI/API sync logic
└── artifacts/           # JSON artifact generation

.opencode/
├── agents/
│   └── muti-mind-po.md  # The Product Owner AI persona
└── command/
    ├── muti-mind.backlog-*.md
    ├── muti-mind.sync-*.md
    ├── muti-mind.prioritize.md
    └── muti-mind.generate-stories.md

schemas/hero-manifest/
└── muti-mind-hero.json  # Hero manifest declaring MCP dependencies
```

**Structure Decision**: The project uses a hybrid layout. The Go application logic lives under `cmd/mutimind` and `internal/` to keep it strictly separated from other tools in the monorepo. The user interface and AI delegation logic live inside `.opencode/` as native commands and agents, aligned with the rest of the Unbound Force ecosystem.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
