---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:40:53Z
identity: issue-triage-20261102T034053-jay-flowers
tier: draft
---

When triaging prose-consolidation issues, verify the duplication by grepping the repository instead of trusting the issue's file list, and check whether tool evolution has superseded the issue's mechanism. Verified on unbound-force issue #663 (expand dispatch-advisor skill to full lifecycle): the fail-closed dispatch protocol prose it names in 3 command files is verbatim-duplicated across exactly those 3 (uf.review-council, uf.triage-issue, uf.address-feedback), but the same protocol family also lives in uf.review-pr.md, uf.unleash.md, and speckit.testreview.md; and its proposed step-4 mechanism (per-run dispatch_agent_run invocation) was superseded by the dispatch-review-runs change, which collapsed the sequence into the aggregate dispatch_review_runs tool that the skill's Invocation Tool section now mandates. The issue's intent (single-source the lifecycle in the skill) remained valid while its mechanism prescription and file scope had aged. Consolidation issues age in scope and mechanism faster than in intent — re-verify both at triage time.
