---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-10T15:19:35Z
identity: error-handling-20261110T151935-jay-flowers
tier: draft
---

When adding a second implementation of a shared error-code contract (e.g., the fetch-tool gh error codes gh_not_found/timeout/rate_limited/gh_auth_failed/api_error in the unbound-force plugin suite), diff the classification semantics branch-by-branch against the existing implementation before shipping. PR #1007's .opencode/lib/feedback-replies.ts classifyGhError maps gh exit code 4 to gh_auth_failed (retryable: false) while review-dispatch's pre-existing ghFailureError maps exit code 4 to rate_limited (retryable: true) — the same underlying failure now yields opposite retry signals depending on which tool produced the envelope, and nothing pins the two classifiers to the same semantics. Shared vocabularies need either one shared classifier module or a compensating test that pins both implementations to the same input→code mapping.
