# Unbound Force

The meta/organizational repository for the [Unbound Force](https://github.com/unbound-force) organization -- a superhero-themed AI agent swarm for software engineering.

## What is Unbound Force?

Unbound Force is an organization of AI agent personas (heroes) that collaborate as a software development swarm. Each hero is a separate repository with a distinct role:

| Hero | Role | Status |
|------|------|--------|
| **Gaze** | Tester (Quality Sentinel) | Implemented |
| **Muti-Mind** | Product Owner (Vision Keeper) | Implemented |
| **Cobalt-Crush** | Developer (Engineering Core) | Implemented (embedded in `unbound-force`) |
| **The Divisor** | PR Reviewer (Council) | Implemented (embedded in `unbound-force`) |
| **Mx F** | Manager (Flow Facilitator) | Implemented (`mxf` CLI + coaching agent) |

## Constitution

This organization is governed by a [constitution](.specify/memory/constitution.md) that defines four core principles:

1. **Autonomous Collaboration** -- Heroes communicate through well-defined artifacts, not runtime coupling. Every hero completes its primary function independently.
2. **Composability First** -- Every hero is independently installable and usable alone. Combining heroes produces additive value without mandatory dependencies.
3. **Observable Quality** -- Every hero produces machine-parseable output (JSON minimum) with provenance metadata. Quality claims are backed by automated evidence.
4. **Testability** -- Every component MUST be testable in isolation without requiring external services or shared mutable state.

All hero repositories must maintain constitutions that align with (and never contradict) these org-level principles.

## Getting Started

### macOS (Homebrew)

```bash
brew install unbound-force/tap/unbound-force
```

### Fedora / RHEL (dnf)

```bash
sudo dnf install -y "$(
  curl -fsSL \
    https://api.github.com/repos/unbound-force/unbound-force/releases/latest |
  grep -o 'https://[^"]*linux_amd64\.rpm'
)"
```

See **[QUICKSTART.md](QUICKSTART.md)** for full installation instructions (macOS and Fedora/RHEL), first-use walkthrough, and platform-specific guidance. See **[Usage Guide](docs/usage.md)** for common workflows and command reference.

## Specification Framework

The framework provides:

- **Speckit** (strategic): Full 9-phase pipeline for architectural work (`/speckit.specify` through `/speckit.implement`)
- **OpenSpec** (tactical): Lightweight workflow for bug fixes and small changes (`/opsx-propose` through `/opsx-archive`)
- **Workflow orchestration**: Hero lifecycle commands (`/workflow start`, `/workflow status`, `/workflow list`, `/workflow advance`) for managing the 6-stage feature lifecycle
- **Constitution governance bridge**: Every proposal includes alignment assessment against the four org principles

`uf init` scaffolds 50 files into your repository: templates, scripts, commands, agents, Divisor review personas, convention packs, and the custom `unbound-force` OpenSpec schema. Use `uf init --divisor` to deploy only the PR review agents and convention packs. Use `--lang` to override language auto-detection for convention pack selection. User-owned files are skipped on re-run; tool-owned files are auto-updated when content changes.

See [AGENTS.md](AGENTS.md) for full workflow documentation and boundary guidelines. See [docs/architecture.md](docs/architecture.md) for how all components connect, [docs/cli-reference.md](docs/cli-reference.md) for CLI reference, and [docs/configuration.md](docs/configuration.md) for configuration guide.

## Repository Contents

This repo contains architectural design specs for all heroes and shared standards:

- **`specs/`** -- 37 architectural specifications covering constitution, hero architectures, swarm orchestration, tooling, and workflows
- **`images/`** -- Published sandbox images, including the FullSend OpenCode image
- **`cmd/unbound-force/`** -- Go CLI binary for framework distribution
- **`cmd/mutimind/`** -- Muti-Mind product owner backend CLI
- **`internal/`** -- Business logic packages (scaffold, sandbox, gateway, config, doctor, setup, orchestration, schemas, artifacts, backlog, sync, coaching, dashboard, impediment, metrics, sprint)
- **`.specify/memory/constitution.md`** -- The org constitution (highest authority)
- **`openspec/`** -- OpenSpec tactical workflow configuration and schema
- **`schemas/`** -- JSON Schema registry
- **`.opencode/`** -- 18 agents, ~40 commands, skills, convention packs
- **`docs/`** -- [Architecture](docs/architecture.md), [usage guide](docs/usage.md), [CLI reference](docs/cli-reference.md), [configuration](docs/configuration.md), [hero descriptions](docs/heroes.md)
- **`opencode.json`** -- MCP server configuration (Dewey, Replicator)
- **`AGENTS.md`** -- Development conventions and workflow guide

## Knowledge Layer

Project knowledge is indexed and queryable via [Dewey](https://github.com/unbound-force/dewey), a semantic knowledge layer that combines graph traversal with vector-based semantic search. Hero agents can search specs, find similar documents, traverse cross-references, and query document metadata via MCP tools without loading entire files into their context windows. See `specs/014-dewey-architecture/` for the design and `specs/015-dewey-integration/` for agent integration details.

See [AGENTS.md](AGENTS.md) for full project structure, spec organization, and development workflow.

## License

[Apache 2.0](LICENSE)
