---
tag: documentation-drift
author: jay-flowers
category: gotcha
created_at: 2026-11-04T22:58:31Z
identity: documentation-drift-20261104T225831-jay-flowers
tier: draft
---

When a PR extends a registered JSON Schema payload with a new property, the registry directory's README contract summary is a silently-drifting documentation surface: the schema file, its .uf/schemas/ and internal/scaffold/assets mirrors, and parity tests all get updated, but no test enforces README-to-schema parity, so the prose enumeration of payload contents goes stale. Verified on unbound-force PR #949 (review-dispatch-consumer-bounds): schemas/review-dispatch/v1.0.0.schema.json gained the optional prop
