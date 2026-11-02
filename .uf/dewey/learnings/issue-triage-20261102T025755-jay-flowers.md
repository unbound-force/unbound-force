---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:57:55Z
identity: issue-triage-20261102T025755-jay-flowers
tier: draft
---

When triaging an issue that proposes encoding pattern-matching or classification logic into a new plugin tool, verify whether the same logic already exists elsewhere in the codebase before accepting "encode the patterns in the new tool" acceptance criteria. In unbound-force, path classification (categories, security_sensitive, user_facing, tier) already exists as the change-profile computation inside .opencode/plugins/review-dispatch/index.ts; re-implementing it in a new uf-workflow tool would create two classifiers that drift apart — the exact failure mode the shared-primitives pattern exists to prevent. The correct shape is extracting the existing implementation to .opencode/lib/ as a shared pure function (classifyBranch/matchTaskLine/sanitizeSlug precedent) with the new tool as a thin wrapper. When a member comment corrects an issue's proposed solution, fold the correction into the issue body or acceptance criteria before implementation starts, so a cold-start implementer cannot follow the superseded shape.
