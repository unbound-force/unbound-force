---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-06T18:13:51Z
identity: review-dispatch-20261106T181351-jay-flowers
tier: draft
---

Reviewer-capability scopes are category-based (documentation, standard, cli-ux, ci-cd, security, dependencies, test-quality) while persona review surfaces are path-based, so the minimal category set covering a surface can over-fire on out-of-surface PRs. On PR 959, divisor-determinism ([documentation, standard]) dispatches on Go-only PRs just to return its mandated no-findings APPROVE, and the sibling fact that divisor-sre ([ci-cd, dependencies, security]) scope-missed a [standard, test-quality] dispatch shows the same imprecision causes both over- and under-firing. The trap: narrowing a persona's scopes to save the per-run cost silently drops MUST-rule coverage for surface paths that share a category with out-of-surface code — dropping `standard` from divisor-determinism would orphan DE-006/DE-007 on .opencode/plugins and .opencode/lib PRs. The durable fix is path-gated include/skip in the dispatch planner (plan_review_dispatch already computes per-agent include/skip decisions with reason codes from changed file paths; the surface-intersection check is mechanical path matching — a DE-001 extraction target), pursued as its own gatekeeping-sensitive change rather than per-persona scope narrowing.
