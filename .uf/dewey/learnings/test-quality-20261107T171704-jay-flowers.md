---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-07T17:17:04Z
identity: test-quality-20261107T171704-jay-flowers
tier: draft
---

Vitest tests over discriminated-union result envelopes ({status: "ok", data} | {status: "error", code, message}) must assert the status discriminant before any data assertion. Two masking forms to reject at review: (1) compound boolean expects with a falsy expected literal — expect(r.status === "ok" && r.data?.updated).toBe(false) passes both when the result is ok-with-updated:false and when it is any error envelope, because && short-circuits to false; (2) if (r.status === "ok") { expect(...) } guards with no preceding status assertion and no else branch — the data asserts are skipped and the test passes vacuously on error envelopes. Compound expects with truthy expected literals (toBe(sha), toBe(true), toBeNull()) do not mask today, but they collapse two assertions into one and reintroduce masking when copied with a falsy literal. Correct forms, both modeled in .opencode/test/uf-workflow.test.ts: expect(r.status).toBe("ok") before the narrowing guard, or close guards with else { expect.fail("...") }; split compound expects into separate status and data assertions, and prefer full-envelope toEqual(success({...})) where the whole payload is known. Review sweep: grep a test file for `.toBe(false)` compounds and for `if (result.status === "ok")` guards lacking an else-fail.
