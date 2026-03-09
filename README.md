# Unbound Force

The meta/organizational repository for the [Unbound Force](https://github.com/unbound-force) organization -- a superhero-themed AI agent swarm for software engineering.

## What is Unbound Force?

Unbound Force is an organization of AI agent personas (heroes) that collaborate as a software development swarm. Each hero is a separate repository with a distinct role:

| Hero | Role | Status |
|------|------|--------|
| **Gaze** | Tester (Quality Sentinel) | Implemented |
| **Muti-Mind** | Product Owner (Vision Keeper) | Spec only |
| **Cobalt-Crush** | Developer (Engineering Core) | Spec only |
| **The Divisor** | PR Reviewer (Council) | Spec only |
| **Mx F** | Manager (Flow Facilitator) | Spec only |

## Constitution

This organization is governed by a [constitution](.specify/memory/constitution.md) (v1.0.0) that defines three core principles:

1. **Autonomous Collaboration** -- Heroes communicate through well-defined artifacts, not runtime coupling. Every hero completes its primary function independently.
2. **Composability First** -- Every hero is independently installable and usable alone. Combining heroes produces additive value without mandatory dependencies.
3. **Observable Quality** -- Every hero produces machine-parseable output (JSON minimum) with provenance metadata. Quality claims are backed by automated evidence.

All hero repositories must maintain constitutions that align with (and never contradict) these org-level principles.

## Specification Framework

This repo distributes a unified two-tier specification framework via the `unbound` CLI:

```bash
# Install
brew install unbound-force/tap/unbound

# Scaffold into any repository
unbound init
```

The framework provides:

- **Speckit** (strategic): Full 9-phase pipeline for architectural work (`/speckit.specify` through `/speckit.implement`)
- **OpenSpec** (tactical): Lightweight workflow for bug fixes and small changes (`/opsx:propose` through `/opsx:archive`)
- **Constitution governance bridge**: Every proposal includes alignment assessment against the three org principles

`unbound init` scaffolds 29 files into your repository: templates, scripts, commands, agents, and the custom `unbound-force` OpenSpec schema. User-owned files are skipped on re-run; tool-owned files are auto-updated when content changes.

See [AGENTS.md](AGENTS.md) for full workflow documentation and boundary guidelines.

## Repository Contents

This repo contains architectural design specs for all heroes and shared standards:

- **`specs/`** -- 10 architectural specifications organized in three phases
- **`cmd/unbound/`** -- Go CLI binary for framework distribution
- **`internal/scaffold/`** -- Scaffold engine with embedded assets
- **`.specify/memory/constitution.md`** -- The org constitution (highest authority)
- **`openspec/`** -- OpenSpec tactical workflow configuration and schema
- **`opencode.json`** -- MCP server configuration (knowledge graph via graphthulhu)
- **`unbound-force.md`** -- Hero descriptions and team vision
- **`AGENTS.md`** -- Development conventions and workflow guide

## Knowledge Graph

Project knowledge is indexed and queryable via [graphthulhu](https://github.com/skridlevsky/graphthulhu), an MCP-based knowledge graph server. Hero agents can search specs, traverse cross-references, and query document metadata without loading entire files into their context windows. See `specs/010-knowledge-graph-integration/` for the full specification.

See [AGENTS.md](AGENTS.md) for full project structure, spec organization, and development workflow.

## License

[Apache 2.0](LICENSE)
