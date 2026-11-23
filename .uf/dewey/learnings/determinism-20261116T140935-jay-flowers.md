---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-16T14:09:35Z
identity: determinism-20261116T140935-jay-flowers
tier: draft
as_of: "unbound-force@c9e35b82b9b5a4718abfc2f4bd320eb849aa60c1"
---

When a review fix migrates mechanical prose into a tool mid-PR, the fix commit usually brings the seam tests but not a change-record update: proposal.md and tasks.md keep describing the pre-fix prose design. On PR #1157 (per-task-delegation-budgets), commits ca333391/c9e35b82 moved the DE-005 retry routing and the DE-001 remaining_ms fill into executePerTaskDelegation and the checkpoint tool with green seam tests, while proposal.md item 3 and tasks.md task 3.2 still described the caller-relays-retry design the fixes replaced. Amend the change record (and its seam-verification task) in the same fix commit so no future reader reconstructs the violating delivery shape from the proposal.
