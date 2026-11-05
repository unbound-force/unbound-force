---
tag: uf-workflow-testing
author: jay-flowers
category: gotcha
created_at: 2026-11-05T18:35:05Z
identity: uf-workflow-testing-20261105T183505-jay-flowers
tier: draft
---

Tool-envelope vitest tests narrow results with `if (result.status === "ok") { expect(...) }` for TypeScript type narrowing, but without a preceding `expect(result.status).toBe("ok")` or an `else { expect.fail(...) }` the test passes vacuously when the tool returns a failure envelope — vitest does not fail zero-assertion tests. The same trap hits compound assertions: `expect(result.status === "ok" && result.data?.flag).toBe(false)` evaluates to the expected `false` on an error status, so (error) and (ok, flag=false) are indistinguishable. Observed in PR #948 round 2 (.opencode/test/uf-workflow.test.ts:1610, :1746, :1867): the `updated: false` up-to-date contract of the push tools had no fail-closed coverage until review flagged it. Fix pattern already used in-file: assert status first, then narrow; or split compounds into a status assertion followed by a field assertion.
