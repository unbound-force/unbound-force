---
tag: scaffold-testing
author: jay-flowers
category: pattern
created_at: 2026-11-06T18:09:14Z
identity: scaffold-testing-20261106T180914-jay-flowers
tier: draft
---

When a preservation/merge behavior is wired into a deployment loop (scaffold.Run), unit tests of the pure helper do not protect the wiring. In unbound-force, the TestRun_*PreservedOnRerun family (constitution, reviewer capabilities, preflight checks, model matrix, override stubs) is the established integration pattern for refresh-preservation semantics: pre-populate the target file, run full Run(), assert the content survives byte-identical AND the disposition is Skipped rather than Updated. New preservation behaviors must get the same Run-level test — a pure-function test plus fresh-target Run tests leaves the glue condition untested, and for tool-owned files a wiring regression is a silent overwrite (file listed in result.Updated) that no unit test catches. Observed in stop-shipping-plugin-lock: preservePluginDevDependencies had exemplary unit coverage (including the repo round-trip) but zero integration coverage of its scaffold.go Run-loop call site.
