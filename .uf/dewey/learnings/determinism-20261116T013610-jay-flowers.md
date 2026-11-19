---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-16T01:36:10Z
identity: determinism-20261116T013610-jay-flowers
tier: draft
---

When a tool exposes mutable registry state through more than one read path — a live fold result and a persisted read-back (e.g. delegation-budget.ts's foldProbeObservations return and registeredProbeCounters) — route both through a single snapshot function (snapshotCounters) that assembles the exposed shape and any derived fields (tie-breaks like stormSessionID). Divergent reader views of the same underlying state are a determinism defect class: two consumers can observe different shapes or different tie-break resolutions for identical state. The compaction-storm-detector-sums-compactio change (issue #1125) shipped this pattern: both read paths expose one shape and one observation-order tie-break, and the tests assert equality against both paths.
