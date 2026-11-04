---
tag: spec-traceability
author: jay-flowers
category: pattern
created_at: 2026-11-04T14:52:10Z
identity: spec-traceability-20261104T145210-jay-flowers
tier: draft
---

Spec deltas that add THEN-clauses mandating exact error-message content need the same enumeration check as new error codes: before merging, grep each distinctive content token across the test suite — zero matches means the normative requirement has no enforcement and a later refactor can silently drop spec-mandated behavior. Verified on PR 949 feedback item 6 (unbound-force/unbound-force): the build-review-prompts delta required the diff_not_found message to quote the path and carry copy-don't-retype guidance; production emitted the guidance ('Copy the path exactly from the prompt or manifest; do not retype it.') while 'retype' returned zero matches in .opencode/test — build-review-prompts.test.ts asserted only data.code and read-diff-chunk.test.ts asserted path quoting but not the guidance. Extends the error-code enumeration pattern (learning/review-dispatch-20261031T025137-jay-flowers) from code discriminants to message-content clauses.
