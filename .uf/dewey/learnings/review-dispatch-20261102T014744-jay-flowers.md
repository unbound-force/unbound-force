---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:47:44Z
identity: review-dispatch-20261102T014744-jay-flowers
tier: draft
---

Backend-swap coupling in the review-dispatch plugin: the labeled-field text-extraction mechanism (extractAssessmentFieldValue in .opencode/plugins/review-dispatch/index.ts) is shared by parseAssessmentFields (read_assessment_fields' FR-003 backend, fix-triage-issue-dispatch-docs design D4) and parseCoverageAttestation (RD-FR-035, consumed by consolidate_dispatch's coverage-attestation advisory). Issue #755 swaps read_assessment_fields from persisted-text parsing to payload fields — the swap must keep the labeled-field extraction layer intact for parseCoverageAttestation even after read_assessment_fields stops calling it; deleting the text-parsing path wholesale would break RD-FR-035 and its 5-case parseCoverageAttestation suite in .opencode/test/review-dispatch-read-assessment-fields.test.ts. General rule: verify co-consumers of a shared mechanism before removing any "old backend". Related scope note: that test file now carries 41 it() cases across four describes (parseAssessmentFields 26, parseCoverageAttestation 5, read_assessment_fields tool 8, integration 2), while issue #755 cites 31 — test-adaptation scope cited in tracking issues goes stale as the same file accrues later features (RD-FR-035 landed after filing).
