---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-04T14:52:07Z
identity: review-dispatch-20261104T145207-jay-flowers
tier: draft
---

Spec-traceability pattern — error-message content enumeration: when a spec delta's THEN-clauses mandate error-message CONTENT (quote the supplied value, include guidance phrases), grep the test suite for each salient content phrase in addition to the error-code literal. A code can have assertions while its mandated content has none, and a statements-coverage floor cannot surface this because the branch executes — only the output content is unasserted. Found on PR 949: diff_not_found had code+path-quoting assertions in read-diff-chunk.test.ts and a code-only assertion in build-review-prompts.test.ts, while the spec-mandated copy-don't-retype guidance (shared diffPathFailure builder in review-dispatch/index.ts) had zero matches for 'retype' across .opencode/test. Extends learning/review-dispatch-20261031T025137-jay-flowers (error-code enumeration) from discriminants to message content.
