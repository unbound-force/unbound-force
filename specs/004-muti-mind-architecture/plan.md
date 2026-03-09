# Implementation Plan: Muti-Mind Architecture (Product Owner)

**Branch**: `004-muti-mind-architecture` | **Date**: 2026-03-09 | **Spec**: [[specs/004-muti-mind-architecture/spec]]
**Input**: Feature specification from `/specs/004-muti-mind-architecture/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Design the architecture for Muti-Mind, the Product Owner hero. Muti-Mind acts as the Vision Keeper and Prioritization Engine, providing an AI persona, backlog management via OpenCode, GitHub Issues/Projects synchronization, and speckit pipeline integration for driving specifications and serving as the acceptance authority.

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Go 1.24+ (for tooling/MCP if any, though primarily OpenCode agents/commands)
**Primary Dependencies**: OpenCode runtime, GitHub CLI (`gh`) or GitHub API
**Storage**: Local Markdown files (YAML frontmatter) in `.muti-mind/backlog/` indexed by graphthulhu
**Testing**: Go standard library `testing` (if Go code), OpenCode simulation testing
**Target Platform**: OpenCode runtime environment (Swarm)
**Project Type**: AI Agent Persona / OpenCode Commands / MCP Server
**Performance Goals**: Fast local file operations, transparent LLM delegation, non-blocking GitHub sync
**Constraints**: Fully compatible with Hero Interface Contract (Spec 002) and Org Constitution
**Scale/Scope**: Single project backlog management, inter-agent coordination within the Swarm

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Autonomous Collaboration**: PASS - Muti-Mind collaborates via well-defined artifacts (`backlog-item`, `acceptance-decision`) and integrates via OpenCode commands rather than tight coupling.
- **II. Composability First**: PASS - Muti-Mind can be deployed independently for backlog management without requiring other heroes (though value increases with Gaze/Divisor).
- **III. Observable Quality**: PASS - Produces machine-parseable JSON artifacts (acceptance decisions) alongside human-readable markdown files.
- **IV. Testability**: PASS - Functional requirements define clear inputs (Gaze reports, user prompts) and observable outputs (acceptance decisions, updated MD files).

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
.muti-mind/
├── backlog/                 # Local MD files for backlog items (indexed by graphthulhu)
│   ├── BI-001.md
│   └── BI-002.md
├── config.yaml              # Muti-Mind specific configuration

.opencode/
├── agents/
│   ├── muti-mind-po.md      # The AI persona and decision framework
├── command/
│   ├── muti-mind.init.md
│   ├── muti-mind.backlog-add.md
│   ├── muti-mind.backlog-list.md
│   ├── muti-mind.backlog-update.md
│   ├── muti-mind.backlog-show.md
│   ├── muti-mind.sync.md
│   ├── muti-mind.sync-push.md
│   ├── muti-mind.sync-pull.md
│   ├── muti-mind.sync-project.md
│   ├── muti-mind.sync-status.md
│   ├── muti-mind.prioritize.md
│   └── muti-mind.generate-stories.md

cmd/mutimind/                # (Optional) Go binary if complex logic requires MCP/CLI backend
└── main.go
```

**Structure Decision**: Muti-Mind is primarily an OpenCode-native application. The core logic resides in `.opencode/agents/` and `.opencode/command/`. Data is stored in `.muti-mind/backlog/` as Markdown files to ensure graphthulhu indexing. If GitHub sync or complex operations exceed bash scripting limits within OpenCode commands, a Go binary under `cmd/mutimind/` may be introduced as a backend/MCP server.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
<!-- scaffolded by unbound vdev -->
