---
tag: documentation-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-05T17:48:41Z
identity: documentation-pipeline-20261105T174841-jay-flowers
tier: draft
---

Verifying a checked documentation-gate task in unbound-force/unbound-force requires searching both the `docs` and `documentation` GitHub labels or running a keyword search across all labels: the label namespace is unstandardized (tracked by issue #825), and the review-dispatch-consumer-bounds gate issue #960 carries only `documentation`, so the standard `gh issue list --label docs` duplicate check misses it. When a tasks.md docs-gate task records no issue numbers (consumer-bounds task 5.3, unlike preflight-declared-output task 4.4 which records #951 and website#373), the verification path is a `--state all --search "<change-name>"` keyword search in both unbound-force/unbound-force and unbound-force/website, then confirming the issue body names the change and covers its user-facing surface.
