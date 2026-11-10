---
tag: openspec-task-planning
author: jay-flowers
category: gotcha
created_at: 2026-11-10T15:21:19Z
identity: openspec-task-planning-20261110T152119-jay-flowers
tier: draft
---

OpenSpec tasks.md [P] markers must be validated against the delivered file layout, not the task descriptions: in merge-plugin-runtime-deps, tasks 3.1 (table-driven unit tests) and 3.2 (acceptance test) both carried [P] as independent deliverables, but both landed in the same new file (internal/scaffold/plugin_manifest_merge_test.go) and 3.2 reused 3.1's helpers (loadShippedPluginManifest, decodePluginManifest) — violating MP-017's different-files and no-prior-dependency clauses. When an acceptance test reuses a unit-test file's helpers, either place it in its own file, drop [P] from the dependent task, or hoist the shared helpers into a pre-task.
