---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:14:34Z
identity: issue-triage-20261102T011434-jay-flowers
tier: draft
---

Deferred-work follow-up issues triage-verifiably when they carry bidirectional traceability anchors. Issue #770 (dispatch prompt/sibling-evidence context budgets) validated on four checks: (1) the deferring change's proposal names the tracking issue explicitly ("Prompt/evidence trimming is tracked in #770, not implemented here" — fix-review-dispatch-recovery/proposal.md), (2) the parent incident #751 documents the concrete root cause (a child session hit the ~363k input-token context ceiling during a 5-reviewer dispatch, bricking finalization), (3) the observability prerequisite shipped first (D6 per-run usage reporting in dispatch_status/consolidate_dispatch makes budget overruns diagnosable before enforcement — measure-then-enforce sequencing), and (4) the gap is confirmed against live tool surfaces: budget_usd skipping exists on dispatch_review_runs but no token-budget parameter exists, and acquire_sibling_evidence enforces per-file size bounds only, with no aggregate per-run budget or truncate/omit policy. Adjacent-but-complementary shipped work (RD-FR-032 manifest diff delivery, the #880 direction) bounds only the diff dimension and does not close the request — grep active openspec/changes for the issue number to rule out in-flight implementation before validating.
