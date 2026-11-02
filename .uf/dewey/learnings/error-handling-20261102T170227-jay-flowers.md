---
tag: error-handling
author: jay-flowers
category: pattern
created_at: 2026-11-02T17:02:27Z
identity: error-handling-20261102T170227-jay-flowers
tier: draft
---

In the unbound-force shared plugin library, failedResult() (.opencode/lib/agent-executor.ts, error: { code, message: sanitizeInvocationError(message), retryable }) is the single sanitization seam for tool-result errors. Explicit sanitizeInvocationError(...) wrappers around text passed to failedResult() are redundant double-passes (the redaction regexes are idempotent), and catch blocks re-routed through a shared classification helper can safely pass raw errorText() as long as the final construction goes through failedResult(). When reviewing such a re-route, verify the seam directly (read the constructor) instead of trusting the helper's doc comment — the spec's "single sanitizeInvocationError() pass" requirement is satisfied at the constructor, not at the call site. Verified in split-host-resolution-failure-classes (RD-FR-038): dropping the explicit wrapper at the host-resolution catch seam preserved the sanitized transient form exactly (Bearer/query-string/home-path redaction pinned by tests).
