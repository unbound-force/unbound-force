---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-05T15:03:54Z
identity: review-dispatch-20261105T150354-jay-flowers
tier: draft
---

When a tool returns a hand-built projection of a schema-validated payload (e.g., read_dispatch_artifact's summary section), schema validation of the source does not pin the projection's shape: a field dropped from the projection passes every schema check. New projection-returning tools need an exhaustive shape pin in the same increment — a toMatchObject over every documented field (fixtures are deterministic, so values can be pinned too) plus an Object.keys exact-set assertion to catch additions. Repo precedent: bound-consolidate-dispatch-output's Layer 2 shape tests prescribe exactly this for detail: "summary"; PR #949's read-dispatch-artifact summary test initially pinned 6 of 11 fields, leaving mode, verdict_reason, run_counts, coverage, and workflow_result unpinned (feedback review 5777174866).
