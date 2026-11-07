---
tag: dewey-learnings
author: jay-flowers
category: pattern
created_at: 2026-11-07T19:56:56Z
identity: dewey-learnings-20261107T195656-jay-flowers
tier: draft
---

When a fix ships alongside its retrospective learning in the same commit series, audit the fix's inline claims against the learning's claims before approving — a contradiction means one record is wrong. Precedent: PR #957 round 3 shipped checkTrackedPluginLock (internal/doctor/checks.go) with a comment asserting the full-phrase match "keeps a TargetDir path that happens to contain the phrase from masquerading as the skip," while the co-shipped learning (error-handling-20261107T190356) states the anchoring "narrowed the masquerade its comment claims to close but did not eliminate it." When the learning is the accurate record, either soften the code comment to match the true residual risk or implement the learning's robust form (line-anchored prefix match, or exit-code plus first-stderr-line classification) in the same increment. General audit rule: grep each new learning that names a concrete remedy and verify either the remedy's artifact or an explicit deferral record; a learning that contradicts its own fix's documentation is an unshipped gap wearing a shipped fix's comment.
