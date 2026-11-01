---
tag: content-pipeline
author: jay-flowers
category: pattern
created_at: 2026-10-31T21:32:19Z
identity: content-pipeline-20261031T213219-jay-flowers
tier: draft
---

Spec-change documentation phases scoped to docs/configuration.md and CHANGELOG.md miss AGENTS.md updates for structural additions: new tracked .uf config files, schema registry directories, and repo-local convention packs all change AGENTS.md's Project Structure or Convention Packs content. The Convention Packs list in AGENTS.md is the discovery surface for repo-local governance packs (the determinism.md precedent): a pack authored without a list entry exists on disk but is undiscoverable to agents that read AGENTS.md to determine applicable packs. Run the full documentation-gate triad (CHANGELOG / AGENTS.md / README) against the change's structural surface — every new tracked path and every new pack file — not just the user-facing configuration docs.
