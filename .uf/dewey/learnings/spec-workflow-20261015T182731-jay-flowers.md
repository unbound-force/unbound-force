---
tag: spec-workflow
author: jay-flowers
category: gotcha
created_at: 2026-10-15T18:27:31Z
identity: spec-workflow-20261015T182731-jay-flowers
tier: draft
---

When a review-phase Dewey learning produced during the same change contradicts a design.md decision or a proposal claim, reconcile the artifact text before merge. Observed on opsx/address-feedback-model-normalization (PR #733): design D4 claimed the delta spec's scenarios were "enforced by existing drift tests plus the existing tool tests," while the command-contracts learning shipped in the same PR established that content-absence requirements (MUST NOT clauses on command text) are enforced by neither and need negative command_contract_test.go clauses; the proposal's "zero behavioral change" claim was likewise contradicted by the tool-extraction learning documenting the tool's guarded -N strip. Shipping both the stale rationale and its draft-tier disproving learning leaves the repo with an internally inconsistent change record — future contributors trust the design doc, not the learning. Reconciliation means either acting on the learning (add the missing enforcement clause) or amending the rationale to cite the learning as the tracked follow-up.
