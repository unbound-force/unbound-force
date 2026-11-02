---
tag: error-handling
author: jay-flowers
category: pattern
created_at: 2026-11-02T17:02:20Z
identity: error-handling-20261102T170220-jay-flowers
tier: draft
---

When consolidating duplicated error-handling catch blocks into a shared classification helper (like hostResolutionFailureResult in review-dispatch), the explicit sanitizeInvocationError() wrapper at the old call sites becomes redundant ONLY because failedResult() in .opencode/lib/agent-executor.ts applies sanitizeInvocationError() internally to every message (L184). This makes failedResult the single sanitization chokepoint for all dispatch error results. Before dropping an explicit sanitize call during such an extraction, verify the downstream constructor sanitizes internally, and pin the property with redaction assertions in tests (Bearer value, URL query string, home path must not survive) driven through the new shared helper on every branch that uses it. The split-host-resolution-failure-classes change (RD-FR-038) did both: the transient test throws an error carrying a fake Bearer token, query string, and home path, and asserts the sanitized markers appear in the surfaced message.
