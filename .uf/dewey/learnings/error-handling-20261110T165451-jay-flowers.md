---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-10T16:54:51Z
identity: error-handling-20261110T165451-jay-flowers
tier: draft
---

When a new tool claims to implement an established error-code contract, code-name parity is not contract parity: diff every observable failure signal (exit codes, stderr signatures, signal names) against the incumbent classifier's branch conditions. In unbound-force PR #1007, feedback-replies.ts classifyGhError maps gh exit code 4 to gh_auth_failed (retryable: false) while review-dispatch's ghFailureError (index.ts:8534) maps the same exit code 4 to rate_limited (retryable: true) — opposite codes, opposite retryability, one documented fetch-tool contract. Failure-stop logic and operator remediation branch on these values, so the divergence produces wrong guidance. Verification recipe for reviewers: grep the shared code names across plugins, then compare the condition lines above each return, exit code by exit code.
