---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:18:35Z
identity: issue-triage-20261102T011835-jay-flowers
tier: draft
---

Advisory-derived issues can be resolved by later iterations of the same fix loop before triage runs. When triaging an issue that originated from a review-council advisory on a parent change, grep the current tree for the gate, outcome, or behavior the issue requests before classifying it as actionable work. Issue #771 (verify replacement run-file persistence before reporting redispatched) was filed from an iteration-2 divisor-guard advisory on fix-review-dispatch-recovery; by triage time, PR #776 had already landed the persisted-flag gate (write-temp-then-rename in executeDispatchRun), the restore-on-!persisted path in redispatch_failed_runs, and the persistence-failure test in dispatch-recovery.test.ts — verified in both the active plugin and the byte-identical scaffold twin (the scaffold parity obligation means the twin only carries merged canonical work). Triage that skips the code check schedules duplicate work; the correct disposition was close-as-completed referencing the fixing PR.
