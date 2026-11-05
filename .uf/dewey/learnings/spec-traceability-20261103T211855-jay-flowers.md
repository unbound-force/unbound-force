---
tag: spec-traceability
author: jay-flowers
category: pattern
created_at: 2026-11-03T21:18:55Z
identity: spec-traceability-20261103T211855-jay-flowers
tier: draft
---

When a documentation-only change records existing tool behavior as spec acceptance scenarios (e.g., OpenSpec delta edits for already-shipped error codes), enumerate every new THEN-clause and grep the test suite for each observable claim it makes (error code, message content, side-effect guarantees). Clauses with zero assertion matches become traceability debt introduced by the doc change itself: the spec then demands behavior no test pins, so a later refactor can silently drop it while the archived spec sti
