---
tag: change-records
author: jay-flowers
category: gotcha
created_at: 2026-11-10T20:28:32Z
identity: change-records-20261110T202832-jay-flowers
tier: draft
---

When a gatekeeping-exception security fix-commit rides a spec-governed branch, fold it into the active change's record before review: acknowledge it in the proposal's Impact (a "no pins changed" claim must stay true at PR/branch level, not just at individual-change level), add its task, and amend any unarchived requirement that enumerates the affected pin set. unbound-force precedent: stop-shipping-plugin-lock task 4.7 sweeps "both manifests, the provenance record, the spec delta, the proposal, and the tests" for every pin bump; an effect-override fix-commit on PR #1006 that swept only manifests/provenance/tests/CHANGELOG left the proposal disclaiming a pin change the branch carried.
