---
tag: test-architecture
author: jay-flowers
category: gotcha
created_at: 2026-11-08T22:40:34Z
identity: test-architecture-20261108T224034-jay-flowers
tier: draft
---

An expect() placed inside a callback passed to the code under test (e.g., an injected io.err writer) silently never runs when the callback is never invoked; the test passes on its other assertions while the callback-mediated contract stays unverified. Always capture callback output into an array during the call and assert on the captured text afterwards, so a never-invoked callback fails the assertion instead of skipping it.
