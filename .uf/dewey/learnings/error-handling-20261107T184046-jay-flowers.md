---
tag: error-handling
author: jay-flowers
category: pattern
created_at: 2026-11-07T18:40:46Z
identity: error-handling-20261107T184046-jay-flowers
tier: draft
---

When classifying subprocess failures by matching the command's combined output, anchor the signature to the diagnostic prefix the tool actually emits (git's "fatal: not a git repository") rather than a bare substring, and echo the captured output in the surfaced message. Two failure modes otherwise appear: a target path that itself contains the signature phrase (git echoes the path in "fatal: cannot change to ..." errors) misclassifies a real failure as the benign skip; and a warning that reports only the exec error ("exit status 128") discards the stderr that names the actual cause, forcing the operator to re-run the command manually to diagnose. The output is already in hand at the classification point — use it for both the match and the message. Precedent: checkTrackedPluginLock in internal/doctor/checks.go (stop-shipping-plugin-lock follow-up, PR #957).
