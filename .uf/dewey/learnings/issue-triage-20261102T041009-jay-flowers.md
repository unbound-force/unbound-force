---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T04:10:09Z
identity: issue-triage-20261102T041009-jay-flowers
tier: draft
---

Cross-repo scaffold-refresh issues filed in a meta repo (e.g. "re-run uf init across an external org's repositories") are tracking issues, not implementable work for an agent dispatched against the meta repo. Effective triage confirms three structural facts: the precedent PRs that define the per-repo procedure, the enumerable list of remaining target repos, and — the sequencing gotcha — whether open defects in the scaffold templates themselves (hardcoded org framing, constitution leakage, stale command names) land before the sweep. A mechanical re-init run ahead of the template fixes propagates known-defective content into every adopter repo, converting one template fix into N per-repo manual corrections.
