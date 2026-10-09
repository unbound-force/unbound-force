---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-10-06T00:45:37Z
identity: review-dispatch-20261006T004537-jay-flowers
tier: draft
---

The review-dispatch plugin's buildPlan() function uses a concept of "matrix mode" (the mode resolved from the review matrix, e.g. "code", "spec", "triage", "feedback") versus "command mode" (the mode from the user's input: "code", "specs", "triage", "feedback", "test"). These two mode spaces overlap but are not identical — notably, test mode maps to code matrix mode via matrixMode(). When adding mode-dependent guard conditions (like tier cap applicability in FR-005), always use input.mode (command mode) for the guard, not the resolved matrix mode variable. Using matrix mode causes test mode to incorrectly inherit code mode behavior. The code review guard agent caught this as a MEDIUM finding: "FR-005 mode guard uses matrix mode not command mode." The fix is to reference input.mode directly in the conditional.
