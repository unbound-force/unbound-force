---
tag: quickfetch-failed-run-logs-test-sets-tmpdir-t
author: jay-flowers
category: gotcha
created_at: 2026-11-15T00:52:10Z
identity: quickfetch-failed-run-logs-test-sets-tmpdir-t-20261115T005210-jay-flowers
tier: draft
---

Learning from branch quick/fetch-failed-run-logs-test-sets-tmpdir-t (issue #1051), 2026-11-14, category gotcha, topic gaze-reporter subagent depth fallback. When /uf.unleash runs as a delegated child session (for example from /uf.run-pipeline-plan), the Step 6 Phase 1b gaze-reporter invocation through the Task tool fails with "Subagent depth limit reached (1)". The faithful fallback is to reproduce the gaze-reporter agent's full mode directly: run gaze crap --format=json ./..., gaze quality --format=json ./..., gaze analyze --classify --format=json ./..., and gaze docscan ./... via bash, then condense the JSON summaries (CRAP summary, quadrant counts, quality summary, baseline comparison) into the informational Gaze Report for the review dispatch. Also noted: the local .gaze/baseline.json can be stale enough that the crap baseline comparison reports most functions as new with several new-function violations — a pre-existing repo condition, not a change regression; gaze metrics for a branch that touches zero Go files equal the merge-base by construction.
