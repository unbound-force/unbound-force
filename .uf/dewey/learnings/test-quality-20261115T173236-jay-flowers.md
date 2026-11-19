---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-15T17:32:36Z
identity: test-quality-20261115T173236-jay-flowers
tier: draft
---

Error-path tests for a plugin tool's structured error envelope must pin all four documented fields — error, code, message, retryable — not just error: true. The envelope is the tool's documented output shape that instruction text consumes for degradation routing (/uf.review-pr Step 8a skips the duplicate-review gate on a structured error envelope), so a wrong code or message silently misroutes while tests stay green. Pin static messages exactly and prefix/contains-match dynamic ones (stderr- and validator-derived). Gotcha: the fallback code passed to classifyError is not always the emitted code — ghFailureError annotates api_error, which wins over the fallback, so pr_not_found fires only on unparseable metadata; pin observed behavior, not the assumed fallback (PR #1077, feedback item 2).
