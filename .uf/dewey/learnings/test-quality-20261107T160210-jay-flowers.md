---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:02:10Z
identity: test-quality-20261107T160210-jay-flowers
tier: draft
---

In TypeScript tool-result tests, `expect(result.status === "ok" && result.data?.field).toBe(false)` is a false-pass assertion: when the result is an error envelope the boolean short-circuits to false and matches `.toBe(false)`, so an error-regression passes silently. The same masking applies to `if (result.status === "ok") { ... }` blocks with no preceding `expect(result.status).toBe("ok")` and no `else expect.fail(...)`. Observed on PR #948's push-tool tests (.opencode/test/uf-workflow.test.ts): the up-to-date contract (success with updated:false) was guarded only by these weak forms while neighboring tests used the strong `expect(result).toEqual(success({...}))` form. Fix pattern: assert status explicitly before conditional data assertions, split compound boolean assertions into separate status and field expects, and prefer toEqual(success(...)) for full-envelope contracts.
