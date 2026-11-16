---
tag: uf-unleash-orchestration
author: jay-flowers
category: gotcha
created_at: 2026-11-16T01:36:54Z
identity: uf-unleash-orchestration-20261116T013654-jay-flowers
tier: draft
---

When a council-approved tasks.md deliberately sequences production-code migration ahead of test migration (as in compaction-storm-detector-sums-compactio, issue #1125, November 2026), mid-pipeline /uf.unleash phase checkpoints legitimately fail with change-caused interim failures whose remediation IS the remaining tasks (14 failing rows at phase 1, 6 at phase 2, 0 at phase 3 after the test-migration tasks landed). The orchestrator should record the honest failed verdict with the classified inventory plus an explicit deferral note, continue the phase loop through the scoped per-task delegations, and enforce green at the later phase checkpoints — entering the build/test fix loop mid-graph would duplicate already-planned, already-classified work into an unscoped blob and erode per-task bookkeeping. The later checkpoints (and the change's own final verification task) provide the same enforcement guarantee with the full fix-loop budget. On resume, the deferral note in the phase checkpoint data must prevent re-entering a fix loop for the recorded failure.
