---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-10T15:19:27Z
identity: error-handling-20261110T151927-jay-flowers
tier: draft
---

When a tool retries a non-idempotent create (gh issue create, POST endpoints) and records success only after the response parses, align the retry predicate with the "mutation may have landed" test: timeout-classified failures are outcome-unknown (a retry duplicates the side effect), while rate-limit and explicit HTTP rejections are safe to retry. Observed in unbound-force PR #1007 (.opencode/lib/feedback-replies.ts createIssue): the same module refuses in-call retry for reply posts and refuses retry on unparseable create output for exactly this reason (design.md D3: "a timed-out POST may have landed"), yet createIssue's loop retries timeout-classified failures blindly, reintroducing the duplicate-issue window the rest of the design closes. Fix pattern: retry only when the failure proves the request was rejected, or search for an existing artifact (e.g., issue by title) before retrying.
