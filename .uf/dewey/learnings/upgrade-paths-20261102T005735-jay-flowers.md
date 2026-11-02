---
tag: upgrade-paths
author: jay-flowers
category: gotcha
created_at: 2026-11-02T00:57:35Z
identity: upgrade-paths-20261102T005735-jay-flowers
tier: draft
---

An unpinned installer is an implicit distribution channel for a new major version. During the OpenCode v1→v2 org-wide migration (unbound-force/unbound-force#797), the umbrella identifies the hard sequencing constraint: `uf setup` installs the latest opencode unpinned, so any repo running setup receives the v2 binary immediately — before dual-compatibility work lands. Major-runtime migration plans MUST audit every install-latest path first and treat it as a forced edge in the migration sequencing graph: compatibility work (plugin port, scaffold dual-emission) must land before any promotion of the new runtime, because the unpinned installer promotes it automatically the moment upstream ships it.
