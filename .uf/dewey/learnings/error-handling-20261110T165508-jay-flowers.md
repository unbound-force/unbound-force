---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-10T16:55:08Z
identity: error-handling-20261110T165508-jay-flowers
tier: draft
---

gh error-classifier divergence gotcha (PR #1007, unbound-force): `gh api` surfaces HTTP 4xx failures as process exit code 4, which is overloaded — it covers both auth failures (401/403) and rate limits (429). Two classifiers in the same plugin suite mapped exit 4 oppositely: review-dispatch/index.ts (exit 4 → rate_limited, retryable) and lib/feedback-replies.ts classifyGhError (exit 4 → gh_auth_failed, not retryable, with a comment asserting exit 4 means auth). A rate limit classified as gh_auth_failed reports the wrong remediation and suppresses retry logic keyed on retryable. When writing a new gh error classifier: key on stderr text first ("rate limit", "gh auth login") and treat exit code as a fallback only, and prefer extracting one shared classifier both plugins import rather than copying per-plugin.
