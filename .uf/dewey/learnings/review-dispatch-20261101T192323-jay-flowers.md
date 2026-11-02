---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-01T19:23:23Z
identity: review-dispatch-20261101T192323-jay-flowers
tier: draft
---

When a regression test for a validation relaxation relies on a shared JSON fixture's implicit null default for the relaxed field (rather than assigning null explicitly in the test body), the test's coverage claim is coupled to fixture state it does not control. The pattern is legitimate — leaving the field untouched exercises the full null path including schema validation and artifact persistence — but it erodes silently: a later fixture edit that introduces a non-null default keeps the test green while removing coverage of exactly the relaxation the test was written to protect. Mitigation: either assign the relaxed field explicitly to null in the test body (payload state is identical after structuredClone), or add an inline comment naming the fixture file and line whose null default the test depends on, so a fixture edit surfaces the coupling. Observed reviewing accept-variantless-parent-model (PR #898): both variantless-acceptance tests in finalize-error-boundaries.test.ts depend on sample-review-dispatch.json:49.
