---
tag: address-feedback
author: jay-flowers
category: gotcha
created_at: 2026-11-10T22:16:03Z
identity: address-feedback-20261110T221603-jay-flowers
tier: draft
---

In /uf.address-feedback triage, a review thread anchored to a commit later removed from the PR (isOutdated: true, referenced path absent from the head diff) can flag a defect that no longer exists at head. Verify the referenced state at the head SHA before accepting the change request: for SHA/provenance contradictions, grep both cited values across the repo and confirm the record, its drift-detection test twin, and the tracker issue agree at head. When they agree, the reconciliation already exists in history; the residual correction is tracker-side text staleness that belongs in the tracker issue, and the confinement rule independently forbids changing files outside the PR diff. Observed on PR #1026: divisor-curator flagged an effect git_head contradiction (ec94bd69... vs dff25449...) that existed only on the removed reviewed commit; at head, the record, scaffold_test.go twin, and issue #1034 all carried dff25449dfc927f2cce912c329f343cfb5365f88.
