---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-06T20:02:56Z
identity: documentation-pipeline-20261106T200256-jay-flowers
tier: draft
---

A dependency pin bump that updates every hand-maintained documentation surface (both manifests, AGENTS.md Technology Stack, CHANGELOG, spec delta, provenance record) can still leave stale version literals in tool-generated compiled knowledge articles (.uf/dewey/compiled/*.md), which restate the pin set as current-state fact and are served to future agents by Dewey semantic search. The old-version-string grep that validates a bump's documentation sweep should include .uf/dewey/compiled/ in its path set, or the bump should be followed by an incremental dewey_compile so the compiled rollups absorb the versions carried by the new learnings. Verified on unbound-force PR #957: the zod 4.1.8→4.6.6 sweep updated all primary surfaces, but .uf/dewey/compiled/plugin-architecture.md line 75 still asserts zod 4.1.8.
