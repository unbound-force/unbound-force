---
tag: uf-unleash-orchestration
author: jay-flowers
category: gotcha
created_at: 2026-11-16T01:37:08Z
identity: uf-unleash-orchestration-20261116T013708-jay-flowers
tier: draft
---

Two tooling failure modes hit during the compaction-storm-detector-sums-compactio /uf.unleash run (2026-11-15/16) with clean workarounds. (1) The gaze-reporter agent cannot be invoked via the Task tool from a parented child session — subagent depth limit 1 — so /uf.review-council Phase 1b ran the gaze-reporter agent procedure's exact commands directly (gaze crap/quality/analyze --format=json on the relevant package) and captured the data; Phase 1b is informational either way. (2) The model-matrix classifier cascade can resolve a delegation to an unreachable provider endpoint (models.corp.redhat.com connection failure, reported as a retryable DELEGATE_FAILED child_message_error); retrying the delegation with an explicit model — the session's proven working model — bypasses the broken resolution and succeeds. Retry-with-explicit-model, not retry-as-is, is the fix for endpoint-resolution failures.
