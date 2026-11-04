---
tag: seam-verification
author: jay-flowers
category: pattern
created_at: 2026-11-04T15:16:42Z
identity: seam-verification-20261104T151642-jay-flowers
tier: draft
---

A reader tool added over an existing persisted artifact format needs a production writer→reader round-trip test (determinism pack DE-007 layer 3); compositional fixtures on each side do not close the seam. Hand-built envelope fixtures in the reader's tests plus a raw readFile+JSON.parse helper in the writer's tests both stay green when the envelope contract between production writer and production reader drifts (artifact_type value, filename convention, envelope keys only the writer emits) — payload-level drift is caught only when both sides share the payload schema, which masks the envelope-wrapper gap. The pattern: persist through the production writer in a scratch root, read back through the production reader, and assert the sections consumers depend on. The seam-verification change's "Dispatch artifact parse round-trip" scenario (finalize_review_dispatch → consumer parser) is the shipped precedent; PR 949's read_dispatch_artifact initially shipped with only compositional coverage, caught as review feedback.
