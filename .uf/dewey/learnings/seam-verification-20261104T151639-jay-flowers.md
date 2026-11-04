---
tag: seam-verification
author: jay-flowers
category: gotcha
created_at: 2026-11-04T15:16:39Z
identity: seam-verification-20261104T151639-jay-flowers
tier: draft
---

A new consumer-reader tool over a producer-persisted artifact needs its own write→read round-trip test even when the payload contract is compile-time shared. PR 949's read_dispatch_artifact (unbound-force/unbound-force) validated payloads with the same zod schema its finalizer uses, yet shipped with only compositional coverage: reader tests hand-built minimal envelopes (omitting the production envelope's version/timestamp/schema_version/context) and writer tests parsed persisted bytes with a bare JSON.parse helper, while the only existing round-trip (finalize → parseReviewDispatchArtifact) pinned the resume consumer's lenient field subset — so persisted-bytes → strict-schema validation, the section service over real writer output, and the duplicated `<correlation_id>.json` filename convention had zero connecting coverage. Recipe: for every new reader over a persisted artifact, add one connecting test that drives the production writer (consolidate/finalize with provenance into a scratch root) and reads back through the production reader, asserting the served sections; and derive the increment's tasks.md seam-verification task from DE-007's three named layers (registration surface, documented output shape, round-trips) so a checked checkbox cannot mask an unverified layer.
