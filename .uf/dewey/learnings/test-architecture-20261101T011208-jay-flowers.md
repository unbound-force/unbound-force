---
tag: test-architecture
author: jay-flowers
category: pattern
created_at: 2026-11-01T01:12:08Z
identity: test-architecture-20261101T011208-jay-flowers
tier: draft
---

When a spec moves a MUST's guarantee into a wiring layer by passing a pre-computed outcome into a pure function (e.g., resolveSelection taking a ClassifierOutcome computed by the caller), unit tests on the pure function can only prove the outcome is *ignored* — they cannot detect that the producing subsystem (classifier cascade, engine, network call) was *invoked anyway*. A MUST phrased as "X bypasses the cascade entirely" then has no failing-test vehicle unless the plan adds a wiring-level assertion: a spy on the injected spawn/fetch seams, or an audit-record field that distinguishes "not run" from "result ignored". Verified on the model-matrix-delegation spec review: the landed unit tests passed synthetic selected outcomes alongside explicit params and asserted pack: null, while the wiring task's wording ("resolve the selection through resolveSelection with the classifier cascade for implement/unknown kinds") permitted an eager implementation that would violate the 45s-latency guard with every enumerated test green. Rule of thumb: for every "subsystem must not run" MUST, require an enumerated test that observes the subsystem's invocation seam, not just the selection record's final values.
