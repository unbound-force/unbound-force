---
tag: uf-workflow-testing
author: jay-flowers
category: gotcha
created_at: 2026-11-06T23:58:52Z
identity: uf-workflow-testing-20261106T235852-jay-flowers
tier: draft
---

In TypeScript tool-result tests, two assertion shapes silently pass when the tool returns an error result: (1) `if (result.status === "ok") { expect(result.data)... }` with no else branch — vitest passes zero-assertion tests when expect.hasAssertions is not configured, so the test cannot fail on the error path; and (2) expected-false compound assertions like `expect(result.status === "ok" && result.data?.updated).toBe(false)` — the compound short-circuits to `false` on an error result and coincidentally matches the expected value. Observed on PR #948 (uf-workflow push tools): every assertion of the no-op contract (`updated: false`) could not fail on an error result, leaving a documented contract with no fail-proof coverage. Fix pattern: assert `expect(result.status).toBe("ok")` before the narrowing guard (or add `else { expect.fail("expected success") }`, the file's own established pattern), and split expected-false compound assertions into an explicit status check plus a separate value check so the error path fails loudly.
