---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:49:50Z
identity: issue-triage-20261102T014950-jay-flowers
tier: draft
---

Issue-triage verification pattern: when an issue pins its evidence to a specific commit, verify the cited lines at current HEAD and search the surrounding code for issue-number references in design comments — remediation changes in this organization cite their driving issues (e.g., "design D5 of fix-review-dispatch-recovery, issue #751"), so a landed fix leaves a discoverable trail from the reported defect to follow-up issues and design rationale. Applied to unbound-force#746 (conversation-keyed dispatch session silently breaking per-item Tier 2 finalization): verification connected the report to root-cause issue #794, confirmed both of the issue's proposed fix directions now ship (per-invocation correlation minting in dispatch_review_runs at index.ts:5588-5603; session metadata persisted whenever supplied with replace-with-latest semantics at index.ts:3467-3482), and let the triage verdict carry resolved-by lifecycle state in addition to validity classification.
