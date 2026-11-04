---
tag: seam-verification
author: jay-flowers
category: gotcha
created_at: 2026-11-04T15:16:36Z
identity: seam-verification-20261104T151636-jay-flowers
tier: draft
---

A strict-schema reader verified only against hand-built fixture envelopes is not pinned to the production writer: two green suites — reader tests on hand-built envelopes plus writer tests asserting persisted fields through a raw JSON.parse helper — can coexist with a broken writer→reader contract, because the raw-parse helper bypasses the reader's strict validation and the hand-built envelope is a subset of the writer's real envelope. Verified on unbound-force PR 949 (read-dispatch-artifact.test.ts vs consolidate-dispatch.test.ts, reviewer feedback thread PRRT_kwDORVoNbc6zVoHJ): DE-007's write→read round-trip was the only one of the three seam layers absent from the increment. Recipe: for every new consumer of a persisted artifact, add one test that goes through the production writer (e.g., the consolidate/finalize tool with provenance into a scratch root) and then the production reader, asserting the sections consumers depend on. Field-level assertions through a raw-parse helper are write-path verification, not round-trip verification — and a tasks.md seam-verification task scoped by assertion surface ('with tests') instead of DE-007's three named layers lets the checkbox read [x] while the round-trip layer stays missing.
