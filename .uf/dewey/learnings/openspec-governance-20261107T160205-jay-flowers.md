---
tag: openspec-governance
author: jay-flowers
category: pattern
created_at: 2026-11-07T16:02:05Z
identity: openspec-governance-20261107T160205-jay-flowers
tier: draft
---

When a quick-tier changeset trips tier-routing risk floors (data-loss path, breaking plugin-tool contract) and the council demands a durable spec record, the workable retroactive pattern from PR 948's retroactive-push-skill-contracts change is: one retroactive OpenSpec change carrying (1) MODIFIED requirement restatements derived from the latest-generation delta text (no new FR ids, numbering allocation carried forward), (2) light inline superseded-by annotations at every contradicted live delta — placed inside the span that archivers materialize, not in a separate index — each naming the superseding change and the archive order, and (3) archive-order constraints recorded redundantly in every surface a future archiver consults (delta header comments, requirement prose, proposal.md, design.md), because OpenSpec has no mechanical cross-change ordering guard. Validation should run openspec validate --strict across the whole unarchived stack, not just the new change: PR 948's validation surfaced a pre-existing mis-pathed sibling delta (specs/plugin-tools.md instead of specs/plugin-tools/spec.md) whose repair is outside the retroactive change's scope but must be documented as an archive precondition so the failure mode is fail-safe (archive refuses: target spec does not exist) rather than silent.
