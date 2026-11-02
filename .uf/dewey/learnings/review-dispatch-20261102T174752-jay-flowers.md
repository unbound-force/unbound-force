---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-02T17:47:52Z
identity: review-dispatch-20261102T174752-jay-flowers
tier: draft
---

When a new vitest describe suite consumes fixtures from a sibling suite, hoisting one shared fixture to module level while copy-defining another in the same PR leaves inconsistent sharing discipline. In split-host-resolution-failure-classes (PR #920), `hostLookupClient` was correctly hoisted so the RD-FR-037 and RD-FR-038 suites share it, but `emptyStandardProfileDeps` was re-implemented suite-locally with a JSDoc comment pointing at the advisor-fallback suite's identical empty-profile matrix — a copy whose comment names its origin is a drift risk, because a future change to the origin fixture will not propagate to the copy. Audit every fixture a new suite borrows and hoist them together in the same edit.
