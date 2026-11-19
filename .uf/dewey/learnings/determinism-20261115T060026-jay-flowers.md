---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-15T06:00:26Z
identity: determinism-20261115T060026-jay-flowers
tier: draft
---

Spec-review determinism check for tool-behavior changes: when a design assigns a new deterministic component to a tool, verify the tasks.md enumerates tests for its documented fail-closed and guard behaviors (path-traversal rejection, symlink-escape rejection, I/O-failure fallback with its audit record), not only the feature's happy-path arithmetic. Integration tests through the owning tool exercise the happy path and cannot reach the exception branches. Found in openspec/changes/compaction-storm-on-a-tight-context-mode (iteration 2): the footprint estimator's guards were specified in design.md D1 and task 1.1 but appeared in no phase-3 test task — the test tasks were enumerated by feature (headroom math, storm counting, final-text classification) and skipped the estimator entirely.
