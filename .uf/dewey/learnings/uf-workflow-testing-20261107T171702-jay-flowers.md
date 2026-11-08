---
tag: uf-workflow-testing
author: jay-flowers
category: gotcha
created_at: 2026-11-07T17:17:02Z
identity: uf-workflow-testing-20261107T171702-jay-flowers
tier: draft
---

Result-envelope assertions that fold the status check into the data check pass vacuously on error envelopes. In vitest over `{ status: "ok", data } | { status: "error", code, message }` tool results, `expect(result.status === "ok" && result.data?.flag).toBe(false)` short-circuits to `false` when the result is an error envelope and satisfies the expectation — the suite stays green when the tool regresses from success to failure. The same masking occurs when data assertions sit inside `if (result.status === "ok")` with no preceding status assertion and no `else expect.fail(...)`; a test whose entire body is such a gate asserts nothing on the error path. Discipline that prevents it: assert the status envelope explicitly first, then assert each data field as a separate expectation, and give every conditional assertion branch an else that fails with context. Found in the uf-workflow push-tool contract tests (PR #948, thread PRRT_kwDORVoNbc604Lr3); adjacent tests in the same file already model both correct forms.
