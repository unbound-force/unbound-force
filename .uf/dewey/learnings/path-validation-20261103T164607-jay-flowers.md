---
tag: path-validation
author: jay-flowers
category: pattern
created_at: 2026-11-03T16:46:07Z
identity: path-validation-20261103T164607-jay-flowers
tier: draft
---

When improving error taxonomy in a fail-closed path validator, preserve the read invariant rather than just the error codes: content is read only when realpath fully resolves the entire path inside the allowed roots (validateDiffPath returns null), and every error classification aborts before any read. Missing paths can then be classified safely by canonicalizing the nearest existing ancestor and joining the missing suffix (issue #932): dangling symlinks or unreadable segments in the missing suffix cannot change the classification: only the suffix's existence matters.
