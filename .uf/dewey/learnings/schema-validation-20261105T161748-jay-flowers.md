---
tag: schema-validation
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:17:48Z
identity: schema-validation-20261105T161748-jay-flowers
tier: draft
---

A cap declared on two parallel validation surfaces needs a boundary rejection test on each surface. On PR #949 (RD-FR-054), the payload.proposals 512 cap existed as Zod `.max(512)` in the plugin and `maxItems: 512` in schemas/review-dispatch/v1.0.0.schema.json, but the only >512-entry test exercised consolidate_dispatch's truncation layer (keep-first-512 plus overflow warning), which slices before either schema sees the array — so dropping either bound, or letting the two bounds drift apart, would have failed no test. Truncation-layer tests verify the graceful path, not schema enforcement; an N+1-element array asserted to fail parse on each surface is the drift alarm that keeps both representations of the bound verified and in sync.
