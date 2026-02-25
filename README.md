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

## Repository Contents

This repo contains architectural design specs for all heroes and shared standards:

- **`specs/`** -- 9 architectural specifications organized in three phases
- **`.specify/memory/constitution.md`** -- The org constitution (highest authority)
- **`unbound-force.md`** -- Hero descriptions and team vision
- **`AGENTS.md`** -- Development conventions and workflow guide

See [AGENTS.md](AGENTS.md) for full project structure, spec organization, and development workflow.

## License

[Apache 2.0](LICENSE)
