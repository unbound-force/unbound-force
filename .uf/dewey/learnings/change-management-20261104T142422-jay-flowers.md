---
tag: change-management
author: jay-flowers
category: pattern
created_at: 2026-11-04T14:24:22Z
identity: change-management-20261104T142422-jay-flowers
tier: draft
---

A spec-governed branch that bundles issue-tracked bug fixes beyond the governing change's recorded scope produces a detectable record-vs-ship mismatch: the PR closes-list is longer than the change proposal's Fixes line, and the bundled fixes' CHANGELOG entries appear without the change's (Spec: ..., Fixes: #N) tail. The fix is documentation-only (MP-009 — a change must not silently outgrow its artifact shape): a What Changes bullet or 'Bundled fixes' note per fix, a tasks.md entry each for traceability, and the Spec: tail on the CHANGELOG entries; no code change. Verified on PR 949 (review-dispatch-consumer-bounds): the change recorded 'Fixes #941, #945, #943' while the PR also shipped #946 (dispatch_review_runs summary error suffix, dispatchSummaryLine) and #942 (uf.review-pr self-authored verdict mapping) as standalone commits with Spec-less CHANGELOG entries. Extends learning/change-management-20261019T204943-jay-flowers (the change record must match what ships at merge; fix the record rather than defer).
