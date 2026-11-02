---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-01T22:39:17Z
identity: issue-triage-20261101T223917-jay-flowers
tier: draft
---

A documentation-gate tracking issue's body can go stale against its own change's final dispositions when a review round re-disposes a content item during the implementing PR. Observed triaging unbound-force issue #862 (uf-fix-ci docs impact, 2026-11-01): the issue body as fetched records "Content coverage: tutorial no (bounded single-command repair workflow...)" — the original disposition — while the change's proposal (openspec/changes/uf-fix-ci/proposal) records that round-1 review re-disposed the tutorial assessment (the command meets the tutorial significance threshold, mirroring the /uf.rebase-pr precedent #855) and filed tutorial issue #870; the CHANGELOG separately confirms the "(Docs: #862, unbound-force/website#342)" footer executed as claimed. Because the tracking issue is the auditable record the documentation gate points at, a re-disposition made during PR review should be written back to the tracking-issue body (or appended as a comment) at re-disposition time. Triage corroboration rule: cross-check each content-coverage disposition in a docs-gate issue body against the change's proposal/tasks final state, not only against repository file state — a body/proposal mismatch on a disposition signals the audit trail desynced and warrants a body-update recommendation, not an invalidation. Complements issue-triage-20261101T221608 (disposition-matrix verification against repo state) and content-pipeline-20261029T154738 (disposition-rationale expiry) by adding the review-round re-disposition case.
