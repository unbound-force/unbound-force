---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:30:54Z
identity: issue-triage-20261102T013054-jay-flowers
tier: draft
---

A narrow out-of-scope rationale in a corrective change rules out only the defect class it names. fix-triage-issue-dispatch-docs excluded /speckit.testreview because its command file has no proposals-array references, yet the command still carries lesson-proposal processing prose and a findings-consumption flow on a different tool path (plan_review_dispatch → invoke_agent → child submit_review_findings → parent dedup → finalize_review_dispatch, with zero dispatch_review_runs or consolidate_dispatch references). Two triage corollaries: (1) follow-up issues scoped to auditing the remaining prose are valid bug-class follow-ups, not duplicates of the original fix — duplicate comparison must match artifact and defect class, not topic (issue #760's correlation-ID-reuse plugin bug shares the dispatch topic but is TypeScript plugin runtime behavior, a different artifact and problem class than command-markdown accuracy). (2) "with scaffold mirror sync" clauses inherit the original change's mirror set — speckit.testreview.md has no scaffold mirror in this repo (internal/scaffold/assets/opencode/commands/ mirrors only the 13 uf.* commands), so triage and implementation should verify a mirror exists before treating mirror sync as in-scope work.
