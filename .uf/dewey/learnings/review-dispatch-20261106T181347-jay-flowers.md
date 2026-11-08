---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-06T18:13:47Z
identity: review-dispatch-20261106T181347-jay-flowers
tier: draft
---

Category-based reviewer-capability scopes cannot express path-bounded persona review surfaces, forcing a cost-vs-coverage trade-off on every path-narrow persona registration. Registering under a broad category (divisor-determinism under `standard`) fires the persona on every council in that category — including PRs entirely outside its surface where its own instructions require a no-findings APPROVE — a recurring no-op child run per council iteration that also marginally enlarges the dispatch failure surface. Narrowing the category (dropping `standard`) orphans the surfaces that share the category with unrelated code: .opencode/plugins+lib tool code shares `standard` with all Go code, so DE-006/DE-007 coverage on pure tool-code PRs is lost, and no path-scoped preflight lane covers tool-code seams. The mismatch under-selects too — divisor-sre scope-missed a standard+test-quality feedback dispatch because its scopes (ci-cd, dependencies, security) did not intersect the item's categories. Structural fix: extend the reviewer-capability schema with path-based scopes or add planner path-gating; the in-repo precedent is .uf/preflight.yaml declared checks, which already bind checks to file paths. Until then, prefer the coverage-preserving registration, accept the bounded no-op cost, and record the trade-off with measured run/token cost in the change record.
