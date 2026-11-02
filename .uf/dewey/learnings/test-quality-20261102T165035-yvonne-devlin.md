---
tag: test-quality
author: yvonne-devlin
category: pattern
created_at: 2026-11-02T16:50:35Z
identity: test-quality-20261102T165035-yvonne-devlin
tier: draft
---

A resolution semantic applied to some checks but not all is invisible to per-check test suites: each check's tests pass in single-change fixtures, so cross-check inconsistency only surfaces when a full phase runs against a multi-change fixture. In the uf-gate review (PR #582), checkArtifactExists and findTasksFile got branch-aware strict resolution plus multi-change regression tests, while checkCoverageStrategy kept an unrestricted glob across openspec/changes/* and specs/* — an unrelated change's coverage keywords satisfied the active change's Constitution IV gate, undetected at 93.7% statement coverage because no test ran a phase-level or cross-check comparison. When a change introduces a shared resolution rule (branch-awareness, ordering, scoping), extend the established multi-change test pattern to every glob that feeds a check, and add at least one phase-level integration test against a multi-change repository so resolution semantics are compared across checks, not just verified per check.
