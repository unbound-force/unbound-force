---
tag: review-dispatch
author: jay-flowers
category: context
created_at: 2026-11-02T01:30:48Z
identity: review-dispatch-20261102T013048-jay-flowers
tier: draft
---

Triage verification of unbound-force/unbound-force#756 (speckit-testreview dispatch-data-flow audit): .opencode/commands/speckit.testreview.md contains zero dispatch_review_runs and zero consolidate_dispatch references — it orchestrates per-run invoke_agent calls, manual consolidation (§3.5), and direct finalize_review_dispatch (§3.7) — so audit checklist items targeting those two tools are null checks. Its lesson-proposal prose is accurate against the plugin contracts: submit_lesson_proposal takes {information, tag, optional category} with no reference field, and prepare_lesson_learning returns information, a generated sibling-* tag, and the hardcoded category "reference" (.opencode/lib/review-dispatch-lesson-proposal.ts, LESSON_CATEGORY). The substantive open audit surface is the findings-retrieval prose: the command never names the mechanism by which the parent obtains child-submitted findings for deduplication and finalize-payload construction (invoke_agent returns child text directly; submit_review_findings writes the dispatch store), unlike the corrected uf.review-pr §5a/5b which names the no-provenance consolidate_dispatch retrieval. Mirror caveat: speckit.testreview.md is absent from this repo's internal/scaffold/assets/opencode/commands/ mirror set and carries the "scaffolded by gaze v1.4.6" marker — its byte-identical mirror pair lives in the gaze repo, so remediation "scaffold mirror sync" is a cross-repo gaze concern, not the in-repo sync pattern from fix-triage-issue-dispatch-docs tasks 3.1-3.2.
