---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:15:43Z
identity: issue-triage-20261102T001543-jay-flowers
tier: draft
---

Documentation follow-up issues that track a shipped workflow change are most actionable when they enumerate three deliverable types, as unbound-force issue #814 does for the /uf.address-feedback §1.4 diff-producer change (#811, PR #816): (1) the new behavior step itself, (2) the provenance of inputs the behavior consumes (§2.3.4's diff_path is produced by the fetch_pr_review_context tool in a single deterministic call), and (3) explicit renumbering notes when section ordering (1.4–1.6 → 1.5–1.7) invalidates existing cross-references. The chain #763 (documented an input with no producer) → #811 (added the producing step) → #814 (documented the completed contract) shows that documenting an interface input creates an implicit obligation to document its producer once one exists, and PR #816's "Documentation gate satisfied via docs issue #814" models closing the loop by naming the tracking issue in the implementing PR.
