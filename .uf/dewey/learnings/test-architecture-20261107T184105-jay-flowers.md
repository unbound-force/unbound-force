---
tag: test-architecture
author: jay-flowers
category: pattern
created_at: 2026-11-07T18:41:05Z
identity: test-architecture-20261107T184105-jay-flowers
tier: draft
---

Composed Run()-level tests for merge-before-compare preservation behaviors must assert both dispositions: the refresh run (differing seed → asset reported Updated, merged content on disk) and the idempotent re-run (unchanged state → asset reported Skipped, bytes stable). The second run is the only orchestrator-level signal that the merge executes before the bytes.Equal skip comparison; reordering the merge after the comparison makes every run rewrite and re-report Updated while helper-level round-trip tests stay green, because the helper still produces byte-identical output in isolation. Precedent: TestRun_PreservesTargetDevDependenciesWhenUpdatingManifest (internal/scaffold) covers the Updated half for preservePluginDevDependencies with an older-pin seed, but the Skipped half — prescribed by the change's own retrospective learning ("asserts the deployed bytes plus Created/Updated/skip reporting") — was omitted, leaving the scaffold deploy loop's preserve-before-skip ordering (scaffold.go:299-301 vs the comparison at :331) verified only by code reading. The fix is one appended second Run() asserting the asset lands in result.Skipped and the file bytes are unchanged.
