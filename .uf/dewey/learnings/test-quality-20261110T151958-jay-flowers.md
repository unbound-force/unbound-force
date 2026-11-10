---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-10T15:19:58Z
identity: test-quality-20261110T151958-jay-flowers
tier: draft
---

Fail-closed fake-executor route tables implicitly pin negative behavior in tool tests. A fake gh (or git) executor that throws "unexpected call" on any invocation it has no scripted route for turns every "this call must never happen" into an implicit assertion: resolving a REJECT decision's thread, re-filing a deduplicated issue, or re-posting a cached reply all fail the test without dedicated negative routes. Observed on PR #1007 (.opencode/test/feedback-replies.test.ts): reject-stays-open is pinned not by an absence assertion but because a resolveReviewThread call for the reject thread would find no route and fail. Pair the pattern with one or two explicit absence assertions (calls.some(...) === false) where the guarantee is spec-critical, so the intent is readable without archaeology into the route table.
