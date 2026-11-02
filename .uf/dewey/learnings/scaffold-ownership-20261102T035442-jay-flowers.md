---
tag: scaffold-ownership
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:54:42Z
identity: scaffold-ownership-20261102T035442-jay-flowers
tier: draft
---

Scaffold extension-point classification needs a drift signal when the deployed file carries a versioned schema. A project-owned file (isToolOwned -> false) is skipped on plain re-init with no content comparison, which protects local customizations but silently preserves stale schema versions — unbound-force issue #649: .uf/review-matrix.yaml stayed at v2 in generated projects while the canonical embedded asset moved to v3 (profiles/defaults/advisor/runs/risk_augmentation/limits), silently breaking review dispatch. When classifying a schema-bearing config file as a project-owned extension point, pair the classification with a version-field or canonical-hash drift check that warns on plain re-init, preserving the extension-point semantics (issues #884/#886 precedent) while making schema evolution visible. Companion rule: ownership-decision comments must state --force semantics explicitly — the uf/model-matrix.yaml branch's wording ("--force remains the explicit replace path") is the correct pattern; the review-matrix branch's "without --force replacing their policy" reads as force-protection the code does not provide (only isNeverOverwrite files, i.e. constitution.md, survive --force).
