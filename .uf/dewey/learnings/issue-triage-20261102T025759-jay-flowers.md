---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:57:59Z
identity: issue-triage-20261102T025759-jay-flowers
tier: draft
---

Tool-extraction issues (Constitution Principle VI / DE-001 prose→tool migrations) should be triaged against existing internal implementations before accepting a "create new tool" proposal. unbound-force issue #673 proposed a new classify_paths tool in uf-workflow, but .opencode/plugins/review-dispatch/index.ts already contained the same path→category classifier (change-profile computation feeding plan_review_dispatch, lines ~1195-1252), and .opencode/skills/review-context/SKILL.md carried a parallel glob-pattern table (Protocol 3) applied by the LLM at runtime. The member redirect — extract the existing classifier to .opencode/lib/ as a shared pure function (the classifyBranch/sanitizeSlug/matchTaskLine precedent) and expose only a thin wrapper — prevents the two-classifiers-drift failure mode the shared-primitives pattern exists to stop. Triage signal: location-pinned acceptance criteria ("tool added to <plugin>, both copies") are refinable when comments surface an existing implementation; the verifiable problem statement, not the proposed mechanism, drives the VALID verdict.
