---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:40:46Z
identity: issue-triage-20261102T034046-jay-flowers
tier: draft
---

Triage pattern for instruction-consolidation issues that predate tool evolution — verify three things before rendering a verdict. (1) Premise liveness: re-check the duplication the issue describes; in unbound-force issue #663 the "Shared Dispatch, Evidence, and Finalization Protocol" prose was repeated across four command files (uf.review-council, uf.triage-issue, uf.address-feedback, uf.review-pr), larger than the issue's ~500-token-per-file estimate. (2) Solution-sketch currency: the issue's enumerated lifecycle named dispatch_agent_run (step 4) and finalize_review_dispatch (step 7), both since superseded by the DE-005 intent-level surface — dispatch_review_runs (the dispatch-advisor skill itself now forbids individual dispatch_agent_run calls) and consolidate_dispatch-with-provenance finalization (every command contract states "Never call finalize_review_dispatch directly"). Stale tool names do not invalidate an issue whose "no behavioral change" acceptance criterion anchors implementation to current behavior; flag the reconciliation as an implementation constraint in the triage reasoning. (3) Drift evidence: duplicated protocol prose that has already diverged (uf.address-feedback.md line 1241 still mandated dispatch_agent_run while sibling commands mandated dispatch_review_runs) is the strongest objective evidence that the consolidation is needed — copy-paste protocol sections drift apart, which is the exact failure mode single-source documentation prevents.
