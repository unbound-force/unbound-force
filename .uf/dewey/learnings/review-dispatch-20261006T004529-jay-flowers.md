---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-10-06T00:45:29Z
identity: review-dispatch-20261006T004529-jay-flowers
tier: draft
---

When implementing changes to the review-dispatch plugin that adds new fields to the DispatchPlan interface (like the advisories field added in the tier-gated-agent-count change), always check the failedPlan() helper function. This function constructs a DispatchPlan object for error cases and must include every required field from the interface. Missing a field there causes runtime TypeErrors when consumers access properties on failed plans. The code review council caught this as a HIGH consensus finding across three of four agents (adversary, architect, guard). The fix is trivial (add the field with its zero value, e.g. advisories: []) but the bug is insidious because failedPlan() is only exercised on error paths that may not have test coverage.
