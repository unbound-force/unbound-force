---
tag: risk-assessment
author: jay-flowers
category: gotcha
created_at: 2026-11-04T22:49:37Z
identity: risk-assessment-20261104T224937-jay-flowers
tier: draft
---

When a design documents compensating controls for a removed guarantee, enumerate every surface the change touches, not just the primary one. PR 957 removed the npm lock from both the scaffold activation path and `make plugin-test`; the design's mitigation list (exact pins, toml/uuid overrides, --ignore-scripts, provider-free probes) applies only to activation (--omit=dev, four direct packages), while plugin-test installs the full dev-dependency tree — the widest transitive surface — without --ignore-scripts
