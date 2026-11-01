---
tag: model-matrix
author: jay-flowers
category: pattern
created_at: 2026-11-01T13:46:38Z
identity: model-matrix-20261101T134638-jay-flowers
tier: draft
---

Pattern: when a spec requires explicit parameters to bypass an auxiliary engine (classifier, enricher, recommender) while the primary resolution function stays pure, place the bypass predicate INSIDE the engine's entry point rather than at call sites. In the model-matrix-delegation change (PR #894, .opencode/lib/model-matrix.ts), ClassifierCascade.classify() internally computes a run state via classifierRunState() — explicit model/profile/pack, a yielding agents pin, static plan/tasks kinds, classifier.enabled: false, or no configured matrix all resolve to not-run without touching either engine or the cache — so the plugin invokes cascade.classify() unconditionally before resolveSelection(). The spec's bypass requirement cannot be violated by a caller forgetting a guard, the not-run reason feeds the audit trail, and the cache is never consulted for bypassed calls. When reviewing cascade systems, verify the bypass lives in the engine, not the caller; a caller-side guard is an intent-drift risk because every new call site must remember it.
