---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:18:32Z
identity: issue-triage-20261102T011832-jay-flowers
tier: draft
---

When triaging an issue filed from a review-council advisory (an APPROVE WITH ADVISORIES follow-up), verify the current tree before routing it to a new fix: the advisory may already have been resolved by a later iteration of the same change that filed it. Observed with unbound-force#771 (redispatch run-file persistence verification, filed 2026-10-20 from fix-review-dispatch-recovery iteration 2): the requested persisted-flag gate, restore-on-failure path, and dispatch-recovery.test.ts coverage all landed the same day via PR #776's iteration 3, leaving the issue describing a defect that no longer exists in the tree. In-code comments citing the PR/requirement IDs (e.g., "PR #776, DR-FR-002") are the fastest verification signal. Triaging from the issue text alone would have opened a duplicate work stream for already-fixed work; check the fix locus in the current tree first and close or fast-verify instead.
