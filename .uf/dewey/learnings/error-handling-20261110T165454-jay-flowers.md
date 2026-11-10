---
tag: error-handling
author: jay-flowers
category: pattern
created_at: 2026-11-10T16:54:54Z
identity: error-handling-20261110T165454-jay-flowers
tier: draft
---

Retrying a non-idempotent mutation on timeout duplicates the server-side effect: a timeout means the request's fate is unknown, so the server may already have acted. The same ambiguity as unparseable-success-output, which feedback-replies.ts createIssue already handles conservatively (no retry when no issue URL prints, "the issue may exist"), yet the same loop retries on timeout-classified failures — a timed-out gh issue create that landed server-side gets filed twice, and the second result silently wins the cache. Rule: for non-idempotent creates, retry only failure classes that provably precede the server-side effect (executable missing, auth failure, connection refused); treat timeout exactly like unparseable output — fail with a check-before-re-running hint and let the re-run be the retry. The cache dedup protects re-runs, never the in-call retry.
