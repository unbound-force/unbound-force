---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-06T18:13:53Z
identity: review-dispatch-20261106T181353-jay-flowers
tier: draft
---

Category-based reviewer scopes in .uf/reviewer-capabilities.yaml cannot express path-declared review surfaces: .opencode/lib/*.ts tool code and internal/**/*.go both classify as `standard`, while .opencode/test/*.test.ts classifies as `test-quality`. A persona registered under `standard` to cover its tool-code surface therefore over-fires on Go-only PRs (contractually zero findings) and under-fires on test-only plugin PRs. The coarseness is symmetric and was demonstrated live during PR 959 feedback review: divisor-sre ([ci-cd, dependencies, security]) scope-missed a [standard, test-quality] dispatch. The durable fix is path-aware scope selection in the dispatch planner, not scope-list tuning — changing scope lists only trades over-firing for under-firing.
