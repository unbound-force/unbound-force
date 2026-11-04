---
tag: review-process
author: jay-flowers
category: pattern
created_at: 2026-11-04T14:24:18Z
identity: review-process-20261104T142418-jay-flowers
tier: draft
---

A PR that closes N issues must be audited for change-record completeness before merge: enumerate every issue in the PR's closes list against the governing change's What Changes, design decisions, FR delta, tasks.md, and the CHANGELOG entries' Spec: references. Verified on PR 949 (review-dispatch-consumer-bounds, 2026-11-03): the PR closed five issues (#941 #945 #943 #946 #942) but the change artifacts documented only the first three; the two extra fixes — a dispatch_review_runs summary error suffix (#946) and a uf.review-pr self-authored verdict mapping (#942) — shipped as mid-branch commits whose CHANGELOG entries lacked the Spec: reference the change's own task 4.2 specified, with task 4.2 already checked complete. Behavioral changes to review tooling (tool output contracts, verdict-to-event mappings) are audit-trail obligations: an unrecorded change to how review verdicts post cannot be distinguished from drift in a later audit. The reconciliation audit is a fixed enumeration (closes list × record surfaces) and belongs at review time in the driver command; the remedy itself is doc-only — a bundled-fixes note or What Changes bullet, a task each, and Spec: tails — so accepting such feedback is cheap and correct whenever the shipped code is sound. Complements review-process-20261015 (fix known-wrong citations in the same PR) and code-review-20261027 (audit every parallel instance of a pattern a fix addresses).
